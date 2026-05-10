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

	let loading = $state(true);
	let saving = $state(false);
	let errorMessage = $state('');
	let successMessage = $state('');
	let transactions = $state<Transaction[]>([]);
	let categories = $state<Category[]>([]);
	let budgetWarnings = $state<BudgetCheck[]>([]);

	let type = $state<TransactionType>('expense');
	let category = $state('');
	let amount = $state('');
	let date = $state(new Date().toISOString().slice(0, 10));
	let note = $state('');

	function clearMessages(): void {
		errorMessage = '';
		successMessage = '';
	}

	async function loadData(): Promise<void> {
		loading = true;
		clearMessages();

		try {
			const [transactionsResponse, categoriesResponse] = await Promise.all([
				api.get<Transaction[]>('/transactions'),
				api.get<Category[]>('/categories')
			]);

			transactions = transactionsResponse;
			categories = categoriesResponse;

			if (!category && categoriesResponse.length > 0) {
				category = categoriesResponse[0].name;
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

		saving = true;
		try {
			const response = await api.post<TransactionMutationResponse>('/transactions', {
				type,
				category,
				amount: numericAmount,
				date,
				note
			});

			transactions = [response.transaction, ...transactions];
			successMessage = 'Transaksi berhasil disimpan.';
			amount = '';
			note = '';

			if (response.warning) {
				budgetWarnings = response.budget_checks.filter((item) => item.over_budget);
			}
		} catch (error) {
			if (error instanceof ApiError) {
				errorMessage = error.message;
			} else {
				errorMessage = 'Gagal menambah transaksi.';
			}
		} finally {
			saving = false;
		}
	}

	async function handleDelete(id: number): Promise<void> {
		clearMessages();

		try {
			await api.delete<void>(`/transactions/${id}`);
			transactions = transactions.filter((item) => item.id !== id);
			successMessage = 'Transaksi berhasil dihapus.';
		} catch (error) {
			if (error instanceof ApiError) {
				errorMessage = error.message;
			} else {
				errorMessage = 'Gagal menghapus transaksi.';
			}
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
								onclick={() => handleDelete(transaction.id)}
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
