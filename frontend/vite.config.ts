import path from 'node:path';
import fs from 'node:fs';
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import svgr from "vite-plugin-svgr";
import { VitePWA } from 'vite-plugin-pwa';
import mkcert from 'vite-plugin-mkcert';

// https://vite.dev/config/
export default defineConfig({
  server: {
    host: "0.0.0.0",
    port: 3000,
    strictPort: true,
    proxy: {
      '/api': {
        target: 'https://192.168.1.150:8000',
        changeOrigin: true
      }
    },
    https: {
      key: fs.readFileSync(path.resolve(__dirname, 'cert.key')),
      cert: fs.readFileSync(path.resolve(__dirname, 'cert.crt')),
    },
  },
  plugins: [
    react(),
    svgr(),
    mkcert(),
    VitePWA({
      registerType: 'autoUpdate',
      manifest: {
        "name": "VETRYAKI",
        "short_name": "VETRYAKI",
        "start_url": "/RIP/",
        "display": "standalone",
        "background_color": "#ffffff",
        "theme_color": "#ffffff",
        "orientation": "portrait-primary",
        "icons": [
          {
            "src": "/RIP/logo128.png",
            "type": "image/png", "sizes": "128x128"
          },
          {
            "src": "/RIP/logo512.png",
            "type": "image/png", "sizes": "512x512"
          }
        ]
      }
    }),
  ],
})
