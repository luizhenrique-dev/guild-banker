import { expect, test } from '@playwright/test'

test.use({ timezoneId: 'America/Bahia' })

test.beforeEach(async ({ page }) => {
  await page.clock.setFixedTime(new Date('2026-11-15T02:30:00Z')) // Still November 14 locally.
  await page.addInitScript(() => localStorage.setItem('guildbanker:last-guild:demo:1', '1'))
})

for (const width of [375, 1280]) {
  test(`category card shows five rows and scrolls the remaining categories at ${width}px`, async ({ page }) => {
    await page.setViewportSize({ width, height: 900 })
    await page.addInitScript(() => {
      const original = window.fetch.bind(window)
      window.fetch = async (input, options) => {
        const response = await original(input, options)
        const path = new URL(input instanceof Request ? input.url : String(input), location.origin).pathname
        if (!path.endsWith('/fixed-expenses')) return response
        return new Response(JSON.stringify(['HOUSING', 'GROCERY', 'UTILITIES', 'INSURANCE', 'TRANSPORTATION', 'HEALTH', 'OTHER'].map((category, index) => ({
          id: index + 1, guild_id: 1, name: category, amount: `${700 - index * 100}.00`, due_day: 1, category, status: 'ACTIVE', responsible_user_id: null,
        }))), { headers: { 'Content-Type': 'application/json' } })
      }
    })
    await page.goto('/home?guild=1')
    const card = page.getByRole('region', { name: 'Despesas fixas - Por categoria' })
    const list = card.getByRole('list')
    await expect(list.getByRole('listitem')).toHaveCount(7)
    const visibleRows = await list.evaluate((element) => {
      const box = element.getBoundingClientRect()
      return [...element.children].filter((child) => {
        const row = child.getBoundingClientRect()
        return row.top >= box.top && row.bottom <= box.bottom + 1
      }).length
    })
    expect(visibleRows).toBe(5)
    await list.focus()
    await page.keyboard.press('End')
    await expect.poll(() => list.evaluate((element) => element.scrollTop)).toBeGreaterThan(0)
    expect(await list.evaluate((element) => element.clientHeight)).toBe(280)
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
  })

  test(`shows the five nearest local monthly expenses without overflow at ${width}px`, async ({ page }) => {
    await page.setViewportSize({ width, height: 900 })
    await page.goto('/fixed-expenses?guild=1')
    // Create through the existing UI so the browser exercises the real shared query and mock API.
    for (const [name, day] of [['Telefone', 14], ['Condomínio', 15], ['Energia', 16], ['Água', 18], ['Outra conta', 25]] as const) {
      await page.getByRole('button', { name: 'Nova despesa fixa', exact: true }).click()
      const dialog = page.getByRole('dialog')
      await dialog.getByRole('textbox', { name: 'Nome', exact: true }).fill(name)
  await page.getByRole('dialog').getByRole('combobox', { name: 'Forma de pagamento', exact: true }).selectOption('CREDIT_CARD')
      await dialog.getByRole('textbox', { name: 'Valor (R$)' }).fill('123,45')
      await dialog.getByRole('spinbutton', { name: 'Dia de vencimento' }).fill(String(day))
      await dialog.getByRole('button', { name: 'Salvar', exact: true }).click()
      await expect(dialog).toHaveCount(0)
    }
    await page.getByRole('link', { name: 'Início', exact: true }).click()
    const section = page.getByRole('region', { name: 'Despesas fixas - Próximos vencimentos' })
    await expect(section.getByRole('listitem')).toHaveCount(5)
    await expect(section.getByRole('listitem').locator('> p')).toHaveText(['Telefone', 'Netflix', 'Condomínio', 'Energia', 'Água'])
    await expect(section.getByText('Hoje', { exact: true })).toHaveCount(1)
    await expect(section.getByText('Amanhã', { exact: true })).toHaveCount(2)
    await expect(section.getByText('14/11/2026')).toBeVisible()
    await expect(section.getByRole('listitem').first()).toContainText('123,45')
    const categories = page.getByRole('region', { name: 'Despesas fixas - Por categoria' })
    await expect(categories.getByRole('listitem').first()).toContainText('Moradia')
    await expect(categories.getByRole('listitem').first()).toContainText('2.500,00')
    const upcomingBox = await section.boundingBox()
    const categoryBox = await categories.boundingBox()
    expect(Math.abs(upcomingBox!.height - categoryBox!.height)).toBeLessThan(1)
    await expect(section.getByRole('listitem').first().locator('> p')).toHaveCSS('font-size', '14px')
    await expect(categories.getByRole('listitem').first().locator('> p').first()).toHaveCSS('font-size', '14px')
    expect(upcomingBox!.width).toBeLessThanOrEqual(440)
    expect(categoryBox!.width).toBeLessThanOrEqual(440)
    if (width === 1280) {
      expect(Math.abs(upcomingBox!.y - categoryBox!.y)).toBeLessThan(1)
      expect(categoryBox!.x).toBeGreaterThan(upcomingBox!.x + upcomingBox!.width)
    } else expect(categoryBox!.y).toBeGreaterThan(upcomingBox!.y + upcomingBox!.height)
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
    await page.getByRole('button', { name: 'Família Silva', exact: true }).click()
    await page.getByRole('button', { name: 'República dos Amigos', exact: true }).click()
    await expect(section.getByText('Nenhum próximo gasto fixo neste mês.')).toBeVisible()
    await expect(section.getByRole('listitem')).toHaveCount(0)
    await expect(categories.getByRole('listitem')).toHaveCount(1)
    await expect(categories.getByRole('listitem')).toContainText('Outros')
    await expect(categories.getByRole('listitem')).toContainText('150,00')
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
