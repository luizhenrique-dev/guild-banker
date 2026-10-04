import { expect, test } from '@playwright/test'

test('uploads CSV as multipart and confirms using the handler route', async ({ page }) => {
  await page.goto('/import')
  await page.locator('input[type=file]').setInputFiles({ name: 'statement.csv', mimeType: 'text/csv', buffer: Buffer.from('Data,Descrição,Valor\n2026-07-15,Compra,149.90') })
  const upload = page.waitForResponse((response) => response.url().endsWith('/imports') && response.request().method() === 'POST')
  await page.getByRole('button', { name: 'Enviar arquivo' }).click()
  const response = await upload
  expect(response.status()).toBe(201)
  expect(response.request().headers()['content-type']).toContain('multipart/form-data; boundary=')
  await expect(page.getByRole('heading', { name: 'Revisar itens importados' })).toBeVisible()
  const confirmation = page.waitForResponse((response) => response.url().endsWith('/imports/1/confirm'))
  await page.getByRole('button', { name: 'Confirmar importação' }).click()
  expect((await confirmation).status()).toBe(200)
  await expect(page.getByRole('heading', { name: 'Revisar itens importados' })).toHaveCount(0)
})
