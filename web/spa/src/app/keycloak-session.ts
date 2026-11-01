import Keycloak from 'keycloak-js'

interface PublicConfig {
  VITE_KEYCLOAK_URL?: string
  VITE_KEYCLOAK_REALM?: string
  VITE_KEYCLOAK_CLIENT_ID?: string
}

export class SessionExpiredError extends Error {}

export function createKeycloakSession(env: PublicConfig) {
  const { VITE_KEYCLOAK_URL: url, VITE_KEYCLOAK_REALM: realm, VITE_KEYCLOAK_CLIENT_ID: clientId } = env
  if (!url || !realm?.trim() || !clientId?.trim()) throw new Error('Missing Keycloak configuration')
  const server = new URL(url)
  if (!['http:', 'https:'].includes(server.protocol) || server.username || server.password || server.search || server.hash) {
    throw new Error('Invalid Keycloak URL')
  }
  const recoveryKey = `oidc-recovery:${url}:${realm}:${clientId}`
  function clearRecovery() {
    try { sessionStorage.removeItem(recoveryKey) }
    catch { /* Storage can be unavailable; successful API work must remain successful. */ }
  }
  const sdk = new Keycloak({ url, realm, clientId })
  const redirectUri = () => window.location.origin + window.location.pathname + window.location.search + window.location.hash
  let expirationListener: (() => void) | undefined
  sdk.onTokenExpired = () => expirationListener?.()
  let initialization: ReturnType<typeof initialize> | undefined
  let refreshing: Promise<string> | undefined
  let logoutUrl: string | undefined
  let ended = false
  async function initialize() {
    const callback = [window.location.search.slice(1), window.location.hash.slice(1)].some((part) => {
      const params = new URLSearchParams(part)
      return params.has('code') || params.has('error') || params.has('state')
    })
    const authenticated = await sdk.init({ onLoad: callback ? undefined : 'login-required', flow: 'standard', pkceMethod: 'S256', checkLoginIframe: false, redirectUri: redirectUri() })
    const sub = sdk.tokenParsed?.sub
    if (!authenticated || !sub) throw new SessionExpiredError('Unauthenticated')
    // The provider owns reauthentication after initialization, including refresh rejection.
    sdk.loginRequired = false
    const claims = sdk.tokenParsed!
    return { id: sub, name: typeof claims.name === 'string' ? claims.name : typeof claims.preferred_username === 'string' ? claims.preferred_username : '', email: typeof claims.email === 'string' ? claims.email : '' }
  }
  return {
    initialize: () => initialization ??= initialize(),
    async getToken() {
      if (ended) throw new SessionExpiredError('Session ended')
      return refreshing ??= (async () => {
        try {
          await sdk.updateToken(30)
          if (ended) {
            await sdk.clearToken()
            throw new SessionExpiredError('Session ended')
          }
          if (!sdk.authenticated || !sdk.token) throw new SessionExpiredError('Session expired')
          return sdk.token
        } catch (error) {
          if (!sdk.authenticated) throw new SessionExpiredError('Session expired')
          throw error
        }
      })().finally(() => { refreshing = undefined })
    },
    async login(explicit = false) {
      if (explicit) {
        try { sessionStorage.setItem(recoveryKey, 'attempted') }
        catch { /* A user-initiated login may proceed without automatic recovery. */ }
      } else {
        if (sessionStorage.getItem(recoveryKey)) throw new Error('Automatic login already attempted')
        sessionStorage.setItem(recoveryKey, 'attempted')
      }
      return sdk.login({ redirectUri: redirectUri() })
    },
    accepted: clearRecovery,
    onExpiration(listener: () => void) {
      expirationListener = listener
      return () => { if (expirationListener === listener) expirationListener = undefined }
    },
    async logout() {
      ended = true
      clearRecovery()
      logoutUrl ??= sdk.createLogoutUrl({ redirectUri: window.location.origin + '/' })
      expirationListener = undefined
      await sdk.clearToken()
      // Detect an unavailable server before leaving the UI, retaining an explicit retry.
      const response = await fetch(`${url.replace(/\/+$/, '')}/realms/${encodeURIComponent(realm)}/.well-known/openid-configuration`, { cache: 'no-store', signal: AbortSignal.timeout(5000) })
      if (!response.ok) throw new Error('Logout server unavailable')
      window.location.replace(logoutUrl)
    },
  }
}
export type KeycloakSession = ReturnType<typeof createKeycloakSession>
