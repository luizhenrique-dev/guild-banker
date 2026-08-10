// ============================================================
// API CLIENT — GuildBanker
// ============================================================
// MODO DESENVOLVIMENTO: MSW intercepta todas as chamadas abaixo.
// Para conectar na API REAL, siga os passos:
//
// 1. Defina no .env.local:
//    NEXT_PUBLIC_API_URL=http://localhost:8080
//    NEXT_PUBLIC_USE_MOCK=false
//
// 2. Autenticação:
//    A API usa JWT Bearer via Keycloak. Após login, armazene
//    o access_token no Zustand/localStorage e injete no header:
//    Authorization: Bearer <token>
//
//    ATENÇÃO: A validação JWT ainda não está implementada no
//    backend (retorna 501). Enquanto isso, passe obrigatoriamente:
//    X-User-ID: <id numérico do usuário>
//    X-User-Email: <email do usuário>
//    Esses headers são a única forma de autenticação funcional no momento.
//
// 3. Remova ou desabilite o MSW (NEXT_PUBLIC_USE_MOCK=false)
//    e todos os endpoints funcionarão contra a API real.
// ============================================================
import axios from 'axios'
import type { AxiosError, InternalAxiosRequestConfig } from 'axios'
import { useAuthStore } from '@/src/shared/store/auth.store'
import { useGuildStore } from '@/src/shared/store/guild.store'
import { toast } from 'sonner'

// 💡 REAL API: Em produção, troque para a URL do seu backend.
const baseURL = process.env.NEXT_PUBLIC_API_URL ?? ''

export const apiClient = axios.create({
  baseURL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 15_000,
})

// ============================================================
// REQUEST INTERCEPTOR
// Injeta automaticamente os headers de autenticação em toda request.
// ============================================================
apiClient.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const { user } = useAuthStore.getState()

  if (user) {
    // ⚠️ ATENÇÃO: X-User-ID e X-User-Email são a forma temporária de auth.
    // Quando o JWT estiver funcional no backend, remova esses headers
    // e use apenas o Authorization: Bearer <token>.
    config.headers.set('X-User-ID', String(user.id))
    config.headers.set('X-User-Email', user.email)

    if (user.token) {
      config.headers.set('Authorization', `Bearer ${user.token}`)
    }
  }

  // 📌 PADRÃO: Log de requests em dev para debug
  if (process.env.NODE_ENV === 'development') {
    console.log(`[API] ${config.method?.toUpperCase()} ${config.url}`)
  }

  return config
})

// ============================================================
// RESPONSE INTERCEPTOR
// Tratamento global de erros HTTP.
// ============================================================
apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError<{ error?: string }>) => {
    const status = error?.response?.status
    const message = error?.response?.data?.error ?? 'Erro inesperado'

    switch (status) {
      case 401:
        // 💡 REAL API: Limpar sessão e redirecionar para login
        useAuthStore.getState().clearUser()
        toast.error('Sessão expirada. Faça login novamente.')
        break
      case 403:
        toast.error('Você não tem permissão para esta ação.')
        break
      case 422:
        // Erro de regra de negócio
        toast.error(message)
        break
      case 500:
        toast.error('Erro interno do servidor. Tente novamente mais tarde.')
        break
      default:
        // 400, 404, 409 etc. — tratados individualmente pelos hooks de mutation
        break
    }

    return Promise.reject(error)
  }
)

// ============================================================
// HELPER: Obter o guildID ativo para rotas que dependem dele
// ============================================================
export function getActiveGuildId(): number {
  const { activeGuild } = useGuildStore.getState()
  return activeGuild?.id ?? 1
}
