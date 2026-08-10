import { apiClient } from '@/src/shared/lib/api-client'
import type { Guild } from '@/src/shared/store/guild.store'

export const guildsApi = {
  list: async (): Promise<Guild[]> => {
    const { data } = await apiClient.get<Guild[]>('/api/v1/guilds')
    return data
  },

  invite: async (guildId: number, email: string): Promise<void> => {
    await apiClient.post(`/api/v1/guilds/${guildId}/invites`, { email })
  },

  removeMember: async (guildId: number, userId: number): Promise<void> => {
    await apiClient.delete(`/api/v1/guilds/${guildId}/members/${userId}`)
  },
}
