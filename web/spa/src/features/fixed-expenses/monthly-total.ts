import type { FixedExpenseResponse } from './types'

export function monthlyTotal(expenses: FixedExpenseResponse[], locale: string): string {
  const cents = expenses.reduce((sum, expense) => {
    if (expense.status !== 'ACTIVE') return sum
    const [units, fraction = ''] = expense.amount.split('.')
    return sum + BigInt(units!) * 100n + BigInt(fraction.padEnd(2, '0'))
  }, 0n)
  return new Intl.NumberFormat(locale, { style: 'currency', currency: 'BRL' })
    .formatToParts(cents / 100n)
    .map((part) => part.type === 'fraction' ? (cents % 100n).toString().padStart(2, '0') : part.value)
    .join('')
}
