import { useTranslation } from 'react-i18next'
import { useActiveGuild } from '../use-active-guild'

export function GuildSelectionState() {
  const { t } = useTranslation()
  const { isLoading, error, guilds, refetch } = useActiveGuild()
  return <main className="p-8" role="status">
    {isLoading ? t('common.loading') : error ? <>
      {t('fx.toast.error')} <button onClick={() => void refetch()}>{t('common.retry')}</button>
    </> : guilds.length ? t('guild.selectValid') : t('guild.empty')}
  </main>
}
