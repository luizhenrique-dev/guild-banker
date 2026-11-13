import { guildLabel } from '../guild-label'
import { useEffect, useRef, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useTranslation } from 'react-i18next'
import { ArrowRight, Users } from 'lucide-react'
import { useSession } from '@/src/app/session-provider'
import { APIError } from '@/src/shared/lib/api-client'
import { AccessDeniedPage } from '@/src/shared/components/access-denied-page'
import { useOnboarding } from '../api/use-onboarding'
import { createGuildSchema, type CreateGuildInput } from '../schemas/create-guild.schema'

const button = 'rounded-full bg-[#0052ff] px-5 py-3 font-semibold text-white disabled:opacity-50'
export function GuildOnboarding({ showHomeLink = true }: { showHomeLink?: boolean }) {
  const { t } = useTranslation()
  const { user } = useSession()
  const [started, setStarted] = useState(false)
  const panelRef = useRef<HTMLElement>(null)
  const { context, invitations, create, respond, actionError } = useOnboarding(started)
  const pendingInvites = invitations.data ?? []
  const showWelcome = !started && context.guilds.length === 0
  const step = context.guilds.length ? 'selection' : showWelcome ? 'welcome' : pendingInvites.length ? 'invitations' : 'create'
  const firstName = user?.name.trim().split(/\s+/)[0]
  const error = context.error ?? (started && context.guilds.length === 0 ? invitations.error : null)
  const forbidden = error instanceof APIError && error.status === 403
  const loading = context.isLoading || (started && !context.guilds.length && (invitations.isPending || invitations.isFetching) && !error)
  useEffect(() => {
    if (started && !loading && !error) panelRef.current?.querySelector('h1')?.focus({ preventScroll: true })
  }, [step, started, loading, error])
  if (forbidden) return <AccessDeniedPage showHomeLink={showHomeLink} />
  if (loading) return <main className="flex flex-1 items-center justify-center p-8" role="status">{t('common.loading')}</main>
  if (error) return <main className="flex flex-1 flex-col items-center justify-center gap-4 p-8" role="alert"><p>{t('onboarding.loadError')}</p><button className={button} onClick={() => void (context.error ? context.refetch() : invitations.refetch())}>{t('common.retry')}</button></main>
  return <main className="flex flex-1 items-center justify-center bg-[#f7f7f7] px-4 py-10 sm:px-8">
    <section key={step} ref={panelRef} className="h-fit w-full min-w-0 max-w-xl break-words rounded-3xl border border-[#dee1e6] bg-white p-6 shadow-sm motion-safe:animate-fade-in sm:p-10" aria-labelledby="onboarding-title">
      <Users className={`mb-5 h-12 w-12 rounded-full bg-[#eef0f3] p-3 text-[#0052ff] ${showWelcome ? 'mx-auto' : ''}`} aria-hidden="true" />
      {actionError && <p role="alert" className="mb-4 text-red-700">{t(actionError)}</p>}
      {context.guilds.length > 0 ? <>
        <h1 id="onboarding-title" tabIndex={-1} className="text-3xl outline-none">{t('onboarding.select')}</h1>
        <p className="my-3 text-sm text-[#5b616e]">{t('guild.selectValid')}</p>
        <div className="grid gap-3">{context.guilds.map((guild) => <button key={guild.id} className="min-w-0 rounded-2xl border p-4 text-left hover:border-[#0052ff] focus-visible:outline-[#0052ff]" onClick={() => context.setActiveGuild(guild)} aria-label={t('onboarding.enter', { name: guildLabel(guild, context.guilds) })}>
          <strong>{guild.display_name}</strong>
          {guild.description && <p className="mt-1 text-sm text-[#5b616e]">{guild.description}</p>}
          {context.guilds.filter((candidate) => candidate.display_name === guild.display_name).length > 1 && <p className="text-sm">{guild.name}</p>}
        </button>)}</div>
      </> : showWelcome ? <div className="text-center">
        <h1 id="onboarding-title" tabIndex={-1} className="text-3xl font-semibold tracking-tight outline-none">{firstName ? t('onboarding.welcome', { name: firstName }) : t('onboarding.welcomeAnonymous')}</h1>
        <p className="mt-5 text-base font-medium">{t('onboarding.firstStep')}</p>
        <p className="mt-3 text-sm leading-6 text-[#5b616e]">{t('onboarding.guildExplanation')}</p>
        <p className="mt-3 text-sm leading-6 text-[#5b616e]">{t('onboarding.welcomeInviteHint')}</p>
        <button type="button" className={`${button} mt-8 inline-flex w-full items-center justify-center gap-2`} onClick={() => setStarted(true)}>
          {t('onboarding.start')}<ArrowRight size={18} aria-hidden="true" />
        </button>
      </div> : pendingInvites.length ? <>
        <h1 id="onboarding-title" tabIndex={-1} className="text-3xl outline-none">{t('onboarding.invites')}</h1>
        <p className="my-3 text-sm text-[#5b616e]">{t('onboarding.inviteHint')}</p>
        <div className="grid gap-4">{pendingInvites.map((invitation) => <article key={invitation.id} className="min-w-0 rounded-2xl border p-4" aria-label={guildLabel(invitation.guild, pendingInvites.map((item) => item.guild))}>
          <h2 className="text-lg font-semibold">{invitation.guild.display_name}</h2>
          {pendingInvites.filter((item) => item.guild.display_name === invitation.guild.display_name).length > 1 && <p className="text-sm">{invitation.guild.name}</p>}
          {invitation.guild.description && <p className="my-2 text-sm">{invitation.guild.description}</p>}
          <p className="my-3 break-words text-sm text-[#5b616e]">{t('onboarding.invitedBy', { name: invitation.invited_by.name, email: invitation.invited_by.email })}</p>
          <div className="flex flex-wrap gap-3">
            <button className={button} disabled={respond.isPending} onClick={() => respond.mutate({ id: invitation.id, accept: true })}>{t('onboarding.accept')}</button>
            <button className="rounded-full border px-5 py-3 disabled:opacity-50" disabled={respond.isPending} onClick={() => respond.mutate({ id: invitation.id, accept: false })}>{t('onboarding.decline')}</button>
          </div>
        </article>)}</div>
      </> : <CreateGuildForm pending={create.isPending} submit={(input) => create.mutate(input)} />}
    </section>
  </main>
}

function CreateGuildForm({ pending, submit }: { pending: boolean; submit: (input: CreateGuildInput) => void }) {
  const { t } = useTranslation()
  const { register, handleSubmit, formState: { errors } } = useForm<CreateGuildInput>({ resolver: zodResolver(createGuildSchema), defaultValues: { display_name: '', description: '' } })
  return <>
    <h1 id="onboarding-title" tabIndex={-1} className="text-3xl outline-none">{t('onboarding.create')}</h1>
    <p className="my-3 text-sm text-[#5b616e]">{t('onboarding.createHint')}</p>
    <form className="mt-6 grid gap-5" onSubmit={handleSubmit((input) => { if (!pending) submit(input) })}>
      <div>
        <label htmlFor="guild-display-name" className="mb-2 block font-semibold">{t('onboarding.name')}</label>
        <input id="guild-display-name" autoComplete="organization" className="w-full rounded-xl border p-3" aria-invalid={!!errors.display_name} aria-describedby={errors.display_name ? 'guild-name-error' : undefined} {...register('display_name')} />
        {errors.display_name && <p id="guild-name-error" role="alert" className="mt-2 text-sm text-red-700">{t(errors.display_name.message!)}</p>}
      </div>
      <div>
        <label htmlFor="guild-description" className="mb-2 block font-semibold">{t('onboarding.description')}</label>
        <textarea id="guild-description" className="min-h-24 w-full rounded-xl border p-3" {...register('description')} />
      </div>
      <button type="submit" className={button} disabled={pending}>{t(pending ? 'onboarding.creating' : 'onboarding.createButton')}</button>
    </form>
  </>
}
