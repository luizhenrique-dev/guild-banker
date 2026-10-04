import { beforeEach, expect, it, vi } from 'vitest'
import { startMocks } from './start-mocks'

const { start } = vi.hoisted(() => ({ start: vi.fn() }))
vi.mock('../mocks/browser', () => ({ worker: { start } }))
beforeEach(() => { start.mockReset() })

it('requires explicit opt-in in development', async () => {
  expect(await startMocks(true, undefined)).toBe(false)
  expect(await startMocks(true, 'false')).toBe(false)
  expect(start).not.toHaveBeenCalled()
  expect(await startMocks(true, 'true')).toBe(true)
  expect(start).toHaveBeenCalledOnce()
})
it('never starts mocks in production even with opt-in', async () => {
  expect(await startMocks(false, 'true')).toBe(false)
  expect(start).not.toHaveBeenCalled()
})
it('propagates startup failure so bootstrap cannot mount demo', async () => {
  start.mockRejectedValue(new Error('worker unavailable'))
  await expect(startMocks(true, 'true')).rejects.toThrow('worker unavailable')
})

it('rejects unknown API requests and bypasses fonts, assets and HMR', async () => {
  await startMocks(true, 'true')
  const options = start.mock.calls[0]![0]
  const print = { error: vi.fn() }
  for (const path of ['/assets/font.woff2', '/@vite/client', '/favicon.svg']) {
    options.onUnhandledRequest(new Request('http://localhost' + path), print)
  }
  expect(print.error).not.toHaveBeenCalled()
  options.onUnhandledRequest(new Request('http://localhost/api/v1/unknown'), print)
  expect(print.error).toHaveBeenCalledOnce()
})
