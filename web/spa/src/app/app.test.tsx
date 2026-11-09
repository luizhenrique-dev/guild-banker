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
  render(<MemoryRouter initialEntries={['/missing']}><App initialUser={{ id: 1, name: 'Ana', email: 'ana@example.com' }} demo /></MemoryRouter>)
  expect(await screen.findByRole('heading', { name: 'Página não encontrada' })).toBeVisible()
})

it('opens Home from root and menu, preserving query without requiring a guild', async () => {
  render(<MemoryRouter initialEntries={['/?guild=999']}><App initialUser={{ id: 99, name: '<b>Ana</b>', email: 'ana@example.com' }} demo /></MemoryRouter>)
  expect(await screen.findByRole('heading', { name: 'Bem-vindo(a) <b>Ana</b>' })).toBeVisible()
  const home = screen.getByRole('link', { name: 'Início' })
  expect(home).toHaveAttribute('href', '/home?guild=999')
  fireEvent.click(home)
  expect(screen.getByRole('heading', { name: 'Bem-vindo(a) <b>Ana</b>' }).querySelector('b')).toBeNull()
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
it.each(['/fixed-expenses', '/transactions?guild=1', '/import?guild=1', '/guild?guild=1'])('keeps shell and Home for forbidden loading on %s', async (path) => {
  let calls = 0
  const endpoint = path.startsWith('/fixed-expenses') ? '*/api/v1/guilds/:guildID/fixed-expenses' : '*/api/v1/guilds'
  server.use(http.get(endpoint, () => { calls++; return HttpResponse.json({ error: 'secret internal cause' }, { status: 403 }) }))
  render(<MemoryRouter initialEntries={[path]}><App initialUser={demoUser} demo /></MemoryRouter>)
  expect(await screen.findByRole('heading', { name: 'Acesso negado' })).toBeVisible()
  expect(screen.getByRole('navigation')).toBeVisible()
  expect(screen.getByRole('button', { name: 'Sair' })).toBeEnabled()
  expect(screen.queryByText('secret internal cause')).toBeNull()
  expect(calls).toBe(1)
  fireEvent.click(screen.getByRole('link', { name: 'Voltar ao início' }))
  expect(await screen.findByRole('heading', { name: 'Bem-vindo(a) Ana' })).toBeVisible()
  fireEvent.click(screen.getByRole('button', { name: 'Sair' }))
  expect(await screen.findByRole('heading', { name: 'Sessão encerrada' })).toBeVisible()
})
it.each(['/home'])('does not block independent %s when guilds are forbidden', async (path) => {
  server.use(http.get('*/api/v1/guilds', () => new HttpResponse(null, { status: 403 })))
  render(<MemoryRouter initialEntries={[path]}><App initialUser={demoUser} demo /></MemoryRouter>)
  const title = path === '/home' ? 'Bem-vindo(a) Ana' : 'Despesas fixas da guild'
  expect(await screen.findByRole('heading', { name: title })).toBeVisible()
  if (path === '/fixed-expenses') expect(await screen.findByRole('row', { name: /Aluguel/ })).toBeVisible()
  expect(screen.queryByRole('heading', { name: 'Acesso negado' })).toBeNull()
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
