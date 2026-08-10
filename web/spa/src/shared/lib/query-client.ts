// ============================================================
// 📌 PADRÃO: TanStack Query Client
// ============================================================
// Por que TanStack Query em vez de useEffect + fetch?
// 1. Cache automático — evita re-fetches desnecessários
// 2. Deduplicação — múltiplos componentes usando o mesmo hook
//    geram apenas uma request
// 3. Revalidação inteligente — staleTime, refetchOnWindowFocus
// 4. Mutations com invalidação — ao criar/editar, invalida a lista
// 5. Loading/error states consistentes
// ============================================================
import { QueryClient } from '@tanstack/react-query'

export function makeQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: {
        // 📌 PADRÃO: staleTime de 60s evita re-fetches agressivos
        staleTime: 60 * 1000,
        // 📌 PADRÃO: retry 1x em caso de erro de rede
        retry: 1,
        // 📌 PADRÃO: refetch quando a aba volta ao foco
        refetchOnWindowFocus: false,
      },
    },
  })
}
