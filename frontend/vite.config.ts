import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';
import { fileURLToPath, URL } from 'node:url';

const rootDir = fileURLToPath(new URL('..', import.meta.url));

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, rootDir, '');
  const agentPort = env.AGENT_PORT || env.PORT || '13001';
  const agentTarget = env.VITE_AGENT_BASE_URL || `http://localhost:${agentPort}`;
  const backendTarget = env.VITE_BACKEND_PROXY_TARGET || 'http://localhost:8080';

  return {
    envDir: '..',
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: backendTarget,
        changeOrigin: true,
      },
      '/agent': {
        target: agentTarget,
        changeOrigin: true,
        rewrite: (path: string) => path.replace(/^\/agent/, '/api'),
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['vue', 'vue-router', 'pinia'],
          d3: ['d3'],
        },
      },
    },
  },
  };
});
