<script lang="ts">
	import { onMount } from 'svelte';
	import { api, ApiError } from '$lib/api';
	import { formatPeriod, formatRupiah } from '$lib/format';
	import type {
		BudgetCheck,
		Category,
		Transaction,
		TransactionMutationResponse,
		TransactionType
	} from '$lib/types';

	interface TransactionDraft {
		type: TransactionType;
		category: string;
		amount: number;
		date: string;
		note: string;
	}

	let loading = $state(true);
	let saving = $state(false);
	let errorMessage = $state('');
	let successMessage = $state('');
	let transactions = $state<Transaction[]>([]);
	let categories = $state<Category[]>([]);
	let budgetWarnings = $state<BudgetCheck[]>([]);
	let showAddConfirmModal = $state(false);
	let showDeleteConfirmModal = $state(false);
	let pendingDraft = $state<TransactionDraft | null>(null);
	let pendingDelete = $state<Transaction | null>(null);

	let type = $state<TransactionType>('expense');
	let category = $state('');
	let amount = $state('');
	let date = $state(new Date().toISOString().slice(0, 10));
	let note = $state('');

	function toTransactionArray(input: unknown): Transaction[] {
		return Array.isArray(input) ? (input as Transaction[]) : [];
	}

	function toCategoryArray(input: unknown): Category[] {
		return Array.isArray(input) ? (input as Category[]) : [];
	}

	function clearMessages(): void {
		errorMessage = '';
		successMessage = '';
	}

	function formatTransactionType(typeValue: TransactionType): string {
		return typeValue === 'income' ? 'Pemasukan' : 'Pengeluaran';
	}

	function handleAddBackdropClick(event: MouseEvent): void {
		if (event.target === event.currentTarget) {
			closeAddConfirmModal();
		}
	}

	function handleDeleteBackdropClick(event: MouseEvent): void {
		if (event.target === event.currentTarget) {
			closeDeleteConfirmModal();
		}
	}

	function handleBackdropKeydown(event: KeyboardEvent, kind: 'add' | 'delete'): void {
		if (event.key !== 'Escape') {
			return;
		}

		if (kind === 'add') {
			closeAddConfirmModal();
			return;
		}

		closeDeleteConfirmModal();
	}

	function closeAddConfirmModal(): void {
		if (saving) {
			return;
		}
		showAddConfirmModal = false;
		pendingDraft = null;
	}

	function openDeleteConfirmModal(transaction: Transaction): void {
		pendingDelete = transaction;
		showDeleteConfirmModal = true;
	}

	function closeDeleteConfirmModal(): void {
		if (saving) {
			return;
		}
		showDeleteConfirmModal = false;
		pendingDelete = null;
	}

	async function loadData(): Promise<void> {
		loading = true;
		clearMessages();

		try {
			const [transactionsResponse, categoriesResponse] = await Promise.all([
				api.get<Transaction[]>('/transactions'),
				api.get<Category[]>('/categories')
			]);

			transactions = toTransactionArray(transactionsResponse);
			categories = toCategoryArray(categoriesResponse);

			if (!category && categories.length > 0) {
				category = categories[0].name;
			}
		} catch (error) {
			if (error instanceof ApiError) {
				errorMessage = error.message;
			} else {
				errorMessage = 'Gagal memuat transaksi.';
			}
		} finally {
			loading = false;
		}
	}

	async function handleSubmit(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		clearMessages();
		budgetWarnings = [];

		const numericAmount = Number(amount);
		if (!Number.isFinite(numericAmount) || numericAmount <= 0) {
			errorMessage = 'Nominal transaksi harus lebih dari 0.';
			return;
		}
		if (!category) {
			errorMessage = 'Pilih kategori terlebih dahulu.';
			return;
		}

		pendingDraft = {
			type,
			category,
			amount: numericAmount,
			date,
			note
		};
		showAddConfirmModal = true;
	}

	async function confirmAddTransaction(): Promise<void> {
		if (!pendingDraft) {
			return;
		}

		saving = true;
		try {
			const response = await api.post<TransactionMutationResponse>('/transactions', pendingDraft);

			const createdTransaction = response?.transaction;
			if (createdTransaction) {
				if (!Array.isArray(transactions)) {
					transactions = [];
				}
				transactions.unshift(createdTransaction);
			}

			successMessage = 'Transaksi berhasil disimpan.';
			amount = '';
			note = '';
			showAddConfirmModal = false;
			pendingDraft = null;

			const budgetChecks = Array.isArray(response?.budget_checks) ? response.budget_checks : [];
			if (response?.warning) {
				budgetWarnings = budgetChecks.filter((item) => item.over_budget);
			}
		} catch (error) {
			if (error instanceof ApiError) {
				errorMessage = error.message;
			} else if (error instanceof Error) {
				errorMessage = `Gagal menambah transaksi. (${error.message})`;
			} else {
				errorMessage = 'Gagal menambah transaksi.';
			}
		} finally {
			saving = false;
		}
	}

	async function confirmDeleteTransaction(): Promise<void> {
		if (!pendingDelete) {
			return;
		}

		clearMessages();
		saving = true;

		try {
			await api.delete<void>(`/transactions/${pendingDelete.id}`);
			const index = transactions.findIndex((item) => item.id === pendingDelete?.id);
			if (index >= 0) {
				transactions.splice(index, 1);
			}
			successMessage = 'Transaksi berhasil dihapus.';
			showDeleteConfirmModal = false;
			pendingDelete = null;
		} catch (error) {
			if (error instanceof ApiError) {
				errorMessage = error.message;
			} else {
				errorMessage = 'Gagal menghapus transaksi.';
			}
		} finally {
			saving = false;
		}
	}

	onMount(() => {
		loadData();
	});
</script>

<section class="page">
	<header class="page-header">
		<div>
			<h1 class="page-title">Transaksi</h1>
			<p class="page-subtitle">Catat pemasukan dan pengeluaran harian Anda.</p>
		</div>
	</header>

	<section class="section-card">
		<h2 class="section-title">Tambah Transaksi</h2>

		{#if errorMessage}
			<p class="error">{errorMessage}</p>
		{/if}
		{#if successMessage}
			<p class="success">{successMessage}</p>
		{/if}
		{#if budgetWarnings.length > 0}
			<div class="notice">
				<strong>Peringatan budget:</strong>
				{#each budgetWarnings as warning}
					<p>
						{warning.category} ({formatPeriod(warning.period)}) melebihi limit. Sisa:
						{formatRupiah(warning.remaining)}
					</p>
				{/each}
			</div>
		{/if}

		<form class="form-grid" onsubmit={handleSubmit}>
			<div class="form-row">
				<label class="field">
					<span>Tipe</span>
					<select bind:value={type}>
						<option value="income">Pemasukan</option>
						<option value="expense">Pengeluaran</option>
					</select>
				</label>

				<label class="field">
					<span>Kategori</span>
					<select bind:value={category} required>
						<option value="" disabled>Pilih kategori</option>
						{#each categories as item}
							<option value={item.name}>{item.name}</option>
						{/each}
					</select>
				</label>
			</div>

			<div class="form-row">
				<label class="field">
					<span>Nominal</span>
					<input type="number" bind:value={amount} min="1" placeholder="Contoh: 50000" required />
				</label>
				<label class="field">
					<span>Tanggal</span>
					<input type="date" bind:value={date} required />
				</label>
			</div>

			<label class="field">
				<span>Catatan</span>
				<textarea bind:value={note} placeholder="Contoh: makan siang, isi bensin, gaji bulan ini"
				></textarea>
			</label>

			<div class="button-row">
				<button class="button-primary" type="submit" disabled={saving}>
					{saving ? 'Menyimpan...' : 'Simpan Transaksi'}
				</button>
				<button class="button-secondary" type="button" onclick={loadData}>Muat Ulang</button>
			</div>
		</form>
	</section>

	<section class="section-card">
		<h2 class="section-title">Daftar Transaksi</h2>
		{#if loading}
			<p class="muted">Memuat transaksi...</p>
		{:else if transactions.length === 0}
			<p class="muted">Belum ada transaksi.</p>
		{:else}
			<div class="list">
				{#each transactions as transaction}
					<article class="list-item">
						<div class="list-item-header">
							<div>
								<strong>{transaction.category}</strong>
								<p class="muted">
									{transaction.date}
									{transaction.note ? `• ${transaction.note}` : ''}
								</p>
							</div>
							<div style="text-align:right;">
								<span class="badge {transaction.type}">
									{transaction.type === 'income' ? 'Pemasukan' : 'Pengeluaran'}
								</span>
								<p class="card-value">
									{transaction.type === 'income' ? '+' : '-'}{formatRupiah(transaction.amount)}
								</p>
							</div>
						</div>
						<div class="button-row">
							<button
								class="button-danger"
								type="button"
								onclick={() => openDeleteConfirmModal(transaction)}
							>
								Hapus
							</button>
						</div>
					</article>
				{/each}
			</div>
		{/if}
	</section>
</section>

{#if showAddConfirmModal && pendingDraft}
	<div
		class="modal-backdrop"
		role="button"
		tabindex="0"
		aria-label="Tutup popup konfirmasi tambah transaksi"
		onclick={handleAddBackdropClick}
		onkeydown={(event) => handleBackdropKeydown(event, 'add')}
	>
		<div
			class="modal-card"
			role="dialog"
			aria-modal="true"
			aria-labelledby="confirm-add-title"
			tabindex="-1"
		>
			<h3 id="confirm-add-title" class="section-title">Konfirmasi Tambah Transaksi</h3>
			<div class="confirm-detail-list">
				<div class="confirm-detail-row">
					<span>Tipe</span>
					<strong>{formatTransactionType(pendingDraft.type)}</strong>
				</div>
				<div class="confirm-detail-row">
					<span>Kategori</span>
					<strong>{pendingDraft.category}</strong>
				</div>
				<div class="confirm-detail-row">
					<span>Nominal</span>
					<strong>{formatRupiah(pendingDraft.amount)}</strong>
				</div>
				<div class="confirm-detail-row">
					<span>Tanggal</span>
					<strong>{pendingDraft.date}</strong>
				</div>
				<div class="confirm-detail-row">
					<span>Catatan</span>
					<strong>{pendingDraft.note || '-'}</strong>
				</div>
			</div>
			<div class="button-row">
				<button
					class="button-secondary"
					type="button"
					onclick={closeAddConfirmModal}
					disabled={saving}
				>
					Batal
				</button>
				<button
					class="button-primary"
					type="button"
					onclick={confirmAddTransaction}
					disabled={saving}
				>
					{saving ? 'Menyimpan...' : 'Ya, Simpan Transaksi'}
				</button>
			</div>
		</div>
	</div>
{/if}

{#if showDeleteConfirmModal && pendingDelete}
	<div
		class="modal-backdrop"
		role="button"
		tabindex="0"
		aria-label="Tutup popup konfirmasi hapus transaksi"
		onclick={handleDeleteBackdropClick}
		onkeydown={(event) => handleBackdropKeydown(event, 'delete')}
	>
		<div
			class="modal-card"
			role="dialog"
			aria-modal="true"
			aria-labelledby="confirm-delete-title"
			tabindex="-1"
		>
			<h3 id="confirm-delete-title" class="section-title">Konfirmasi Hapus Transaksi</h3>
			<div class="confirm-detail-list">
				<div class="confirm-detail-row">
					<span>Kategori</span>
					<strong>{pendingDelete.category}</strong>
				</div>
				<div class="confirm-detail-row">
					<span>Tipe</span>
					<strong>{formatTransactionType(pendingDelete.type)}</strong>
				</div>
				<div class="confirm-detail-row">
					<span>Nominal</span>
					<strong>{formatRupiah(pendingDelete.amount)}</strong>
				</div>
				<div class="confirm-detail-row">
					<span>Tanggal</span>
					<strong>{pendingDelete.date}</strong>
				</div>
				<div class="confirm-detail-row">
					<span>Catatan</span>
					<strong>{pendingDelete.note || '-'}</strong>
				</div>
			</div>
			<p class="muted">Data yang dihapus tidak bisa dikembalikan.</p>
			<div class="button-row">
				<button
					class="button-secondary"
					type="button"
					onclick={closeDeleteConfirmModal}
					disabled={saving}
				>
					Batal
				</button>
				<button
					class="button-danger"
					type="button"
					onclick={confirmDeleteTransaction}
					disabled={saving}
				>
					{saving ? 'Menghapus...' : 'Ya, Hapus Transaksi'}
				</button>
			</div>
		</div>
	</div>
{/if}
