import { expect, it } from 'vitest'
import { monthlyTotal } from './monthly-total'
import type { FixedExpenseResponse } from './types'

const expense: FixedExpenseResponse = { id: 1, guild_id: 1, name: 'Expense', amount: '0.10', due_day: 1, category: 'OTHER', responsible_user_id: null, status: 'ACTIVE' }

it('adds cents exactly and excludes paused and cancelled expenses', () => {
  const expenses: FixedExpenseResponse[] = [expense, { ...expense, amount: '0.20' }, { ...expense, amount: '900', status: 'PAUSED' }, { ...expense, amount: '800', status: 'CANCELLED' }]
  expect(monthlyTotal(expenses, 'pt-BR').replace(/\s/g, ' ')).toBe('R$ 0,30')
})

it('shows zero for an empty forecast and formats decimal strings without a fractional part', () => {
  expect(monthlyTotal([], 'pt-BR').replace(/\s/g, ' ')).toBe('R$ 0,00')
  expect(monthlyTotal([{ ...expense, amount: '100' }], 'en')).toBe('R$100.00')
})
