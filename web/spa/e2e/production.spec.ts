import { expect, test } from '@playwright/test'

for (const route of ['/fixed-expenses', '/transactions', '/import', '/guild']) {
  test(`compiled ${route} loads directly without demo, API calls or mock worker`, async ({ page, context }) => {
    const requests: string[] = []
    const failedAssets: string[] = []
    const errors: string[] = []
    page.on('request', (request) => { requests.push(request.url()) })
    page.on('response', (response) => { if (response.url().includes('/assets/') && !response.ok()) failedAssets.push(response.url()) })
    page.on('requestfailed', (request) => { failedAssets.push(request.url()) })
    page.on('pageerror', (error) => { errors.push(error.message) })
    const response = await page.goto(route)
    expect(response?.status()).toBe(200)
    await expect(page.getByRole('heading', { name: 'Sessão não integrada' })).toBeVisible()
    await page.evaluate(() => document.fonts.ready)
    expect(requests.some((url) => /\/assets\/.*\.js/.test(url))).toBe(true)
    expect(requests.some((url) => /\/assets\/.*\.css/.test(url))).toBe(true)
    expect(requests.filter((url) => /mockServiceWorker|\/api\//.test(url))).toEqual([])
    expect(context.serviceWorkers()).toEqual([])
    expect(await page.evaluate(() => navigator.serviceWorker.getRegistrations().then((registrations) => registrations.length))).toBe(0)
    await expect(page.getByText('Ana Silva', { exact: true })).toHaveCount(0)
    await expect(page.getByText('Família Silva', { exact: true })).toHaveCount(0)
    await expect(page.getByText('Aluguel', { exact: true })).toHaveCount(0)
    expect(failedAssets).toEqual([])
    expect(errors).toEqual([])
  })
}
test('compiled root redirects and unknown UI route is visible', async ({ page }) => {
  await page.goto('/')
  await expect(page).toHaveURL(/\/fixed-expenses$/)
  await expect(page.getByRole('heading', { name: 'Sessão não integrada' })).toBeVisible()
  await page.goto('/route-does-not-exist')
  await expect(page.getByRole('heading', { name: 'Página não encontrada' })).toBeVisible()
})
