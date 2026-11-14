import type { Member } from '../../features/guilds/api/use-guilds'
export const membersByGuild: Record<number, Member[]> = {
  1: [
    { id: 1, name: 'Ana Silva', email: 'ana@familia.com', last_access_at: '2026-11-12T12:00:00Z', member_since: '2026-01-10T12:00:00Z' },
    { id: 2, name: 'Carlos Silva', email: 'carlos@familia.com', last_access_at: '2026-11-12T12:00:00Z', member_since: '2026-01-10T12:00:00Z' },
    { id: 3, name: 'Maria Silva', email: 'maria@familia.com', last_access_at: '2026-11-12T12:00:00Z', member_since: '2026-01-10T12:00:00Z' },
  ],
  2: [
    { id: 1, name: 'Ana Silva', email: 'ana@familia.com', last_access_at: '2026-11-12T12:00:00Z', member_since: '2026-01-10T12:00:00Z' },
    { id: 4, name: 'Bruno Lima', email: 'bruno@republica.com', last_access_at: '2026-11-12T12:00:00Z', member_since: '2026-01-10T12:00:00Z' },
  ],
}
