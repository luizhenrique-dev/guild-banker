import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { APIError, type RequestOptions } from '../shared/lib/api-client'
import { safeID } from '../shared/api/id'

export interface SessionUser { id: number; name: string; email: string; token?: string }
interface Session {
  user: SessionUser | null
  demo: boolean
  logout: () => void
  active: () => boolean
  request: <T>(operation: (options: RequestOptions) => Promise<T>, signal?: AbortSignal) => Promise<T>
}
const SessionContext = createContext<Session | null>(null)

export function SessionProvider({ children, initialUser = null, demo = false }: { children: ReactNode; initialUser?: SessionUser | null; demo?: boolean }) {
  const qc = useQueryClient()
  const [user, setUser] = useState(initialUser)
  const [controller] = useState(() => new AbortController())
  const logout = useCallback(() => {
    controller.abort()
    void qc.cancelQueries()
    qc.clear()
    setUser(null)
  }, [controller, qc])
  useEffect(() => () => {
    controller.abort()
    void qc.cancelQueries()
    qc.clear()
  }, [controller, qc])
  const active = useCallback(() => !!user && !controller.signal.aborted, [user, controller])
  const request = useCallback(async <T,>(operation: (options: RequestOptions) => Promise<T>, extraSignal?: AbortSignal): Promise<T> => {
    if (!user) throw new DOMException('Session ended', 'AbortError')
    const signal = extraSignal ? AbortSignal.any([controller.signal, extraSignal]) : controller.signal
    signal.throwIfAborted()
    try {
      const result = await operation({
        signal,
        headers: { 'X-User-ID': String(safeID(user.id)), 'X-User-Email': user.email, ...(user.token ? { Authorization: `Bearer ${user.token}` } : {}) },
      })
      signal.throwIfAborted()
      return result
    } catch (error) {
      if (error instanceof APIError && error.status === 401) logout()
      throw error
    }
  }, [user, controller, logout])
  return <SessionContext.Provider value={{ user, demo, logout, active, request }}>{children}</SessionContext.Provider>
}

export function useSession() {
  const session = useContext(SessionContext)
  if (!session) throw new Error('SessionProvider required')
  return session
}
