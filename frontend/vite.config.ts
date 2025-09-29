import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import vueDevTools from 'vite-plugin-vue-devtools'
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
    tailwindcss(),
    VitePWA({
      registerType: 'autoUpdate', // 👈 ensures SW updates silently
      devOptions: {
        enabled: true, // allow testing in dev
      },
      manifest: {
        name: 'My App',
        short_name: 'App',
        description: 'My PWA using Vite',
        theme_color: '#000000',
        background_color: '#ffffff',
        display: 'standalone',
        start_url: '/',
        icons: [
          {
            src: '/pwa-192x192.png',
            sizes: '192x192',
            type: 'image/png',
          },
          {
            src: '/pwa-512x512.png',
            sizes: '512x512',
            type: 'image/png',
          },
          {
            src: '/pwa-512x512.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'any maskable',
          },
        ],
      },
      workbox: {
        clientsClaim: true, // take control immediately
        skipWaiting: true, // activate new SW without reload
      },
    }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    watch: {
      usePolling: true,
      interval: 1000,
    },
    allowedHosts: true,
    cors: {
      origin: '*',
      methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
      allowedHeaders: ['Content-Type', 'Authorization'],
      credentials: true,
    },
    proxy: {
      '/api': {
        target: process.env.VITE_BACKEND_URL || 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''),
        configure(proxy, options) {
          proxy.on('proxyRes', (proxyRes, req, res) => {
            if (proxyRes.statusCode && proxyRes.statusCode >= 400) {
              res.statusCode = proxyRes.statusCode
              Object.entries(proxyRes.headers).forEach(([key, value]) => {
                if (value !== undefined) {
                  res.setHeader(key, value)
                }
              })
              proxyRes.pipe(res)
            }
          })
        },
      },
    },
  },
})
