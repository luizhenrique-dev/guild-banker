'use client'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useGuildStore } from '@/src/shared/store/guild.store'
import { useInviteMember, useRemoveMember } from '../api/use-guilds'
import { Users, Mail, Trash2, CheckCircle, XCircle } from 'lucide-react'

// Membros mock (não há endpoint GET /members na spec)
const mockMembers = [
  { id: 1, name: 'Ana Silva', email: 'ana@familia.com', role: 'Criadora' },
  { id: 2, name: 'Carlos Silva', email: 'carlos@familia.com', role: 'Membro' },
  { id: 3, name: 'Maria Silva', email: 'maria@familia.com', role: 'Membro' },
]

export function GuildPageClient() {
  const { t } = useTranslation()
  const activeGuild = useGuildStore((s) => s.activeGuild)
  const inviteMutation = useInviteMember()
  const removeMutation = useRemoveMember()
  const [inviteEmail, setInviteEmail] = useState('')

  const handleInvite = (e: React.FormEvent) => {
    e.preventDefault()
    if (!activeGuild || !inviteEmail) return
    inviteMutation.mutate({ guildId: activeGuild.id, email: inviteEmail })
    setInviteEmail('')
  }

  return (
    <main className="flex-1 px-8 py-8">
      <div className="mb-6">
        <h1 className="text-[32px] font-normal tracking-[-0.8px]">{t('guild.title')}</h1>
        <p className="text-[13px] text-[#5b616e]">{t('guild.subtitle')}</p>
      </div>

      {/* Guild info */}
      <div className="mb-8 rounded-[24px] border border-[#dee1e6] bg-white p-6">
        <div className="flex items-center justify-between">
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
        <form onSubmit={handleInvite} className="flex gap-3">
          <input
            type="email"
            value={inviteEmail}
            onChange={(e) => setInviteEmail(e.target.value)}
            placeholder={t('guild.invitePlaceholder')}
            className="flex-1 rounded-xl border border-[#dee1e6] px-4 py-2.5 text-[14px] outline-none focus:border-[#0052ff]"
          />
          <button
            type="submit"
            disabled={inviteMutation.isPending || !inviteEmail}
            className="rounded-full bg-[#0052ff] px-5 py-2.5 text-[14px] font-semibold text-white transition-colors hover:bg-[#003ecc] disabled:opacity-50"
          >
            {t('guild.inviteButton')}
          </button>
        </form>
      </div>

      {/* Membros */}
      <div className="rounded-[24px] border border-[#dee1e6] bg-white">
        <div className="border-b border-[#dee1e6] px-6 py-4">
          <h3 className="flex items-center gap-2 text-[15px] font-semibold">
            <Users size={18} className="text-[#0052ff]" />
            {t('guild.members')}
          </h3>
        </div>
        <div className="divide-y divide-[#eef0f3]">
          {mockMembers.map((member) => (
            <div key={member.id} className="flex items-center justify-between px-6 py-4">
              <div className="flex items-center gap-3">
                <span className="inline-flex h-9 w-9 items-center justify-center rounded-full bg-[#0052ff] text-[13px] font-semibold text-white">
                  {member.name.split(' ').map((n) => n[0]).join('').slice(0, 2)}
                </span>
                <div>
                  <p className="text-[14px] font-medium">{member.name}</p>
                  <p className="text-[12px] text-[#7c828a]">{member.email}</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <span className="rounded-full bg-[#eef0f3] px-2.5 py-1 text-[11px] font-semibold text-[#5b616e]">
                  {member.role}
                </span>
                {member.id !== 1 && (
                  <button
                    type="button"
                    onClick={() =>
                      activeGuild &&
                      removeMutation.mutate({ guildId: activeGuild.id, userId: member.id })
                    }
                    className="text-[#7c828a] transition-colors hover:text-[#cf202f]"
                    title={t('guild.remove')}
                  >
                    <Trash2 size={15} />
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>
    </main>
  )
}
