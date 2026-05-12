<script lang="ts">
	import { onMount } from 'svelte';
	import { api, ApiError } from '$lib/api';
	import type { AccountResponse } from '$lib/types';

	let accounts = $state<AccountResponse[]>([]);
	let loading = $state(true);
	let errorMessage = $state('');

	// Modal state
	let modalOpen = $state(false);
	let modalMode = $state<'create' | 'edit'>('create');
	let editingAccount = $state<AccountResponse | null>(null);

	// Form fields
	let formName = $state('');
	let formEmail = $state('');
	let formPassword = $state('');
	let formRole = $state('user');
	let formError = $state('');
	let formSaving = $state(false);

	// Delete confirm
	let deleteTarget = $state<AccountResponse | null>(null);
	let deleteConfirmOpen = $state(false);
	let deleting = $state(false);

	async function loadAccounts(): Promise<void> {
		loading = true;
		errorMessage = '';
		try {
			accounts = await api.get<AccountResponse[]>('/admin/accounts');
		} catch (err) {
			errorMessage = err instanceof ApiError ? err.message : 'Gagal memuat daftar akun.';
		} finally {
			loading = false;
		}
	}

	function openCreate(): void {
		modalMode = 'create';
		editingAccount = null;
		formName = '';
		formEmail = '';
		formPassword = '';
		formRole = 'user';
		formError = '';
		modalOpen = true;
	}

	function openEdit(account: AccountResponse): void {
		modalMode = 'edit';
		editingAccount = account;
		formName = account.name;
		formEmail = account.email;
		formPassword = '';
		formRole = account.role;
		formError = '';
		modalOpen = true;
	}

	function closeModal(): void {
		modalOpen = false;
	}

	async function handleSubmit(e: Event): Promise<void> {
		e.preventDefault();
		formError = '';
		formSaving = true;

		try {
			if (modalMode === 'create') {
				const created = await api.post<AccountResponse>('/admin/accounts', {
					name: formName,
					email: formEmail,
					password: formPassword,
					role: formRole
				});
				accounts = [...accounts, created];
			} else if (editingAccount) {
			const body: Record<string, string> = {
					name: formName,
					email: formEmail,
					role: formRole
				};
				if (formPassword) body.password = formPassword;

				const updated = await api.put<AccountResponse>(
					`/admin/accounts/${editingAccount.id}`,
					body
				);
				accounts = accounts.map((a) => (a.id === updated.id ? updated : a));
			}
			modalOpen = false;
		} catch (err) {
			formError = err instanceof ApiError ? err.message : 'Gagal menyimpan akun.';
		} finally {
			formSaving = false;
		}
	}

	function confirmDelete(account: AccountResponse): void {
		deleteTarget = account;
		deleteConfirmOpen = true;
	}

	async function handleDelete(): Promise<void> {
		if (!deleteTarget) return;
		deleting = true;
		try {
			await api.delete(`/admin/accounts/${deleteTarget.id}`);
			accounts = accounts.filter((a) => a.id !== deleteTarget!.id);
			deleteConfirmOpen = false;
			deleteTarget = null;
		} catch (err) {
			errorMessage = err instanceof ApiError ? err.message : 'Gagal menghapus akun.';
			deleteConfirmOpen = false;
		} finally {
			deleting = false;
		}
	}

	function getInitials(name: string): string {
		return name
			.split(' ')
			.map((n) => n[0])
			.join('')
			.substring(0, 2)
			.toUpperCase();
	}

	onMount(loadAccounts);
</script>

<div class="page">
	<div class="page-header">
		<div class="header-text">
			<p class="page-eyebrow mono">Admin · Manajemen Akun</p>
			<h1 class="page-title">Kelola Akun</h1>
			<p class="page-desc">Buat dan atur akun pengguna. Setiap akun memiliki data terpisah.</p>
		</div>
		<button class="button-primary add-btn" type="button" onclick={openCreate}>
			<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
			Tambah Akun
		</button>
	</div>

	{#if errorMessage}
		<div class="alert-error">
			<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
			{errorMessage}
		</div>
	{/if}

	{#if loading}
		<div class="loading-state">
			<div class="loading-spinner"></div>
			<p>Memuat akun…</p>
		</div>
	{:else if accounts.length === 0}
		<div class="empty-state">
			<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round" width="48" height="48"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>
			<p>Belum ada akun selain akun Anda.</p>
			<button class="button-primary" type="button" onclick={openCreate}>Tambah Akun Pertama</button>
		</div>
	{:else}
		<div class="accounts-grid">
			{#each accounts as account (account.id)}
				<div class="account-card" class:is-admin={account.role === 'admin'}>
					<div class="card-top">
						<div class="account-avatar" class:avatar-admin={account.role === 'admin'}>
							{getInitials(account.name)}
						</div>
						<div class="account-badge" class:badge-admin={account.role === 'admin'}>
							{account.role === 'admin' ? 'Admin' : 'User'}
						</div>
					</div>
					<div class="account-info">
						<p class="account-name">{account.name}</p>
						<p class="account-email">{account.email}</p>
					</div>
					<div class="card-actions">
						<button
							class="action-btn edit-btn"
							type="button"
							onclick={() => openEdit(account)}
							title="Edit akun"
						>
							<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="14" height="14"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
							Edit
						</button>
						<button
							class="action-btn delete-btn"
							type="button"
							onclick={() => confirmDelete(account)}
							title="Hapus akun"
						>
							<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="14" height="14"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/><path d="M10 11v6M14 11v6"/><path d="M9 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/></svg>
							Hapus
						</button>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Create / Edit Modal -->
{#if modalOpen}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
	<div class="modal-backdrop" onclick={closeModal} role="presentation">
		<div class="modal-card" onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1">
			<p class="muted mono" style="margin-bottom: 0.35rem; font-size: 0.6rem; letter-spacing: 0.18em; text-transform: uppercase;">
				Admin · {modalMode === 'create' ? 'Akun Baru' : 'Edit Akun'}
			</p>
			<h2 class="section-title" style="margin-bottom: 1.25rem;">
				{modalMode === 'create' ? 'Tambah Akun' : 'Edit Akun'}
			</h2>

			<form class="form-grid" onsubmit={handleSubmit}>
				<label class="field">
					<span>Nama Lengkap</span>
					<input type="text" bind:value={formName} placeholder="Nama pengguna" required />
				</label>

				<label class="field">
					<span>Email</span>
					<input type="email" bind:value={formEmail} placeholder="email@contoh.com" required />
				</label>

				<label class="field">
					<span>{modalMode === 'create' ? 'Password' : 'Password Baru'}</span>
					<input
						type="password"
						bind:value={formPassword}
						placeholder={modalMode === 'edit' ? 'Kosongkan jika tidak diubah' : 'Minimal 6 karakter'}
						required={modalMode === 'create'}
					/>
				</label>

				<label class="field">
					<span>Role</span>
					<select bind:value={formRole}>
						<option value="user">User</option>
						<option value="admin">Admin</option>
					</select>
				</label>

				{#if formError}
					<p class="error">{formError}</p>
				{/if}

				<div class="button-row" style="justify-content: flex-end; margin-top: 0.5rem;">
					<button
						class="button-secondary"
						type="button"
						onclick={closeModal}
						disabled={formSaving}
					>Batal</button>
					<button class="button-primary" type="submit" disabled={formSaving}>
						{formSaving ? 'Menyimpan…' : modalMode === 'create' ? 'Buat Akun' : 'Simpan'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Delete Confirm Modal -->
{#if deleteConfirmOpen && deleteTarget}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
	<div class="modal-backdrop" onclick={() => (deleteConfirmOpen = false)} role="presentation">
		<div class="modal-card modal-sm" onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1">
			<p class="muted mono" style="margin-bottom: 0.35rem; font-size: 0.6rem; letter-spacing: 0.18em; text-transform: uppercase;">Konfirmasi Hapus</p>
			<h2 class="section-title" style="margin-bottom: 0.75rem;">Hapus Akun?</h2>
			<p style="font-size: 0.9rem; color: var(--ink-soft); margin-bottom: 1.25rem; line-height: 1.5;">
				Akun <strong>{deleteTarget.name}</strong> ({deleteTarget.email}) beserta semua datanya
				(transaksi, kategori, budget) akan dihapus permanen. Tindakan ini tidak bisa dibatalkan.
			</p>
			<div class="button-row" style="justify-content: flex-end;">
				<button
					class="button-secondary"
					type="button"
					onclick={() => (deleteConfirmOpen = false)}
					disabled={deleting}
				>Batal</button>
				<button class="button-danger" type="button" onclick={handleDelete} disabled={deleting}>
					{deleting ? 'Menghapus…' : 'Ya, Hapus'}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.page {
		padding: 1.5rem 1rem 4rem;
		max-width: 900px;
		margin: 0 auto;
	}

	.page-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 1.75rem;
		flex-wrap: wrap;
	}

	.page-eyebrow {
		margin: 0 0 0.25rem;
		font-size: 0.6rem;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		color: var(--ink-faint);
	}

	.page-title {
		margin: 0 0 0.35rem;
		font-family: var(--font-display);
		font-size: clamp(1.75rem, 5vw, 2.5rem);
		font-weight: 400;
		line-height: 1;
		letter-spacing: -0.02em;
	}

	.page-desc {
		margin: 0;
		font-size: 0.88rem;
		color: var(--ink-soft);
		font-style: italic;
		font-family: var(--font-display);
	}

	.add-btn {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		white-space: nowrap;
		flex-shrink: 0;
	}

	.alert-error {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1rem;
		background: color-mix(in srgb, var(--oxblood) 10%, var(--paper));
		border: 1px solid var(--oxblood);
		color: var(--oxblood);
		font-size: 0.88rem;
		margin-bottom: 1.25rem;
	}

	.loading-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1rem;
		padding: 3rem;
		color: var(--ink-soft);
		font-style: italic;
		font-family: var(--font-display);
	}

	.loading-spinner {
		width: 2rem;
		height: 2rem;
		border: 2px solid var(--ink-faint);
		border-top-color: var(--ink);
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1rem;
		padding: 3rem 1rem;
		text-align: center;
		color: var(--ink-soft);
		border: 1.5px dashed var(--ink-faint);
	}

	.empty-state p {
		margin: 0;
		font-style: italic;
		font-family: var(--font-display);
	}

	/* Accounts grid */
	.accounts-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
		gap: 1rem;
	}

	.account-card {
		border: 1.5px solid var(--ink);
		padding: 1.25rem;
		background: var(--paper);
		display: flex;
		flex-direction: column;
		gap: 0.85rem;
		transition: box-shadow 0.15s, transform 0.15s;
		position: relative;
	}

	.account-card:hover {
		box-shadow: 4px 4px 0 var(--ink);
		transform: translate(-2px, -2px);
	}

	.account-card.is-admin {
		border-color: var(--oxblood);
	}

	.card-top {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.account-avatar {
		width: 2.75rem;
		height: 2.75rem;
		border-radius: 50%;
		background: var(--ink);
		color: var(--paper);
		display: grid;
		place-items: center;
		font-family: var(--font-display);
		font-style: italic;
		font-size: 1.1rem;
		letter-spacing: -0.02em;
		border: 1.5px solid var(--ink);
		flex-shrink: 0;
	}

	.account-avatar.avatar-admin {
		background: var(--oxblood);
		border-color: var(--oxblood);
	}

	.account-badge {
		font-family: var(--font-mono);
		font-size: 0.6rem;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		padding: 0.2rem 0.5rem;
		border: 1px solid var(--ink-faint);
		color: var(--ink-soft);
	}

	.account-badge.badge-admin {
		border-color: var(--oxblood);
		color: var(--oxblood);
	}

	.account-info {
		flex: 1;
	}

	.account-name {
		margin: 0 0 0.2rem;
		font-family: var(--font-display);
		font-size: 1.1rem;
		font-weight: 400;
		color: var(--ink);
		line-height: 1.2;
	}

	.account-email {
		margin: 0;
		font-family: var(--font-mono);
		font-size: 0.72rem;
		color: var(--ink-soft);
		letter-spacing: 0.02em;
		word-break: break-all;
	}

	.card-actions {
		display: flex;
		gap: 0.5rem;
		padding-top: 0.5rem;
		border-top: 1px solid var(--ink-faint);
	}

	.action-btn {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		padding: 0.45rem 0.75rem;
		font-size: 0.78rem;
		font-family: var(--font-mono);
		letter-spacing: 0.05em;
		border: 1px solid var(--ink-faint);
		background: transparent;
		color: var(--ink-soft);
		cursor: pointer;
		transition: all 0.15s;
	}

	.action-btn:hover {
		border-color: var(--ink);
		color: var(--ink);
		background: var(--ink);
		color: var(--paper);
	}

	.delete-btn:hover {
		border-color: var(--oxblood);
		background: var(--oxblood);
		color: var(--paper);
	}

	/* Modal */
	.modal-sm {
		width: min(100%, 420px);
	}

	@media (min-width: 640px) {
		.page {
			padding: 2rem 1.5rem 4rem;
		}
	}
</style>
