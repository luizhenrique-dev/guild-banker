import type { FixedExpenseResponse } from '../fixed-expenses/types'

export function upcomingExpenses(expenses: FixedExpenseResponse[], reference: Date) {
  const year = reference.getFullYear()
  const month = reference.getMonth()
  const lastDay = new Date(year, month + 1, 0).getDate()
  return expenses
    .filter((expense) => expense.status === 'ACTIVE')
    .map((expense) => ({ expense, dueDate: new Date(year, month, Math.min(expense.due_day, lastDay)) }))
    .filter(({ dueDate }) => dueDate.getDate() >= reference.getDate())
    .sort((a, b) => a.dueDate.getTime() - b.dueDate.getTime() || a.expense.id - b.expense.id)
    .slice(0, 5)
}
