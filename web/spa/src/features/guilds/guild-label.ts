import type { Guild } from './types'

export function guildLabel(guild: Guild, guilds: Guild[]) {
  return guilds.filter((candidate) => candidate.display_name === guild.display_name).length > 1 ? `${guild.display_name} (${guild.name})` : guild.display_name
}
