<script lang="ts">
	import { onMount } from 'svelte';
	import { api, ApiError } from '$lib/api';
	import { formatPeriod, formatRupiah } from '$lib/format';
	import type { BudgetPeriod, BudgetRule, Category } from '$lib/types';

	let loading = $state(true);
	let saving = $state(false);
	let errorMessage = $state('');
	let successMessage = $state('');
	let rules = $state<BudgetRule[]>([]);
	let categories = $state<Category[]>([]);

	let editingId = $state<number | null>(null);
	let category = $state('');
	let period = $state<BudgetPeriod>('daily');
	let limit = $state('');

	function resetForm(): void {
		editingId = null;
		period = 'daily';
		limit = '';
		if (categories.length > 0) {
			category = categories[0].name;
		}
	}

	function clearMessages(): void {
		errorMessage = '';
		successMessage = '';
	}

	async function loadData(): Promise<void> {
		loading = true;
		clearMessages();

		try {
			const [rulesResponse, categoriesResponse] = await Promise.all([
				api.get<BudgetRule[]>('/budget-rules'),
				api.get<Category[]>('/categories')
			]);
			rules = rulesResponse;
			categories = categoriesResponse;
			if (!category && categoriesResponse.length > 0) {
				category = categoriesResponse[0].name;
			}
		} catch (error) {
			if (error instanceof ApiError) {
				errorMessage = error.message;
			} else {
				errorMessage = 'Gagal memuat aturan budget.';
			}
		} finally {
			loading = false;
		}
	}

	function handleEdit(rule: BudgetRule): void {
		editingId = rule.id;
		category = rule.category;
		period = rule.period;
		limit = String(rule.limit);
	}

	function handleCancel(): void {
		resetForm();
	}

	async function handleSubmit(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		clearMessages();

		const numericLimit = Number(limit);
		if (!Number.isFinite(numericLimit) || numericLimit <= 0) {
			errorMessage = 'Limit harus lebih dari 0.';
			return;
		}
		if (!category) {
			errorMessage = 'Pilih kategori.';
			return;
		}

		saving = true;
		try {
			if (editingId) {
				await api.put<BudgetRule>(`/budget-rules/${editingId}`, {
					category,
					period,
					limit: numericLimit
				});
				successMessage = 'Aturan budget berhasil diperbarui.';
			} else {
				await api.post<BudgetRule>('/budget-rules', {
					category,
					period,
					limit: numericLimit
				});
				successMessage = 'Aturan budget berhasil ditambahkan.';
			}

			await loadData();
			resetForm();
		} catch (error) {
			if (error instanceof ApiError) {
				errorMessage = error.message;
			} else {
				errorMessage = 'Gagal menyimpan aturan budget.';
			}
		} finally {
			saving = false;
		}
	}

	async function handleDelete(id: number): Promise<void> {
		clearMessages();
		try {
			await api.delete<void>(`/budget-rules/${id}`);
			successMessage = 'Aturan budget berhasil dihapus.';
			await loadData();
			if (editingId === id) {
				resetForm();
			}
		} catch (error) {
			if (error instanceof ApiError) {
				errorMessage = error.message;
			} else {
				errorMessage = 'Gagal menghapus aturan budget.';
			}
		}
	}

	onMount(async () => {
		await loadData();
		resetForm();
	});
</script>

<section class="page">
	<header class="page-header">
		<div>
			<h1 class="page-title">Budget</h1>
			<p class="page-subtitle">Atur batas pengeluaran harian, mingguan, dan bulanan.</p>
		</div>
	</header>

	<section class="section-card">
		<h2 class="section-title">{editingId ? 'Edit Aturan Budget' : 'Tambah Aturan Budget'}</h2>

		{#if errorMessage}
			<p class="error">{errorMessage}</p>
		{/if}
		{#if successMessage}
			<p class="success">{successMessage}</p>
		{/if}

		<form class="form-grid" onsubmit={handleSubmit}>
			<div class="form-row">
				<label class="field">
					<span>Kategori</span>
					<select bind:value={category}>
						<option value="" disabled>Pilih kategori</option>
						{#each categories as item}
							<option value={item.name}>{item.name}</option>
						{/each}
					</select>
				</label>
				<label class="field">
					<span>Periode</span>
					<select bind:value={period}>
						<option value="daily">Harian</option>
						<option value="weekly">Mingguan</option>
						<option value="monthly">Bulanan</option>
					</select>
				</label>
			</div>

			<label class="field">
				<span>Limit Budget</span>
				<input type="number" bind:value={limit} min="1" placeholder="Contoh: 60000" required />
			</label>

			<div class="button-row">
				<button class="button-primary" type="submit" disabled={saving}>
					{saving ? 'Menyimpan...' : editingId ? 'Perbarui Budget' : 'Tambah Budget'}
				</button>
				{#if editingId}
					<button class="button-secondary" type="button" onclick={handleCancel}>Batal Edit</button>
				{/if}
			</div>
		</form>
	</section>

	<section class="section-card">
		<h2 class="section-title">Daftar Aturan Budget</h2>
		{#if loading}
			<p class="muted">Memuat aturan budget...</p>
		{:else if rules.length === 0}
			<p class="muted">Belum ada aturan budget.</p>
		{:else}
			<div class="list">
				{#each rules as rule}
					<article class="list-item">
						<div class="list-item-header">
							<div>
								<strong>{rule.category}</strong>
								<p class="muted">{formatPeriod(rule.period)}</p>
							</div>
							<p class="card-value">{formatRupiah(rule.limit)}</p>
						</div>
						<div class="button-row">
							<button class="button-secondary" type="button" onclick={() => handleEdit(rule)}
								>Edit</button
							>
							<button class="button-danger" type="button" onclick={() => handleDelete(rule.id)}
								>Hapus</button
							>
						</div>
					</article>
				{/each}
			</div>
		{/if}
	</section>
</section>
