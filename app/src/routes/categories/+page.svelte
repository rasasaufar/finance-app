<script lang="ts">
	import { onMount } from 'svelte';
	import { api, ApiError } from '$lib/api';
	import type { Category } from '$lib/types';
	import { categoryColor } from '$lib/format';

	let loading = $state(true);
	let saving = $state(false);
	let categories = $state<Category[]>([]);
	let editingId = $state<number | null>(null);
	let name = $state('');
	let errorMessage = $state('');
	let successMessage = $state('');

	const todayLabel = new Intl.DateTimeFormat('id-ID', {
		weekday: 'long',
		day: 'numeric',
		month: 'long',
		year: 'numeric'
	}).format(new Date());

	function clearMessages(): void {
		errorMessage = '';
		successMessage = '';
	}

	async function loadCategories(): Promise<void> {
		loading = true;
		clearMessages();
		try {
			categories = await api.get<Category[]>('/categories');
		} catch (error) {
			if (error instanceof ApiError) {
				errorMessage = error.message;
			} else {
				errorMessage = 'Gagal memuat data kategori.';
			}
		} finally {
			loading = false;
		}
	}

	function startEdit(category: Category): void {
		editingId = category.id;
		name = category.name;
	}

	function cancelEdit(): void {
		editingId = null;
		name = '';
	}

	async function handleSubmit(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		clearMessages();

		const trimmedName = name.trim();
		if (!trimmedName) {
			errorMessage = 'Nama kategori wajib diisi.';
			return;
		}

		saving = true;
		try {
			if (editingId) {
				await api.put<Category>(`/categories/${editingId}`, { name: trimmedName });
				successMessage = 'Kategori berhasil diperbarui.';
			} else {
				await api.post<Category>('/categories', { name: trimmedName });
				successMessage = 'Kategori berhasil ditambahkan.';
			}

			await loadCategories();
			cancelEdit();
		} catch (error) {
			if (error instanceof ApiError) {
				errorMessage = error.message;
			} else {
				errorMessage = 'Gagal menyimpan kategori.';
			}
		} finally {
			saving = false;
		}
	}

	async function handleDelete(id: number): Promise<void> {
		clearMessages();
		try {
			await api.delete<void>(`/categories/${id}`);
			successMessage = 'Kategori berhasil dihapus.';
			await loadCategories();
			if (editingId === id) {
				cancelEdit();
			}
		} catch (error) {
			if (error instanceof ApiError) {
				errorMessage = error.message;
			} else {
				errorMessage = 'Gagal menghapus kategori.';
			}
		}
	}

	onMount(() => {
		loadCategories();
	});
</script>

<section class="page">
	<header class="page-header">
		<div class="page-header-top">
			<span><span class="issue-mark">§</span> 05 · Daftar Istilah</span>
			<span>{todayLabel}</span>
		</div>
		<div class="page-header-main">
			<div>
				<h1 class="page-title">Daftar <em>Kategori</em></h1>
				<p class="page-subtitle">
					Taksonomi pribadi untuk mengelompokkan setiap transaksi.
				</p>
			</div>
		</div>
	</header>

	<section class="section-card">
		<h2 class="section-title">{editingId ? 'Edit Kategori' : 'Tambah Kategori'}</h2>
		<p class="section-lede">
			Gunakan nama yang ringkas dan konsisten agar laporan tetap rapi.
		</p>

		{#if errorMessage}
			<p class="error">{errorMessage}</p>
		{/if}
		{#if successMessage}
			<p class="success">{successMessage}</p>
		{/if}

		<form class="form-grid" onsubmit={handleSubmit}>
			<label class="field">
				<span>Nama Kategori</span>
				<input type="text" bind:value={name} placeholder="Contoh: Pendidikan" required />
			</label>

			<div class="button-row">
				<button class="button-primary" type="submit" disabled={saving}>
					{saving ? 'Menyimpan…' : editingId ? 'Perbarui Kategori' : 'Tambah Kategori'}
				</button>
				{#if editingId}
					<button class="button-secondary" type="button" onclick={cancelEdit}>Batal Edit</button>
				{/if}
			</div>
		</form>
	</section>

	<section class="section-card">
		<h2 class="section-title">Daftar Kategori</h2>
		<p class="section-lede">
			Total {categories.length} kategori aktif.
		</p>
		{#if loading}
			<p class="muted mono">Memuat daftar kategori…</p>
		{:else if categories.length === 0}
			<p class="muted">Belum ada kategori. Silakan tambah yang pertama.</p>
		{:else}
			<ul class="cat-list">
				{#each categories as category, i}
					<li class="cat-row" class:is-editing={editingId === category.id}>
						<span class="cat-num">{String(i + 1).padStart(2, '0')}</span>
						<span class="cat-dot cat-dot-lg" style="background: {categoryColor(category.name)};"></span>
						<span class="cat-name">{category.name}</span>
						<span class="leader"></span>
						<div class="cat-actions">
							<button class="button-ghost" type="button" onclick={() => startEdit(category)}
								>Edit</button
							>
							<button
								class="button-ghost danger"
								type="button"
								onclick={() => handleDelete(category.id)}>Hapus</button
							>
						</div>
					</li>
				{/each}
			</ul>
		{/if}
	</section>
</section>

<style>
	.cat-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: grid;
		gap: 0;
		border-top: 1px solid var(--rule);
	}

	.cat-dot-lg {
		width: 0.7rem !important;
		height: 0.7rem !important;
		margin-right: 0 !important;
	}

	.cat-row {
		display: flex;
		align-items: baseline;
		gap: 0.75rem;
		padding: 0.85rem 0;
		border-bottom: 1px solid var(--rule);
	}

	.cat-row.is-editing {
		background: var(--ochre-soft);
		margin: 0 -1rem;
		padding-left: 1rem;
		padding-right: 1rem;
	}

	.cat-num {
		font-family: var(--font-mono);
		font-size: 0.7rem;
		letter-spacing: 0.15em;
		color: var(--ink-faint);
		min-width: 1.5rem;
	}

	.cat-name {
		font-family: var(--font-display);
		font-size: 1.25rem;
		color: var(--ink);
		line-height: 1;
	}

	.cat-actions {
		display: flex;
		gap: 1rem;
		flex-shrink: 0;
	}

	.button-ghost.danger {
		color: var(--oxblood);
		border-bottom-color: var(--oxblood);
	}

	.button-ghost.danger:hover {
		color: var(--ink);
		border-bottom-color: var(--ink);
	}
</style>
