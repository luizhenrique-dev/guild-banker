import { act, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '@/src/mocks/server'
import { App } from '@/src/app/app'
import i18n from '@/src/shared/lib/i18n'
import { fixedExpensesApi } from '../api/fixed-expenses.api'
import type { FixedExpenseResponse } from '../types'

beforeEach(async () => { await i18n.changeLanguage('pt-BR') })
afterEach(() => vi.restoreAllMocks())
function page() {
  return render(<MemoryRouter initialEntries={['/fixed-expenses']}><App initialUser={{ id: 1, name: 'Ana', email: 'ana@familia.com' }} demo /></MemoryRouter>)
}
async function fill() {
  const user = userEvent.setup()
  await user.click(screen.getByRole('button', { name: 'Nova despesa fixa' }))
  const dialog = within(screen.getByRole('dialog'))
  await user.type(dialog.getByRole('textbox', { name: 'Nome' }), 'Internet')
  await user.type(dialog.getByRole('textbox', { name: 'Valor (R$)' }), '123,45')
  return { user, dialog }
}
it('shows loading, a GET error with manual retry and an empty list', async () => {
  let finish!: (rows: FixedExpenseResponse[]) => void
  const list = vi.spyOn(fixedExpensesApi, 'list').mockImplementationOnce(() => new Promise((resolve) => { finish = resolve })).mockRejectedValueOnce(new Error('offline')).mockResolvedValue([])
  const view = page()
  expect(screen.getByRole('status', { name: 'Carregando...' })).toBeVisible()
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
