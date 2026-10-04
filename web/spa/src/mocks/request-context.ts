export function mockUserID(request: Request): number {
  const id = Number(request.headers.get('X-User-ID'))
  return Number.isSafeInteger(id) && id > 0 ? id : 0
}

export function mockGuildMember(request: Request, guildID: number): boolean {
  const userID = mockUserID(request)
  return userID === 1 && (guildID === 1 || guildID === 2) || userID === 2 && guildID === 2
}
