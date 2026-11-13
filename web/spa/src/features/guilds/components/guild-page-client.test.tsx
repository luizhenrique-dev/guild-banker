import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '@/src/mocks/server'
import { App } from '@/src/app/app'
import i18n from '@/src/shared/lib/i18n'
import { toast } from 'sonner'
import { apiClient } from '@/src/shared/lib/api-client'

const ownerHeaders = { 'X-User-ID': '1', 'X-User-Email': 'ana@familia.com' }
function renderGuild() {
  return render(<MemoryRouter initialEntries={['/guild?guild=1']}><App initialUser={{ id: 1, name: 'Ana Silva', email: 'ana@familia.com' }} demo /></MemoryRouter>)
}

it('lists full names, emails, access and membership dates including the current user', async () => {
  renderGuild()
  const section = await screen.findByRole('region', { name: 'Membros' })
  expect(await within(section).findByText('Ana Silva')).toBeVisible()
  expect(within(section).getByText('ana@familia.com')).toBeVisible()
  expect(within(section).getByText('Você')).toBeVisible()
  expect(within(section).queryByRole('button', { name: 'Remover Ana Silva' })).toBeNull()
  expect(within(section).getAllByText('Último acesso')).toHaveLength(3)
  expect(within(section).getAllByText('Membro desde')).toHaveLength(3)
  expect(section.querySelectorAll('time')).toHaveLength(6)
  expect(await screen.findByText('Nenhum convite em aberto.')).toBeVisible()
})

it('shows missing access history explicitly', async () => {
  server.use(http.get('*/api/v1/guilds/1/members', () => HttpResponse.json([{ id: 1, name: 'Ana Silva', email: 'ana@familia.com', last_access_at: null, member_since: '2026-01-10T12:00:00Z' }])))
  renderGuild()
  expect(await screen.findByText('Sem registro')).toBeVisible()
})

it('hides member information when the member query is forbidden', async () => {
  server.use(http.get('*/api/v1/guilds/1/members', () => new HttpResponse(null, { status: 403 })))
  renderGuild()
  expect(await screen.findByRole('heading', { name: 'Acesso negado' })).toBeVisible()
  expect(screen.queryByText('Carlos Silva')).toBeNull()
  expect(screen.queryByRole('region', { name: 'Membros' })).toBeNull()
})

it('refreshes both lists if an invitation was accepted before cancellation', async () => {
  let answered = false
  server.use(
    http.get('*/api/v1/guilds/1/invites', () => HttpResponse.json(answered ? [] : [{ id: 9, email: 'target@example.test', invited_by: { name: 'Ana Silva', email: 'ana@familia.com' }, created_at: '2026-11-12T12:00:00Z' }])),
    http.get('*/api/v1/guilds/1/members', () => HttpResponse.json(answered ? [{ id: 9, name: 'Target Full Name', email: 'target@example.test', last_access_at: null, member_since: '2026-11-13T00:00:00Z' }] : [])),
    http.delete('*/api/v1/guilds/1/invites/9', () => { answered = true; return HttpResponse.json({ error: 'already answered', code: 'invitation_already_answered' }, { status: 409 }) }),
  )
  renderGuild()
  await userEvent.click(await screen.findByRole('button', { name: 'Cancelar convite para target@example.test' }))
  const dialog = await screen.findByRole('alertdialog')
  await userEvent.click(within(dialog).getByRole('button', { name: 'Confirmar cancelamento' }))
  expect(await within(dialog).findByRole('alert')).toHaveTextContent('Este convite já foi respondido ou cancelado.')
  expect(await screen.findByText('Nenhum convite em aberto.')).toBeVisible()
  expect(await screen.findByText('Target Full Name')).toBeVisible()
})

it('keeps the invitation after cancellation failure, then cancels and permits inviting again', async () => {
  await apiClient('/api/v1/guilds/1/invites', { method: 'POST', headers: ownerHeaders, body: { email: 'target@example.test' } })
  renderGuild()
  const section = await screen.findByRole('region', { name: 'Convites em aberto' })
  expect(await within(section).findByText('target@example.test')).toBeVisible()
  expect(within(section).getByText('Convidado por: Ana Silva (ana@familia.com)')).toBeVisible()
  expect(section.querySelector('time')).toHaveAttribute('datetime')
  server.use(http.delete('*/api/v1/guilds/1/invites/:id', () => new HttpResponse(null, { status: 500 })))
  await userEvent.click(within(section).getByRole('button', { name: 'Cancelar convite para target@example.test' }))
  const dialog = await screen.findByRole('alertdialog')
  await userEvent.click(within(dialog).getByRole('button', { name: 'Confirmar cancelamento' }))
  expect(await within(dialog).findByRole('alert')).toHaveTextContent('Não foi possível cancelar o convite.')
  expect(within(section).getByText('target@example.test')).toBeVisible()
  server.resetHandlers()
  await userEvent.click(within(dialog).getByRole('button', { name: 'Confirmar cancelamento' }))
  expect(await screen.findByText('Nenhum convite em aberto.')).toBeVisible()
  expect(screen.queryByRole('alertdialog')).toBeNull()
  await userEvent.type(screen.getByRole('textbox', { name: 'Convidar membro' }), 'target@example.test')
  await userEvent.click(screen.getByRole('button', { name: 'Enviar convite' }))
  expect(await within(section).findByText('target@example.test')).toBeVisible()
})

it('retries members and invitations independently without replacing failures with empty states', async () => {
  server.use(http.get('*/api/v1/guilds/1/members', () => new HttpResponse(null, { status: 500 })), http.get('*/api/v1/guilds/1/invites', () => new HttpResponse(null, { status: 500 })))
  renderGuild()
  const members = await screen.findByRole('region', { name: 'Membros' })
  const invitations = await screen.findByRole('region', { name: 'Convites em aberto' })
  expect(await within(members).findByRole('alert')).toHaveTextContent('Não foi possível carregar os membros.')
  expect(await within(invitations).findByRole('alert')).toHaveTextContent('Não foi possível carregar os convites.')
  expect(screen.queryByText('Nenhum convite em aberto.')).toBeNull()
  server.resetHandlers()
  await userEvent.click(within(members).getByRole('button', { name: 'Tentar novamente' }))
  expect(await within(members).findByText('Ana Silva')).toBeVisible()
  expect(within(invitations).getByRole('alert')).toBeVisible()
  await userEvent.click(within(invitations).getByRole('button', { name: 'Tentar novamente' }))
  expect(await within(invitations).findByText('Nenhum convite em aberto.')).toBeVisible()
})

beforeEach(async () => { await i18n.changeLanguage('pt-BR') })
it('lowercases the invitation field and sends the normalized email', async () => {
  let submitted: unknown
  server.use(http.post('*/api/v1/guilds/1/invites', async ({ request }) => {
    submitted = await request.json()
    return new HttpResponse(null, { status: 204 })
  }))
  render(<MemoryRouter initialEntries={['/guild?guild=1']}><App initialUser={{ id: 1, name: 'Ana', email: 'ana@familia.com' }} demo /></MemoryRouter>)
  const user = userEvent.setup()
  const input = await screen.findByRole('textbox', { name: 'Convidar membro' })
  await user.type(input, 'Member.Name+bank@EXAMPLE.TEST')
  expect(input).toHaveValue('member.name+bank@example.test')
  await user.click(screen.getByRole('button', { name: 'Enviar convite' }))
  await waitFor(() => expect(submitted).toEqual({ email: 'member.name+bank@example.test' }))
})

it('rejects quotes in the invitation field before making a request', async () => {
  let calls = 0
  server.use(http.post('*/api/v1/guilds/1/invites', () => { calls++; return new HttpResponse(null, { status: 204 }) }))
  render(<MemoryRouter initialEntries={['/guild?guild=1']}><App initialUser={{ id: 1, name: 'Ana', email: 'ana@familia.com' }} demo /></MemoryRouter>)
  const user = userEvent.setup()
  const input = await screen.findByRole('textbox', { name: 'Convidar membro' })
  await user.type(input, '"member"@example.test')
  expect(await screen.findByRole('alert')).toHaveTextContent('Informe um email válido.')
  expect(input).toHaveAttribute('aria-invalid', 'true')
  await user.click(screen.getByRole('button', { name: 'Enviar convite' }))
  expect(calls).toBe(0)
})

it('keeps the invitation draft after failure and clears it only on success', async () => {
  let calls = 0
  server.use(http.post('*/api/v1/guilds/1/invites', () => { calls++; return new HttpResponse(null, { status: 500 }) }))
  render(<MemoryRouter initialEntries={['/guild?guild=1']}><App initialUser={{ id: 1, name: 'Ana', email: 'ana@familia.com' }} demo /></MemoryRouter>)
  const user = userEvent.setup()
  const input = await screen.findByRole('textbox', { name: 'Convidar membro' })
  await user.type(input, 'test@example.com')
  await user.click(screen.getByRole('button', { name: 'Enviar convite' }))
  await waitFor(() => expect(calls).toBe(1))
  await waitFor(() => expect(screen.getByRole('button', { name: 'Enviar convite' })).toBeEnabled())
  expect(input).toHaveValue('test@example.com')
  server.use(http.post('*/api/v1/guilds/1/invites', () => new HttpResponse(null, { status: 204 })))
  await user.click(screen.getByRole('button', { name: 'Enviar convite' }))
  await waitFor(() => expect(input).toHaveValue(''))
})
it('offers retry when guild listing fails', async () => {
  server.use(http.get('*/api/v1/guilds', () => new HttpResponse(null, { status: 400 })))
  render(<MemoryRouter initialEntries={['/guild']}><App initialUser={{ id: 1, name: 'Ana', email: 'ana@familia.com' }} demo /></MemoryRouter>)
  const retry = await screen.findByRole('button', { name: 'Tentar novamente' })
  server.resetHandlers()
  await userEvent.click(retry)
  await screen.findByText('Selecione uma guild válida.')
  await userEvent.click(screen.getByRole('button', { name: 'Entrar em Família Silva' }))
  expect(await screen.findByRole('textbox', { name: 'Convidar membro' })).toBeVisible()
})
it('shows the empty guild state without requesting financial data', async () => {
  let financialRequests = 0
  server.use(
    http.get('*/api/v1/guilds', () => HttpResponse.json([])),
    http.get('*/api/v1/guilds/:id/transactions', () => { financialRequests++; return HttpResponse.json([]) }),
  )
  render(<MemoryRouter initialEntries={['/transactions']}><App initialUser={{ id: 1, name: 'Ana', email: 'ana@familia.com' }} demo /></MemoryRouter>)
  expect(await screen.findByRole('heading', { name: 'Olá, Ana!' })).toBeVisible()
  await userEvent.click(screen.getByRole('button', { name: 'Continuar' }))
  expect(await screen.findByRole('heading', { name: 'Crie sua guild' })).toBeVisible()
  expect(financialRequests).toBe(0)
})

afterEach(() => vi.restoreAllMocks())

it.each(['invite', 'remove'])('reports forbidden guild %s without dropping state', async (action) => {
  const error = vi.spyOn(toast, 'error'), success = vi.spyOn(toast, 'success')
  let calls = 0
  const denied = () => { calls++; return new HttpResponse(null, { status: 403 }) }
  server.use(http.post('*/api/v1/guilds/1/invites', denied), http.delete('*/api/v1/guilds/1/members/:id', denied))
  render(<MemoryRouter initialEntries={['/guild?guild=1']}><App initialUser={{ id: 1, name: 'Ana', email: 'ana@example.test' }} demo /></MemoryRouter>)
  const user = userEvent.setup()
  const input = await screen.findByRole('textbox', { name: 'Convidar membro' })
  await user.type(input, 'draft@example.test')
  if (action === 'invite') await user.click(screen.getByRole('button', { name: 'Enviar convite' }))
  else await user.click((await screen.findAllByRole('button', { name: /^Remover/ }))[0]!)
  await waitFor(() => expect(error).toHaveBeenCalledWith('Você não tem permissão para realizar esta ação.'))
  expect(input).toHaveValue('draft@example.test')
  expect(calls).toBe(1); expect(success).not.toHaveBeenCalled()
  expect(screen.getByRole('heading', { name: 'Guild / Membros' })).toBeVisible()
  if (action === 'remove') expect(screen.getAllByRole('button', { name: /^Remover/ }).length).toBeGreaterThan(0)
})

it.each([
  ['invitation_already_pending', 'O usuário já foi convidado para esta guild.'],
  ['guild_member_already_set', 'O usuário já é membro desta guild.'],
])('localizes %s without losing the invitation draft', async (code, message) => {
  const error = vi.spyOn(toast, 'error')
  server.use(http.post('*/api/v1/guilds/1/invites', () => HttpResponse.json({ error: 'conflict', code }, { status: 409 })))
  render(<MemoryRouter initialEntries={['/guild?guild=1']}><App initialUser={{ id: 1, name: 'Ana', email: 'ana@familia.com' }} demo /></MemoryRouter>)
  const user = userEvent.setup()
  const input = await screen.findByRole('textbox', { name: 'Convidar membro' })
  await user.type(input, 'draft@example.test')
  await user.click(screen.getByRole('button', { name: 'Enviar convite' }))
  await waitFor(() => expect(error).toHaveBeenCalledWith(message))
  expect(input).toHaveValue('draft@example.test')
})
