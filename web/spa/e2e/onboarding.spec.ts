import { expect, test, type Page } from '@playwright/test'
import { guildsFixtures } from '../src/mocks/fixtures/guilds'
import type { components } from '../src/shared/api/schema'

type Guild = components['schemas']['GuildResponse']
type Invitation = components['schemas']['InvitationResponse']
async function scenario(page: Page, invitations: Invitation[] = [], failCreateOnce = false) {
  await page.addInitScript((initial) => {
    const state = JSON.parse(sessionStorage.getItem('issue5-e2e-state') || JSON.stringify(initial)) as typeof initial
    const save = () => sessionStorage.setItem('issue5-e2e-state', JSON.stringify(state))
    const original = window.fetch.bind(window)
    window.fetch = async (input, options) => {
      const path = new URL(input instanceof Request ? input.url : String(input), location.origin).pathname
      const method = options?.method ?? (input instanceof Request ? input.method : 'GET')
      if (path.startsWith('/api/')) {
        const requests = JSON.parse(sessionStorage.getItem('issue5-e2e-requests') || '[]') as string[]
        requests.push(path)
        sessionStorage.setItem('issue5-e2e-requests', JSON.stringify(requests))
      }
      const json = (value: unknown, status = 200) => new Response(JSON.stringify(value), { status, headers: { 'Content-Type': 'application/json' } })
      if (path === '/api/v1/guilds' && method === 'GET') return json(state.guilds)
      if (path === '/api/v1/guild-invites') return json(state.invitations)
      if (path === '/api/v1/guilds' && method === 'POST') {
        state.attempts++
        save()
        if (state.failCreateOnce && state.attempts === 1) return json({ error: 'unavailable' }, 500)
        const body = JSON.parse(String(options?.body)) as { display_name: string; description?: string }
        const guild: Guild = { id: 50, name: 'new-guild', display_name: body.display_name, description: body.description ?? '', enabled: true }
        state.guilds.push(guild)
        save()
        return json(guild, 201)
      }
      const response = path.match(/^\/api\/v1\/guild-invites\/(\d+)\/(accept|decline)$/)
      if (response && method === 'POST') {
        const invitation = state.invitations.find((invite) => invite.id === Number(response[1]))!
        state.invitations = state.invitations.filter((invite) => invite.id !== invitation.id)
        if (response[2] === 'accept') state.guilds.push(invitation.guild)
        save()
        return response[2] === 'accept' ? json(invitation.guild) : new Response(null, { status: 204 })
      }
      return original(input, options)
    }
  }, { guilds: [] as Guild[], invitations, failCreateOnce, attempts: 0 })
}

test('chooses a guild and restores it after logout and another demo login', async ({ page }) => {
  await page.goto('/home')
  await expect(page.getByRole('heading', { name: 'Escolha sua guild' })).toBeVisible()
  await expect(page.getByRole('navigation')).toHaveCount(0)
  await page.getByRole('button', { name: 'Entrar em República dos Amigos' }).click()
  await expect(page.getByRole('heading', { name: 'Bem-vindo(a) Ana Silva' })).toBeVisible()
  await page.getByRole('button', { name: 'Sair' }).click()
  await expect(page.getByRole('heading', { name: 'Sessão encerrada' })).toBeVisible()
  await page.goto('/home')
  await expect(page.getByRole('button', { name: 'República dos Amigos', exact: true })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Escolha sua guild' })).toHaveCount(0)
  await expect(page).toHaveURL(/guild=2/)
})

test('lists invitations, declines one and accepts the other before loading business data', async ({ page }) => {
  await scenario(page, guildsFixtures.map((guild, index) => ({ id: index+1, guild, invited_by: { name: 'Ana', email: 'inviter@example.test' } })))
  await page.goto('/transactions?category=OTHER')
  await expect(page.getByRole('heading', { name: 'Seus convites' })).toBeVisible()
  await expect(page.getByRole('article')).toHaveCount(2)
  const requests = await page.evaluate(() => JSON.parse(sessionStorage.getItem('issue5-e2e-requests') || '[]') as string[])
  expect(requests.some((path) => /\/guilds\/\d+\/transactions/.test(path))).toBe(false)
  await page.getByRole('article', { name: 'Família Silva' }).getByRole('button', { name: 'Negar' }).click()
  await expect(page.getByRole('article')).toHaveCount(1)
  await page.getByRole('button', { name: 'Aceitar' }).click()
  await expect(page.getByRole('heading', { name: 'Transações', exact: true })).toBeVisible()
  await expect(page).toHaveURL(/category=OTHER&guild=2/)
  await expect(page.getByRole('button', { name: 'República dos Amigos', exact: true })).toBeVisible()
})

for (const width of [375, 1280]) {
  test(`creates a guild after refusal with keyboard, retry and layout at ${width}px`, async ({ page }) => {
    await page.setViewportSize({ width, height: 900 })
    await scenario(page, [{ id: 1, guild: guildsFixtures[0]!, invited_by: { name: 'Ana', email: 'inviter@example.test' } }], true)
    await page.goto('/home')
    await page.getByRole('button', { name: 'Negar' }).click()
    await expect(page.getByRole('heading', { name: 'Crie sua guild' })).toBeVisible()
    const name = page.getByRole('textbox', { name: 'Nome de exibição' })
    await name.fill('Minha casa')
    await name.press('Tab')
    await expect(page.getByRole('textbox', { name: 'Descrição (opcional)' })).toBeFocused()
    await page.getByRole('button', { name: 'Criar guild' }).click()
    await expect(page.getByRole('alert')).toContainText('Não foi possível concluir')
    await expect(name).toHaveValue('Minha casa')
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
    await page.getByRole('button', { name: 'Criar guild' }).click()
    await expect(page.getByRole('heading', { name: 'Bem-vindo(a) Ana Silva' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Minha casa', exact: true })).toBeVisible()
    await expect(page).toHaveURL(/guild=50/)
  })
}
