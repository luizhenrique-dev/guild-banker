// @vitest-environment node
import { afterEach, expect, it, vi } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '../../mocks/server'
import { apiClient } from './api-client'

vi.stubGlobal('window', { location: { origin: 'http://localhost' } })

afterEach(() => { vi.useRealTimers(); vi.restoreAllMocks(); vi.unstubAllEnvs() })

it('serializes JSON, query strings and explicit headers', async () => {
  server.use(http.post('*/api/test', async ({ request }) => {
    expect(request.headers.get('X-User-ID')).toBe('7')
    expect(request.headers.get('Content-Type')).toBe('application/json')
    expect(new URL(request.url).searchParams.get('cursor')).toBe('opaque+/=')
    return HttpResponse.json(await request.json())
  }))
  expect(await apiClient('*/api/test', { method: 'POST', body: { amount: '10.20' }, headers: { 'X-User-ID': '7' }, params: { cursor: 'opaque+/=' } })).toEqual({ amount: '10.20' })
})
it('handles 204 without parsing JSON', async () => {
  server.use(http.patch('*/api/test', () => new HttpResponse(null, { status: 204 })))
  expect(await apiClient('*/api/test', { method: 'PATCH' })).toBeUndefined()
})
it('leaves multipart boundary to the browser', async () => {
  server.use(http.post('*/api/test', async ({ request }) => {
    expect(request.headers.get('Content-Type')).toContain('multipart/form-data; boundary=')
    expect((await request.formData()).get('file')).toBe('csv content')
    return HttpResponse.json({ ok: true })
  }))
  const body = new FormData(); body.append('file', 'csv content')
  expect(await apiClient('*/api/test', { method: 'POST', body })).toEqual({ ok: true })
})
it.each([400, 401, 403, 404, 409, 422, 501, 500, 503])('exposes status %s and server message', async (status) => {
  server.use(http.get('*/api/test', () => HttpResponse.json({ error: 'rejected' }, { status })))
  await expect(apiClient('*/api/test')).rejects.toMatchObject({ status, message: 'rejected', kind: 'http' })
})
it('handles non-JSON errors and invalid success JSON', async () => {
  server.use(http.get('*/api/test', () => new HttpResponse('upstream failed', { status: 502 })))
  await expect(apiClient('*/api/test')).rejects.toMatchObject({ status: 502, kind: 'http' })
  server.use(http.get('*/api/test', () => new HttpResponse('{bad json')))
  await expect(apiClient('*/api/test')).rejects.toMatchObject({ kind: 'invalid-response' })
})
it('classifies network failure without retrying', async () => {
  let calls = 0
  server.use(http.post('*/api/test', () => { calls++; return HttpResponse.error() }))
  await expect(apiClient('*/api/test', { method: 'POST' })).rejects.toMatchObject({ kind: 'network' })
  expect(calls).toBe(1)
})
it('cancels active requests and removes listeners', async () => {
  const controller = new AbortController()
  const remove = vi.spyOn(controller.signal, 'removeEventListener')
  vi.spyOn(globalThis, 'fetch').mockImplementation((_url, options) => new Promise((_resolve, reject) => {
    options?.signal?.addEventListener('abort', () => reject(options.signal?.reason), { once: true })
  }))
  const result = apiClient('*/api/test', { signal: controller.signal })
  controller.abort()
  await expect(result).rejects.toMatchObject({ name: 'AbortError' })
  expect(remove).toHaveBeenCalledWith('abort', expect.any(Function))
})
it('times out after 15s and cleans timers', async () => {
  vi.useFakeTimers()
  vi.spyOn(globalThis, 'fetch').mockImplementation((_url, options) => new Promise((_resolve, reject) => {
    options?.signal?.addEventListener('abort', () => reject(options.signal?.reason), { once: true })
  }))
  const result = expect(apiClient('*/api/test')).rejects.toMatchObject({ kind: 'timeout' })
  await vi.advanceTimersByTimeAsync(15000)
  await result
  expect(vi.getTimerCount()).toBe(0)
})

it('normalizes a configured API origin with a trailing slash', async () => {
  vi.stubEnv('VITE_API_URL', 'http://api.example/')
  server.use(http.get('http://api.example/api/test', () => HttpResponse.json({ ok: true })))
  expect(await apiClient('/api/test')).toEqual({ ok: true })
})
