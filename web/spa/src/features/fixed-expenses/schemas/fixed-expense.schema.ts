// ============================================================
// Zod Schemas: Fixed Expenses — derivados da OpenAPI spec
// ============================================================
// 📌 PADRÃO: Schemas Zod validam tanto no form (client-side)
// quanto poderiam validar no server-side (API routes).
// Derivados do CreateFixedExpenseRequest da spec.
// ============================================================
import { z } from 'zod'
import { FIXED_EXPENSE_CATEGORIES } from '../types'

export const createFixedExpenseSchema = z.object({
  name: z
    .string()
    .min(1, 'Nome obrigatório')
    .max(100, 'Máximo 100 caracteres'),
  amount: z
    .string()
    .min(1, 'Valor obrigatório')
    .regex(/^\d+([.,]\d{1,2})?$/, 'Valor inválido (ex: 100,00)'),
  due_day: z
    .number({ invalid_type_error: 'Informe o dia' })
    .int('Deve ser número inteiro')
    .min(1, 'Mínimo 1')
    .max(31, 'Máximo 31'),
  category: z.enum(
    FIXED_EXPENSE_CATEGORIES as unknown as [string, ...string[]],
    { errorMap: () => ({ message: 'Selecione uma categoria' }) }
  ),
})

export type CreateFixedExpenseInput = z.infer<typeof createFixedExpenseSchema>

// 📌 PADRÃO: Schema de edição reutiliza o de criação.
// Na spec, UpdateFixedExpenseRequest tem campos opcionais,
// mas no form sempre enviamos todos.
export const updateFixedExpenseSchema = createFixedExpenseSchema
export type UpdateFixedExpenseInput = z.infer<typeof updateFixedExpenseSchema>
