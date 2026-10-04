import { expect, it } from 'vitest'
import { importSchema } from './import.schema'
import { inviteSchema } from '../../guilds/schemas/invite.schema'

it('accepts CSV up to 5 MiB and rejects empty, other formats and oversized files', () => {
  for (const file of [new File([], 'empty.csv'), new File(['csv'], 'test.pdf'), new File([new Uint8Array(5 * 1024 * 1024 + 1)], 'large.csv')]) {
    expect(importSchema.safeParse({ file }).success).toBe(false)
  }
  expect(importSchema.safeParse({ file: new File([new Uint8Array(5 * 1024 * 1024)], 'test.CSV') }).success).toBe(true)
})
it('validates and trims invitation emails', () => {
  expect(inviteSchema.safeParse({ email: 'invalid' }).success).toBe(false)
  expect(inviteSchema.parse({ email: '  test@example.com  ' })).toEqual({ email: 'test@example.com' })
})
