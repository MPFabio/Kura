import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  optimizeDeps: {
    include: ['react-syntax-highlighter', 'react-syntax-highlighter/dist/esm/styles/prism'],
  },
  ssr: {
    noExternal: ['react-syntax-highlighter'],
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8000',
        changeOrigin: true,
        ws: true, // Activer le support WebSocket
      },
    },
  },
})
