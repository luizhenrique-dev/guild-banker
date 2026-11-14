// ============================================================
// Fixtures: Despesas fixas mock (realistas, PT-BR)
// Dados iguais ao protótipo HTML de referência.
// ============================================================
import type { FixedExpenseResponse } from '@/src/features/fixed-expenses/types'

export const fixedExpensesFixtures: FixedExpenseResponse[] = [
  {
    guild_id: 1,
    id: 1,
    name: 'Aluguel',
    amount: '2500.00',
    due_day: 10,
    category: 'HOUSING',
    responsible_user_id: null, status: 'ACTIVE',
  },
  {
    guild_id: 1,
    id: 2,
    name: 'Netflix',
    amount: '55.90',
    due_day: 15,
    category: 'SUBSCRIPTIONS',
    responsible_user_id: null, status: 'ACTIVE',
  },
  {
    guild_id: 1,
    id: 3,
    name: 'Plano de saúde',
    amount: '320.00',
    due_day: 5,
    category: 'INSURANCE',
    responsible_user_id: null, status: 'ACTIVE',
  },
  {
    guild_id: 1,
    id: 4,
    name: 'Academia',
    amount: '89.90',
    due_day: 1,
    category: 'HEALTH',
    responsible_user_id: null, status: 'PAUSED',
  },
  {
    guild_id: 1,
    id: 5,
    name: 'Seguro auto',
    amount: '180.00',
    due_day: 20,
    category: 'INSURANCE',
    responsible_user_id: null, status: 'ACTIVE',
  },
]

export const sharedFixedExpensesFixtures: FixedExpenseResponse[] = [
 { id: 20, guild_id: 2, name: 'Internet da República', amount: '150.00', due_day: 12, category: 'OTHER', responsible_user_id: null, status: 'ACTIVE' },
]
