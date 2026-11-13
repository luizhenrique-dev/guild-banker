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
  'GROCERY',
  'HOUSING',
  'UTILITIES',
  'SUBSCRIPTIONS',
  'INSURANCE',
  'EDUCATION',
  'TRANSPORTATION',
  'HEALTH',
  'PERSONAL',
  'TAXES',
  'OTHER',
  'FOOD_AND_DINING',
  'ENTERTAINMENT',
  'SHOPPING',
  'PETS',
  'TRAVEL',
  'INVESTMENTS',
]

// Emojis para apresentar as categorias no select nativo do formulário.
export const CATEGORY_ICONS: Record<FixedExpenseCategory, string> = {
  GROCERY: '🛒',
  HOUSING: '🏠',
  UTILITIES: '💡',
  SUBSCRIPTIONS: '🔁',
  INSURANCE: '🛡️',
  EDUCATION: '🎓',
  TRANSPORTATION: '🚗',
  HEALTH: '🩺',
  PERSONAL: '👤',
  TAXES: '🧾',
  OTHER: '📦',
  FOOD_AND_DINING: '🍽️',
  ENTERTAINMENT: '🎬',
  SHOPPING: '🛍️',
  PETS: '🐾',
  TRAVEL: '✈️',
  INVESTMENTS: '📈',
}
