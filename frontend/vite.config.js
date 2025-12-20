
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['icon.svg'],
      manifest: {
        name: 'AvidLearner — Software Engineering Coach',
        short_name: 'AvidLearner',
        description: 'Learn software engineering concepts, take quizzes, and sharpen typing speed.',
        theme_color: '#0b1020',
        background_color: '#0b1020',
        display: 'standalone',
        start_url: '/',
        scope: '/',
        icons: [
          {
            src: 'icon.svg',
            sizes: '512x512',
            type: 'image/svg+xml',
            purpose: 'any maskable'
          }
        ]
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,png,svg,webmanifest}'],
        runtimeCaching: [
          {
            urlPattern: /^https?:\/\/[^/]+\/api\/lessons$/,
            handler: 'NetworkFirst',
            options: {
              cacheName: 'lessons-cache',
              expiration: {
                maxEntries: 10,
                maxAgeSeconds: 86400 // 24 hours
              },
              networkTimeoutSeconds: 10
            }
          },
          {
            urlPattern: /^https?:\/\/[^/]+\/api\/random\?/,
            handler: 'NetworkFirst',
            options: {
              cacheName: 'random-lessons-cache',
              expiration: {
                maxEntries: 50,
                maxAgeSeconds: 3600 // 1 hour
              },
              networkTimeoutSeconds: 10
            }
          },
          {
            urlPattern: /^https?:\/\/[^/]+\/api\/session\?/,
            handler: 'NetworkFirst',
            options: {
              cacheName: 'session-cache',
              expiration: {
                maxEntries: 30,
                maxAgeSeconds: 1800 // 30 minutes
              },
              networkTimeoutSeconds: 10
            }
          }
        ]
      },
      devOptions: {
        enabled: true,
        suppressWarnings: true
      }
    })
  ],
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:8081'
    }
  },
  build: {
    outDir: 'dist'
  }
})
