# ISSUE-3 — T4: Exibir página de acesso negado em consultas

Spec: [ISSUE-3](./ISSUE-3-cadastro-sincronizacao-usuarios.md)

- **Status:** ⬜ pendente
- **Depende de:** T3

## O que fazer

1. Criar `AccessDeniedPage` compartilhada, acessível, com título/mensagem genéricos e link Home. Renderizar no conteúdo da rota atual, mantendo Header/Sidebar/Sair. Mensagens PT/EN; não criar rota `/403` nem expor detalhes da conta/conflito/verificação de e-mail.
2. Reconhecer `APIError` com status 403 nas consultas necessárias à tela: lista de despesas pessoais, lista de transações e lista de guilds para `/transactions`, `/import`, `/guild`. A importação atual não consulta revisão por GET remoto; nela a falha de carregamento tratada é de guilds. 403 prevalece sobre dados antigos em cache e não dispara retry/redirect automático.
3. Ajustar `useActiveGuild`/guarda para não entregar guilds/seleção recusadas nem reescrever URL a partir do cache após 403; impedir carga dependente desse estado. Header não exibe lista antiga. A Home permanece acessível, e despesas pessoais com GET bem sucedido não são bloqueadas só pelo GET de guilds. Guild vazia/ID inválido e erros não 403 mantêm seus fluxos anteriores.
4. Preservar separação 401/403 de `SessionProvider`: nenhum logout, novo login, cache global limpo ou estado global de bloqueio por causa de 403. Link Home e Sair devem funcionar da página negada. Não transformar erro de mutation em página nesta task.
5. Testes **planejados**: casos descritivos em Vitest/Testing Library com MSW para cada rota, 403 inicial/refetch com cache, Home/despesas independentes de guilds, ausência de conteúdo recusado/seletor antigo, navegação Home e logout sem ciclo. Novo teste de transações e componente compartilhado; estender E2E navegação para página negada em viewport existente, usando interceptação de teste, sem bypass de produção.
6. Atualizar README da SPA com o comportamento de carregamento 403 e seu limite; manter demais erros inalterados.

## Arquivos

- [criar] `web/spa/src/shared/components/access-denied-page.tsx`, `access-denied-page.test.tsx` — página e acessibilidade.
- [alterar] `web/spa/src/app/app.tsx`, `app.test.tsx`, `session-provider.test.tsx` — guardas/rotas, sessão preservada.
- [alterar] `web/spa/src/features/guilds/use-active-guild.ts` — erro antes de cache/auto seleção.
- [alterar] `web/spa/src/features/guilds/components/guild-selection-state.tsx`, `guild-page-client.test.tsx` — 403 e estados de guild.
- [alterar] `web/spa/src/shared/components/header.tsx` — não exibir guilds recusadas.
- [alterar] `web/spa/src/features/fixed-expenses/components/fixed-expenses-page-client.tsx`, `fixed-expenses-page-client.test.tsx` — consulta pessoal 403.
- [alterar] `web/spa/src/features/transactions/components/transactions-page-client.tsx` — consulta 403.
- [criar] `web/spa/src/features/transactions/components/transactions-page-client.test.tsx` — carga inicial/refetch e cache.
- [alterar] `web/spa/src/features/imports/components/import-page-client.test.tsx` — guarda da rota dependente de guild.
- [alterar] `web/spa/src/shared/lib/i18n.ts` — PT/EN.
- [alterar] `web/spa/e2e/navigation.spec.ts`, `web/spa/README.md` — jornada/contrato UI.

## Critérios de aceite

- **T4-AC1:** Cada rota afetada mostra página genérica no 403 da consulta necessária, inclusive com dados anteriores; não mostra dados financeiros recusados → CA9.
- **T4-AC2:** Menu/Sair/link Home continuam operáveis, sem logout/login/retry automático por 403 e sem redirect global → CA8/CA9/CA10.
- **T4-AC3:** Guilds recusadas não permanecem no seletor nem habilitam consulta dependente; Home/despesas independentes continuam disponíveis; vazio/não 403 preservados → CA10.
- **T4-AC4:** Página acessível e PT/EN; testes cobrem mensagem sem inferência da causa, rotas diretas, refetch e navegação → CA9/CA10.

## Verificação

Na raiz, conforme README/package.json: `npm --prefix web/spa run api:types:check` → `npm --prefix web/spa run lint` → `npm --prefix web/spa run build` → `npm --prefix web/spa test` → `npm --prefix web/spa run test:e2e` → `npm --prefix web/spa run test:e2e:prod`. Pré-requisitos: Node compatível, dependências instaladas com `npm --prefix web/spa ci`, Chromium e portas livres; config pública OIDC no build conforme README. Esperado: página/guardas validadas e suítes demo/produção existentes passando, sem API/Keycloak reais para os testes interceptados.

Manual: não é gate adicional; acessibilidade, Home/Sair, ausência de lista obsoleta e largura de viewport devem ser observados nos testes de componente/E2E. A demo não comprova autorização backend, coberta por T3.

## Contexto adicional

D6 confirmado: página para carregamento, erro local para ação, Home disponível. `AppRoutes` e `Header` hoje consultam guilds mesmo na Home; não interpretar essa consulta auxiliar como proibição global. O corpo 403 é genérico por desenho e não permite deduzir conflito/desativação/perfil incompleto. Reaproveitar layout e padrões de i18n; não redesenhar telas nem introduzir middleware frontend global. Testes de sessão já cobrem 403 sem logout: preservar e ampliar a evidência.

## Evidências

<!-- Preenchido pelo dev-mage: tabela rota/consulta/cache → resultado, navegação, acessibilidade, gates e commit local. -->
