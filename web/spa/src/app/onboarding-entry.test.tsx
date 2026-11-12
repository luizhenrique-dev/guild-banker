import { render, screen, waitFor, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '@/src/mocks/server'
import { guildsFixtures } from '@/src/mocks/fixtures/guilds'
import i18n from '@/src/shared/lib/i18n'
import { App } from './app'

beforeEach(async () => { await i18n.changeLanguage('pt-BR') })
it('requires a guild before showing Home and navigation while retaining identity controls', async () => {
  render(<MemoryRouter initialEntries={['/home']}><App initialUser={{ id: 1, name: 'Ana', email: 'ana@familia.com' }} demo /></MemoryRouter>)
  expect(screen.queryByRole('heading', { name: 'Bem-vindo(a) Ana' })).toBeNull()
  expect(screen.queryByRole('navigation')).toBeNull()
  expect(screen.getByRole('button', { name: 'Sair' })).toBeVisible()
  expect(await screen.findByRole('heading', { name: 'Escolha sua guild' })).toBeVisible()
  await userEvent.click(screen.getByRole('button', { name: 'Entrar em Família Silva' }))
  expect(await screen.findByRole('heading', { name: 'Bem-vindo(a) Ana' })).toBeVisible()
  expect(screen.getByRole('navigation')).toBeVisible()
})
it('blocks Home on membership 403 without a link back into the same block', async () => {
  server.use(http.get('*/api/v1/guilds', () => new HttpResponse(null, { status: 403 })))
  render(<MemoryRouter initialEntries={['/home']}><App initialUser={{ id: 1, name: 'Ana', email: 'ana@familia.com' }} demo /></MemoryRouter>)
  expect(await screen.findByRole('heading', { name: 'Acesso negado' })).toBeVisible()
  expect(screen.queryByRole('navigation')).toBeNull()
  expect(screen.queryByRole('link', { name: 'Voltar ao início' })).toBeNull()
  expect(screen.getByRole('button', { name: 'Sair' })).toBeVisible()
})
it('does not mount financial pages while membership lookup is pending', async () => {
  let finish!: (response: Response) => void
  let financial = 0
  server.use(http.get('*/api/v1/guilds', () => new Promise<Response>((resolve) => { finish = resolve })), http.get('*/api/v1/guilds/:id/fixed-expenses', () => { financial++; return HttpResponse.json([]) }))
  render(<MemoryRouter initialEntries={['/fixed-expenses']}><App initialUser={{ id: 1, name: 'Ana', email: 'ana@familia.com' }} demo /></MemoryRouter>)
  await waitFor(() => expect(finish).toBeTypeOf('function'))
  expect(screen.queryByRole('navigation')).toBeNull()
  expect(financial).toBe(0)
  await act(async () => { finish(HttpResponse.json([guildsFixtures[0]])) })
  expect(await screen.findByRole('heading', { name: 'Despesas fixas da guild' })).toBeVisible()
  await waitFor(() => expect(financial).toBe(1))
})

it('identifies the active guild and switcher options when visual names repeat', async () => {
  server.use(http.get('*/api/v1/guilds', () => HttpResponse.json(guildsFixtures.map((guild) => ({ ...guild, display_name: 'Família' })))))
  render(<MemoryRouter initialEntries={['/home?guild=1']}><App initialUser={{ id: 1, name: 'Ana', email: 'ana@familia.com' }} demo /></MemoryRouter>)
  await userEvent.click(await screen.findByRole('button', { name: 'Família (familia-silva)' }))
  await userEvent.click(screen.getByRole('button', { name: 'Família (republica-amigos)' }))
  expect(screen.getByRole('button', { name: 'Família (republica-amigos)' })).toBeVisible()
  expect(screen.getByRole('link', { name: 'Início' })).toHaveAttribute('href', '/home?guild=2')
})
