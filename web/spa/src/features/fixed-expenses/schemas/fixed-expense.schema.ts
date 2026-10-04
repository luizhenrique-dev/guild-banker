import { z } from 'zod'
import { FIXED_EXPENSE_CATEGORIES, type FixedExpenseCategory } from '../types'

const amount = z.string()
  .regex(/^\d+([.,]\d{1,2})?$/, 'validation.amount')
  .refine((value) => /[1-9]/.test(value), 'validation.amount')
  .transform((value) => value.replace(',', '.'))

export const createFixedExpenseSchema = z.object({
  // The 100-character limit is a local UI policy.
  name: z.string().trim().min(1, 'validation.name').max(100, 'validation.nameLength'),
  amount,
  due_day: z.number({ invalid_type_error: 'validation.dueDay' })
    .int('validation.dueDay').min(1, 'validation.dueDay').max(31, 'validation.dueDay'),
  category: z.enum(FIXED_EXPENSE_CATEGORIES as [FixedExpenseCategory, ...FixedExpenseCategory[]], {
    errorMap: () => ({ message: 'validation.category' }),
  }),
})
export const updateFixedExpenseSchema = createFixedExpenseSchema.omit({ name: true })
export type CreateFixedExpenseInput = z.infer<typeof createFixedExpenseSchema>
export type UpdateFixedExpenseInput = z.infer<typeof updateFixedExpenseSchema>
