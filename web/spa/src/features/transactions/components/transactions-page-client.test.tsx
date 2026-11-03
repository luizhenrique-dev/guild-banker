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
