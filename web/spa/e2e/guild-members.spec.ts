import { expect, test } from '@playwright/test'

test('lists pending invitations separately and cancels them while retaining all current members', async ({ page }) => {
  await page.setViewportSize({ width: 375, height: 900 })
  await page.goto('/guild?guild=1')
  const members = page.getByRole('region', { name: 'Membros', exact: true })
  await expect(members.getByText('Ana Silva', { exact: true })).toBeVisible()
  await expect(members.getByText('Você', { exact: true })).toBeVisible()
  await expect(members.getByRole('button', { name: 'Remover Ana Silva' })).toHaveCount(0)
  await expect(members.getByText('Último acesso', { exact: true })).toHaveCount(3)
  await expect(members.getByText('Membro desde', { exact: true })).toHaveCount(3)
  const input = page.getByRole('textbox', { name: 'Convidar membro' })
  await input.fill('pending@example.test')
  await page.getByRole('button', { name: 'Enviar convite' }).click()
  await expect(input).toHaveValue('')
  const invitations = page.getByRole('region', { name: 'Convites em aberto' })
  await expect(invitations.getByText('pending@example.test', { exact: true })).toBeVisible()
  await expect(invitations.getByText('Convidado por: Ana Silva (ana@familia.com)')).toBeVisible()
  await expect(invitations.locator('time')).toHaveAttribute('datetime', /T/)
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
  await invitations.getByRole('button', { name: 'Cancelar convite para pending@example.test' }).click()
  await page.getByRole('alertdialog').getByRole('button', { name: 'Confirmar cancelamento' }).click()
  await expect(invitations.getByText('Nenhum convite em aberto.')).toBeVisible()
  await expect(members.getByText('Ana Silva', { exact: true })).toBeVisible()
  await input.fill('pending@example.test')
  await page.getByRole('button', { name: 'Enviar convite' }).click()
  await expect(invitations.getByText('pending@example.test', { exact: true })).toBeVisible()
})

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
  await expect(page.getByRole('heading', { name: 'Membros', exact: true })).toBeVisible()
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
