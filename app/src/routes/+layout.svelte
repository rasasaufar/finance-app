<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { clearAuthToken } from '$lib/auth';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';

	let { children, data } = $props<{
		children: import('svelte').Snippet;
		data: { isLoggedIn: boolean };
	}>();

	const navItems = [
		{
			href: '/dashboard',
			label: 'Dashboard',
			shortLabel: 'Dash',
			icon: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path><polyline points="9 22 9 12 15 12 15 22"></polyline></svg>'
		},
		{
			href: '/transactions',
			label: 'Transaksi',
			shortLabel: 'Trans',
			icon: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="8" y1="6" x2="21" y2="6"></line><line x1="8" y1="12" x2="21" y2="12"></line><line x1="8" y1="18" x2="21" y2="18"></line><line x1="3" y1="6" x2="3.01" y2="6"></line><line x1="3" y1="12" x2="3.01" y2="12"></line><line x1="3" y1="18" x2="3.01" y2="18"></line></svg>'
		},
		{
			href: '/budget-rules',
			label: 'Budget',
			shortLabel: 'Budget',
			icon: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>'
		},
		{
			href: '/salary-masters',
			label: 'Master Gaji',
			shortLabel: 'Gaji',
			icon: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="6" width="20" height="12" rx="2"></rect><path d="M2 10h20"></path><circle cx="12" cy="14" r="2"></circle></svg>'
		},
		{
			href: '/categories',
			label: 'Kategori',
			shortLabel: 'Kat',
			icon: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="7" height="7"></rect><rect x="14" y="3" width="7" height="7"></rect><rect x="14" y="14" width="7" height="7"></rect><rect x="3" y="14" width="7" height="7"></rect></svg>'
		},
		{
			href: '/reports',
			label: 'Laporan',
			shortLabel: 'Lap',
			icon: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="20" x2="18" y2="10"></line><line x1="12" y1="20" x2="12" y2="4"></line><line x1="6" y1="20" x2="6" y2="14"></line></svg>'
		}
	];

	const currentPath = $derived(page.url.pathname);

	function isActive(path: string): boolean {
		return currentPath === path;
	}

	let profileOpen = $state(false);
	let settingsOpen = $state(false);
	
	let username = $state('Rasa Saufar');
	let userEmail = $state('rasas@example.com');

	let editUsername = $state(username);
	let editEmail = $state(userEmail);
	let editPassword = $state('');

	let userInitials = $derived(username.split(' ').map(n => n[0]).join('').substring(0, 2).toUpperCase() || 'U');

	function handleSaveSettings(e: Event) {
		e.preventDefault();
		username = editUsername;
		userEmail = editEmail;
		settingsOpen = false;
	}

	async function handleLogout(): Promise<void> {
		clearAuthToken();
		await goto('/login');
	}
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>Dompet Pribadi</title>
	<meta name="viewport" content="width=device-width, initial-scale=1" />
</svelte:head>

<div class="app-shell">
	{#if data.isLoggedIn}
		<aside class="sidebar">
			<div class="brand">
				<span class="brand-badge">DP</span>
				<div>
					<p class="brand-title">Dompet Pribadi</p>
					<p class="brand-subtitle">Kontrol Budget Harian</p>
				</div>
			</div>

			<nav class="sidebar-nav">
				{#each navItems as item}
					<a class:active={isActive(item.href)} href={item.href}>
						{@html item.icon}
						{item.label}
					</a>
				{/each}
			</nav>

			<div class="sidebar-bottom">
				<div class="profile-widget">
					<button class="profile-button" type="button" onclick={() => profileOpen = !profileOpen}>
						<div class="profile-avatar">{userInitials}</div>
						<div class="profile-info">
							<p class="profile-name">{username}</p>
							<p class="profile-role">Admin</p>
						</div>
					</button>

					{#if profileOpen}
						<div class="profile-popup">
							<div class="popup-header">
								<p class="popup-email">{userEmail}</p>
							</div>
							<button class="popup-item" type="button" onclick={() => { editUsername = username; editEmail = userEmail; editPassword = ''; settingsOpen = true; profileOpen = false; }}>Pengaturan Akun</button>
							<button class="popup-item" type="button">Bantuan</button>
						</div>
					{/if}
				</div>

				<button class="logout-button" type="button" onclick={handleLogout}>Keluar</button>
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
		<div class="modal-backdrop" onclick={() => settingsOpen = false} role="presentation">
			<div class="modal-card" onclick={(e) => e.stopPropagation()} role="dialog">
				<h2 class="section-title">Pengaturan Akun</h2>
				<p class="muted" style="margin-bottom: 1.5rem;">Update informasi profil Anda di sini.</p>

				<form class="form-grid" onsubmit={handleSaveSettings}>
					<div style="display: flex; gap: 1rem; align-items: center; margin-bottom: 0.5rem;">
						<div class="profile-avatar" style="width: 4rem; height: 4rem; font-size: 1.5rem;">{editUsername.split(' ').map(n => n[0]).join('').substring(0, 2).toUpperCase() || 'U'}</div>
						<button class="button-secondary" type="button" style="padding: 0.4rem 0.8rem; font-size: 0.8rem; border-radius: 0.5rem; min-height: unset;">Ubah Avatar</button>
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
						<input type="password" bind:value={editPassword} placeholder="Kosongkan jika tidak ingin diubah" />
					</label>
					
					<div style="display: flex; gap: 0.5rem; justify-content: flex-end; margin-top: 1rem;">
						<button class="button-secondary" type="button" onclick={() => settingsOpen = false}>Batal</button>
						<button class="button-primary" type="submit">Simpan</button>
					</div>
				</form>
			</div>
		</div>
	{/if}
</div>
