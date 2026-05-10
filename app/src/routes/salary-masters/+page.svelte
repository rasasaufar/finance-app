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
			// Biarkan user tetap lanjut kerja walau ringkasan sementara gagal disegarkan.
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
		<div>
			<h1 class="page-title">Master Saldo Gaji</h1>
			<p class="page-subtitle">Kelola nominal gaji bulanan untuk perhitungan saldo otomatis.</p>
		</div>
		<button class="button-secondary" type="button" onclick={loadData}>Segarkan</button>
	</header>

	{#if errorMessage}
		<p class="error">{errorMessage}</p>
	{/if}
	{#if successMessage}
		<p class="success">{successMessage}</p>
	{/if}

	{#if loading}
		<p class="muted">Memuat data master gaji...</p>
	{:else}
		{#if summary}
			<div class="card-grid">
				<article class="card">
					<p class="card-title">Master Gaji Bulan Ini</p>
					<p class="card-value">{formatRupiah(summary.salary_current_month)}</p>
				</article>
				<article class="card">
					<p class="card-title">Akumulasi Master Gaji</p>
					<p class="card-value">{formatRupiah(summary.salary_total_to_date)}</p>
				</article>
			</div>
		{/if}

		<section class="section-card">
			<h2 class="section-title">{editingSalaryId ? 'Edit Master Gaji' : 'Tambah Master Gaji'}</h2>
			<form class="form-grid" onsubmit={handleSubmit}>
				<div class="form-row">
					<label class="field">
						<span>Bulan Gaji</span>
						<input type="month" bind:value={salaryMonth} required />
					</label>
					<label class="field">
						<span>Nominal Gaji</span>
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
					<input type="text" bind:value={salaryNote} placeholder="Contoh: gaji pokok + bonus" />
				</label>
				<div class="button-row">
					<button class="button-primary" type="submit" disabled={saving}>
						{saving
							? 'Menyimpan...'
							: editingSalaryId
								? 'Perbarui Master Gaji'
								: 'Tambah Master Gaji'}
					</button>
					{#if editingSalaryId}
						<button class="button-secondary" type="button" onclick={resetForm}>
							Batal Edit
						</button>
					{/if}
				</div>
			</form>
		</section>

		<section class="section-card">
			<h2 class="section-title">Daftar Master Gaji</h2>
			<div class="list">
				{#if salaryMasters.length === 0}
					<div class="list-item">
						<p class="muted">Belum ada master gaji.</p>
					</div>
				{:else}
					{#each salaryMasters as item (item.id)}
						<div class="list-item">
							<div class="list-item-header">
								<div>
									<strong>{formatMonthLabel(item.month)}</strong>
									<p class="muted">{item.note || 'Tanpa catatan'}</p>
								</div>
								<p class="card-value">{formatRupiah(item.amount)}</p>
							</div>
							<div class="button-row">
								<button class="button-secondary" type="button" onclick={() => handleEdit(item)}>
									Edit
								</button>
								<button
									class="button-danger"
									type="button"
									onclick={() => openDeleteConfirmModal(item)}
									disabled={saving}
								>
									Hapus
								</button>
							</div>
						</div>
					{/each}
				{/if}
			</div>
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
			<h3 id="confirm-delete-master-gaji-title" class="section-title">Konfirmasi Hapus Master Gaji</h3>
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
					<strong>{pendingDelete.note || '-'}</strong>
				</div>
			</div>
			<p class="muted">Master gaji yang dihapus akan mengubah perhitungan saldo saat ini.</p>
			<div class="button-row">
				<button class="button-secondary" type="button" onclick={closeDeleteConfirmModal} disabled={saving}>
					Batal
				</button>
				<button class="button-danger" type="button" onclick={confirmDelete} disabled={saving}>
					{saving ? 'Menghapus...' : 'Ya, Hapus Master Gaji'}
				</button>
			</div>
		</div>
	</div>
{/if}
