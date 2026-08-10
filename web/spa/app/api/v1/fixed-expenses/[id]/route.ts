import { NextResponse } from 'next/server'

export const dynamic = 'force-dynamic'

export async function PATCH(request: Request) {
  const body = await request.json()
  return NextResponse.json({ id: 1, ...body })
}
