import { useTranslation } from 'react-i18next'
import { useSession } from '@/src/app/session-provider'

export function HomePage() {
  const { user } = useSession()
  const { t } = useTranslation()
  return <main className="p-8"><h1 className="text-[32px]">{user?.name ? t('home.welcome', { name: user.name }) : t('home.welcomeAnonymous')}</h1></main>
}
