'use client'
// ============================================================
// FixedExpensesPageClient — Client component da página principal
// Contém toda a interatividade (hooks, estado local, modais).
// ============================================================
import { useRef, useState } from 'react'
import { Button } from '@/components/ui/button'
import { AlertDialog, AlertDialogContent, AlertDialogTitle, AlertDialogDescription, AlertDialogFooter, AlertDialogCancel, AlertDialogAction } from '@/components/ui/alert-dialog'
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
  const { data: expenses, isLoading, error, refetch, isFetching } = useFixedExpenses()
  const deactivateMutation = useDeactivateFixedExpense()

  const [modalOpen, setModalOpen] = useState(false)
  const [editingExpense, setEditingExpense] = useState<FixedExpenseResponse | null>(null)

  const [deactivating, setDeactivating] = useState<FixedExpenseResponse | null>(null)
  const [saved, setSaved] = useState(false)
  const returnFocus = useRef<HTMLElement | null>(null)
  const newButton = useRef<HTMLButtonElement | null>(null)

  const handleEdit = (expense: FixedExpenseResponse) => {
    setEditingExpense(expense)
    setModalOpen(true)
  }

  const handleDeactivate = (expense: FixedExpenseResponse) => {
    returnFocus.current = document.activeElement as HTMLElement
    deactivateMutation.reset()
    setDeactivating(expense)
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
            <span className="rounded-full bg-secondary px-3 py-1 text-[12px] font-semibold text-secondary-foreground">
              {t('fx.badge')}
            </span>
          </div>
          <Button
            ref={newButton}
            type="button"
            onClick={() => {
              setEditingExpense(null)
              setModalOpen(true)
            }}
            className="flex items-center gap-2 rounded-full bg-primary px-5 py-3 text-[14px] font-semibold text-white transition-colors hover:bg-primary/90"
          >
            <Plus size={16} />
            {t('fx.newButton')}
          </Button>
        </div>

        {/* Info banner */}
        <div className="mb-6 flex items-center gap-2 rounded-[12px] bg-secondary px-4 py-3 text-[13px] text-secondary-foreground">
          <Info size={16} className="shrink-0 text-muted-foreground" />
          {t('fx.info')}
        </div>

        {/* Tabela */}
        {isLoading ? (
          <div role="status" aria-label={t('common.loading')}><TableSkeleton rows={3} /></div>
        ) : error ? (
          <div role="alert" className="rounded-[24px] border border-border bg-card p-8 text-center text-destructive">
            <p>{t(saved ? 'fx.savedRefreshFailed' : 'fx.toast.error')}</p>
            <Button type="button" variant="outline" disabled={isFetching} onClick={() => void refetch()} className="mt-4">{t('common.retry')}</Button>
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
          onSaved={() => setSaved(true)}
          editingExpense={editingExpense}
        />
        <AlertDialog open={!!deactivating} onOpenChange={(open) => { if (!open && !deactivateMutation.isPending) setDeactivating(null) }}>
          <AlertDialogContent className="w-[calc(100%-2rem)] rounded-2xl" onCloseAutoFocus={(event) => {
            event.preventDefault()
            const target = returnFocus.current
            if (target?.isConnected) target.focus()
            else newButton.current?.focus()
          }}>
            <AlertDialogTitle>{t('fx.deactivate.title')}</AlertDialogTitle>
            <AlertDialogDescription>{t('fx.deactivate.message', { name: deactivating?.name })}</AlertDialogDescription>
            {deactivateMutation.error && <p role="alert" className="text-sm text-destructive">{t('fx.toast.error')}</p>}
            <AlertDialogFooter>
              <AlertDialogCancel disabled={deactivateMutation.isPending}>{t('fx.modal.cancel')}</AlertDialogCancel>
              <AlertDialogAction disabled={deactivateMutation.isPending} onClick={(event) => {
                event.preventDefault()
                if (!deactivating || deactivateMutation.isPending) return
                deactivateMutation.mutate({ id: deactivating.id, status: 'PAUSED' }, { onSuccess: () => { setSaved(true); setDeactivating(null) } })
              }}>{t('fx.deactivate.confirm')}</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </main>
    </ErrorBoundary>
  )
}
