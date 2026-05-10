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

		<p class="notice">Akun demo: rasas@example.com / password123</p>
	</div>
</section>

<style>
	.login-page {
		min-height: calc(100dvh - 2rem);
		display: grid;
		align-items: center;
	}

	.login-card {
		background: rgba(255, 255, 255, 0.96);
		border: 1px solid var(--line);
		border-radius: 1.2rem;
		padding: 1.1rem;
		max-width: 430px;
		width: 100%;
		margin: 0 auto;
		box-shadow: 0 16px 28px -24px rgba(11, 30, 60, 0.55);
	}

	.login-tag {
		margin: 0;
		color: var(--primary);
		font-size: 0.74rem;
		font-weight: 800;
		letter-spacing: 0.08em;
		text-transform: uppercase;
	}

	.login-title {
		margin: 0.65rem 0 0;
		font-size: 1.42rem;
		font-weight: 800;
	}

	.login-subtitle {
		margin: 0.45rem 0 1rem;
		color: var(--text-secondary);
		font-size: 0.9rem;
	}

	.login-button {
		min-height: 2.9rem;
	}

	@media (min-width: 768px) {
		.login-card {
			padding: 1.4rem;
		}
	}
</style>
