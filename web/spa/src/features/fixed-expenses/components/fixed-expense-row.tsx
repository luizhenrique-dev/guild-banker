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
  Trash2,
  Pause,
  Play,
  ShoppingBasket,
  Zap,
  Utensils,
  Clapperboard,
  ShoppingBag,
  PawPrint,
  Plane,
  TrendingUp,
} from 'lucide-react'
import { CATEGORY_ICONS, type FixedExpenseResponse, type FixedExpenseCategory } from '../types'
import { useSetFixedExpenseStatus } from '../api/use-fixed-expenses'
import { APIError } from '@/src/shared/lib/api-client'

interface Props {
  expense: FixedExpenseResponse
  responsibleName?: string
  onEdit: (expense: FixedExpenseResponse) => void
  onDelete: (expense: FixedExpenseResponse) => void
}

// 📌 PADRÃO: Mapa de ícones por categoria (Lucide React)
const categoryIconMap: Record<FixedExpenseCategory, React.ReactNode> = {
  GROCERY: <ShoppingBasket size={16} />,
  HOUSING: <Home size={16} />,
  UTILITIES: <Zap size={16} />,
  SUBSCRIPTIONS: <Monitor size={16} />,
  INSURANCE: <HeartPulse size={16} />,
  EDUCATION: <GraduationCap size={16} />,
  TRANSPORTATION: <Car size={16} />,
  HEALTH: <HeartPulse size={16} />,
  PERSONAL: <User size={16} />,
  TAXES: <Receipt size={16} />,
  OTHER: <MoreHorizontal size={16} />,
  FOOD_AND_DINING: <Utensils size={16} />,
  ENTERTAINMENT: <Clapperboard size={16} />,
  SHOPPING: <ShoppingBag size={16} />,
  PETS: <PawPrint size={16} />,
  TRAVEL: <Plane size={16} />,
  INVESTMENTS: <TrendingUp size={16} />,
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

export function FixedExpenseRow({ expense, responsibleName, onEdit, onDelete }: Props) {
  const { t } = useTranslation()
  const statusMutation = useSetFixedExpenseStatus()

  const icon = categoryIconMap[expense.category] ?? <MoreHorizontal size={16} />
  const categoryLabel = t(`category.${expense.category}`)
  const statusColor = expense.status === 'ACTIVE' ? '#05b169' : expense.status === 'PAUSED' ? '#a16207' : '#7c828a'

  return (
    <tr className="transition-colors hover:bg-muted/50">
      {/* Nome com ícone */}
      <td className="py-4 pl-6">
        <div className="flex items-center gap-3">
          <span className="inline-flex h-9 w-9 items-center justify-center rounded-full bg-secondary text-secondary-foreground">
            {icon}
          </span>
          <span className="font-semibold">{expense.name}</span>
        </div>
      </td>

      {/* Categoria badge */}
      <td className="py-4">
        <span className="inline-flex items-center gap-1.5 rounded-full bg-secondary px-2.5 py-1 text-[12px] font-semibold text-secondary-foreground">
          <span aria-hidden="true">{CATEGORY_ICONS[expense.category]}</span>
          {categoryLabel}
        </span>
      </td>

      <td className="px-4 py-4 text-secondary-foreground">
        {expense.responsible_user_id == null ? t('fx.unassigned') : responsibleName ?? t('fx.responsibleUnavailable')}
      </td>

      {/* Valor em JetBrains Mono */}
      <td className="py-4 text-right font-mono font-medium">
        {formatBRL(expense.amount)}
      </td>

      {/* Vencimento */}
      <td className="py-4 pl-6 font-mono text-secondary-foreground">
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
          {t(`fx.status.${expense.status}`)}
        </span>
      </td>

      {/* Ações */}
      <td className="py-4 pr-6 text-right">
        <button
          type="button"
          onClick={() => onEdit(expense)}
          className="mr-2 text-muted-foreground transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring hover:text-primary"
          title={t('fx.edit', { name: expense.name })}
          aria-label={t('fx.edit', { name: expense.name })}
        >
          <Pencil size={15} />
        </button>
        <button
          type="button"
          disabled={statusMutation.isPending}
          onClick={() => statusMutation.mutate({ guildId: expense.guild_id, id: expense.id, status: expense.status === 'ACTIVE' ? 'PAUSED' : 'ACTIVE' })}
          className="mr-2 text-muted-foreground transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring hover:text-primary disabled:opacity-50"
          title={t(expense.status === 'ACTIVE' ? 'fx.pause.action' : 'fx.resume.action', { name: expense.name })}
          aria-label={t(expense.status === 'ACTIVE' ? 'fx.pause.action' : 'fx.resume.action', { name: expense.name })}
        >
          {expense.status === 'ACTIVE' ? <Pause size={15} /> : <Play size={15} />}
        </button>
        <button
          type="button"
          onClick={() => onDelete(expense)}
          className="text-muted-foreground transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring hover:text-destructive"
          title={t('fx.delete.action', { name: expense.name })}
          aria-label={t('fx.delete.action', { name: expense.name })}
        >
          <Trash2 size={15} />
        </button>
        {statusMutation.error && <p role="alert" className="mt-2 text-sm text-destructive">{t(statusMutation.error instanceof APIError && statusMutation.error.status === 403 ? 'permission.denied' : 'fx.toast.error')}</p>}
      </td>
    </tr>
  )
}
