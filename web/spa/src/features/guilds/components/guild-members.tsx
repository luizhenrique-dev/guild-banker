import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useSession } from '@/src/app/session-provider'
import { APIError } from '@/src/shared/lib/api-client'
import { AccessDeniedPage } from '@/src/shared/components/access-denied-page'
import { AlertDialog, AlertDialogContent, AlertDialogTitle, AlertDialogDescription, AlertDialogFooter, AlertDialogCancel, AlertDialogAction } from '@/components/ui/alert-dialog'
import { useGuildMembers, useSentInvitations, useCancelInvitation, useRemoveMember, type SentInvitation } from '../api/use-guilds'

const section = 'mb-8 min-w-0 rounded-3xl border border-[#dee1e6] bg-white p-6'
const action = 'rounded-full border border-[#dee1e6] px-4 py-2 text-sm font-semibold disabled:opacity-50'

export function GuildMembers({ guildID }: { guildID: number }) {
  const { t, i18n } = useTranslation()
  const { user } = useSession()
  const members = useGuildMembers(guildID)
  const invitations = useSentInvitations(guildID)
  const cancel = useCancelInvitation()
  const remove = useRemoveMember()
  const [cancelling, setCancelling] = useState<SentInvitation | null>(null)
  const date = (value: string | null) => value
    ? <time dateTime={value}>{new Intl.DateTimeFormat(i18n.language, { dateStyle: 'short', timeStyle: 'short' }).format(new Date(value))}</time>
    : t('guild.noAccess')
  const forbidden = [members.error, invitations.error].some((error) => error instanceof APIError && error.status === 403)
  if (forbidden) return <AccessDeniedPage />
  const cancelError = cancel.error instanceof APIError && [404, 409].includes(cancel.error.status) ? 'guild.cancelUnavailable' : cancel.error instanceof APIError && cancel.error.status === 403 ? 'permission.denied' : 'guild.cancelError'
  return <>
    <section className={section} aria-labelledby="pending-invitations-title">
      <h2 id="pending-invitations-title" className="mb-4 text-lg font-semibold">{t('guild.pendingInvitations')}</h2>
      {invitations.isPending ? <p role="status">{t('common.loading')}</p> : invitations.error ? <div role="alert"><p>{t('guild.invitationsError')}</p><button className={action} onClick={() => void invitations.refetch()}>{t('common.retry')}</button></div> : !invitations.data?.length ? <p className="text-sm text-[#5b616e]">{t('guild.noInvitations')}</p> : <ul className="divide-y divide-[#eef0f3]">
        {invitations.data.map((invitation) => <li key={invitation.id} className="flex min-w-0 flex-wrap items-center justify-between gap-4 py-4">
          <div className="min-w-0 break-words">
            <p className="font-medium">{invitation.email}</p>
            <p className="mt-1 text-sm text-[#5b616e]">{t('guild.invitedBy')}: {invitation.invited_by.name} ({invitation.invited_by.email})</p>
            <p className="mt-1 text-sm text-[#5b616e]">{t('guild.invitedAt')}: {date(invitation.created_at)}</p>
          </div>
          <button className="rounded-full border border-red-200 bg-red-50 px-4 py-2 text-sm font-semibold text-[#cf202f] transition-colors hover:border-red-300 hover:bg-red-100 focus-visible:outline-[#cf202f]" onClick={() => { cancel.reset(); setCancelling(invitation) }} aria-label={t('guild.cancelFor', { email: invitation.email })}>{t('guild.cancelInvitation')}</button>
        </li>)}
      </ul>}
    </section>
    <section className={section} aria-labelledby="guild-members-title">
      <h2 id="guild-members-title" className="mb-4 text-lg font-semibold">{t('guild.members')}</h2>
      {members.isPending ? <p role="status">{t('common.loading')}</p> : members.error ? <div role="alert"><p>{t('guild.membersError')}</p><button className={action} onClick={() => void members.refetch()}>{t('common.retry')}</button></div> : !members.data?.length ? <p>{t('guild.noMembers')}</p> : <ul className="divide-y divide-[#eef0f3]">
        {members.data.map((member) => {
          const current = member.email === user?.email.trim().toLowerCase()
          return <li key={member.id} className="flex min-w-0 flex-wrap items-center justify-between gap-4 py-4">
            <div className="min-w-0 break-words">
              <p className="font-medium"><span>{member.name}</span> {current && <span className="rounded-full bg-[#eef0f3] px-2 py-1 text-xs">{t('guild.you')}</span>}</p>
              <p className="mt-1 text-sm text-[#5b616e]">{member.email}</p>
              <dl className="mt-3 grid gap-2 text-sm sm:grid-cols-2 sm:gap-8">
                <div><dt className="text-[#5b616e]">{t('guild.lastAccess')}</dt><dd>{date(member.last_access_at)}</dd></div>
                <div><dt className="text-[#5b616e]">{t('guild.memberSince')}</dt><dd>{date(member.member_since)}</dd></div>
              </dl>
            </div>
            {!current && <button className={action} aria-label={t('guild.remove') + ' ' + member.name} disabled={remove.isPending} onClick={() => remove.mutate({ guildId: guildID, userId: member.id })}>{t('guild.remove')}</button>}
          </li>
        })}
      </ul>}
    </section>
    <AlertDialog open={!!cancelling} onOpenChange={(open) => { if (!open && !cancel.isPending) setCancelling(null) }}>
      <AlertDialogContent>
        <AlertDialogTitle>{t('guild.cancelInvitation')}</AlertDialogTitle>
        <AlertDialogDescription className="break-words">{t('guild.cancelMessage', { email: cancelling?.email })}</AlertDialogDescription>
        {cancel.error && <p role="alert">{t(cancelError)}</p>}
        <AlertDialogFooter>
          <AlertDialogCancel disabled={cancel.isPending}>{t('guild.keepInvitation')}</AlertDialogCancel>
          <AlertDialogAction disabled={cancel.isPending} onClick={(event) => {
            event.preventDefault()
            if (!cancelling || cancel.isPending) return
            cancel.mutate({ guildId: guildID, id: cancelling.id }, { onSuccess: () => setCancelling(null) })
          }}>{t('guild.confirmCancellation')}</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </>
}
