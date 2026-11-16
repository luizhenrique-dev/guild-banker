import { describe, expect, it } from 'vitest'
import { dueNotice } from './due-notice'
import type { FixedExpenseResponse } from './types'

const expense: FixedExpenseResponse = {
  id: 1, guild_id: 1, name: 'Internet', amount: '100.00', due_day: 15,
  category: 'OTHER', payment_method: 'CREDIT_CARD', responsible_user_id: null, status: 'ACTIVE',
}

describe('dueNotice', () => {
  it.each([
    [14, null], [15, 'today'], [16, 'tomorrow'], [17, 'soon'],
    [18, 'soon'], [19, 'soon'], [20, 'soon'], [21, null],
  ])('classifies due day %s relative to the 15th', (day, expected) => {
    expect(dueNotice({ ...expense, due_day: day }, new Date(2026, 10, 15, 23, 59))).toBe(expected)
  })

  it.each(['PAUSED', 'CANCELLED'] as const)('hides notices for %s expenses', (status) => {
    expect(dueNotice({ ...expense, status }, new Date(2026, 10, 15))).toBeNull()
  })

  it.each([
    [2026, 10, 30, 1, 'tomorrow'],
    [2026, 11, 31, 1, 'tomorrow'],
    [2026, 11, 31, 5, 'soon'],
    [2026, 11, 31, 6, null],
    [2027, 1, 27, 31, 'tomorrow'],
    [2027, 1, 28, 31, 'today'],
    [2028, 1, 28, 31, 'tomorrow'],
    [2028, 1, 29, 31, 'today'],
    [2026, 3, 30, 31, 'today'],
    [2027, 0, 31, 5, 'soon'],
  ])('handles month ends: %s/%s/%s, due day %s', (year, month, day, dueDay, expected) => {
    expect(dueNotice({ ...expense, due_day: dueDay }, new Date(year, month, day, 23, 59))).toBe(expected)
  })

  it.each([[2026, 2, 8, 9], [2026, 10, 1, 2]])('counts calendar days across daylight saving transitions', (year, month, day, dueDay) => {
    expect(dueNotice({ ...expense, due_day: dueDay }, new Date(year, month, day, 0, 1))).toBe('tomorrow')
  })
})
