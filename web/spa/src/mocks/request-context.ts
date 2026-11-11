export function mockUserID(request: Request): number {
  const id = Number(request.headers.get('X-User-ID'))
  return Number.isSafeInteger(id) && id > 0 ? id : 0
}

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
export function resetMockMemberships() { memberships = initialMemberships() }
