import { createContext, useEffect, useMemo, type ReactNode } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useSearchParams } from 'react-router-dom'
import { useSession } from '@/src/app/session-provider'
import { APIError } from '@/src/shared/lib/api-client'
import { guildsApi } from './api/guilds.api'
import type { Guild } from './types'

export const ActiveGuildContext = createContext<ReturnType<typeof useGuildResolution> | null>(null)

function validID(value: string | null) {
  if (!value || !/^[1-9]\d*$/.test(value)) return undefined
  const id = Number(value)
  return Number.isSafeInteger(id) ? id : undefined
}
function preference(key: string) {
  try { return validID(localStorage.getItem(key)) }
  catch { return undefined }
}
function useGuildResolution() {
  const session = useSession()
  const qc = useQueryClient()
  const [params, setParams] = useSearchParams()
  const queryKey = ['guilds', session.user?.id] as const
  const storageKey = `guildbanker:last-guild:${session.demo ? 'demo' : 'oidc'}:${session.user?.id}`
  const query = useQuery({
    queryKey,
    queryFn: ({ signal }) => session.request((options) => guildsApi.list(options), signal),
    enabled: !!session.user && session.phase === 'ready',
  })
  const isForbidden = query.error instanceof APIError && query.error.status === 403
  const guilds = useMemo(() => isForbidden ? [] : query.data ?? [], [query.data, isForbidden])
  const selected = params.get('guild')
  const preferred = guilds.find((guild) => guild.id === preference(storageKey))
  const automatic = preferred ?? (guilds.length === 1 ? guilds[0] : undefined)
  const activeGuild = selected === null ? automatic : guilds.find((guild) => guild.id === validID(selected))
  useEffect(() => {
    if (!activeGuild || !session.active()) return
    try { localStorage.setItem(storageKey, String(activeGuild.id)) } catch { /* Preference is best effort. */ }
    if (selected === null) {
      const next = new URLSearchParams(params)
      next.set('guild', String(activeGuild.id))
      setParams(next, { replace: true })
    }
  }, [activeGuild, selected, params, setParams, storageKey, session])
  const setActiveGuild = (guild: { id: number }) => {
    if (!session.active() || !guilds.some((candidate) => candidate.id === guild.id)) return
    const next = new URLSearchParams(params)
    next.set('guild', String(guild.id))
    setParams(next)
  }
  const activateJoinedGuild = (guild: Guild) => {
    if (!session.active() || !Number.isSafeInteger(guild.id) || guild.id <= 0) return
    const latestError = qc.getQueryState(queryKey)?.error
    if (latestError instanceof APIError && latestError.status === 403) return
    qc.setQueryData<Guild[]>(queryKey, (previous = []) => [...previous.filter((candidate) => candidate.id !== guild.id), guild])
    const next = new URLSearchParams(params)
    next.set('guild', String(guild.id))
    setParams(next, { replace: true })
  }
  return { activeGuild, guilds, isForbidden, setActiveGuild, activateJoinedGuild, isLoading: query.isLoading, error: query.error, refetch: query.refetch }
}
export function ActiveGuildProvider({ children }: { children: ReactNode }) {
  const value = useGuildResolution()
  return <ActiveGuildContext.Provider value={value}>{children}</ActiveGuildContext.Provider>
}
