import { defineConfig } from "vite";
import tailwindcss from '@tailwindcss/vite';
import elmPlugin from 'vite-plugin-elm'
import { resolve } from 'node:path'

// https://vitejs.dev/config/
export default defineConfig(async () => ({
  plugins: [elmPlugin({

  }), tailwindcss()],
  build: {
    rollupOptions: {
      external: ['@wailsio/runtime', '/wails/runtime.js'],
      input: {
        main: resolve(__dirname, 'index.html'),
      },
    },
  },
}));
