import { useEffect, useRef } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Modal } from '@/src/shared/components/modal'
import { createFixedExpenseSchema, updateFixedExpenseSchema, type UpdateFixedExpenseInput } from '../schemas/fixed-expense.schema'
import { FIXED_EXPENSE_CATEGORIES, type FixedExpenseResponse } from '../types'
import { useCreateFixedExpense, useUpdateFixedExpense } from '../api/use-fixed-expenses'

interface Props {
  open: boolean
  onClose: () => void
  onSaved: () => void
  editingExpense?: FixedExpenseResponse | null
}
type FormValues = UpdateFixedExpenseInput & { name?: string }

export function FixedExpenseModal({ open, onClose, onSaved, editingExpense }: Props) {
  const { t } = useTranslation()
  const createMutation = useCreateFixedExpense()
  const updateMutation = useUpdateFixedExpense()
  const isEditing = !!editingExpense
  const isSubmitting = createMutation.isPending || updateMutation.isPending
  const sending = useRef(false)
  const { register, handleSubmit, reset, formState: { errors } } = useForm<FormValues>({
    resolver: zodResolver(isEditing ? updateFixedExpenseSchema : createFixedExpenseSchema),
    defaultValues: { name: '', amount: '', due_day: 1, category: 'OTHER' },
  })
  const resetCreate = createMutation.reset
  const resetUpdate = updateMutation.reset
  useEffect(() => {
    if (!open) return
    reset(editingExpense ? { ...editingExpense, amount: editingExpense.amount.replace('.', ',') } : { name: '', amount: '', due_day: 1, category: 'OTHER' })
    resetCreate()
    resetUpdate()
  }, [open, editingExpense, reset, resetCreate, resetUpdate])

  const onSubmit = handleSubmit(async (data) => {
    if (sending.current) return
    sending.current = true
    try {
      if (editingExpense) await updateMutation.mutateAsync({ id: editingExpense.id, input: data })
      else await createMutation.mutateAsync(createFixedExpenseSchema.parse(data))
      onSaved()
      onClose()
    } catch {
      // Keep the draft; the mutation error is presented below.
    } finally {
      sending.current = false
    }
  })
  const field = (name: keyof FormValues) => ({
    id: `expense-${name}`,
    'aria-invalid': !!errors[name],
    'aria-describedby': errors[name] ? `expense-${name}-error` : undefined,
  })
  const error = (name: keyof FormValues) => errors[name] && <p id={`expense-${name}-error`} className="mt-1 text-[12px] text-destructive">{t(errors[name]?.message ?? '')}</p>
  const inputClass = 'h-auto rounded-xl px-4 py-2.5 text-[14px]'
  return <Modal open={open} onOpenChange={(value) => { if (!value && !isSubmitting) onClose() }} title={t(isEditing ? 'fx.modal.editTitle' : 'fx.modal.createTitle')}>
    <form onSubmit={onSubmit} noValidate className="space-y-4">
      <fieldset disabled={isSubmitting} className="space-y-4">
        <div>
          <Label htmlFor="expense-name" className="mb-1 block text-[13px] text-secondary-foreground">{t('fx.modal.name')}</Label>
          <Input {...field('name')} readOnly={isEditing} {...register('name')} placeholder={t('fx.modal.namePlaceholder')} className={inputClass} />
          {error('name')}
        </div>
        <div>
          <Label htmlFor="expense-amount" className="mb-1 block text-[13px] text-secondary-foreground">{t('fx.modal.amount')}</Label>
          <Input {...field('amount')} inputMode="decimal" {...register('amount')} placeholder={t('fx.modal.amountPlaceholder')} className={`${inputClass} font-mono`} />
          {error('amount')}
        </div>
        <div>
          <Label htmlFor="expense-due_day" className="mb-1 block text-[13px] text-secondary-foreground">{t('fx.modal.dueDay')}</Label>
          <Input {...field('due_day')} type="number" min={1} max={31} step={1} {...register('due_day', { valueAsNumber: true })} placeholder={t('fx.modal.dueDayPlaceholder')} className={inputClass} />
          {error('due_day')}
        </div>
        <div>
          <Label htmlFor="expense-category" className="mb-1 block text-[13px] text-secondary-foreground">{t('fx.modal.category')}</Label>
          <select {...field('category')} {...register('category')} className="w-full rounded-xl border border-input bg-background px-4 py-2.5 text-[14px] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
            {FIXED_EXPENSE_CATEGORIES.map((cat) => <option key={cat} value={cat}>{t(`category.${cat}`)}</option>)}
          </select>
          {error('category')}
        </div>
      </fieldset>
      {(isEditing ? updateMutation.error : createMutation.error) && <p role="alert" className="text-sm text-destructive">{t('fx.toast.error')}</p>}
      <div className="flex justify-end gap-3 pt-2">
        <Button type="button" variant="secondary" onClick={onClose} disabled={isSubmitting} className="h-auto rounded-full px-5 py-2.5">{t('fx.modal.cancel')}</Button>
        <Button type="submit" disabled={isSubmitting} className="h-auto rounded-full px-5 py-2.5">{t(isSubmitting ? isEditing ? 'fx.modal.saving' : 'fx.modal.creating' : 'fx.modal.save')}</Button>
      </div>
    </form>
  </Modal>
}
