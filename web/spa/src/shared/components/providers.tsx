import { useState, type ReactNode } from 'react'
import { QueryClientProvider } from '@tanstack/react-query'
import { makeQueryClient } from '../lib/query-client'
import { I18nProvider } from './i18n-provider'
import { SessionProvider, type SessionUser } from '../../app/session-provider'

export function Providers({ children, initialUser, demo }: { children: ReactNode; initialUser?: SessionUser | null; demo?: boolean }) {
  const [queryClient] = useState(() => makeQueryClient())
  return <I18nProvider><QueryClientProvider client={queryClient}>
    <SessionProvider initialUser={initialUser} demo={demo}>{children}</SessionProvider>
  </QueryClientProvider></I18nProvider>
}
