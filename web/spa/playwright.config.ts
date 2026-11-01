import { defineConfig } from '@playwright/test'

const port = Number(process.env.PLAYWRIGHT_PORT || 5173)

export default defineConfig({
  testDir: './e2e',
  testIgnore: '**/production.spec.ts',
  use: { baseURL: `http://127.0.0.1:${port}`, locale: 'pt-BR', trace: 'retain-on-failure' },
  webServer: {
    command: `npm run dev:test -- --port ${port}`,
    url: `http://127.0.0.1:${port}`,
    reuseExistingServer: false,
  },
})
