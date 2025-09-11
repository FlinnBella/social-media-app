import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tsconfigPaths from 'vite-tsconfig-paths';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react(), tsconfigPaths()],
  optimizeDeps: {
    exclude: ['lucide-react'],
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        configure(proxy) {
        proxy.on('proxyReq', (_proxyReq, req) => console.log('→', req.method, req.url));
        proxy.on('proxyRes', (res, req) => console.log('←', res.statusCode, req.url));
        },
      },
    },
  }
});
