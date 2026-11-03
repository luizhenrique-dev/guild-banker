import { useTranslation } from 'react-i18next'
import { AccessDeniedPage } from '../../../shared/components/access-denied-page'
import { useActiveGuild } from '../use-active-guild'

export function GuildSelectionState() {
  const { t } = useTranslation()
  const { isLoading, error, guilds, refetch, isForbidden } = useActiveGuild()
  if (isForbidden) return <AccessDeniedPage />
  return <main className="p-8" role="status">
    {isLoading ? t('common.loading') : error ? <>
      {t('fx.toast.error')} <button onClick={() => void refetch()}>{t('common.retry')}</button>
    </> : guilds.length ? t('guild.selectValid') : t('guild.empty')}
  </main>
}
