export function mockUserID(request: Request): number {
  const id = Number(request.headers.get('X-User-ID'))
  if (Number.isSafeInteger(id) && id > 0) {
    const email = request.headers.get('X-User-Email')?.trim().toLowerCase()
    if (email) emails.set(id, email)
    return id
  }
  return 0
}

let emails = new Map([[1, 'ana@familia.com'], [2, 'b@example.test']])
let memberships = initialMemberships()
function initialMemberships() { return new Map<number, Set<number>>([[1, new Set([1, 2])], [2, new Set([2])]]) }
export function mockGuildMember(request: Request, guildID: number): boolean {
  return memberships.get(mockUserID(request))?.has(guildID) ?? false
}
export function addMockGuildMember(userID: number, guildID: number) {
  const guilds = memberships.get(userID) ?? new Set<number>()
  guilds.add(guildID)
  memberships.set(userID, guilds)
}
export function mockGuildMemberByEmail(email: string, guildID: number) {
  return [...emails].some(([id, knownEmail]) => knownEmail === email && memberships.get(id)?.has(guildID))
}
export function resetMockMemberships() {
  memberships = initialMemberships()
  emails = new Map([[1, 'ana@familia.com'], [2, 'b@example.test']])
}
