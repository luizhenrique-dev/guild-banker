import { expect, test, type Page } from '@playwright/test'

const username = process.env.E2E_USERNAME || 'foo.bar@email.com'
const password = process.env.E2E_PASSWORD || process.env.KC_TEST_USER_PASSWORD!
const welcome = process.env.E2E_NAME || 'Foo Bar'
const tokenEndpoint = '/realms/guild-banker/protocol/openid-connect/token'

async function login(page: Page) {
  await expect(page.locator('#username')).toBeVisible()
  await page.locator('#username').fill(username)
  await page.locator('#password').fill(password).catch(() => { throw new Error('Unable to fill the test password') })
  await page.locator('#kc-login').click()
}

test('real login opens Home, reuses SSO on reload and UI logout ends SSO', async ({ page, context }) => {
  await page.goto('/')
  await login(page)
  await expect(page.getByRole('heading', { name: `Bem-vindo(a) ${welcome}` })).toBeVisible()
  await expect(page).toHaveURL(/\/home(?:\?|$)/)
  expect(context.serviceWorkers()).toHaveLength(0)
  expect(await page.evaluate(() => [localStorage, sessionStorage].some((storage) => Object.values(storage).some((value) => /^eyJ[\w-]+\./.test(String(value)))))).toBe(false)
  await page.reload()
  await expect(page.getByRole('heading', { name: `Bem-vindo(a) ${welcome}` })).toBeVisible()
  await page.getByRole('button', { name: 'Sair', exact: true }).click()
  await expect(page.locator('#username')).toBeVisible()
  await page.goto('/home')
  await expect(page.locator('#username')).toBeVisible()
  await page.goBack()
  await expect(page.getByRole('navigation')).toHaveCount(0)
})

test('real deep link preserves business query and sends only Bearer to the API', async ({ page }) => {
  await page.goto('/transactions?guild=999&category=OTHER')
  const api = page.waitForRequest((r) => r.url().includes('/api/v1/guilds'))
  await login(page)
  const request = await api
  expect(Boolean(request.headers().authorization?.startsWith('Bearer '))).toBe(true)
  expect(Boolean(request.headers()['x-user-id'] || request.headers()['x-user-email'])).toBe(false)
  await expect(page).toHaveURL(/\/transactions\?guild=999&category=OTHER$/)
  await expect(page.getByRole('navigation')).toBeVisible()
  await page.reload()
  await expect(page.getByRole('navigation')).toBeVisible()
  await expect(page).toHaveURL(/\/transactions\?guild=999&category=OTHER$/)
})

test('real API authorizes the linked identity', async ({ page }) => {
  await page.goto('/fixed-expenses')
  const api = page.waitForResponse((r) => r.url().includes('/api/v1/fixed-expenses'))
  await login(page)
  const response = await api
  expect(response.status(), 'Requires an active user_account linked to the Keycloak sub; this suite never seeds data').toBe(200)
  await expect(page.getByRole('heading', { name: 'Minhas despesas fixas', exact: true })).toBeVisible()
})

test('real token renewal occurs on expiration without changing the server session policy', async ({ page }) => {
  await page.clock.install()
  await page.goto('/home')
  const issued = page.waitForResponse((r) => r.url().endsWith(tokenEndpoint) && r.request().postData()?.includes('grant_type=authorization_code') === true)
  await login(page)
  const lifetime = Number((await (await issued).json()).expires_in)
  expect(Number.isFinite(lifetime) && lifetime > 0).toBe(true)
  await expect(page.getByRole('heading', { name: `Bem-vindo(a) ${welcome}` })).toBeVisible()
  const refreshed = page.waitForResponse((r) => r.url().endsWith(tokenEndpoint) && r.request().postData()?.includes('grant_type=refresh_token') === true)
  await page.clock.fastForward((lifetime + 1) * 1000)
  expect((await refreshed).status()).toBe(200)
  await expect(page.getByRole('heading', { name: `Bem-vindo(a) ${welcome}` })).toBeVisible()
})

test('an expired token requires login after SSO is ended in another tab', async ({ page, context }) => {
  await page.clock.install()
  await page.goto('/home')
  const issued = page.waitForResponse((r) => r.url().endsWith(tokenEndpoint) && r.request().postData()?.includes('grant_type=authorization_code') === true)
  await login(page)
  const lifetime = Number((await (await issued).json()).expires_in)
  await expect(page.getByRole('heading', { name: `Bem-vindo(a) ${welcome}` })).toBeVisible()
  const otherTab = await context.newPage()
  await otherTab.goto('/home')
  await expect(otherTab.getByRole('button', { name: 'Sair', exact: true })).toBeVisible()
  await otherTab.getByRole('button', { name: 'Sair', exact: true }).click()
  await expect(otherTab.locator('#username')).toBeVisible()
  await page.clock.fastForward((lifetime + 1) * 1000)
  await expect(page.locator('#username')).toBeVisible()
  await expect(page.getByRole('navigation')).toHaveCount(0)
})

test('logout transport failure keeps local content closed and permits explicit retry', async ({ page }) => {
  await page.goto('/home')
  await login(page)
  await expect(page.getByRole('button', { name: 'Sair', exact: true })).toBeVisible()
  const discovery = '**/realms/guild-banker/.well-known/openid-configuration'
  await page.route(discovery, (route) => route.abort())
  await page.getByRole('button', { name: 'Sair', exact: true }).click()
  await expect(page.getByRole('alert')).toContainText('Sessão local encerrada')
  await expect(page.getByRole('navigation')).toHaveCount(0)
  await page.unroute(discovery)
  await page.getByRole('button', { name: 'Tentar novamente' }).click()
  await expect(page.locator('#username')).toBeVisible()
})
