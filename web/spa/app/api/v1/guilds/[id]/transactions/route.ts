import { NextResponse } from 'next/server'

export const dynamic = 'force-dynamic'

const txFixtures = [
  { id: 1, type: 'EXPENSE', description: 'Supermercado Extra', amount: '347.82', category: 'GROCERY', status: 'ACTIVE', source: 'MANUAL', visibility: 'PUBLIC', occurredAt: '2026-07-20T10:30:00Z', guildId: 1, userAccountId: 1, createdAt: '2026-07-20T10:35:00Z' },
  { id: 2, type: 'EXPENSE', description: 'Conta de luz', amount: '215.40', category: 'UTILITIES', status: 'ACTIVE', source: 'MANUAL', visibility: 'PUBLIC', occurredAt: '2026-07-18T14:00:00Z', guildId: 1, userAccountId: 1, createdAt: '2026-07-18T14:05:00Z' },
  { id: 3, type: 'INCOME', description: 'Salário — julho', amount: '8500.00', category: 'OTHER', status: 'ACTIVE', source: 'MANUAL', visibility: 'PRIVATE', occurredAt: '2026-07-05T08:00:00Z', guildId: 1, userAccountId: 1, createdAt: '2026-07-05T08:01:00Z' },
]

export async function GET() {
  return NextResponse.json({ items: txFixtures })
}

export async function POST(request: Request) {
  const body = await request.json()
  const tx = { id: 999, ...body, status: 'ACTIVE', source: 'MANUAL', guildId: 1, userAccountId: 1, createdAt: new Date().toISOString() }
  return NextResponse.json(tx, { status: 201 })
}
