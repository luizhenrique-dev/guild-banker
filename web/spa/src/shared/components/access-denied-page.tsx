import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

export function AccessDeniedPage() {
  const { t } = useTranslation()
  return <main className="p-8" aria-labelledby="access-denied-title">
    <h1 id="access-denied-title" className="text-3xl">{t('accessDenied.title')}</h1>
    <p className="my-4">{t('accessDenied.message')}</p>
    <Link className="text-primary underline" to="/home">{t('accessDenied.home')}</Link>
  </main>
}
