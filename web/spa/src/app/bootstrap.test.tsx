import { http, HttpResponse } from 'msw'
import { server } from '../mocks/server'
import { guildsFixtures } from '../mocks/fixtures/guilds'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { act, screen } from '@testing-library/react'
import { createRoot, type Root } from 'react-dom/client'
import i18n from '../shared/lib/i18n'
import { bootstrap } from './bootstrap'
import { createKeycloakSession } from './keycloak-session'
import { startMocks } from './start-mocks'

vi.mock('./start-mocks', () => ({ startMocks: vi.fn().mockResolvedValue(false) }))
vi.mock('./keycloak-session', () => ({ createKeycloakSession: vi.fn() }))
let root: Root
beforeEach(async () => { await i18n.changeLanguage('pt-BR'); vi.clearAllMocks(); const el = document.createElement('div'); document.body.append(el); root = createRoot(el) })
afterEach(() => { act(() => root.unmount()) })
it('does not mount shell or issue financial requests until OIDC resolves', async () => {
  server.use(http.get('*/api/v1/guilds', () => HttpResponse.json([guildsFixtures[0]])))
  let resolve!: (user: { id: string; name: string; email: string }) => void
  const auth = { initialize: vi.fn(() => new Promise((r) => { resolve = r })), getToken: vi.fn().mockResolvedValue('access'), accepted: vi.fn(), onExpiration: () => () => {} }
  vi.mocked(createKeycloakSession).mockReturnValue(auth as unknown as ReturnType<typeof createKeycloakSession>)
  const fetchSpy = vi.spyOn(globalThis, 'fetch')
  let pending!: Promise<void>
  await act(async () => { pending = bootstrap(root) })
  expect(screen.getByRole('status')).toHaveTextContent('Autenticando')
  expect(screen.queryByRole('navigation')).toBeNull()
  expect(fetchSpy).not.toHaveBeenCalled()
  window.history.replaceState(null, '', '/')
  await act(async () => { resolve({ id: 'sub', name: 'Ana', email: '' }); await pending })
  expect(await screen.findByRole('heading', { name: 'Bem-vindo(a) Ana' })).toBeVisible()
  expect(window.location.pathname).toBe('/home')
  fetchSpy.mockRestore()
})
it('fails closed with an explicit retry on initialization error', async () => {
  vi.mocked(createKeycloakSession).mockImplementation(() => { throw new Error('invalid config') })
  await act(async () => { await bootstrap(root) })
  expect(screen.getByRole('alert')).toHaveTextContent('Não foi possível autenticar')
  expect(screen.getByRole('button', { name: 'Tentar novamente' })).toBeVisible()
  expect(screen.queryByRole('navigation')).toBeNull()
})
it('does not fall back to demo when the mock worker fails', async () => {
  vi.mocked(startMocks).mockRejectedValueOnce(new Error('worker failed'))
  await act(async () => { await bootstrap(root) })
  expect(screen.getByRole('alert')).toBeVisible()
  expect(createKeycloakSession).not.toHaveBeenCalled()
})
