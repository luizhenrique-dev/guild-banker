import { useEffect, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router-dom'
import { useSession } from '../../app/session-provider'
import { guildsApi } from './api/guilds.api'

export function useActiveGuild() {
  const session = useSession()
  const [params, setParams] = useSearchParams()
  const query = useQuery({
    queryKey: ['guilds', session.user?.id],
    queryFn: ({ signal }) => session.request((options) => guildsApi.list(options), signal),
    enabled: !!session.user,
  })
  const guilds = useMemo(() => query.data ?? [], [query.data])
  const selected = params.get('guild')
  const selectedID = selected && /^[1-9]\d*$/.test(selected) ? Number(selected) : undefined
  const activeGuild = Number.isSafeInteger(selectedID) ? guilds.find((guild) => guild.id === selectedID) : undefined
  useEffect(() => {
    const first = guilds[0]
    if (selected === null && first && Number.isSafeInteger(first.id) && first.id > 0) {
      const next = new URLSearchParams(params)
      next.set('guild', String(first.id))
      setParams(next, { replace: true })
    }
  }, [selected, guilds, params, setParams])
  const setActiveGuild = (guild: { id: number }) => {
    const next = new URLSearchParams(params)
    next.set('guild', String(guild.id))
    setParams(next)
  }
  return { activeGuild, guilds, setActiveGuild, isLoading: query.isLoading, error: query.error, refetch: query.refetch }
}
