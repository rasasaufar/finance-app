<script lang="ts">
	import { goto } from '$app/navigation';
	import { api, ApiError } from '$lib/api';
	import { setAuthToken } from '$lib/auth';
	import type { LoginResponse } from '$lib/types';

	let email = $state('rasas@example.com');
	let password = $state('password123');
	let loading = $state(false);
	let errorMessage = $state('');

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

<section class="login-page">
	<div class="login-card">
		<p class="login-tag">Aplikasi Keuangan Pribadi</p>
		<h1 class="login-title">Masuk ke Dompet Pribadi</h1>
		<p class="login-subtitle">
			Pantau transaksi harian, budget, dan laporan bulanan dalam satu layar.
		</p>

		{#if errorMessage}
			<p class="error">{errorMessage}</p>
		{/if}

		<form class="form-grid" onsubmit={handleSubmit}>
			<label class="field">
				<span>Email</span>
				<input type="email" bind:value={email} placeholder="email@contoh.com" required />
			</label>

			<label class="field">
				<span>Password</span>
				<input type="password" bind:value={password} placeholder="Masukkan password" required />
			</label>

			<button class="button-primary login-button" type="submit" disabled={loading}>
				{loading ? 'Memproses...' : 'Masuk'}
			</button>
		</form>

	</div>
</section>

<style>
	.login-page {
		min-height: calc(100dvh - 2rem);
		display: grid;
		align-items: center;
		justify-content: center;
		padding: 1rem;
	}

	.login-card {
		background: var(--surface);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		border: 1px solid var(--line);
		border-radius: 1.5rem;
		padding: 1.75rem;
		max-width: 420px;
		width: 100%;
		box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
		position: relative;
		overflow: hidden;
	}

	.login-card::before {
		content: '';
		position: absolute;
		top: 0; left: 0; right: 0;
		height: 1px;
		background: linear-gradient(90deg, transparent, rgba(255,255,255,0.15), transparent);
	}

	.login-tag {
		margin: 0 0 1rem;
		color: #93c5fd;
		background: rgba(59, 130, 246, 0.15);
		display: inline-block;
		padding: 0.35rem 0.75rem;
		border-radius: 2rem;
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.05em;
		text-transform: uppercase;
		border: 1px solid rgba(59, 130, 246, 0.2);
	}

	.login-title {
		margin: 0;
		font-size: 1.75rem;
		font-weight: 800;
		letter-spacing: -0.02em;
		background: linear-gradient(to right, #fff, #a1a1aa);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
	}

	.login-subtitle {
		margin: 0.5rem 0 1.5rem;
		color: var(--text-secondary);
		font-size: 0.95rem;
		line-height: 1.5;
	}

	.login-button {
		min-height: 3.2rem;
		margin-top: 0.5rem;
		font-size: 1rem;
	}

	@media (min-width: 768px) {
		.login-card {
			padding: 2.5rem;
		}
	}
</style>
