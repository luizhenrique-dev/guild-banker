import { Navigate, Route, Routes, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '../shared/store/auth.store'
import { ThemeProvider } from '@/components/theme-provider'
import { Toaster } from '@/components/ui/sonner'
import { Providers } from '../shared/components/providers'
import { Header } from '../shared/components/header'
import { Sidebar } from '../shared/components/sidebar'
import { FixedExpensesPageClient } from '../features/fixed-expenses/components/fixed-expenses-page-client'
import { TransactionsPageClient } from '../features/transactions/components/transactions-page-client'
import { ImportPageClient } from '../features/imports/components/import-page-client'
import { GuildPageClient } from '../features/guilds/components/guild-page-client'

function AppRoutes() {
  const user = useAuthStore((state) => state.user)
  const { search } = useLocation()
  const { t } = useTranslation()
  const protectedPage = (page: React.ReactNode) => user ? page : <main className="p-8"><h1>{t('session.missing')}</h1></main>
  return <Routes>
    <Route path="/" element={<Navigate to={'/fixed-expenses' + search} replace />} />
    <Route path="/fixed-expenses" element={protectedPage(<FixedExpensesPageClient />)} />
    <Route path="/transactions" element={protectedPage(<TransactionsPageClient />)} />
    <Route path="/import" element={protectedPage(<ImportPageClient />)} />
    <Route path="/guild" element={protectedPage(<GuildPageClient />)} />
    <Route path="*" element={<main className="p-8"><h1>{t('route.missing')}</h1></main>} />
  </Routes>
}

export function App() {
  return <ThemeProvider><Providers>
    <Header />
    <div className="flex flex-col md:flex-row">
      <Sidebar />
      <div className="min-w-0 flex-1 min-h-[calc(100vh-64px)]"><AppRoutes /></div>
    </div>
    <Toaster />
  </Providers></ThemeProvider>
}
