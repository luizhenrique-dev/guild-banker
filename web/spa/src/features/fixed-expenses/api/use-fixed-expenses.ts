import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { fixedExpensesApi } from './fixed-expenses.api'
import { useSession } from '@/src/app/session-provider'
import type { CreateFixedExpenseInput } from '../schemas/fixed-expense.schema'
import type { RequestOptions } from '@/src/shared/lib/api-client'
import { toast } from 'sonner'
import i18n from '@/src/shared/lib/i18n'

export const fixedExpenseKeys = { list: (userID?: number) => ['fixed-expenses', userID] as const }
export function useFixedExpenses() {
  const session = useSession()
  return useQuery({
    queryKey: fixedExpenseKeys.list(session.user?.id),
    queryFn: ({ signal }) => session.request((options) => fixedExpensesApi.list(options), signal),
    enabled: !!session.user,
    staleTime: 5 * 60 * 1000,
  })
}

function useExpenseMutation<V, R>(operation: (input: V, options: RequestOptions) => Promise<R>, message: string) {
  const session = useSession()
  const qc = useQueryClient()
  const userID = session.user?.id
  return useMutation({
    mutationFn: (input: V) => session.request((options) => operation(input, options)),
    onSuccess: () => {
      if (!session.active()) return
      void qc.invalidateQueries({ queryKey: fixedExpenseKeys.list(userID) })
      toast.success(i18n.t(message))
    },
    onError: () => { if (session.active()) toast.error(i18n.t('fx.toast.error')) },
  })
}
export function useCreateFixedExpense() {
  return useExpenseMutation((input: CreateFixedExpenseInput, options) => fixedExpensesApi.create(input, options), 'fx.toast.created')
}
export function useUpdateFixedExpense() {
  return useExpenseMutation(({ id, input }: { id: number; input: CreateFixedExpenseInput }, options) => fixedExpensesApi.update(id, input, options), 'fx.toast.updated')
}
export function useDeactivateFixedExpense() {
  return useExpenseMutation(({ id, status }: { id: number; status: 'PAUSED' | 'CANCELLED' }, options) => fixedExpensesApi.deactivate(id, status, options), 'fx.toast.deactivated')
}
