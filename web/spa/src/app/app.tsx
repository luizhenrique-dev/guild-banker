import { Navigate, Route, Routes, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import type { KeycloakSession } from './keycloak-session'
import { useSession, type SessionUser } from './session-provider'
import { useActiveGuild } from '../features/guilds/use-active-guild'
import { GuildSelectionState } from '../features/guilds/components/guild-selection-state'
import { ThemeProvider } from '@/components/theme-provider'
import { Toaster } from '@/components/ui/sonner'
import { Providers } from '../shared/components/providers'
import { Header } from '../shared/components/header'
import { Sidebar } from '../shared/components/sidebar'
import { FixedExpensesPageClient } from '../features/fixed-expenses/components/fixed-expenses-page-client'
import { TransactionsPageClient } from '../features/transactions/components/transactions-page-client'
import { ImportPageClient } from '../features/imports/components/import-page-client'
import { GuildPageClient } from '../features/guilds/components/guild-page-client'

import { HomePage } from '../features/home/components/home-page'

function AppRoutes() {
  const { user } = useSession()
  const { activeGuild } = useActiveGuild()
  const { search } = useLocation()
  const { t } = useTranslation()
  const protectedPage = (page: React.ReactNode) => user ? page : <main className="p-8"><h1>{t('session.missing')}</h1></main>
  const guildPage = (page: React.ReactNode) => protectedPage(activeGuild ? <div key={activeGuild.id}>{page}</div> : <GuildSelectionState />)
  return <Routes>
    <Route path="/" element={<Navigate to={'/home' + search} replace />} />
    <Route path="/home" element={protectedPage(<HomePage />)} />
    <Route path="/fixed-expenses" element={protectedPage(<FixedExpensesPageClient />)} />
    <Route path="/transactions" element={guildPage(<TransactionsPageClient />)} />
    <Route path="/import" element={guildPage(<ImportPageClient />)} />
    <Route path="/guild" element={guildPage(<GuildPageClient />)} />
    <Route path="*" element={<main className="p-8"><h1>{t('route.missing')}</h1></main>} />
  </Routes>
}

function AuthenticatedApp() {
  const { user, phase, retry } = useSession()
  const { t } = useTranslation()
  if (phase !== 'ready') return <main role={phase.endsWith('error') ? 'alert' : 'status'} className="p-8"><h1>{t(`session.${phase}`)}</h1>{phase.endsWith('error') && <button onClick={retry}>{t('common.retry')}</button>}</main>
  if (!user) return <main className="p-8"><h1>{t('session.missing')}</h1></main>
  return <>
    <Header />
    <div className="flex flex-col md:flex-row">
      <Sidebar />
      <div className="min-w-0 flex-1 min-h-[calc(100vh-64px)]"><AppRoutes /></div>
    </div>
    <Toaster />
  </>
}

export function App({ initialUser = null, demo = false, auth }: { initialUser?: SessionUser | null; demo?: boolean; auth?: KeycloakSession }) {
  return <ThemeProvider><Providers key={initialUser?.id} initialUser={initialUser} demo={demo} auth={auth}>
    <AuthenticatedApp />
  </Providers></ThemeProvider>
}
