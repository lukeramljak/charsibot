/// <reference types="vitest/config" />
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { type Plugin, defineConfig } from 'vite';

const eagerInit = (): Plugin => ({
  name: 'eager-init',
  configureServer(server) {
    server.httpServer?.once('listening', () => {
      const addr = server.httpServer!.address();
      if (addr && typeof addr === 'object') {
        fetch(`http://localhost:${addr.port}/health`).catch(() => {});
      }
    });
  },
});

export default defineConfig({
  plugins: [tailwindcss(), sveltekit(), eagerInit()],
  resolve: process.env.VITEST
    ? {
        conditions: ['browser'],
      }
    : undefined,
  test: {
    silent: 'passed-only',
  },
});
