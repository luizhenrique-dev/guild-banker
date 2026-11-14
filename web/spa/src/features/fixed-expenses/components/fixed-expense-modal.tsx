import { useGuildMembers } from '../../guilds/api/use-guilds'
import { useActiveGuild } from '../../guilds/use-active-guild'
import { APIError } from '@/src/shared/lib/api-client'
import { useEffect, useRef } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Modal } from '@/src/shared/components/modal'
import { createFixedExpenseSchema, updateFixedExpenseSchema, type UpdateFixedExpenseInput } from '../schemas/fixed-expense.schema'
import { CATEGORY_ICONS, FIXED_EXPENSE_CATEGORIES, type FixedExpenseResponse } from '../types'
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
  const { activeGuild } = useActiveGuild()
  const members = useGuildMembers(activeGuild!.id)
  const mounted = useRef(true)
  useEffect(() => { mounted.current = true; return () => { mounted.current = false } }, [])
  const createMutation = useCreateFixedExpense()
  const updateMutation = useUpdateFixedExpense()
  const isEditing = !!editingExpense
  const isSubmitting = createMutation.isPending || updateMutation.isPending
  const sending = useRef(false)
  const { register, handleSubmit, reset, watch, formState: { errors } } = useForm<FormValues>({
    resolver: zodResolver(isEditing ? updateFixedExpenseSchema : createFixedExpenseSchema),
    defaultValues: { name: '', amount: '', due_day: 1, category: 'OTHER', responsible_user_id: null },
  })
  const responsibleUserID = watch('responsible_user_id')
  const resetCreate = createMutation.reset
  const resetUpdate = updateMutation.reset
  useEffect(() => {
    if (!open) return
    reset(editingExpense ? { ...editingExpense, amount: editingExpense.amount.replace('.', ',') } : { name: '', amount: '', due_day: 1, category: 'OTHER', responsible_user_id: null })
    resetCreate()
    resetUpdate()
  }, [open, editingExpense, reset, resetCreate, resetUpdate])

  const onSubmit = handleSubmit(async (data) => {
    if (sending.current || !activeGuild) return
    sending.current = true
    try {
      if (editingExpense) await updateMutation.mutateAsync({ guildId: activeGuild.id, id: editingExpense.id, input: data })
      else await createMutation.mutateAsync({ guildId: activeGuild.id, input: createFixedExpenseSchema.parse(data) })
      if (!mounted.current) return
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
  const error = (name: keyof FormValues) => errors[name] && <p id={`expense-${name}-error`} className="mt-1 text-[14px] text-destructive">{t(errors[name]?.message ?? '')}</p>
  const mutationError = isEditing ? updateMutation.error : createMutation.error
  const permissionDenied = mutationError instanceof APIError && mutationError.status === 403
  const inputClass = 'h-auto rounded-xl px-4 py-2.5 text-[16px]'
  return <Modal open={open} onOpenChange={(value) => { if (!value && !isSubmitting) onClose() }} title={t(isEditing ? 'fx.modal.editTitle' : 'fx.modal.createTitle')}>
    <form onSubmit={onSubmit} noValidate className="space-y-4">
      <fieldset disabled={isSubmitting} className="space-y-4">
        <div>
          <Label htmlFor="expense-name" className="mb-1 block text-[15px] text-secondary-foreground">{t('fx.modal.name')}</Label>
          <Input {...field('name')} readOnly={isEditing} {...register('name')} placeholder={t('fx.modal.namePlaceholder')} className={inputClass} />
          {error('name')}
        </div>
        <div>
          <Label htmlFor="expense-amount" className="mb-1 block text-[15px] text-secondary-foreground">{t('fx.modal.amount')}</Label>
          <Input {...field('amount')} inputMode="decimal" {...register('amount')} placeholder={t('fx.modal.amountPlaceholder')} className={`${inputClass} font-mono`} />
          {error('amount')}
        </div>
        <div>
          <Label htmlFor="expense-due_day" className="mb-1 block text-[15px] text-secondary-foreground">{t('fx.modal.dueDay')}</Label>
          <Input {...field('due_day')} type="number" min={1} max={31} step={1} {...register('due_day', { valueAsNumber: true })} placeholder={t('fx.modal.dueDayPlaceholder')} className={inputClass} />
          {error('due_day')}
        </div>
        <div>
          <Label htmlFor="expense-category" className="mb-1 block text-[15px] text-secondary-foreground">{t('fx.modal.category')}</Label>
          <select {...field('category')} {...register('category')} className="w-full rounded-xl border border-input bg-background px-4 py-2.5 text-[16px] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
            {FIXED_EXPENSE_CATEGORIES.map((cat) => <option key={cat} value={cat} aria-label={t(`category.${cat}`)}>{CATEGORY_ICONS[cat]} {t(`category.${cat}`)}</option>)}
          </select>
          {error('category')}
        </div>
        <div>
          <Label htmlFor="expense-responsible_user_id" className="mb-1 block text-[15px] text-secondary-foreground">{t('fx.responsible')}</Label>
          <select value={responsibleUserID ?? ''} {...field('responsible_user_id')} {...register('responsible_user_id', { setValueAs: (value) => value == null || value === '' ? null : Number(value) })} disabled={members.isPending || !!members.error} className="w-full rounded-xl border border-input bg-background px-4 py-2.5 text-[16px] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
            <option value="">{t('fx.unassigned')}</option>
            {responsibleUserID && !members.data?.some((member) => member.id === responsibleUserID) && <option value={responsibleUserID}>{t('fx.responsibleUnavailable')}</option>}
            {members.data?.map((member) => <option key={member.id} value={member.id}>{member.name}</option>)}
          </select>
          {members.isPending && <p role="status">{t('fx.membersLoading')}</p>}
          {members.error && <div role="alert"><p>{t('guild.membersError')}</p><button type="button" onClick={() => void members.refetch()}>{t('common.retry')}</button></div>}
          {error('responsible_user_id')}
        </div>
      </fieldset>
      {mutationError && <p role="alert" className="text-base text-destructive">{t(permissionDenied ? 'permission.denied' : 'fx.toast.error')}</p>}
      <div className="flex justify-end gap-3 pt-2">
        <Button type="button" variant="secondary" onClick={onClose} disabled={isSubmitting} className="h-auto rounded-full px-5 py-2.5 text-[16px]">{t('fx.modal.cancel')}</Button>
        <Button type="submit" disabled={isSubmitting} className="h-auto rounded-full px-5 py-2.5 text-[16px]">{t(isSubmitting ? isEditing ? 'fx.modal.saving' : 'fx.modal.creating' : 'fx.modal.save')}</Button>
      </div>
    </form>
  </Modal>
}
