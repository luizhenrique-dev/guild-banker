import { expect, it } from 'vitest'
import { inviteSchema } from './invite.schema'

it.each([
  'member@example.test',
  ' MEMBER@EXAMPLE.TEST ',
  'Member.Name+bank@My-Group.Example.test',
])('normalizes an invitation email %s', (email) => {
  expect(inviteSchema.parse({ email }).email).toBe(email.trim().toLowerCase())
})

it.each([
  '"member"@example.test', '"a b"@example.test', '"a@b"@example.test',
  'Member <member@example.test>', 'member..name@example.test', '.member@example.test',
  'member.@example.test', 'member@example', 'member@-example.test',
  'member@example-.test', 'member@exam_ple.test',
])('rejects unsupported invitation email %s', (email) => {
  expect(inviteSchema.safeParse({ email }).success).toBe(false)
})
