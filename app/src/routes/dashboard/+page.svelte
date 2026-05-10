<script lang="ts">
	import { onMount } from 'svelte';
	import { api, ApiError } from '$lib/api';
	import { formatPeriod, formatRupiah } from '$lib/format';
	import type { BudgetCheck, DashboardSummary } from '$lib/types';

	let loading = $state(true);
	let errorMessage = $state('');
	let summary = $state<DashboardSummary | null>(null);

	function budgetPercent(value: number): number {
		return Math.min(100, Math.max(0, value));
	}

	function findBudget(category: string, period: string): BudgetCheck | null {
		if (!summary) {
			return null;
		}

		if (category === 'Makan' && period === 'daily' && summary.makan_today) {
			return summary.makan_today;
		}
		if (category === 'Bensin' && period === 'monthly' && summary.bensin_month) {
			return summary.bensin_month;
		}

		return (
			summary.budget_usage.find((item) => item.category === category && item.period === period) ??
			null
		);
	}

	const makanBudget = $derived(findBudget('Makan', 'daily'));
	const bensinBudget = $derived(findBudget('Bensin', 'monthly'));

	async function loadSummary(): Promise<void> {
		loading = true;
		errorMessage = '';

		try {
			summary = await api.get<DashboardSummary>('/dashboard/summary');
		} catch (error) {
			if (error instanceof ApiError) {
				errorMessage = error.message;
			} else {
				errorMessage = 'Gagal memuat data dashboard.';
			}
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadSummary();
	});
</script>

<section class="page">
	<header class="page-header">
		<div>
			<h1 class="page-title">Dashboard</h1>
			<p class="page-subtitle">Ringkasan keuangan hari ini dan bulan ini.</p>
		</div>
		<button class="button-secondary" type="button" onclick={loadSummary}>Segarkan</button>
	</header>

	{#if errorMessage}
		<p class="error">{errorMessage}</p>
	{/if}

	{#if loading}
		<p class="muted">Memuat ringkasan...</p>
	{:else if summary}
		<div class="card-grid wide">
			<article class="card">
				<p class="card-title">Saldo Saat Ini</p>
				<p class="card-value">{formatRupiah(summary.current_balance)}</p>
			</article>
			<article class="card">
				<p class="card-title">Pengeluaran Hari Ini</p>
				<p class="card-value">{formatRupiah(summary.today_expense)}</p>
			</article>
			<article class="card">
				<p class="card-title">Pengeluaran Bulan Ini</p>
				<p class="card-value">{formatRupiah(summary.month_expense)}</p>
			</article>
			<article class="card">
				<p class="card-title">Sisa Budget</p>
				<p class="card-value">{formatRupiah(summary.remaining_budget)}</p>
			</article>
		</div>

		<section class="section-card">
			<h2 class="section-title">Limit Utama</h2>
			<div class="card-grid">
				{#if makanBudget}
					<div class="list-item">
						<p class="card-title">Makan Hari Ini (Limit Rp 60.000)</p>
						<p class="card-value">
							{formatRupiah(makanBudget.used)} / {formatRupiah(makanBudget.limit)}
						</p>
						<div class="progress {makanBudget.over_budget ? 'over' : ''}">
							<span style={`width: ${budgetPercent(makanBudget.percentage)}%`}></span>
						</div>
						<p class="muted">Sisa: {formatRupiah(makanBudget.remaining)}</p>
					</div>
				{/if}

				{#if bensinBudget}
					<div class="list-item">
						<p class="card-title">Bensin Bulan Ini (Limit Rp 240.000)</p>
						<p class="card-value">
							{formatRupiah(bensinBudget.used)} / {formatRupiah(bensinBudget.limit)}
						</p>
						<div class="progress {bensinBudget.over_budget ? 'over' : ''}">
							<span style={`width: ${budgetPercent(bensinBudget.percentage)}%`}></span>
						</div>
						<p class="muted">Sisa: {formatRupiah(bensinBudget.remaining)}</p>
					</div>
				{/if}
			</div>
		</section>

		<section class="section-card">
			<h2 class="section-title">Pemakaian Budget Lainnya</h2>
			<div class="list">
				{#each summary.budget_usage as usage}
					<div class="list-item">
						<div class="list-item-header">
							<strong>{usage.category}</strong>
							<span class="muted">{formatPeriod(usage.period)}</span>
						</div>
						<p class="card-value">{formatRupiah(usage.used)} / {formatRupiah(usage.limit)}</p>
						<div class="progress {usage.over_budget ? 'over' : ''}">
							<span style={`width: ${budgetPercent(usage.percentage)}%`}></span>
						</div>
					</div>
				{/each}
			</div>
		</section>
	{/if}
</section>
