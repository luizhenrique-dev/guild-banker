import { safeID } from '@/src/shared/api/id'
import { apiClient, type RequestOptions } from '@/src/shared/lib/api-client'
import type { Guild } from '../types'
import type { components } from '@/src/shared/api/schema'

export const guildsApi = {
  members: (guildId: number, options: RequestOptions) => apiClient<components['schemas']['MemberResponse'][]>(`/api/v1/guilds/${safeID(guildId)}/members`, options),
  sentInvitations: (guildId: number, options: RequestOptions) => apiClient<components['schemas']['SentInvitationResponse'][]>(`/api/v1/guilds/${safeID(guildId)}/invites`, options),
  cancelInvitation: (guildId: number, id: number, options: RequestOptions) => apiClient(`/api/v1/guilds/${safeID(guildId)}/invites/${safeID(id)}`, { ...options, method: 'DELETE' }),
  create: (body: components['schemas']['CreateGuildRequest'], options: RequestOptions) => apiClient<Guild>('/api/v1/guilds', { ...options, method: 'POST', body }),
  invitations: (options: RequestOptions) => apiClient<components['schemas']['InvitationResponse'][]>('/api/v1/guild-invites', options),
  respond: (id: number, accept: boolean, options: RequestOptions) => apiClient<Guild | undefined>(`/api/v1/guild-invites/${safeID(id)}/${accept ? 'accept' : 'decline'}`, { ...options, method: 'POST' }),
  list: (options: RequestOptions) => apiClient<Guild[]>('/api/v1/guilds', options),
  invite: (guildId: number, email: string, options: RequestOptions) => {
    const body: components['schemas']['InviteUserRequest'] = { email }
    return apiClient(`/api/v1/guilds/${safeID(guildId)}/invites`, { ...options, method: 'POST', body })
  },
  removeMember: (guildId: number, userId: number, options: RequestOptions) =>
    apiClient(`/api/v1/guilds/${safeID(guildId)}/members/${safeID(userId)}`, { ...options, method: 'DELETE' }),
}
