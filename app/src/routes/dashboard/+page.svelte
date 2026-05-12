<script lang="ts">
	import { onMount } from 'svelte';
	import { api, ApiError } from '$lib/api';
	import { formatPeriod, formatRupiah } from '$lib/format';
	import type { BudgetCheck, DashboardSummary, Transaction } from '$lib/types';

	let loading = $state(true);
	let errorMessage = $state('');
	let summary = $state<DashboardSummary | null>(null);
	let recentTransactions = $state<Transaction[]>([]);

	const today = new Date();
	const todayLabel = new Intl.DateTimeFormat('id-ID', {
		weekday: 'long',
		day: 'numeric',
		month: 'long',
		year: 'numeric'
	}).format(today);

	const monthLabel = new Intl.DateTimeFormat('id-ID', {
		month: 'long',
		year: 'numeric'
	}).format(today);

	const dayOfMonth = today.getDate();
	const lastDayOfMonth = new Date(today.getFullYear(), today.getMonth() + 1, 0).getDate();
	const monthProgress = Math.round((dayOfMonth / lastDayOfMonth) * 100);

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
	const makanMonthlyBudget = $derived(summary?.makan_month ?? null);
	const bensinMonthlyBudget = $derived(summary?.bensin_month ?? null);

	function splitRupiah(value: number): { mark: string; number: string; negative: boolean } {
		const negative = value < 0;
		const full = formatRupiah(Math.abs(value));
		const match = full.match(/^(Rp)\s*(.+)$/);
		if (!match) {
			return { mark: '', number: full, negative };
		}
		return { mark: match[1], number: match[2], negative };
	}

	// Build a 7-day spending sparkline from recent transactions
	const sparklineData = $derived.by(() => {
		const days: { date: Date; total: number; label: string }[] = [];
		for (let i = 6; i >= 0; i--) {
			const d = new Date(today);
			d.setHours(0, 0, 0, 0);
			d.setDate(d.getDate() - i);
			// Use local date parts to avoid UTC offset shifting the date
			const year = d.getFullYear();
			const month = String(d.getMonth() + 1).padStart(2, '0');
			const day = String(d.getDate()).padStart(2, '0');
			const iso = `${year}-${month}-${day}`;
			const total = recentTransactions
				.filter((t) => t.type === 'expense' && t.date === iso)
				.reduce((sum, t) => sum + t.amount, 0);
			days.push({
				date: d,
				total,
				label: new Intl.DateTimeFormat('id-ID', { weekday: 'narrow' }).format(d)
			});
		}
		return days;
	});

	const sparklineMax = $derived(
		Math.max(1, ...sparklineData.map((d) => d.total))
	);

	const sparklinePath = $derived.by(() => {
		if (sparklineData.length === 0) return '';
		const w = 100;
		const h = 30;
		const step = w / (sparklineData.length - 1);
		const points = sparklineData.map((d, i) => {
			const x = i * step;
			const y = h - (d.total / sparklineMax) * (h - 4) - 2;
			return [x, y] as const;
		});
		let path = `M ${points[0][0]},${points[0][1]}`;
		for (let i = 1; i < points.length; i++) {
			const [px, py] = points[i - 1];
			const [cx, cy] = points[i];
			const midX = (px + cx) / 2;
			path += ` Q ${midX},${py} ${midX},${(py + cy) / 2}`;
			path += ` Q ${midX},${cy} ${cx},${cy}`;
		}
		return path;
	});

	const sparklineAreaPath = $derived(
		sparklinePath ? `${sparklinePath} L 100,30 L 0,30 Z` : ''
	);

	const weekTotal = $derived(
		sparklineData.reduce((s, d) => s + d.total, 0)
	);

	async function loadSummary(): Promise<void> {
		loading = true;
		errorMessage = '';

		try {
			const [sum, txs] = await Promise.all([
				api.get<DashboardSummary>('/dashboard/summary'),
				api.get<Transaction[]>('/transactions').catch(() => [])
			]);
			summary = sum;
			recentTransactions = Array.isArray(txs) ? txs : [];
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
		<div class="page-header-top">
			<span><span class="issue-mark">§</span> 01 · Ringkasan Harian</span>
			<span>{todayLabel}</span>
		</div>
		<div class="page-header-main">
			<div>
				<h1 class="page-title">Ringkasan <em>Hari&nbsp;Ini</em></h1>
				<p class="page-subtitle">
					ppPotret cepat saldo, pengeluaran, dan pemakaian budget untuk menjaga irama keuangan.
				</p>
			</div>
			<button class="button-secondary" type="button" onclick={loadSummary}>↻ Segarkan</button>
		</div>
	</header>

	{#if errorMessage}
		<p class="error">{errorMessage}</p>
	{/if}

	{#if loading}
		<div class="skeleton-hero">
			<p class="muted mono">Memuat lembar ringkasan…</p>
			<div class="skeleton-bar"></div>
			<div class="skeleton-bar short"></div>
		</div>
	{:else if summary}
		{@const balance = splitRupiah(summary.current_balance)}
		<!-- HERO: the big feature spread -->
		<article class="hero">
			<div class="hero-stamp">
				<div class="seal">
					<span>₨</span>
				</div>
			</div>

			<div class="hero-meta">
				<span class="mono tiny">Saldo Kas</span>
				<span class="hero-meta-rule"></span>
				<span class="mono tiny">{monthLabel}</span>
			</div>

			<p class="hero-number torn-edge" class:is-negative={balance.negative}>
				<span class="hero-mark">{balance.mark}</span>
				{#if balance.negative}<span class="hero-neg">−</span>{/if}
				<span class="hero-digits">{balance.number}</span>
			</p>

			<div class="hero-caption">
				<em>"Catatan harian,"</em> tertulis pada hari ke-{dayOfMonth}
				dari {lastDayOfMonth} · bulan {monthProgress}% berjalan.
			</div>

			<div class="hero-grid">
				<div class="hero-item">
					<span class="mono tiny">Gaji · Bulan Ini</span>
					<span class="hero-item-num money-display">{formatRupiah(summary.salary_current_month)}</span>
				</div>
				<div class="hero-item">
					<span class="mono tiny">Akumulasi Gaji</span>
					<span class="hero-item-num money-display">{formatRupiah(summary.salary_total_to_date)}</span>
				</div>
				<div class="hero-item">
					<span class="mono tiny">Sisa Budget</span>
					<span class="hero-item-num money-display">{formatRupiah(summary.remaining_budget)}</span>
				</div>
			</div>

			<!-- Month progress bar baked into hero -->
			<div class="month-progress">
				<div class="mp-rail">
					<span style="width: {monthProgress}%"></span>
				</div>
				<div class="mp-labels">
					<span class="mono tiny">Awal bulan</span>
					<span class="mono tiny mp-marker">Hari {dayOfMonth}</span>
					<span class="mono tiny">Akhir bulan</span>
				</div>
			</div>
		</article>

		<!-- Weekly sparkline -->
		<article class="spark-article">
			<div class="spark-head">
				<div>
					<p class="mono tiny">Tujuh Hari Terakhir · Pengeluaran</p>
					<p class="spark-total money-display">{formatRupiah(weekTotal)}</p>
				</div>
				<div class="spark-note">
					<span class="mono tiny">Tertinggi</span>
					<span class="money-mono">{formatRupiah(sparklineMax)}</span>
				</div>
			</div>

			<svg class="spark-svg" viewBox="0 0 100 30" preserveAspectRatio="none" aria-hidden="true">
				<defs>
					<pattern id="sparkHatch" width="3" height="3" patternUnits="userSpaceOnUse" patternTransform="rotate(45)">
						<line x1="0" y1="0" x2="0" y2="3" stroke="var(--oxblood)" stroke-width="0.4" opacity="0.5"/>
					</pattern>
				</defs>
				<!-- grid lines -->
				<line x1="0" y1="10" x2="100" y2="10" stroke="var(--rule)" stroke-width="0.2" stroke-dasharray="0.4 0.6"/>
				<line x1="0" y1="20" x2="100" y2="20" stroke="var(--rule)" stroke-width="0.2" stroke-dasharray="0.4 0.6"/>
				<path d={sparklineAreaPath} fill="url(#sparkHatch)"/>
				<path d={sparklinePath} fill="none" stroke="var(--oxblood)" stroke-width="0.8" stroke-linejoin="round" stroke-linecap="round"/>
				{#each sparklineData as d, i}
					{@const x = (i / (sparklineData.length - 1)) * 100}
					{@const y = 30 - (d.total / sparklineMax) * 26 - 2}
					<circle cx={x} cy={y} r="0.7" fill="var(--paper)" stroke="var(--ink)" stroke-width="0.5"/>
				{/each}
			</svg>

			<div class="spark-labels">
				{#each sparklineData as d, i}
					<span class="spark-day" class:is-today={i === sparklineData.length - 1}>
						{d.label}
					</span>
				{/each}
			</div>
		</article>

		<!-- Summary cards -->
		<div class="card-grid wide">
			<article class="card expense">
				<p class="card-title">Pengeluaran · Hari Ini</p>
				<p class="card-value money">{formatRupiah(summary.today_expense)}</p>
				<p class="card-note">{todayLabel}</p>
			</article>
			<article class="card expense">
				<p class="card-title">Pengeluaran · Bulan Ini</p>
				<p class="card-value money">{formatRupiah(summary.month_expense)}</p>
				<p class="card-note">{monthLabel}</p>
			</article>
			{#if makanMonthlyBudget}
				<article class="card info">
					<p class="card-title">Sisa · Budget Makan</p>
					<p class="card-value money">{formatRupiah(makanMonthlyBudget.remaining)}</p>
					<p class="card-note">
						{formatRupiah(makanMonthlyBudget.used)} / {formatRupiah(makanMonthlyBudget.limit)}
					</p>
				</article>
			{/if}
			{#if bensinMonthlyBudget}
				<article class="card info">
					<p class="card-title">Sisa · Budget Bensin</p>
					<p class="card-value money">{formatRupiah(bensinMonthlyBudget.remaining)}</p>
					<p class="card-note">
						{formatRupiah(bensinMonthlyBudget.used)} / {formatRupiah(bensinMonthlyBudget.limit)}
					</p>
				</article>
			{/if}
		</div>

		<section class="section-card">
			<h2 class="section-title">Limit Utama</h2>
			<p class="section-lede">
				Dua kantong yang wajib dijaga setiap hari.
			</p>
			<div class="limits">
				{#if makanBudget}
					<article class="limit-row">
						<header class="limit-head">
							<div>
								<p class="mono tiny">Kantong · Harian</p>
								<h3 class="limit-name">Makan</h3>
							</div>
							<p class="limit-ratio money-display">
								{formatRupiah(makanBudget.used)}
								<span class="ratio-slash">/</span>
								<span class="ratio-limit">{formatRupiah(makanBudget.limit)}</span>
							</p>
						</header>
						<div class="progress {makanBudget.over_budget ? 'over' : ''}">
							<span style="width: {budgetPercent(makanBudget.percentage)}%"></span>
						</div>
						<div class="progress-label">
							<span>{Math.round(makanBudget.percentage)}% terpakai</span>
							<span>Sisa {formatRupiah(makanBudget.remaining)}</span>
						</div>
					</article>
				{/if}

				{#if bensinBudget}
					<article class="limit-row">
						<header class="limit-head">
							<div>
								<p class="mono tiny">Kantong · Bulanan</p>
								<h3 class="limit-name">Bensin</h3>
							</div>
							<p class="limit-ratio money-display">
								{formatRupiah(bensinBudget.used)}
								<span class="ratio-slash">/</span>
								<span class="ratio-limit">{formatRupiah(bensinBudget.limit)}</span>
							</p>
						</header>
						<div class="progress {bensinBudget.over_budget ? 'over' : ''}">
							<span style="width: {budgetPercent(bensinBudget.percentage)}%"></span>
						</div>
						<div class="progress-label">
							<span>{Math.round(bensinBudget.percentage)}% terpakai</span>
							<span>Sisa {formatRupiah(bensinBudget.remaining)}</span>
						</div>
					</article>
				{/if}
			</div>
		</section>

		<section class="section-card">
			<h2 class="section-title">Pemakaian Budget Lainnya</h2>
			<p class="section-lede">
				Setiap kategori diukur terhadap batas yang sudah ditetapkan.
			</p>
			{#if summary.budget_usage.filter(u => u.category !== 'Makan' && u.category !== 'Bensin').length === 0}
				<p class="muted">Belum ada aturan budget tambahan.</p>
			{:else}
				<div class="usage-list">
					{#each summary.budget_usage.filter(u => u.category !== 'Makan' && u.category !== 'Bensin') as usage, i}
						<article class="usage-row">
							<div class="usage-head">
								<span class="ledger-num">{String(i + 1).padStart(2, '0')}</span>
								<div class="usage-text">
									<strong>{usage.category}</strong>
									<span class="mono tiny">{formatPeriod(usage.period)}</span>
								</div>
								<span class="leader"></span>
								<span class="money-mono usage-amount">
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

		<div class="fleuron">
			<span class="fleuron-mark">❦</span>
		</div>

		<p class="colophon mono tiny">
			Dicetak digital · <em>Sunday Edition</em> · Tinta hitam &amp; oxblood
		</p>
	{/if}
</section>

<style>
	.tiny {
		font-size: 0.6rem !important;
		letter-spacing: 0.15em;
		text-transform: uppercase;
	}

	.skeleton-hero {
		padding: 2rem 1rem;
		border: 1px solid var(--rule);
		display: grid;
		gap: 0.75rem;
	}

	.skeleton-bar {
		height: 2.5rem;
		background: linear-gradient(
			90deg,
			var(--paper-deep),
			var(--paper-fold),
			var(--paper-deep)
		);
		background-size: 200% 100%;
		animation: shimmer 1.6s ease-in-out infinite;
	}

	.skeleton-bar.short {
		height: 1rem;
		width: 60%;
	}

	@keyframes shimmer {
		0% { background-position: -100% 0; }
		100% { background-position: 200% 0; }
	}

	/* Hero — the main feature */
	.hero {
		border: 1.5px solid var(--ink);
		padding: 1.5rem 1.25rem 1.25rem;
		background: var(--paper);
		position: relative;
		overflow: hidden;
	}

	.hero::before {
		content: '';
		position: absolute;
		top: 0.4rem;
		left: 0.4rem;
		right: 0.4rem;
		bottom: 0.4rem;
		border: 1px solid var(--ink);
		pointer-events: none;
		opacity: 0.15;
	}

	.hero-stamp {
		position: absolute;
		top: 1.25rem;
		right: 1.25rem;
		z-index: 2;
		transform: rotate(8deg);
		opacity: 0.85;
	}

	.hero-stamp .seal {
		width: 3rem;
		height: 3rem;
		font-size: 1.1rem;
	}

	.hero-meta {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		color: var(--ink);
		margin-bottom: 0.85rem;
		font-family: var(--font-mono);
		padding-right: 4rem;
	}

	.hero-meta-rule {
		flex: 1;
		max-width: 6rem;
		height: 1px;
		background: var(--ink);
		opacity: 0.35;
	}

	.hero-number {
		margin: 0;
		font-family: var(--font-display);
		font-weight: 400;
		font-size: clamp(3rem, 13vw, 6rem);
		line-height: 0.9;
		letter-spacing: -0.035em;
		color: var(--ink);
		font-variant-numeric: tabular-nums;
		display: flex;
		align-items: flex-start;
		gap: 0.1em;
		padding-bottom: 0.5rem;
	}

	.hero-number.is-negative {
		color: var(--oxblood);
	}

	.hero-mark {
		font-family: var(--font-mono);
		font-size: 0.2em;
		letter-spacing: 0.1em;
		color: var(--ink-soft);
		margin-top: 0.85em;
		text-transform: uppercase;
		font-weight: 500;
	}

	.hero-neg {
		font-style: italic;
		color: var(--oxblood);
	}

	.hero-digits {
		font-style: italic;
	}

	.hero-caption {
		margin: 0.5rem 0 0;
		font-family: var(--font-display);
		font-style: italic;
		font-size: 0.95rem;
		color: var(--ink-soft);
		max-width: 46ch;
		line-height: 1.45;
	}

	.hero-caption em {
		color: var(--oxblood);
	}

	.hero-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 0;
		margin-top: 1.25rem;
		padding-top: 1rem;
		border-top: 1.5px solid var(--ink);
	}

	.hero-item {
		display: flex;
		flex-direction: column;
		gap: 0.3rem;
		padding: 0 0.5rem 0 0;
		border-right: 1px dashed var(--rule);
	}

	.hero-item:first-child {
		padding-left: 0;
	}

	.hero-item:not(:first-child) {
		padding-left: 0.5rem;
	}

	.hero-item:last-child {
		border-right: 0;
	}

	.hero-item-num {
		font-size: 1rem;
		color: var(--ink);
		line-height: 1;
		letter-spacing: -0.01em;
	}

	.month-progress {
		margin-top: 1rem;
		padding-top: 0.85rem;
		border-top: 1px dashed var(--rule);
	}

	.mp-rail {
		height: 4px;
		background: var(--paper-fold);
		position: relative;
		overflow: hidden;
		border: 1px solid var(--rule);
	}

	.mp-rail > span {
		display: block;
		height: 100%;
		background: var(--ink);
		background-image: repeating-linear-gradient(
			90deg,
			var(--ink),
			var(--ink) 4px,
			var(--paper) 4px,
			var(--paper) 5px
		);
		transition: width 0.8s cubic-bezier(0.22, 1, 0.36, 1);
	}

	.mp-labels {
		display: flex;
		justify-content: space-between;
		margin-top: 0.5rem;
		color: var(--ink-soft);
	}

	.mp-marker {
		color: var(--oxblood) !important;
		font-weight: 600;
	}

	/* Sparkline */
	.spark-article {
		border: 1px solid var(--ink);
		padding: 1.25rem 1.25rem 0.85rem;
		background: var(--paper);
	}

	.spark-head {
		display: flex;
		justify-content: space-between;
		align-items: flex-end;
		gap: 1rem;
		margin-bottom: 0.6rem;
	}

	.spark-total {
		margin: 0.2rem 0 0;
		font-size: 1.5rem;
		line-height: 1;
		color: var(--ink);
	}

	.spark-note {
		text-align: right;
		display: flex;
		flex-direction: column;
		gap: 0.15rem;
	}

	.spark-svg {
		width: 100%;
		height: 4rem;
		display: block;
	}

	.spark-labels {
		display: flex;
		justify-content: space-between;
		padding-top: 0.35rem;
		border-top: 1px dashed var(--rule);
		font-family: var(--font-mono);
		font-size: 0.65rem;
		color: var(--ink-faint);
		letter-spacing: 0.1em;
		text-transform: uppercase;
	}

	.spark-day {
		flex: 1;
		text-align: center;
	}

	.spark-day.is-today {
		color: var(--oxblood);
		font-weight: 700;
	}

	/* Limits */
	.limits {
		display: grid;
		gap: 1.25rem;
	}

	.limit-row {
		padding: 1rem 0;
		border-top: 1px solid var(--rule);
	}

	.limit-row:first-child {
		border-top: 0;
		padding-top: 0;
	}

	.limit-head {
		display: flex;
		justify-content: space-between;
		align-items: flex-end;
		gap: 1rem;
		margin-bottom: 0.75rem;
		flex-wrap: wrap;
	}

	.limit-name {
		margin: 0.15rem 0 0;
		font-family: var(--font-display);
		font-size: 1.5rem;
		line-height: 1;
		color: var(--ink);
	}

	.limit-ratio {
		font-size: 1.35rem;
		color: var(--ink);
	}

	.ratio-slash {
		color: var(--ink-faint);
		font-style: italic;
		margin: 0 0.15em;
	}

	.ratio-limit {
		color: var(--ink-soft);
		font-size: 0.85em;
	}

	/* Usage list */
	.usage-list {
		display: grid;
		gap: 1rem;
	}

	.usage-row {
		padding: 0.75rem 0;
		border-top: 1px solid var(--rule);
	}

	.usage-row:first-child {
		border-top: 0;
		padding-top: 0;
	}

	.usage-head {
		display: flex;
		align-items: baseline;
		gap: 0.5rem;
		margin-bottom: 0.5rem;
	}

	.usage-text {
		display: flex;
		flex-direction: column;
	}

	.usage-text strong {
		font-weight: 600;
		font-size: 0.95rem;
	}

	.usage-amount {
		font-size: 0.8rem;
		color: var(--ink);
	}

	.colophon {
		text-align: center;
		color: var(--ink-faint);
	}

	.colophon em {
		font-family: var(--font-display);
		font-size: 0.85rem;
		text-transform: none;
		letter-spacing: 0;
		color: var(--oxblood);
	}

	@media (min-width: 768px) {
		.hero {
			padding: 2rem 1.75rem 1.5rem;
		}

		.hero-stamp {
			top: 1.75rem;
			right: 1.75rem;
		}

		.hero-stamp .seal {
			width: 4rem;
			height: 4rem;
			font-size: 1.5rem;
		}

		.hero-item-num {
			font-size: 1.15rem;
		}

		.spark-svg {
			height: 5rem;
		}
	}
</style>
