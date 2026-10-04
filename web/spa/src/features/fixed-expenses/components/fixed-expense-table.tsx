'use client'
// ============================================================
// FixedExpenseTable — Tabela principal de despesas fixas
// Replica o design da seção fx-table do protótipo.
// ============================================================
import { useTranslation } from 'react-i18next'
import { FixedExpenseRow } from './fixed-expense-row'
import type { FixedExpenseResponse } from '../types'
import { Table } from '@/components/ui/table'
import { Repeat } from 'lucide-react'

interface Props {
  expenses: FixedExpenseResponse[]
  onEdit: (expense: FixedExpenseResponse) => void
  onDeactivate: (expense: FixedExpenseResponse) => void
}

export function FixedExpenseTable({ expenses, onEdit, onDeactivate }: Props) {
  const { t } = useTranslation()

  // Empty state
  if (!expenses?.length) {
    return (
      <div className="flex flex-col items-center justify-center gap-4 rounded-[24px] border border-border bg-card py-16">
        <div className="inline-flex h-16 w-16 items-center justify-center rounded-full bg-secondary">
          <Repeat size={28} className="text-muted-foreground" />
        </div>
        <p className="text-[15px] font-medium text-secondary-foreground">{t('fx.empty')}</p>
        <p className="text-[13px] text-muted-foreground">{t('fx.emptyHint')}</p>
      </div>
    )
  }

  return (
    <section aria-label={t('fx.title')} className="overflow-x-auto focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring rounded-[24px] border border-border bg-card">
      <Table aria-label={t('fx.title')} className="w-full text-left text-[14px]">
        <thead className="border-b border-border bg-muted text-[12px] uppercase tracking-wide text-muted-foreground">
          <tr>
            <th scope="col" className="py-3 pl-6 font-semibold">{t('fx.col.name')}</th>
            <th scope="col" className="py-3 font-semibold">{t('fx.col.category')}</th>
            <th scope="col" className="py-3 text-right font-semibold">{t('fx.col.amount')}</th>
            <th scope="col" className="py-3 pl-6 font-semibold">{t('fx.col.dueDay')}</th>
            <th scope="col" className="py-3 font-semibold">{t('fx.col.status')}</th>
            <th scope="col" className="py-3 pr-6 text-right font-semibold">{t('fx.col.actions')}</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-border/50">
          {expenses.map((expense) => (
            <FixedExpenseRow
              key={expense.id}
              expense={expense}
              onEdit={onEdit}
              onDeactivate={onDeactivate}
            />
          ))}
        </tbody>
      </Table>
    </section>
  )
}
