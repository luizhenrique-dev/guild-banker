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
