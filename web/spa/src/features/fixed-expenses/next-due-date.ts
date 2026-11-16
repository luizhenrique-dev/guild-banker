import { differenceInCalendarDays } from 'date-fns'

export function nextDueDate(dueDay: number, reference: Date): Date {
  const year = reference.getFullYear()
  const month = reference.getMonth()
  const lastDay = new Date(year, month + 1, 0).getDate()
  const dueDate = new Date(year, month, Math.min(dueDay, lastDay))
  if (differenceInCalendarDays(dueDate, reference) >= 0) return dueDate

  const nextLastDay = new Date(year, month + 2, 0).getDate()
  return new Date(year, month + 1, Math.min(dueDay, nextLastDay))
}
