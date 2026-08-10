// ============================================================
// 📌 PADRÃO: API Route Stub para desenvolvimento
// Quando MSW não está ativo (SSR, testes), este endpoint responde.
// 💡 REAL API: Em produção, remova este arquivo e aponte para o backend.
// ============================================================
import { NextResponse } from 'next/server'

export const dynamic = 'force-dynamic'

const fixtures = [
  { id: 1, name: 'Aluguel', amount: '2500.00', due_day: 10, category: 'HOUSING', status: 'ACTIVE' },
  { id: 2, name: 'Netflix', amount: '55.90', due_day: 15, category: 'SUBSCRIPTIONS', status: 'ACTIVE' },
  { id: 3, name: 'Plano de saúde', amount: '320.00', due_day: 5, category: 'INSURANCE', status: 'ACTIVE' },
  { id: 4, name: 'Seguro auto', amount: '180.00', due_day: 20, category: 'INSURANCE', status: 'ACTIVE' },
]

export async function GET() {
  return NextResponse.json(fixtures)
}

export async function POST(request: Request) {
  const body = await request.json()
  const newExpense = {
    id: Math.floor(Math.random() * 10000),
    name: body?.name ?? '',
    amount: body?.amount ?? '0.00',
    due_day: body?.due_day ?? 1,
    category: body?.category ?? 'OTHER',
    status: 'ACTIVE',
  }
  return NextResponse.json(newExpense, { status: 201 })
}
