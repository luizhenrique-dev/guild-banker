'use client'
import { usePathname } from 'next/navigation'
import Link from 'next/link'
import { useTranslation } from 'react-i18next'
import {
  Home,
  ArrowLeftRight,
  Repeat,
  FileUp,
  Users,
} from 'lucide-react'
import type { ReactNode } from 'react'

interface NavItem {
  href: string
  labelKey: string
  icon: ReactNode
}

export function Sidebar() {
  const pathname = usePathname()
  const { t } = useTranslation()

  const navItems: NavItem[] = [
    { href: '/', labelKey: 'nav.home', icon: <Home size={18} /> },
    { href: '/transactions', labelKey: 'nav.transactions', icon: <ArrowLeftRight size={18} /> },
    { href: '/fixed-expenses', labelKey: 'nav.fixedExpenses', icon: <Repeat size={18} /> },
    { href: '/import', labelKey: 'nav.import', icon: <FileUp size={18} /> },
    { href: '/guild', labelKey: 'nav.guild', icon: <Users size={18} /> },
  ]

  return (
    <aside className="sticky top-16 h-[calc(100vh-64px)] w-60 shrink-0 border-r border-[#dee1e6] bg-white px-3 py-6">
      <nav className="space-y-1">
        {navItems.map((item) => {
          const isActive =
            item.href === '/'
              ? pathname === '/'
              : pathname?.startsWith(item.href)

          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex items-center gap-3 rounded-[12px] px-3 py-2.5 text-[14px] font-medium transition-colors ${
                isActive
                  ? 'bg-[#eef0f3] font-semibold text-[#0a0b0d]'
                  : 'text-[#5b616e] hover:bg-[#f7f7f7]'
              }`}
            >
              <span className={isActive ? 'text-[#0052ff]' : ''}>
                {item.icon}
              </span>
              {t(item.labelKey)}
            </Link>
          )
        })}
      </nav>
    </aside>
  )
}
