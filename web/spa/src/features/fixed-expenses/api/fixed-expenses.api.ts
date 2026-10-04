import { safeID } from '@/src/shared/api/id'
import { apiClient, type RequestOptions } from '@/src/shared/lib/api-client'
import { legacyRequestOptions } from '@/src/shared/lib/legacy-session'
import type { components } from '@/src/shared/api/schema'
import type { FixedExpenseResponse } from '../types'
import type { CreateFixedExpenseInput } from '../schemas/fixed-expense.schema'

export const fixedExpensesApi = {
  list: (options: RequestOptions = legacyRequestOptions()) => apiClient<FixedExpenseResponse[]>('/api/v1/fixed-expenses', options),
  create: (input: CreateFixedExpenseInput, options: RequestOptions = legacyRequestOptions()) => {
    const body: components['schemas']['CreateFixedExpenseRequest'] = { ...input, category: input.category as FixedExpenseResponse['category'], amount: input.amount.replace(',', '.') }
    return apiClient<FixedExpenseResponse>('/api/v1/fixed-expenses', { ...options, method: 'POST', body })
  },
  update: (id: number, input: CreateFixedExpenseInput, options: RequestOptions = legacyRequestOptions()) => {
    const body: components['schemas']['UpdateFixedExpenseRequest'] = { amount: input.amount.replace(',', '.'), due_day: input.due_day, category: input.category as FixedExpenseResponse['category'] }
    return apiClient<FixedExpenseResponse>(`/api/v1/fixed-expenses/${safeID(id)}`, { ...options, method: 'PATCH', body })
  },
  deactivate: (id: number, status: 'PAUSED' | 'CANCELLED' = 'PAUSED', options: RequestOptions = legacyRequestOptions()) => {
    const body: components['schemas']['DeactivateFixedExpenseRequest'] = { status }
    return apiClient(`/api/v1/fixed-expenses/${safeID(id)}/deactivate`, { ...options, method: 'PATCH', body })
  },
}
