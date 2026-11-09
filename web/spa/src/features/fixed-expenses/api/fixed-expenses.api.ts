import { safeID } from '@/src/shared/api/id'
import { apiClient, type RequestOptions } from '@/src/shared/lib/api-client'
import type { components } from '@/src/shared/api/schema'
import type { FixedExpenseResponse } from '../types'
import type { CreateFixedExpenseInput, UpdateFixedExpenseInput } from '../schemas/fixed-expense.schema'

export const fixedExpensesApi = {
  list: (options: RequestOptions, guildId: number) => apiClient<FixedExpenseResponse[]>(`/api/v1/guilds/${safeID(guildId)}/fixed-expenses`, options),
  create: (input: CreateFixedExpenseInput, options: RequestOptions, guildId: number) => {
    const body: components['schemas']['CreateFixedExpenseRequest'] = { ...input, category: input.category, amount: input.amount }
    return apiClient<FixedExpenseResponse>(`/api/v1/guilds/${safeID(guildId)}/fixed-expenses`, { ...options, method: 'POST', body })
  },
  update: (id: number, input: UpdateFixedExpenseInput, options: RequestOptions, guildId: number) => {
    const body: components['schemas']['UpdateFixedExpenseRequest'] = { amount: input.amount, due_day: input.due_day, category: input.category }
    return apiClient<FixedExpenseResponse>(`/api/v1/guilds/${safeID(guildId)}/fixed-expenses/${safeID(id)}`, { ...options, method: 'PATCH', body })
  },
  deactivate: (id: number, status: 'PAUSED' | 'CANCELLED', options: RequestOptions, guildId: number) => {
    const body: components['schemas']['DeactivateFixedExpenseRequest'] = { status }
    return apiClient(`/api/v1/guilds/${safeID(guildId)}/fixed-expenses/${safeID(id)}/deactivate`, { ...options, method: 'PATCH', body })
  },
}
