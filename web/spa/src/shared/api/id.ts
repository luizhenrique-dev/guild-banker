export function safeID(id: number): number {
  if (!Number.isSafeInteger(id) || id <= 0) throw new Error('Invalid API identifier')
  return id
}
