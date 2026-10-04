# GuildBanker SPA

React 18, TypeScript estrito, Vite, React Router e Tailwind 3.

Requisitos: Node 24.6.0 (ou Node >=22.12 compatível), npm e Chromium para E2E.
Na raiz: `npm --prefix web/spa ci`, `npm --prefix web/spa exec -- playwright install chromium`.

`npm --prefix web/spa run dev:demo` inicia a demonstração explícita em localhost:5173.
`npm --prefix web/spa run dev` não inicializa mocks ou identidade demo.

Checks: `make web/lint`, `make web/build`, `make web/test`, `npm --prefix web/spa run test:e2e`.
O build gera web/spa/dist. Não há servidor Next ou stubs API frontend.

A integração real de sessão permanece fora da GB-20: endpoints protegidos respondem 401 sem Bearer e 501 com Bearer. Headers X-User-ID/Email não desbloqueiam autenticação; CORS também exige trabalho separado. API local: 8081, Keycloak: 8080.
