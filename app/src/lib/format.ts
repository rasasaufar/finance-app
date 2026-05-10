const idrFormatter = new Intl.NumberFormat('id-ID', {
	style: 'currency',
	currency: 'IDR',
	maximumFractionDigits: 0
});

export function formatRupiah(value: number): string {
	return idrFormatter.format(Number.isFinite(value) ? value : 0);
}

export function formatPeriod(period: string): string {
	if (period === 'daily') return 'Harian';
	if (period === 'weekly') return 'Mingguan';
	if (period === 'monthly') return 'Bulanan';
	return period;
}

/**
 * Deterministic color for a category name — uses a small curated palette
 * picked to harmonize with the paper/ink ledger aesthetic.
 */
const CATEGORY_PALETTE = [
	'#8b2e2e', // oxblood
	'#2f5d46', // forest
	'#a9812c', // ochre
	'#3a4a7a', // indigo
	'#6b4a8a', // plum
	'#7a3b2e', // sienna
	'#2c6470', // teal
	'#8a5a2c', // bronze
	'#5c4a2a', // olive
	'#983f5c' // rose-ink
];

export function categoryColor(name: string): string {
	if (!name) return CATEGORY_PALETTE[0];
	let hash = 0;
	for (let i = 0; i < name.length; i++) {
		hash = (hash * 31 + name.charCodeAt(i)) >>> 0;
	}
	return CATEGORY_PALETTE[hash % CATEGORY_PALETTE.length];
}
