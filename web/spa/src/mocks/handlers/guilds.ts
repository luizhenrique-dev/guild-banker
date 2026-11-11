import { http, HttpResponse, delay } from 'msw'
import { mockUserID, mockGuildMember, addMockGuildMember, resetMockMemberships } from '../request-context'
import { guildsFixtures } from '../fixtures/guilds'
import type { Guild } from '@/src/features/guilds/types'

let guilds = structuredClone(guildsFixtures)

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
    await delay(400)
    return new HttpResponse(null, { status: 204 })
  }),

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
}
