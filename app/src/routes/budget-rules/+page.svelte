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

	const todayLabel = new Intl.DateTimeFormat('id-ID', {
		weekday: 'long',
		day: 'numeric',
		month: 'long',
		year: 'numeric'
	}).format(new Date());

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
		<div class="page-header-top">
			<span><span class="issue-mark">§</span> 03 · Peraturan Kantong</span>
			<span>{todayLabel}</span>
		</div>
		<div class="page-header-main">
			<div>
				<h1 class="page-title">Aturan <em>Budget</em></h1>
				<p class="page-subtitle">
					Tetapkan pagar pembatas pengeluaran — harian, mingguan, atau bulanan.
				</p>
			</div>
		</div>
	</header>

	<section class="section-card">
		<h2 class="section-title">
			{editingId ? 'Edit Aturan' : 'Tambah Aturan Baru'}
		</h2>
		<p class="section-lede">
			Setiap kategori boleh memiliki beberapa periode budget yang berbeda.
		</p>

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
				<span>Limit Budget (Rp)</span>
				<input type="number" bind:value={limit} min="1" placeholder="Contoh: 60000" required />
			</label>

			<div class="button-row">
				<button class="button-primary" type="submit" disabled={saving}>
					{saving ? 'Menyimpan…' : editingId ? 'Perbarui Aturan' : 'Tambah Aturan'}
				</button>
				{#if editingId}
					<button class="button-secondary" type="button" onclick={handleCancel}>Batal Edit</button>
				{/if}
			</div>
		</form>
	</section>

	<section class="section-card">
		<h2 class="section-title">Daftar Aturan</h2>
		<p class="section-lede">
			Total {rules.length} aturan tercatat saat ini.
		</p>
		{#if loading}
			<p class="muted mono">Memuat aturan budget…</p>
		{:else if rules.length === 0}
			<p class="muted">Belum ada aturan budget yang dibuat.</p>
		{:else}
			<div class="rule-list">
				{#each rules as rule, i}
					<article class="rule-row" class:is-editing={editingId === rule.id}>
						<div class="rule-index">
							<span class="mono tiny">№</span>
							<span class="rule-num">{String(i + 1).padStart(2, '0')}</span>
						</div>
						<div class="rule-body">
							<div class="rule-head">
								<h3 class="rule-name">{rule.category}</h3>
								<span class="badge rule-period">{formatPeriod(rule.period)}</span>
							</div>
							<p class="rule-amount money-display">{formatRupiah(rule.limit)}</p>
						</div>
						<div class="rule-actions">
							<button
								class="button-ghost"
								type="button"
								onclick={() => handleEdit(rule)}
							>
								Edit
							</button>
							<button
								class="button-ghost danger"
								type="button"
								onclick={() => handleDelete(rule.id)}
							>
								Hapus
							</button>
						</div>
					</article>
				{/each}
			</div>
		{/if}
	</section>
</section>

<style>
	.tiny {
		font-size: 0.6rem;
		letter-spacing: 0.15em;
		text-transform: uppercase;
	}

	.badge.rule-period {
		color: var(--indigo);
		background: var(--indigo-soft);
	}

	.rule-list {
		display: grid;
		gap: 0;
		border-top: 1px solid var(--rule);
	}

	.rule-row {
		display: grid;
		grid-template-columns: auto 1fr auto;
		align-items: center;
		gap: 1rem;
		padding: 1rem 0;
		border-bottom: 1px solid var(--rule);
		transition: background 0.2s;
	}

	.rule-row.is-editing {
		background: var(--ochre-soft);
		margin: 0 -1rem;
		padding: 1rem;
		border-bottom-color: var(--ochre);
	}

	.rule-index {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.1rem;
		color: var(--ink-faint);
	}

	.rule-num {
		font-family: var(--font-display);
		font-size: 1.5rem;
		line-height: 1;
		color: var(--ink);
	}

	.rule-body {
		min-width: 0;
	}

	.rule-head {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		flex-wrap: wrap;
		margin-bottom: 0.25rem;
	}

	.rule-name {
		margin: 0;
		font-family: var(--font-display);
		font-size: 1.35rem;
		line-height: 1;
		color: var(--ink);
	}

	.rule-amount {
		margin: 0;
		font-size: 1.1rem;
		color: var(--ink);
		font-variant-numeric: tabular-nums;
		letter-spacing: -0.01em;
	}

	.rule-actions {
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

	@media (max-width: 520px) {
		.rule-row {
			grid-template-columns: auto 1fr;
		}

		.rule-actions {
			grid-column: 1 / -1;
			justify-content: flex-end;
			padding-top: 0.35rem;
			border-top: 1px dashed var(--rule);
		}
	}
</style>
