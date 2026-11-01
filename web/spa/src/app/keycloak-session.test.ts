import { beforeEach, expect, it, vi } from 'vitest'
import { http, HttpResponse } from 'msw'
import { server } from '../mocks/server'
import Keycloak, { type KeycloakTokenParsed } from 'keycloak-js'
import { createKeycloakSession } from './keycloak-session'

vi.mock('keycloak-js', () => ({ default: vi.fn() }))
const env = { VITE_KEYCLOAK_URL: 'http://localhost:8080', VITE_KEYCLOAK_REALM: 'guild-banker', VITE_KEYCLOAK_CLIENT_ID: 'guild-banker-ui' }
const sdk = {
  init: vi.fn().mockResolvedValue(true), tokenParsed: { sub: 'user-sub', name: 'Ana', email: 'ana@example.com' } as KeycloakTokenParsed,
  token: 'access', authenticated: true, login: vi.fn().mockResolvedValue(undefined),
  createLogoutUrl: vi.fn().mockReturnValue('http://localhost:8080/logout?id_token_hint=hint'), clearToken: vi.fn().mockResolvedValue(undefined),
  updateToken: vi.fn().mockResolvedValue(false), logout: vi.fn().mockResolvedValue(undefined),
}
beforeEach(() => { sessionStorage.clear(); vi.clearAllMocks(); vi.mocked(Keycloak).mockImplementation(function () { return sdk as unknown as Keycloak }) })
it('validates public config before constructing the SDK', () => {
  expect(() => createKeycloakSession({})).toThrow()
  expect(() => createKeycloakSession({ ...env, VITE_KEYCLOAK_URL: 'javascript:alert(1)' })).toThrow()
  expect(Keycloak).not.toHaveBeenCalled()
})
it('initializes only once with Code PKCE and a local deep link', async () => {
  window.history.replaceState(null, '', '/transactions?guild=2&category=OTHER#details')
  const auth = createKeycloakSession(env)
  expect(await auth.initialize()).toEqual({ id: 'user-sub', name: 'Ana', email: 'ana@example.com' })
  await auth.initialize()
  expect(sdk.init).toHaveBeenCalledOnce()
  expect((sdk as unknown as Keycloak).loginRequired).toBe(false)
  expect(sdk.init).toHaveBeenCalledWith(expect.objectContaining({ onLoad: 'login-required', flow: 'standard', pkceMethod: 'S256', redirectUri: 'http://localhost/transactions?guild=2&category=OTHER#details' }))
  expect(Keycloak).toHaveBeenCalledWith({ url: env.VITE_KEYCLOAK_URL, realm: 'guild-banker', clientId: 'guild-banker-ui' })
})
it('uses username fallback and rejects missing sub', async () => {
  const previous = sdk.tokenParsed
  sdk.tokenParsed = { sub: 'user-sub', preferred_username: 'ana' } as typeof previous
  expect((await createKeycloakSession(env).initialize()).name).toBe('ana')
  sdk.tokenParsed = { name: 'Ana' } as typeof previous
  await expect(createKeycloakSession(env).initialize()).rejects.toThrow()
  sdk.tokenParsed = previous
})
it('shares concurrent refresh and reads the current access token', async () => {
  const auth = createKeycloakSession(env)
  await auth.initialize()
  const requests = [auth.getToken(), auth.getToken()]
  expect(await Promise.all(requests)).toEqual(['access', 'access'])
  expect(sdk.updateToken).toHaveBeenCalledOnce()
  expect(sdk.updateToken).toHaveBeenCalledWith(30)
})
it('prepares logout before clearing tokens and exposes unavailable server for retry', async () => {
  server.use(http.get('http://localhost:8080/realms/guild-banker/.well-known/openid-configuration', () => new HttpResponse(null, { status: 503 })))
  const auth = createKeycloakSession(env)
  await auth.initialize()
  await expect(auth.logout()).rejects.toThrow('Logout server unavailable')
  expect(sdk.createLogoutUrl).toHaveBeenCalledWith({ redirectUri: 'http://localhost/' })
  expect(sdk.createLogoutUrl.mock.invocationCallOrder[0]).toBeLessThan(sdk.clearToken.mock.invocationCallOrder[0]!)
  await expect(auth.logout()).rejects.toThrow()
  expect(sdk.createLogoutUrl).toHaveBeenCalledOnce()
})
it('does not restart login automatically for an invalid callback', async () => {
  window.history.replaceState(null, '', '/home#code=invalid&state=unknown')
  sdk.init.mockResolvedValueOnce(false)
  await expect(createKeycloakSession(env).initialize()).rejects.toThrow()
  expect(sdk.init).toHaveBeenLastCalledWith(expect.objectContaining({ onLoad: undefined }))
  expect(sdk.login).not.toHaveBeenCalled()
})

it('installs expiration handling before init so the SDK schedules the initial token timer', async () => {
  sdk.init.mockImplementationOnce(async () => {
    expect((sdk as unknown as Keycloak).onTokenExpired).toBeTypeOf('function')
    return true
  })
  const auth = createKeycloakSession(env)
  window.history.replaceState(null, '', '/home')
  await auth.initialize()
  const listener = vi.fn()
  const unsubscribe = auth.onExpiration(listener)
  ;(sdk as unknown as Keycloak).onTokenExpired?.()
  expect(listener).toHaveBeenCalledOnce()
  unsubscribe()
  ;(sdk as unknown as Keycloak).onTokenExpired?.()
  expect(listener).toHaveBeenCalledOnce()
})

it('bounds automatic reauthentication across page lifetimes, resetting only after API success or explicit retry', async () => {
  window.history.replaceState(null, '', '/home')
  const auth = createKeycloakSession(env)
  await auth.login()
  const afterReload = createKeycloakSession(env)
  await expect(afterReload.login()).rejects.toThrow('Automatic login already attempted')
  await afterReload.login(true)
  afterReload.accepted()
  await afterReload.login()
  expect(sdk.login).toHaveBeenCalledTimes(3)
})

it('treats blocked recovery storage as a handled login failure without failing accepted API work', async () => {
  const auth = createKeycloakSession(env)
  const get = vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => { throw new Error('blocked') })
  await expect(auth.login()).rejects.toThrow('blocked')
  get.mockRestore()
  const remove = vi.spyOn(Storage.prototype, 'removeItem').mockImplementation(() => { throw new Error('blocked') })
  expect(() => auth.accepted()).not.toThrow()
  remove.mockRestore()
})
it('discards a refresh response that arrives after logout', async () => {
  server.use(http.get('http://localhost:8080/realms/guild-banker/.well-known/openid-configuration', () => new HttpResponse(null, { status: 503 })))
  const auth = createKeycloakSession(env)
  let finish!: (value: boolean) => void
  sdk.updateToken.mockImplementationOnce(() => new Promise((resolve) => { finish = resolve }))
  const pending = auth.getToken()
  const rejection = expect(pending).rejects.toThrow('Session ended')
  await expect(auth.logout()).rejects.toThrow()
  finish(true)
  await rejection
  expect(sdk.clearToken).toHaveBeenCalledTimes(2)
  await expect(auth.getToken()).rejects.toThrow('Session ended')
})
