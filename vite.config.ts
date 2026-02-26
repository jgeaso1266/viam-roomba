import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig(({ mode }) => ({
	plugins: [tailwindcss(), sveltekit()],
	define:
		mode === 'ssr'
			? {
					// Polyfill for packages that reference window at top level during SSR
					window: 'globalThis'
				}
			: {},
	ssr: {
		noExternal: ['@tanstack/svelte-query', '@viamrobotics/svelte-sdk']
	}
}));
