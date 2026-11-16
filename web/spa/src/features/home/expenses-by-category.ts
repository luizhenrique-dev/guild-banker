import type { FixedExpenseCategory, FixedExpenseResponse } from '../fixed-expenses/types'

export function expensesByCategory(expenses: FixedExpenseResponse[]) {
  const groups = new Map<FixedExpenseCategory, { category: FixedExpenseCategory; expenses: FixedExpenseResponse[]; cents: bigint }>()
  for (const expense of expenses) {
    if (expense.status !== 'ACTIVE') continue
    const group = groups.get(expense.category) ?? { category: expense.category, expenses: [], cents: 0n }
    const [units, fraction = ''] = expense.amount.split('.')
    group.cents += BigInt(units!) * 100n + BigInt(fraction.padEnd(2, '0'))
    group.expenses.push(expense)
    groups.set(expense.category, group)
  }
  return [...groups.values()].sort((a, b) => a.cents === b.cents
    ? a.category.localeCompare(b.category)
    : a.cents > b.cents ? -1 : 1)
}
