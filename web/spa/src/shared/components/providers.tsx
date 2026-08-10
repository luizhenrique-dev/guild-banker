'use client'
// ============================================================
// 📌 PADRÃO: Client-side Providers
// Centraliza todos os providers que precisam de "use client".
// O layout raiz (server component) importa este wrapper.
// ============================================================
import { useState, type ReactNode } from 'react'
import { QueryClientProvider } from '@tanstack/react-query'
import { makeQueryClient } from '@/src/shared/lib/query-client'
import { MSWProvider } from './msw-provider'
import { I18nProvider } from './i18n-provider'

export function Providers({ children }: { children: ReactNode }) {
  // 📌 PADRÃO: QueryClient em useState garante que cada sessão
  // tem sua própria instância (evita compartilhamento entre requests SSR).
  const [queryClient] = useState(() => makeQueryClient())

  return (
    <I18nProvider>
      <MSWProvider>
        <QueryClientProvider client={queryClient}>
          {children}
        </QueryClientProvider>
      </MSWProvider>
    </I18nProvider>
  )
}
