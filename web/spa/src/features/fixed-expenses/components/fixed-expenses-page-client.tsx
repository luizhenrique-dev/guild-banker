'use client'
import { useGuildMembers } from '../../guilds/api/use-guilds'
import { useActiveGuild } from '../../guilds/use-active-guild'
import { AccessDeniedPage } from '@/src/shared/components/access-denied-page'
import { APIError } from '@/src/shared/lib/api-client'
// ============================================================
// FixedExpensesPageClient — Client component da página principal
// Contém toda a interatividade (hooks, estado local, modais).
// ============================================================
import { useRef, useState } from 'react'
import { Button } from '@/components/ui/button'
import { AlertDialog, AlertDialogContent, AlertDialogTitle, AlertDialogDescription, AlertDialogFooter, AlertDialogCancel, AlertDialogAction } from '@/components/ui/alert-dialog'
import { useTranslation } from 'react-i18next'
import { Plus, Info } from 'lucide-react'
import { useFixedExpenses, useDeleteFixedExpense } from '../api/use-fixed-expenses'
import { FixedExpenseTable } from './fixed-expense-table'
import { FixedExpenseModal } from './fixed-expense-modal'
import { TableSkeleton } from '@/src/shared/components/loading-skeleton'
import { ErrorBoundary } from '@/src/shared/components/error-boundary'
import { CATEGORY_ICONS, FIXED_EXPENSE_CATEGORIES, type FixedExpenseCategory, type FixedExpenseResponse } from '../types'
import { monthlyTotal } from '../monthly-total'

export function FixedExpensesPageClient() {
  const { t, i18n } = useTranslation()
  const { activeGuild } = useActiveGuild()
  const { data: expenses, isLoading, error, refetch, isFetching } = useFixedExpenses()
  const members = useGuildMembers(activeGuild!.id)
  const deleteMutation = useDeleteFixedExpense()
  const [category, setCategory] = useState<FixedExpenseCategory | ''>('')
  const [member, setMember] = useState('')
  const filteredExpenses = (expenses ?? []).filter((expense) =>
    (!category || expense.category === category) &&
    (!member || (member === 'unassigned' ? expense.responsible_user_id == null : expense.responsible_user_id === Number(member))))

  const [modalOpen, setModalOpen] = useState(false)
  const [editingExpense, setEditingExpense] = useState<FixedExpenseResponse | null>(null)

  const [deleting, setDeleting] = useState<FixedExpenseResponse | null>(null)
  const [saved, setSaved] = useState(false)
  const returnFocus = useRef<HTMLElement | null>(null)
  const newButton = useRef<HTMLButtonElement | null>(null)

  const handleEdit = (expense: FixedExpenseResponse) => {
    setEditingExpense(expense)
    setModalOpen(true)
  }

  const handleDelete = (expense: FixedExpenseResponse) => {
    returnFocus.current = document.activeElement as HTMLElement
    deleteMutation.reset()
    setDeleting(expense)
  }

  const handleCloseModal = () => {
    setModalOpen(false)
    setEditingExpense(null)
  }

  if ([error, members.error].some((failure) => failure instanceof APIError && failure.status === 403)) return <AccessDeniedPage />

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

        <section aria-label={t('fx.summary')} className="mb-6 flex flex-wrap items-end justify-between gap-5 rounded-3xl border border-border bg-card p-6">
          <div>
            <h2 className="text-sm text-secondary-foreground">{t('fx.expectedMonth', { month: new Intl.DateTimeFormat(i18n.language, { month: 'long', year: 'numeric' }).format(new Date()) })}</h2>
            <p role="status" aria-label={t('fx.total')} className="mt-2 font-mono text-3xl font-semibold">{isLoading || error ? '—' : monthlyTotal(filteredExpenses, i18n.language)}</p>
            <p className="mt-2 text-sm text-muted-foreground">{t('fx.totalHint')}</p>
          </div>
          <div className="flex w-full flex-col gap-4 lg:w-auto lg:flex-row">
            <div className="min-w-0 lg:max-w-sm">
              <label htmlFor="fixed-expense-category-filter" className="mb-2 block text-sm font-medium">{t('fx.filterCategory')}</label>
              <select id="fixed-expense-category-filter" value={category} onChange={(event) => setCategory(event.target.value as FixedExpenseCategory | '')} className="w-full rounded-xl border border-input bg-background px-4 py-2.5 text-base">
                <option value="">{t('fx.allCategories')}</option>
                {FIXED_EXPENSE_CATEGORIES.map((value) => <option key={value} value={value} aria-label={t(`category.${value}`)}>{CATEGORY_ICONS[value]} {t(`category.${value}`)}</option>)}
              </select>
            </div>
            <div className="min-w-0 lg:max-w-sm">
              <label htmlFor="fixed-expense-member-filter" className="mb-2 block text-sm font-medium">{t('fx.filterMember')}</label>
              <select id="fixed-expense-member-filter" value={member} onChange={(event) => setMember(event.target.value)} disabled={members.isPending || !!members.error} className="w-full rounded-xl border border-input bg-background px-4 py-2.5 text-base">
                <option value="">{t('fx.allMembers')}</option>
                <option value="unassigned">{t('fx.unassigned')}</option>
                {member && member !== 'unassigned' && !members.data?.some((value) => value.id === Number(member)) && <option value={member}>{t('fx.responsibleUnavailable')}</option>}
                {members.data?.map((value) => <option key={value.id} value={value.id}>{value.name}</option>)}
              </select>
              {members.isPending && <p role="status">{t('fx.membersLoading')}</p>}
              {members.error && <div role="alert"><p>{t('guild.membersError')}</p><button type="button" onClick={() => void members.refetch()}>{t('common.retry')}</button></div>}
            </div>
          </div>
        </section>

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
            expenses={filteredExpenses}
            filtered={!!category || !!member}
            memberFiltered={!!member}
            members={members.error ? [] : members.data ?? []}
            onEdit={handleEdit}
            onDelete={handleDelete}
          />
        )}

        {/* Modal de criação/edição */}
        <FixedExpenseModal
          open={modalOpen}
          onClose={handleCloseModal}
          onSaved={() => setSaved(true)}
          editingExpense={editingExpense}
        />
        <AlertDialog open={!!deleting} onOpenChange={(open) => { if (!open && !deleteMutation.isPending) setDeleting(null) }}>
          <AlertDialogContent className="w-[calc(100%-2rem)] rounded-2xl" onCloseAutoFocus={(event) => {
            event.preventDefault()
            const target = returnFocus.current
            if (deleteMutation.isSuccess) newButton.current?.focus()
            else if (target?.isConnected) target.focus()
            else newButton.current?.focus()
          }}>
            <AlertDialogTitle>{t('fx.delete.title')}</AlertDialogTitle>
            <AlertDialogDescription>{t('fx.delete.message', { name: deleting?.name })}</AlertDialogDescription>
            {deleteMutation.error && <p role="alert" className="text-sm text-destructive">{t(deleteMutation.error instanceof APIError && deleteMutation.error.status === 403 ? 'permission.denied' : 'fx.toast.error')}</p>}
            <AlertDialogFooter>
              <AlertDialogCancel disabled={deleteMutation.isPending}>{t('fx.modal.cancel')}</AlertDialogCancel>
              <AlertDialogAction disabled={deleteMutation.isPending} onClick={(event) => {
                event.preventDefault()
                if (!activeGuild || !deleting || deleteMutation.isPending) return
                deleteMutation.mutate({ guildId: activeGuild.id, id: deleting.id }, { onSuccess: () => { setSaved(true); setDeleting(null) } })
              }}>{t('fx.delete.confirm')}</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </main>
    </ErrorBoundary>
  )
}
