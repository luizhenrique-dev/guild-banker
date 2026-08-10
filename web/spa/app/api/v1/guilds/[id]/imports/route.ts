import { NextResponse } from 'next/server'
export const dynamic = 'force-dynamic'
export async function POST() {
  return NextResponse.json({
    importId: 1, status: 'PENDING_REVIEW', fileName: 'fatura.csv',
    summary: { parsed: 3, candidates: 2, duplicates: 1, skippedZero: 0 },
    items: [
      { itemId: 1, occurredAt: '2026-07-15T00:00:00Z', description: 'COMPRA', amount: '149.90', type: 'EXPENSE', category: 'SHOPPING', bankCategory: 'Compras', cardLast4: '4321', installment: 'Única', status: 'READY' },
    ],
  }, { status: 201 })
}
