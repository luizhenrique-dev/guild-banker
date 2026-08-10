import { http, HttpResponse, delay } from 'msw'
import { guildsFixtures } from '../fixtures/guilds'
import type { Guild } from '@/src/shared/store/guild.store'

let guilds = [...guildsFixtures]

export const guildsHandlers = [
  // operationId: listGuildsByMember
  http.get('/api/v1/guilds', async () => {
    await delay(500)
    return HttpResponse.json(guilds)
  }),

  // operationId: createGuild
  http.post('/api/v1/guilds', async ({ request }) => {
    await delay(500)
    const body = (await request.json()) as { name: string; display_name: string }
    const newGuild: Guild = {
      id: guilds.length + 10,
      name: body.name,
      display_name: body.display_name,
      enabled: true,
    }
    guilds.push(newGuild)
    return HttpResponse.json(newGuild, { status: 201 })
  }),

  // operationId: updateGuildName
  http.put('/api/v1/guilds/:id', async ({ params, request }) => {
    await delay(400)
    const id = Number(params['id'])
    const body = (await request.json()) as { name: string }
    const guild = guilds.find((g) => g.id === id)
    if (!guild) return HttpResponse.json({ error: 'Não encontrado' }, { status: 404 })
    guild.display_name = body.name
    return HttpResponse.json(guild)
  }),

  // operationId: enableGuild
  http.patch('/api/v1/guilds/:id/enable', async ({ params }) => {
    await delay(300)
    const id = Number(params['id'])
    const guild = guilds.find((g) => g.id === id)
    if (guild) guild.enabled = true
    return new HttpResponse(null, { status: 204 })
  }),

  // operationId: disableGuild
  http.patch('/api/v1/guilds/:id/disable', async ({ params }) => {
    await delay(300)
    const id = Number(params['id'])
    const guild = guilds.find((g) => g.id === id)
    if (guild) guild.enabled = false
    return new HttpResponse(null, { status: 204 })
  }),

  // operationId: inviteGuildMember
  http.post('/api/v1/guilds/:id/invites', async () => {
    await delay(400)
    return new HttpResponse(null, { status: 204 })
  }),

  // operationId: removeGuildMember
  http.delete('/api/v1/guilds/:id/members/:userID', async () => {
    await delay(400)
    return new HttpResponse(null, { status: 204 })
  }),
]
