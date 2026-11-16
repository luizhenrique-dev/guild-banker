import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { MemoryRouter, useNavigate } from 'react-router-dom'
import { http, HttpResponse } from 'msw'
import { App } from '@/src/app/app'
import { server } from '@/src/mocks/server'
import i18n from '@/src/shared/lib/i18n'
import * as queryClients from '@/src/shared/lib/query-client'
import { fixedExpensesApi } from '../../fixed-expenses/api/fixed-expenses.api'
import type { FixedExpenseResponse } from '../../fixed-expenses/types'

const user = { id: 1, name: 'Ana', email: 'ana@example.test' }
const rows: FixedExpenseResponse[] = [
  { id: 1, guild_id: 1, name: 'Internet', amount: '123.45', due_day: 15, status: 'ACTIVE', responsible_user_id: null, category: 'OTHER' },
  { id: 2, guild_id: 1, name: 'Aluguel', amount: '2500.00', due_day: 16, status: 'ACTIVE', responsible_user_id: null, category: 'HOUSING' },
]
function open(path = '/home?guild=1') {
  return render(<MemoryRouter initialEntries={[path]}><App initialUser={user} demo /></MemoryRouter>)
}
function section() { return screen.getByRole('region', { name: 'Despesas fixas - Próximos vencimentos' }) }

beforeEach(async () => {
  vi.useFakeTimers({ toFake: ['Date'] })
  vi.setSystemTime(new Date(2026, 10, 15, 23, 59))
  await i18n.changeLanguage('pt-BR')
  server.use(http.get('*/api/v1/guilds/:guildID/fixed-expenses', () => HttpResponse.json(rows)))
})
afterEach(() => { vi.useRealTimers(); vi.restoreAllMocks() })

it('shows the current month, money, dates, today and tomorrow below the welcome', async () => {
  open()
  expect(await screen.findByText('Internet')).toBeVisible()
  expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Bem-vindo(a) Ana')
  const content = within(section())
  expect(content.getByText('novembro de 2026')).toBeVisible()
  expect(content.getByText('Hoje')).toBeVisible()
  expect(content.getByText('Amanhã')).toBeVisible()
  expect(content.getByText(/123,45/)).toHaveTextContent('R$')
  expect(content.getByText('15/11/2026')).toBeVisible()
  expect(content.getByText('16/11/2026')).toBeVisible()
  expect(content.getAllByRole('listitem')).toHaveLength(2)
  await act(async () => { await i18n.changeLanguage('en') })
  const english = within(screen.getByRole('region', { name: 'Fixed expenses - Upcoming due dates' }))
  expect(english.getByText('Today')).toBeVisible()
  expect(english.getByText('Tomorrow')).toBeVisible()
  expect(english.getByText('November 2026')).toBeVisible()
})

it('shows category totals for the whole guild month rather than only upcoming expenses', async () => {
  server.use(http.get('*/api/v1/guilds/:guildID/fixed-expenses', () => HttpResponse.json([
    ...rows,
    { ...rows[0], id: 3, amount: '200', due_day: 1 },
    { ...rows[0], id: 4, amount: '9999', category: 'HEALTH', status: 'PAUSED' },
  ])))
  open()
  const categories = within(await screen.findByRole('region', { name: 'Despesas fixas - Por categoria' }))
  const groups = await categories.findAllByRole('listitem')
  expect(groups).toHaveLength(2)
  expect(groups[0]).toHaveTextContent('Moradia')
  expect(groups[0]).toHaveTextContent('2.500,00')
  expect(groups[1]).toHaveTextContent('Outros')
  expect(groups[1]).toHaveTextContent('323,45')
  expect(within(section()).getAllByRole('listitem')).toHaveLength(2)
})

it('distinguishes loading, technical failure, retry and an empty month', async () => {
  let finish!: (response: Response) => void
  server.use(http.get('*/api/v1/guilds/:guildID/fixed-expenses', () => new Promise<Response>((resolve) => { finish = resolve })))
  open()
  await waitFor(() => expect(finish).toBeTypeOf('function'))
  expect(within(section()).getByRole('status')).toHaveTextContent('Carregando')
  server.use(http.get('*/api/v1/guilds/:guildID/fixed-expenses', () => new HttpResponse(null, { status: 500 })))
  await act(async () => { finish(new HttpResponse(null, { status: 500 })) })
  expect(await within(section()).findByRole('alert')).toHaveTextContent('Não foi possível carregar os gastos fixos.')
  expect(screen.queryByText('Nenhum próximo gasto fixo neste mês.')).toBeNull()
  server.use(http.get('*/api/v1/guilds/:guildID/fixed-expenses', () => HttpResponse.json([])))
  fireEvent.click(within(section()).getByRole('button', { name: 'Tentar novamente' }))
  expect(await screen.findByText('Nenhum próximo gasto fixo neste mês.')).toBeVisible()
  expect(screen.getByText('Nenhuma despesa fixa ativa na guild.')).toBeVisible()
})

it('hides cached expenses on a forbidden refetch while keeping identity controls', async () => {
  const client = queryClients.makeQueryClient()
  vi.spyOn(queryClients, 'makeQueryClient').mockReturnValue(client)
  open()
  await screen.findByText('Internet')
  server.use(http.get('*/api/v1/guilds/:guildID/fixed-expenses', () => new HttpResponse(null, { status: 403 })))
  await act(async () => { await client.invalidateQueries({ queryKey: ['fixed-expenses'] }) })
  expect(await within(section()).findByText('Acesso negado')).toBeVisible()
  expect(within(screen.getByRole('region', { name: 'Despesas fixas - Por categoria' })).queryByRole('list')).toBeNull()
  expect(screen.queryByText('Internet')).toBeNull()
  expect(screen.getByRole('button', { name: 'Sair' })).toBeEnabled()
})

it('reopens using the new local day and month while reusing cached expenses', async () => {
  open()
  await screen.findByText('Internet')
  const list = vi.spyOn(fixedExpensesApi, 'list').mockImplementation(() => new Promise(() => {}))
  fireEvent.click(screen.getByRole('link', { name: 'Despesas fixas' }))
  await screen.findByRole('heading', { name: 'Despesas fixas da guild' })
  vi.setSystemTime(new Date(2026, 10, 16, 0, 1))
  fireEvent.click(screen.getByRole('link', { name: 'Início' }))
  expect(within(section()).queryByText('Internet')).toBeNull()
  expect(within(section()).getByText('Hoje')).toBeVisible()
  fireEvent.click(screen.getByRole('link', { name: 'Despesas fixas' }))
  vi.setSystemTime(new Date(2026, 11, 1))
  fireEvent.click(screen.getByRole('link', { name: 'Início' }))
  expect(within(section()).getByText('dezembro de 2026')).toBeVisible()
  expect(within(section()).getByText('Internet')).toBeVisible()
  await waitFor(() => expect(list).toHaveBeenCalledOnce())
})

it('ignores a late response from the previous guild', async () => {
  let finish!: (value: FixedExpenseResponse[]) => void
  vi.spyOn(fixedExpensesApi, 'list').mockImplementation((_options, guild) => guild === 1
    ? new Promise((resolve) => { finish = resolve })
    : Promise.resolve([{ ...rows[0]!, guild_id: 2, name: 'República' }]))
  function Navigate() {
    const navigate = useNavigate()
    return <button onClick={() => navigate('/home?guild=2')}>Switch</button>
  }
  render(<MemoryRouter initialEntries={['/home?guild=1']}><Navigate /><App initialUser={user} demo /></MemoryRouter>)
  await waitFor(() => expect(finish).toBeTypeOf('function'))
  fireEvent.click(screen.getByRole('button', { name: 'Switch' }))
  expect(await screen.findByText('República', { exact: true })).toBeVisible()
  await act(async () => { finish(rows) })
  expect(screen.queryByText('Internet')).toBeNull()
})

it('clears old user expenses before the new user response', async () => {
  const { rerender } = open('/home?guild=2')
  await screen.findByText('Internet')
  let finish!: (value: FixedExpenseResponse[]) => void
  vi.spyOn(fixedExpensesApi, 'list').mockImplementation(() => new Promise((resolve) => { finish = resolve }))
  rerender(<MemoryRouter initialEntries={['/home?guild=2']}><App initialUser={{ id: 2, name: 'Other', email: 'other@example.test' }} demo /></MemoryRouter>)
  expect(screen.queryByText('Internet')).toBeNull()
  await waitFor(() => expect(finish).toBeTypeOf('function'))
  await act(async () => { finish([{ ...rows[0]!, guild_id: 2, name: 'New user expense' }]) })
  expect(await screen.findByText('New user expense')).toBeVisible()
})

it.each(['/home', '/home?guild=999'])('does not query finances without a valid guild on %s', async (path) => {
  const list = vi.spyOn(fixedExpensesApi, 'list')
  open(path)
  expect(await screen.findByRole('heading', { name: 'Escolha sua guild' })).toBeVisible()
  expect(list).not.toHaveBeenCalled()
})
