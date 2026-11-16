import { differenceInCalendarDays } from 'date-fns'
import type { FixedExpenseResponse } from './types'
import { nextDueDate } from './next-due-date'

export function dueNotice(expense: FixedExpenseResponse, reference: Date): 'today' | 'tomorrow' | 'soon' | null {
  if (expense.status !== 'ACTIVE') return null

  const days = differenceInCalendarDays(nextDueDate(expense.due_day, reference), reference)
  if (days === 0) return 'today'
  if (days === 1) return 'tomorrow'
  return days <= 5 ? 'soon' : null
}
