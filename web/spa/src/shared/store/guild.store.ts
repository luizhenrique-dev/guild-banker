// ============================================================
// 📌 PADRÃO: Zustand Store — Guild selecionado
// ============================================================
// O guild ativo determina o escopo de transações e importações.
// O header exibe um seletor de guild (dropdown).
// ============================================================
import { create } from 'zustand'

export interface Guild {
  id: number
  name: string
  display_name: string
  enabled: boolean
}

interface GuildState {
  activeGuild: Guild | null
  guilds: Guild[]
  setActiveGuild: (guild: Guild) => void
  setGuilds: (guilds: Guild[]) => void
}

export const useGuildStore = create<GuildState>((set) => ({
  activeGuild: {
    id: 1,
    name: 'familia-silva',
    display_name: 'Família Silva',
    enabled: true,
  },
  guilds: [
    {
      id: 1,
      name: 'familia-silva',
      display_name: 'Família Silva',
      enabled: true,
    },
  ],
  setActiveGuild: (guild: Guild) => set({ activeGuild: guild }),
  setGuilds: (guilds: Guild[]) => set({ guilds }),
}))
