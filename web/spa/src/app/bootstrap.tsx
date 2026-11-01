import type { Root } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { App } from './app'
import { startMocks } from './start-mocks'
import { createKeycloakSession } from './keycloak-session'

export async function bootstrap(root: Root) {
  root.render(<main role="status">Autenticando…</main>)
  try {
    const demo = await startMocks(import.meta.env.DEV, import.meta.env.VITE_USE_MOCKS)
    if (import.meta.env.DEV && demo) {
      const { demoUser } = await import('../mocks/fixtures/session')
      root.render(<BrowserRouter><App initialUser={demoUser} demo /></BrowserRouter>)
      return
    }
    const auth = createKeycloakSession(import.meta.env)
    const user = await auth.initialize()
    root.render(<BrowserRouter><App initialUser={user} auth={auth} /></BrowserRouter>)
  } catch {
    root.render(<main role="alert"><h1>Não foi possível autenticar.</h1><p>Verifique a conexão e a configuração de autenticação.</p><button onClick={() => {
      const url = new URL(window.location.href)
      for (const key of ['code', 'state', 'session_state', 'iss', 'error', 'error_description']) url.searchParams.delete(key)
      if (/^(#|.*&)(code|state|error)=/.test(url.hash)) url.hash = ''
      window.location.replace(url.href)
    }}>Tentar novamente</button></main>)
  }
}
