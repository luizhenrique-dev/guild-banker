'use client'
// ============================================================
// 📌 PADRÃO: MSW Provider
// Inicializa o Mock Service Worker no browser (client-side only).
// Em produção, o MSW não é carregado (NEXT_PUBLIC_USE_MOCK=false).
//
// ⚠️ ATENÇÃO: Para conectar na API real, basta setar
// NEXT_PUBLIC_USE_MOCK=false no .env.local.
// ============================================================
import { useEffect, useState, type ReactNode } from 'react'

export function MSWProvider({ children }: { children: ReactNode }) {
  const [ready, setReady] = useState(false)

  useEffect(() => {
    const useMock = process.env.NEXT_PUBLIC_USE_MOCK !== 'false'

    if (useMock && typeof window !== 'undefined') {
      import('@/src/mocks/browser').then(({ worker }) => {
        worker
          .start({
            onUnhandledRequest: 'bypass',
            serviceWorker: {
              url: '/mockServiceWorker.js',
            },
          })
          .then(() => {
            console.log('[MSW] Mock Service Worker ativo ✔')
            setReady(true)
          })
          .catch((err: unknown) => {
            console.warn('[MSW] Falha ao iniciar:', err)
            setReady(true)
          })
      })
    } else {
      setReady(true)
    }
  }, [])

  if (!ready) {
    return (
      <div className="flex h-screen items-center justify-center bg-[#f7f7f7]">
        <div className="flex items-center gap-3">
          <div className="h-5 w-5 animate-spin rounded-full border-2 border-[#0052ff] border-t-transparent" />
          <span className="text-sm text-[#5b616e]">Iniciando...</span>
        </div>
      </div>
    )
  }

  return <>{children}</>
}
