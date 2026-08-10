import { http, HttpResponse, delay } from 'msw'

// 📌 PADRÃO: Handlers simplificados para a feature de import
let mockImportItems = [
  {
    itemId: 1,
    occurredAt: '2026-07-15T00:00:00Z',
    description: 'MERCADO LIVRE *COMPRA',
    amount: '149.90',
    type: 'EXPENSE' as const,
    category: 'SHOPPING' as const,
    bankCategory: 'Compras',
    cardLast4: '4321',
    installment: 'Única',
    status: 'READY' as const as string,
  },
  {
    itemId: 2,
    occurredAt: '2026-07-14T00:00:00Z',
    description: 'IFOOD *RESTAURANTE',
    amount: '45.80',
    type: 'EXPENSE' as const,
    category: 'FOOD_AND_DINING' as const,
    bankCategory: 'Alimentação',
    cardLast4: '4321',
    installment: 'Única',
    status: 'READY' as const as string,
  },
  {
    itemId: 3,
    occurredAt: '2026-07-13T00:00:00Z',
    description: 'UBER *TRIP',
    amount: '32.50',
    type: 'EXPENSE' as const,
    category: 'TRANSPORTATION' as const,
    bankCategory: 'Transporte',
    cardLast4: '4321',
    installment: 'Única',
    status: 'DUPLICATE' as const as string,
  },
]

export const importsHandlers = [
  // operationId: uploadImport
  http.post('/api/v1/guilds/:guildID/imports', async () => {
    await delay(800)
    return HttpResponse.json(
      {
        importId: 1,
        status: 'PENDING_REVIEW',
        fileName: 'fatura-c6-julho.csv',
        summary: { parsed: 3, candidates: 2, duplicates: 1, skippedZero: 0 },
        items: mockImportItems,
      },
      { status: 201 }
    )
  }),

  // operationId: getImportById
  http.get('/api/v1/guilds/:guildID/imports/:importID', async () => {
    await delay(500)
    return HttpResponse.json({
      importId: 1,
      status: 'PENDING_REVIEW',
      fileName: 'fatura-c6-julho.csv',
      items: mockImportItems,
    })
  }),

  // operationId: updateImportItem
  http.patch('/api/v1/guilds/:guildID/imports/:importID/items/:itemID', async ({ params, request }) => {
    await delay(300)
    const itemId = Number(params['itemID'])
    const body = (await request.json()) as Record<string, unknown>
    const item = mockImportItems.find((i) => i.itemId === itemId)
    if (!item) return HttpResponse.json({ error: 'Não encontrado' }, { status: 404 })
    Object.assign(item, body)
    return HttpResponse.json(item)
  }),

  // operationId: discardImportItem
  http.delete('/api/v1/guilds/:guildID/imports/:importID/items/:itemID', async ({ params }) => {
    await delay(300)
    const itemId = Number(params['itemID'])
    const idx = mockImportItems.findIndex((i) => i.itemId === itemId)
    if (idx !== -1) {
      const item = mockImportItems[idx]
      if (item) (item as { status: string }).status = 'DISCARDED'
    }
    return new HttpResponse(null, { status: 204 })
  }),

  // operationId: confirmImport
  http.post('/api/v1/guilds/:guildID/imports/:importID\\:confirm', async () => {
    await delay(600)
    return HttpResponse.json({
      importId: 1,
      created: 2,
      skipped: 1,
      status: 'COMPLETED',
    })
  }),
]
