import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// Built assets are embedded into the mbsecli Go binary (see web/embed.go),
// so paths must be relative — the app can be served from any mount point.
export default defineConfig({
  plugins: [react()],
  base: './',
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:4173',
      '/ws': { target: 'ws://localhost:4173', ws: true },
    },
  },
  build: {
    outDir: 'dist',
    // The dist/ directory is embedded into the Go binary and lives inside a
    // mounted project folder that disallows deleting/renaming files once
    // written — emptying the dir on every build would fail. Vite overwrites
    // files in place instead, which works fine here.
    emptyOutDir: false,
  },
});
