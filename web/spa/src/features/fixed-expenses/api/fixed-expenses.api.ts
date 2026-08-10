// ============================================================
// API Layer: Fixed Expenses
// Funções puras que fazem as chamadas HTTP via apiClient.
// ============================================================
import { apiClient } from '@/src/shared/lib/api-client'
import type { FixedExpenseResponse } from '../types'
import type { CreateFixedExpenseInput } from '../schemas/fixed-expense.schema'

export const fixedExpensesApi = {
  // 💡 REAL API: GET /api/v1/fixed-expenses
  // Retorna apenas despesas ACTIVE do usuário (filtrado pelo X-User-ID no backend)
  list: async (): Promise<FixedExpenseResponse[]> => {
    const { data } = await apiClient.get<FixedExpenseResponse[]>('/api/v1/fixed-expenses')
    return data
  },

  // 💡 REAL API: POST /api/v1/fixed-expenses
  create: async (input: CreateFixedExpenseInput): Promise<FixedExpenseResponse> => {
    // 📌 PADRÃO: Normalizar amount de "2500,00" para "2500.00"
    const payload = {
      ...input,
      amount: input.amount.replace(',', '.'),
    }
    const { data } = await apiClient.post<FixedExpenseResponse>('/api/v1/fixed-expenses', payload)
    return data
  },

  // 💡 REAL API: PATCH /api/v1/fixed-expenses/:id
  update: async (id: number, input: CreateFixedExpenseInput): Promise<FixedExpenseResponse> => {
    const payload = {
      ...input,
      amount: input.amount.replace(',', '.'),
    }
    const { data } = await apiClient.patch<FixedExpenseResponse>(`/api/v1/fixed-expenses/${id}`, payload)
    return data
  },

  // 💡 REAL API: PATCH /api/v1/fixed-expenses/:id/deactivate
  deactivate: async (id: number, status: 'PAUSED' | 'CANCELLED' = 'PAUSED'): Promise<void> => {
    await apiClient.patch(`/api/v1/fixed-expenses/${id}/deactivate`, { status })
  },
}
