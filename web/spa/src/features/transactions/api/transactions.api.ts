import { safeID } from '@/src/shared/api/id'
import { apiClient, type RequestOptions } from '@/src/shared/lib/api-client'
import type { components } from '@/src/shared/api/schema'

export interface ListParams { cursor?: string; limit?: number; category?: string; type?: string }
export const transactionsApi = {
  list: (params: ListParams, options: RequestOptions, guildId: number) =>
    apiClient<components['schemas']['ListTransactionsResponse']>(`/api/v1/guilds/${safeID(guildId)}/transactions`, { ...options, params: { limit: 10, ...params } }),
  cancel: (id: number, options: RequestOptions, guildId: number) =>
    apiClient(`/api/v1/guilds/${safeID(guildId)}/transactions/${safeID(id)}`, { ...options, method: 'DELETE' }),
  setVisibility: (id: number, visibility: 'PRIVATE' | 'PUBLIC', options: RequestOptions, guildId: number) => {
    const body: components['schemas']['SetVisibilityRequest'] = { visibility }
    return apiClient<components['schemas']['SetVisibilityResponse']>(`/api/v1/guilds/${safeID(guildId)}/transactions/${safeID(id)}/visibility`, { ...options, method: 'PATCH', body })
  },
}
