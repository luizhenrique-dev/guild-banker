import { beforeEach, expect, it, vi } from 'vitest'
import Keycloak, { type KeycloakTokenParsed } from 'keycloak-js'
import { createKeycloakSession } from './keycloak-session'

vi.mock('keycloak-js', () => ({ default: vi.fn() }))
const env = { VITE_KEYCLOAK_URL: 'http://localhost:8080', VITE_KEYCLOAK_REALM: 'guild-banker', VITE_KEYCLOAK_CLIENT_ID: 'guild-banker-ui' }
const sdk = {
  init: vi.fn().mockResolvedValue(true), tokenParsed: { sub: 'user-sub', name: 'Ana', email: 'ana@example.com' } as KeycloakTokenParsed,
  token: 'access', authenticated: true, login: vi.fn().mockResolvedValue(undefined),
  updateToken: vi.fn().mockResolvedValue(false), logout: vi.fn().mockResolvedValue(undefined),
}
beforeEach(() => { vi.clearAllMocks(); vi.mocked(Keycloak).mockImplementation(function () { return sdk as unknown as Keycloak }) })
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
it('logs out to the UI root, independently from the current path', async () => {
  const auth = createKeycloakSession(env)
  await auth.initialize()
  await auth.logout()
  expect(sdk.logout).toHaveBeenCalledWith({ redirectUri: 'http://localhost/' })
})

it('does not restart login automatically for an invalid callback', async () => {
  window.history.replaceState(null, '', '/home#code=invalid&state=unknown')
  sdk.init.mockResolvedValueOnce(false)
  await expect(createKeycloakSession(env).initialize()).rejects.toThrow()
  expect(sdk.init).toHaveBeenLastCalledWith(expect.objectContaining({ onLoad: undefined }))
  expect(sdk.login).not.toHaveBeenCalled()
})
