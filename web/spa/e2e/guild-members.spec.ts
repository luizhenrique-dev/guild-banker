import { expect, test } from '@playwright/test'

test('invitation validates the email in the browser and sends lowercase input', async ({ page }) => {
  const sentEmails: string[] = []
  page.on('request', (request) => {
    if (request.method() === 'POST' && request.url().endsWith('/api/v1/guilds/1/invites')) {
      sentEmails.push((request.postDataJSON() as { email: string }).email)
    }
  })
  await page.goto('/guild?guild=1')
  const input = page.getByRole('textbox', { name: 'Convidar membro' })
  for (const invalid of ['"member"@example.test', 'member..name@example.test']) {
    await input.fill(invalid)
    await expect(page.getByRole('alert')).toHaveText('Informe um email válido.')
    expect(await input.evaluate((element) => (element as HTMLInputElement).validity.patternMismatch)).toBe(true)
    await input.press('Enter')
    expect(sentEmails).toEqual([])
  }
  await input.fill('Member.Name+bank@EXAMPLE.TEST')
  await expect(input).toHaveValue('member.name+bank@example.test')
  await page.getByRole('button', { name: 'Enviar convite' }).click()
  await expect(input).toHaveValue('')
  expect(sentEmails).toEqual(['member.name+bank@example.test'])
})

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

// Existing business journeys begin with an explicit saved guild preference.
test.beforeEach(async ({ page }) => { await page.addInitScript(() => localStorage.setItem('guildbanker:last-guild:demo:1', '1')) })
