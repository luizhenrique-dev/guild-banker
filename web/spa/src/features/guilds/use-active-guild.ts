import { useContext } from 'react'
import { ActiveGuildContext } from './active-guild-provider'

export function useActiveGuild() {
  const context = useContext(ActiveGuildContext)
  if (!context) throw new Error('ActiveGuildProvider required')
  return context
}
