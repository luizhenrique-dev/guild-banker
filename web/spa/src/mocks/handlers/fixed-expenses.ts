// ============================================================
// MSW Handlers: Fixed Expenses
// Espelham os endpoints da OpenAPI spec (tags: FixedExpenses)
// ============================================================
import { http, HttpResponse, delay } from 'msw'
import { fixedExpensesFixtures } from '../fixtures/fixed-expenses'
import type { FixedExpenseResponse } from '@/src/features/fixed-expenses/types'

// 📌 PADRÃO: Estado in-memory para que mutations persistam durante a sessão.
// Em produção, isso é o banco de dados.
let expenses = [...fixedExpensesFixtures]
let nextId = 100

export const fixedExpensesHandlers = [
  // --------------------------------------------------------
  // operationId: listActiveFixedExpenses
  // GET /api/v1/fixed-expenses
  // Produção: requer header X-User-ID para filtrar despesas do usuário logado
  // Retorna apenas despesas com status ACTIVE
  // --------------------------------------------------------
  http.get('/api/v1/fixed-expenses', async () => {
    await delay(600)
    const active = expenses.filter((e) => e.status === 'ACTIVE')
    return HttpResponse.json(active)
  }),

  // --------------------------------------------------------
  // operationId: createFixedExpense
  // POST /api/v1/fixed-expenses
  // Produção: requer X-User-ID e X-User-Email
  // Body: { name, amount, due_day, category }
  // --------------------------------------------------------
  http.post('/api/v1/fixed-expenses', async ({ request }) => {
    await delay(500)
    const body = (await request.json()) as {
      name: string
      amount: string
      due_day: number
      category: string
    }

    const newExpense: FixedExpenseResponse = {
      id: nextId++,
      name: body.name,
      amount: body.amount,
      due_day: body.due_day,
      category: body.category as FixedExpenseResponse['category'],
      status: 'ACTIVE',
    }

    expenses.push(newExpense)
    return HttpResponse.json(newExpense, { status: 201 })
  }),

  // --------------------------------------------------------
  // operationId: updateFixedExpense
  // PATCH /api/v1/fixed-expenses/:id
  // Produção: campos opcionais — só os enviados são atualizados
  // --------------------------------------------------------
  http.patch('/api/v1/fixed-expenses/:id', async ({ params, request }) => {
    await delay(400)
    const id = Number(params['id'])
    const body = (await request.json()) as Partial<FixedExpenseResponse>
    const index = expenses.findIndex((e) => e.id === id)

    if (index === -1) {
      return HttpResponse.json({ error: 'Despesa fixa não encontrada' }, { status: 404 })
    }

    const existing = expenses[index]
    if (!existing) {
      return HttpResponse.json({ error: 'Despesa fixa não encontrada' }, { status: 404 })
    }

    const updated: FixedExpenseResponse = { ...existing, ...body, id }
    expenses[index] = updated
    return HttpResponse.json(updated)
  }),

  // --------------------------------------------------------
  // operationId: deactivateFixedExpense
  // PATCH /api/v1/fixed-expenses/:id/deactivate
  // Body: { status: 'PAUSED' | 'CANCELLED' }
  // Produção: só aceita PAUSED ou CANCELLED; outros retornam 400
  // --------------------------------------------------------
  http.patch('/api/v1/fixed-expenses/:id/deactivate', async ({ params, request }) => {
    await delay(400)
    const id = Number(params['id'])
    const body = (await request.json()) as { status: string }

    if (body.status !== 'PAUSED' && body.status !== 'CANCELLED') {
      return HttpResponse.json(
        { error: 'Status inválido para desativação' },
        { status: 400 }
      )
    }

    const index = expenses.findIndex((e) => e.id === id)
    if (index === -1) {
      return HttpResponse.json({ error: 'Não encontrada' }, { status: 404 })
    }

    const existing = expenses[index]
    if (existing) {
      expenses[index] = { ...existing, status: body.status as FixedExpenseResponse['status'] }
    }

    return new HttpResponse(null, { status: 204 })
  }),
]
