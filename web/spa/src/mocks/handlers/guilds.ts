import { http, HttpResponse, delay } from 'msw'
import { mockUserID, mockGuildMember, addMockGuildMember, resetMockMemberships, mockGuildMemberByEmail } from '../request-context'
import { guildsFixtures } from '../fixtures/guilds'
import type { Guild } from '@/src/features/guilds/types'
import { inviteSchema } from '@/src/features/guilds/schemas/invite.schema'

let guilds = structuredClone(guildsFixtures)
let invitations: { id: number; guildID: number; email: string; invitedBy: number; inviterEmail: string; status: string }[] = []

export const guildsHandlers = [
  // operationId: listGuildsByMember
  http.get('*/api/v1/guilds', async ({ request }) => {
    await delay(500)
    return HttpResponse.json(guilds.filter((guild) => mockGuildMember(request, guild.id)))
  }),

  // operationId: createGuild
  http.post('*/api/v1/guilds', async ({ request }) => {
    await delay(500)
    const body = (await request.json()) as { display_name: string; description?: string }
    const displayName = body.display_name?.trim() ?? ''
    if (!displayName || [...displayName].length > 70) return HttpResponse.json({ error: 'invalid display name' }, { status: 400 })
    const base = displayName.normalize('NFD').replace(/\p{M}/gu, '').toLowerCase().replace(/\s+/g, '-').replace(/[^a-z0-9-]/g, '').replace(/-+/g, '-').replace(/^-|-$/g, '') || 'guild'
    let name = base
    for (let attempt = 0; guilds.some((guild) => guild.name === name); attempt++) {
      if (attempt === 5) return HttpResponse.json({ error: 'guild name generation temporarily unavailable' }, { status: 503 })
      name = base.slice(0,61).replace(/-+$/, '') + '-' + crypto.randomUUID().replaceAll('-', '').slice(0,8)
    }
    const newGuild: Guild = {
      id: guilds.length + 10,
      name,
      display_name: displayName,
      description: body.description ?? "",
      enabled: true,
    }
    guilds.push(newGuild)
    addMockGuildMember(mockUserID(request), newGuild.id)
    return HttpResponse.json(newGuild, { status: 201 })
  }),

  // operationId: updateGuildName
  http.put('*/api/v1/guilds/:id', async ({ params, request }) => {
    await delay(400)
    const id = Number(params['id'])
    const body = (await request.json()) as { display_name: string; description?: string }
    const guild = guilds.find((g) => g.id === id)
    if (!guild) return HttpResponse.json({ error: 'Não encontrado' }, { status: 404 })
    if (!body.display_name?.trim() || [...body.display_name.trim()].length > 70) return HttpResponse.json({ error: "invalid display name" }, { status: 400 })
    guild.display_name = body.display_name.trim()
    if (body.description !== undefined) guild.description = body.description
    return HttpResponse.json(guild)
  }),

  // operationId: enableGuild
  http.patch('*/api/v1/guilds/:id/enable', async ({ params }) => {
    await delay(300)
    const id = Number(params['id'])
    const guild = guilds.find((g) => g.id === id)
    if (guild) guild.enabled = true
    return new HttpResponse(null, { status: 204 })
  }),

  // operationId: disableGuild
  http.patch('*/api/v1/guilds/:id/disable', async ({ params }) => {
    await delay(300)
    const id = Number(params['id'])
    const guild = guilds.find((g) => g.id === id)
    if (guild) guild.enabled = false
    return new HttpResponse(null, { status: 204 })
  }),

  // operationId: inviteGuildMember
  http.post('*/api/v1/guilds/:id/invites', async ({ params, request }) => {
    if (!mockGuildMember(request, Number(params['id']))) return HttpResponse.json({ error: 'forbidden' }, { status: 403 })
    const guildID = Number(params['id'])
    const body = await request.json() as { email: string }
    const parsed = inviteSchema.safeParse(body)
    if (!parsed.success) return HttpResponse.json({ error: 'invalid invitation email' }, { status: 400 })
    const { email } = parsed.data
    if (mockGuildMemberByEmail(email, guildID)) return HttpResponse.json({ error: 'user is already a guild member', code: 'guild_member_already_set' }, { status: 409 })
    if (invitations.some((invite) => invite.guildID === guildID && invite.email === email && invite.status === 'PENDING')) return HttpResponse.json({ error: 'user already invited to guild', code: 'invitation_already_pending' }, { status: 409 })
    invitations.push({ id: invitations.length + 1, guildID, email, invitedBy: mockUserID(request), inviterEmail: request.headers.get('X-User-Email') ?? '', status: 'PENDING' })
    await delay(400)
    return new HttpResponse(null, { status: 204 })
  }),

  // operationId: listGuildInvitations
  http.get('*/api/v1/guild-invites', ({ request }) => {
    if (!mockUserID(request)) return HttpResponse.json({ error: 'unauthorized' }, { status: 401 })
    const email = request.headers.get('X-User-Email')?.trim().toLowerCase()
    return HttpResponse.json(invitations.filter((invite) => invite.email === email && invite.status === 'PENDING').map((invite) => ({
      id: invite.id, guild: guilds.find((guild) => guild.id === invite.guildID)!,
      invited_by: { name: invite.invitedBy === 1 ? 'Ana Silva' : 'Convidador', email: invite.inviterEmail },
    })))
  }),
  // operationId: acceptGuildInvitation
  http.post('*/api/v1/guild-invites/:id/accept', ({ request, params }) => respondInvitation(request, Number(params['id']), true)),
  // operationId: declineGuildInvitation
  http.post('*/api/v1/guild-invites/:id/decline', ({ request, params }) => respondInvitation(request, Number(params['id']), false)),

  // operationId: removeGuildMember
  http.delete('*/api/v1/guilds/:id/members/:userID', async ({ params, request }) => {
    if (!mockGuildMember(request, Number(params['id'])) || Number(params['userID']) === mockUserID(request)) return HttpResponse.json({ error: 'forbidden' }, { status: 403 })
    await delay(400)
    return new HttpResponse(null, { status: 204 })
  }),
]

export function resetGuildsMocks() {
  guilds = structuredClone(guildsFixtures)
  resetMockMemberships()
  invitations = []
}

function respondInvitation(request: Request, id: number, accept: boolean) {
  const userID = mockUserID(request)
  if (!userID) return HttpResponse.json({ error: 'unauthorized' }, { status: 401 })
  const email = request.headers.get('X-User-Email')?.trim().toLowerCase()
  const invitation = invitations.find((invite) => invite.id === id && invite.email === email)
  if (!invitation) return HttpResponse.json({ error: 'invitation not found' }, { status: 404 })
  if (invitation.status !== 'PENDING') return HttpResponse.json({ error: 'invitation already answered', code: 'invitation_already_answered' }, { status: 409 })
  invitation.status = accept ? 'ACCEPTED' : 'DECLINED'
  if (!accept) return new HttpResponse(null, { status: 204 })
  addMockGuildMember(userID, invitation.guildID)
  return HttpResponse.json(guilds.find((guild) => guild.id === invitation.guildID))
}
