import { expect, test } from '@playwright/test'

test.use({ timezoneId: 'America/Bahia' })

test.beforeEach(async ({ page }) => {
  await page.clock.setFixedTime(new Date('2026-11-15T02:30:00Z')) // Still November 14 locally.
  await page.addInitScript(() => localStorage.setItem('guildbanker:last-guild:demo:1', '1'))
})

for (const width of [375, 1280]) {
  test(`shows the five nearest local monthly expenses without overflow at ${width}px`, async ({ page }) => {
    await page.setViewportSize({ width, height: 900 })
    await page.goto('/fixed-expenses?guild=1')
    // Create through the existing UI so the browser exercises the real shared query and mock API.
    for (const [name, day] of [['Telefone', 14], ['Condomínio', 15], ['Energia', 16], ['Água', 18], ['Outra conta', 25]] as const) {
      await page.getByRole('button', { name: 'Nova despesa fixa', exact: true }).click()
      const dialog = page.getByRole('dialog')
      await dialog.getByRole('textbox', { name: 'Nome', exact: true }).fill(name)
      await dialog.getByRole('textbox', { name: 'Valor (R$)' }).fill('123,45')
      await dialog.getByRole('spinbutton', { name: 'Dia de vencimento' }).fill(String(day))
      await dialog.getByRole('button', { name: 'Salvar', exact: true }).click()
      await expect(dialog).toHaveCount(0)
    }
    await page.getByRole('link', { name: 'Início', exact: true }).click()
    const section = page.getByRole('region', { name: 'Gastos fixos do mês' })
    await expect(section.getByRole('listitem')).toHaveCount(5)
    await expect(section.getByRole('listitem').locator('> p')).toHaveText(['Telefone', 'Netflix', 'Condomínio', 'Energia', 'Água'])
    await expect(section.getByText('Hoje', { exact: true })).toHaveCount(1)
    await expect(section.getByText('Amanhã', { exact: true })).toHaveCount(2)
    await expect(section.getByText('14/11/2026')).toBeVisible()
    await expect(section.getByRole('listitem').first()).toContainText('123,45')
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
    await page.getByRole('button', { name: 'Família Silva', exact: true }).click()
    await page.getByRole('button', { name: 'República dos Amigos', exact: true }).click()
    await expect(section.getByText('Nenhum próximo gasto fixo neste mês.')).toBeVisible()
    await expect(section.getByRole('listitem')).toHaveCount(0)
    await page.goBack()
    await expect(section.getByRole('listitem')).toHaveCount(5)
    await page.clock.setFixedTime(new Date('2026-12-01T12:00:00Z'))
    await page.getByRole('link', { name: 'Despesas fixas', exact: true }).click()
    await page.getByRole('link', { name: 'Início', exact: true }).click()
    await expect(section.getByText('dezembro de 2026')).toBeVisible()
    await expect(section.getByRole('listitem').locator('> p')).toHaveText(['Plano de saúde', 'Aluguel', 'Telefone', 'Netflix', 'Condomínio'])
    await expect(section.getByText('Hoje', { exact: true })).toHaveCount(0)
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
  })
}
