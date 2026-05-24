import path from "node:path"
import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import { defineConfig } from "vite"

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  // Polling watcher: native filesystem events don't propagate through Docker
  // bind mounts on macOS/Windows, so HMR misses edits. Polling is heavier on
  // CPU but reliable across hosts.
  server: {
    watch: {
      usePolling: true,
      interval: 100,
    },
  },
})
