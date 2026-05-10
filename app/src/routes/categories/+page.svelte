<script lang="ts">
	import { onMount } from 'svelte';
	import { api, ApiError } from '$lib/api';
	import type { Category } from '$lib/types';

	let loading = $state(true);
	let saving = $state(false);
	let categories = $state<Category[]>([]);
	let editingId = $state<number | null>(null);
	let name = $state('');
	let errorMessage = $state('');
	let successMessage = $state('');

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
		<div>
			<h1 class="page-title">Kategori</h1>
			<p class="page-subtitle">Kelola kategori transaksi sesuai kebutuhan pribadi Anda.</p>
		</div>
	</header>

	<section class="section-card">
		<h2 class="section-title">{editingId ? 'Edit Kategori' : 'Tambah Kategori'}</h2>

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
					{saving ? 'Menyimpan...' : editingId ? 'Perbarui Kategori' : 'Tambah Kategori'}
				</button>
				{#if editingId}
					<button class="button-secondary" type="button" onclick={cancelEdit}>Batal Edit</button>
				{/if}
			</div>
		</form>
	</section>

	<section class="section-card">
		<h2 class="section-title">Daftar Kategori</h2>
		{#if loading}
			<p class="muted">Memuat kategori...</p>
		{:else if categories.length === 0}
			<p class="muted">Belum ada kategori.</p>
		{:else}
			<div class="list">
				{#each categories as category}
					<article class="list-item">
						<div class="list-item-header">
							<strong>{category.name}</strong>
						</div>
						<div class="button-row">
							<button class="button-secondary" type="button" onclick={() => startEdit(category)}
								>Edit</button
							>
							<button class="button-danger" type="button" onclick={() => handleDelete(category.id)}>
								Hapus
							</button>
						</div>
					</article>
				{/each}
			</div>
		{/if}
	</section>
</section>
