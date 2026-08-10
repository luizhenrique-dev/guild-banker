'use client'
// ============================================================
// Custom Hooks: Fixed Expenses (TanStack Query)
// ============================================================
// 📌 PADRÃO: Cada hook encapsula uma operação da API.
// O componente só precisa chamar useFixedExpenses() para ter:
// - data (lista de despesas)
// - isLoading (estado de carregamento)
// - error (erro da request)
// Sem useEffect, sem estado local, sem race conditions.
// ============================================================
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { fixedExpensesApi } from './fixed-expenses.api'
import type { CreateFixedExpenseInput } from '../schemas/fixed-expense.schema'
import { toast } from 'sonner'
import i18n from '@/src/shared/lib/i18n'

// 📌 PADRÃO: Query Key Factory
// Centraliza as chaves do cache para evitar typos e facilitar invalidação.
export const fixedExpenseKeys = {
  all: ['fixed-expenses'] as const,
  lists: () => [...fixedExpenseKeys.all, 'list'] as const,
  list: () => [...fixedExpenseKeys.lists()] as const,
}

// -------------------------------------------------------
// QUERY: Listar despesas fixas ativas
// -------------------------------------------------------
export function useFixedExpenses() {
  return useQuery({
    queryKey: fixedExpenseKeys.list(),
    queryFn: () => fixedExpensesApi.list(),
    staleTime: 1000 * 60 * 5, // 5 min de cache fresco
  })
}

// -------------------------------------------------------
// MUTATION: Criar despesa fixa
// -------------------------------------------------------
export function useCreateFixedExpense() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: CreateFixedExpenseInput) => fixedExpensesApi.create(input),
    onSuccess: () => {
      // 📌 PADRÃO: Invalidar a lista após mutation força um refetch
      queryClient.invalidateQueries({ queryKey: fixedExpenseKeys.lists() })
      toast.success(i18n.t('fx.toast.created'))
    },
    onError: () => {
      toast.error(i18n.t('fx.toast.error'))
    },
  })
}

// -------------------------------------------------------
// MUTATION: Atualizar despesa fixa
// -------------------------------------------------------
export function useUpdateFixedExpense() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, input }: { id: number; input: CreateFixedExpenseInput }) =>
      fixedExpensesApi.update(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: fixedExpenseKeys.lists() })
      toast.success(i18n.t('fx.toast.updated'))
    },
    onError: () => {
      toast.error(i18n.t('fx.toast.error'))
    },
  })
}

// -------------------------------------------------------
// MUTATION: Desativar despesa fixa
// -------------------------------------------------------
export function useDeactivateFixedExpense() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, status }: { id: number; status: 'PAUSED' | 'CANCELLED' }) =>
      fixedExpensesApi.deactivate(id, status),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: fixedExpenseKeys.lists() })
      toast.success(i18n.t('fx.toast.deactivated'))
    },
    onError: () => {
      toast.error(i18n.t('fx.toast.error'))
    },
  })
}
