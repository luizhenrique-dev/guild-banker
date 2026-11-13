import { http, HttpResponse, delay } from 'msw'
import { mockGuildMember } from '../request-context'
import { fixedExpensesFixtures, sharedFixedExpensesFixtures } from '../fixtures/fixed-expenses'
import type { FixedExpenseResponse } from '@/src/features/fixed-expenses/types'

const expensesByGuild = new Map<number, FixedExpenseResponse[]>()
function guildExpenses(guildID: number) {
  if (!expensesByGuild.has(guildID)) expensesByGuild.set(guildID, structuredClone(guildID === 1 ? fixedExpensesFixtures : guildID === 2 ? sharedFixedExpensesFixtures : []))
  return expensesByGuild.get(guildID)!
}
const forbidden = () => HttpResponse.json({ error: 'forbidden' }, { status: 403 })
const notFound = () => HttpResponse.json({ error: 'fixed expense not found' }, { status: 404 })
const root = '*/api/v1/guilds/:guildID/fixed-expenses'
let nextId = 100

export const fixedExpensesHandlers = [
  // operationId: listActiveFixedExpenses
  http.get(root, async ({ request, params }) => {
    const guildID = Number(params['guildID'])
    if (!mockGuildMember(request, guildID)) return forbidden()
    await delay(600)
    return HttpResponse.json(guildExpenses(guildID).filter((expense) => expense.status === 'ACTIVE' || new URL(request.url).searchParams.get('include_paused') === 'true' && expense.status === 'PAUSED').sort((a, b) => a.due_day - b.due_day || a.name.localeCompare(b.name)))
  }),
  // operationId: createFixedExpense
  http.post(root, async ({ request, params }) => {
    const guildID = Number(params['guildID'])
    if (!mockGuildMember(request, guildID)) return forbidden()
    await delay(500)
    const body = await request.json() as FixedExpenseResponse
    const expense: FixedExpenseResponse = { id: nextId++, guild_id: guildID, name: body.name, amount: body.amount, due_day: body.due_day, category: body.category, status: 'ACTIVE' }
    guildExpenses(guildID).push(expense)
    return HttpResponse.json(expense, { status: 201 })
  }),
  // operationId: updateFixedExpense
  http.patch(root + '/:id', async ({ request, params }) => {
    const guildID = Number(params['guildID'])
    if (!mockGuildMember(request, guildID)) return forbidden()
    await delay(400)
    const expenses = guildExpenses(guildID)
    const index = expenses.findIndex((expense) => expense.id === Number(params['id']))
    const existing = expenses[index]
    if (!existing) return notFound()
    const body = await request.json() as Partial<FixedExpenseResponse>
    const updated = { ...existing, amount: body.amount ?? existing.amount, due_day: body.due_day ?? existing.due_day, category: body.category ?? existing.category, status: body.status ?? existing.status }
    expenses[index] = updated
    return HttpResponse.json(updated)
  }),
  // operationId: deleteFixedExpense
  http.delete(root + '/:id', async ({ request, params }) => {
    const guildID = Number(params['guildID'])
    if (!mockGuildMember(request, guildID)) return forbidden()
    await delay(400)
    const expenses = guildExpenses(guildID)
    const index = expenses.findIndex((expense) => expense.id === Number(params['id']))
    if (index < 0) return notFound()
    expenses.splice(index, 1)
    return new HttpResponse(null, { status: 204 })
  }),
  // operationId: deactivateFixedExpense
  http.patch(root + '/:id/deactivate', async ({ request, params }) => {
    const guildID = Number(params['guildID'])
    if (!mockGuildMember(request, guildID)) return forbidden()
    await delay(400)
    const body = await request.json() as { status: string }
    if (body.status !== 'PAUSED' && body.status !== 'CANCELLED') return HttpResponse.json({ error: 'invalid fixed expense status' }, { status: 400 })
    const expense = guildExpenses(guildID).find((row) => row.id === Number(params['id']))
    if (!expense) return notFound()
    expense.status = body.status
    return new HttpResponse(null, { status: 204 })
  }),
]

export function resetFixedExpensesMocks() {
  expensesByGuild.clear()
  nextId = 100
}
