// ============================================================
// Types: Fixed Expenses — derivados da OpenAPI spec
// ============================================================

// 📌 PADRÃO: Enums como union types garantem type-safety
// e IntelliSense no VS Code.
export type FixedExpenseCategory =
  | 'HOUSING'
  | 'SUBSCRIPTIONS'
  | 'INSURANCE'
  | 'EDUCATION'
  | 'TRANSPORTATION'
  | 'HEALTH'
  | 'PERSONAL'
  | 'TAXES'
  | 'OTHER'

export type FixedExpenseStatus = 'ACTIVE' | 'PAUSED' | 'CANCELLED'

// 📌 PADRÃO: Interface alinhada 1:1 com o schema FixedExpenseResponse da spec.
// O campo `amount` é string decimal (shopspring/decimal no Go).
export interface FixedExpenseResponse {
  id: number
  name: string
  amount: string // decimal string, ex: "2500.00"
  due_day: number
  category: FixedExpenseCategory
  status: FixedExpenseStatus
}

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
