import { NextResponse } from 'next/server'

export const dynamic = 'force-dynamic'

export async function PATCH() {
  return new NextResponse(null, { status: 204 })
}
