import { http, HttpResponse, delay } from 'msw'
import { mockUserID, mockGuildMember } from '../request-context'
import { addImportedTransactions } from './transactions'
import type { ImportItemResponse, UploadImportResponse, ConfirmImportResponse } from '@/src/features/imports/types'

// 📌 PADRÃO: Handlers simplificados para a feature de import
const initialImportItems: ImportItemResponse[] = [
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
    status: 'READY' as const,
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
    status: 'READY' as const,
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
    status: 'DUPLICATE' as const,
  },
]

const batches = new Map<string, UploadImportResponse>()
let nextImportID = 1
function batchKey(request: Request, guildID: unknown, importID: unknown) { return `${mockUserID(request)}:${guildID}:${importID}` }
const forbidden = () => HttpResponse.json({ error: 'forbidden' }, { status: 403 })
const missing = () => HttpResponse.json({ error: 'not found' }, { status: 404 })

export const importsHandlers = [
  http.post('*/api/v1/guilds/:guildID/imports', async ({ params, request }) => {
    if (!mockGuildMember(request, Number(params['guildID']))) return forbidden()
    const form = await request.formData()
    const file = form.get('file')
    if (!(file instanceof File) || file.size === 0 || file.size > 5 * 1024 * 1024) return HttpResponse.json({ error: 'invalid file size' }, { status: 400 })
    await delay(300)
    const batch: UploadImportResponse = {
      importId: nextImportID++, status: 'PENDING_REVIEW', fileName: file.name,
      summary: { parsed: 3, candidates: 2, duplicates: 1, skippedZero: 0 }, items: structuredClone(initialImportItems),
    }
    batches.set(batchKey(request, params['guildID'], batch.importId), batch)
    return HttpResponse.json(batch, { status: 201 })
  }),
  http.get('*/api/v1/guilds/:guildID/imports/:importID', ({ params, request }) => {
    if (!mockGuildMember(request, Number(params['guildID']))) return forbidden()
    const batch = batches.get(batchKey(request, params['guildID'], params['importID']))
    if (!batch) return missing()
    return HttpResponse.json({ importId: batch.importId, status: batch.status, fileName: batch.fileName, items: batch.items })
  }),
  http.patch('*/api/v1/guilds/:guildID/imports/:importID/items/:itemID', async ({ params, request }) => {
    if (!mockGuildMember(request, Number(params['guildID']))) return forbidden()
    const batch = batches.get(batchKey(request, params['guildID'], params['importID']))
    const item = batch?.items.find((item) => item.itemId === Number(params['itemID']))
    if (!item) return missing()
    Object.assign(item, await request.json(), { status: 'READY' })
    return HttpResponse.json(item)
  }),
  http.delete('*/api/v1/guilds/:guildID/imports/:importID/items/:itemID', ({ params, request }) => {
    if (!mockGuildMember(request, Number(params['guildID']))) return forbidden()
    const item = batches.get(batchKey(request, params['guildID'], params['importID']))?.items.find((item) => item.itemId === Number(params['itemID']))
    if (!item) return missing()
    item.status = 'DISCARDED'
    return new HttpResponse(null, { status: 204 })
  }),
  http.post('*/api/v1/guilds/:guildID/imports/:importID/confirm', async ({ params, request }) => {
    if (!mockGuildMember(request, Number(params['guildID']))) return forbidden()
    const batch = batches.get(batchKey(request, params['guildID'], params['importID']))
    if (!batch) return missing()
    if (batch.status !== 'PENDING_REVIEW') return HttpResponse.json({ error: 'batch is not pending review' }, { status: 400 })
    await delay(300)
    const created = addImportedTransactions(mockUserID(request), Number(params['guildID']), batch.items)
    batch.status = 'COMPLETED'
    return HttpResponse.json<ConfirmImportResponse>({ importId: batch.importId, created, skipped: batch.items.length - created, status: 'COMPLETED' })
  }),
]
export function resetImportsMocks() { batches.clear(); nextImportID = 1 }
