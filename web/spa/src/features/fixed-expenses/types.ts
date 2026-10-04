// ============================================================
// Types: Fixed Expenses — derivados da OpenAPI spec
// ============================================================

// 📌 PADRÃO: Enums como union types garantem type-safety
// e IntelliSense no VS Code.
import type { components } from '@/src/shared/api/schema'

export type FixedExpenseCategory = components['schemas']['FixedExpenseCategory']
export type FixedExpenseStatus = components['schemas']['FixedExpenseStatus']
export type FixedExpenseResponse = components['schemas']['FixedExpenseResponse']

// Constante com todas as categorias para iteração
export const FIXED_EXPENSE_CATEGORIES: FixedExpenseCategory[] = [
  'HOUSING',
  'SUBSCRIPTIONS',
  'INSURANCE',
  'EDUCATION',
  'TRANSPORTATION',
  'HEALTH',
  'PERSONAL',
  'TAXES',
  'OTHER',
]

// 📌 PADRÃO: Mapa de ícones por categoria para o componente de tabela
export const CATEGORY_ICONS: Record<FixedExpenseCategory, string> = {
  HOUSING: 'Home',
  SUBSCRIPTIONS: 'Monitor',
  INSURANCE: 'HeartPulse',
  EDUCATION: 'GraduationCap',
  TRANSPORTATION: 'Car',
  HEALTH: 'HeartPulse',
  PERSONAL: 'User',
  TAXES: 'Receipt',
  OTHER: 'MoreHorizontal',
}
