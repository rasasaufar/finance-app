<script lang="ts">
	import { goto } from '$app/navigation';
	import { api, ApiError } from '$lib/api';
	import { setAuthToken } from '$lib/auth';
	import type { LoginResponse } from '$lib/types';

	let email = $state('rasas@example.com');
	let password = $state('password123');
	let loading = $state(false);
	let errorMessage = $state('');

	const todayLabel = new Intl.DateTimeFormat('id-ID', {
		weekday: 'long',
		day: 'numeric',
		month: 'long',
		year: 'numeric'
	}).format(new Date());

	async function handleSubmit(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		errorMessage = '';
		loading = true;

		try {
			const response = await api.post<LoginResponse>(
				'/auth/login',
				{
					email,
					password
				},
				false
			);
			setAuthToken(response.token);
			await goto('/dashboard');
		} catch (error) {
			if (error instanceof ApiError) {
				errorMessage = error.message;
			} else {
				errorMessage = 'Tidak bisa login. Coba lagi.';
			}
		} finally {
			loading = false;
		}
	}
</script>

<section class="cover">
	<div class="cover-plate">
		<!-- corner registration marks -->
		<span class="reg-mark top-left"></span>
		<span class="reg-mark top-right"></span>
		<span class="reg-mark bottom-left"></span>
		<span class="reg-mark bottom-right"></span>

		<div class="cover-head">
			<p class="cover-eyebrow">Buku Kas Pribadi · Edisi Harian</p>
			<div class="cover-rule"></div>
			<p class="cover-date">{todayLabel}</p>
		</div>

		<div class="cover-body">
			<h1 class="cover-title">
				Buku Kas <em>Pribadi</em>
				<span class="cover-and">&amp;</span>
				<span class="cover-sub">Catatan Budget</span>
			</h1>

			<div class="cover-meta">
				<div class="meta-col">
					<p class="meta-label">Jilid</p>
					<p class="meta-value">MMXXVI</p>
				</div>
				<div class="meta-col">
					<p class="meta-label">Mata Uang</p>
					<p class="meta-value">IDR</p>
				</div>
				<div class="meta-col">
					<p class="meta-label">Pemilik</p>
					<p class="meta-value">Tunggal</p>
				</div>
			</div>

			<p class="cover-lede">
				<span class="drop-cap">S</span>etiap rupiah dicatat dengan ketelitian seorang kurator.
				Masuk untuk membuka lembaran baru — catat pemasukan, amati pengeluaran, dan jaga
				batas budget harian serta bulanan.
			</p>
		</div>

		<div class="fleuron">
			<span class="fleuron-mark">❦</span>
		</div>

		<form class="cover-form" onsubmit={handleSubmit}>
			<p class="form-header">Masuk ke Lembar Pencatat</p>

			{#if errorMessage}
				<p class="error">{errorMessage}</p>
			{/if}

			<label class="field">
				<span>Alamat Email</span>
				<input type="email" bind:value={email} placeholder="email@contoh.com" required />
			</label>

			<label class="field">
				<span>Kata Sandi</span>
				<input
					type="password"
					bind:value={password}
					placeholder="••••••••"
					required
				/>
			</label>

			<button class="button-primary cover-submit" type="submit" disabled={loading}>
				{loading ? 'Membuka buku…' : 'Buka Buku Kas'}
			</button>

			<p class="form-fine">
				Akses hanya untuk pemilik. Sistem ini tidak memiliki pendaftaran terbuka.
			</p>
		</form>

		<!-- wax seal -->
		<div class="wax-seal" aria-hidden="true">
			<div class="seal">
				<span>R·S</span>
			</div>
			<p class="wax-caption mono">Est. MMXXVI</p>
		</div>
	</div>

	<div class="cover-foot">
		<span>№ 01 · Jilid Utama</span>
		<span class="foot-rule"></span>
		<span>Dicetak Digital · id-ID</span>
	</div>
</section>

<style>
	.cover {
		min-height: 100dvh;
		display: flex;
		flex-direction: column;
		align-items: stretch;
		justify-content: center;
		padding: 1.25rem;
		gap: 1rem;
		position: relative;
		overflow: hidden;
	}

	.cover::before,
	.cover::after {
		content: '';
		position: absolute;
		font-family: var(--font-display);
		font-style: italic;
		color: var(--ink);
		opacity: 0.04;
		pointer-events: none;
		user-select: none;
		font-size: clamp(12rem, 30vw, 22rem);
		line-height: 0.8;
		letter-spacing: -0.03em;
	}

	.cover::before {
		content: 'Rp';
		top: -2rem;
		left: -2rem;
	}

	.cover::after {
		content: '₨';
		bottom: -4rem;
		right: -3rem;
		transform: rotate(-12deg);
	}

	.cover-plate {
		width: min(100%, 560px);
		margin: 0 auto;
		background: var(--paper);
		border: 1.5px solid var(--ink);
		padding: 1.5rem 1.35rem 2rem;
		position: relative;
		box-shadow: 8px 8px 0 var(--ink);
		animation: plate-in 0.7s cubic-bezier(0.22, 1, 0.36, 1);
	}

	@keyframes plate-in {
		from {
			opacity: 0;
			transform: translateY(12px) rotate(-0.5deg);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}

	.cover-plate::before,
	.cover-plate::after {
		content: '';
		position: absolute;
		background: var(--ink);
	}

	.cover-plate::before {
		top: 10px;
		left: 10px;
		right: 10px;
		height: 1px;
	}

	.cover-plate::after {
		bottom: 10px;
		left: 10px;
		right: 10px;
		height: 1px;
	}

	/* Corner registration marks (like printer marks) */
	.reg-mark {
		position: absolute;
		width: 14px;
		height: 14px;
		border: 1.5px solid var(--ink);
	}

	.reg-mark.top-left {
		top: -1px;
		left: -1px;
		border-right: 0;
		border-bottom: 0;
	}

	.reg-mark.top-right {
		top: -1px;
		right: -1px;
		border-left: 0;
		border-bottom: 0;
	}

	.reg-mark.bottom-left {
		bottom: -1px;
		left: -1px;
		border-right: 0;
		border-top: 0;
	}

	.reg-mark.bottom-right {
		bottom: -1px;
		right: -1px;
		border-left: 0;
		border-top: 0;
	}

	.cover-head {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 0.75rem;
		padding-bottom: 0.9rem;
		border-bottom: 1.5px solid var(--ink);
	}

	.cover-eyebrow {
		margin: 0;
		font-family: var(--font-mono);
		font-size: 0.6rem;
		letter-spacing: 0.22em;
		text-transform: uppercase;
		color: var(--ink);
	}

	.cover-rule {
		flex: 1;
		height: 1px;
		background: var(--ink);
		margin: 0 0.5rem 0.25rem;
		opacity: 0.25;
	}

	.cover-date {
		margin: 0;
		font-family: var(--font-mono);
		font-size: 0.6rem;
		letter-spacing: 0.15em;
		color: var(--ink-soft);
		text-transform: uppercase;
		text-align: right;
	}

	.cover-body {
		padding: 1.5rem 0 0.5rem;
	}

	.cover-title {
		margin: 0;
		font-family: var(--font-display);
		font-weight: 400;
		font-size: clamp(2.75rem, 9vw, 3.75rem);
		line-height: 0.88;
		letter-spacing: -0.025em;
		color: var(--ink);
	}

	.cover-title em {
		font-style: italic;
		color: var(--oxblood);
	}

	.cover-and {
		display: inline-block;
		font-style: italic;
		color: var(--ochre);
		margin: 0 0.1em;
		transform: translateY(-0.05em);
	}

	.cover-sub {
		display: block;
		font-style: italic;
		font-size: 0.6em;
		color: var(--ink-soft);
		margin-top: 0.1em;
	}

	.cover-meta {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 0.5rem;
		margin: 1.25rem 0 1rem;
		padding: 0.75rem 0;
		border-top: 1px solid var(--ink);
		border-bottom: 1px solid var(--ink);
	}

	.meta-label {
		margin: 0;
		font-family: var(--font-mono);
		font-size: 0.55rem;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		color: var(--ink-faint);
	}

	.meta-value {
		margin: 0.2rem 0 0;
		font-family: var(--font-display);
		font-size: 1.1rem;
		color: var(--ink);
		line-height: 1;
	}

	.cover-lede {
		margin: 0;
		font-size: 0.95rem;
		line-height: 1.55;
		color: var(--ink-soft);
	}

	.cover-lede :global(.drop-cap) {
		color: var(--oxblood);
	}

	.cover-form {
		display: grid;
		gap: 1rem;
		padding-top: 0.5rem;
		position: relative;
	}

	.form-header {
		margin: 0 0 0.25rem;
		font-family: var(--font-mono);
		font-size: 0.65rem;
		letter-spacing: 0.2em;
		text-transform: uppercase;
		color: var(--ink);
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.form-header::before,
	.form-header::after {
		content: '';
		flex: 1;
		height: 1px;
		background: var(--ink);
		opacity: 0.35;
	}

	.cover-submit {
		margin-top: 0.5rem;
		padding: 1rem 1.25rem;
		font-size: 1rem;
	}

	.form-fine {
		margin: 0.35rem 0 0;
		font-size: 0.8rem;
		color: var(--ink-faint);
		text-align: center;
		font-style: italic;
		font-family: var(--font-display);
	}

	/* Wax seal */
	.wax-seal {
		position: absolute;
		bottom: 0.85rem;
		right: 0.85rem;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.25rem;
		transform: rotate(-8deg);
		animation: seal-stamp 0.9s cubic-bezier(0.34, 1.56, 0.64, 1) 0.3s backwards;
	}

	@keyframes seal-stamp {
		0% {
			opacity: 0;
			transform: rotate(-8deg) scale(2);
		}
		60% {
			opacity: 1;
			transform: rotate(-8deg) scale(0.92);
		}
		100% {
			opacity: 1;
			transform: rotate(-8deg) scale(1);
		}
	}

	.wax-seal .seal {
		width: 3.25rem;
		height: 3.25rem;
		font-size: 1rem;
		letter-spacing: 0.05em;
		box-shadow: 2px 2px 0 rgba(139, 46, 46, 0.25);
	}

	.wax-caption {
		font-size: 0.55rem;
		color: var(--oxblood);
		letter-spacing: 0.15em;
		text-transform: uppercase;
		margin: 0;
		opacity: 0.75;
	}

	.cover-foot {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		width: min(100%, 560px);
		margin: 0 auto;
		font-family: var(--font-mono);
		font-size: 0.6rem;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		color: var(--ink-soft);
	}

	.foot-rule {
		flex: 1;
		height: 1px;
		background: var(--ink);
		opacity: 0.3;
	}

	@media (min-width: 768px) {
		.cover-plate {
			padding: 2.25rem 2.25rem 2rem;
		}

		.wax-seal {
			bottom: 1.35rem;
			right: 1.35rem;
		}

		.wax-seal .seal {
			width: 4rem;
			height: 4rem;
			font-size: 1.2rem;
		}
	}
</style>
