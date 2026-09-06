import path from 'node:path'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(import.meta.dirname, './src'),
    },
  },
  server: {
    port: 3000,
    // ポートが埋まっていたら黙って別ポートにずらさず、エラーで止める
    // （ずれると Caddy の転送先 3000 と食い違って原因が分かりにくくなる）
    strictPort: true,
    proxy: {
      // Caddy を起動しなくても Go(8080) に届くようにする。
      // Caddy 経由の場合はそちらが先に /api/* を処理するので競合しない。
      '/api': 'http://localhost:8080',
    },
  },
})
