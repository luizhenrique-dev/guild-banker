// ============================================================
// 📌 PADRÃO: Zustand Store — Guild selecionado
// ============================================================
// O guild ativo determina o escopo de transações e importações.
// O header exibe um seletor de guild (dropdown).
// ============================================================
import { create } from 'zustand'

import type { Guild } from '@/src/features/guilds/types'
export type { Guild } from '@/src/features/guilds/types'

interface GuildState {
  activeGuild: Guild | null
  guilds: Guild[]
  setActiveGuild: (guild: Guild) => void
  setGuilds: (guilds: Guild[]) => void
}

export const useGuildStore = create<GuildState>((set) => ({
  activeGuild: null,
  guilds: [],
  setActiveGuild: (guild: Guild) => set({ activeGuild: guild }),
  setGuilds: (guilds: Guild[]) => set({ guilds }),
}))
