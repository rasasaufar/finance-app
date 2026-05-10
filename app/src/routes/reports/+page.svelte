<script lang="ts">
	import { onMount } from 'svelte';
	import { api, ApiError } from '$lib/api';
	import { formatPeriod, formatRupiah, categoryColor } from '$lib/format';
	import type { MonthlyReport } from '$lib/types';

	let loading = $state(true);
	let errorMessage = $state('');
	let report = $state<MonthlyReport | null>(null);
	let month = $state(new Date().toISOString().slice(0, 7));

	const todayLabel = new Intl.DateTimeFormat('id-ID', {
		weekday: 'long',
		day: 'numeric',
		month: 'long',
		year: 'numeric'
	}).format(new Date());

	function budgetPercent(value: number): number {
		return Math.min(100, Math.max(0, value));
	}

	function formatMonthLabel(value: string): string {
		const parsed = new Date(`${value}-01T00:00:00`);
		if (Number.isNaN(parsed.getTime())) return value;
		return new Intl.DateTimeFormat('id-ID', { month: 'long', year: 'numeric' }).format(parsed);
	}

	// Largest category share for proportional bars
	const maxCategorySpending = $derived(
		report && report.spending_by_category.length > 0
			? Math.max(...report.spending_by_category.map((c) => c.total))
			: 0
	);

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
		<div class="page-header-top">
			<span><span class="issue-mark">§</span> 06 · Laporan Bulanan</span>
			<span>{todayLabel}</span>
		</div>
		<div class="page-header-main">
			<div>
				<h1 class="page-title">Laporan <em>Bulanan</em></h1>
				<p class="page-subtitle">
					Ringkasan editorial: pemasukan, pengeluaran, dan kategori paling dominan.
				</p>
			</div>
		</div>
	</header>

	<section class="section-card report-picker">
		<div class="picker-row">
			<label class="field">
				<span>Pilih Bulan Laporan</span>
				<input type="month" bind:value={month} />
			</label>
			<button class="button-primary" type="button" onclick={loadReport}>
				Terbitkan Laporan
			</button>
		</div>
		{#if report}
			<p class="muted mono" style="margin-top: 0.75rem;">
				Menampilkan edisi · {formatMonthLabel(report.month)}
			</p>
		{/if}
	</section>

	{#if errorMessage}
		<p class="error">{errorMessage}</p>
	{/if}

	{#if loading}
		<p class="muted mono">Menyiapkan laporan…</p>
	{:else if report}
		<!-- Headline numbers -->
		<article class="report-hero">
			<div class="hero-piece" data-kind="income">
				<span class="mono tiny">Total Pemasukan</span>
				<span class="hero-num money-display">{formatRupiah(report.total_income)}</span>
			</div>
			<div class="hero-piece" data-kind="expense">
				<span class="mono tiny">Total Pengeluaran</span>
				<span class="hero-num money-display">{formatRupiah(report.total_expense)}</span>
			</div>
			<div class="hero-piece" data-kind="net">
				<span class="mono tiny">Selisih Bersih</span>
				<span class="hero-num money-display" data-positive={report.net >= 0}>
					{report.net >= 0 ? '+' : ''}{formatRupiah(report.net)}
				</span>
			</div>
		</article>

		<section class="section-card">
			<h2 class="section-title">Pengeluaran per Kategori</h2>
			<p class="section-lede">
				Peringkat kategori dari yang paling banyak menyerap dana.
			</p>
			{#if report.spending_by_category.length === 0}
				<p class="muted">Belum ada pengeluaran pada bulan ini.</p>
			{:else}
				<ol class="cat-chart">
					{#each report.spending_by_category as item, i}
						{@const share =
							maxCategorySpending > 0 ? (item.total / maxCategorySpending) * 100 : 0}
						<li class="cat-chart-row">
							<div class="chart-head">
								<span class="mono tiny">{String(i + 1).padStart(2, '0')}</span>
								<span class="cat-dot" style="background: {categoryColor(item.category)};"></span>
								<span class="chart-cat">{item.category}</span>
								<span class="leader"></span>
								<span class="chart-amount money-mono">{formatRupiah(item.total)}</span>
							</div>
							<div class="chart-bar">
								<span style="width: {share}%; background-color: {categoryColor(item.category)};"></span>
							</div>
						</li>
					{/each}
				</ol>
			{/if}
		</section>

		<section class="section-card">
			<h2 class="section-title">Pemakaian Budget</h2>
			<p class="section-lede">
				Status tiap aturan budget untuk edisi {formatMonthLabel(report.month)}.
			</p>
			{#if report.budget_usage.length === 0}
				<p class="muted">Belum ada aturan budget.</p>
			{:else}
				<div class="budget-report">
					{#each report.budget_usage as usage, i}
						<article class="budget-report-row">
							<div class="br-head">
								<span class="ledger-num">{String(i + 1).padStart(2, '0')}</span>
								<div class="br-title">
									<strong>{usage.category}</strong>
									<span class="mono tiny">{formatPeriod(usage.period)}</span>
								</div>
								<span class="leader"></span>
								<span class="money-mono br-amount">
									{formatRupiah(usage.used)} / {formatRupiah(usage.limit)}
								</span>
							</div>
							<div class="progress {usage.over_budget ? 'over' : ''}">
								<span style="width: {budgetPercent(usage.percentage)}%"></span>
							</div>
							<div class="progress-label">
								<span>{Math.round(usage.percentage)}% terpakai</span>
								<span>Sisa {formatRupiah(usage.remaining)}</span>
							</div>
						</article>
					{/each}
				</div>
			{/if}
		</section>
	{/if}
</section>

<style>
	.tiny {
		font-size: 0.6rem;
		letter-spacing: 0.15em;
		text-transform: uppercase;
	}

	.report-picker .picker-row {
		display: grid;
		grid-template-columns: 1fr auto;
		align-items: end;
		gap: 1rem;
	}

	.report-hero {
		display: grid;
		grid-template-columns: 1fr;
		gap: 0;
		border: 1.5px solid var(--ink);
		background: var(--rule);
	}

	.hero-piece {
		background: var(--paper);
		padding: 1.25rem 1.15rem;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.hero-num {
		font-size: clamp(1.75rem, 6vw, 2.5rem);
		line-height: 1;
		letter-spacing: -0.02em;
		color: var(--ink);
	}

	.hero-piece[data-kind='income'] .hero-num {
		color: var(--forest);
	}

	.hero-piece[data-kind='expense'] .hero-num {
		color: var(--oxblood);
	}

	.hero-piece[data-kind='net'] .hero-num[data-positive='true'] {
		color: var(--forest);
	}

	.hero-piece[data-kind='net'] .hero-num[data-positive='false'] {
		color: var(--oxblood);
	}

	/* Category chart */
	.cat-chart {
		list-style: none;
		padding: 0;
		margin: 0;
		display: grid;
		gap: 0.85rem;
		border-top: 1px solid var(--rule);
		padding-top: 0.75rem;
	}

	.cat-chart-row {
		display: grid;
		gap: 0.35rem;
	}

	.chart-head {
		display: flex;
		align-items: baseline;
		gap: 0.5rem;
	}

	.chart-cat {
		font-weight: 600;
		color: var(--ink);
	}

	.chart-amount {
		color: var(--ink);
	}

	.chart-bar {
		height: 10px;
		background: var(--paper-fold);
		border: 1px solid var(--rule);
		position: relative;
		overflow: hidden;
	}

	.chart-bar > span {
		display: block;
		height: 100%;
		background: var(--ink);
		background-image: repeating-linear-gradient(
			-45deg,
			transparent,
			transparent 4px,
			rgba(243, 236, 222, 0.25) 4px,
			rgba(243, 236, 222, 0.25) 8px
		);
		transition: width 0.6s cubic-bezier(0.22, 1, 0.36, 1);
	}

	/* Budget report */
	.budget-report {
		display: grid;
		gap: 1rem;
		border-top: 1px solid var(--rule);
		padding-top: 0.75rem;
	}

	.budget-report-row {
		padding: 0.5rem 0;
		border-bottom: 1px dashed var(--rule);
	}

	.budget-report-row:last-child {
		border-bottom: 0;
	}

	.br-head {
		display: flex;
		align-items: baseline;
		gap: 0.5rem;
		margin-bottom: 0.5rem;
	}

	.br-title {
		display: flex;
		flex-direction: column;
	}

	.br-title strong {
		font-weight: 600;
	}

	.br-amount {
		font-size: 0.8rem;
	}

	@media (min-width: 640px) {
		.report-hero {
			grid-template-columns: repeat(3, 1fr);
			gap: 1px;
		}
	}
</style>
