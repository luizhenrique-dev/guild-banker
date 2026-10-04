import { act, renderHook, waitFor } from '@testing-library/react'
import { QueryClientProvider } from '@tanstack/react-query'
import { afterEach, expect, it, vi } from 'vitest'
import { toast } from 'sonner'
import { SessionProvider } from '@/src/app/session-provider'
import { makeQueryClient } from '@/src/shared/lib/query-client'
import { importsApi } from './imports.api'
import { useImports } from './use-imports'
import { useToggleVisibility } from '../../transactions/api/use-transactions'
import { transactionsApi } from '../../transactions/api/transactions.api'
import type { UploadImportResponse } from '../types'

afterEach(() => { vi.restoreAllMocks() })
function setup() {
  const qc = makeQueryClient()
  const wrapper = ({ children }: { children: React.ReactNode }) => <QueryClientProvider client={qc}><SessionProvider initialUser={{ id: 1, email: 'user@example.com', name: 'User' }}>{children}</SessionProvider></QueryClientProvider>
  return { qc, wrapper }
}
const batch: UploadImportResponse = { importId: 1, status: 'PENDING_REVIEW', fileName: 'test.csv', summary: { parsed: 0, candidates: 0, duplicates: 0, skippedZero: 0 }, items: [] }

it('changing guild discards the revision and prevents late upload from recreating it', async () => {
  const { qc, wrapper } = setup()
  let finish!: (value: UploadImportResponse) => void
  vi.spyOn(importsApi, 'upload').mockImplementation(() => new Promise((resolve) => { finish = resolve }))
  const success = vi.spyOn(toast, 'success')
  const error = vi.spyOn(toast, 'error')
  const { result, rerender } = renderHook(({ guildID }) => useImports(guildID), { wrapper, initialProps: { guildID: 1 } })
  let pending!: Promise<unknown>
  act(() => { pending = result.current.upload.mutateAsync(new File(['csv'], 'test.csv')) })
  const rejection = expect(pending).rejects.toMatchObject({ name: 'AbortError' })
  await waitFor(() => expect(finish).toBeTypeOf('function'))
  rerender({ guildID: 2 })
  await act(async () => { finish(batch); await rejection })
  expect(qc.getQueryData(['imports', 1, 1])).toBeUndefined()
  expect(result.current.revision).toBeNull()
  expect(success).not.toHaveBeenCalled()
  expect(error).not.toHaveBeenCalled()
})
it('confirmation invalidates only the original guild transaction cache', async () => {
  const { qc, wrapper } = setup()
  vi.spyOn(importsApi, 'confirm').mockResolvedValue({ importId: 1, status: 'COMPLETED', created: 2, skipped: 0 })
  qc.setQueryData(['transactions', 1, 1, {}], ['original'])
  qc.setQueryData(['transactions', 1, 2, {}], ['other guild'])
  const { result } = renderHook(() => useImports(1), { wrapper })
  await act(async () => { await result.current.confirm.mutateAsync(1) })
  expect(qc.getQueryState(['transactions', 1, 1, {}])?.isInvalidated).toBe(true)
  expect(qc.getQueryState(['transactions', 1, 2, {}])?.isInvalidated).toBe(false)
})
it('late transaction mutation invalidates captured guild and leaves other contexts fresh', async () => {
  const { qc, wrapper } = setup()
  let finish!: (value: { id: number; visibility: 'PRIVATE' }) => void
  vi.spyOn(transactionsApi, 'setVisibility').mockImplementation(() => new Promise((resolve) => { finish = resolve }))
  for (const [userID, guildID] of [[1, 1], [1, 2], [2, 1]]) qc.setQueryData(['transactions', userID, guildID, {}], ['data'])
  const { result } = renderHook(() => useToggleVisibility(), { wrapper })
  let pending!: Promise<unknown>
  act(() => { pending = result.current.mutateAsync({ id: 1, guildId: 1, visibility: 'PRIVATE' }) })
  await waitFor(() => expect(finish).toBeTypeOf('function'))
  await act(async () => { finish({ id: 1, visibility: 'PRIVATE' }); await pending })
  expect(qc.getQueryState(['transactions', 1, 1, {}])?.isInvalidated).toBe(true)
  expect(qc.getQueryState(['transactions', 1, 2, {}])?.isInvalidated).toBe(false)
  expect(qc.getQueryState(['transactions', 2, 1, {}])?.isInvalidated).toBe(false)
})
it('confirmation finishing after a guild switch invalidates the original list without restoring its revision', async () => {
  const { qc, wrapper } = setup()
  let finish!: (value: { importId: number; status: 'COMPLETED'; created: number; skipped: number }) => void
  vi.spyOn(importsApi, 'confirm').mockImplementation(() => new Promise((resolve) => { finish = resolve }))
  const success = vi.spyOn(toast, 'success')
  qc.setQueryData(['transactions', 1, 1, {}], ['original'])
  qc.setQueryData(['transactions', 1, 2, {}], ['other guild'])
  const { result, rerender } = renderHook(({ guildID }) => useImports(guildID), { wrapper, initialProps: { guildID: 1 } })
  let pending!: Promise<unknown>
  act(() => { pending = result.current.confirm.mutateAsync(1) })
  const rejection = expect(pending).rejects.toMatchObject({ name: 'AbortError' })
  await waitFor(() => expect(finish).toBeTypeOf('function'))
  rerender({ guildID: 2 })
  await act(async () => { finish({ importId: 1, status: 'COMPLETED', created: 2, skipped: 0 }); await rejection })
  expect(qc.getQueryState(['transactions', 1, 1, {}])?.isInvalidated).toBe(true)
  expect(qc.getQueryState(['transactions', 1, 2, {}])?.isInvalidated).toBe(false)
  expect(qc.getQueryData(['imports', 1, 1])).toBeUndefined()
  expect(success).not.toHaveBeenCalled()
})
