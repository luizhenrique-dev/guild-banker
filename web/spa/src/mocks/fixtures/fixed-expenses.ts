// ============================================================
// Fixtures: Despesas fixas mock (realistas, PT-BR)
// Dados iguais ao protótipo HTML de referência.
// ============================================================
import type { FixedExpenseResponse } from '@/src/features/fixed-expenses/types'

export const fixedExpensesFixtures: FixedExpenseResponse[] = [
  {
    id: 1,
    name: 'Aluguel',
    amount: '2500.00',
    due_day: 10,
    category: 'HOUSING',
    status: 'ACTIVE',
  },
  {
    id: 2,
    name: 'Netflix',
    amount: '55.90',
    due_day: 15,
    category: 'SUBSCRIPTIONS',
    status: 'ACTIVE',
  },
  {
    id: 3,
    name: 'Plano de saúde',
    amount: '320.00',
    due_day: 5,
    category: 'INSURANCE',
    status: 'ACTIVE',
  },
  {
    id: 4,
    name: 'Academia',
    amount: '89.90',
    due_day: 1,
    category: 'HEALTH',
    status: 'PAUSED',
  },
  {
    id: 5,
    name: 'Seguro auto',
    amount: '180.00',
    due_day: 20,
    category: 'INSURANCE',
    status: 'ACTIVE',
  },
]
