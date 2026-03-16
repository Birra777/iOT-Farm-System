import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api/events': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        // SSE requires the proxy to not buffer the response
        configure: (proxy) => {
          proxy.on('proxyReq', (proxyReq) => {
            proxyReq.setHeader('Accept', 'text/event-stream')
          })
        },
      },
      '/api': 'http://localhost:8080',
      '/health': 'http://localhost:8080',
    },
  },
})
