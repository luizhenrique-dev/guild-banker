import { ActiveGuildProvider } from '@/src/features/guilds/active-guild-provider'
import { useState, type ReactNode } from 'react'
import { QueryClientProvider } from '@tanstack/react-query'
import { makeQueryClient } from '../lib/query-client'
import { I18nProvider } from './i18n-provider'
import type { KeycloakSession } from '../../app/keycloak-session'
import { SessionProvider, type SessionUser } from '../../app/session-provider'

export function Providers({ children, initialUser, demo, auth }: { children: ReactNode; initialUser?: SessionUser | null; demo?: boolean; auth?: KeycloakSession }) {
  const [queryClient] = useState(() => makeQueryClient())
  return <I18nProvider><QueryClientProvider client={queryClient}>
    <SessionProvider initialUser={initialUser} demo={demo} auth={auth}><ActiveGuildProvider>{children}</ActiveGuildProvider></SessionProvider>
  </QueryClientProvider></I18nProvider>
}
