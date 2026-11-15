import { useTranslation } from 'react-i18next'
import { useSession } from '@/src/app/session-provider'
import { useState } from 'react'
import { CalendarClock } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { APIError } from '@/src/shared/lib/api-client'
import { useFixedExpenses } from '../../fixed-expenses/api/use-fixed-expenses'
import { monthlyTotal } from '../../fixed-expenses/monthly-total'
import { upcomingExpenses } from '../upcoming-expenses'

export function HomePage() {
  const { user } = useSession()
  const { t, i18n } = useTranslation()
  const [reference] = useState(() => new Date())
  const { data, isLoading, error, refetch, isFetching } = useFixedExpenses()
  const items = upcomingExpenses(data ?? [], reference)
  const forbidden = error instanceof APIError && error.status === 403
  const dates = new Intl.DateTimeFormat(i18n.language, { dateStyle: 'short' })
  const month = new Intl.DateTimeFormat(i18n.language, { month: 'long', year: 'numeric' }).format(reference)

  return <main className="p-6 sm:p-8">
    <h1 className="text-[32px]">{user?.name ? t('home.welcome', { name: user.name }) : t('home.welcomeAnonymous')}</h1>
    <section aria-labelledby="home-expenses-title" className="mt-8 rounded-3xl border border-border bg-card p-5 sm:p-6">
      <div className="mb-6 flex items-center gap-3">
        <CalendarClock aria-hidden="true" className="shrink-0 text-primary" size={24} />
        <div className="min-w-0">
          <h2 id="home-expenses-title" className="text-xl font-semibold">{t('home.expensesTitle')}</h2>
          <p className="mt-1 text-sm text-muted-foreground">{month}</p>
        </div>
      </div>
      {isLoading ? <p role="status" className="text-muted-foreground">{t('common.loading')}</p>
        : error ? <div role="alert">
          <p className="text-destructive">{t(forbidden ? 'accessDenied.title' : 'home.expensesError')}</p>
          {forbidden ? <p className="mt-2 text-sm text-muted-foreground">{t('accessDenied.message')}</p>
            : <Button type="button" variant="outline" disabled={isFetching} className="mt-4" onClick={() => void refetch()}>{t('common.retry')}</Button>}
        </div>
        : items.length === 0 ? <p className="text-muted-foreground">{t('home.expensesEmpty')}</p>
          : <ol className="divide-y divide-border">
            {items.map(({ expense, dueDate }) => {
              const dayDifference = dueDate.getDate() - reference.getDate()
              return <li key={expense.id} className="flex flex-col gap-2 py-4 first:pt-0 last:pb-0 sm:flex-row sm:items-center sm:justify-between sm:gap-6">
                <p className="min-w-0 break-words font-medium">{expense.name}</p>
                <div className="flex shrink-0 flex-wrap items-center gap-x-4 gap-y-1 sm:justify-end">
                  <p className="font-mono font-semibold">{monthlyTotal([expense], i18n.language)}</p>
                  <p className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
                    {dayDifference <= 1 && <span className="rounded-full bg-secondary px-2 py-1 font-medium text-secondary-foreground">{t(dayDifference === 0 ? 'home.today' : 'home.tomorrow')}</span>}
                    <time dateTime={`${dueDate.getFullYear()}-${String(dueDate.getMonth() + 1).padStart(2, '0')}-${String(dueDate.getDate()).padStart(2, '0')}`}>{dates.format(dueDate)}</time>
                  </p>
                </div>
              </li>
            })}
          </ol>}
    </section>
  </main>
}
