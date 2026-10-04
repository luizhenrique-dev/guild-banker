import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  testIgnore: '**/production.spec.ts',
  use: { baseURL: 'http://127.0.0.1:5173', locale: 'pt-BR', trace: 'retain-on-failure' },
  webServer: {
    command: 'npm run dev:test',
    url: 'http://127.0.0.1:5173',
    reuseExistingServer: false,
  },
})
