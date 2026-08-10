'use client'
// ============================================================
// FixedExpenseTable — Tabela principal de despesas fixas
// Replica o design da seção fx-table do protótipo.
// ============================================================
import { useTranslation } from 'react-i18next'
import { FixedExpenseRow } from './fixed-expense-row'
import type { FixedExpenseResponse } from '../types'
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
      <div className="flex flex-col items-center justify-center gap-4 rounded-[24px] border border-[#dee1e6] bg-white py-16">
        <div className="inline-flex h-16 w-16 items-center justify-center rounded-full bg-[#eef0f3]">
          <Repeat size={28} className="text-[#7c828a]" />
        </div>
        <p className="text-[15px] font-medium text-[#5b616e]">{t('fx.empty')}</p>
        <p className="text-[13px] text-[#7c828a]">{t('fx.emptyHint')}</p>
      </div>
    )
  }

  return (
    <section className="overflow-hidden rounded-[24px] border border-[#dee1e6] bg-white">
      <table className="w-full text-left text-[14px]">
        <thead className="border-b border-[#dee1e6] bg-[#f7f7f7] text-[12px] uppercase tracking-wide text-[#7c828a]">
          <tr>
            <th className="py-3 pl-6 font-semibold">{t('fx.col.name')}</th>
            <th className="py-3 font-semibold">{t('fx.col.category')}</th>
            <th className="py-3 text-right font-semibold">{t('fx.col.amount')}</th>
            <th className="py-3 pl-6 font-semibold">{t('fx.col.dueDay')}</th>
            <th className="py-3 font-semibold">{t('fx.col.status')}</th>
            <th className="py-3 pr-6 text-right font-semibold">{t('fx.col.actions')}</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-[#eef0f3]">
          {expenses.map((expense) => (
            <FixedExpenseRow
              key={expense.id}
              expense={expense}
              onEdit={onEdit}
              onDeactivate={onDeactivate}
            />
          ))}
        </tbody>
      </table>
    </section>
  )
}
