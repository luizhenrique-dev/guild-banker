import { useTranslation } from 'react-i18next'
import { useSession } from '@/src/app/session-provider'
import { useState } from 'react'
import { CalendarClock, ChartPie } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { APIError } from '@/src/shared/lib/api-client'
import { useFixedExpenses } from '../../fixed-expenses/api/use-fixed-expenses'
import { monthlyTotal } from '../../fixed-expenses/monthly-total'
import { upcomingExpenses } from '../upcoming-expenses'
import { expensesByCategory } from '../expenses-by-category'
import { CATEGORY_ICONS } from '../../fixed-expenses/types'

export function HomePage() {
  const { user } = useSession()
  const { t, i18n } = useTranslation()
  const [reference] = useState(() => new Date())
  const { data, isLoading, error, refetch, isFetching } = useFixedExpenses()
  const items = upcomingExpenses(data ?? [], reference)
  const categories = expensesByCategory(data ?? [])
  const forbidden = error instanceof APIError && error.status === 403
  const dates = new Intl.DateTimeFormat(i18n.language, { dateStyle: 'short' })
  const month = new Intl.DateTimeFormat(i18n.language, { month: 'long', year: 'numeric' }).format(reference)
  const expenseStatus = isLoading ? <p role="status" className="text-muted-foreground">{t('common.loading')}</p>
    : error ? <div role="alert">
      <p className="text-destructive">{t(forbidden ? 'accessDenied.title' : 'home.expensesError')}</p>
      {forbidden ? <p className="mt-2 text-sm text-muted-foreground">{t('accessDenied.message')}</p>
        : <Button type="button" variant="outline" disabled={isFetching} className="mt-4" onClick={() => void refetch()}>{t('common.retry')}</Button>}
    </div> : null

  return <main className="p-6 sm:p-8">
    <h1 className="text-[32px]">{user?.name ? t('home.welcome', { name: user.name }) : t('home.welcomeAnonymous')}</h1>
    <div className="mt-8 grid max-w-[904px] auto-rows-fr items-stretch gap-6 lg:grid-cols-2">
      <section aria-labelledby="home-expenses-title" className="w-full max-w-[440px] rounded-3xl border border-border bg-card p-5 sm:p-6">
        <div className="mb-2 flex items-center gap-3">
          <CalendarClock aria-hidden="true" className="shrink-0 text-primary" size={24} />
          <div className="min-w-0">
            <h2 id="home-expenses-title" className="text-lg font-semibold">{t('home.expensesTitle')}</h2>
            <p className="mt-1 text-sm text-muted-foreground">{month}</p>
          </div>
        </div>
        {expenseStatus ?? (items.length === 0 ? <p className="text-muted-foreground">{t('home.expensesEmpty')}</p>
            : <ol className="divide-y divide-border">
              {items.map(({ expense, dueDate }) => {
                const dayDifference = dueDate.getDate() - reference.getDate()
                return <li key={expense.id} className="flex h-14 items-center justify-between gap-4 py-2">
                  <p className="min-w-0 line-clamp-2 break-words text-sm font-medium" title={expense.name}>{expense.name}</p>
                  <div className="flex shrink-0 flex-col items-end">
                    <p className="font-mono font-semibold">{monthlyTotal([expense], i18n.language)}</p>
                    <p className="flex items-center gap-2 text-xs text-muted-foreground">
                      {dayDifference <= 1 && <span className="rounded-full bg-secondary px-2 font-medium text-secondary-foreground">{t(dayDifference === 0 ? 'home.today' : 'home.tomorrow')}</span>}
                      <time dateTime={`${dueDate.getFullYear()}-${String(dueDate.getMonth() + 1).padStart(2, '0')}-${String(dueDate.getDate()).padStart(2, '0')}`}>{dates.format(dueDate)}</time>
                    </p>
                  </div>
                </li>
              })}
            </ol>)}
      </section>
      <section aria-labelledby="home-categories-title" className="w-full max-w-[440px] rounded-3xl border border-border bg-card p-5 sm:p-6">
        <div className="mb-2 flex items-center gap-3">
          <ChartPie aria-hidden="true" className="shrink-0 text-primary" size={24} />
          <div className="min-w-0">
            <h2 id="home-categories-title" className="text-lg font-semibold">{t('home.categoriesTitle')}</h2>
            <p className="mt-1 text-sm text-muted-foreground">{month}</p>
          </div>
        </div>
        {expenseStatus ?? (categories.length === 0 ? <p className="text-muted-foreground">{t('home.categoriesEmpty')}</p>
          : <ol tabIndex={categories.length > 5 ? 0 : undefined} aria-label={t('home.categoriesTitle')} className="max-h-[280px] overflow-y-auto overscroll-contain pr-3 divide-y divide-border focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
            {categories.map(({ category, expenses }) => <li key={category} className="flex h-14 items-center justify-between gap-4 py-2">
              <p className="flex min-w-0 items-center gap-2 text-sm font-medium">
                <span aria-hidden="true">{CATEGORY_ICONS[category]}</span>
                <span className="line-clamp-2 break-words" title={t(`category.${category}`)}>{t(`category.${category}`)}</span>
              </p>
              <p className="shrink-0 font-mono font-semibold text-orange-600 dark:text-orange-400">{monthlyTotal(expenses, i18n.language)}</p>
            </li>)}
          </ol>)}
      </section>
    </div>
  </main>
}
