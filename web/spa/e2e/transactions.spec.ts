import { expect, test } from '@playwright/test'

test('filters restore with history and mutations update the selected guild', async ({ page }) => {
  await page.goto('/transactions?guild=1')
  await expect(page.getByText('Supermercado Extra')).toBeVisible()
  await page.getByRole('combobox', { name: 'Tipo', exact: true }).selectOption('INCOME')
  await expect(page).toHaveURL(/type=INCOME/)
  await expect(page.getByText('Supermercado Extra')).toHaveCount(0)
  await page.goBack()
  const row = page.getByRole('row').filter({ hasText: 'Supermercado Extra' })
  await expect(row).toBeVisible()
  await row.getByTitle('Pública').click()
  await expect(row.getByTitle('Privada')).toBeVisible()
  page.once('dialog', (dialog) => dialog.accept())
  await row.getByTitle('Cancelar transação').click()
  await expect(row).toHaveCount(0)
})
