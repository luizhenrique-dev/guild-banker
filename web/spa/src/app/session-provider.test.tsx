import { act, renderHook, render, screen, waitFor } from '@testing-library/react'
import { QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { expect, it, vi, afterEach } from 'vitest'
import { http, HttpResponse } from 'msw'
import { toast } from 'sonner'
import { SessionProvider, useSession } from './session-provider'
import { makeQueryClient } from '../shared/lib/query-client'
import { APIError, apiClient } from '../shared/lib/api-client'
import { useCreateFixedExpense } from '../features/fixed-expenses/api/use-fixed-expenses'
import { fixedExpensesApi } from '../features/fixed-expenses/api/fixed-expenses.api'
import { server } from '../mocks/server'
import { App } from './app'
import i18n from '../shared/lib/i18n'

afterEach(() => { vi.restoreAllMocks() })
const user = { id: 1, name: 'Ana', email: 'ana@familia.com' }
function setup() {
  const qc = makeQueryClient()
  const wrapper = ({ children }: { children: React.ReactNode }) => <QueryClientProvider client={qc}><SessionProvider initialUser={user}>{children}</SessionProvider></QueryClientProvider>
  return { qc, wrapper }
}

it('logout aborts active IO, clears caches and rejects a late response', async () => {
  const { qc, wrapper } = setup()
  const { result } = renderHook(() => useSession(), { wrapper })
  qc.setQueryData(['fixed-expenses', 1], ['private'])
  let finish!: (value: string) => void
  let signal!: AbortSignal
  const pending = result.current.request((options) => {
    signal = options.signal!
    return new Promise<string>((resolve) => { finish = resolve })
  })
  const rejection = expect(pending).rejects.toMatchObject({ name: 'AbortError' })
  act(() => result.current.logout())
  expect(signal.aborted).toBe(true)
  expect(qc.getQueryCache().getAll()).toHaveLength(0)
  finish('late private data')
  await rejection
  expect(result.current.user).toBeNull()
})
it('a 401 ends the session and clears cached data', async () => {
  const { qc, wrapper } = setup()
  const { result } = renderHook(() => useSession(), { wrapper })
  qc.setQueryData(['imports', 1, 1], { secret: true })
  server.use(http.get('*/api/session', () => HttpResponse.json({ error: 'expired' }, { status: 401 })))
  await act(async () => {
    await expect(result.current.request((options) => apiClient('/api/session', options))).rejects.toMatchObject({ status: 401 })
  })
  expect(result.current.user).toBeNull()
  expect(qc.getQueryCache().getAll()).toHaveLength(0)
})
it('late mutation after logout cannot recreate cache or show notifications', async () => {
  const { qc, wrapper } = setup()
  let finish!: (value: Awaited<ReturnType<typeof fixedExpensesApi.create>>) => void
  vi.spyOn(fixedExpensesApi, 'create').mockImplementation(() => new Promise((resolve) => { finish = resolve }))
  const success = vi.spyOn(toast, 'success')
  const failure = vi.spyOn(toast, 'error')
  const { result } = renderHook(() => ({ session: useSession(), create: useCreateFixedExpense() }), { wrapper })
  let pending!: Promise<unknown>
  act(() => { pending = result.current.create.mutateAsync({ name: 'Rent', amount: '10', due_day: 1, category: 'HOUSING' }) })
  const rejection = expect(pending).rejects.toMatchObject({ name: 'AbortError' })
  await waitFor(() => expect(finish).toBeTypeOf('function'))
  act(() => result.current.session.logout())
  await act(async () => {
    finish({ id: 100, name: 'Rent', amount: '10', due_day: 1, category: 'HOUSING', status: 'ACTIVE' })
    await rejection
  })
  expect(qc.getQueryCache().getAll()).toHaveLength(0)
  expect(success).not.toHaveBeenCalled()
  expect(failure).not.toHaveBeenCalled()
})
it('combines query cancellation with the session lifetime', async () => {
  const { wrapper } = setup()
  const { result } = renderHook(() => useSession(), { wrapper })
  const controller = new AbortController()
  let finish!: () => void
  let signal!: AbortSignal
  const pending = result.current.request((options) => {
    signal = options.signal!
    return new Promise<void>((resolve) => { finish = resolve })
  }, controller.signal)
  const rejection = expect(pending).rejects.toMatchObject({ name: 'AbortError' })
  controller.abort(); finish()
  await rejection
  expect(signal.aborted).toBe(true)
  expect(result.current.active()).toBe(true)
})
it('switching users cannot display the previous personal list', async () => {
  await i18n.changeLanguage('pt-BR')
  const { rerender } = render(<MemoryRouter initialEntries={['/fixed-expenses']}><App initialUser={user} demo /></MemoryRouter>)
  await screen.findByText('Aluguel')
  rerender(<MemoryRouter initialEntries={['/fixed-expenses']}><App initialUser={{ id: 2, name: 'Other', email: 'other@example.com' }} demo /></MemoryRouter>)
  await screen.findByText('Nenhuma despesa fixa cadastrada.')
  expect(screen.queryByText('Aluguel')).not.toBeInTheDocument()
})
it('retries only eligible GET failures once and never retries mutations', async () => {
  const { qc } = setup()
  for (const status of [400, 401, 403, 404, 409, 422, 501]) {
    const query = vi.fn().mockRejectedValue(new APIError('failure', status, 'http'))
    await expect(qc.fetchQuery({ queryKey: [status], queryFn: query })).rejects.toBeInstanceOf(APIError)
    expect(query).toHaveBeenCalledOnce()
  }
  const query = vi.fn().mockRejectedValue(new APIError('failure', 503, 'http'))
  await expect(qc.fetchQuery({ queryKey: [503], queryFn: query })).rejects.toBeInstanceOf(APIError)
  expect(query).toHaveBeenCalledTimes(2)
  const mutation = vi.fn().mockRejectedValue(new APIError('failure', 503, 'http'))
  await expect(qc.getMutationCache().build(qc, { mutationFn: mutation }).execute(undefined)).rejects.toBeInstanceOf(APIError)
  expect(mutation).toHaveBeenCalledOnce()
})
