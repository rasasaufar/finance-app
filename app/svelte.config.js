// ⚠️ DEPLOYMENT: Harus pakai adapter-node (self-hosted di VPS Docker)
// JANGAN ganti ke adapter-cloudflare — Dockerfile expect output di build/
import adapter from '@sveltejs/adapter-node';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	compilerOptions: {
		// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
		runes: ({ filename }) => (filename.split(/[/\\]/).includes('node_modules') ? undefined : true)
	},
	kit: { adapter: adapter() }
};

export default config;
