import { useEffect, useRef, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useSession } from '@/src/app/session-provider'
import { APIError } from '@/src/shared/lib/api-client'
import { useActiveGuild } from '../use-active-guild'
import { guildsApi } from './guilds.api'
import type { components } from '@/src/shared/api/schema'

type Invitation = components['schemas']['InvitationResponse']
export function useOnboarding(started: boolean) {
  const session = useSession()
  const context = useActiveGuild()
  const qc = useQueryClient()
  const mounted = useRef(true)
  useEffect(() => { mounted.current = true; return () => { mounted.current = false } }, [])
  const [actionError, setActionError] = useState<string | null>(null)
  const key = ['guild-invites', session.user?.id] as const
  const alive = () => {
    const latestError = qc.getQueryState(key)?.error
    return mounted.current && session.active() && !(latestError instanceof APIError && latestError.status === 403)
  }
  const invitations = useQuery({
    queryKey: key,
    queryFn: ({ signal }) => session.request((options) => guildsApi.invitations(options), signal),
    enabled: started && !!session.user && !context.isLoading && !context.error && context.guilds.length === 0,
  })
  const report = (error: unknown) => {
    if (!alive()) return
    setActionError(error instanceof APIError && error.status === 403 ? 'permission.denied' : 'onboarding.actionError')
    if (error instanceof APIError && (error.status === 409 || error.status === 404)) {
      void qc.invalidateQueries({ queryKey: key })
      void qc.invalidateQueries({ queryKey: ['guilds', session.user?.id] })
    }
  }
  const create = useMutation({
    mutationFn: (input: components['schemas']['CreateGuildRequest']) => session.request((options) => guildsApi.create(input, options)),
    onMutate: () => { if (alive()) setActionError(null) },
    onSuccess: (guild) => { if (alive()) context.activateJoinedGuild(guild) },
    onError: report,
  })
  const respond = useMutation({
    mutationFn: ({ id, accept }: { id: number; accept: boolean }) => session.request((options) => guildsApi.respond(id, accept, options)),
    onMutate: () => { if (alive()) setActionError(null) },
    onSuccess: async (guild, { id, accept }) => {
      if (!alive()) return
      if (accept && guild) {
        qc.setQueryData<Invitation[]>(key, (previous) => previous?.filter((invite) => invite.id !== id) ?? [])
        context.activateJoinedGuild(guild)
      } else {
        await qc.invalidateQueries({ queryKey: key })
      }
    },
    onError: report,
  })
  return { context, invitations, create, respond, actionError }
}
