// ============================================================
// Página: /fixed-expenses
// 📌 PADRÃO: A página é um Server Component.
// A interatividade fica no FixedExpensesPageClient ("use client").
// ============================================================
import { FixedExpensesPageClient } from '@/src/features/fixed-expenses/components/fixed-expenses-page-client'

export default function FixedExpensesPage() {
  return <FixedExpensesPageClient />
}
