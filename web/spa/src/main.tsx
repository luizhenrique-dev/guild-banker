import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import '@fontsource/inter/latin-400.css'
import '@fontsource/inter/latin-500.css'
import '@fontsource/inter/latin-600.css'
import '@fontsource/jetbrains-mono/latin-400.css'
import './app/styles.css'
import { App } from './app/app'
import { startMocks } from './app/start-mocks'
import { useAuthStore } from './shared/store/auth.store'
import { useGuildStore } from './shared/store/guild.store'

async function bootstrap() {
  const root = createRoot(document.getElementById('root')!)
  try {
    const demo = await startMocks(import.meta.env.DEV, import.meta.env.VITE_USE_MOCKS)
    if (import.meta.env.DEV && demo) {
      useAuthStore.getState().setUser({ id: 1, name: 'Ana Silva', email: 'ana@familia.com' })
      const { guildsFixtures } = await import('./mocks/fixtures/guilds')
      useGuildStore.getState().setGuilds(guildsFixtures)
      useGuildStore.getState().setActiveGuild(guildsFixtures[0]!)
    }
    root.render(<BrowserRouter><App /></BrowserRouter>)
  } catch {
    root.render(<main role="alert">Falha ao iniciar a demonstração. Recarregue para tentar novamente.</main>)
  }
}

void bootstrap()
