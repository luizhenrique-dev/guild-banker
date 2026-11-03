import { act, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, useNavigate } from 'react-router-dom'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { http, HttpResponse } from 'msw'
import { guildsFixtures } from '@/src/mocks/fixtures/guilds'
import { mockGuildMember } from '@/src/mocks/request-context'
import { server } from '@/src/mocks/server'
import { App } from '@/src/app/app'
import i18n from '@/src/shared/lib/i18n'
import { fixedExpensesApi } from '../api/fixed-expenses.api'
import { toast } from 'sonner'
import type { FixedExpenseResponse } from '../types'
import * as queryClients from '@/src/shared/lib/query-client'

beforeEach(async () => { await i18n.changeLanguage('pt-BR') })
afterEach(() => vi.restoreAllMocks())
function page() {
  return render(<MemoryRouter initialEntries={['/fixed-expenses']}><App initialUser={{ id: 1, name: 'Ana', email: 'ana@familia.com' }} demo /></MemoryRouter>)
}
async function fill() {
  const user = userEvent.setup()
  await user.click(await screen.findByRole('button', { name: 'Nova despesa fixa' }))
  const dialog = within(screen.getByRole('dialog'))
  await user.type(dialog.getByRole('textbox', { name: 'Nome' }), 'Internet')
  await user.type(dialog.getByRole('textbox', { name: 'Valor (R$)' }), '123,45')
  return { user, dialog }
}
it('shows loading, a GET error with manual retry and an empty list', async () => {
  let finish!: (rows: FixedExpenseResponse[]) => void
  const list = vi.spyOn(fixedExpensesApi, 'list').mockImplementationOnce(() => new Promise((resolve) => { finish = resolve })).mockRejectedValueOnce(new Error('offline')).mockResolvedValue([])
  const view = page()
  expect(await screen.findByRole('status', { name: 'Carregando...' })).toBeVisible()
  await waitFor(() => expect(finish).toBeTypeOf('function'))
  await act(async () => { finish([]) })
  expect(await screen.findByText('Nenhuma despesa fixa cadastrada.')).toBeVisible()
  view.unmount()
  page()
  const retry = await screen.findByRole('button', { name: 'Tentar novamente' })
  await userEvent.click(retry)
  expect(await screen.findByText('Nenhuma despesa fixa cadastrada.')).toBeVisible()
  expect(list).toHaveBeenCalledTimes(3)
})
it('preserves values and exposes server error after a single failed write', async () => {
  let calls = 0
  server.use(http.post('*/api/v1/fixed-expenses', () => { calls++; return new HttpResponse(null, { status: 500 }) }))
  page()
  const { user, dialog } = await fill()
  await user.click(dialog.getByRole('button', { name: 'Salvar' }))
  expect(await dialog.findByRole('alert')).toHaveTextContent('Ocorreu um erro')
  expect(dialog.getByRole('textbox', { name: 'Nome' })).toHaveValue('Internet')
  expect(dialog.getByRole('textbox', { name: 'Valor (R$)' })).toHaveValue('123,45')
  expect(calls).toBe(1)
})
it('disables duplicate sends while pending and refreshes the list after success', async () => {
  page()
  const { user, dialog } = await fill()
  const create = vi.spyOn(fixedExpensesApi, 'create')
  await user.dblClick(dialog.getByRole('button', { name: 'Salvar' }))
  expect(dialog.getByRole('button', { name: 'Criando...' })).toBeDisabled()
  expect(dialog.getByRole('textbox', { name: 'Nome' })).toBeDisabled()
  expect(create).toHaveBeenCalledTimes(1)
  expect(await screen.findByRole('row', { name: /Internet/ }, { timeout: 3000 })).toBeVisible()
  expect(screen.queryByRole('dialog')).toBeNull()
})
it('reports saved changes when refetch fails and retries only GET', async () => {
  page()
  await screen.findByRole('row', { name: /Aluguel/ })
  const { user, dialog } = await fill()
  let writes = 0
  server.use(
    http.post('*/api/v1/fixed-expenses', async ({ request }) => { writes++; return HttpResponse.json({ ...(await request.json() as object), id: 100, status: 'ACTIVE' }, { status: 201 }) }),
    http.get('*/api/v1/fixed-expenses', () => new HttpResponse(null, { status: 400 })),
  )
  await user.click(dialog.getByRole('button', { name: 'Salvar' }))
  expect(await screen.findByRole('alert')).toHaveTextContent('A alteração foi salva, mas a lista não foi atualizada.')
  expect(screen.queryByRole('dialog')).toBeNull()
  server.use(http.get('*/api/v1/fixed-expenses', () => HttpResponse.json([{ id: 100, name: 'Internet', amount: '123.45', due_day: 1, category: 'OTHER', status: 'ACTIVE' }])))
  await user.click(screen.getByRole('button', { name: 'Tentar novamente' }))
  expect(await screen.findByRole('row', { name: /Internet/ }, { timeout: 3000 })).toBeVisible()
  expect(writes).toBe(1)
})
it('associates validation errors and updates them with selected language', async () => {
  page()
  const { user, dialog } = await fill()
  const amount = dialog.getByRole('textbox', { name: 'Valor (R$)' })
  await user.clear(amount)
  await user.type(amount, '0')
  await user.click(dialog.getByRole('button', { name: 'Salvar' }))
  expect(amount).toHaveAttribute('aria-invalid', 'true')
  expect(amount).toHaveAccessibleDescription('Informe um valor maior que zero, com até duas casas decimais.')
  await act(async () => { await i18n.changeLanguage('en') })
  expect(amount).toHaveAccessibleDescription('Enter an amount greater than zero with up to two decimal places.')
})

it('replaces cached personal expenses with access denied after forbidden refetch', async () => {
  const client = queryClients.makeQueryClient()
  vi.spyOn(queryClients, 'makeQueryClient').mockReturnValue(client)
  page()
  expect(await screen.findByRole('row', { name: /Aluguel/ })).toBeVisible()
  let calls = 0
  server.use(http.get('*/api/v1/fixed-expenses', () => { calls++; return new HttpResponse(null, { status: 403 }) }))
  await act(async () => { await client.invalidateQueries({ queryKey: ['fixed-expenses'] }) })
  expect(await screen.findByRole('heading', { name: 'Acesso negado' })).toBeVisible()
  expect(screen.queryByRole('table')).toBeNull()
  expect(calls).toBe(1)
})


it.each(['create', 'edit', 'deactivate'])('preserves local state on forbidden expense %s', async (action) => {
  const success = vi.spyOn(toast, 'success')
  let calls = 0
  const denied = () => { calls++; return HttpResponse.json({ error: 'private cause' }, { status: 403 }) }
  server.use(http.post('*/api/v1/fixed-expenses', denied), http.patch('*/api/v1/fixed-expenses/:id', denied), http.patch('*/api/v1/fixed-expenses/:id/deactivate', denied))
  page()
  const user = userEvent.setup()
  if (action === 'create') {
    const { dialog } = await fill()
    await user.click(dialog.getByRole('button', { name: 'Salvar' }))
    expect(await dialog.findByRole('alert')).toHaveTextContent('Você não tem permissão para realizar esta ação.')
    expect(dialog.getByRole('textbox', { name: 'Nome' })).toHaveValue('Internet')
    await act(async () => { await i18n.changeLanguage('en') })
    expect(dialog.getByRole('alert')).toHaveTextContent('You do not have permission to perform this action.')
  } else {
    await screen.findByRole('row', { name: /Aluguel/ })
    await user.click(screen.getByRole('button', { name: `${action === 'edit' ? 'Editar' : 'Desativar'} Aluguel` }))
    const dialog = within(screen.getByRole(action === 'edit' ? 'dialog' : 'alertdialog'))
    if (action === 'edit') {
      const amount = dialog.getByRole('textbox', { name: 'Valor (R$)' })
      await user.clear(amount); await user.type(amount, '321,00')
      await user.click(dialog.getByRole('button', { name: 'Salvar' }))
      expect(await dialog.findByRole('alert')).toHaveTextContent('Você não tem permissão para realizar esta ação.')
      expect(amount).toHaveValue('321,00')
    } else {
      await user.click(dialog.getByRole('button', { name: 'Desativar' }))
      expect(await dialog.findByRole('alert')).toHaveTextContent('Você não tem permissão para realizar esta ação.')
      expect(screen.getByRole('alertdialog')).toBeVisible()
    }
  }
  expect(calls).toBe(1)
  expect(screen.queryByRole('heading', { name: 'Acesso negado' })).toBeNull()
  expect(screen.queryByText('private cause')).toBeNull()
  expect(success).not.toHaveBeenCalled()
})

beforeEach(() => server.use(http.get('*/api/v1/guilds', ({ request }) => HttpResponse.json(guildsFixtures.filter((guild) => mockGuildMember(request, guild.id))))))

it('a late create from G1 leaves the new G2 draft intact and invalidates only G1', async () => {
  const client = queryClients.makeQueryClient()
  vi.spyOn(queryClients, 'makeQueryClient').mockReturnValue(client)
  const invalidate = vi.spyOn(client, 'invalidateQueries')
  const success = vi.spyOn(toast, 'success')
  let navigate!: ReturnType<typeof useNavigate>
  function Navigation() { navigate = useNavigate(); return null }
  let finish!: (value: FixedExpenseResponse) => void
  vi.spyOn(fixedExpensesApi, 'create').mockImplementation(() => new Promise((resolve) => { finish = resolve }))
  render(<MemoryRouter initialEntries={['/fixed-expenses?guild=1']}><Navigation /><App initialUser={{ id: 1, name: 'Ana', email: 'ana@example.test' }} demo /></MemoryRouter>)
  const { user, dialog } = await fill()
  await user.click(dialog.getByRole('button', { name: 'Salvar' }))
  await waitFor(() => expect(finish).toBeTypeOf('function'))
  act(() => navigate('/fixed-expenses?guild=2'))
  expect(screen.queryByRole('dialog')).toBeNull()
  const next = await fill()
  await act(async () => { finish({ id: 100, name: 'Old', amount: '123.45', due_day: 1, category: 'OTHER', status: 'ACTIVE' }) })
  expect(screen.getByRole('dialog')).toBeVisible()
  expect(next.dialog.getByRole('textbox', { name: 'Nome' })).toHaveValue('Internet')
  expect(invalidate).toHaveBeenCalledExactlyOnceWith({ queryKey: ['fixed-expenses', 1, 1] })
  expect(success).not.toHaveBeenCalled()
})
it('changing guild closes the previous deactivation confirmation', async () => {
  let navigate!: ReturnType<typeof useNavigate>
  function Navigation() { navigate = useNavigate(); return null }
  render(<MemoryRouter initialEntries={['/fixed-expenses?guild=1']}><Navigation /><App initialUser={{ id: 1, name: 'Ana', email: 'ana@example.test' }} demo /></MemoryRouter>)
  await screen.findByRole('row', { name: /Aluguel/ })
  await userEvent.click(screen.getByRole('button', { name: 'Desativar Aluguel' }))
  expect(screen.getByRole('alertdialog')).toBeVisible()
  act(() => navigate('/fixed-expenses?guild=2'))
  expect(screen.queryByRole('alertdialog')).toBeNull()
})
