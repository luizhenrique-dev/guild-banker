'use client'
import { useTranslation } from 'react-i18next'
import { useSession } from '@/src/app/session-provider'
import { useActiveGuild } from '@/src/features/guilds/use-active-guild'
import { LogOut, ChevronDown, Shield, Users } from 'lucide-react'
import { useState, useRef, useEffect } from 'react'

export function Header() {
  const { t, i18n } = useTranslation()
  const { user, logout } = useSession()
  const { activeGuild, guilds, setActiveGuild } = useActiveGuild()
  const [guildOpen, setGuildOpen] = useState(false)
  const guildRef = useRef<HTMLDivElement>(null)

  const currentLang = i18n.language?.startsWith('pt') ? 'pt-BR' : 'en'

  // Close dropdown on outside click
  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (guildRef.current && !guildRef.current.contains(e.target as Node)) {
        setGuildOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [])

  const initials = user?.name
    ?.split(' ')
    ?.map((n) => n?.[0])
    ?.join('')
    ?.toUpperCase()
    ?.slice(0, 2) ?? '??'

  return (
    <header className="sticky top-0 z-20 flex min-h-16 flex-wrap gap-2 items-center justify-between border-b border-[#dee1e6] bg-white px-3 py-2 md:px-6">
      {/* Left: Logo + Guild switcher */}
      <div className="flex items-center gap-2 md:gap-6">
        <div className="flex items-center gap-2">
          <span className="inline-flex h-8 w-8 items-center justify-center rounded-full bg-[#0052ff] text-white">
            <Shield size={14} />
          </span>
          <span className="text-[17px] font-semibold">GuildBanker</span>
        </div>

        {/* Guild Switcher */}
        <div className="relative" ref={guildRef} onKeyDown={(e) => e.key === 'Escape' && setGuildOpen(false)}>
          <button
            type="button"
            aria-expanded={guildOpen}
            onClick={() => setGuildOpen(!guildOpen)}
            className="flex items-center gap-2 rounded-full bg-[#eef0f3] px-3 py-1.5 text-[13px] font-semibold transition-colors hover:bg-[#dee1e6]"
          >
            <Users size={14} className="text-[#0052ff]" />
            {activeGuild?.display_name ?? t('guild.select')}
            <ChevronDown size={12} className="text-[#7c828a]" />
          </button>

          {guildOpen && (
            <div className="absolute right-0 top-full mt-1 w-56 rounded-xl border border-[#dee1e6] bg-white py-1 shadow-lg">
              {guilds.map((g) => (
                <button
                  key={g.id}
                  type="button"
                  onClick={() => {
                    setActiveGuild(g)
                    setGuildOpen(false)
                  }}
                  className={`flex w-full items-center gap-2 px-4 py-2 text-left text-[13px] transition-colors hover:bg-[#f7f7f7] ${
                    g.id === activeGuild?.id ? 'bg-[#eef0f3] font-semibold' : ''
                  }`}
                >
                  <Users size={14} className="text-[#0052ff]" />
                  {g.display_name}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Right: Language switcher + User + Logout */}
      <div className="flex items-center gap-2 md:gap-4">
        {/* Language Switcher */}
        <div className="flex items-center gap-1 rounded-full bg-[#f7f7f7] p-1">
          <button
            type="button"
            onClick={() => i18n.changeLanguage('pt-BR')}
            className={`rounded-full px-3 py-1 text-[12px] font-semibold transition-all ${
              currentLang === 'pt-BR'
                ? 'bg-white shadow-sm'
                : 'text-[#7c828a] hover:text-[#5b616e]'
            }`}
          >
            PT
          </button>
          <button
            type="button"
            onClick={() => i18n.changeLanguage('en')}
            className={`rounded-full px-3 py-1 text-[12px] font-semibold transition-all ${
              currentLang === 'en'
                ? 'bg-white shadow-sm'
                : 'text-[#7c828a] hover:text-[#5b616e]'
            }`}
          >
            EN
          </button>
        </div>

        <span className="h-6 w-px bg-[#dee1e6]" />

        {/* User info */}
        <div className="flex items-center gap-2">
          <span className="inline-flex h-8 w-8 items-center justify-center rounded-full bg-[#0052ff] text-[13px] font-semibold text-white">
            {initials}
          </span>
          <div className="hidden md:block leading-tight">
            <div className="text-[13px] font-semibold">{user?.name ?? ''}</div>
            <div className="text-[11px] text-[#7c828a]" suppressHydrationWarning>
              {user?.email ?? ''}
            </div>
          </div>
        </div>

        {/* Logout */}
        <button
          type="button"
          onClick={logout}
          className="inline-flex h-9 w-9 items-center justify-center rounded-full text-[#7c828a] transition-colors hover:bg-[#f7f7f7] hover:text-[#0a0b0d]"
          aria-label={t('header.logout')} title={t('header.logout')}
        >
          <LogOut size={16} />
        </button>
      </div>
    </header>
  )
}
