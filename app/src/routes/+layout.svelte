<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { clearAuthToken } from '$lib/auth';
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { User } from '$lib/types';

	let { children, data } = $props<{
		children: import('svelte').Snippet;
		data: { isLoggedIn: boolean };
	}>();

	const navItems = [
		{
			href: '/dashboard',
			label: 'Dashboard',
			shortLabel: 'Dash',
			num: '01',
			icon: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path><polyline points="9 22 9 12 15 12 15 22"></polyline></svg>'
		},
		{
			href: '/transactions',
			label: 'Transaksi',
			shortLabel: 'Trans',
			num: '02',
			icon: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M7 17l10-10M17 7H9M17 7v8"></path><path d="M17 17H7M7 17V9"></path></svg>'
		},
		{
			href: '/budget-rules',
			label: 'Budget',
			shortLabel: 'Budget',
			num: '03',
			icon: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="9"></circle><path d="M12 3v9l6 3"></path></svg>'
		},
		{
			href: '/salary-masters',
			label: 'Master Gaji',
			shortLabel: 'Gaji',
			num: '04',
			icon: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="6" width="20" height="12" rx="1"></rect><circle cx="12" cy="12" r="2"></circle><path d="M6 10v4M18 10v4"></path></svg>'
		},
		{
			href: '/categories',
			label: 'Kategori',
			shortLabel: 'Kat',
			num: '05',
			icon: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M4 6h16M4 12h16M4 18h10"></path></svg>'
		},
		{
			href: '/reports',
			label: 'Laporan',
			shortLabel: 'Lap',
			num: '06',
			icon: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M3 3v18h18"></path><path d="M7 14l4-4 3 3 5-6"></path></svg>'
		}
	];

	const currentPath = $derived(page.url.pathname);
	const todayLabel = new Intl.DateTimeFormat('id-ID', {
		weekday: 'long',
		day: 'numeric',
		month: 'long',
		year: 'numeric'
	}).format(new Date());

	const issueNumber = $derived(
		new Intl.DateTimeFormat('id-ID', { year: 'numeric', month: '2-digit' })
			.format(new Date())
			.replace('/', '.')
	);

	function isActive(path: string): boolean {
		return currentPath === path;
	}

	let profileOpen = $state(false);
	let settingsOpen = $state(false);

	let username = $state('Rasa Saufar');
	let userEmail = $state('rasas@example.com');

	let editUsername = $state('');
	let editEmail = $state('');
	let editPassword = $state('');

	let userInitials = $derived(
		username
			.split(' ')
			.map((n) => n[0])
			.join('')
			.substring(0, 2)
			.toUpperCase() || 'U'
	);

	let saveError = $state('');
	let saving = $state(false);

	async function handleSaveSettings(e: Event): Promise<void> {
		e.preventDefault();
		saveError = '';
		saving = true;

		try {
			const updated = await api.put<User>('/me', {
				name: editUsername,
				email: editEmail,
				password: editPassword || undefined
			});
			username = updated.name;
			userEmail = updated.email;
			editPassword = '';
			settingsOpen = false;
		} catch (err: unknown) {
			if (err instanceof Error) {
				saveError = err.message;
			} else {
				saveError = 'Gagal menyimpan perubahan.';
			}
		} finally {
			saving = false;
		}
	}

	async function loadProfile(): Promise<void> {
		if (!data.isLoggedIn) {
			return;
		}

		try {
			const me = await api.get<User>('/me');
			username = me.name;
			userEmail = me.email;
			editUsername = me.name;
			editEmail = me.email;
		} catch {
			// Biarkan UX tetap lanjut walau profile gagal dimuat sementara.
		}
	}

	async function handleLogout(): Promise<void> {
		clearAuthToken();
		await goto('/login');
	}

	onMount(() => {
		loadProfile();
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>Buku Kas Pribadi — Dompet</title>
	<meta name="viewport" content="width=device-width, initial-scale=1" />
</svelte:head>

<div class="app-shell">
	{#if data.isLoggedIn}
		<aside class="sidebar">
			<div class="sidebar-masthead">
				<p class="masthead-eyebrow">Buku Kas · Edisi {issueNumber}</p>
				<h1 class="masthead-title">Dompet <em>Pribadi</em></h1>
				<div class="masthead-meta">
					<span>{todayLabel}</span>
					<span>№ {issueNumber}</span>
				</div>
			</div>

			<nav class="sidebar-nav">
				{#each navItems as item}
					<a class:active={isActive(item.href)} href={item.href}>
						<span class="nav-num">{item.num}</span>
						{@html item.icon}
						<span>{item.label}</span>
						<span class="nav-arrow">→</span>
					</a>
				{/each}
			</nav>

			<div class="sidebar-bottom">
				<div class="profile-widget">
					<button
						class="profile-button"
						type="button"
						onclick={() => (profileOpen = !profileOpen)}
					>
						<div class="profile-avatar">{userInitials}</div>
						<div class="profile-info">
							<p class="profile-name">{username}</p>
							<p class="profile-role">Pemilik Buku Kas</p>
						</div>
					</button>

					{#if profileOpen}
						<div class="profile-popup">
							<div class="popup-header">
								<p class="popup-email">{userEmail}</p>
							</div>
							<button
								class="popup-item"
								type="button"
								onclick={() => {
									editUsername = username;
									editEmail = userEmail;
									editPassword = '';
									settingsOpen = true;
									profileOpen = false;
								}}>Pengaturan Akun</button
							>
							<button class="popup-item" type="button">Bantuan</button>
						</div>
					{/if}
				</div>

				<button class="logout-button" type="button" onclick={handleLogout}>Keluar dari Buku</button>
			</div>
		</aside>
	{/if}

	<main class:withNavigation={data.isLoggedIn}>
		{@render children()}
	</main>

	{#if data.isLoggedIn}
		<nav class="bottom-nav">
			{#each navItems as item}
				<a class:active={isActive(item.href)} href={item.href}>
					{@html item.icon}
					<span>{item.shortLabel}</span>
				</a>
			{/each}
		</nav>
	{/if}

	{#if settingsOpen}
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
		<div class="modal-backdrop" onclick={() => (settingsOpen = false)} role="presentation">
			<div class="modal-card" onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1">
				<p class="muted mono" style="margin-bottom: 0.35rem;">Preferensi · Akun</p>
				<h2 class="section-title" style="margin-bottom: 1rem;">Pengaturan Akun</h2>
				<p class="muted" style="margin-bottom: 1.5rem; padding-bottom: 1rem; border-bottom: 1px dashed var(--rule);">
					Perbarui informasi pemilik buku kas.
				</p>

				<form class="form-grid" onsubmit={handleSaveSettings}>
					<div style="display: flex; gap: 1rem; align-items: center; margin-bottom: 0.25rem;">
						<div
							class="profile-avatar"
							style="width: 3.5rem; height: 3.5rem; font-size: 1.5rem;"
						>
							{editUsername
								.split(' ')
								.map((n) => n[0])
								.join('')
								.substring(0, 2)
								.toUpperCase() || 'U'}
						</div>
						<div>
							<p class="muted mono" style="margin: 0 0 0.25rem;">Avatar inisial</p>
							<p class="muted" style="font-size: 0.8rem;">
								Dihitung otomatis dari nama lengkap.
							</p>
						</div>
					</div>

					<label class="field">
						<span>Nama Lengkap</span>
						<input type="text" bind:value={editUsername} />
					</label>
					<label class="field">
						<span>Email</span>
						<input type="email" bind:value={editEmail} />
					</label>
					<label class="field">
						<span>Password Baru</span>
						<input
							type="password"
							bind:value={editPassword}
							placeholder="Kosongkan jika tidak diubah"
						/>
					</label>

					{#if saveError}
						<p class="error">{saveError}</p>
					{/if}

					<div class="button-row" style="justify-content: flex-end;">
						<button
							class="button-secondary"
							type="button"
							onclick={() => (settingsOpen = false)}
							disabled={saving}>Batal</button
						>
						<button class="button-primary" type="submit" disabled={saving}
							>{saving ? 'Menyimpan…' : 'Simpan Perubahan'}</button
						>
					</div>
				</form>
			</div>
		</div>
	{/if}
</div>
