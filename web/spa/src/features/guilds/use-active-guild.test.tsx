import { useSession } from '@/src/app/session-provider'
import { useQueryClient } from '@tanstack/react-query'
import { http, HttpResponse } from 'msw'
import { server } from '@/src/mocks/server'
import { act, renderHook, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { expect, it, vi, afterEach } from 'vitest'
import { Providers } from '@/src/shared/components/providers'
import { useActiveGuild } from './use-active-guild'

const key = 'guildbanker:last-guild:demo:1'
afterEach(() => { localStorage.clear(); vi.restoreAllMocks() })
function setup(path = '/home', userID = 1) {
  const wrapper = ({ children }: { children: React.ReactNode }) => <MemoryRouter initialEntries={[path]}><Providers initialUser={{ id: userID, name: 'Ana', email: 'ana@familia.com' }} demo>{children}</Providers></MemoryRouter>
  return renderHook(useActiveGuild, { wrapper })
}
it('requires an explicit choice among multiple guilds', async () => {
  const { result } = setup()
  await waitFor(() => expect(result.current.guilds).toHaveLength(2))
  expect(result.current.activeGuild).toBeUndefined()
  act(() => result.current.setActiveGuild({ id: 2 }))
  await waitFor(() => expect(result.current.activeGuild?.id).toBe(2))
  expect(localStorage.getItem(key)).toBe('2')
})
it('restores the last choice after remount', async () => {
  localStorage.setItem(key, '2')
  const { result } = setup()
  await waitFor(() => expect(result.current.activeGuild?.id).toBe(2))
})
it('uses a valid URL before the saved choice', async () => {
  localStorage.setItem(key, '2')
  const { result } = setup('/transactions?guild=1&category=OTHER')
  await waitFor(() => expect(result.current.activeGuild?.id).toBe(1))
  await waitFor(() => expect(localStorage.getItem(key)).toBe('1'))
})
it('does not recover an invalid explicit URL from storage', async () => {
  localStorage.setItem(key, '2')
  const { result } = setup('/home?guild=999')
  await waitFor(() => expect(result.current.guilds).toHaveLength(2))
  expect(result.current.activeGuild).toBeUndefined()
})
it('ignores unavailable storage and isolates users', async () => {
  localStorage.setItem(key, '1')
  vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => { throw new Error('blocked') })
  vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => { throw new Error('blocked') })
  const { result } = setup('/home', 2)
  await waitFor(() => expect(result.current.activeGuild?.id).toBe(2))
})
it.each(['999', 'corrupted', '0', '9007199254740993'])('does not select stale or corrupted saved ID %s', async (id) => {
  localStorage.setItem(key, id)
  const { result } = setup()
  await waitFor(() => expect(result.current.guilds).toHaveLength(2))
  expect(result.current.activeGuild).toBeUndefined()
})

it('does not apply another user preference to a user without memberships', async () => {
  localStorage.setItem(key, '2')
  const { result } = setup('/home', 8)
  await waitFor(() => expect(result.current.isLoading).toBe(false))
  expect(result.current.activeGuild).toBeUndefined()
  expect(result.current.guilds).toEqual([])
  expect(localStorage.getItem(key)).toBe('2')
})

function lifecycleProbe() {
  const wrapper = ({ children }: { children: React.ReactNode }) => <MemoryRouter><Providers initialUser={{ id: 1, name: 'Ana', email: 'ana@familia.com' }} demo>{children}</Providers></MemoryRouter>
  return renderHook(() => ({ guild: useActiveGuild(), session: useSession(), client: useQueryClient() }), { wrapper })
}
it('keeps a newer 403 authoritative over an earlier mutation response', async () => {
  const { result } = lifecycleProbe()
  await waitFor(() => expect(result.current.guild.guilds).toHaveLength(2))
  const activate = result.current.guild.activateJoinedGuild
  const guild = result.current.guild.guilds[0]!
  server.use(http.get('*/api/v1/guilds', () => HttpResponse.json({ error: 'forbidden' }, { status: 403 })))
  await act(async () => { await result.current.client.invalidateQueries({ queryKey: ['guilds'] }) })
  await waitFor(() => expect(result.current.guild.isForbidden).toBe(true))
  act(() => activate(guild))
  expect(result.current.client.getQueryState(['guilds', 1])?.status).toBe('error')
  expect(result.current.client.getQueryState(['guilds', 1])?.error).toMatchObject({ status: 403 })
  expect(result.current.guild.isForbidden).toBe(true)
  expect(result.current.guild.activeGuild).toBeUndefined()
})
it('retains preference on logout and ignores late activation from that session', async () => {
  const { result } = lifecycleProbe()
  await waitFor(() => expect(result.current.guild.guilds).toHaveLength(2))
  const activate = result.current.guild.activateJoinedGuild
  const guild = result.current.guild.guilds[0]!
  act(() => result.current.guild.setActiveGuild({ id: 2 }))
  await waitFor(() => expect(localStorage.getItem(key)).toBe('2'))
  act(() => result.current.session.logout())
  act(() => activate(guild))
  expect(result.current.guild.activeGuild).toBeUndefined()
  expect(localStorage.getItem(key)).toBe('2')
})
