// ============================================================
// 📌 PADRÃO: Zustand Store — Estado do usuário autenticado
// ============================================================
// Por que Zustand?
// 1. API minimalista — sem boilerplate de reducers/actions
// 2. Funciona fora de componentes React (interceptors Axios)
// 3. Suporte nativo a selectors com shallow comparison
// 4. Tipagem forte com TypeScript
//
// 💡 REAL API: Em produção, o token JWT viria do Keycloak.
// O store armazenaria o access_token e refresh_token.
// ============================================================
import { create } from 'zustand'

export interface AuthUser {
  id: number
  email: string
  name: string
  token?: string // JWT Bearer token (futuro)
}

interface AuthState {
  user: AuthUser | null
  setUser: (user: AuthUser) => void
  clearUser: () => void
}

// ⚠️ ATENÇÃO: Usuário mock é inicializado aqui para desenvolvimento.
// Em produção, o user seria null até o login via Keycloak.
export const useAuthStore = create<AuthState>((set) => ({
  user: {
    id: 1,
    email: 'ana@familia.com',
    name: 'Ana Silva',
  },
  setUser: (user: AuthUser) => set({ user }),
  clearUser: () => set({ user: null }),
}))
