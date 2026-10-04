import { expect, it } from 'vitest'
import { createFixedExpenseSchema, updateFixedExpenseSchema } from './fixed-expense.schema'

const valid = { name: ' Internet ', amount: '123,45', due_day: 1, category: 'OTHER' }
it.each(['0', '0.00', '00,00', '-1', '1.234', '1,234', '', '1e2', '1.2.3'])('rejects invalid or nonpositive amount %s', (amount) => {
  expect(createFixedExpenseSchema.safeParse({ ...valid, amount }).success).toBe(false)
})
it.each(['0.01', '10', '123,45', '123.45'])('normalizes a positive decimal %s as a string', (amount) => {
  const parsed = createFixedExpenseSchema.parse({ ...valid, amount })
  expect(parsed.amount).toBe(amount.replace(',', '.'))
  expect(parsed.name).toBe('Internet')
})
it.each([0, 32, 1.5, NaN])('rejects invalid due day %s', (due_day) => {
  expect(createFixedExpenseSchema.safeParse({ ...valid, due_day }).success).toBe(false)
})
it.each([1, 31])('accepts boundary day %s', (due_day) => {
  expect(createFixedExpenseSchema.safeParse({ ...valid, due_day }).success).toBe(true)
})
it.each([{ category: 'INVALID' }, { name: '' }, { name: '  ' }, { name: 'x'.repeat(101) }])('rejects invalid category/name %j', (input) => {
  expect(createFixedExpenseSchema.safeParse({ ...valid, ...input }).success).toBe(false)
})
it('edit schema strips name and only validates supported fields', () => {
  expect(updateFixedExpenseSchema.parse({ ...valid, name: '' })).toEqual({ amount: '123.45', due_day: 1, category: 'OTHER' })
})
