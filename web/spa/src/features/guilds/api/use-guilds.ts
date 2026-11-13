import { APIError } from '@/src/shared/lib/api-client'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { guildsApi } from './guilds.api'
import { useSession } from '@/src/app/session-provider'
import { toast } from 'sonner'
import i18n from '@/src/shared/lib/i18n'
import type { components } from '@/src/shared/api/schema'

export type Member = components['schemas']['MemberResponse']
export type SentInvitation = components['schemas']['SentInvitationResponse']
const memberKey = (userID?: string | number, guildID?: number) => ['guild-members', userID, guildID] as const
const invitationKey = (userID?: string | number, guildID?: number) => ['sent-invitations', userID, guildID] as const
export function useGuildMembers(guildID: number) {
  const session = useSession()
  return useQuery({
    queryKey: memberKey(session.user?.id, guildID),
    enabled: !!session.user,
    queryFn: ({ signal }) => session.request((options) => guildsApi.members(guildID, options), signal),
  })
}
export function useSentInvitations(guildID: number) {
  const session = useSession()
  return useQuery({
    queryKey: invitationKey(session.user?.id, guildID),
    enabled: !!session.user,
    queryFn: ({ signal }) => session.request((options) => guildsApi.sentInvitations(guildID, options), signal),
  })
}
export function useCancelInvitation() {
  const session = useSession()
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ guildId, id }: { guildId: number; id: number }) => session.request((options) => guildsApi.cancelInvitation(guildId, id, options)),
    onSuccess: (_result, { guildId, id }) => {
      if (!session.active()) return
      qc.setQueryData<SentInvitation[]>(invitationKey(session.user?.id, guildId), (invitations) => invitations?.filter((invitation) => invitation.id !== id))
      void qc.invalidateQueries({ queryKey: invitationKey(session.user?.id, guildId) })
      toast.success(i18n.t('guild.toast.cancelled'))
    },
    onError: (error, { guildId }) => {
      if (!session.active()) return
      if (error instanceof APIError && (error.status === 409 || error.status === 404)) {
        void qc.invalidateQueries({ queryKey: invitationKey(session.user?.id, guildId) })
        void qc.invalidateQueries({ queryKey: memberKey(session.user?.id, guildId) })
      }
    },
  })
}
export function useInviteMember() {
  const session = useSession()
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ guildId, email }: { guildId: number; email: string }) => session.request((options) => guildsApi.invite(guildId, email, options)),
    onSuccess: (_result, { guildId }) => {
      if (!session.active()) return
      void qc.invalidateQueries({ queryKey: invitationKey(session.user?.id, guildId) })
      toast.success(i18n.t('guild.toast.invited'))
    },
    onError: (error) => {
      if (!session.active()) return
      let message = 'fx.toast.error'
      if (error instanceof APIError) {
        if (error.status === 403) message = 'permission.denied'
        if (error.code === 'invitation_already_pending') message = 'guild.invite.alreadyPending'
        if (error.code === 'guild_member_already_set') message = 'guild.invite.alreadyMember'
      }
      toast.error(i18n.t(message))
    },
  })
}
export function useRemoveMember() {
  const session = useSession()
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ guildId, userId }: { guildId: number; userId: number }) => session.request((options) => guildsApi.removeMember(guildId, userId, options)),
    onSuccess: (_result, { guildId, userId }) => {
      if (!session.active()) return
      qc.setQueryData<Member[]>(memberKey(session.user?.id, guildId), (members) => members?.filter((member) => member.id !== userId))
      void qc.invalidateQueries({ queryKey: memberKey(session.user?.id, guildId) })
      toast.success(i18n.t('guild.toast.removed'))
    },
    onError: (error) => { if (session.active()) toast.error(i18n.t(error instanceof APIError && error.status === 403 ? 'permission.denied' : 'fx.toast.error')) },
  })
}
