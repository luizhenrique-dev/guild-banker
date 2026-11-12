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

async function completeHomeOnboarding(page: Page) {
  await expect(page.getByRole('heading', { name: /^(Bem-vindo\(a\)|Crie sua guild|Escolha sua guild)/ })).toBeVisible()
  if (await page.getByRole('heading', { name: 'Crie sua guild' }).isVisible()) {
    if (process.env.E2E_ALLOW_GUILD_CREATE !== 'true') throw new Error('Provide test membership or E2E_ALLOW_GUILD_CREATE=true in a dedicated environment')
    await page.getByRole('textbox', { name: 'Nome de exibição' }).fill('Auth test ' + Date.now())
    await page.getByRole('button', { name: 'Criar guild', exact: true }).click()
  } else if (await page.getByRole('heading', { name: 'Escolha sua guild' }).isVisible()) {
    await page.getByRole('button', { name: /^Entrar em / }).first().click()
  }
  await expect(page.getByRole('heading', { name: `Bem-vindo(a) ${welcome}` })).toBeVisible()
}

test('real login opens Home, reuses SSO on reload and UI logout ends SSO', async ({ page, context }) => {
  await page.goto('/')
  await login(page)
  await completeHomeOnboarding(page)
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
  await expect(page.getByRole('heading', { name: 'Escolha sua guild' })).toBeVisible()
  await expect(page.getByRole('navigation')).toHaveCount(0)
  await page.reload()
  await expect(page.getByRole('heading', { name: 'Escolha sua guild' })).toBeVisible()
  await expect(page).toHaveURL(/\/transactions\?guild=999&category=OTHER$/)
  await page.getByRole('button', { name: /^Entrar em / }).first().click()
  await expect(page.getByRole('navigation')).toBeVisible()
  expect(new URL(page.url()).searchParams.get('category')).toBe('OTHER')
  expect(new URL(page.url()).searchParams.get('guild')).not.toBe('999')
})

test('real API provisions and authorizes a verified identity', async ({ page }) => {
  await page.goto('/home')
  const guildResponse = page.waitForResponse((r) => r.url().endsWith('/api/v1/guilds') && r.request().method() === 'GET')
  await login(page)
  const guilds = await (await guildResponse).json() as { id: number; display_name: string }[]
  let guild = guilds[0]
  if (!guild) {
    if (process.env.E2E_ALLOW_GUILD_CREATE !== 'true') throw new Error('Provide test membership or E2E_ALLOW_GUILD_CREATE=true in a dedicated environment')
    const created = page.waitForResponse((r) => r.url().endsWith('/api/v1/guilds') && r.request().method() === 'POST')
    await completeHomeOnboarding(page)
    guild = await (await created).json()
  } else await completeHomeOnboarding(page)

  expect(guild && Number.isSafeInteger(guild.id) && guild.id > 0).toBe(true)
  const api = page.waitForResponse((r) => r.url().endsWith('/api/v1/guilds/' + guild!.id + '/fixed-expenses'))
  await page.goto('/fixed-expenses?guild=' + guild!.id)
  await expect(page).toHaveURL(new RegExp('guild=' + guild!.id + '(?:&|$)'))
  expect((await api).status()).toBe(200)
  await expect(page.getByRole('heading', { name: 'Despesas fixas da guild', exact: true })).toBeVisible()
})

test('real token renewal occurs on expiration without changing the server session policy', async ({ page }) => {
  await page.clock.install()
  await page.goto('/home')
  const issued = page.waitForResponse((r) => r.url().endsWith(tokenEndpoint) && r.request().postData()?.includes('grant_type=authorization_code') === true)
  await login(page)
  await completeHomeOnboarding(page)
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
  await completeHomeOnboarding(page)
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
  await completeHomeOnboarding(page)
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
