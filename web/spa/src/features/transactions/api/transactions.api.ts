import { apiClient, getActiveGuildId } from '@/src/shared/lib/api-client'
import type { TransactionResponse } from '../types'

interface ListParams {
  cursor?: string
  limit?: number
  category?: string
  type?: string
}

export const transactionsApi = {
  list: async (params: ListParams = {}) => {
    const guildId = getActiveGuildId()
    const { data } = await apiClient.get<{ items: TransactionResponse[]; nextCursor?: string }>(
      `/api/v1/guilds/${guildId}/transactions`,
      { params: { limit: 10, ...params } }
    )
    return data
  },

  cancel: async (id: number) => {
    const guildId = getActiveGuildId()
    await apiClient.delete(`/api/v1/guilds/${guildId}/transactions/${id}`)
  },

  setVisibility: async (id: number, visibility: 'PRIVATE' | 'PUBLIC') => {
    const guildId = getActiveGuildId()
    const { data } = await apiClient.patch(
      `/api/v1/guilds/${guildId}/transactions/${id}/visibility`,
      { visibility }
    )
    return data
  },
}
