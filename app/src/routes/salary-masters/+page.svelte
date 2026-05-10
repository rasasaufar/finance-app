<script lang="ts">
	import { onMount } from 'svelte';
	import { api, ApiError } from '$lib/api';
	import { formatRupiah } from '$lib/format';
	import type { DashboardSummary, SalaryMaster } from '$lib/types';

	let loading = $state(true);
	let saving = $state(false);
	let errorMessage = $state('');
	let successMessage = $state('');
	let summary = $state<DashboardSummary | null>(null);
	let salaryMasters = $state<SalaryMaster[]>([]);
	let editingSalaryId = $state<number | null>(null);
	let salaryMonth = $state(new Date().toISOString().slice(0, 7));
	let salaryAmount = $state('');
	let salaryNote = $state('');
	let showDeleteConfirmModal = $state(false);
	let pendingDelete = $state<SalaryMaster | null>(null);

	const todayLabel = new Intl.DateTimeFormat('id-ID', {
		weekday: 'long',
		day: 'numeric',
		month: 'long',
		year: 'numeric'
	}).format(new Date());

	function toSalaryMasterArray(input: unknown): SalaryMaster[] {
		return Array.isArray(input) ? (input as SalaryMaster[]) : [];
	}

	function sortSalaryMasters(items: SalaryMaster[]): SalaryMaster[] {
		return items.sort((a, b) => b.month.localeCompare(a.month));
	}

	function clearMessages(): void {
		errorMessage = '';
		successMessage = '';
	}

	function resetForm(): void {
		editingSalaryId = null;
		salaryMonth = new Date().toISOString().slice(0, 7);
		salaryAmount = '';
		salaryNote = '';
	}

	function formatMonthLabel(monthValue: string): string {
		const parsed = new Date(`${monthValue}-01T00:00:00`);
		if (Number.isNaN(parsed.getTime())) {
			return monthValue;
		}
		return new Intl.DateTimeFormat('id-ID', { month: 'long', year: 'numeric' }).format(parsed);
	}

	function handleEdit(item: SalaryMaster): void {
		editingSalaryId = item.id;
		salaryMonth = item.month;
		salaryAmount = String(item.amount);
		salaryNote = item.note ?? '';
	}

	function openDeleteConfirmModal(item: SalaryMaster): void {
		pendingDelete = item;
		showDeleteConfirmModal = true;
	}

	function closeDeleteConfirmModal(): void {
		if (saving) {
			return;
		}
		showDeleteConfirmModal = false;
		pendingDelete = null;
	}

	function handleDeleteBackdropClick(event: MouseEvent): void {
		if (event.target === event.currentTarget) {
			closeDeleteConfirmModal();
		}
	}

	function handleDeleteBackdropKeydown(event: KeyboardEvent): void {
		if (event.key === 'Escape') {
			closeDeleteConfirmModal();
		}
	}

	async function loadData(): Promise<void> {
		loading = true;
		errorMessage = '';

		try {
			const [mastersResponse, summaryResponse] = await Promise.all([
				api.get<SalaryMaster[]>('/salary-masters'),
				api.get<DashboardSummary>('/dashboard/summary')
			]);
			salaryMasters = sortSalaryMasters(toSalaryMasterArray(mastersResponse));
			summary = summaryResponse;
		} catch (error) {
			if (error instanceof ApiError) {
				errorMessage = error.message;
			} else {
				errorMessage = 'Gagal memuat data master gaji.';
			}
		} finally {
			loading = false;
		}
	}

	async function refreshSummarySafely(): Promise<void> {
		try {
			summary = await api.get<DashboardSummary>('/dashboard/summary');
		} catch {
			// ignored
		}
	}

	async function handleSubmit(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		clearMessages();

		const numericAmount = Number(salaryAmount);
		if (!salaryMonth) {
			errorMessage = 'Bulan gaji wajib diisi.';
			return;
		}
		if (!Number.isFinite(numericAmount) || numericAmount <= 0) {
			errorMessage = 'Nominal gaji harus lebih dari 0.';
			return;
		}

		saving = true;
		try {
			if (editingSalaryId) {
				const updated = await api.put<SalaryMaster>(`/salary-masters/${editingSalaryId}`, {
					month: salaryMonth,
					amount: numericAmount,
					note: salaryNote
				});
				const index = salaryMasters.findIndex((item) => item.id === editingSalaryId);
				if (index >= 0) {
					salaryMasters[index] = updated;
					salaryMasters = sortSalaryMasters(salaryMasters);
				}
				successMessage = 'Master gaji berhasil diperbarui.';
			} else {
				const created = await api.post<SalaryMaster>('/salary-masters', {
					month: salaryMonth,
					amount: numericAmount,
					note: salaryNote
				});
				salaryMasters.unshift(created);
				salaryMasters = sortSalaryMasters(salaryMasters);
				successMessage = 'Master gaji berhasil ditambahkan.';
			}

			resetForm();
			await refreshSummarySafely();
		} catch (error) {
			if (error instanceof ApiError) {
				errorMessage = error.message;
			} else {
				errorMessage = 'Gagal menyimpan master gaji.';
			}
		} finally {
			saving = false;
		}
	}

	async function confirmDelete(): Promise<void> {
		if (!pendingDelete) {
			return;
		}

		const deletingID = pendingDelete.id;
		const deletingMonth = pendingDelete.month;
		closeDeleteConfirmModal();
		clearMessages();
		saving = true;
		try {
			await api.delete<void>(`/salary-masters/${deletingID}`);
			salaryMasters = salaryMasters.filter((item) => item.id !== deletingID);
			successMessage = 'Master gaji berhasil dihapus.';
			if (editingSalaryId === deletingID) {
				resetForm();
			}
			await refreshSummarySafely();
		} catch (error) {
			if (error instanceof ApiError) {
				errorMessage = error.message;
			} else {
				errorMessage = `Gagal menghapus master gaji ${formatMonthLabel(deletingMonth)}.`;
			}
		} finally {
			saving = false;
		}
	}

	onMount(() => {
		loadData();
	});
</script>

<section class="page">
	<header class="page-header">
		<div class="page-header-top">
			<span><span class="issue-mark">§</span> 04 · Arsip Gaji</span>
			<span>{todayLabel}</span>
		</div>
		<div class="page-header-main">
			<div>
				<h1 class="page-title">Master <em>Gaji</em></h1>
				<p class="page-subtitle">
					Arsip pendapatan bulanan sebagai dasar perhitungan saldo otomatis.
				</p>
			</div>
			<button class="button-secondary" type="button" onclick={loadData}>↻ Segarkan</button>
		</div>
	</header>

	{#if errorMessage}
		<p class="error">{errorMessage}</p>
	{/if}
	{#if successMessage}
		<p class="success">{successMessage}</p>
	{/if}

	{#if loading}
		<p class="muted mono">Memuat arsip gaji…</p>
	{:else}
		{#if summary}
			<div class="card-grid">
				<article class="card master">
					<p class="card-title">Gaji · Bulan Ini</p>
					<p class="card-value money">{formatRupiah(summary.salary_current_month)}</p>
				</article>
				<article class="card master">
					<p class="card-title">Akumulasi · Total</p>
					<p class="card-value money">{formatRupiah(summary.salary_total_to_date)}</p>
				</article>
			</div>
		{/if}

		<section class="section-card">
			<h2 class="section-title">{editingSalaryId ? 'Edit Master Gaji' : 'Tambah Master Gaji'}</h2>
			<p class="section-lede">
				Setiap bulan dapat memiliki satu catatan gaji resmi.
			</p>
			<form class="form-grid" onsubmit={handleSubmit}>
				<div class="form-row">
					<label class="field">
						<span>Bulan Gaji</span>
						<input type="month" bind:value={salaryMonth} required />
					</label>
					<label class="field">
						<span>Nominal Gaji (Rp)</span>
						<input
							type="number"
							bind:value={salaryAmount}
							min="1"
							placeholder="Contoh: 7500000"
							required
						/>
					</label>
				</div>
				<label class="field">
					<span>Catatan (opsional)</span>
					<input
						type="text"
						bind:value={salaryNote}
						placeholder="Contoh: gaji pokok + bonus kuartal"
					/>
				</label>
				<div class="button-row">
					<button class="button-primary" type="submit" disabled={saving}>
						{saving
							? 'Menyimpan…'
							: editingSalaryId
								? 'Perbarui Master Gaji'
								: 'Arsipkan Master Gaji'}
					</button>
					{#if editingSalaryId}
						<button class="button-secondary" type="button" onclick={resetForm}>Batal Edit</button>
					{/if}
				</div>
			</form>
		</section>

		<section class="section-card">
			<h2 class="section-title">Arsip Master Gaji</h2>
			<p class="section-lede">
				Disusun dari bulan terbaru ke paling lama.
			</p>
			{#if salaryMasters.length === 0}
				<p class="muted">Belum ada master gaji yang diarsipkan.</p>
			{:else}
				<div class="salary-list">
					{#each salaryMasters as item (item.id)}
						<article class="salary-row" class:is-editing={editingSalaryId === item.id}>
							<div class="salary-month">
								<span class="mono tiny">Bulan</span>
								<span class="salary-month-name">{formatMonthLabel(item.month)}</span>
							</div>
							<div class="salary-mid">
								<p class="salary-amount money-display">{formatRupiah(item.amount)}</p>
								<p class="salary-note">{item.note || 'Tanpa catatan'}</p>
							</div>
							<div class="salary-actions">
								<button class="button-ghost" type="button" onclick={() => handleEdit(item)}>
									Edit
								</button>
								<button
									class="button-ghost danger"
									type="button"
									onclick={() => openDeleteConfirmModal(item)}
									disabled={saving}
								>
									Hapus
								</button>
							</div>
						</article>
					{/each}
				</div>
			{/if}
		</section>
	{/if}
</section>

{#if showDeleteConfirmModal && pendingDelete}
	<div
		class="modal-backdrop"
		role="button"
		tabindex="0"
		aria-label="Tutup popup konfirmasi hapus master gaji"
		onclick={handleDeleteBackdropClick}
		onkeydown={handleDeleteBackdropKeydown}
	>
		<div
			class="modal-card"
			role="dialog"
			aria-modal="true"
			aria-labelledby="confirm-delete-master-gaji-title"
			tabindex="-1"
		>
			<p class="muted mono" style="margin-bottom: 0.35rem;">Konfirmasi · Hapus Arsip</p>
			<h3 id="confirm-delete-master-gaji-title" class="section-title" style="margin-bottom: 1rem;">
				Hapus Master Gaji?
			</h3>
			<div class="confirm-detail-list">
				<div class="confirm-detail-row">
					<span>Bulan</span>
					<strong>{formatMonthLabel(pendingDelete.month)}</strong>
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
				Menghapus master gaji akan mengubah perhitungan saldo saat ini.
			</p>
			<div class="button-row" style="justify-content: flex-end;">
				<button
					class="button-secondary"
					type="button"
					onclick={closeDeleteConfirmModal}
					disabled={saving}
				>
					Batal
				</button>
				<button class="button-danger" type="button" onclick={confirmDelete} disabled={saving}>
					{saving ? 'Menghapus…' : 'Ya, Hapus'}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.tiny {
		font-size: 0.6rem;
		letter-spacing: 0.15em;
		text-transform: uppercase;
	}

	.salary-list {
		display: grid;
		gap: 0;
		border-top: 1px solid var(--rule);
	}

	.salary-row {
		display: grid;
		grid-template-columns: minmax(8rem, auto) 1fr auto;
		align-items: center;
		gap: 1rem;
		padding: 1rem 0;
		border-bottom: 1px solid var(--rule);
	}

	.salary-row.is-editing {
		background: var(--ochre-soft);
		margin: 0 -1rem;
		padding-left: 1rem;
		padding-right: 1rem;
	}

	.salary-month {
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
		color: var(--ink-faint);
	}

	.salary-month-name {
		font-family: var(--font-display);
		font-size: 1.25rem;
		color: var(--ink);
		line-height: 1;
	}

	.salary-mid {
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
		min-width: 0;
	}

	.salary-amount {
		margin: 0;
		font-size: 1.35rem;
		color: var(--ink);
		line-height: 1;
		letter-spacing: -0.015em;
	}

	.salary-note {
		margin: 0;
		font-size: 0.85rem;
		color: var(--ink-soft);
		font-style: italic;
		font-family: var(--font-display);
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.salary-actions {
		display: flex;
		gap: 1rem;
		flex-shrink: 0;
	}

	.button-ghost.danger {
		color: var(--oxblood);
		border-bottom-color: var(--oxblood);
	}

	.button-ghost.danger:hover {
		color: var(--ink);
		border-bottom-color: var(--ink);
	}

	@media (max-width: 560px) {
		.salary-row {
			grid-template-columns: 1fr;
			gap: 0.5rem;
		}

		.salary-actions {
			padding-top: 0.35rem;
			border-top: 1px dashed var(--rule);
			justify-content: flex-end;
		}
	}
</style>
