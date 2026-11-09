import { act, render, renderHook, waitFor } from '@testing-library/react'
import { QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, expect, it, vi } from 'vitest'
import { toast } from 'sonner'
import { SessionProvider } from '@/src/app/session-provider'
import { makeQueryClient } from '@/src/shared/lib/query-client'
import { fixedExpensesApi } from './fixed-expenses.api'
import { fixedExpenseKeys, useCreateFixedExpense, useFixedExpenses } from './use-fixed-expenses'

afterEach(() => vi.restoreAllMocks())
function setup(guild = '999') {
  const client = makeQueryClient()
  const wrapper = ({ children }: { children: React.ReactNode }) => <MemoryRouter initialEntries={['/fixed-expenses?guild=' + guild]}><QueryClientProvider client={client}><SessionProvider initialUser={{ id: 1, name: 'Ana', email: 'ana@example.test' }} demo>{children}</SessionProvider></QueryClientProvider></MemoryRouter>
  return { client, wrapper }
}
it('keys distinguish both user and guild', () => {
  expect(fixedExpenseKeys.list(1, 1)).toEqual(['fixed-expenses', 1, 1])
  expect(fixedExpenseKeys.list(1, 2)).not.toEqual(fixedExpenseKeys.list(1, 1))
  expect(fixedExpenseKeys.list(2, 1)).not.toEqual(fixedExpenseKeys.list(1, 1))
})
it('does not query expenses for an invalid selection', async () => {
  const list = vi.spyOn(fixedExpensesApi, 'list')
  const { wrapper, client } = setup()
  const { result } = renderHook(useFixedExpenses, { wrapper })
  await waitFor(() => expect(client.getQueryData(['guilds', 1])).toBeDefined())
  expect(result.current.fetchStatus).toBe('idle')
  expect(list).not.toHaveBeenCalled()
})
it('a late mutation invalidates only its origin and never toasts after unmount', async () => {
  const { wrapper, client } = setup('1')
  const invalidate = vi.spyOn(client, 'invalidateQueries')
  const success = vi.spyOn(toast, 'success')
  let finish!: (value: Awaited<ReturnType<typeof fixedExpensesApi.create>>) => void
  vi.spyOn(fixedExpensesApi, 'create').mockImplementation(() => new Promise((resolve) => { finish = resolve }))
  let mutation!: ReturnType<typeof useCreateFixedExpense>
  function Probe() { mutation = useCreateFixedExpense(); return null }
  const view = render(<Probe />, { wrapper })
  let pending!: Promise<unknown>
  act(() => { pending = mutation.mutateAsync({ guildId: 1, input: { name: 'Rent', amount: '10', due_day: 1, category: 'HOUSING' } }) })
  await waitFor(() => expect(finish).toBeTypeOf('function'))
  view.rerender(<></>)
  await act(async () => { finish({ guild_id: 1, id: 100, name: 'Rent', amount: '10', due_day: 1, category: 'HOUSING', status: 'ACTIVE' }); await pending })
  expect(invalidate).toHaveBeenCalledExactlyOnceWith({ queryKey: ['fixed-expenses', 1, 1] })
  expect(success).not.toHaveBeenCalled()
})
