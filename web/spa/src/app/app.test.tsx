import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { act, render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { App } from './app'
import i18n from '../shared/lib/i18n'

beforeEach(async () => {
  await i18n.changeLanguage('pt-BR')
})
it.each(['/', '/home', '/fixed-expenses', '/transactions', '/import', '/guild'])('gates %s without requests when session is absent', async (path) => {
  const fetchSpy = vi.spyOn(globalThis, 'fetch')
  render(<MemoryRouter initialEntries={[path]}><App /></MemoryRouter>)
  expect(await screen.findByRole('heading', { name: 'Sessão encerrada' })).toBeVisible()
  expect(screen.queryByText('Aluguel')).not.toBeInTheDocument()
  expect(fetchSpy).not.toHaveBeenCalled()
  fetchSpy.mockRestore()
})
it('handles unknown routes', async () => {
  render(<MemoryRouter initialEntries={['/missing?guild=1']}><App initialUser={{ id: 1, name: 'Ana', email: 'ana@example.com' }} demo /></MemoryRouter>)
  expect(await screen.findByRole('heading', { name: 'Página não encontrada' })).toBeVisible()
})

it('opens Home only after choosing a guild and escapes the session name', async () => {
  render(<MemoryRouter initialEntries={['/?guild=999']}><App initialUser={{ id: 1, name: '<b>Ana</b>', email: 'ana@example.com' }} demo /></MemoryRouter>)
  expect(await screen.findByRole('heading', { name: 'Escolha sua guild' })).toBeVisible()
  expect(screen.queryByRole('navigation')).toBeNull()
  fireEvent.click(screen.getByRole('button', { name: 'Entrar em Família Silva' }))
  const heading = await screen.findByRole('heading', { name: 'Bem-vindo(a) <b>Ana</b>' })
  expect(heading.querySelector('b')).toBeNull()
  const home = screen.getByRole('link', { name: 'Início' })
  expect(home).toHaveAttribute('href', '/home?guild=1')
})

it('does not mount the shell even on unknown routes without a session', async () => {
  const fetchSpy = vi.spyOn(globalThis, 'fetch')
  render(<MemoryRouter initialEntries={['/missing']}><App /></MemoryRouter>)
  expect(screen.queryByRole('navigation')).not.toBeInTheDocument()
  expect(screen.queryByRole('button', { name: 'Sair' })).not.toBeInTheDocument()
  expect(fetchSpy).not.toHaveBeenCalled()
  fetchSpy.mockRestore()
})

import { http, HttpResponse } from 'msw'
import { server } from '../mocks/server'
import * as queryClients from '../shared/lib/query-client'

afterEach(() => vi.restoreAllMocks())
const demoUser = { id: 1, name: 'Ana', email: 'ana@example.test' }
it.each(['/fixed-expenses?guild=1', '/transactions?guild=1', '/import?guild=1', '/guild?guild=1'])('keeps identity controls for forbidden loading on %s', async (path) => {
  let calls = 0
  const endpoint = path.startsWith('/fixed-expenses') ? '*/api/v1/guilds/:guildID/fixed-expenses' : '*/api/v1/guilds'
  server.use(http.get(endpoint, () => { calls++; return HttpResponse.json({ error: 'secret internal cause' }, { status: 403 }) }))
  render(<MemoryRouter initialEntries={[path]}><App initialUser={demoUser} demo /></MemoryRouter>)
  expect(await screen.findByRole('heading', { name: 'Acesso negado' })).toBeVisible()
  if (path.startsWith('/fixed-expenses')) expect(screen.getByRole('navigation')).toBeVisible()
  else expect(screen.queryByRole('navigation')).toBeNull()
  expect(screen.getByRole('button', { name: 'Sair' })).toBeEnabled()
  expect(screen.queryByText('secret internal cause')).toBeNull()
  expect(calls).toBe(1)
  if (path.startsWith('/fixed-expenses')) {
    fireEvent.click(screen.getByRole('link', { name: 'Voltar ao início' }))
    expect(await screen.findByRole('heading', { name: 'Bem-vindo(a) Ana' })).toBeVisible()
  } else expect(screen.queryByRole('link', { name: 'Voltar ao início' })).toBeNull()
  fireEvent.click(screen.getByRole('button', { name: 'Sair' }))
  expect(await screen.findByRole('heading', { name: 'Sessão encerrada' })).toBeVisible()
})
it('blocks Home when memberships are forbidden', async () => {
  server.use(http.get('*/api/v1/guilds', () => new HttpResponse(null, { status: 403 })))
  render(<MemoryRouter initialEntries={['/home']}><App initialUser={demoUser} demo /></MemoryRouter>)
  expect(await screen.findByRole('heading', { name: 'Acesso negado' })).toBeVisible()
  expect(screen.queryByRole('heading', { name: 'Bem-vindo(a) Ana' })).toBeNull()
  expect(screen.queryByRole('navigation')).toBeNull()
})
it('removes refused guild cache and stops dependent loading after refetch', async () => {
  const client = queryClients.makeQueryClient()
  vi.spyOn(queryClients, 'makeQueryClient').mockReturnValue(client)
  let financialCalls = 0
  server.events.on('request:start', ({ request }) => { if (request.url.includes('/transactions')) financialCalls++ })
  render(<MemoryRouter initialEntries={['/transactions?guild=1']}><App initialUser={demoUser} demo /></MemoryRouter>)
  expect(await screen.findByRole('button', { name: 'Família Silva' })).toBeVisible()
  await waitFor(() => expect(financialCalls).toBeGreaterThan(0))
  const previousCalls = financialCalls
  server.use(http.get('*/api/v1/guilds', () => new HttpResponse(null, { status: 403 })))
  await act(async () => { await client.invalidateQueries({ queryKey: ['guilds'] }) })
  expect(await screen.findByRole('heading', { name: 'Acesso negado' })).toBeVisible()
  expect(screen.queryByRole('button', { name: 'Família Silva' })).toBeNull()
  expect(screen.queryByRole('table')).toBeNull()
  expect(financialCalls).toBe(previousCalls)
  server.events.removeAllListeners('request:start')
})

it.each(['/fixed-expenses', '/fixed-expenses?guild=999'])('blocks expenses without authorized guild on %s', async (path) => {
  let calls = 0
  server.use(http.get('*/api/v1/guilds', () => new HttpResponse(null, { status: 403 })), http.get('*/api/v1/guilds/:guildID/fixed-expenses', () => { calls++; return HttpResponse.json([]) }))
  render(<MemoryRouter initialEntries={[path]}><App initialUser={demoUser} demo /></MemoryRouter>)
  expect(await screen.findByRole('heading', { name: 'Acesso negado' })).toBeVisible()
  expect(calls).toBe(0)
})
