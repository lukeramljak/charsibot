import { defineConfig } from 'vite';

export default defineConfig({
  build: {
    emptyOutDir: false,
    minify: false,
    outDir: 'build',
    ssr: 'server/index.ts',
    target: 'node22',
    rollupOptions: {
      external: ['./handler.js'],
      output: {
        entryFileNames: 'server.js',
      },
    },
  },
});
