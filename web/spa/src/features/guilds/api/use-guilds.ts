import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { guildsApi } from './guilds.api'
import { useSession } from '@/src/app/session-provider'
import { toast } from 'sonner'
import i18n from '@/src/shared/lib/i18n'

export interface DemoMember { id: number; name: string; email: string; role: string }
const memberKey = (userID?: number, guildID?: number) => ['demo-members', userID, guildID] as const
export function useDemoMembers(guildID: number) {
  const session = useSession()
  return useQuery({
    queryKey: memberKey(session.user?.id, guildID),
    enabled: import.meta.env.DEV && session.demo,
    queryFn: async () => {
      if (!import.meta.env.DEV || !session.demo) return []
      const { membersByGuild } = await import('@/src/mocks/fixtures/members')
      return structuredClone(membersByGuild[guildID] ?? [])
    },
  })
}
export function useInviteMember() {
  const session = useSession()
  return useMutation({
    mutationFn: ({ guildId, email }: { guildId: number; email: string }) => session.request((options) => guildsApi.invite(guildId, email, options)),
    onSuccess: () => { if (session.active()) toast.success(i18n.t('guild.toast.invited')) },
    onError: () => { if (session.active()) toast.error(i18n.t('fx.toast.error')) },
  })
}
export function useRemoveMember() {
  const session = useSession()
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ guildId, userId }: { guildId: number; userId: number }) => session.request((options) => guildsApi.removeMember(guildId, userId, options)),
    onSuccess: (_result, { guildId, userId }) => {
      if (!session.active()) return
      qc.setQueryData<DemoMember[]>(memberKey(session.user?.id, guildId), (members) => members?.filter((member) => member.id !== userId))
      toast.success(i18n.t('guild.toast.removed'))
    },
    onError: () => { if (session.active()) toast.error(i18n.t('fx.toast.error')) },
  })
}
