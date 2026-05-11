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

	// Confirm modal state
	let confirmModal = $state<{
		open: boolean;
		type: 'add' | 'delete';
		categoryName: string;
		categoryId: number | null;
		pendingName: string;
	}>({
		open: false,
		type: 'add',
		categoryName: '',
		categoryId: null,
		pendingName: ''
	});

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

	function closeModal(): void {
		confirmModal.open = false;
	}

	function handleSubmit(event: SubmitEvent): void {
		event.preventDefault();
		clearMessages();

		const trimmedName = name.trim();
		if (!trimmedName) {
			errorMessage = 'Nama kategori wajib diisi.';
			return;
		}

		if (editingId) {
			// Edit langsung tanpa konfirmasi
			doSave(trimmedName);
		} else {
			// Tampilkan konfirmasi tambah
			confirmModal = {
				open: true,
				type: 'add',
				categoryName: trimmedName,
				categoryId: null,
				pendingName: trimmedName
			};
		}
	}

	async function doSave(trimmedName: string): Promise<void> {
		saving = true;
		clearMessages();
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

	function requestDelete(category: Category): void {
		confirmModal = {
			open: true,
			type: 'delete',
			categoryName: category.name,
			categoryId: category.id,
			pendingName: ''
		};
	}

	async function confirmAction(): Promise<void> {
		closeModal();
		if (confirmModal.type === 'add') {
			await doSave(confirmModal.pendingName);
		} else if (confirmModal.type === 'delete' && confirmModal.categoryId !== null) {
			await doDelete(confirmModal.categoryId);
		}
	}

	async function doDelete(id: number): Promise<void> {
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
								onclick={() => requestDelete(category)}>Hapus</button
							>
						</div>
					</li>
				{/each}
			</ul>
		{/if}
	</section>
</section>

{#if confirmModal.open}
	<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
	<div class="modal-backdrop" onclick={closeModal}>
		<div class="modal-box" onclick={(e) => e.stopPropagation()} role="dialog" aria-modal="true">
			{#if confirmModal.type === 'delete'}
				<p class="modal-icon">⚠</p>
				<h3 class="modal-title">Hapus Kategori?</h3>
				<p class="modal-body">
					Kategori <strong>"{confirmModal.categoryName}"</strong> akan dihapus secara permanen.
					Tindakan ini tidak dapat dibatalkan.
				</p>
				<div class="modal-actions">
					<button class="button-secondary" type="button" onclick={closeModal}>Batal</button>
					<button class="button-danger" type="button" onclick={confirmAction}>Ya, Hapus</button>
				</div>
			{:else}
				<p class="modal-icon">✦</p>
				<h3 class="modal-title">Tambah Kategori?</h3>
				<p class="modal-body">
					Kategori baru <strong>"{confirmModal.categoryName}"</strong> akan ditambahkan ke daftar.
				</p>
				<div class="modal-actions">
					<button class="button-secondary" type="button" onclick={closeModal}>Batal</button>
					<button class="button-primary" type="button" onclick={confirmAction}>Ya, Tambah</button>
				</div>
			{/if}
		</div>
	</div>
{/if}

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

	/* ── Confirm Modal ── */
	.modal-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.65);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 200;
		padding: 1.5rem;
		backdrop-filter: blur(2px);
	}

	.modal-box {
		background: var(--paper);
		border: 1px solid var(--rule);
		padding: 2rem 1.75rem 1.75rem;
		max-width: 22rem;
		width: 100%;
		text-align: center;
	}

	.modal-icon {
		font-size: 1.75rem;
		margin: 0 0 0.75rem;
		color: var(--ochre);
		line-height: 1;
	}

	.modal-title {
		font-family: var(--font-display);
		font-size: 1.35rem;
		color: var(--ink);
		margin: 0 0 0.6rem;
	}

	.modal-body {
		font-size: 0.875rem;
		color: var(--ink-muted);
		line-height: 1.55;
		margin: 0 0 1.5rem;
	}

	.modal-body strong {
		color: var(--ink);
	}

	.modal-actions {
		display: flex;
		gap: 0.75rem;
		justify-content: center;
	}

	.modal-actions .button-secondary,
	.modal-actions .button-primary,
	.modal-actions .button-danger {
		flex: 1;
	}

	.button-danger {
		background: var(--oxblood);
		color: var(--paper);
		border: 1px solid var(--oxblood);
		padding: 0.6rem 1.25rem;
		font-family: var(--font-mono);
		font-size: 0.75rem;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		cursor: pointer;
		transition: opacity 0.15s;
	}

	.button-danger:hover {
		opacity: 0.85;
	}
</style>
