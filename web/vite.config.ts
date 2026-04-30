import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  base: '/admin/',
  build: { outDir: 'dist', emptyOutDir: true, sourcemap: false },
  server: { proxy: { '/cms.v1.CMSService': 'http://localhost:8080', '/uploads': 'http://localhost:8080' } }
})
