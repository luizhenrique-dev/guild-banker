'use client'
import { useMutation } from '@tanstack/react-query'
import { guildsApi } from './guilds.api'
import { toast } from 'sonner'
import i18n from '@/src/shared/lib/i18n'

export function useInviteMember() {
  return useMutation({
    mutationFn: ({ guildId, email }: { guildId: number; email: string }) =>
      guildsApi.invite(guildId, email),
    onSuccess: () => toast.success(i18n.t('guild.toast.invited')),
    onError: () => toast.error(i18n.t('fx.toast.error')),
  })
}

export function useRemoveMember() {
  return useMutation({
    mutationFn: ({ guildId, userId }: { guildId: number; userId: number }) =>
      guildsApi.removeMember(guildId, userId),
    onSuccess: () => toast.success(i18n.t('guild.toast.removed')),
    onError: () => toast.error(i18n.t('fx.toast.error')),
  })
}
