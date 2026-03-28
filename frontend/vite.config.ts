import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig(({ mode }) => {
  // load env and allow calling with VITE_BACKEND_URL
  const env = loadEnv(mode, process.cwd(), '')
  const backend = env.VITE_BACKEND_URL || 'http://localhost:8080'

  return {
    plugins: [react(), tailwindcss()],
    server: {
      proxy: {
        '/api': backend,
        '/health': backend,
      },
    },
  }
})
