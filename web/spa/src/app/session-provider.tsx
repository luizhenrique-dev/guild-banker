import { createContext, useCallback, useContext, useEffect, useRef, useState, type ReactNode } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { APIError, type RequestOptions } from '../shared/lib/api-client'
import { SessionExpiredError, type KeycloakSession } from './keycloak-session'
import { safeID } from '../shared/api/id'

export interface SessionUser { id: string | number; name: string; email: string }
type Phase = 'ready' | 'login' | 'logout' | 'login-error' | 'logout-error' | 'refresh-error'
interface Session {
  user: SessionUser | null
  demo: boolean
  phase: Phase
  retry: () => void
  logout: () => void
  active: () => boolean
  request: <T>(operation: (options: RequestOptions) => Promise<T>, signal?: AbortSignal) => Promise<T>
}
const SessionContext = createContext<Session | null>(null)

export function SessionProvider({ children, initialUser = null, demo = false, auth }: { children: ReactNode; initialUser?: SessionUser | null; demo?: boolean; auth?: KeycloakSession }) {
  const qc = useQueryClient()
  const [user, setUser] = useState(initialUser)
  const [phase, setPhase] = useState<Phase>('ready')
  const [controller] = useState(() => new AbortController())
  const mounted = useRef(true)
  const invalidate = useCallback(() => {
    controller.abort()
    void qc.cancelQueries()
    qc.clear()
    setUser(null)
  }, [controller, qc])
  const redirect = useCallback((action: 'login' | 'logout', explicit = false) => {
    if (!auth) return
    setPhase(action)
    void (action === 'login' ? auth.login(explicit) : auth.logout()).catch(() => { if (mounted.current) setPhase(`${action}-error`) })
  }, [auth])
  const end = useCallback((action: 'login' | 'logout') => {
    if (controller.signal.aborted) return
    invalidate()
    redirect(action)
  }, [controller, invalidate, redirect])
  const authError = useCallback((error: unknown) => {
    if (controller.signal.aborted) return
    if (error instanceof SessionExpiredError) end('login')
    else setPhase('refresh-error')
  }, [controller, end])
  useEffect(() => {
    const unsubscribe = auth?.onExpiration(() => { void auth.getToken().catch(authError) })
    return () => { unsubscribe?.() }
  }, [auth, authError])
  useEffect(() => () => {
    mounted.current = false
    controller.abort()
    void qc.cancelQueries()
    qc.clear()
  }, [controller, qc])
  const active = useCallback(() => !!user && !controller.signal.aborted, [user, controller])
  const request = useCallback(async <T,>(operation: (options: RequestOptions) => Promise<T>, extraSignal?: AbortSignal): Promise<T> => {
    if (!user) throw new DOMException('Session ended', 'AbortError')
    const signal = extraSignal ? AbortSignal.any([controller.signal, extraSignal]) : controller.signal
    signal.throwIfAborted()
    let token: string | undefined
    if (auth) {
      try { token = await auth.getToken() }
      catch (error) { authError(error); throw error }
    }
    signal.throwIfAborted()
    try {
      const result = await operation({
        signal,
        headers: auth ? { Authorization: `Bearer ${token}` } : import.meta.env.DEV && demo && typeof user.id === 'number' ? { 'X-User-ID': String(safeID(user.id)), 'X-User-Email': user.email } : {},
      })
      signal.throwIfAborted()
      auth?.accepted()
      return result
    } catch (error) {
      if (!signal.aborted && error instanceof APIError && error.status === 401) end('login')
      throw error
    }
  }, [user, controller, end, authError, auth, demo])
  const retry = () => {
    if (phase === 'logout-error') redirect('logout')
    else if (phase === 'login-error') redirect('login', true)
    else if (phase === 'refresh-error' && auth) {
      setPhase('login')
      void auth.getToken().then(() => { if (!controller.signal.aborted) setPhase('ready') }, authError)
    }
  }
  return <SessionContext.Provider value={{ user, demo, phase, retry, logout: () => end('logout'), active, request }}>{children}</SessionContext.Provider>
}

export function useSession() {
  const session = useContext(SessionContext)
  if (!session) throw new Error('SessionProvider required')
  return session
}
