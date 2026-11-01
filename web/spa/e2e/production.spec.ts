import { expect, test } from '@playwright/test'

const issuer = 'http://localhost:8080/realms/guild-banker'
for (const route of ['/', '/home', '/fixed-expenses', '/transactions?guild=2&category=OTHER', '/import', '/guild', '/missing']) {
  test(`compiled ${route} requires OIDC without private content or demo`, async ({ page, context }) => {
    const requests: string[] = []
    const errors: string[] = []
    page.on('request', (request) => { requests.push(request.url()) })
    page.on('pageerror', (error) => { errors.push(error.message) })
    await page.route(`${issuer}/.well-known/openid-configuration`, (route) => route.fulfill({ json: {
      issuer, authorization_endpoint: `${issuer}/protocol/openid-connect/auth`, token_endpoint: `${issuer}/protocol/openid-connect/token`, end_session_endpoint: `${issuer}/protocol/openid-connect/logout`,
    } }))
    await page.route(`${issuer}/protocol/openid-connect/auth?*`, (route) => route.fulfill({ contentType: 'text/html', body: '<h1>OIDC navigation intercepted</h1>' }))
    await page.goto(route)
    await expect(page.getByRole('heading', { name: 'OIDC navigation intercepted' })).toBeVisible()
    const login = new URL(page.url())
    expect(login.searchParams.get('client_id')).toBe('guild-banker-ui')
    expect(login.searchParams.get('response_type')).toBe('code')
    expect(login.searchParams.get('code_challenge_method')).toBe('S256')
    expect(login.searchParams.get('redirect_uri')).toBe(`http://127.0.0.1:4173${route}`)
    expect(requests.some((url) => /\/assets\/.*\.js/.test(url))).toBe(true)
    expect(requests.some((url) => /\/assets\/.*\.css/.test(url))).toBe(true)
    expect(requests.filter((url) => /mockServiceWorker|\/api\//.test(url))).toEqual([])
    expect(context.serviceWorkers()).toEqual([])
    expect(errors).toEqual([])
    await expect(page.getByRole('navigation')).toHaveCount(0)
    await expect(page.getByText('Ana Silva', { exact: true })).toHaveCount(0)
  })
}

test('invalid callback stays closed and offers an explicit retry without an automatic login loop', async ({ page }) => {
  let logins = 0
  await page.route(`${issuer}/.well-known/openid-configuration`, (route) => route.fulfill({ json: {
    issuer, authorization_endpoint: `${issuer}/protocol/openid-connect/auth`, token_endpoint: `${issuer}/protocol/openid-connect/token`, end_session_endpoint: `${issuer}/protocol/openid-connect/logout`,
  } }))
  await page.route(`${issuer}/protocol/openid-connect/auth?*`, (route) => { logins++; return route.abort() })
  await page.goto('/home#code=invalid&state=unknown')
  await expect(page.getByRole('alert')).toContainText('Não foi possível autenticar')
  await expect(page.getByRole('button', { name: 'Tentar novamente' })).toBeVisible()
  await expect(page.getByRole('navigation')).toHaveCount(0)
  expect(logins).toBe(0)
})
