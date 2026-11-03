import userEvent from '@testing-library/user-event'
import { toast } from 'sonner'
import { act, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '@/src/mocks/server'
import { App } from '@/src/app/app'
import * as queryClients from '@/src/shared/lib/query-client'
import i18n from '@/src/shared/lib/i18n'

beforeEach(async () => { await i18n.changeLanguage('pt-BR') })
afterEach(() => vi.restoreAllMocks())
it.each([false, true])('prioritizes forbidden transactions over cached data: %s', async (cached) => {
  const client = queryClients.makeQueryClient()
  vi.spyOn(queryClients, 'makeQueryClient').mockReturnValue(client)
  const transaction = { id: 1, description: 'Private transaction', type: 'EXPENSE', amount: '10.00', category: 'OTHER', status: 'ACTIVE', source: 'MANUAL', visibility: 'PUBLIC', occurredAt: '2026-01-01', userAccountId: 1, guildId: 1 }
  let forbidden = !cached, calls = 0
  server.use(http.get('*/api/v1/guilds/1/transactions', () => { calls++; return forbidden ? new HttpResponse(null, { status: 403 }) : HttpResponse.json({ items: [transaction] }) }))
  render(<MemoryRouter initialEntries={['/transactions?guild=1']}><App initialUser={{ id: 1, name: 'Ana', email: 'ana@example.test' }} demo /></MemoryRouter>)
  if (cached) {
    expect(await screen.findByText('Private transaction')).toBeVisible()
    forbidden = true
    await act(async () => { await client.invalidateQueries({ queryKey: ['transactions'] }) })
  }
  expect(await screen.findByRole('heading', { name: 'Acesso negado' })).toBeVisible()
  expect(screen.queryByText('Private transaction')).toBeNull()
  expect(calls).toBe(cached ? 2 : 1)
  expect(screen.getByRole('button', { name: 'Sair' })).toBeEnabled()
})

it.each(['cancel', 'visibility'])('reports forbidden %s locally without changing the transaction', async (action) => {
  const success = vi.spyOn(toast, 'success')
  const error = vi.spyOn(toast, 'error')
  vi.spyOn(window, 'confirm').mockReturnValue(true)
  let calls = 0
  const denied = () => { calls++; return new HttpResponse(null, { status: 403 }) }
  server.use(http.delete('*/api/v1/guilds/1/transactions/:id', denied), http.patch('*/api/v1/guilds/1/transactions/:id/visibility', denied))
  render(<MemoryRouter initialEntries={['/transactions?guild=1']}><App initialUser={{ id: 1, name: 'Ana', email: 'ana@example.test' }} demo /></MemoryRouter>)
  expect(await screen.findByText('Supermercado Extra', {}, { timeout: 3000 })).toBeVisible()
  const label = action === 'cancel' ? 'Cancelar transação' : 'Pública'
  const buttons = await screen.findAllByTitle(label)
  await userEvent.click(buttons[0]!)
  await vi.waitFor(() => expect(error).toHaveBeenCalledWith('Você não tem permissão para realizar esta ação.'))
  expect(success).not.toHaveBeenCalled()
  expect(calls).toBe(1)
  expect(screen.getByText('Supermercado Extra')).toBeVisible()
  expect(screen.getByRole('heading', { name: 'Transações' })).toBeVisible()
  expect(screen.getByRole('button', { name: 'Sair' })).toBeEnabled()
})
