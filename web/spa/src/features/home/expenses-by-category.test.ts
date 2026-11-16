import { expect, it } from 'vitest'
import type { FixedExpenseResponse } from '../fixed-expenses/types'
import { monthlyTotal } from '../fixed-expenses/monthly-total'
import { expensesByCategory } from './expenses-by-category'

const expense: FixedExpenseResponse = { id: 1, guild_id: 1, name: 'Expense', amount: '0.10', due_day: 1, category: 'OTHER', responsible_user_id: null, status: 'ACTIVE' }

it('sums every active expense in each category, ordered by descending total with exact cents', () => {
  const groups = expensesByCategory([
    expense,
    { ...expense, id: 2, amount: '0.20', due_day: 31 },
    { ...expense, id: 3, amount: '0.29', category: 'HOUSING' },
    { ...expense, id: 4, amount: '5000', category: 'HOUSING', status: 'PAUSED' },
    { ...expense, id: 5, amount: '9000', category: 'HEALTH', status: 'CANCELLED' },
  ])
  expect(groups.map(({ category }) => category)).toEqual(['OTHER', 'HOUSING'])
  expect(groups[0]!.cents).toBe(30n)
  expect(monthlyTotal(groups[0]!.expenses, 'pt-BR')).toContain('0,30')
  expect(groups[1]!.cents).toBe(29n)
})

it('returns no categories without active expenses', () => {
  expect(expensesByCategory([])).toEqual([])
  expect(expensesByCategory([{ ...expense, status: 'PAUSED' }])).toEqual([])
})
