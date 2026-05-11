<script lang="ts">
	import { tick } from 'svelte';

	interface Props {
		value: string; // "YYYY-MM"
		label?: string;
		required?: boolean;
		onchange?: (value: string) => void;
	}

	let { value = $bindable(), label, required = false, onchange }: Props = $props();

	let open = $state(false);
	let triggerEl = $state<HTMLButtonElement | null>(null);
	let panelEl = $state<HTMLDivElement | null>(null);

	let panelStyle = $state('');

	const parsed = $derived.by(() => {
		if (!value) return { year: new Date().getFullYear(), month: new Date().getMonth() };
		const [y, m] = value.split('-').map(Number);
		return { year: y, month: m - 1 };
	});

	let viewYear = $state(new Date().getFullYear());

	const MONTHS_ID = ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des'];
	const MONTHS_FULL_ID = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'];

	function displayValue(): string {
		if (!value) return '—';
		return `${MONTHS_FULL_ID[parsed.month]} ${parsed.year}`;
	}

	async function openPanel() {
		viewYear = parsed.year;
		open = true;
		await tick();
		positionPanel();
	}

	function positionPanel() {
		if (!triggerEl) return;
		const rect = triggerEl.getBoundingClientRect();
		const panelW = 280;
		let left = rect.left + window.scrollX;
		let top = rect.bottom + window.scrollY + 4;
		// Prevent right overflow
		if (rect.left + panelW > window.innerWidth - 8) {
			left = window.innerWidth - panelW - 8 + window.scrollX;
		}
		panelStyle = `position:absolute;top:${top}px;left:${left}px;width:${panelW}px;z-index:9999;`;
	}

	function selectMonth(monthIndex: number) {
		const mm = String(monthIndex + 1).padStart(2, '0');
		value = `${viewYear}-${mm}`;
		onchange?.(value);
		open = false;
	}

	function prevYear() { viewYear -= 1; }
	function nextYear() { viewYear += 1; }

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') open = false;
	}

	function handleOutsideClick(e: MouseEvent) {
		const target = e.target as Node;
		if (!triggerEl?.contains(target) && !panelEl?.contains(target)) {
			open = false;
		}
	}

	// Portal: move panel to document.body to escape overflow clipping
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

<div class="mp-wrap">
	{#if label}
		<span class="mp-label">{label}{required ? ' *' : ''}</span>
	{/if}

	<button
		bind:this={triggerEl}
		type="button"
		class="mp-trigger"
		class:is-open={open}
		onclick={(e) => { e.stopPropagation(); open ? (open = false) : openPanel(); }}
		aria-haspopup="true"
		aria-expanded={open}
	>
		<span class="mp-trigger-value">{displayValue()}</span>
		<span class="mp-trigger-icon" aria-hidden="true">
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
		class="mp-panel"
		style={panelStyle}
		role="dialog"
		aria-label="Pilih bulan"
		onclick={(e) => e.stopPropagation()}
	>
		<div class="mp-year-nav">
			<button type="button" class="mp-nav-btn" onclick={prevYear} aria-label="Tahun sebelumnya">←</button>
			<span class="mp-year-label">{viewYear}</span>
			<button type="button" class="mp-nav-btn" onclick={nextYear} aria-label="Tahun berikutnya">→</button>
		</div>

		<div class="mp-grid" role="grid" aria-label="Pilih bulan">
			{#each MONTHS_ID as name, i}
				{@const isSelected = parsed.year === viewYear && parsed.month === i}
				{@const isCurrentMonth = new Date().getFullYear() === viewYear && new Date().getMonth() === i}
				<button
					type="button"
					class="mp-month-btn"
					class:is-selected={isSelected}
					class:is-current={isCurrentMonth && !isSelected}
					onclick={() => selectMonth(i)}
					aria-pressed={isSelected}
					aria-label={`${MONTHS_FULL_ID[i]} ${viewYear}`}
				>
					{name}
				</button>
			{/each}
		</div>

		<div class="mp-footer">
			<button
				type="button"
				class="mp-footer-btn"
				onclick={() => {
					const now = new Date();
					viewYear = now.getFullYear();
					selectMonth(now.getMonth());
				}}
			>
				Bulan ini
			</button>
			<button type="button" class="mp-footer-btn mp-close" onclick={() => (open = false)}>
				Tutup
			</button>
		</div>
	</div>
{/if}

<style>
	.mp-wrap {
		display: grid;
		gap: 0.35rem;
	}

	.mp-label {
		font-family: var(--font-mono);
		font-size: 0.65rem;
		font-weight: 500;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		color: var(--ink-soft);
	}

	.mp-trigger {
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

	.mp-trigger:hover,
	.mp-trigger.is-open {
		border-bottom-color: var(--oxblood);
	}

	.mp-trigger-value { flex: 1; }

	.mp-trigger-icon {
		color: var(--ink-faint);
		display: flex;
		align-items: center;
		flex-shrink: 0;
		transition: color 0.2s;
	}

	.mp-trigger:hover .mp-trigger-icon,
	.mp-trigger.is-open .mp-trigger-icon {
		color: var(--oxblood);
	}

	/* Panel styles — applied globally since panel is portalled to body */
	:global(.mp-panel) {
		background: var(--paper);
		border: 1.5px solid var(--ink);
		box-shadow: 6px 6px 0 var(--ink);
		animation: mp-drop 0.18s cubic-bezier(0.22, 1, 0.36, 1);
	}

	@keyframes mp-drop {
		from { opacity: 0; transform: translateY(-6px); }
		to { opacity: 1; transform: none; }
	}

	:global(.mp-year-nav) {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0.65rem 0.85rem;
		border-bottom: 1px solid var(--ink);
		background: var(--paper-deep);
	}

	:global(.mp-nav-btn) {
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

	:global(.mp-nav-btn:hover) {
		background: var(--ink);
		color: var(--paper);
		border-color: var(--ink);
	}

	:global(.mp-year-label) {
		font-family: var(--font-mono);
		font-size: 0.85rem;
		font-weight: 600;
		letter-spacing: 0.12em;
		color: var(--ink);
	}

	:global(.mp-grid) {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 0;
		padding: 0.5rem;
	}

	:global(.mp-month-btn) {
		background: transparent;
		border: 1px solid transparent;
		color: var(--ink-soft);
		padding: 0.6rem 0.25rem;
		font-family: var(--font-mono);
		font-size: 0.72rem;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		cursor: pointer;
		text-align: center;
		transition: all 0.12s;
		position: relative;
	}

	:global(.mp-month-btn:hover) {
		background: var(--paper-deep);
		color: var(--ink);
		border-color: var(--rule);
	}

	:global(.mp-month-btn.is-current::after) {
		content: '';
		position: absolute;
		bottom: 4px;
		left: 50%;
		transform: translateX(-50%);
		width: 3px;
		height: 3px;
		background: var(--oxblood);
		border-radius: 50%;
	}

	:global(.mp-month-btn.is-selected) {
		background: var(--ink);
		color: var(--paper);
		border-color: var(--ink);
		font-weight: 600;
	}

	:global(.mp-month-btn.is-selected::after) { display: none; }

	:global(.mp-footer) {
		display: flex;
		justify-content: space-between;
		padding: 0.5rem 0.85rem;
		border-top: 1px dashed var(--rule);
	}

	:global(.mp-footer-btn) {
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

	:global(.mp-footer-btn:hover) {
		color: var(--oxblood);
		border-bottom-color: var(--oxblood);
	}

	:global(.mp-close) { color: var(--ink-faint); }
</style>
