<script lang="ts">
	import { tick } from 'svelte';

	interface Props {
		value: string; // "YYYY-MM-DD"
		label?: string;
		required?: boolean;
		onchange?: (value: string) => void;
	}

	let { value = $bindable(), label, required = false, onchange }: Props = $props();

	let open = $state(false);
	let triggerEl = $state<HTMLButtonElement | null>(null);
	let panelEl = $state<HTMLDivElement | null>(null);

	let panelStyle = $state('');

	const DAYS_ID = ['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab'];
	const MONTHS_FULL_ID = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'];

	const parsed = $derived.by(() => {
		if (!value) {
			const now = new Date();
			return { year: now.getFullYear(), month: now.getMonth(), day: now.getDate() };
		}
		const [y, m, d] = value.split('-').map(Number);
		return { year: y, month: m - 1, day: d };
	});

	let viewYear = $state(new Date().getFullYear());
	let viewMonth = $state(new Date().getMonth());

	function displayValue(): string {
		if (!value) return '—';
		const d = new Date(`${value}T00:00:00`);
		if (isNaN(d.getTime())) return value;
		return new Intl.DateTimeFormat('id-ID', { day: 'numeric', month: 'long', year: 'numeric' }).format(d);
	}

	async function openPanel() {
		viewYear = parsed.year;
		viewMonth = parsed.month;
		open = true;
		await tick();
		positionPanel();
	}

	function positionPanel() {
		if (!triggerEl) return;
		const rect = triggerEl.getBoundingClientRect();
		const panelW = 300;
		let left = rect.left + window.scrollX;
		let top = rect.bottom + window.scrollY + 4;
		if (rect.left + panelW > window.innerWidth - 8) {
			left = window.innerWidth - panelW - 8 + window.scrollX;
		}
		panelStyle = `position:absolute;top:${top}px;left:${left}px;width:${panelW}px;z-index:9999;`;
	}

	const calendarDays = $derived.by(() => {
		const firstDay = new Date(viewYear, viewMonth, 1).getDay();
		const daysInMonth = new Date(viewYear, viewMonth + 1, 0).getDate();
		const daysInPrev = new Date(viewYear, viewMonth, 0).getDate();
		const cells: { day: number; type: 'prev' | 'current' | 'next' }[] = [];

		for (let i = firstDay - 1; i >= 0; i--) cells.push({ day: daysInPrev - i, type: 'prev' });
		for (let d = 1; d <= daysInMonth; d++) cells.push({ day: d, type: 'current' });
		const remaining = 42 - cells.length;
		for (let d = 1; d <= remaining; d++) cells.push({ day: d, type: 'next' });
		return cells;
	});

	function isSelected(day: number, type: string): boolean {
		if (type !== 'current') return false;
		return parsed.year === viewYear && parsed.month === viewMonth && parsed.day === day;
	}

	function isToday(day: number, type: string): boolean {
		if (type !== 'current') return false;
		const now = new Date();
		return now.getFullYear() === viewYear && now.getMonth() === viewMonth && now.getDate() === day;
	}

	function selectDay(day: number, type: 'prev' | 'current' | 'next') {
		let y = viewYear, m = viewMonth;
		if (type === 'prev') { m--; if (m < 0) { m = 11; y--; } }
		else if (type === 'next') { m++; if (m > 11) { m = 0; y++; } }
		const mm = String(m + 1).padStart(2, '0');
		const dd = String(day).padStart(2, '0');
		value = `${y}-${mm}-${dd}`;
		onchange?.(value);
		open = false;
	}

	function prevMonth() { viewMonth--; if (viewMonth < 0) { viewMonth = 11; viewYear--; } }
	function nextMonth() { viewMonth++; if (viewMonth > 11) { viewMonth = 0; viewYear++; } }

	function goToday() {
		const now = new Date();
		const mm = String(now.getMonth() + 1).padStart(2, '0');
		const dd = String(now.getDate()).padStart(2, '0');
		value = `${now.getFullYear()}-${mm}-${dd}`;
		onchange?.(value);
		open = false;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') open = false;
	}

	function handleOutsideClick(e: MouseEvent) {
		const target = e.target as Node;
		if (!triggerEl?.contains(target) && !panelEl?.contains(target)) {
			open = false;
		}
	}

	// Portal: move panel to document.body
	$effect(() => {
		if (open && panelEl) {
			document.body.appendChild(panelEl);
			return () => {
				if (panelEl && panelEl.parentNode === document.body) {
					document.body.removeChild(panelEl);
				}
			};
		}
	});
</script>

<svelte:window onkeydown={handleKeydown} onclick={handleOutsideClick} />

<div class="dp-wrap">
	{#if label}
		<span class="dp-label">{label}{required ? ' *' : ''}</span>
	{/if}

	<button
		bind:this={triggerEl}
		type="button"
		class="dp-trigger"
		class:is-open={open}
		onclick={(e) => { e.stopPropagation(); open ? (open = false) : openPanel(); }}
		aria-haspopup="true"
		aria-expanded={open}
	>
		<span class="dp-trigger-value">{displayValue()}</span>
		<span class="dp-trigger-icon" aria-hidden="true">
			<svg width="12" height="12" viewBox="0 0 12 12" fill="none">
				<rect x="1" y="2" width="10" height="9" rx="0" stroke="currentColor" stroke-width="1.2"/>
				<line x1="1" y1="5" x2="11" y2="5" stroke="currentColor" stroke-width="1.2"/>
				<line x1="4" y1="0.5" x2="4" y2="3.5" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/>
				<line x1="8" y1="0.5" x2="8" y2="3.5" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/>
			</svg>
		</span>
	</button>
</div>

{#if open}
	<div
		bind:this={panelEl}
		class="dp-panel"
		style={panelStyle}
		role="dialog"
		aria-label="Pilih tanggal"
		onclick={(e) => e.stopPropagation()}
	>
		<div class="dp-nav">
			<button type="button" class="dp-nav-btn" onclick={prevMonth} aria-label="Bulan sebelumnya">←</button>
			<span class="dp-nav-label">{MONTHS_FULL_ID[viewMonth]} {viewYear}</span>
			<button type="button" class="dp-nav-btn" onclick={nextMonth} aria-label="Bulan berikutnya">→</button>
		</div>

		<div class="dp-day-headers">
			{#each DAYS_ID as d}
				<span>{d}</span>
			{/each}
		</div>

		<div class="dp-grid" role="grid">
			{#each calendarDays as cell}
				<button
					type="button"
					class="dp-day"
					class:is-other={cell.type !== 'current'}
					class:is-selected={isSelected(cell.day, cell.type)}
					class:is-today={isToday(cell.day, cell.type)}
					onclick={() => selectDay(cell.day, cell.type)}
					aria-pressed={isSelected(cell.day, cell.type)}
				>
					{cell.day}
				</button>
			{/each}
		</div>

		<div class="dp-footer">
			<button type="button" class="dp-footer-btn" onclick={goToday}>Hari ini</button>
			<button type="button" class="dp-footer-btn dp-close" onclick={() => (open = false)}>Tutup</button>
		</div>
	</div>
{/if}

<style>
	.dp-wrap {
		display: grid;
		gap: 0.35rem;
	}

	.dp-label {
		font-family: var(--font-mono);
		font-size: 0.65rem;
		font-weight: 500;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		color: var(--ink-soft);
	}

	.dp-trigger {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem;
		width: 100%;
		background: transparent;
		border: 0;
		border-bottom: 1.5px solid var(--rule);
		border-radius: 0;
		padding: 0.55rem 0.1rem 0.5rem;
		color: var(--ink);
		font-family: var(--font-mono);
		font-size: 0.95rem;
		min-height: 2.5rem;
		cursor: pointer;
		text-align: left;
		transition: border-color 0.2s;
	}

	.dp-trigger:hover,
	.dp-trigger.is-open {
		border-bottom-color: var(--oxblood);
	}

	.dp-trigger-value { flex: 1; }

	.dp-trigger-icon {
		color: var(--ink-faint);
		display: flex;
		align-items: center;
		flex-shrink: 0;
		transition: color 0.2s;
	}

	.dp-trigger:hover .dp-trigger-icon,
	.dp-trigger.is-open .dp-trigger-icon {
		color: var(--oxblood);
	}

	:global(.dp-panel) {
		background: var(--paper);
		border: 1.5px solid var(--ink);
		box-shadow: 6px 6px 0 var(--ink);
		animation: dp-drop 0.18s cubic-bezier(0.22, 1, 0.36, 1);
	}

	@keyframes dp-drop {
		from { opacity: 0; transform: translateY(-6px); }
		to { opacity: 1; transform: none; }
	}

	:global(.dp-nav) {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.65rem 0.85rem;
		border-bottom: 1px solid var(--ink);
		background: var(--paper-deep);
	}

	:global(.dp-nav-btn) {
		background: transparent;
		border: 1px solid var(--rule);
		color: var(--ink);
		width: 1.75rem;
		height: 1.75rem;
		display: grid;
		place-items: center;
		cursor: pointer;
		font-family: var(--font-display);
		font-size: 1rem;
		line-height: 1;
		transition: all 0.12s;
	}

	:global(.dp-nav-btn:hover) {
		background: var(--ink);
		color: var(--paper);
		border-color: var(--ink);
	}

	:global(.dp-nav-label) {
		font-family: var(--font-mono);
		font-size: 0.78rem;
		font-weight: 600;
		letter-spacing: 0.08em;
		color: var(--ink);
		text-transform: uppercase;
	}

	:global(.dp-day-headers) {
		display: grid;
		grid-template-columns: repeat(7, 1fr);
		padding: 0.4rem 0.5rem 0.2rem;
		border-bottom: 1px dashed var(--rule);
	}

	:global(.dp-day-headers span) {
		font-family: var(--font-mono);
		font-size: 0.6rem;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--ink-faint);
		text-align: center;
	}

	:global(.dp-grid) {
		display: grid;
		grid-template-columns: repeat(7, 1fr);
		padding: 0.35rem 0.5rem 0.5rem;
		gap: 1px;
	}

	:global(.dp-day) {
		background: transparent;
		border: 1px solid transparent;
		color: var(--ink);
		padding: 0.4rem 0.1rem;
		font-family: var(--font-mono);
		font-size: 0.75rem;
		font-variant-numeric: tabular-nums;
		cursor: pointer;
		text-align: center;
		line-height: 1;
		transition: all 0.1s;
		position: relative;
		aspect-ratio: 1;
		display: grid;
		place-items: center;
	}

	:global(.dp-day:hover:not(.is-selected)) {
		background: var(--paper-deep);
		border-color: var(--rule);
	}

	:global(.dp-day.is-other) { color: var(--ink-faint); }

	:global(.dp-day.is-today:not(.is-selected)::after) {
		content: '';
		position: absolute;
		bottom: 3px;
		left: 50%;
		transform: translateX(-50%);
		width: 3px;
		height: 3px;
		background: var(--oxblood);
		border-radius: 50%;
	}

	:global(.dp-day.is-selected) {
		background: var(--ink);
		color: var(--paper);
		border-color: var(--ink);
		font-weight: 700;
	}

	:global(.dp-footer) {
		display: flex;
		justify-content: space-between;
		padding: 0.5rem 0.85rem;
		border-top: 1px dashed var(--rule);
	}

	:global(.dp-footer-btn) {
		background: transparent;
		border: 0;
		padding: 0.3rem 0;
		font-family: var(--font-display);
		font-style: italic;
		font-size: 0.88rem;
		color: var(--ink-soft);
		cursor: pointer;
		border-bottom: 1px dotted transparent;
		transition: all 0.15s;
	}

	:global(.dp-footer-btn:hover) {
		color: var(--oxblood);
		border-bottom-color: var(--oxblood);
	}

	:global(.dp-close) { color: var(--ink-faint); }
</style>
