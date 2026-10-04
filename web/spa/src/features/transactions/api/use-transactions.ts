import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { transactionsApi, type ListParams } from './transactions.api'
import { useSession } from '@/src/app/session-provider'
import { useActiveGuild } from '../../guilds/use-active-guild'
import { toast } from 'sonner'
import i18n from '@/src/shared/lib/i18n'

export const transactionKeys = {
  lists: (userID?: number, guildID?: number) => ['transactions', userID, guildID] as const,
  list: (userID?: number, guildID?: number, filters: ListParams = {}) => ['transactions', userID, guildID, filters] as const,
}
export function useTransactions(filters: ListParams = {}) {
  const session = useSession()
  const { activeGuild } = useActiveGuild()
  return useQuery({
    queryKey: transactionKeys.list(session.user?.id, activeGuild?.id, filters),
    queryFn: ({ signal }) => session.request((options) => transactionsApi.list(filters, options, activeGuild!.id), signal),
    enabled: !!session.user && !!activeGuild,
    staleTime: 2 * 60 * 1000,
  })
}
export function useCancelTransaction() {
  const session = useSession()
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, guildId }: { id: number; guildId: number }) => session.request((options) => transactionsApi.cancel(id, options, guildId)),
    onSuccess: (_result, { guildId }) => {
      if (!session.active()) return
      void qc.invalidateQueries({ queryKey: transactionKeys.lists(session.user?.id, guildId) })
      toast.success(i18n.t('tx.toast.cancelled'))
    },
    onError: () => { if (session.active()) toast.error(i18n.t('fx.toast.error')) },
  })
}
export function useToggleVisibility() {
  const session = useSession()
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, guildId, visibility }: { id: number; guildId: number; visibility: 'PRIVATE' | 'PUBLIC' }) => session.request((options) => transactionsApi.setVisibility(id, visibility, options, guildId)),
    onSuccess: (_result, { guildId }) => {
      if (!session.active()) return
      void qc.invalidateQueries({ queryKey: transactionKeys.lists(session.user?.id, guildId) })
      toast.success(i18n.t('tx.toast.visibilityChanged'))
    },
    onError: () => { if (session.active()) toast.error(i18n.t('fx.toast.error')) },
  })
}
