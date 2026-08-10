'use client'
// ============================================================
// FixedExpensesPageClient — Client component da página principal
// Contém toda a interatividade (hooks, estado local, modais).
// ============================================================
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Plus, Info } from 'lucide-react'
import { useFixedExpenses, useDeactivateFixedExpense } from '../api/use-fixed-expenses'
import { FixedExpenseTable } from './fixed-expense-table'
import { FixedExpenseModal } from './fixed-expense-modal'
import { TableSkeleton } from '@/src/shared/components/loading-skeleton'
import { ErrorBoundary } from '@/src/shared/components/error-boundary'
import type { FixedExpenseResponse } from '../types'

export function FixedExpensesPageClient() {
  const { t } = useTranslation()
  const { data: expenses, isLoading, error } = useFixedExpenses()
  const deactivateMutation = useDeactivateFixedExpense()

  const [modalOpen, setModalOpen] = useState(false)
  const [editingExpense, setEditingExpense] = useState<FixedExpenseResponse | null>(null)

  const handleEdit = (expense: FixedExpenseResponse) => {
    setEditingExpense(expense)
    setModalOpen(true)
  }

  const handleDeactivate = (expense: FixedExpenseResponse) => {
    if (window.confirm(t('fx.deactivate.message', { name: expense.name }))) {
      deactivateMutation.mutate({ id: expense.id, status: 'PAUSED' })
    }
  }

  const handleCloseModal = () => {
    setModalOpen(false)
    setEditingExpense(null)
  }

  return (
    <ErrorBoundary>
      <main className="flex-1 px-8 py-8">
        {/* Header com título e botão */}
        <div className="mb-2 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <h1 className="text-[32px] font-normal tracking-[-0.8px]">
              {t('fx.title')}
            </h1>
            <span className="rounded-full bg-[#eef0f3] px-3 py-1 text-[12px] font-semibold text-[#5b616e]">
              {t('fx.badge')}
            </span>
          </div>
          <button
            type="button"
            onClick={() => {
              setEditingExpense(null)
              setModalOpen(true)
            }}
            className="flex items-center gap-2 rounded-full bg-[#0052ff] px-5 py-3 text-[14px] font-semibold text-white transition-colors hover:bg-[#003ecc]"
          >
            <Plus size={16} />
            {t('fx.newButton')}
          </button>
        </div>

        {/* Info banner */}
        <div className="mb-6 flex items-center gap-2 rounded-[12px] bg-[#eef0f3] px-4 py-3 text-[13px] text-[#5b616e]">
          <Info size={16} className="shrink-0 text-[#7c828a]" />
          {t('fx.info')}
        </div>

        {/* Tabela */}
        {isLoading ? (
          <TableSkeleton rows={3} />
        ) : error ? (
          <div className="rounded-[24px] border border-[#dee1e6] bg-white p-8 text-center text-[#cf202f]">
            {t('fx.toast.error')}
          </div>
        ) : (
          <FixedExpenseTable
            expenses={expenses ?? []}
            onEdit={handleEdit}
            onDeactivate={handleDeactivate}
          />
        )}

        {/* Modal de criação/edição */}
        <FixedExpenseModal
          open={modalOpen}
          onClose={handleCloseModal}
          editingExpense={editingExpense}
        />
      </main>
    </ErrorBoundary>
  )
}
