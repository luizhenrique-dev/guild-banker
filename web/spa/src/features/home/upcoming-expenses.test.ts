import { describe, expect, it } from 'vitest'
import type { FixedExpenseResponse } from '../fixed-expenses/types'
import { upcomingExpenses } from './upcoming-expenses'

function expense(id: number, day: number, status: FixedExpenseResponse['status'] = 'ACTIVE'): FixedExpenseResponse {
  return { id, guild_id: 1, name: `Expense ${id}`, amount: '123.45', due_day: day, payment_method: 'CREDIT_CARD', category: 'OTHER', responsible_user_id: null, status }
}

describe('upcomingExpenses', () => {
  it('filters status and past days, sorts ties by ID, then limits to five without changing input', () => {
    const input = [expense(9, 30), expense(6, 17), expense(3, 15), expense(2, 15), expense(5, 16), expense(4, 15), expense(1, 14), expense(7, 15, 'PAUSED'), expense(8, 15, 'CANCELLED')]
    const before = structuredClone(input)
    expect(upcomingExpenses(input, new Date(2026, 10, 15, 23, 59)).map(({ expense }) => expense.id)).toEqual([2, 3, 4, 5, 6])
    expect(input).toEqual(before)
  })

  it.each([[2027, 1, 28], [2028, 1, 29], [2026, 3, 30], [2026, 11, 31]])('projects days onto the end of %s/%s', (year, month, last) => {
    const result = upcomingExpenses([expense(3, 31), expense(1, 30), expense(2, 29)], new Date(year, month, last))
    expect(result.map(({ expense }) => expense.id)).toEqual(last === 31 ? [3] : last === 30 ? [1, 3] : [1, 2, 3])
    for (const { dueDate } of result) {
      expect([dueDate.getFullYear(), dueDate.getMonth(), dueDate.getDate()]).toEqual([year, month, last])
    }
  })

  it('returns fewer than five and never rolls past expenses into next month', () => {
    expect(upcomingExpenses([expense(1, 1)], new Date(2026, 11, 31))).toEqual([])
    expect(upcomingExpenses([expense(1, 1)], new Date(2027, 0, 1))).toHaveLength(1)
    expect(upcomingExpenses([], new Date(2027, 0, 1))).toEqual([])
  })
})
