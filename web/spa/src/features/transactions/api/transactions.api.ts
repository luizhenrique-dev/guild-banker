import { safeID } from '@/src/shared/api/id'
import { apiClient, type RequestOptions } from '@/src/shared/lib/api-client'
import { legacyRequestOptions, getActiveGuildId } from '@/src/shared/lib/legacy-session'
import type { components } from '@/src/shared/api/schema'

export interface ListParams { cursor?: string; limit?: number; category?: string; type?: string }
export const transactionsApi = {
  list: (params: ListParams = {}, options: RequestOptions = legacyRequestOptions(), guildId = getActiveGuildId()) =>
    apiClient<components['schemas']['ListTransactionsResponse']>(`/api/v1/guilds/${safeID(guildId)}/transactions`, { ...options, params: { limit: 10, ...params } }),
  cancel: (id: number, options: RequestOptions = legacyRequestOptions(), guildId = getActiveGuildId()) =>
    apiClient(`/api/v1/guilds/${safeID(guildId)}/transactions/${safeID(id)}`, { ...options, method: 'DELETE' }),
  setVisibility: (id: number, visibility: 'PRIVATE' | 'PUBLIC', options: RequestOptions = legacyRequestOptions(), guildId = getActiveGuildId()) => {
    const body: components['schemas']['SetVisibilityRequest'] = { visibility }
    return apiClient<components['schemas']['SetVisibilityResponse']>(`/api/v1/guilds/${safeID(guildId)}/transactions/${safeID(id)}/visibility`, { ...options, method: 'PATCH', body })
  },
}
