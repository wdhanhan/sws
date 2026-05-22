import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 13000,
    host: '0.0.0.0',
    allowedHosts: ['seer6000.vip', '.seer6000.vip'],
    proxy: {
      '/api': 'http://localhost:8080',
      '/health': 'http://localhost:8080',
      '/ws': {
        target: 'http://localhost:8080',
        ws: true,
      },
    }
  }
})
