/// <reference types="vitest/config" />
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig, type Plugin } from 'vite';

const eagerServerHooks = (): Plugin => ({
  name: 'eager-server-hooks',
  configureServer(server) {
    server.httpServer?.once('listening', () => {
      const address = server.httpServer!.address();
      if (address && typeof address === 'object') {
        fetch(`http://localhost:${address.port}/health`).catch(() => {});
      }
    });
  },
});

export default defineConfig({
  plugins: [tailwindcss(), eagerServerHooks(), sveltekit()],
  resolve: process.env.VITEST
    ? {
        conditions: ['browser'],
      }
    : undefined,
  test: {
    silent: 'passed-only',
  },
});
