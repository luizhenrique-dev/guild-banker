import { expect, test } from '@playwright/test'

test('URL guild selection isolates data, restores history and rejects invalid IDs', async ({ page }) => {
  await page.goto('/transactions?guild=1')
  await expect(page.getByText('Supermercado Extra')).toBeVisible()
  await page.getByRole('button', { name: 'Família Silva', exact: true }).click()
  await page.getByRole('button', { name: 'República dos Amigos', exact: true }).click()
  await expect(page).toHaveURL(/guild=2/)
  await expect(page.getByText('Compras da República')).toBeVisible()
  await expect(page.getByText('Supermercado Extra')).toHaveCount(0)
  await page.goBack()
  await expect(page.getByText('Supermercado Extra')).toBeVisible()
  const requests: string[] = []
  page.on('request', (request) => { if (/\/transactions(?:\?|$)/.test(request.url()) && request.url().includes('/api/')) requests.push(request.url()) })
  for (const guild of ['999', '-1', '9007199254740992', 'abc']) {
    await page.goto('/transactions?guild=' + guild)
    await expect(page.getByText('Selecione uma guild válida.')).toBeVisible()
  }
  expect(requests).toEqual([])
})

test('expenses close the old guild modal and preserve selection through history', async ({ page }) => {
  await page.goto('/fixed-expenses?guild=1')
  await expect(page.getByText('Aluguel', { exact: true })).toBeVisible()
  await page.getByRole('button', { name: 'Nova despesa fixa' }).click()
  await expect(page.getByRole('dialog')).toBeVisible()
  await page.evaluate(() => { history.pushState({}, '', '/fixed-expenses?guild=2'); dispatchEvent(new PopStateEvent('popstate')) })
  await expect(page).toHaveURL(/guild=2/)
  await expect(page.getByText('Internet da República', { exact: true })).toBeVisible()
  await expect(page.getByText('Aluguel', { exact: true })).toHaveCount(0)
  await expect(page.getByRole('dialog')).toHaveCount(0)
  await page.goBack()
  await expect(page).toHaveURL(/guild=1/)
  await expect(page.getByText('Aluguel', { exact: true })).toBeVisible()
  await expect(page.getByText('Internet da República', { exact: true })).toHaveCount(0)
  await expect(page.getByRole('dialog')).toHaveCount(0)
  await page.goto('/fixed-expenses?guild=999')
  await expect(page.getByText('Selecione uma guild válida.')).toBeVisible()
  await expect(page.getByRole('button', { name: 'Nova despesa fixa' })).toHaveCount(0)
})
