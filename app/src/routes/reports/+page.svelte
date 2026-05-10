<script lang="ts">
	import { onMount } from 'svelte';
	import { api, ApiError } from '$lib/api';
	import { formatPeriod, formatRupiah } from '$lib/format';
	import type { MonthlyReport } from '$lib/types';

	let loading = $state(true);
	let errorMessage = $state('');
	let report = $state<MonthlyReport | null>(null);
	let month = $state(new Date().toISOString().slice(0, 7));

	function budgetPercent(value: number): number {
		return Math.min(100, Math.max(0, value));
	}

	async function loadReport(): Promise<void> {
		loading = true;
		errorMessage = '';

		try {
			report = await api.get<MonthlyReport>(`/reports/monthly?month=${month}`);
		} catch (error) {
			if (error instanceof ApiError) {
				errorMessage = error.message;
			} else {
				errorMessage = 'Gagal memuat laporan bulanan.';
			}
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadReport();
	});
</script>

<section class="page">
	<header class="page-header">
		<div>
			<h1 class="page-title">Laporan</h1>
			<p class="page-subtitle">Ringkasan pengeluaran bulanan dan pemakaian budget.</p>
		</div>
	</header>

	<section class="section-card">
		<h2 class="section-title">Pilih Bulan Laporan</h2>
		<div class="form-row">
			<label class="field">
				<span>Bulan</span>
				<input type="month" bind:value={month} />
			</label>
			<div class="button-row" style="align-items:end;">
				<button class="button-primary" type="button" onclick={loadReport}>Tampilkan Laporan</button>
			</div>
		</div>
	</section>

	{#if errorMessage}
		<p class="error">{errorMessage}</p>
	{/if}

	{#if loading}
		<p class="muted">Memuat laporan...</p>
	{:else if report}
		<div class="card-grid wide">
			<article class="card">
				<p class="card-title">Total Pemasukan</p>
				<p class="card-value">{formatRupiah(report.total_income)}</p>
			</article>
			<article class="card">
				<p class="card-title">Total Pengeluaran</p>
				<p class="card-value">{formatRupiah(report.total_expense)}</p>
			</article>
			<article class="card">
				<p class="card-title">Selisih</p>
				<p class="card-value">{formatRupiah(report.net)}</p>
			</article>
		</div>

		<section class="section-card">
			<h2 class="section-title">Pengeluaran per Kategori</h2>
			{#if report.spending_by_category.length === 0}
				<p class="muted">Belum ada pengeluaran di bulan ini.</p>
			{:else}
				<div style="overflow:auto;">
					<table>
						<thead>
							<tr>
								<th>Kategori</th>
								<th>Total Pengeluaran</th>
							</tr>
						</thead>
						<tbody>
							{#each report.spending_by_category as item}
								<tr>
									<td>{item.category}</td>
									<td>{formatRupiah(item.total)}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</section>

		<section class="section-card">
			<h2 class="section-title">Ringkasan Pemakaian Budget</h2>
			{#if report.budget_usage.length === 0}
				<p class="muted">Belum ada aturan budget.</p>
			{:else}
				<div class="list">
					{#each report.budget_usage as usage}
						<article class="list-item">
							<div class="list-item-header">
								<div>
									<strong>{usage.category}</strong>
									<p class="muted">{formatPeriod(usage.period)}</p>
								</div>
								<p class="card-value">{formatRupiah(usage.used)} / {formatRupiah(usage.limit)}</p>
							</div>
							<div class="progress {usage.over_budget ? 'over' : ''}">
								<span style={`width: ${budgetPercent(usage.percentage)}%`}></span>
							</div>
							<p class="muted">Sisa: {formatRupiah(usage.remaining)}</p>
						</article>
					{/each}
				</div>
			{/if}
		</section>
	{/if}
</section>
