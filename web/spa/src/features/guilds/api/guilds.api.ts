import { safeID } from '@/src/shared/api/id'
import { apiClient, type RequestOptions } from '@/src/shared/lib/api-client'
import type { Guild } from '../types'
import type { components } from '@/src/shared/api/schema'

export const guildsApi = {
  list: (options: RequestOptions) => apiClient<Guild[]>('/api/v1/guilds', options),
  invite: (guildId: number, email: string, options: RequestOptions) => {
    const body: components['schemas']['InviteUserRequest'] = { email }
    return apiClient(`/api/v1/guilds/${safeID(guildId)}/invites`, { ...options, method: 'POST', body })
  },
  removeMember: (guildId: number, userId: number, options: RequestOptions) =>
    apiClient(`/api/v1/guilds/${safeID(guildId)}/members/${safeID(userId)}`, { ...options, method: 'DELETE' }),
}
