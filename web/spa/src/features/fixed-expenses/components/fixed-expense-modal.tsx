'use client'
// ============================================================
// FixedExpenseModal — Modal de criação/edição de despesa fixa
// Usa React Hook Form + Zod para validação.
// ============================================================
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useTranslation } from 'react-i18next'
import { Modal } from '@/src/shared/components/modal'
import {
  createFixedExpenseSchema,
  type CreateFixedExpenseInput,
} from '../schemas/fixed-expense.schema'
import { FIXED_EXPENSE_CATEGORIES, type FixedExpenseResponse } from '../types'
import { useCreateFixedExpense, useUpdateFixedExpense } from '../api/use-fixed-expenses'
import { useEffect } from 'react'

interface Props {
  open: boolean
  onClose: () => void
  editingExpense?: FixedExpenseResponse | null
}

export function FixedExpenseModal({ open, onClose, editingExpense }: Props) {
  const { t } = useTranslation()
  const createMutation = useCreateFixedExpense()
  const updateMutation = useUpdateFixedExpense()

  const isEditing = !!editingExpense
  const isSubmitting = createMutation.isPending || updateMutation.isPending

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateFixedExpenseInput>({
    resolver: zodResolver(createFixedExpenseSchema),
    defaultValues: {
      name: '',
      amount: '',
      due_day: 1,
      category: 'OTHER',
    },
  })

  // 📌 PADRÃO: Preencher form quando editando
  useEffect(() => {
    if (editingExpense) {
      reset({
        name: editingExpense.name,
        amount: editingExpense.amount.replace('.', ','),
        due_day: editingExpense.due_day,
        category: editingExpense.category,
      })
    } else {
      reset({ name: '', amount: '', due_day: 1, category: 'OTHER' })
    }
  }, [editingExpense, reset])

  const onSubmit = handleSubmit(async (data: CreateFixedExpenseInput) => {
    try {
      if (isEditing && editingExpense) {
        await updateMutation.mutateAsync({ id: editingExpense.id, input: data })
      } else {
        await createMutation.mutateAsync(data)
      }
      onClose()
    } catch {
      // Erro já tratado pelo hook (toast)
    }
  })

  return (
    <Modal
      open={open}
      onOpenChange={(isOpen: boolean) => !isOpen && onClose()}
      title={isEditing ? t('fx.modal.editTitle') : t('fx.modal.createTitle')}
    >
      <form onSubmit={onSubmit} className="space-y-4">
        {/* Nome */}
        <div>
          <label className="mb-1 block text-[13px] font-medium text-[#5b616e]">
            {t('fx.modal.name')}
          </label>
          <input
            type="text"
            {...register('name')}
            placeholder={t('fx.modal.namePlaceholder')}
            className="w-full rounded-xl border border-[#dee1e6] px-4 py-2.5 text-[14px] outline-none transition-colors focus:border-[#0052ff] focus:ring-1 focus:ring-[#0052ff]"
          />
          {errors.name && (
            <p className="mt-1 text-[12px] text-[#cf202f]">{errors.name.message}</p>
          )}
        </div>

        {/* Valor */}
        <div>
          <label className="mb-1 block text-[13px] font-medium text-[#5b616e]">
            {t('fx.modal.amount')}
          </label>
          <input
            type="text"
            {...register('amount')}
            placeholder={t('fx.modal.amountPlaceholder')}
            className="w-full rounded-xl border border-[#dee1e6] px-4 py-2.5 font-mono text-[14px] outline-none transition-colors focus:border-[#0052ff] focus:ring-1 focus:ring-[#0052ff]"
          />
          {errors.amount && (
            <p className="mt-1 text-[12px] text-[#cf202f]">{errors.amount.message}</p>
          )}
        </div>

        {/* Dia de vencimento */}
        <div>
          <label className="mb-1 block text-[13px] font-medium text-[#5b616e]">
            {t('fx.modal.dueDay')}
          </label>
          <input
            type="number"
            {...register('due_day', { valueAsNumber: true })}
            min={1}
            max={31}
            placeholder={t('fx.modal.dueDayPlaceholder')}
            className="w-full rounded-xl border border-[#dee1e6] px-4 py-2.5 text-[14px] outline-none transition-colors focus:border-[#0052ff] focus:ring-1 focus:ring-[#0052ff]"
          />
          {errors.due_day && (
            <p className="mt-1 text-[12px] text-[#cf202f]">{errors.due_day.message}</p>
          )}
        </div>

        {/* Categoria */}
        <div>
          <label className="mb-1 block text-[13px] font-medium text-[#5b616e]">
            {t('fx.modal.category')}
          </label>
          <select
            {...register('category')}
            className="w-full rounded-xl border border-[#dee1e6] px-4 py-2.5 text-[14px] outline-none transition-colors focus:border-[#0052ff] focus:ring-1 focus:ring-[#0052ff]"
          >
            <option value="" disabled>
              {t('fx.modal.selectCategory')}
            </option>
            {FIXED_EXPENSE_CATEGORIES.map((cat) => (
              <option key={cat} value={cat}>
                {t(`category.${cat}`)}
              </option>
            ))}
          </select>
          {errors.category && (
            <p className="mt-1 text-[12px] text-[#cf202f]">{errors.category.message}</p>
          )}
        </div>

        {/* Botões */}
        <div className="flex justify-end gap-3 pt-2">
          <button
            type="button"
            onClick={onClose}
            disabled={isSubmitting}
            className="rounded-full bg-[#eef0f3] px-5 py-2.5 text-[14px] font-semibold text-[#5b616e] transition-colors hover:bg-[#dee1e6]"
          >
            {t('fx.modal.cancel')}
          </button>
          <button
            type="submit"
            disabled={isSubmitting}
            className="rounded-full bg-[#0052ff] px-5 py-2.5 text-[14px] font-semibold text-white transition-colors hover:bg-[#003ecc] disabled:opacity-50"
          >
            {isSubmitting
              ? isEditing
                ? t('fx.modal.saving')
                : t('fx.modal.creating')
              : t('fx.modal.save')}
          </button>
        </div>
      </form>
    </Modal>
  )
}
