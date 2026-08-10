import { NextResponse } from 'next/server'

export const dynamic = 'force-dynamic'

export async function GET() {
  return NextResponse.json([
    { id: 1, name: 'familia-silva', display_name: 'Família Silva', enabled: true },
  ])
}

export async function POST(request: Request) {
  const body = await request.json()
  return NextResponse.json({ id: 10, ...body, enabled: true }, { status: 201 })
}
