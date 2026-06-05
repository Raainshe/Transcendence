import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

/** Backend URL for the Vite dev proxy (not exposed to the browser). */
const apiProxyTarget = process.env.API_PROXY_TARGET ?? 'http://localhost:8080'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  server: {
    proxy: {
      '/api': { target: apiProxyTarget, changeOrigin: true, ws: true },
      '/uploads': { target: apiProxyTarget, changeOrigin: true },
    },
  },
})
