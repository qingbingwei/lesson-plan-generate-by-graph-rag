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
          manualChunks(id: string) {
            if (!id.includes('node_modules')) return undefined;

            if (id.includes('node_modules/d3')) return 'd3';
            if (id.includes('node_modules/marked')) return 'markdown';
            if (id.includes('node_modules/@element-plus/icons-vue')) {
              return 'element-plus-icons';
            }
            if (id.includes('node_modules/element-plus')) return 'element-plus';
            if (id.includes('node_modules/dayjs')) return 'dayjs';
            if (id.includes('node_modules/axios')) return 'axios';
            if (id.includes('node_modules/@vueuse/core')) return 'vueuse';

            return 'vendor';
          },
        },
      },
      // P0 阶段先完成主入口拆分，后续可继续做 Element Plus 按需加载。
      chunkSizeWarningLimit: 800,
    },
  };
});
