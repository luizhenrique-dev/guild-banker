import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '@/src/mocks/server'
import { App } from '@/src/app/app'
import i18n from '@/src/shared/lib/i18n'
import { toast } from 'sonner'

beforeEach(async () => { await i18n.changeLanguage('pt-BR') })
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
  expect(await screen.findByRole('textbox', { name: 'Convidar membro' })).toBeVisible()
})
it('shows the empty guild state without requesting financial data', async () => {
  let financialRequests = 0
  server.use(
    http.get('*/api/v1/guilds', () => HttpResponse.json([])),
    http.get('*/api/v1/guilds/:id/transactions', () => { financialRequests++; return HttpResponse.json([]) }),
  )
  render(<MemoryRouter initialEntries={['/transactions']}><App initialUser={{ id: 1, name: 'Ana', email: 'ana@familia.com' }} demo /></MemoryRouter>)
  expect(await screen.findByText('Nenhuma guild disponível.')).toBeVisible()
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
