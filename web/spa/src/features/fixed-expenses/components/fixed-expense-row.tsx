'use client'
// ============================================================
// FixedExpenseRow — Uma linha da tabela de despesas fixas
// Replica fielmente o design do protótipo HTML.
// ============================================================
import { useTranslation } from 'react-i18next'
import {
  Home,
  Monitor,
  HeartPulse,
  GraduationCap,
  Car,
  User,
  Receipt,
  MoreHorizontal,
  Pencil,
  Power,
} from 'lucide-react'
import type { FixedExpenseResponse, FixedExpenseCategory } from '../types'

interface Props {
  expense: FixedExpenseResponse
  onEdit: (expense: FixedExpenseResponse) => void
  onDeactivate: (expense: FixedExpenseResponse) => void
}

// 📌 PADRÃO: Mapa de ícones por categoria (Lucide React)
const categoryIconMap: Record<FixedExpenseCategory, React.ReactNode> = {
  HOUSING: <Home size={16} />,
  SUBSCRIPTIONS: <Monitor size={16} />,
  INSURANCE: <HeartPulse size={16} />,
  EDUCATION: <GraduationCap size={16} />,
  TRANSPORTATION: <Car size={16} />,
  HEALTH: <HeartPulse size={16} />,
  PERSONAL: <User size={16} />,
  TAXES: <Receipt size={16} />,
  OTHER: <MoreHorizontal size={16} />,
}

// 📌 PADRÃO: Formatação monetária BRL com Intl.NumberFormat
function formatBRL(amount: string): string {
  const num = parseFloat(amount)
  if (isNaN(num)) return 'R$ 0,00'
  return new Intl.NumberFormat('pt-BR', {
    style: 'currency',
    currency: 'BRL',
  }).format(num)
}

export function FixedExpenseRow({ expense, onEdit, onDeactivate }: Props) {
  const { t } = useTranslation()

  const icon = categoryIconMap[expense.category] ?? <MoreHorizontal size={16} />
  const categoryLabel = t(`category.${expense.category}`)
  const statusColor = expense.status === 'ACTIVE' ? '#05b169' : '#7c828a'

  return (
    <tr className="transition-colors hover:bg-[#f7f7f7]/50">
      {/* Nome com ícone */}
      <td className="py-4 pl-6">
        <div className="flex items-center gap-3">
          <span className="inline-flex h-9 w-9 items-center justify-center rounded-full bg-[#eef0f3] text-[#5b616e]">
            {icon}
          </span>
          <span className="font-semibold">{expense.name}</span>
        </div>
      </td>

      {/* Categoria badge */}
      <td className="py-4">
        <span className="rounded-full bg-[#eef0f3] px-2.5 py-1 text-[12px] font-semibold text-[#5b616e]">
          {categoryLabel}
        </span>
      </td>

      {/* Valor em JetBrains Mono */}
      <td className="py-4 text-right font-mono font-medium">
        {formatBRL(expense.amount)}
      </td>

      {/* Vencimento */}
      <td className="py-4 pl-6 font-mono text-[#5b616e]">
        {t('fx.dueDay', { day: expense.due_day })}
      </td>

      {/* Status badge */}
      <td className="py-4">
        <span
          className="inline-flex items-center gap-1.5 text-[12px] font-semibold"
          style={{ color: statusColor }}
        >
          <span
            className="h-2 w-2 rounded-full"
            style={{ backgroundColor: statusColor }}
          />
          {expense.status}
        </span>
      </td>

      {/* Ações */}
      <td className="py-4 pr-6 text-right">
        <button
          type="button"
          onClick={() => onEdit(expense)}
          className="mr-2 text-[#7c828a] transition-colors hover:text-[#0052ff]"
          title="Editar"
        >
          <Pencil size={15} />
        </button>
        <button
          type="button"
          onClick={() => onDeactivate(expense)}
          className="text-[#7c828a] transition-colors hover:text-[#cf202f]"
          title="Desativar"
        >
          <Power size={15} />
        </button>
      </td>
    </tr>
  )
}
