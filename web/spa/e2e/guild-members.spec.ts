import { expect, test } from '@playwright/test'

test('demo invitation and member removal report actual mutation outcomes', async ({ page }) => {
  await page.goto('/guild?guild=1')
  await expect(page.getByRole('heading', { name: 'Membros demonstrativos' })).toBeVisible()
  await page.getByRole('textbox', { name: 'Convidar membro' }).fill('new@example.com')
  await page.getByRole('button', { name: 'Enviar convite' }).click()
  await expect(page.getByRole('textbox', { name: 'Convidar membro' })).toHaveValue('')
  await page.getByRole('button', { name: 'Remover Carlos Silva' }).click()
  await expect(page.getByText('Carlos Silva')).toHaveCount(0)
  await page.getByRole('button', { name: 'Família Silva', exact: true }).click()
  await page.getByRole('button', { name: 'República dos Amigos', exact: true }).click()
  await expect(page.getByText('Bruno Lima')).toBeVisible()
  await expect(page.getByText('Maria Silva')).toHaveCount(0)
})
