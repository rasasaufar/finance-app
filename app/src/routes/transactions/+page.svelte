<script lang="ts">
	import { onMount } from 'svelte';
	import { api, ApiError } from '$lib/api';
	import { formatPeriod, formatRupiah, categoryColor } from '$lib/format';
	import type {
		BudgetCheck,
		Category,
		Transaction,
		TransactionMutationResponse,
		TransactionType
	} from '$lib/types';
	import DatePicker from '$lib/DatePicker.svelte';

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

	let filterType = $state<'all' | TransactionType>('all');
	let filterCategory = $state('');


	const todayLabel = new Intl.DateTimeFormat('id-ID', {
		weekday: 'long',
		day: 'numeric',
		month: 'long',
		year: 'numeric'
	}).format(new Date());

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

	function formatDateLong(value: string): string {
		const d = new Date(`${value}T00:00:00`);
		if (Number.isNaN(d.getTime())) return value;
		return new Intl.DateTimeFormat('id-ID', {
			day: 'numeric',
			month: 'short',
			year: 'numeric'
		}).format(d);
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

	// Group transactions by date
	const filteredTransactions = $derived(
		transactions.filter((t) => {
			if (filterType !== 'all' && t.type !== filterType) return false;
			if (filterCategory && t.category !== filterCategory) return false;
			return true;
		})
	);

	const groupedTransactions = $derived.by(() => {
		const groups = new Map<string, Transaction[]>();
		for (const t of filteredTransactions) {
			if (!groups.has(t.date)) {
				groups.set(t.date, []);
			}
			groups.get(t.date)!.push(t);
		}
		return Array.from(groups.entries()).sort((a, b) => b[0].localeCompare(a[0]));
	});

	const totalIncome = $derived(
		transactions.filter((t) => t.type === 'income').reduce((s, t) => s + t.amount, 0)
	);
	const totalExpense = $derived(
		transactions.filter((t) => t.type === 'expense').reduce((s, t) => s + t.amount, 0)
	);

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

			successMessage = 'Transaksi berhasil dicatat di buku kas.';
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
		<div class="page-header-top">
			<span><span class="issue-mark">§</span> 02 · Lembar Transaksi</span>
			<span>{todayLabel}</span>
		</div>
		<div class="page-header-main">
			<div>
				<h1 class="page-title">Lembar <em>Transaksi</em></h1>
				<p class="page-subtitle">
					Setiap rupiah yang masuk dan keluar dicatat pada lembar ini.
				</p>
			</div>
		</div>
	</header>

	<!-- Totals at the top -->
	<div class="tx-totals">
		<div class="tx-total">
			<span class="mono tiny">Total Pemasukan</span>
			<span class="tx-total-num money-display" data-kind="income">
				+{formatRupiah(totalIncome)}
			</span>
		</div>
		<div class="tx-total">
			<span class="mono tiny">Total Pengeluaran</span>
			<span class="tx-total-num money-display" data-kind="expense">
				−{formatRupiah(totalExpense)}
			</span>
		</div>
		<div class="tx-total">
			<span class="mono tiny">Saldo Bersih</span>
			<span class="tx-total-num money-display">
				{formatRupiah(totalIncome - totalExpense)}
			</span>
		</div>
	</div>

	<section class="section-card">
		<h2 class="section-title">Catat Transaksi Baru</h2>
		<p class="section-lede">
			Isi kolom-kolom di bawah, lalu konfirmasi sebelum masuk ke buku kas.
		</p>

		{#if errorMessage}
			<p class="error">{errorMessage}</p>
		{/if}
		{#if successMessage}
			<p class="success">{successMessage}</p>
		{/if}
		{#if budgetWarnings.length > 0}
			<div class="notice">
				<strong>Peringatan Budget</strong>
				{#each budgetWarnings as warning}
					<p>
						{warning.category} ({formatPeriod(warning.period)}) melebihi limit. Sisa
						{formatRupiah(warning.remaining)}.
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
				<DatePicker bind:value={date} label="Tanggal" required />
			</div>

			<label class="field">
				<span>Catatan</span>
				<textarea
					bind:value={note}
					placeholder="Contoh: makan siang, isi bensin, gaji bulan ini"
				></textarea>
			</label>

			<div class="button-row">
				<button class="button-primary" type="submit" disabled={saving}>
					{saving ? 'Menyimpan…' : 'Catat di Buku Kas'}
				</button>
				<button class="button-secondary" type="button" onclick={loadData}>Muat Ulang</button>
			</div>
		</form>
	</section>

	<section class="section-card">
		<h2 class="section-title">Daftar Transaksi</h2>
		<p class="section-lede">
			Disusun menurut tanggal, dari yang paling baru.
		</p>

		<!-- Filter strip -->
		<div class="filter-strip">
			<div class="filter-group" role="group" aria-label="Filter tipe">
				<button
					class="filter-chip"
					class:is-active={filterType === 'all'}
					type="button"
					onclick={() => (filterType = 'all')}
				>
					Semua · {transactions.length}
				</button>
				<button
					class="filter-chip"
					class:is-active={filterType === 'income'}
					type="button"
					onclick={() => (filterType = 'income')}
				>
					Pemasukan · {transactions.filter((t) => t.type === 'income').length}
				</button>
				<button
					class="filter-chip"
					class:is-active={filterType === 'expense'}
					type="button"
					onclick={() => (filterType = 'expense')}
				>
					Pengeluaran · {transactions.filter((t) => t.type === 'expense').length}
				</button>
			</div>
			{#if categories.length > 0}
				<label class="filter-cat">
					<span class="mono tiny">Kategori</span>
					<select bind:value={filterCategory}>
						<option value="">Semua kategori</option>
						{#each categories as item}
							<option value={item.name}>{item.name}</option>
						{/each}
					</select>
				</label>
			{/if}
		</div>

		{#if loading}
			<p class="muted mono">Memuat daftar transaksi…</p>
		{:else if transactions.length === 0}
			<p class="muted">Belum ada transaksi. Silakan catat yang pertama.</p>
		{:else if filteredTransactions.length === 0}
			<p class="muted">Tidak ada transaksi yang cocok dengan filter.</p>
		{:else}
			<div class="tx-ledger">
				{#each groupedTransactions as [day, items]}
					<div class="tx-day">
						<div class="tx-day-head">
							<span class="day-date">{formatDateLong(day)}</span>
							<span class="leader"></span>
							<span class="mono tiny">{items.length} entri</span>
						</div>
						{#each items as transaction}
							<article class="tx-row" data-type={transaction.type}>
								<div class="tx-info">
									<div class="tx-main">
										<span class="cat-dot" style="background: {categoryColor(transaction.category)};"></span>
										<span class="tx-category">{transaction.category}</span>
										<span class="badge {transaction.type}">
											{transaction.type === 'income' ? 'Masuk' : 'Keluar'}
										</span>
									</div>
									{#if transaction.note}
										<p class="tx-note">{transaction.note}</p>
									{/if}
								</div>
								<div class="tx-amount">
									<p class="money-display tx-amount-num" data-type={transaction.type}>
										{transaction.type === 'income' ? '+' : '−'}{formatRupiah(transaction.amount)}
									</p>
									<button
										class="tx-delete"
										type="button"
										onclick={() => openDeleteConfirmModal(transaction)}
										aria-label="Hapus transaksi"
									>
										Hapus
									</button>
								</div>
							</article>
						{/each}
					</div>
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
			<p class="muted mono" style="margin-bottom: 0.35rem;">Konfirmasi · Transaksi Baru</p>
			<h3 id="confirm-add-title" class="section-title" style="margin-bottom: 1rem;">
				Catat Transaksi?
			</h3>
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
					<strong>{formatDateLong(pendingDraft.date)}</strong>
				</div>
				<div class="confirm-detail-row">
					<span>Catatan</span>
					<strong>{pendingDraft.note || '—'}</strong>
				</div>
			</div>
			<div class="button-row" style="justify-content: flex-end;">
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
					{saving ? 'Menyimpan…' : 'Ya, Catat'}
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
			<p class="muted mono" style="margin-bottom: 0.35rem;">Konfirmasi · Hapus Entri</p>
			<h3 id="confirm-delete-title" class="section-title" style="margin-bottom: 1rem;">
				Hapus Transaksi?
			</h3>
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
					<strong>{formatDateLong(pendingDelete.date)}</strong>
				</div>
				<div class="confirm-detail-row">
					<span>Catatan</span>
					<strong>{pendingDelete.note || '—'}</strong>
				</div>
			</div>
			<p class="muted" style="font-style: italic; font-family: var(--font-display); font-size: 0.95rem;">
				Entri yang dihapus tidak bisa dikembalikan.
			</p>
			<div class="button-row" style="justify-content: flex-end;">
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
					{saving ? 'Menghapus…' : 'Ya, Hapus'}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.tiny {
		font-size: 0.6rem;
		letter-spacing: 0.15em;
		text-transform: uppercase;
	}

	/* Totals strip */
	.tx-totals {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 0;
		border: 1.5px solid var(--ink);
		background: var(--rule);
	}

	.tx-total {
		background: var(--paper);
		padding: 0.85rem 0.85rem;
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
		min-width: 0;
	}

	.tx-total-num {
		font-size: clamp(1rem, 3.5vw, 1.35rem);
		color: var(--ink);
		line-height: 1;
		letter-spacing: -0.02em;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.tx-total-num[data-kind='income'] {
		color: var(--forest);
	}

	.tx-total-num[data-kind='expense'] {
		color: var(--oxblood);
	}

	/* Ledger list */
	.tx-ledger {
		display: grid;
		gap: 1.5rem;
	}

	.tx-day {
		display: grid;
		gap: 0;
	}

	.tx-day-head {
		display: flex;
		align-items: baseline;
		gap: 0.5rem;
		padding-bottom: 0.5rem;
		border-bottom: 1.5px solid var(--ink);
		margin-bottom: 0.5rem;
	}

	.day-date {
		font-family: var(--font-display);
		font-size: 1.15rem;
		color: var(--ink);
		line-height: 1;
	}

	.tx-row {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
		padding: 0.75rem 0;
		border-bottom: 1px dashed var(--rule);
		position: relative;
	}

	.tx-row:last-child {
		border-bottom: 0;
	}

	.tx-info {
		flex: 1;
		min-width: 0;
	}

	.tx-main {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		flex-wrap: wrap;
	}

	.tx-category {
		font-weight: 600;
		font-size: 1rem;
		color: var(--ink);
	}

	.tx-note {
		margin: 0.2rem 0 0;
		font-size: 0.85rem;
		color: var(--ink-soft);
		font-style: italic;
		font-family: var(--font-display);
	}

	.tx-amount {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 0.35rem;
		flex-shrink: 0;
	}

	.tx-amount-num {
		font-size: 1.1rem;
		color: var(--ink);
		line-height: 1;
		letter-spacing: -0.02em;
	}

	.tx-amount-num[data-type='income'] {
		color: var(--forest);
	}

	.tx-amount-num[data-type='expense'] {
		color: var(--oxblood);
	}

	.tx-delete {
		background: transparent;
		border: 0;
		padding: 0;
		color: var(--ink-faint);
		font-family: var(--font-display);
		font-style: italic;
		font-size: 0.85rem;
		cursor: pointer;
		border-bottom: 1px dotted var(--rule);
		line-height: 1.3;
	}

	.tx-delete:hover {
		color: var(--oxblood);
		border-bottom-color: var(--oxblood);
	}

	/* Filter strip */
	.filter-strip {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 1rem;
		padding: 0.75rem 0;
		margin-bottom: 0.5rem;
		border-bottom: 1.5px solid var(--ink);
		flex-wrap: wrap;
	}

	.filter-group {
		display: flex;
		gap: 0;
		flex-wrap: wrap;
		border: 1px solid var(--ink);
	}

	.filter-chip {
		background: var(--paper);
		border: 0;
		border-right: 1px solid var(--ink);
		padding: 0.5rem 0.85rem;
		font-family: var(--font-mono);
		font-size: 0.65rem;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: var(--ink-soft);
		cursor: pointer;
		transition: all 0.15s;
	}

	.filter-chip:last-child {
		border-right: 0;
	}

	.filter-chip:hover {
		background: var(--paper-deep);
		color: var(--ink);
	}

	.filter-chip.is-active {
		background: var(--ink);
		color: var(--paper);
	}

	.filter-cat {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.8rem;
	}

	.filter-cat span {
		color: var(--ink-soft);
		flex-shrink: 0;
	}

	.filter-cat select {
		min-height: 2rem;
		padding: 0.3rem 1.5rem 0.3rem 0.1rem;
		font-size: 0.85rem;
		background-position: right 0 center;
	}

	@media (min-width: 640px) {
		.tx-totals {
			grid-template-columns: repeat(3, 1fr);
		}

		.tx-total {
			padding: 1rem 1.15rem;
		}

		.tx-total-num {
			font-size: 1.5rem;
		}
	}
</style>
