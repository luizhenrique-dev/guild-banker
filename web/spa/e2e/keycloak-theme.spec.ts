import { createHash } from 'node:crypto'
import { expect, test, type Page } from '@playwright/test'

const issuer = 'http://localhost:8080/realms/guild-banker'
const password = process.env.E2E_PASSWORD || process.env.KC_TEST_USER_PASSWORD!

async function fillPassword(page: Page, value: string) {
  await page.locator('#password').fill(value).catch(() => { throw new Error('Unable to fill the test password') })
}

function authorizationURL() {
  const url = new URL(`${issuer}/protocol/openid-connect/auth`)
  url.search = new URLSearchParams({
    client_id: 'guild-banker-ui', redirect_uri: 'http://localhost:5173/',
    response_type: 'code', response_mode: 'query', scope: 'openid',
    code_challenge: createHash('sha256').update('issue6-theme-test-verifier-with-at-least-43-characters').digest('base64url'),
    code_challenge_method: 'S256',
  }).toString()
  return url.toString()
}

test('branded login is responsive, accessible and serves its resources locally', async ({ page }) => {
  const failedResources: string[] = []
  page.on('response', (r) => { if (r.url().includes('/resources/') && r.status() >= 400) failedResources.push(r.url()) })
  for (const width of [1440, 375]) {
    await page.setViewportSize({ width, height: 900 })
    await page.goto(authorizationURL())
    await expect(page.locator('#kc-header-wrapper')).toHaveText('GuildBanker')
    await expect(page.locator('#kc-header-wrapper')).toHaveCSS('color', 'rgb(10, 11, 13)')
    await expect(page.locator('#username')).toHaveAccessibleName(/Email|E-mail/i)
    await expect(page.locator('#password')).toHaveAccessibleName(/Password|Senha/i)
    await expect(page.locator('#kc-login')).toHaveCSS('background-color', 'rgb(0, 82, 255)')
    if (width > 900) {
      await expect(page.locator('.guild-hero')).toBeVisible()
      await expect(page.locator('.guild-hero p')).toHaveCSS('color', 'rgb(168, 172, 179)')
    }
    else await expect(page.locator('.guild-hero')).toBeHidden()
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
    await page.locator('#username').focus()
    await page.keyboard.press('Tab')
    await expect(page.locator('#password')).toBeFocused()
    expect(await page.locator('#password').evaluate((el) => getComputedStyle(el).outlineStyle)).toBe('solid')
    await page.evaluate(() => document.fonts.ready)
    expect(await page.evaluate(() => document.fonts.check('16px Inter'))).toBe(true)
    const resources = await page.evaluate(() => performance.getEntriesByType('resource').map((r) => r.name))
    expect(resources.every((url) => new URL(url).origin === 'http://localhost:8080')).toBe(true)
  }
  expect(failedResources).toEqual([])
})

test('invalid credentials associate the error with both fields', async ({ page }) => {
  await page.goto(authorizationURL())
  await page.locator('#username').fill(process.env.E2E_USERNAME || 'foo.bar@email.com')
  await page.locator('#password').fill('invalid-issue6-password')
  await page.locator('#kc-login').click()
  await expect(page.locator('#input-error')).toBeVisible()
  for (const field of ['#username', '#password']) {
    await expect(page.locator(field)).toHaveAttribute('aria-invalid', 'true')
    await expect(page.locator(field)).toHaveAttribute('aria-describedby', 'input-error')
  }
  await expect(page.locator('#kc-login')).toBeVisible()
})

test('server error page retains the brand and only native recovery links', async ({ page }) => {
  const url = new URL(authorizationURL())
  url.searchParams.set('redirect_uri', 'https://invalid.example/blocked')
  await page.goto(url.toString())
  await expect(page.locator('body')).toHaveAttribute('data-page-id', 'login-error')
  await expect(page.locator('#kc-header-wrapper')).toHaveText('GuildBanker')
  await expect(page.locator('#kc-error-message')).toContainText(/redirect_uri/)
  expect(await page.locator('a[href^="https://invalid.example"]').count()).toBe(0)
})

test('valid credentials preserve the authorization code callback', async ({ page, request }) => {
  await page.route(/^http:\/\/localhost:5173\//, (route) => route.fulfill({ contentType: 'text/html', body: '<p>OIDC callback received</p>' }))
  await page.goto(authorizationURL())
  await page.locator('#username').fill(process.env.E2E_USERNAME || 'foo.bar@email.com')
  await fillPassword(page, password)
  await page.locator('#kc-login').click()
  await expect.poll(() => new URL(page.url()).origin === 'http://localhost:5173' && new URL(page.url()).searchParams.has('code')).toBe(true)
  const response = await request.post(`${issuer}/protocol/openid-connect/token`, { form: {
    grant_type: 'authorization_code', client_id: 'guild-banker-ui',
    redirect_uri: 'http://localhost:5173/', code: new URL(page.url()).searchParams.get('code')!,
    code_verifier: 'issue6-theme-test-verifier-with-at-least-43-characters',
  } })
  expect(response.status()).toBe(200)
  const tokens = await response.json()
  const claims = JSON.parse(Buffer.from(tokens.access_token.split('.')[1], 'base64url').toString())
  expect(claims.aud).toContain('guild-banker-api')
})

test('auxiliary logout confirmation is inherited without the product layout', async ({ page }) => {
  await page.goto(`${issuer}/protocol/openid-connect/logout`)
  await expect(page.locator('body')).toHaveAttribute('data-page-id', 'login-logout-confirm')
  await expect(page.locator('#kc-logout')).toBeVisible()
  await expect(page.locator('.guild-hero')).toHaveCount(0)
})

test('password entry failure replaces sensitive Playwright diagnostics', async ({ page }) => {
  await page.setContent('<input id="password" type="password" readonly>')
  page.setDefaultTimeout(200)
  let message = ''
  try {
    await fillPassword(page, 'issue6-sentinel-must-not-leak')
  } catch (error) {
    message = (error as Error).message
  }
  expect(message).toBe('Unable to fill the test password')
  expect(message).not.toContain('issue6-sentinel-must-not-leak')
})
