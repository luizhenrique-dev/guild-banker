import type { DemoMember } from '../../features/guilds/api/use-guilds'
export const membersByGuild: Record<number, DemoMember[]> = {
  1: [
    { id: 1, name: 'Ana Silva', email: 'ana@familia.com', role: 'Criadora' },
    { id: 2, name: 'Carlos Silva', email: 'carlos@familia.com', role: 'Membro' },
    { id: 3, name: 'Maria Silva', email: 'maria@familia.com', role: 'Membro' },
  ],
  2: [
    { id: 1, name: 'Ana Silva', email: 'ana@familia.com', role: 'Membro' },
    { id: 4, name: 'Bruno Lima', email: 'bruno@republica.com', role: 'Criador' },
  ],
}
