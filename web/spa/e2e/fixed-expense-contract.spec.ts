import { expect, test } from '@playwright/test'

test('name stays read-only and PATCH sends only supported fields', async ({ page }) => {
  await page.goto('/fixed-expenses')
  await page.getByRole('row').filter({ hasText: 'Aluguel' }).getByRole('button', { name: 'Editar Aluguel' }).click()
  await expect(page.getByRole('textbox', { name: 'Nome', exact: true })).toHaveAttribute('readonly', '')
  const update = page.waitForRequest((request) => /fixed-expenses\/1$/.test(request.url()) && request.method() === 'PATCH')
  await page.getByRole('button', { name: 'Salvar', exact: true }).click()
  expect((await update).postDataJSON()).toEqual({ amount: '2500.00', due_day: 10, category: 'HOUSING', responsible_user_id: null })
  await expect(page.getByRole('dialog')).toHaveCount(0)
})

// Existing business journeys begin with an explicit saved guild preference.
test.beforeEach(async ({ page }) => { await page.addInitScript(() => localStorage.setItem('guildbanker:last-guild:demo:1', '1')) })
