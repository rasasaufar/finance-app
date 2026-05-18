<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { formatRupiah } from '$lib/format';
	import DatePicker from '$lib/DatePicker.svelte';
	import { api, ApiError } from '$lib/api';
	import type { WeddingSummary, WeddingDeposit, WeddingConfig } from '$lib/types';

	// ──────────────────────────────────────────────────────────
	// TYPES & CONSTANTS
	// ──────────────────────────────────────────────────────────

	const FROM_LABEL: Record<WeddingDeposit['source'], string> = {
		self: 'Tabungan Pribadi',
		partner: 'Pasangan',
		gift: 'Hadiah / Hibah',
		other: 'Sumber Lain'
	};

	const FROM_GLYPH: Record<WeddingDeposit['source'], string> = {
		self: '𓂀',
		partner: '∞',
		gift: '✦',
		other: '◆'
	};

	// ──────────────────────────────────────────────────────────
	// STATE
	// ──────────────────────────────────────────────────────────
	let mounted = $state(false);
	let loading = $state(true);
	let saving = $state(false);

	let config = $state<WeddingConfig>({
		target_amount: 50_000_000,
		target_date: '',
		bride_name: '',
		groom_name: '',
		venue: ''
	});
	let deposits = $state<WeddingDeposit[]>([]);

	// Form (deposit)
	let depositAmount = $state('');
	let depositDate = $state(new Date().toISOString().slice(0, 10));
	let depositFrom = $state<WeddingDeposit['source']>('self');
	let depositNote = $state('');
	let editingId = $state<number | null>(null);

	// Confetti
	let confettiPieces = $state<{ id: number; left: number; delay: number; rotate: number; color: string; size: number }[]>([]);
	let lastConfettiMilestone = $state(0);

	// Toast
	let toast = $state<{ message: string; tone: 'good' | 'warn' } | null>(null);
	let toastTimer: ReturnType<typeof setTimeout> | null = null;
	function showToast(message: string, tone: 'good' | 'warn' = 'good'): void {
		toast = { message, tone };
		if (toastTimer) clearTimeout(toastTimer);
		toastTimer = setTimeout(() => (toast = null), 2400);
	}

	// Delete confirmation
	let pendingDelete = $state<WeddingDeposit | null>(null);

	// ──────────────────────────────────────────────────────────
	// DERIVED
	// ──────────────────────────────────────────────────────────
	const total = $derived(deposits.reduce((sum: number, d: WeddingDeposit) => sum + d.amount, 0));
	const progressPct = $derived(
		config.target_amount > 0 ? Math.min((total / config.target_amount) * 100, 100) : 0
	);

	const sortedDeposits = $derived(
		[...deposits].sort((a: WeddingDeposit, b: WeddingDeposit) => {
			const cmp = b.date.localeCompare(a.date);
			return cmp !== 0 ? cmp : b.id - a.id;
		})
	);

	const todayLabel = new Intl.DateTimeFormat('id-ID', {
		weekday: 'long',
		day: 'numeric',
		month: 'long',
		year: 'numeric'
	}).format(new Date());

	const issueNumber = new Intl.DateTimeFormat('id-ID', { year: 'numeric', month: '2-digit' })
		.format(new Date())
		.replace('/', '.');

	const targetDateLabel = $derived.by(() => {
		if (!config.target_date) return '—';
		const d = new Date(`${config.target_date}T00:00:00`);
		if (isNaN(d.getTime())) return config.target_date;
		return new Intl.DateTimeFormat('id-ID', {
			weekday: 'long',
			day: 'numeric',
			month: 'long',
			year: 'numeric'
		}).format(d);
	});

	const avgPerMonth = $derived.by(() => {
		if (deposits.length === 0) return 0;
		const months = new Set(deposits.map((d: WeddingDeposit) => d.date.slice(0, 7)));
		return total / Math.max(months.size, 1);
	});

	const lastDeposit = $derived(sortedDeposits[0] ?? null);

	// SVG ring math
	const RING_RADIUS = 92;
	const RING_CIRC = 2 * Math.PI * RING_RADIUS;
	const ringDashOffset = $derived(RING_CIRC - (progressPct / 100) * RING_CIRC);

	// ──────────────────────────────────────────────────────────
	// CONFETTI
	// ──────────────────────────────────────────────────────────
	function fireConfetti(): void {
		const colors = ['#a9812c', '#c9973a', '#e8c46a', '#8b2e2e', '#2f5d46', '#f3ecde'];
		confettiPieces = Array.from({ length: 36 }, (_, i) => ({
			id: Date.now() + i,
			left: Math.random() * 100,
			delay: Math.random() * 0.4,
			rotate: Math.random() * 360,
			color: colors[Math.floor(Math.random() * colors.length)],
			size: 6 + Math.random() * 8
		}));
		setTimeout(() => (confettiPieces = []), 2800);
	}

	// ──────────────────────────────────────────────────────────
	// ACTIONS
	// ──────────────────────────────────────────────────────────
	function resetDepositForm(): void {
		depositAmount = '';
		depositDate = new Date().toISOString().slice(0, 10);
		depositFrom = 'self';
		depositNote = '';
		editingId = null;
	}

	async function loadData(): Promise<void> {
		loading = true;
		try {
			const summary = await api.get<WeddingSummary>('/wedding');
			config = summary.config;
			deposits = summary.deposits ?? [];
		} catch (err) {
			if (err instanceof ApiError) {
				showToast(err.message, 'warn');
			} else {
				showToast('Gagal memuat data tabungan nikah', 'warn');
			}
		} finally {
			loading = false;
		}
	}

	async function handleSubmitDeposit(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		const numeric = Number(depositAmount);
		if (!Number.isFinite(numeric) || numeric <= 0) {
			showToast('Nominal harus lebih dari 0', 'warn');
			return;
		}
		if (!depositDate) {
			showToast('Tanggal wajib diisi', 'warn');
			return;
		}

		const before = progressPct;
		saving = true;

		try {
			if (editingId != null) {
				const updated = await api.put<WeddingDeposit>(`/wedding/deposits/${editingId}`, {
					date: depositDate,
					amount: numeric,
					source: depositFrom,
					note: depositNote.trim()
				});
				const idx = deposits.findIndex((d: WeddingDeposit) => d.id === editingId);
				if (idx >= 0) {
					deposits[idx] = updated;
					deposits = [...deposits];
				}
				showToast('Setoran berhasil diperbarui');
			} else {
				const created = await api.post<WeddingDeposit>('/wedding/deposits', {
					date: depositDate,
					amount: numeric,
					source: depositFrom,
					note: depositNote.trim()
				});
				deposits = [...deposits, created];
				showToast('Setoran tercatat di brankas');
			}

			resetDepositForm();

			// Confetti on big progress jumps (every 25%)
			await tick();
			const after = progressPct;
			const beforeBucket = Math.floor(before / 25);
			const afterBucket = Math.floor(after / 25);
			if (afterBucket > beforeBucket && afterBucket > lastConfettiMilestone) {
				lastConfettiMilestone = afterBucket;
				fireConfetti();
			}
		} catch (err) {
			if (err instanceof ApiError) {
				showToast(err.message, 'warn');
			} else {
				showToast('Gagal menyimpan setoran', 'warn');
			}
		} finally {
			saving = false;
		}
	}

	function handleEditDeposit(d: WeddingDeposit): void {
		editingId = d.id;
		depositAmount = String(d.amount);
		depositDate = d.date;
		depositFrom = d.source;
		depositNote = d.note;
		// Scroll to form
		const el = document.getElementById('form');
		if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' });
	}

	function requestDelete(d: WeddingDeposit): void {
		pendingDelete = d;
	}

	async function confirmDelete(): Promise<void> {
		if (!pendingDelete) return;
		const deletedId = pendingDelete.id;
		saving = true;
		try {
			await api.delete(`/wedding/deposits/${deletedId}`);
			deposits = deposits.filter((x: WeddingDeposit) => x.id !== deletedId);
			pendingDelete = null;
			showToast('Setoran dihapus dari brankas');
			if (editingId === deletedId) resetDepositForm();
		} catch (err) {
			if (err instanceof ApiError) {
				showToast(err.message, 'warn');
			} else {
				showToast('Gagal menghapus setoran', 'warn');
			}
		} finally {
			saving = false;
		}
	}

	function cancelDelete(): void {
		pendingDelete = null;
	}

	function formatDateShort(value: string): string {
		const d = new Date(`${value}T00:00:00`);
		if (isNaN(d.getTime())) return value;
		return new Intl.DateTimeFormat('id-ID', {
			day: 'numeric',
			month: 'short',
			year: 'numeric'
		}).format(d);
	}

	// ──────────────────────────────────────────────────────────
	// LIFECYCLE
	// ──────────────────────────────────────────────────────────
	onMount(async () => {
		await loadData();
		mounted = true;
		lastConfettiMilestone = Math.floor(progressPct / 25);
	});
</script>

<svelte:head>
	<title>Tabungan Nikah — Edisi Khusus</title>
</svelte:head>

<section class="page wedding-page">
	<!-- ─── MASTHEAD ─── -->
	<header class="vault-masthead">
		<div class="masthead-flourish left" aria-hidden="true">
			<svg viewBox="0 0 80 24" fill="none">
				<path d="M0 12 H30 M50 12 H80" stroke="currentColor" stroke-width="0.8" />
				<circle cx="40" cy="12" r="3" stroke="currentColor" stroke-width="0.8" fill="none" />
				<circle cx="40" cy="12" r="1" fill="currentColor" />
				<path d="M30 12 L36 8 M30 12 L36 16 M50 12 L44 8 M50 12 L44 16" stroke="currentColor" stroke-width="0.8" />
			</svg>
		</div>

		<p class="vault-eyebrow">Edisi Istimewa · Brankas {issueNumber}</p>
		<h1 class="vault-title">
			Tabungan
			<em>Nikah</em>
		</h1>
		<p class="vault-tagline">
			«&nbsp;Janji yang dirayakan dengan kesabaran, ditabung sehelai demi sehelai.&nbsp;»
		</p>

		<div class="vault-couple">
			<span class="couple-name">{config.bride_name}</span>
			<span class="couple-amp">&amp;</span>
			<span class="couple-name">{config.groom_name}</span>
		</div>

		<div class="vault-meta">
			<span>{todayLabel}</span>
			<span class="meta-bullet">·</span>
			<span>Akad direncanakan {targetDateLabel}</span>
		</div>

		<div class="masthead-flourish right" aria-hidden="true">
			<svg viewBox="0 0 80 24" fill="none">
				<path d="M0 12 H30 M50 12 H80" stroke="currentColor" stroke-width="0.8" />
				<circle cx="40" cy="12" r="3" stroke="currentColor" stroke-width="0.8" fill="none" />
				<circle cx="40" cy="12" r="1" fill="currentColor" />
				<path d="M30 12 L36 8 M30 12 L36 16 M50 12 L44 8 M50 12 L44 16" stroke="currentColor" stroke-width="0.8" />
			</svg>
		</div>
	</header>

	<!-- ─── HERO RING + STATS ─── -->
	<section class="vault-hero">
		<div class="ring-stage">
			<svg class="ring-svg" viewBox="0 0 220 220" aria-hidden="true">
				<defs>
					<linearGradient id="goldGrad" x1="0" y1="0" x2="1" y2="1">
						<stop offset="0%" stop-color="#e8c46a" />
						<stop offset="50%" stop-color="#c9973a" />
						<stop offset="100%" stop-color="#a9812c" />
					</linearGradient>
					<filter id="goldGlow" x="-20%" y="-20%" width="140%" height="140%">
						<feGaussianBlur stdDeviation="2" result="blur" />
						<feMerge>
							<feMergeNode in="blur" />
							<feMergeNode in="SourceGraphic" />
						</feMerge>
					</filter>
				</defs>

				<!-- decorative outer dotted ring -->
				<circle cx="110" cy="110" r="106" fill="none" stroke="currentColor" stroke-width="0.6" stroke-dasharray="1 4" opacity="0.4" />

				<!-- track -->
				<circle cx="110" cy="110" r={RING_RADIUS} fill="none" stroke="var(--rule)" stroke-width="6" />

				<!-- progress -->
				<circle
					cx="110"
					cy="110"
					r={RING_RADIUS}
					fill="none"
					stroke="url(#goldGrad)"
					stroke-width="8"
					stroke-linecap="round"
					stroke-dasharray={RING_CIRC}
					stroke-dashoffset={mounted ? ringDashOffset : RING_CIRC}
					transform="rotate(-90 110 110)"
					filter="url(#goldGlow)"
					style="transition: stroke-dashoffset 1.4s cubic-bezier(0.22, 1, 0.36, 1);"
				/>

				<!-- tick marks at quarters -->
				{#each [0, 25, 50, 75] as t}
					{@const angle = (t / 100) * 360 - 90}
					{@const rad = (angle * Math.PI) / 180}
					{@const x1 = 110 + Math.cos(rad) * (RING_RADIUS - 14)}
					{@const y1 = 110 + Math.sin(rad) * (RING_RADIUS - 14)}
					{@const x2 = 110 + Math.cos(rad) * (RING_RADIUS - 4)}
					{@const y2 = 110 + Math.sin(rad) * (RING_RADIUS - 4)}
					<line {x1} {y1} {x2} {y2} stroke="var(--ink)" stroke-width="0.8" opacity="0.5" />
				{/each}

				<!-- centerpiece flourish -->
				<path d="M105 95 Q110 88 115 95 Q113 100 110 100 Q107 100 105 95 Z" fill="var(--oxblood)" opacity="0.7" />
			</svg>

			<div class="ring-center">
				<p class="ring-eyebrow">Terkumpul</p>
				<p class="ring-amount">{formatRupiah(total)}</p>
				<p class="ring-pct">{progressPct.toFixed(1).replace('.', ',')}%</p>
				<p class="ring-of">dari {formatRupiah(config.target_amount)}</p>
			</div>
		</div>

		<div class="hero-stats">
			<article class="stat-card">
				<p class="stat-label">Rata-rata Aktual</p>
				<p class="stat-value money">{formatRupiah(Math.round(avgPerMonth))}</p>
				<p class="stat-note">per bulan dari riwayat setoran</p>
			</article>

			<article class="stat-card">
				<p class="stat-label">Setoran Terakhir</p>
				<p class="stat-value money small">
					{lastDeposit ? formatRupiah(lastDeposit.amount) : '—'}
				</p>
				<p class="stat-note">
					{lastDeposit ? formatDateShort(lastDeposit.date) : 'Belum ada riwayat'}
				</p>
			</article>
		</div>
	</section>

	<!-- ─── DEPOSIT FORM ─── -->
	<section id="form" class="section-card form-card">
		<div class="form-card-head">
			<div>
				<p class="muted mono tiny-eyebrow">Kanal Setoran · Brankas Pribadi</p>
				<h2 class="section-title">{editingId != null ? 'Edit Setoran' : 'Setor Tabungan Nikah'}</h2>
			</div>
			<div class="seal" aria-hidden="true">
				<span class="seal-inner">SEAL</span>
				<span class="seal-outer">· DOMPET PRIBADI · BRANKAS NIKAH ·</span>
			</div>
		</div>

		<form class="form-grid" onsubmit={handleSubmitDeposit}>
			<div class="form-row">
				<label class="field">
					<span>Nominal Setoran (Rp)</span>
					<input
						type="number"
						bind:value={depositAmount}
						min="1"
						placeholder="Contoh: 1500000"
						required
					/>
				</label>
				<DatePicker bind:value={depositDate} label="Tanggal Setor" required />
			</div>

			<div class="form-row">
				<label class="field">
					<span>Sumber Dana</span>
					<select bind:value={depositFrom}>
						<option value="self">Tabungan Pribadi</option>
						<option value="partner">Pasangan</option>
						<option value="gift">Hadiah / Hibah</option>
						<option value="other">Sumber Lain</option>
					</select>
				</label>
				<label class="field">
					<span>Catatan (opsional)</span>
					<input
						type="text"
						bind:value={depositNote}
						placeholder="Misal: Bonus tahunan, hasil jual emas"
					/>
				</label>
			</div>

			<div class="quick-amounts">
				<span class="muted mono tiny-eyebrow">Setor cepat</span>
				{#each [100_000, 500_000, 1_000_000, 2_500_000, 5_000_000] as q}
					<button type="button" class="quick-chip" onclick={() => (depositAmount = String(q))}>
						{formatRupiah(q).replace('Rp', 'Rp ')}
					</button>
				{/each}
			</div>

			<div class="button-row">
				<button class="button-primary lux-primary" type="submit">
					<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="14" height="14"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
					{editingId != null ? 'Perbarui Setoran' : 'Simpan ke Brankas'}
				</button>
				{#if editingId != null}
					<button class="button-secondary" type="button" onclick={resetDepositForm}>Batal Edit</button>
				{/if}
			</div>
		</form>
	</section>

	<!-- ─── DEPOSIT LEDGER ─── -->
	<section class="section-card ledger-card">
		<div class="form-card-head">
			<div>
				<p class="muted mono tiny-eyebrow">Buku Setoran · Diurutkan dari yang terbaru</p>
				<h2 class="section-title">Riwayat Brankas</h2>
			</div>
			<p class="ledger-count mono">
				{deposits.length} catatan
			</p>
		</div>

		{#if sortedDeposits.length === 0}
			<div class="ledger-empty">
				<div class="empty-art" aria-hidden="true">
					<svg viewBox="0 0 120 120" fill="none">
						<circle cx="40" cy="65" r="22" stroke="currentColor" stroke-width="1.2" />
						<circle cx="80" cy="65" r="22" stroke="currentColor" stroke-width="1.2" />
						<path d="M30 40 L40 60 M90 40 L80 60 M50 28 H70" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" />
						<circle cx="40" cy="65" r="2" fill="currentColor" />
						<circle cx="80" cy="65" r="2" fill="currentColor" />
					</svg>
				</div>
				<p class="empty-headline">Brankas masih kosong</p>
				<p class="muted">Mulai dengan setoran pertama; sekecil apapun, tetap berarti.</p>
			</div>
		{:else}
			<ul class="ledger-list">
				{#each sortedDeposits as d, i (d.id)}
					{@const isFirst = i === sortedDeposits.length - 1}
					<li class="ledger-item" class:editing={editingId === d.id}>
						<div class="ledger-glyph" aria-hidden="true">
							<span class="glyph-symbol">{FROM_GLYPH[d.source]}</span>
						</div>
						<div class="ledger-main">
							<div class="ledger-row-top">
								<span class="ledger-from">{FROM_LABEL[d.source]}</span>
								<span class="ledger-date mono">{formatDateShort(d.date)}</span>
							</div>
							{#if d.note}
								<p class="ledger-note">"{d.note}"</p>
							{:else}
								<p class="ledger-note muted-italic">Tanpa catatan khusus</p>
							{/if}
							{#if isFirst}
								<span class="ledger-badge">✦ Setoran Pertama</span>
							{/if}
						</div>
						<div class="ledger-amount-col">
							<p class="ledger-amount money">+{formatRupiah(d.amount)}</p>
							<div class="ledger-actions">
								<button class="button-ghost" type="button" onclick={() => handleEditDeposit(d)}>
									Edit
								</button>
								<button class="button-ghost danger" type="button" onclick={() => requestDelete(d)}>
									Hapus
								</button>
							</div>
						</div>
					</li>
				{/each}
			</ul>

			<div class="ledger-footer">
				<span class="muted mono tiny-eyebrow">Total Brankas</span>
				<span class="ledger-total money">{formatRupiah(total)}</span>
			</div>
		{/if}
	</section>

	<!-- ─── CLOSING QUOTE ─── -->
	<aside class="vault-quote">
		<p class="quote-mark">”</p>
		<p class="quote-body">
			Setiap rupiah di brankas ini bukan sekadar angka, melainkan sehelai janji yang sedang ditenun
			menjadi pelaminan.
		</p>
		<p class="quote-cite">— Catatan kaki, Buku Kas {issueNumber}</p>
	</aside>
</section>

<!-- ─── DELETE CONFIRM MODAL ─── -->
{#if pendingDelete}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div class="modal-backdrop" onclick={cancelDelete}>
		<div class="modal-card" onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1">
			<p class="muted mono" style="margin-bottom: 0.35rem;">Konfirmasi · Hapus Setoran</p>
			<h3 class="section-title" style="margin-bottom: 1rem;">Hapus dari Brankas?</h3>
			<div class="confirm-detail-list">
				<div class="confirm-detail-row">
					<span>Tanggal</span>
					<strong>{formatDateShort(pendingDelete.date)}</strong>
				</div>
				<div class="confirm-detail-row">
					<span>Sumber</span>
					<strong>{FROM_LABEL[pendingDelete.source]}</strong>
				</div>
				<div class="confirm-detail-row">
					<span>Nominal</span>
					<strong>{formatRupiah(pendingDelete.amount)}</strong>
				</div>
				<div class="confirm-detail-row">
					<span>Catatan</span>
					<strong>{pendingDelete.note || '—'}</strong>
				</div>
			</div>
			<p class="muted" style="font-style: italic; font-family: var(--font-display); font-size: 0.95rem;">
				Setoran yang dihapus tidak bisa dikembalikan, namun visi tetap utuh.
			</p>
			<div class="button-row" style="justify-content: flex-end;">
				<button class="button-secondary" type="button" onclick={cancelDelete}>Batal</button>
				<button class="button-danger" type="button" onclick={confirmDelete}>Ya, Hapus</button>
			</div>
		</div>
	</div>
{/if}

<!-- ─── CONFETTI ─── -->
{#if confettiPieces.length > 0}
	<div class="confetti-stage" aria-hidden="true">
		{#each confettiPieces as p (p.id)}
			<span
				class="confetti-piece"
				style="left: {p.left}%; animation-delay: {p.delay}s; --rot: {p.rotate}deg; background: {p.color}; width: {p.size}px; height: {p.size * 1.6}px;"
			></span>
		{/each}
	</div>
{/if}

<!-- ─── TOAST ─── -->
{#if toast}
	<div class="lux-toast" class:warn={toast.tone === 'warn'} role="status" aria-live="polite">
		<span class="toast-glyph">{toast.tone === 'warn' ? '!' : '✦'}</span>
		<span>{toast.message}</span>
	</div>
{/if}

<style>
	/* ─────────────────────────────────────────────────────────
	   PAGE-SCOPED LUXE PALETTE
	   ───────────────────────────────────────────────────────── */
	.wedding-page {
		--gold: #c9973a;
		--gold-soft: #e8c46a;
		--gold-deep: #a9812c;
		--rose: #b9697a;
		--rose-soft: #e6c9d1;
		--ivory: #faf3e3;
	}

	.wedding-page :global(.confirm-detail-row strong) {
		font-family: var(--font-mono);
	}

	/* ─────────────────────────────────────────────────────────
	   MASTHEAD
	   ───────────────────────────────────────────────────────── */
	.vault-masthead {
		text-align: center;
		padding: 2.5rem 1rem 2.25rem;
		border: 1.5px solid var(--ink);
		background:
			radial-gradient(ellipse 70% 80% at 50% 0%, rgba(232, 196, 106, 0.18), transparent 70%),
			radial-gradient(ellipse 60% 50% at 50% 100%, rgba(185, 105, 122, 0.1), transparent 70%),
			var(--paper);
		position: relative;
		overflow: hidden;
		box-shadow: 8px 8px 0 var(--ink);
	}

	.vault-masthead::before,
	.vault-masthead::after {
		content: '';
		position: absolute;
		left: 0;
		right: 0;
		height: 1px;
		background: var(--ink);
		opacity: 0.4;
	}

	.vault-masthead::before {
		top: 8px;
	}

	.vault-masthead::after {
		bottom: 8px;
	}

	.masthead-flourish {
		display: flex;
		justify-content: center;
		color: var(--gold-deep);
		opacity: 0.85;
	}

	.masthead-flourish svg {
		width: clamp(60px, 18vw, 110px);
		height: auto;
	}

	.masthead-flourish.left {
		margin-bottom: 1.1rem;
	}

	.masthead-flourish.right {
		margin-top: 1.1rem;
	}

	.vault-eyebrow {
		font-family: var(--font-mono);
		font-size: 0.62rem;
		letter-spacing: 0.32em;
		text-transform: uppercase;
		color: var(--gold-deep);
		margin: 0 0 0.85rem;
	}

	.vault-title {
		margin: 0 0 0.6rem;
		font-family: var(--font-display);
		font-weight: 400;
		font-size: clamp(2.75rem, 9vw, 5.5rem);
		line-height: 0.95;
		letter-spacing: -0.025em;
		color: var(--ink);
	}

	.vault-title em {
		font-style: italic;
		color: var(--gold-deep);
		display: inline-block;
		position: relative;
		padding: 0 0.15em;
	}

	.vault-title em::before,
	.vault-title em::after {
		content: '';
		position: absolute;
		top: 50%;
		width: 1.4rem;
		height: 1px;
		background: linear-gradient(90deg, transparent, var(--gold-deep), transparent);
	}

	.vault-title em::before {
		right: 100%;
		margin-right: 0.4rem;
	}

	.vault-title em::after {
		left: 100%;
		margin-left: 0.4rem;
	}

	.vault-tagline {
		max-width: 36rem;
		margin: 0.4rem auto 1.4rem;
		font-family: var(--font-display);
		font-style: italic;
		font-size: clamp(0.95rem, 2.4vw, 1.2rem);
		color: var(--ink-soft);
		line-height: 1.4;
	}

	.vault-couple {
		display: inline-flex;
		align-items: baseline;
		gap: 0.65rem;
		padding: 0.65rem 1.4rem;
		background: var(--paper-deep);
		border: 1px solid var(--gold-deep);
		border-radius: 999px;
		box-shadow: 0 0 0 4px var(--paper), 0 0 0 5px var(--gold-deep);
		font-family: var(--font-display);
		font-size: clamp(1rem, 2.4vw, 1.25rem);
		max-width: 100%;
	}

	.couple-name {
		font-style: italic;
		color: var(--ink);
	}

	.couple-amp {
		font-style: italic;
		color: var(--gold-deep);
		font-size: 1.4em;
		line-height: 0;
		transform: translateY(0.05em);
	}

	.vault-meta {
		display: flex;
		justify-content: center;
		flex-wrap: wrap;
		gap: 0.6rem;
		margin-top: 1rem;
		font-family: var(--font-mono);
		font-size: 0.65rem;
		letter-spacing: 0.16em;
		text-transform: uppercase;
		color: var(--ink-soft);
	}

	.meta-bullet {
		color: var(--gold-deep);
	}

	/* ─────────────────────────────────────────────────────────
	   HERO RING + STATS
	   ───────────────────────────────────────────────────────── */
	.vault-hero {
		display: grid;
		grid-template-columns: minmax(220px, 320px) 1fr;
		gap: 2rem;
		align-items: center;
		padding: 2rem 1.5rem;
		background: var(--paper);
		border: 1.5px solid var(--ink);
		border-top: 0;
		position: relative;
	}

	.vault-hero::before {
		content: '';
		position: absolute;
		inset: 6px;
		border: 1px dashed var(--gold-deep);
		opacity: 0.3;
		pointer-events: none;
	}

	.ring-stage {
		position: relative;
		aspect-ratio: 1;
		display: grid;
		place-items: center;
		color: var(--gold-deep);
	}

	.ring-svg {
		width: 100%;
		height: 100%;
		max-width: 320px;
	}

	.ring-center {
		position: absolute;
		text-align: center;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.05rem;
	}

	.ring-eyebrow {
		margin: 0;
		font-family: var(--font-mono);
		font-size: 0.6rem;
		letter-spacing: 0.28em;
		text-transform: uppercase;
		color: var(--gold-deep);
	}

	.ring-amount {
		margin: 0.15rem 0 0.25rem;
		font-family: var(--font-display);
		font-size: clamp(1.4rem, 4.4vw, 2rem);
		line-height: 1;
		letter-spacing: -0.02em;
		color: var(--ink);
		font-variant-numeric: tabular-nums;
	}

	.ring-pct {
		margin: 0;
		font-family: var(--font-display);
		font-style: italic;
		font-size: clamp(2.2rem, 6vw, 3.4rem);
		line-height: 1;
		color: var(--gold-deep);
		font-variant-numeric: tabular-nums;
		letter-spacing: -0.02em;
	}

	.ring-of {
		margin: 0.35rem 0 0;
		font-family: var(--font-mono);
		font-size: 0.62rem;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		color: var(--ink-faint);
	}

	.hero-stats {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 0;
		border: 1px solid var(--rule);
		background: var(--rule);
	}

	.stat-card {
		background: var(--paper);
		padding: 1.1rem 1rem;
		min-height: 7rem;
		position: relative;
	}

	.stat-label {
		margin: 0 0 0.45rem;
		font-family: var(--font-mono);
		font-size: 0.6rem;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		color: var(--ink-soft);
	}

	.stat-value {
		margin: 0;
		font-family: var(--font-display);
		font-size: clamp(1.3rem, 3vw, 1.8rem);
		line-height: 1;
		letter-spacing: -0.015em;
		color: var(--ink);
		font-variant-numeric: tabular-nums;
	}

	.stat-value.small {
		font-size: 1.4rem;
	}

	.stat-note {
		margin: 0.55rem 0 0;
		font-family: var(--font-mono);
		font-size: 0.7rem;
		color: var(--ink-soft);
		font-variant-numeric: tabular-nums;
	}

	/* ─────────────────────────────────────────────────────────
	   FORM CARDS
	   ───────────────────────────────────────────────────────── */
	.form-card-head {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 1rem;
		flex-wrap: wrap;
		margin-bottom: 0.4rem;
	}

	.tiny-eyebrow {
		font-size: 0.6rem;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		margin: 0 0 0.25rem;
	}

	/* Wax seal */
	.seal {
		position: relative;
		width: 4.5rem;
		height: 4.5rem;
		flex-shrink: 0;
		display: grid;
		place-items: center;
		color: var(--paper);
	}

	.seal-inner {
		position: absolute;
		inset: 0.55rem;
		border-radius: 50%;
		background: radial-gradient(ellipse 80% 70% at 30% 30%, var(--rose), var(--oxblood));
		display: grid;
		place-items: center;
		font-family: var(--font-mono);
		font-size: 0.55rem;
		letter-spacing: 0.2em;
		font-weight: 700;
		box-shadow:
			inset 0 0 12px rgba(0, 0, 0, 0.3),
			inset 0 -3px 0 rgba(0, 0, 0, 0.15),
			3px 3px 0 var(--ink);
	}

	.seal-outer {
		position: absolute;
		inset: 0;
		font-family: var(--font-mono);
		font-size: 0.5rem;
		letter-spacing: 0.05em;
		color: transparent;
	}

	/* Quick amount chips */
	.quick-amounts {
		display: flex;
		flex-wrap: wrap;
		gap: 0.45rem;
		align-items: center;
	}

	.quick-amounts .tiny-eyebrow {
		margin: 0 0.25rem 0 0;
	}

	.quick-chip {
		background: var(--paper-deep);
		border: 1px solid var(--rule);
		padding: 0.4rem 0.7rem;
		font-family: var(--font-mono);
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--ink);
		cursor: pointer;
		transition: all 0.15s ease;
		font-variant-numeric: tabular-nums;
	}

	.quick-chip:hover {
		background: var(--gold-soft);
		border-color: var(--gold-deep);
		color: var(--ink);
		transform: translateY(-1px);
	}

	.quick-chip:active {
		transform: translateY(0);
	}

	/* Lux primary button override */
	.lux-primary {
		background: linear-gradient(135deg, var(--ink) 0%, #2a2620 100%);
		box-shadow: 4px 4px 0 var(--gold-deep);
	}

	.lux-primary:hover {
		box-shadow: 6px 6px 0 var(--gold-deep);
	}

	.lux-primary:active {
		box-shadow: 1px 1px 0 var(--gold-deep);
	}

	/* ─────────────────────────────────────────────────────────
	   LEDGER
	   ───────────────────────────────────────────────────────── */
	.ledger-count {
		font-size: 0.7rem;
		letter-spacing: 0.15em;
		color: var(--ink-faint);
		margin: 0;
		text-transform: uppercase;
	}

	.ledger-empty {
		text-align: center;
		padding: 2rem 1rem 1.5rem;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.6rem;
		color: var(--ink-soft);
	}

	.empty-art {
		width: 90px;
		height: 90px;
		color: var(--gold-deep);
		opacity: 0.6;
	}

	.empty-art svg {
		width: 100%;
		height: 100%;
	}

	.empty-headline {
		margin: 0;
		font-family: var(--font-display);
		font-style: italic;
		font-size: 1.4rem;
		color: var(--ink);
	}

	.ledger-list {
		list-style: none;
		margin: 0.75rem 0 0;
		padding: 0;
		display: grid;
		border-top: 1px solid var(--ink);
	}

	.ledger-item {
		display: grid;
		grid-template-columns: 3rem 1fr auto;
		align-items: center;
		gap: 1rem;
		padding: 1rem 0;
		border-bottom: 1px solid var(--rule);
		transition: background 0.18s ease;
		position: relative;
	}

	.ledger-item:hover {
		background: rgba(232, 196, 106, 0.08);
	}

	.ledger-item.editing {
		background: rgba(201, 151, 58, 0.18);
		margin: 0 -1rem;
		padding-left: 1rem;
		padding-right: 1rem;
	}

	.ledger-item.editing::before {
		content: '';
		position: absolute;
		left: 0;
		top: 0;
		bottom: 0;
		width: 3px;
		background: var(--gold-deep);
	}

	.ledger-glyph {
		display: grid;
		place-items: center;
		width: 2.5rem;
		height: 2.5rem;
		border: 1px solid var(--gold-deep);
		border-radius: 50%;
		background: var(--paper-deep);
		color: var(--gold-deep);
	}

	.glyph-symbol {
		font-family: var(--font-display);
		font-style: italic;
		font-size: 1.15rem;
		line-height: 1;
	}

	.ledger-main {
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
	}

	.ledger-row-top {
		display: flex;
		gap: 0.7rem;
		align-items: baseline;
		flex-wrap: wrap;
	}

	.ledger-from {
		font-weight: 600;
		font-size: 0.95rem;
		color: var(--ink);
	}

	.ledger-date {
		font-size: 0.7rem;
		letter-spacing: 0.1em;
		color: var(--ink-faint);
		text-transform: uppercase;
	}

	.ledger-note {
		margin: 0;
		font-family: var(--font-display);
		font-style: italic;
		font-size: 0.92rem;
		color: var(--ink-soft);
		line-height: 1.35;
	}

	.muted-italic {
		color: var(--ink-faint);
	}

	.ledger-badge {
		display: inline-block;
		margin-top: 0.3rem;
		padding: 0.15rem 0.5rem;
		font-family: var(--font-mono);
		font-size: 0.55rem;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		color: var(--gold-deep);
		border: 1px solid var(--gold-deep);
		background: var(--paper-deep);
		width: fit-content;
	}

	.ledger-amount-col {
		text-align: right;
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 0.4rem;
	}

	.ledger-amount {
		margin: 0;
		font-family: var(--font-display);
		font-size: 1.25rem;
		line-height: 1;
		color: var(--forest);
		font-variant-numeric: tabular-nums;
		letter-spacing: -0.015em;
	}

	.ledger-actions {
		display: flex;
		gap: 0.85rem;
	}

	.button-ghost.danger {
		color: var(--oxblood);
		border-bottom-color: var(--oxblood);
	}

	.button-ghost.danger:hover {
		color: var(--ink);
		border-bottom-color: var(--ink);
	}

	.ledger-footer {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		padding: 1rem 0 0.25rem;
		margin-top: 0.5rem;
		border-top: 2px solid var(--ink);
	}

	.ledger-total {
		font-family: var(--font-display);
		font-size: 1.6rem;
		color: var(--ink);
		font-variant-numeric: tabular-nums;
		letter-spacing: -0.02em;
	}

	@media (max-width: 560px) {
		.ledger-item {
			grid-template-columns: 2.5rem 1fr;
			grid-template-rows: auto auto;
			gap: 0.6rem 0.85rem;
		}
		.ledger-amount-col {
			grid-column: 1 / -1;
			text-align: left;
			align-items: flex-start;
			padding-top: 0.35rem;
			border-top: 1px dashed var(--rule);
		}
		.ledger-actions {
			justify-content: flex-end;
			width: 100%;
		}
	}

	/* ─────────────────────────────────────────────────────────
	   QUOTE
	   ───────────────────────────────────────────────────────── */
	.vault-quote {
		text-align: center;
		padding: 2.5rem 1.5rem 2rem;
		border-top: 1px solid var(--ink);
		border-bottom: 1px solid var(--ink);
		background:
			radial-gradient(ellipse 80% 70% at 50% 50%, rgba(232, 196, 106, 0.08), transparent 70%),
			var(--paper);
		position: relative;
	}

	.vault-quote::before,
	.vault-quote::after {
		content: '';
		position: absolute;
		left: 50%;
		width: 40%;
		height: 1px;
		background: var(--gold-deep);
		opacity: 0.4;
		transform: translateX(-50%);
	}

	.vault-quote::before {
		top: -1px;
	}

	.vault-quote::after {
		bottom: -1px;
	}

	.quote-mark {
		margin: 0;
		font-family: var(--font-display);
		font-size: 5rem;
		line-height: 0.5;
		color: var(--gold-deep);
		font-style: italic;
	}

	.quote-body {
		margin: 1rem auto 1.2rem;
		max-width: 36rem;
		font-family: var(--font-display);
		font-style: italic;
		font-size: clamp(1.05rem, 2.6vw, 1.35rem);
		line-height: 1.5;
		color: var(--ink);
	}

	.quote-cite {
		margin: 0;
		font-family: var(--font-mono);
		font-size: 0.65rem;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		color: var(--ink-faint);
	}

	/* ─────────────────────────────────────────────────────────
	   CONFETTI
	   ───────────────────────────────────────────────────────── */
	.confetti-stage {
		position: fixed;
		inset: 0;
		pointer-events: none;
		z-index: 200;
		overflow: hidden;
	}

	.confetti-piece {
		position: absolute;
		top: -20px;
		display: block;
		opacity: 0;
		transform-origin: center;
		animation: confetti-fall 2.5s cubic-bezier(0.22, 1, 0.36, 1) forwards;
		border-radius: 1px;
		box-shadow: 0 0 4px rgba(0, 0, 0, 0.1);
	}

	@keyframes confetti-fall {
		0% {
			opacity: 0;
			transform: translateY(-20px) rotate(0deg);
		}
		10% {
			opacity: 1;
		}
		100% {
			opacity: 0;
			transform: translateY(110vh) rotate(calc(var(--rot) * 5));
		}
	}

	/* ─────────────────────────────────────────────────────────
	   TOAST
	   ───────────────────────────────────────────────────────── */
	.lux-toast {
		position: fixed;
		bottom: 1.5rem;
		left: 50%;
		transform: translateX(-50%);
		z-index: 300;
		display: flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.75rem 1.25rem;
		background: var(--ink);
		color: var(--paper);
		font-family: var(--font-display);
		font-style: italic;
		font-size: 0.95rem;
		border: 1px solid var(--gold-deep);
		box-shadow: 4px 4px 0 var(--gold-deep);
		animation: toast-rise 0.3s cubic-bezier(0.22, 1, 0.36, 1);
		max-width: calc(100vw - 2rem);
	}

	.lux-toast.warn {
		background: var(--oxblood);
		border-color: var(--paper);
		box-shadow: 4px 4px 0 var(--ink);
	}

	.toast-glyph {
		color: var(--gold-soft);
		font-size: 1.1rem;
	}

	.lux-toast.warn .toast-glyph {
		color: var(--paper);
	}

	@keyframes toast-rise {
		from {
			opacity: 0;
			transform: translate(-50%, 12px);
		}
		to {
			opacity: 1;
			transform: translate(-50%, 0);
		}
	}

	/* ─────────────────────────────────────────────────────────
	   RESPONSIVE
	   ───────────────────────────────────────────────────────── */
	@media (max-width: 768px) {
		.vault-hero {
			grid-template-columns: 1fr;
			padding: 1.5rem 1rem;
		}

		.ring-svg {
			max-width: 260px;
		}

		.hero-stats {
			grid-template-columns: 1fr 1fr;
		}
	}

	@media (max-width: 480px) {
		.vault-masthead {
			padding: 1.75rem 0.75rem 1.5rem;
		}

		.hero-stats {
			grid-template-columns: 1fr;
		}

		.seal {
			display: none;
		}
	}

	/* ─────────────────────────────────────────────────────────
	   DARK MODE TWEAKS
	   ───────────────────────────────────────────────────────── */
	:global([data-theme='dark']) .wedding-page {
		--gold: #d4a44a;
		--gold-soft: #e8c46a;
		--gold-deep: #c9973a;
	}

	:global([data-theme='dark']) .vault-couple {
		background: var(--paper-fold);
	}

	:global([data-theme='dark']) .quick-chip:hover {
		color: var(--paper);
	}
</style>
