// ============================================================
// MSW Browser Setup
// 📌 PADRÃO: Inicializa o service worker no browser para interceptar
// requisições em desenvolvimento.
// ============================================================
import { setupWorker } from 'msw/browser'
import { handlers } from './handlers'

export const worker = setupWorker(...handlers)
