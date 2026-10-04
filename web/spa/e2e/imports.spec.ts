import { expect, test } from '@playwright/test'

test('uploads CSV as multipart and confirms using the handler route', async ({ page }) => {
  await page.goto('/transactions?guild=1')
  await expect(page.getByText('Supermercado Extra')).toBeVisible()
  await page.getByRole('navigation').getByRole('link', { name: 'Importar', exact: true }).click()
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
  await page.getByRole('navigation').getByRole('link', { name: 'Transações', exact: true }).click()
  await expect(page.getByText('MERCADO LIVRE *COMPRA', { exact: true })).toBeVisible()
})

test('changing guild discards review and selected file', async ({ page }) => {
  await page.goto('/import?guild=1')
  await page.locator('input[type=file]').setInputFiles({ name: 'statement.csv', mimeType: 'text/csv', buffer: Buffer.from('csv') })
  await page.getByRole('button', { name: 'Enviar arquivo' }).click()
  await expect(page.getByRole('heading', { name: 'Revisar itens importados' })).toBeVisible()
  await page.getByRole('button', { name: 'Família Silva', exact: true }).click()
  await page.getByRole('button', { name: 'República dos Amigos', exact: true }).click()
  await expect(page.getByRole('heading', { name: 'Revisar itens importados' })).toHaveCount(0)
  await expect(page.getByRole('button', { name: 'Enviar arquivo' })).toHaveCount(0)
  await page.getByRole('button', { name: 'República dos Amigos', exact: true }).click()
  await page.getByRole('button', { name: 'Família Silva', exact: true }).click()
  await expect(page.getByRole('heading', { name: 'Revisar itens importados' })).toHaveCount(0)
})
