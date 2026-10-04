// @vitest-environment node
import { expect, it, vi } from 'vitest'
import { apiClient } from '../../shared/lib/api-client'
import type { components } from '../../shared/api/schema'

vi.stubGlobal('window', { location: { origin: 'http://localhost' } })

const headers = { 'X-User-ID': '1', 'X-User-Email': 'ana@familia.com' }
it('fixed-expense mock preserves name on PATCH and rejects ACTIVE deactivation', async () => {
  const updated = await apiClient<components['schemas']['FixedExpenseResponse']>('/api/v1/fixed-expenses/1', { method: 'PATCH', headers, body: { name: 'unsupported rename', amount: '100.50' } })
  expect(updated.name).toBe('Aluguel')
  expect(updated.amount).toBe('100.50')
  await expect(apiClient('/api/v1/fixed-expenses/1/deactivate', { method: 'PATCH', headers, body: { status: 'ACTIVE' } })).rejects.toMatchObject({ status: 400 })
  await apiClient('/api/v1/fixed-expenses/1/deactivate', { method: 'PATCH', headers, body: { status: 'PAUSED' } })
  const list = await apiClient<components['schemas']['FixedExpenseResponse'][]>('/api/v1/fixed-expenses', { headers })
  expect(list.some((expense) => expense.id === 1)).toBe(false)
})
