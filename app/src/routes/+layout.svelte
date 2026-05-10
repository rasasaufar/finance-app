<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { clearAuthToken } from '$lib/auth';
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { User } from '$lib/types';

	const THEME_KEY = 'finance_theme';

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

	// Sidebar quick-action definitions per route
	const sidebarActions: Record<string, { label: string; href: string; primary?: boolean; icon: string }[]> = {
		'/transactions': [
			{
				label: 'Catat Transaksi',
				href: '/transactions#form',
				primary: true,
				icon: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>'
			}
		],
		'/budget-rules': [
			{
				label: 'Tambah Budget',
				href: '/budget-rules#form',
				primary: true,
				icon: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>'
			}
		],
		'/salary-masters': [
			{
				label: 'Tambah Gaji',
				href: '/salary-masters#form',
				primary: true,
				icon: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>'
			}
		],
		'/categories': [
			{
				label: 'Tambah Kategori',
				href: '/categories#form',
				primary: true,
				icon: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>'
			}
		],
		'/dashboard': [
			{
				label: 'Catat Transaksi',
				href: '/transactions',
				primary: true,
				icon: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>'
			},
			{
				label: 'Lihat Laporan',
				href: '/reports',
				icon: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M3 3v18h18"/><path d="M7 14l4-4 3 3 5-6"/></svg>'
			}
		],
		'/reports': [
			{
				label: 'Catat Transaksi',
				href: '/transactions',
				primary: true,
				icon: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>'
			}
		]
	};

	const AVATAR_KEY = 'finance_avatar';

	const currentPath = $derived(page.url.pathname);
	const currentActions = $derived(sidebarActions[currentPath] ?? []);

	// ── Dark mode ──
	let isDark = $state(false);

	function applyTheme(dark: boolean): void {
		document.documentElement.setAttribute('data-theme', dark ? 'dark' : 'light');
		document.body.setAttribute('data-theme', dark ? 'dark' : 'light');
	}

	function toggleTheme(): void {
		isDark = !isDark;
		applyTheme(isDark);
		try {
			localStorage.setItem(THEME_KEY, isDark ? 'dark' : 'light');
		} catch {
			// ignore
		}
	}

	function loadTheme(): void {
		try {
			const saved = localStorage.getItem(THEME_KEY);
			if (saved === 'dark') {
				isDark = true;
			} else if (saved === 'light') {
				isDark = false;
			} else {
				// Respect system preference on first visit
				isDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
			}
		} catch {
			isDark = false;
		}
		applyTheme(isDark);
	}
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

	// Saved avatar (base64 data URL) — persisted in localStorage
	let avatarDataUrl = $state<string | null>(null);

	// Edit form state
	let editUsername = $state('');
	let editEmail = $state('');
	let editPassword = $state('');
	// Preview of newly selected image before saving
	let editAvatarPreview = $state<string | null>(null);
	// Whether user explicitly wants to remove the avatar
	let editAvatarRemoved = $state(false);

	let fileInput: HTMLInputElement | null = $state(null);

	let userInitials = $derived(
		username
			.split(' ')
			.map((n) => n[0])
			.join('')
			.substring(0, 2)
			.toUpperCase() || 'U'
	);

	// The avatar shown in the edit modal preview area
	const editAvatarDisplay = $derived<string | null>(
		editAvatarRemoved ? null : (editAvatarPreview ?? avatarDataUrl)
	);

	// Initials computed from the edit form name (live preview)
	const editInitials = $derived(
		editUsername
			.split(' ')
			.map((n) => n[0])
			.join('')
			.substring(0, 2)
			.toUpperCase() || 'U'
	);

	let saveError = $state('');
	let saving = $state(false);

	function loadAvatarFromStorage(): void {
		try {
			const stored = localStorage.getItem(AVATAR_KEY);
			avatarDataUrl = stored ?? null;
		} catch {
			// ignore
		}
	}

	function saveAvatarToStorage(dataUrl: string | null): void {
		try {
			if (dataUrl) {
				localStorage.setItem(AVATAR_KEY, dataUrl);
			} else {
				localStorage.removeItem(AVATAR_KEY);
			}
		} catch {
			// ignore — storage might be full
		}
	}

	function handleFileChange(e: Event): void {
		const input = e.target as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;

		// Validate type
		if (!file.type.startsWith('image/')) {
			saveError = 'File harus berupa gambar (JPG, PNG, WebP, dll).';
			return;
		}

		// Validate size — max 2 MB
		if (file.size > 2 * 1024 * 1024) {
			saveError = 'Ukuran gambar maksimal 2 MB.';
			return;
		}

		saveError = '';
		editAvatarRemoved = false;

		const reader = new FileReader();
		reader.onload = (ev) => {
			const result = ev.target?.result;
			if (typeof result === 'string') {
				// Crop & resize to 256×256 via canvas to keep localStorage lean
				const img = new Image();
				img.onload = () => {
					const canvas = document.createElement('canvas');
					const SIZE = 256;
					canvas.width = SIZE;
					canvas.height = SIZE;
					const ctx = canvas.getContext('2d');
					if (!ctx) {
						editAvatarPreview = result;
						return;
					}
					// Center-crop
					const side = Math.min(img.width, img.height);
					const sx = (img.width - side) / 2;
					const sy = (img.height - side) / 2;
					ctx.drawImage(img, sx, sy, side, side, 0, 0, SIZE, SIZE);
					editAvatarPreview = canvas.toDataURL('image/jpeg', 0.85);
				};
				img.src = result;
			}
		};
		reader.readAsDataURL(file);
	}

	function handleRemoveAvatar(): void {
		editAvatarPreview = null;
		editAvatarRemoved = true;
		if (fileInput) fileInput.value = '';
	}

	function openSettings(): void {
		editUsername = username;
		editEmail = userEmail;
		editPassword = '';
		editAvatarPreview = null;
		editAvatarRemoved = false;
		saveError = '';
		settingsOpen = true;
		profileOpen = false;
	}

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

			// Persist avatar change
			if (editAvatarRemoved) {
				avatarDataUrl = null;
				saveAvatarToStorage(null);
			} else if (editAvatarPreview) {
				avatarDataUrl = editAvatarPreview;
				saveAvatarToStorage(editAvatarPreview);
			}

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
		loadAvatarFromStorage();
		loadTheme();
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
				<!-- Quick action buttons — context-aware per page -->
				{#if currentActions.length > 0}
					<div class="sidebar-actions">
						<p class="sidebar-action-label">Aksi Cepat</p>
						{#each currentActions as action}
							<a
								href={action.href}
								class="sidebar-action-btn"
								class:primary-action={action.primary}
							>
								{@html action.icon}
								{action.label}
							</a>
						{/each}
					</div>
				{/if}

				<!-- Theme toggle -->
				<div class="theme-toggle">
					<span class="theme-toggle-label">{isDark ? 'Mode Malam' : 'Mode Siang'}</span>
					<button class="theme-toggle-btn" type="button" onclick={toggleTheme} aria-label="Ganti tema">
						{#if isDark}
							<!-- Sun icon -->
							<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>
							Siang
						{:else}
							<!-- Moon icon -->
							<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
							Malam
						{/if}
					</button>
				</div>

				<div class="profile-widget">
					<button
						class="profile-button"
						type="button"
						onclick={() => (profileOpen = !profileOpen)}
					>
						<div class="profile-avatar">
							{#if avatarDataUrl}
								<img src={avatarDataUrl} alt={username} class="avatar-img" />
							{:else}
								{userInitials}
							{/if}
						</div>
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
							<button class="popup-item" type="button" onclick={openSettings}>
								Pengaturan Akun
							</button>
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
			<div class="modal-card settings-modal" onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1">
				<p class="muted mono" style="margin-bottom: 0.35rem; font-size: 0.6rem; letter-spacing: 0.18em; text-transform: uppercase;">Preferensi · Akun</p>
				<h2 class="section-title" style="margin-bottom: 1.25rem;">Pengaturan Akun</h2>

				<form class="form-grid" onsubmit={handleSaveSettings}>

					<!-- Avatar editor -->
					<div class="avatar-editor">
						<div class="avatar-preview-wrap">
							<div class="avatar-preview-circle">
								{#if editAvatarDisplay}
									<img src={editAvatarDisplay} alt="Preview avatar" class="avatar-img" />
								{:else}
									<span class="avatar-initials">{editInitials}</span>
								{/if}
							</div>
							<!-- Overlay edit button -->
							<label class="avatar-edit-overlay" title="Ganti foto">
								<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
								<input
									bind:this={fileInput}
									type="file"
									accept="image/*"
									class="avatar-file-input"
									onchange={handleFileChange}
								/>
							</label>
						</div>

						<div class="avatar-actions">
							<p class="avatar-label mono">Foto Profil</p>
							<p class="avatar-hint">JPG, PNG, WebP · Maks. 2 MB · Dipotong otomatis jadi persegi</p>
							<div class="avatar-btns">
								<label class="button-secondary avatar-upload-btn">
									<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="14" height="14"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
									{editAvatarDisplay ? 'Ganti Foto' : 'Upload Foto'}
									<input
										type="file"
										accept="image/*"
										class="avatar-file-input"
										onchange={handleFileChange}
									/>
								</label>
								{#if editAvatarDisplay}
									<button
										class="button-danger avatar-remove-btn"
										type="button"
										onclick={handleRemoveAvatar}
									>
										<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="14" height="14"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/><path d="M10 11v6M14 11v6"/><path d="M9 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/></svg>
										Hapus Foto
									</button>
								{/if}
							</div>
						</div>
					</div>

					<hr class="hrule dashed" />

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

					<div class="button-row" style="justify-content: flex-end; margin-top: 0.5rem;">
						<button
							class="button-secondary"
							type="button"
							onclick={() => (settingsOpen = false)}
							disabled={saving}>Batal</button
						>
						<button class="button-primary" type="submit" disabled={saving}>
							{saving ? 'Menyimpan…' : 'Simpan Perubahan'}
						</button>
					</div>
				</form>
			</div>
		</div>
	{/if}
</div>

<style>
	/* Avatar image inside the circle */
	.avatar-img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		border-radius: 50%;
		display: block;
	}

	/* ── Settings modal ── */
	.settings-modal {
		width: min(100%, 520px);
		max-height: 90dvh;
		overflow-y: auto;
	}

	/* ── Avatar editor block ── */
	.avatar-editor {
		display: flex;
		align-items: flex-start;
		gap: 1.25rem;
	}

	.avatar-preview-wrap {
		position: relative;
		flex-shrink: 0;
	}

	.avatar-preview-circle {
		width: 5rem;
		height: 5rem;
		border-radius: 50%;
		background: var(--ink);
		color: var(--paper);
		display: grid;
		place-items: center;
		overflow: hidden;
		border: 1.5px solid var(--ink);
		box-shadow: 3px 3px 0 var(--ink);
	}

	.avatar-initials {
		font-family: var(--font-display);
		font-style: italic;
		font-size: 1.75rem;
		line-height: 1;
		letter-spacing: -0.03em;
		color: var(--paper);
		user-select: none;
	}

	/* Pencil overlay on hover */
	.avatar-edit-overlay {
		position: absolute;
		inset: 0;
		border-radius: 50%;
		background: rgba(28, 25, 21, 0.55);
		display: grid;
		place-items: center;
		cursor: pointer;
		opacity: 0;
		transition: opacity 0.2s;
		color: var(--paper);
	}

	.avatar-preview-wrap:hover .avatar-edit-overlay {
		opacity: 1;
	}

	.avatar-file-input {
		display: none;
	}

	.avatar-actions {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
	}

	.avatar-label {
		margin: 0;
		font-size: 0.65rem;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		color: var(--ink);
		font-family: var(--font-mono);
	}

	.avatar-hint {
		margin: 0;
		font-size: 0.78rem;
		color: var(--ink-faint);
		font-style: italic;
		font-family: var(--font-display);
		line-height: 1.4;
	}

	.avatar-btns {
		display: flex;
		gap: 0.6rem;
		flex-wrap: wrap;
		margin-top: 0.25rem;
	}

	.avatar-upload-btn {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
		cursor: pointer;
		font-size: 0.82rem;
		padding: 0.55rem 0.85rem;
		min-height: unset;
	}

	.avatar-remove-btn {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
		font-size: 0.82rem;
		padding: 0.55rem 0.85rem;
		min-height: unset;
	}
</style>
