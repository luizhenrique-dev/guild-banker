import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { fixedExpensesApi } from './fixed-expenses.api'
import { useEffect, useRef } from 'react'
import { useActiveGuild } from '../../guilds/use-active-guild'
import { useSession } from '@/src/app/session-provider'
import type { CreateFixedExpenseInput, UpdateFixedExpenseInput } from '../schemas/fixed-expense.schema'
import type { RequestOptions } from '@/src/shared/lib/api-client'
import { toast } from 'sonner'
import i18n from '@/src/shared/lib/i18n'

export const fixedExpenseKeys = { list: (userID?: string | number, guildID?: number) => ['fixed-expenses', userID, guildID] as const }
export function useFixedExpenses() {
  const session = useSession()
  const { activeGuild } = useActiveGuild()
  return useQuery({
    queryKey: fixedExpenseKeys.list(session.user?.id, activeGuild?.id),
    queryFn: ({ signal }) => session.request((options) => fixedExpensesApi.list(options, activeGuild!.id), signal),
    enabled: !!session.user && !!activeGuild,
    staleTime: 5 * 60 * 1000,
  })
}

function useExpenseMutation<V extends { guildId: number }, R>(operation: (input: V, options: RequestOptions) => Promise<R>, message: string) {
  const session = useSession()
  const qc = useQueryClient()
  const userID = session.user?.id
  const mounted = useRef(true)
  useEffect(() => { mounted.current = true; return () => { mounted.current = false } }, [])
  return useMutation({
    mutationFn: (input: V) => session.request((options) => operation(input, options)),
    onSuccess: (_result, { guildId }) => {
      if (!session.active()) return
      void qc.invalidateQueries({ queryKey: fixedExpenseKeys.list(userID, guildId) })
      if (!mounted.current) return
      toast.success(i18n.t(message))
    },
  })
}
export function useCreateFixedExpense() {
  return useExpenseMutation(({ input, guildId }: { guildId: number; input: CreateFixedExpenseInput }, options) => fixedExpensesApi.create(input, options, guildId), 'fx.toast.created')
}
export function useUpdateFixedExpense() {
  return useExpenseMutation(({ id, input, guildId }: { guildId: number; id: number; input: UpdateFixedExpenseInput }, options) => fixedExpensesApi.update(id, input, options, guildId), 'fx.toast.updated')
}
export function useDeactivateFixedExpense() {
  return useExpenseMutation(({ id, status, guildId }: { guildId: number; id: number; status: 'PAUSED' | 'CANCELLED' }, options) => fixedExpensesApi.deactivate(id, status, options, guildId), 'fx.toast.deactivated')
}

export function useSetFixedExpenseStatus() {
  return useExpenseMutation(({ id, status, guildId }: { guildId: number; id: number; status: 'ACTIVE' | 'PAUSED' }, options) => fixedExpensesApi.setStatus(id, status, options, guildId), 'fx.toast.statusUpdated')
}

export function useDeleteFixedExpense() {
  return useExpenseMutation(({ id, guildId }: { guildId: number; id: number }, options) => fixedExpensesApi.delete(id, options, guildId), 'fx.toast.deleted')
}
