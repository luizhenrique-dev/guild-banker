// 📌 PADRÃO: Página raiz redireciona para /fixed-expenses
// Em um app real, isso poderia ser um dashboard.
import { redirect } from 'next/navigation'

export default function HomePage() {
  redirect('/fixed-expenses')
}
