import { expect, test } from '@playwright/test'

for (const width of [375, 1280]) {
  test(`four routes, history and keyboard at ${width}px`, async ({ page }) => {
    await page.setViewportSize({ width, height: 900 })
    await page.goto('/')
    await expect(page).toHaveURL(/\/fixed-expenses/)
    await expect(page.getByRole('heading', { name: 'Minhas despesas fixas' })).toBeVisible()
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
