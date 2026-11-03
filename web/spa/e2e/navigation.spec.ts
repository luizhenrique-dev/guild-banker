import { expect, test } from '@playwright/test'

for (const width of [375, 1280]) {
  test(`four routes, history and keyboard at ${width}px`, async ({ page }) => {
    await page.setViewportSize({ width, height: 900 })
    await page.goto('/')
    await expect(page).toHaveURL(/\/home/)
    await expect(page.getByRole('heading', { name: 'Bem-vindo(a) Ana Silva' })).toBeVisible()
    await page.getByRole('button', { name: 'Família Silva', exact: true }).click()
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
    await page.getByRole('button', { name: 'República dos Amigos' }).focus()
    await page.keyboard.press('Escape')
    await expect(page.getByRole('button', { name: 'República dos Amigos' })).toHaveCount(0)
    const routes = [
      ['Transações', 'Transações'], ['Importar', 'Importar extrato'],
      ['Guild / Membros', 'Guild / Membros'], ['Despesas fixas', 'Minhas despesas fixas'],
    ] as const
    for (const [link, heading] of routes) {
      const target = page.getByRole('navigation').getByRole('link', { name: link, exact: true })
      await target.focus()
      await page.keyboard.press('Enter')
      await expect(page.getByRole('heading', { name: heading, exact: true })).toBeVisible()
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
    }
    await page.goBack()
    await expect(page.getByRole('heading', { name: 'Guild / Membros', exact: true })).toBeVisible()
    await page.goto('/missing')
    await expect(page.getByRole('heading', { name: 'Página não encontrada' })).toBeVisible()
  })
}

for (const width of [375, 1280]) {
  test(`forbidden loading preserves navigation at ${width}px`, async ({ page }) => {
    await page.setViewportSize({ width, height: 900 })
    await page.addInitScript(() => {
      const original = window.fetch.bind(window)
      window.fetch = (input, options) => {
        const path = new URL(input instanceof Request ? input.url : String(input), location.origin).pathname
        if (path === '/api/v1/fixed-expenses') return Promise.resolve(new Response('{"error":"forbidden"}', { status: 403, headers: { 'Content-Type': 'application/json' } }))
        return original(input, options)
      }
    })
    await page.goto('/fixed-expenses')
    await expect(page.getByRole('heading', { name: 'Acesso negado' })).toBeVisible()
    await expect(page.getByRole('table')).toHaveCount(0)
    await expect(page.getByRole('navigation')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Sair' })).toBeEnabled()
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
    await page.getByRole('link', { name: 'Voltar ao início' }).focus()
    await page.keyboard.press('Enter')
    await expect(page.getByRole('heading', { name: 'Bem-vindo(a) Ana Silva' })).toBeVisible()
    await page.getByRole('button', { name: 'Sair' }).click()
    await expect(page.getByRole('navigation')).toHaveCount(0)
  })
}
