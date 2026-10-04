import { expect, it } from 'vitest'
import { safeID } from './id'

it.each([0, -1, 1.5, Number.MAX_SAFE_INTEGER + 1, NaN])('rejects unsafe identifier %s', (id) => {
  expect(() => safeID(id)).toThrow('Invalid API identifier')
})
it('preserves safe positive identifiers', () => {
  expect(safeID(Number.MAX_SAFE_INTEGER)).toBe(Number.MAX_SAFE_INTEGER)
})
