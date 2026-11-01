# ISSUE-2 — T3: Conectar sessão real e guarda integral

Spec: [ISSUE-2](./ISSUE-2-autenticacao-keycloak.md)

- **Status:** ⬜ pendente
- **Depende de:** T2

## O que fazer
Conectar adaptador no bootstrap antes de BrowserRouter. Só montar shell, providers com consultas e telas internas após autenticação. Estados anteriores: carregamento neutro e erro acessível com tentativa explícita, sem shell nem chamadas financeiras. Proteger também navegação interna e wildcard; após autenticação, manter 404 para rota desconhecida.
Adaptar SessionUser para identidade de cache string | number: sessão real usa sub textual e demo conserva ID numérico. Atualizar apenas assinaturas das keys afetadas. Nunca converter UUID para número nem enviar sub como ID interno de domínio. A API já resolve sub em user_account; não criar /me. Na sessão real enviar apenas Authorization Bearer atual; X-User-ID/X-User-Email são exclusivos da demo DEV com opt-in. Manter IDs de guild/transação/membro numéricos. Fixtures de membros continuam exclusivas da demo.
Injetar dependência de autenticação em Providers/SessionProvider sem global mutável. Nenhum initialUser de teste/demo pode habilitar bypass no bootstrap de produção. Demo explícita continua operando sem Keycloak. Configuração ausente, indisponibilidade e callback inválido não autorizam acesso e não geram redirecionamento automático infinito.
Atualizar testes app/bootstrap/sessão e smoke compilado já nesta task: não esperar mais “Sessão não integrada”. No smoke offline, interceptar no navegador a navegação OIDC e inspecionar destino/ausência de conteúdo; isso não injeta sessão em produção. Cobrir configuração ausente separadamente. Build do smoke recebe configuração pública de teste; não depende de redirects 4173 aceitos pelo servidor real.
Testes planejados: raiz/Home e cada rota direta sem sessão, carregamento, erro/callback, SSO simulado, deep link com guild/filtros, 404 autenticado, claims e ausência dos headers demo nas requests reais; atualizar demo para continuar passando.

## Arquivos
- [alterar] web/spa/src/main.tsx
- [alterar] web/spa/src/app/app.tsx
- [alterar] web/spa/src/app/session-provider.tsx
- [alterar] web/spa/src/shared/components/providers.tsx
- [alterar] web/spa/src/features/fixed-expenses/api/use-fixed-expenses.ts
- [alterar] web/spa/src/features/transactions/api/use-transactions.ts
- [alterar] web/spa/src/features/guilds/api/use-guilds.ts
- [alterar] web/spa/src/app/bootstrap.test.tsx
- [alterar] web/spa/src/app/app.test.tsx
- [alterar] web/spa/src/app/session-provider.test.tsx
- [alterar] web/spa/e2e/production.spec.ts
- [alterar] web/spa/README.md

## Critérios de aceite
- T3-AC1: Sem sessão, shell e requests internos permanecem bloqueados e o bootstrap inicia login; loading/erros não liberam conteúdo → CA1/CA2.
- T3-AC2: Sessão autenticada real chega à Home ou deep link/query original; recarga reinicializa OIDC antes do Router → CA3/CA4.
- T3-AC3: Requests reais usam Bearer e identidade de cache textual; headers de identidade numérica continuam exclusivos da demo DEV → CA8.
- T3-AC4: Unitários, demo e smoke dist atualizados passam sem adicionar bypass de produção → CA2/CA8.

## Verificação
Na raiz: `npm --prefix web/spa run api:types:check`, `npm --prefix web/spa run lint`, `npm --prefix web/spa run build`, `npm --prefix web/spa test`, `npm --prefix web/spa run test:e2e` e `npm --prefix web/spa run test:e2e:prod`, nessa ordem conforme gate local documentado no README. Pré-requisitos: Node compatível, dependências instaladas (`npm --prefix web/spa ci`), Chromium instalado e portas livres. Esperado: tipos sem drift, compilação/lint/testes passando; E2E demo e smoke offline sem Keycloak real. Para smoke OIDC, compilar com configuração pública de teste e interceptar navegação externa no teste, nunca habilitar mocks no produto.

## Contexto adicional
Seguir AGENTS.md e docs/go/{architecture,coding,naming}.md; preservar organização feature-first do React e contratos públicos da API. Decisões D1–D5 e premissas P1–P7 na spec. Não fazer cadastro automático. Artefatos de autenticação não podem expor tokens/senha. Cada task deve terminar compilável e com testes atualizados; não deixar regressões para a task seguinte.

## Evidências
<!-- preenchido pelo dev-mage; testes acima são planejados, não executados -->
