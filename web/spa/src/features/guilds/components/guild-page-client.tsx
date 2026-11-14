'use client'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { inviteSchema, inviteEmailPattern, maxInviteEmailLength, type InviteInput } from '../schemas/invite.schema'
import { useTranslation } from 'react-i18next'
import { useActiveGuild } from '../use-active-guild'
import { useInviteMember } from '../api/use-guilds'
import { GuildMembers } from './guild-members'
import { Users, Mail, CheckCircle, XCircle } from 'lucide-react'

export function GuildPageClient() {
  const { t } = useTranslation()
  const { activeGuild } = useActiveGuild()
  const inviteMutation = useInviteMember()
  const { register, handleSubmit, reset, watch, formState: { errors } } = useForm<InviteInput>({ resolver: zodResolver(inviteSchema), mode: 'onChange', defaultValues: { email: '' } })
  const emailField = register('email')
  const inviteEmail = watch('email')
  const handleInvite = handleSubmit(async ({ email }) => {
    if (!activeGuild) return
    try {
      await inviteMutation.mutateAsync({ guildId: activeGuild.id, email })
      reset()
    } catch { /* Preserve input; the hook reports the failure once. */ }
  })

  return (
    <main className="flex-1 px-8 py-8">
      <div className="mb-6">
        <h1 className="text-[32px] font-normal tracking-[-0.8px]">{t('guild.title')}</h1>
        <p className="text-[13px] text-[#5b616e]">{t('guild.subtitle')}</p>
      </div>

      {/* Guild info */}
      <div className="mb-8 rounded-[24px] border border-[#dee1e6] bg-white p-6">
        <div className="flex flex-wrap gap-2 items-center justify-between">
          <div className="flex items-center gap-4">
            <span className="inline-flex h-12 w-12 items-center justify-center rounded-full bg-[#eef0f3]">
              <Users size={22} className="text-[#0052ff]" />
            </span>
            <div>
              <h2 className="text-lg font-semibold">{activeGuild?.display_name ?? ''}</h2>
              <p className="text-[13px] text-[#7c828a]">{activeGuild?.name ?? ''}</p>
            </div>
          </div>
          <span
            className={`inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-[12px] font-semibold ${
              activeGuild?.enabled
                ? 'bg-green-50 text-[#05b169]'
                : 'bg-red-50 text-[#cf202f]'
            }`}
          >
            {activeGuild?.enabled ? (
              <><CheckCircle size={14} /> {t('guild.enabled')}</>
            ) : (
              <><XCircle size={14} /> {t('guild.disabled')}</>
            )}
          </span>
        </div>
      </div>

      {/* Convite */}
      <div className="mb-8 rounded-[24px] border border-[#dee1e6] bg-white p-6">
        <h3 className="mb-4 flex items-center gap-2 text-[15px] font-semibold">
          <Mail size={18} className="text-[#0052ff]" />
          {t('guild.invite')}
        </h3>
        <form onSubmit={handleInvite} className="flex flex-wrap gap-3">
          <input
            type="email"
            required
            pattern={inviteEmailPattern.source}
            maxLength={maxInviteEmailLength}
            autoCapitalize="none"
            aria-label={t('guild.invite')}
            aria-invalid={!!errors.email}
            aria-describedby={errors.email ? 'invite-error' : undefined}
            {...emailField}
            onChange={(event) => {
              event.target.value = event.target.value.toLowerCase()
              void emailField.onChange(event)
            }}
            placeholder={t('guild.invitePlaceholder')}
            className="min-w-0 flex-1 rounded-xl border border-[#dee1e6] px-4 py-2.5 text-[14px] outline-none focus:border-[#0052ff]"
          />
          <button
            type="submit"
            disabled={inviteMutation.isPending || !inviteEmail || !!errors.email}
            className="rounded-full bg-[#0052ff] px-5 py-2.5 text-[14px] font-semibold text-white transition-colors hover:bg-[#003ecc] disabled:opacity-50"
          >
            {t('guild.inviteButton')}
          </button>
        </form>
        {errors.email && <p id="invite-error" role="alert">{t(errors.email.message ?? 'validation.email')}</p>}
      </div>

      <GuildMembers key={activeGuild!.id} guildID={activeGuild!.id} />
    </main>
  )
}
