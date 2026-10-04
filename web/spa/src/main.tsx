import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import '@fontsource/inter/latin-400.css'
import '@fontsource/inter/latin-500.css'
import '@fontsource/inter/latin-600.css'
import '@fontsource/jetbrains-mono/latin-400.css'
import './app/styles.css'
import { App } from './app/app'
import { startMocks } from './app/start-mocks'
import type { SessionUser } from './app/session-provider'

async function bootstrap() {
  const root = createRoot(document.getElementById('root')!)
  try {
    const demo = await startMocks(import.meta.env.DEV, import.meta.env.VITE_USE_MOCKS)
    let user: SessionUser | null = null
    if (import.meta.env.DEV && demo) {
      const { demoUser } = await import('./mocks/fixtures/session')
      user = demoUser
    }
    root.render(<BrowserRouter><App initialUser={user} demo={import.meta.env.DEV && demo} /></BrowserRouter>)
  } catch {
    root.render(<main role="alert">Falha ao iniciar a demonstração. Recarregue para tentar novamente.</main>)
  }
}

void bootstrap()
