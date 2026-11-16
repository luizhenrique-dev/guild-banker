import { describe, expect, it } from 'vitest'
import { sortedExpenses } from './sorted-expenses'
import type { FixedExpenseResponse } from './types'

function expense(id: number, dueDay: number, name: string, status: FixedExpenseResponse['status'] = 'ACTIVE'): FixedExpenseResponse {
  return { id, due_day: dueDay, name, status, guild_id: 1, amount: '100', payment_method: 'PIX', category: 'OTHER', responsible_user_id: null }
}

describe('sortedExpenses', () => {
  it('starts with today and mixes active and paused expenses without mutating the input', () => {
    const input = [expense(1, 1, 'Next month'), expense(2, 14, 'Past day'), expense(3, 16, 'Tomorrow', 'PAUSED'), expense(4, 15, 'Today'), expense(5, 31, 'Month end')]
    const before = structuredClone(input)
    expect(sortedExpenses(input, new Date(2026, 10, 15, 23, 59), 'pt-BR').map(({ id }) => id)).toEqual([4, 3, 5, 1, 2])
    expect(input).toEqual(before)
  })

  it.each(['pt-BR', 'en'])('sorts tied dates alphabetically in %s, ignoring case and accents, then by ID', (locale) => {
    const input = [expense(1, 15, 'Zebra'), expense(4, 15, 'água'), expense(3, 15, 'Água'), expense(2, 15, 'energia')]
    expect(sortedExpenses(input, new Date(2026, 10, 15), locale).map(({ id }) => id)).toEqual([3, 4, 2, 1])
    expect(sortedExpenses([input[0]!, input[3]!], new Date(2026, 10, 15), locale).map(({ id }) => id)).toEqual([2, 1])
  })

  it.each([[2027, 1, 28], [2028, 1, 29], [2026, 3, 30]])('uses effective dates for months with fewer days: %s/%s', (year, month, lastDay) => {
    const input = [expense(1, lastDay, 'Zebra'), expense(2, 31, 'Água'), expense(3, 1, 'Tomorrow')]
    expect(sortedExpenses(input, new Date(year, month, lastDay), 'pt-BR').map(({ id }) => id)).toEqual([2, 1, 3])
  })

  it('rolls over the year and supports an empty list', () => {
    expect(sortedExpenses([expense(1, 1, 'January'), expense(2, 31, 'December')], new Date(2026, 11, 31), 'pt-BR').map(({ id }) => id)).toEqual([2, 1])
    expect(sortedExpenses([], new Date(), 'pt-BR')).toEqual([])
  })
})
