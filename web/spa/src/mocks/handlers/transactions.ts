import { http, HttpResponse, delay } from 'msw'
import { transactionsFixtures } from '../fixtures/transactions'
import type { TransactionResponse } from '@/src/features/transactions/types'

let transactions = [...transactionsFixtures]
let nextTxId = 200

export const transactionsHandlers = [
  // operationId: listTransactions
  http.get('/api/v1/guilds/:guildID/transactions', async ({ request }) => {
    await delay(600)
    const url = new URL(request.url)
    const limit = Number(url.searchParams.get('limit') ?? '10')
    const cursor = url.searchParams.get('cursor')
    const category = url.searchParams.get('category')
    const type = url.searchParams.get('type')

    let filtered = transactions.filter((t) => t.status !== 'CANCELLED')
    if (category) filtered = filtered.filter((t) => t.category === category)
    if (type) filtered = filtered.filter((t) => t.type === type)

    // 📌 PADRÃO: Paginação por cursor simples (offset-based para o mock)
    const startIndex = cursor ? Number(atob(cursor)) : 0
    const page = filtered.slice(startIndex, startIndex + limit)
    const hasMore = startIndex + limit < filtered.length

    return HttpResponse.json({
      items: page,
      ...(hasMore ? { nextCursor: btoa(String(startIndex + limit)) } : {}),
    })
  }),

  // operationId: createTransaction
  http.post('/api/v1/guilds/:guildID/transactions', async ({ request }) => {
    await delay(500)
    const body = (await request.json()) as {
      type: string
      description: string
      amount: string
      category: string
      visibility: string
      occurredAt: string
    }
    const newTx: TransactionResponse = {
      id: nextTxId++,
      type: body.type as TransactionResponse['type'],
      description: body.description,
      amount: body.amount,
      category: body.category as TransactionResponse['category'],
      status: 'ACTIVE',
      source: 'MANUAL',
      visibility: body.visibility as TransactionResponse['visibility'],
      occurredAt: body.occurredAt,
      guildId: 1,
      userAccountId: 1,
      createdAt: new Date().toISOString(),
    }
    transactions.unshift(newTx)
    return HttpResponse.json(newTx, { status: 201 })
  }),

  // operationId: deleteTransaction
  http.delete('/api/v1/guilds/:guildID/transactions/:id', async ({ params }) => {
    await delay(400)
    const id = Number(params['id'])
    const tx = transactions.find((t) => t.id === id)
    if (!tx) return HttpResponse.json({ error: 'Não encontrada' }, { status: 404 })
    if (tx.status === 'CANCELLED') return HttpResponse.json({ error: 'Já cancelada' }, { status: 422 })
    tx.status = 'CANCELLED'
    return new HttpResponse(null, { status: 204 })
  }),

  // operationId: setTransactionVisibility
  http.patch('/api/v1/guilds/:guildID/transactions/:id/visibility', async ({ params, request }) => {
    await delay(300)
    const id = Number(params['id'])
    const body = (await request.json()) as { visibility: string }
    const tx = transactions.find((t) => t.id === id)
    if (!tx) return HttpResponse.json({ error: 'Não encontrada' }, { status: 404 })
    tx.visibility = body.visibility as TransactionResponse['visibility']
    return HttpResponse.json({
      id: tx.id,
      visibility: tx.visibility,
      updatedAt: new Date().toISOString(),
    })
  }),

  // operationId: updateTransaction
  http.patch('/api/v1/guilds/:guildID/transactions/:id', async ({ params, request }) => {
    await delay(400)
    const id = Number(params['id'])
    const body = (await request.json()) as Partial<TransactionResponse>
    const tx = transactions.find((t) => t.id === id)
    if (!tx) return HttpResponse.json({ error: 'Não encontrada' }, { status: 404 })
    Object.assign(tx, body)
    return HttpResponse.json(tx)
  }),
]
