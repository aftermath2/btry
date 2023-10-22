import { defineConfig } from 'vite';
import solidPlugin from 'vite-plugin-solid';
// import devtools from 'solid-devtools/vite';

const commitHash = require('child_process')
	.execSync('git rev-parse --short HEAD')
	.toString().trimEnd();

export default defineConfig({
	define: {
		__COMMIT_HASH__: JSON.stringify(commitHash)
	},
	plugins: [
		/* 
		Uncomment the following line to enable solid-devtools.
		For more info see https://github.com/thetarnav/solid-devtools/tree/main/packages/extension#readme
		*/
		// devtools(),
		solidPlugin(),
	],
	server: {
		host: '127.0.0.1',
		port: 4000,
	},
	build: {
		target: 'esnext',
	},
	// envPrefix: "ENV"
});