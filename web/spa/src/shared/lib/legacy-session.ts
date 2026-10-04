import { useAuthStore } from '../store/auth.store'
import { useGuildStore } from '../store/guild.store'
import { safeID } from '../api/id'
import type { RequestOptions } from './api-client'

// Temporary bridge until T3 moves session ownership to the provider.
export function legacyRequestOptions(): RequestOptions {
  const { user } = useAuthStore.getState()
  if (!user) throw new Error('Session required')
  return { headers: { 'X-User-ID': String(safeID(user.id)), 'X-User-Email': user.email, ...(user.token ? { Authorization: `Bearer ${user.token}` } : {}) } }
}

export function getActiveGuildId(): number {
  const { activeGuild } = useGuildStore.getState()
  if (!activeGuild || !Number.isSafeInteger(activeGuild.id) || activeGuild.id <= 0) throw new Error('Guild required')
  return activeGuild.id
}
