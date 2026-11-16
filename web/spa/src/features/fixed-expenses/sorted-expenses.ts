import { nextDueDate } from './next-due-date'
import type { FixedExpenseResponse } from './types'

export function sortedExpenses(expenses: FixedExpenseResponse[], reference: Date, locale: string): FixedExpenseResponse[] {
  const names = new Intl.Collator(locale, { sensitivity: 'base' })
  return expenses
    .map((expense) => ({ expense, dueDate: nextDueDate(expense.due_day, reference).getTime() }))
    .sort((a, b) => a.dueDate - b.dueDate || names.compare(a.expense.name, b.expense.name) || a.expense.id - b.expense.id)
    .map(({ expense }) => expense)
}
