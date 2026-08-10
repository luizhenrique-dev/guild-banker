'use client'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useTransactions, useCancelTransaction, useToggleVisibility } from '../api/use-transactions'
import { TRANSACTION_CATEGORIES, type TransactionResponse } from '../types'
import { TableSkeleton } from '@/src/shared/components/loading-skeleton'
import {
  ArrowDownCircle,
  ArrowUpCircle,
  Eye,
  EyeOff,
  Lock,
  Unlock,
  Trash2,
  Upload,
} from 'lucide-react'

function formatBRL(amount: string): string {
  const num = parseFloat(amount)
  if (isNaN(num)) return 'R$ 0,00'
  return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(num)
}

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleDateString('pt-BR', { day: '2-digit', month: '2-digit', year: 'numeric', timeZone: 'UTC' })
  } catch {
    return iso
  }
}

export function TransactionsPageClient() {
  const { t } = useTranslation()
  const [category, setCategory] = useState<string>('')
  const [type, setType] = useState<string>('')

  const filters: Record<string, string> = {}
  if (category) filters['category'] = category
  if (type) filters['type'] = type

  const { data, isLoading, error } = useTransactions(filters)
  const cancelMutation = useCancelTransaction()
  const visibilityMutation = useToggleVisibility()

  const transactions = data?.items ?? []

  return (
    <main className="flex-1 px-8 py-8">
      <div className="mb-2 flex items-center justify-between">
        <div>
          <h1 className="text-[32px] font-normal tracking-[-0.8px]">{t('tx.title')}</h1>
          <p className="text-[13px] text-[#5b616e]">{t('tx.subtitle')}</p>
        </div>
      </div>

      {/* Filtros */}
      <div className="mb-6 flex items-center gap-3">
        <select
          value={type}
          onChange={(e) => setType(e.target.value)}
          className="rounded-xl border border-[#dee1e6] bg-white px-3 py-2 text-[13px] outline-none focus:border-[#0052ff]"
        >
          <option value="">{t('tx.filterAll')} — {t('tx.filterType')}</option>
          <option value="EXPENSE">{t('tx.type.EXPENSE')}</option>
          <option value="INCOME">{t('tx.type.INCOME')}</option>
        </select>
        <select
          value={category}
          onChange={(e) => setCategory(e.target.value)}
          className="rounded-xl border border-[#dee1e6] bg-white px-3 py-2 text-[13px] outline-none focus:border-[#0052ff]"
        >
          <option value="">{t('tx.filterAll')} — {t('tx.filterCategory')}</option>
          {TRANSACTION_CATEGORIES.map((cat) => (
            <option key={cat} value={cat}>
              {t(`txCategory.${cat}`)}
            </option>
          ))}
        </select>
      </div>

      {/* Tabela */}
      {isLoading ? (
        <TableSkeleton rows={5} />
      ) : error ? (
        <div className="rounded-[24px] border border-[#dee1e6] bg-white p-8 text-center text-[#cf202f]">
          {t('fx.toast.error')}
        </div>
      ) : transactions.length === 0 ? (
        <div className="flex flex-col items-center gap-4 rounded-[24px] border border-[#dee1e6] bg-white py-16">
          <ArrowDownCircle size={28} className="text-[#7c828a]" />
          <p className="text-[15px] text-[#5b616e]">{t('tx.empty')}</p>
        </div>
      ) : (
        <section className="overflow-hidden rounded-[24px] border border-[#dee1e6] bg-white">
          <table className="w-full text-left text-[14px]">
            <thead className="border-b border-[#dee1e6] bg-[#f7f7f7] text-[12px] uppercase tracking-wide text-[#7c828a]">
              <tr>
                <th className="py-3 pl-6 font-semibold">Descrição</th>
                <th className="py-3 font-semibold">{t('fx.col.category')}</th>
                <th className="py-3 text-right font-semibold">{t('fx.col.amount')}</th>
                <th className="py-3 pl-6 font-semibold">Data</th>
                <th className="py-3 font-semibold">Fonte</th>
                <th className="py-3 pr-6 text-right font-semibold">{t('fx.col.actions')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#eef0f3]">
              {transactions.map((tx: TransactionResponse) => (
                <tr key={tx.id} className="transition-colors hover:bg-[#f7f7f7]/50">
                  <td className="py-4 pl-6">
                    <div className="flex items-center gap-3">
                      <span className={`inline-flex h-8 w-8 items-center justify-center rounded-full ${
                        tx.type === 'INCOME' ? 'bg-green-50 text-[#05b169]' : 'bg-red-50 text-[#cf202f]'
                      }`}>
                        {tx.type === 'INCOME' ? <ArrowUpCircle size={16} /> : <ArrowDownCircle size={16} />}
                      </span>
                      <span className="font-medium">{tx.description}</span>
                    </div>
                  </td>
                  <td className="py-4">
                    <span className="rounded-full bg-[#eef0f3] px-2.5 py-1 text-[12px] font-semibold text-[#5b616e]">
                      {t(`txCategory.${tx.category}`)}
                    </span>
                  </td>
                  <td className={`py-4 text-right font-mono font-medium ${
                    tx.type === 'INCOME' ? 'text-[#05b169]' : 'text-[#cf202f]'
                  }`}>
                    {tx.type === 'INCOME' ? '+ ' : '- '}{formatBRL(tx.amount)}
                  </td>
                  <td className="py-4 pl-6 font-mono text-[13px] text-[#5b616e]">
                    {formatDate(tx.occurredAt)}
                  </td>
                  <td className="py-4">
                    <span className="inline-flex items-center gap-1 text-[12px] text-[#7c828a]">
                      {tx.source === 'IMPORT' ? <Upload size={12} /> : null}
                      {t(`tx.source.${tx.source}`)}
                    </span>
                  </td>
                  <td className="py-4 pr-6 text-right">
                    <button
                      type="button"
                      onClick={() =>
                        visibilityMutation.mutate({
                          id: tx.id,
                          visibility: tx.visibility === 'PUBLIC' ? 'PRIVATE' : 'PUBLIC',
                        })
                      }
                      className="mr-2 text-[#7c828a] transition-colors hover:text-[#0052ff]"
                      title={tx.visibility === 'PUBLIC' ? t('tx.visibility.PUBLIC') : t('tx.visibility.PRIVATE')}
                    >
                      {tx.visibility === 'PUBLIC' ? <Unlock size={15} /> : <Lock size={15} />}
                    </button>
                    <button
                      type="button"
                      onClick={() => {
                        if (window.confirm(t('tx.cancel') + '?')) {
                          cancelMutation.mutate(tx.id)
                        }
                      }}
                      className="text-[#7c828a] transition-colors hover:text-[#cf202f]"
                      title={t('tx.cancel')}
                    >
                      <Trash2 size={15} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </section>
      )}
    </main>
  )
}
