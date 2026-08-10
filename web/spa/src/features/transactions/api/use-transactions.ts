'use client'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { transactionsApi } from './transactions.api'
import { toast } from 'sonner'
import i18n from '@/src/shared/lib/i18n'

export const transactionKeys = {
  all: ['transactions'] as const,
  lists: () => [...transactionKeys.all, 'list'] as const,
  list: (filters?: Record<string, unknown>) => [...transactionKeys.lists(), filters] as const,
}

export function useTransactions(filters: { cursor?: string; category?: string; type?: string } = {}) {
  return useQuery({
    queryKey: transactionKeys.list(filters),
    queryFn: () => transactionsApi.list(filters),
    staleTime: 1000 * 60 * 2,
  })
}

export function useCancelTransaction() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => transactionsApi.cancel(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: transactionKeys.lists() })
      toast.success(i18n.t('tx.toast.cancelled'))
    },
    onError: () => toast.error(i18n.t('fx.toast.error')),
  })
}

export function useToggleVisibility() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, visibility }: { id: number; visibility: 'PRIVATE' | 'PUBLIC' }) =>
      transactionsApi.setVisibility(id, visibility),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: transactionKeys.lists() })
      toast.success(i18n.t('tx.toast.visibilityChanged'))
    },
    onError: () => toast.error(i18n.t('fx.toast.error')),
  })
}
