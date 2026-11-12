import { guildsFixtures } from '@/src/mocks/fixtures/guilds'
import { useQueryClient, type QueryClient } from '@tanstack/react-query'
import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '@/src/mocks/server'
import { Providers } from '@/src/shared/components/providers'
import { apiClient } from '@/src/shared/lib/api-client'
import { useActiveGuild } from '../use-active-guild'
import { useSession } from '@/src/app/session-provider'
import { GuildSelectionState } from './guild-selection-state'
import i18n from '@/src/shared/lib/i18n'

beforeEach(async () => { await i18n.changeLanguage('pt-BR') })
const owner = { 'X-User-ID': '1', 'X-User-Email': 'ana@familia.com' }
async function invite() { await apiClient('/api/v1/guilds/1/invites', { method: 'POST', headers: owner, body: { email: 'new@example.test' } }) }
function Probe() {
  const { activeGuild } = useActiveGuild()
  const session = useSession()
  if (!session.user) return <h1>Sessão encerrada</h1>
  return <><button onClick={session.logout}>Sair</button>{activeGuild ? <h1>Entrou em {activeGuild.display_name}</h1> : <GuildSelectionState />}</>
}
function setup(children = <Probe />) { return render(<MemoryRouter initialEntries={['/home']}><Providers initialUser={{ id: 8, name: 'New user', email: 'new@example.test' }} demo>{children}</Providers></MemoryRouter>) }
it('shows guild and inviter information and activates the accepted guild', async () => {
  await invite()
  setup()
  expect(await screen.findByRole('heading', { name: 'Seus convites' })).toBeVisible()
  expect(screen.getByText('Família Silva')).toBeVisible()
  expect(screen.getByText(/Ana Silva/)).toBeVisible()
  expect(screen.getByText(/ana@familia.com/)).toBeVisible()
  await userEvent.click(screen.getByRole('button', { name: 'Aceitar' }))
  expect(await screen.findByRole('heading', { name: 'Entrou em Família Silva' })).toBeVisible()
  expect(localStorage.getItem('guildbanker:last-guild:demo:8')).toBe('1')
})
it('declines the last invitation then creates and activates a guild', async () => {
  await invite()
  setup()
  await userEvent.click(await screen.findByRole('button', { name: 'Negar' }))
  expect(await screen.findByRole('heading', { name: 'Crie sua guild' })).toBeVisible()
  const user = userEvent.setup()
  await user.type(screen.getByRole('textbox', { name: 'Nome de exibição' }), 'Minha Família')
  await user.type(screen.getByRole('textbox', { name: 'Descrição (opcional)' }), 'Nossa casa')
  await user.click(screen.getByRole('button', { name: 'Criar guild' }))
  expect(await screen.findByRole('heading', { name: 'Entrou em Minha Família' })).toBeVisible()
})
it('preserves form values and never retries a failed creation automatically', async () => {
  let calls = 0
  server.use(http.post('*/api/v1/guilds', () => { calls++; return new HttpResponse(null, { status: 500 }) }))
  setup()
  const input = await screen.findByRole('textbox', { name: 'Nome de exibição' })
  await userEvent.type(input, 'Draft')
  await userEvent.click(screen.getByRole('button', { name: 'Criar guild' }))
  expect(await screen.findByRole('alert')).toBeVisible()
  expect(input).toHaveValue('Draft')
  expect(calls).toBe(1)
})
it('does not offer creation when invitation loading fails', async () => {
  server.use(http.get('*/api/v1/guild-invites', () => new HttpResponse(null, { status: 403 })))
  setup()
  expect(await screen.findByRole('heading', { name: 'Acesso negado' })).toBeVisible()
  expect(screen.queryByRole('button', { name: 'Criar guild' })).toBeNull()
})
it('ignores a creation response arriving after logout', async () => {
  let finish!: (response: Response) => void
  server.use(http.post('*/api/v1/guilds', async () => new Promise<Response>((resolve) => { finish = resolve })))
  setup()
  await userEvent.type(await screen.findByRole('textbox', { name: 'Nome de exibição' }), 'Draft')
  await userEvent.click(screen.getByRole('button', { name: 'Criar guild' }))
  await waitFor(() => expect(finish).toBeTypeOf('function'))
  await userEvent.click(screen.getByRole('button', { name: 'Sair' }))
  finish(HttpResponse.json({ id: 10, name: 'draft', display_name: 'Draft', description: '', enabled: true }, { status: 201 }))
  expect(await screen.findByRole('heading', { name: 'Sessão encerrada' })).toBeVisible()
  expect(localStorage.getItem('guildbanker:last-guild:demo:8')).toBeNull()
})

it('lists all invitations and keeps the remaining one after a refusal', async () => {
  await invite()
  await apiClient('/api/v1/guilds/2/invites', { method: 'POST', headers: owner, body: { email: 'new@example.test' } })
  setup()
  expect(await screen.findByRole('article', { name: 'Família Silva' })).toBeVisible()
  expect(screen.getByRole('article', { name: 'República dos Amigos' })).toBeVisible()
  await userEvent.click(screen.getAllByRole('button', { name: 'Negar' })[0]!)
  await waitFor(() => expect(screen.queryByRole('article', { name: 'Família Silva' })).toBeNull())
  expect(screen.queryByRole('button', { name: 'Criar guild' })).toBeNull()
  await userEvent.click(screen.getByRole('button', { name: 'Aceitar' }))
  expect(await screen.findByRole('heading', { name: 'Entrou em República dos Amigos' })).toBeVisible()
})
it('validates the display name length before sending creation', async () => {
  let calls = 0
  server.use(http.post('*/api/v1/guilds', () => { calls++; return new HttpResponse(null, { status: 500 }) }))
  setup()
  const input = await screen.findByRole('textbox', { name: 'Nome de exibição' })
  await userEvent.type(input, 'á'.repeat(71))
  await userEvent.click(screen.getByRole('button', { name: 'Criar guild' }))
  expect(await screen.findByRole('alert')).toHaveTextContent('Use no máximo 70 caracteres.')
  expect(calls).toBe(0)
})

it.each(['create', 'accept'])('does not overwrite an invitation 403 with a late %s success', async (action) => {
  if (action === 'accept') await invite()
  let finish!: (response: Response) => void
  let client!: QueryClient
  function WithCache() { client = useQueryClient(); return <Probe /> }
  server.use(http.post(action === 'create' ? '*/api/v1/guilds' : '*/api/v1/guild-invites/:id/accept', async () => new Promise<Response>((resolve) => { finish = resolve })))
  setup(<WithCache />)
  if (action === 'create') {
    await userEvent.type(await screen.findByRole('textbox', { name: 'Nome de exibição' }), 'Draft')
    await userEvent.click(screen.getByRole('button', { name: 'Criar guild' }))
  } else await userEvent.click(await screen.findByRole('button', { name: 'Aceitar' }))
  await waitFor(() => expect(finish).toBeTypeOf('function'))
  server.use(http.get('*/api/v1/guild-invites', () => new HttpResponse(null, { status: 403 })))
  await act(async () => { await client.invalidateQueries({ queryKey: ['guild-invites', 8] }) })
  expect(await screen.findByRole('heading', { name: 'Acesso negado' })).toBeVisible()
  finish(HttpResponse.json({ id: 10, name: 'draft', display_name: 'Draft', description: '', enabled: true }, { status: 201 }))
  await waitFor(() => expect(client.isMutating()).toBe(0))
  expect(client.getQueryState(['guild-invites', 8])?.error).toMatchObject({ status: 403 })
  expect(client.getQueryData(['guilds', 8])).toEqual([])
  expect(localStorage.getItem('guildbanker:last-guild:demo:8')).toBeNull()
})

it('disambiguates invitations with repeated visual names for sighted and screen reader users', async () => {
  server.use(http.get('*/api/v1/guild-invites', () => HttpResponse.json(guildsFixtures.map((guild, index) => ({ id: index+1, guild: { ...guild, display_name: 'Família' }, invited_by: { name: 'Ana', email: 'ana@example.test' } })))))
  setup()
  expect(await screen.findByRole('article', { name: 'Família (familia-silva)' })).toHaveTextContent('familia-silva')
  expect(screen.getByRole('article', { name: 'Família (republica-amigos)' })).toHaveTextContent('republica-amigos')
})
