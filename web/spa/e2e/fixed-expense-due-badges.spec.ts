import { expect, test } from '@playwright/test'

test.beforeEach(async ({ page }) => {
  await page.addInitScript(() => localStorage.setItem('guildbanker:last-guild:demo:1', '1'))
  await page.clock.setFixedTime(new Date(2026, 10, 15, 12))
})

for (const width of [375, 1280]) {
  test(`due badges stay within the title column at ${width}px`, async ({ page }) => {
    await page.setViewportSize({ width, height: 900 })
    await page.goto('/fixed-expenses?guild=1')
    const today = page.getByRole('row').filter({ hasText: 'Netflix' })
    const soon = page.getByRole('row').filter({ hasText: 'Seguro auto' })
    await expect(today.getByRole('cell').first().getByText('Vence hoje', { exact: true })).toBeVisible()
    await expect(soon.getByRole('cell').first().getByText('Próximo do vencimento', { exact: true })).toBeVisible()
    const titles = page.getByRole('table').locator('tbody tr td:first-child span[title]')
    await expect(titles).toHaveText(['Netflix', 'Seguro auto', 'Academia', 'Plano de saúde', 'Aluguel'])
    await expect(page.getByRole('row').filter({ hasText: 'Academia' })).not.toContainText('Vence')

    await page.getByRole('button', { name: 'Nova despesa fixa', exact: true }).click()
    const dialog = page.getByRole('dialog')
    const name = 'Internet residencial com um título longo para validar o espaço disponível na coluna da despesa fixa'
    await dialog.getByRole('textbox', { name: 'Nome', exact: true }).fill(name)
    await dialog.getByRole('textbox', { name: 'Valor (R$)' }).fill('100,00')
    await dialog.getByRole('spinbutton', { name: 'Dia de vencimento' }).fill('16')
    await dialog.getByRole('combobox', { name: 'Forma de pagamento' }).selectOption('CREDIT_CARD')
    await dialog.getByRole('button', { name: 'Salvar', exact: true }).click()
    const tomorrow = page.getByRole('row').filter({ hasText: name })
    await expect(tomorrow.getByRole('cell').first().getByText('Vence amanhã', { exact: true })).toBeVisible()
    await expect(titles.nth(1)).toHaveText(name)
    await page.evaluate(() => document.fonts.ready)
    expect(await page.evaluate(() => document.fonts.check('600 14px Inter'))).toBe(true)

    for (const [row, label] of [[today, 'Vence hoje'], [tomorrow, 'Vence amanhã'], [soon, 'Próximo do vencimento']] as const) {
      await expect(row.getByRole('cell')).toHaveCount(8)
      const cell = row.getByRole('cell').first()
      const badge = cell.getByText(label, { exact: true })
      const cellBox = (await cell.boundingBox())!
      const badgeBox = (await badge.boundingBox())!
      expect(Math.abs(cellBox.x + cellBox.width - 24 - badgeBox.x - badgeBox.width)).toBeLessThan(1)
      const titleBox = (await cell.locator('span[title]').boundingBox())!
      expect(titleBox.x + titleBox.width).toBeLessThanOrEqual(badgeBox.x - 12)
      expect(Math.abs(titleBox.y + titleBox.height / 2 - badgeBox.y - badgeBox.height / 2)).toBeLessThan(1)
    }
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
    await page.screenshot({ path: `test-results/fixed-expense-due-badges-${width}.png`, fullPage: true })
    await today.getByRole('button', { name: 'Pausar Netflix' }).click()
    await expect(today).toContainText('Pausada')
    await expect(today.getByText('Vence hoje', { exact: true })).toHaveCount(0)
    await page.getByRole('button', { name: 'EN', exact: true }).click()
    await expect(tomorrow).toContainText('Due tomorrow')
    await expect(soon).toContainText('Due soon')
  })
}
