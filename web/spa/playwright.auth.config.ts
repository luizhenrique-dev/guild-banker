import { defineConfig } from '@playwright/test'
import { existsSync } from 'node:fs'
import { resolve } from 'node:path'

const envFile = resolve(__dirname, '../../infra/.env')
if (existsSync(envFile)) process.loadEnvFile(envFile)
if (!process.env.E2E_PASSWORD && !process.env.KC_TEST_USER_PASSWORD) {
  throw new Error('Set E2E_PASSWORD or KC_TEST_USER_PASSWORD for the real Keycloak test user')
}

export default defineConfig({
  testDir: './e2e',
  testMatch: ['**/auth.spec.ts', '**/keycloak-theme.spec.ts'],
  workers: 1,
  timeout: 45000,
  // Authentication artifacts must never record passwords, codes or tokens.
  use: { baseURL: 'http://localhost:5173', locale: 'pt-BR', trace: 'off', screenshot: 'off', video: 'off' },
  webServer: {
    command: 'npm run dev -- --host localhost --port 5173 --strictPort',
    url: 'http://localhost:5173',
    reuseExistingServer: false,
    env: { VITE_USE_MOCKS: 'false', VITE_KEYCLOAK_URL: 'http://localhost:8080', VITE_KEYCLOAK_REALM: 'guild-banker', VITE_KEYCLOAK_CLIENT_ID: 'guild-banker-ui' },
  },
})
