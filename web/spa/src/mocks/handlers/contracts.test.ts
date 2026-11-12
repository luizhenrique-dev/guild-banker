// @vitest-environment node
import { expect, it, vi } from 'vitest'
import { apiClient } from '../../shared/lib/api-client'
import type { components } from '../../shared/api/schema'

vi.stubGlobal('window', { location: { origin: 'http://localhost' } })

const headers = { 'X-User-ID': '1', 'X-User-Email': 'ana@familia.com' }
it('fixed-expense mock preserves name on PATCH and rejects ACTIVE deactivation', async () => {
  const updated = await apiClient<components['schemas']['FixedExpenseResponse']>('/api/v1/guilds/1/fixed-expenses/1', { method: 'PATCH', headers, body: { name: 'unsupported rename', amount: '100.50' } })
  expect(updated.name).toBe('Aluguel')
  expect(updated.amount).toBe('100.50')
  await expect(apiClient('/api/v1/guilds/1/fixed-expenses/1/deactivate', { method: 'PATCH', headers, body: { status: 'ACTIVE' } })).rejects.toMatchObject({ status: 400 })
  await apiClient('/api/v1/guilds/1/fixed-expenses/1/deactivate', { method: 'PATCH', headers, body: { status: 'PAUSED' } })
  const list = await apiClient<components['schemas']['FixedExpenseResponse'][]>('/api/v1/guilds/1/fixed-expenses', { headers })
  expect(list.some((expense) => expense.id === 1)).toBe(false)
})

it('members share one guild record while the other guild stays isolated', async () => {
  const root = '/api/v1/guilds/2/fixed-expenses'
  const memberB = { 'X-User-ID': '2', 'X-User-Email': 'b@example.test' }
  const created = await apiClient<components['schemas']['FixedExpenseResponse']>(root, { method: 'POST', headers, body: { name: 'Shared test', amount: '12.34', due_day: 3, category: 'OTHER', guild_id: 1 } })
  expect(created.guild_id).toBe(2)
  const list = await apiClient<components['schemas']['FixedExpenseResponse'][]>(root, { headers: memberB })
  expect(list.find((expense) => expense.id === created.id)).toEqual(created)
  const updated = await apiClient<components['schemas']['FixedExpenseResponse']>(root + '/' + created.id, { method: 'PATCH', headers: memberB, body: { name: 'ignored', guild_id: 1, amount: '20.50' } })
  expect(updated).toMatchObject({ name: 'Shared test', guild_id: 2, amount: '20.50' })
  await expect(apiClient('/api/v1/guilds/1/fixed-expenses/' + created.id, { method: 'PATCH', headers, body: { amount: '999' } })).rejects.toMatchObject({ status: 404 })
  await apiClient(root + '/' + created.id + '/deactivate', { method: 'PATCH', headers: memberB, body: { status: 'CANCELLED' } })
  expect(await apiClient<components['schemas']['FixedExpenseResponse'][]>(root, { headers })).not.toContainEqual(expect.objectContaining({ id: created.id }))
})

it.each(['GET', 'POST', 'PATCH', 'deactivate'])('denies non member %s in demo', async (operation) => {
  const method = operation === 'deactivate' ? 'PATCH' : operation
  const path = '/api/v1/guilds/1/fixed-expenses' + (method === 'PATCH' ? '/1' : '') + (operation === 'deactivate' ? '/deactivate' : '')
  await expect(apiClient(path, { method, headers: { 'X-User-ID': '2', 'X-User-Email': 'b@example.test' }, ...(method === 'GET' ? {} : { body: { status: 'PAUSED' } }) })).rejects.toMatchObject({ status: 403 })
})

it('guild mock generates unique names, retains creator membership and edits only mutable fields', async () => {
  type Guild = components['schemas']['GuildResponse']
  const create = () => apiClient<Guild>('/api/v1/guilds', { method: 'POST', headers, body: { name: 'ignored', display_name: 'Á&B  Amigos', description: 'Casa' } })
  const first = await create()
  const second = await create()
  expect(first).toMatchObject({ name: 'ab-amigos', display_name: 'Á&B  Amigos', description: 'Casa' })
  expect(second.name).toMatch(/^ab-amigos-[a-f0-9]{8}$/)
  expect(await apiClient<Guild[]>('/api/v1/guilds', { headers })).toContainEqual(first)
  const updated = await apiClient<Guild>('/api/v1/guilds/' + first.id, { method: 'PUT', headers, body: { name: 'ignored-again', display_name: 'Novo visual' } })
  expect(updated).toMatchObject({ name: first.name, description: 'Casa' })
  const cleared = await apiClient<Guild>('/api/v1/guilds/' + first.id, { method: 'PUT', headers, body: { display_name: 'Novo visual', description: '' } })
  expect(cleared.description).toBe('')
  await expect(apiClient('/api/v1/guilds', { method: 'POST', headers, body: { display_name: 'á'.repeat(71) } })).rejects.toMatchObject({ status: 400 })
})

it('inviting only creates a pending invitation and distinguishes duplicate from existing member', async () => {
  await apiClient('/api/v1/guilds/1/invites', { method: 'POST', headers, body: { email: 'new@example.test' } })
  await expect(apiClient('/api/v1/guilds/1/invites', { method: 'POST', headers, body: { email: 'NEW@example.test' } })).rejects.toMatchObject({ status: 409, code: 'invitation_already_pending' })
  await expect(apiClient('/api/v1/guilds/1/invites', { method: 'POST', headers, body: { email: 'ana@familia.com' } })).rejects.toMatchObject({ status: 409, code: 'guild_member_already_set' })
  const recipient = { 'X-User-ID': '8', 'X-User-Email': 'new@example.test' }
  expect(await apiClient('/api/v1/guilds', { headers: recipient })).toEqual([])
})

it('invitation mock authorizes the recipient and grants access only after acceptance', async () => {
  type Invitation = components['schemas']['InvitationResponse']
  const recipient = { 'X-User-ID': '8', 'X-User-Email': 'new@example.test' }
  const emit = () => apiClient('/api/v1/guilds/1/invites', { method: 'POST', headers, body: { email: 'new@example.test' } })
  await emit()
  const pending = await apiClient<Invitation[]>('/api/v1/guild-invites', { headers: recipient })
  expect(pending).toHaveLength(1)
  expect(pending[0]?.invited_by).toEqual({ name: 'Ana Silva', email: headers['X-User-Email'] })
  const root = '/api/v1/guild-invites/' + pending[0]!.id
  await expect(apiClient(root + '/accept', { method: 'POST', headers })).rejects.toMatchObject({ status: 404 })
  await apiClient(root + '/decline', { method: 'POST', headers: recipient })
  expect(await apiClient('/api/v1/guilds', { headers: recipient })).toEqual([])
  await emit()
  const again = await apiClient<Invitation[]>('/api/v1/guild-invites', { headers: recipient })
  await apiClient('/api/v1/guild-invites/' + again[0]!.id + '/accept', { method: 'POST', headers: recipient })
  expect(await apiClient('/api/v1/guilds', { headers: recipient })).toEqual([expect.objectContaining({ id: 1 })])
  expect(await apiClient('/api/v1/guild-invites', { headers: recipient })).toEqual([])
})
