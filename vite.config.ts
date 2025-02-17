
import { defineConfig } from 'vite';

export default defineConfig({
	define: {
		__BUILD_INFO__: JSON.stringify({
			sha: process.env.SHA || '',
			name: process.env.BUILD_NAME || '',
		})
	},
	root: 'ui',
	publicDir: 'pub',
	build: {
		outDir: '../pkg/ui/assets',
		assetsDir: '.',
		emptyOutDir: true,
	}
});
