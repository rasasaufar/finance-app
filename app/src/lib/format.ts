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
