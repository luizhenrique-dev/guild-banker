import { NextResponse } from 'next/server'
export const dynamic = 'force-dynamic'
export async function DELETE() {
  return new NextResponse(null, { status: 204 })
}
