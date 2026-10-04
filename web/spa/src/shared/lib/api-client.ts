export class APIError extends Error {
  constructor(message: string, public readonly status: number, public readonly kind: 'http' | 'network' | 'timeout' | 'invalid-response', options?: ErrorOptions) {
    super(message, options)
    this.name = 'APIError'
  }
}

export interface RequestOptions {
  method?: string
  body?: unknown
  params?: Record<string, string | number | undefined>
  headers?: HeadersInit
  signal?: AbortSignal
}

export async function apiClient<T = void>(path: string, options: RequestOptions = {}): Promise<T> {
  const url = new URL((import.meta.env.VITE_API_URL || '').replace(/\/+$/, '') + path, window.location.origin)
  for (const [key, value] of Object.entries(options.params ?? {})) {
    if (value !== undefined) url.searchParams.set(key, String(value))
  }
  const headers = new Headers(options.headers)
  const multipart = options.body instanceof FormData
  if (options.body !== undefined && !multipart) headers.set('Content-Type', 'application/json')
  if (multipart) headers.delete('Content-Type')
  const controller = new AbortController()
  const cancel = () => controller.abort(options.signal?.reason)
  if (options.signal?.aborted) cancel()
  else options.signal?.addEventListener('abort', cancel, { once: true })
  let timedOut = false
  const timeout = setTimeout(() => {
    timedOut = true
    controller.abort(new DOMException('Request timeout', 'TimeoutError'))
  }, 15000)
  try {
    const response = await fetch(url, {
      method: options.method ?? 'GET', headers, signal: controller.signal,
      body: options.body === undefined ? undefined : multipart ? options.body as FormData : JSON.stringify(options.body),
    })
    if (!response.ok) {
      let message = `HTTP ${response.status}`
      try {
        const payload: unknown = await response.json()
        if (payload && typeof payload === 'object' && 'error' in payload && typeof payload.error === 'string') message = payload.error
      } catch { /* HTTP status remains useful without a JSON envelope. */ }
      throw new APIError(message, response.status, 'http')
    }
    if (response.status === 204) return undefined as T
    try { return await response.json() as T }
    catch (cause) { throw new APIError('Invalid JSON response', response.status, 'invalid-response', { cause }) }
  } catch (cause) {
    if (options.signal?.aborted) throw options.signal.reason ?? new DOMException('Request cancelled', 'AbortError')
    if (timedOut) throw new APIError('Request timeout', 0, 'timeout', { cause })
    if (cause instanceof APIError) throw cause
    throw new APIError('Network request failed', 0, 'network', { cause })
  } finally {
    clearTimeout(timeout)
    options.signal?.removeEventListener('abort', cancel)
  }
}
