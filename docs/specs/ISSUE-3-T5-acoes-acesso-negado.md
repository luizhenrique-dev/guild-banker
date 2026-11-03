# ISSUE-3 — T5: Tratar permissão negada nas ações existentes

Spec: [ISSUE-3](./ISSUE-3-cadastro-sincronizacao-usuarios.md)

- **Status:** ⬜ pendente
- **Depende de:** T4

## O que fazer

1. Em 403 das mutations já expostas pela UI, usar mensagem genérica de permissão no feedback local existente: despesas (criar/editar/desativar), transações (cancelar/visibilidade), guilds (convidar/remover), importações (upload/confirmar). Não criar ações novas nem alterar regras/contratos do backend.
2. Preservar tela/URL/sessão e formulário, arquivo ou revisão quando aplicável. Não fechar modal, resetar arquivo/revisão, mostrar sucesso, efetivar atualização otimista ou repetir mutation por 403. Não transformar esses erros na página de T4. Preservar comportamentos de cancelamento/troca de guild/logout e demais erros.
3. Usar `APIError.status` no consumidor que já apresenta o erro; não comparar mensagem do servidor nem alterar `SessionProvider` para interceptar todo 403. Reutilizar chaves i18n PT/EN de permissão. Evitar criar abstração genérica apenas para um check de status; qualquer pequena extração compartilhada deve ficar no módulo específico existente, sem pacote utilitário.
4. Testes **planejados** nas suítes de componentes/hooks de cada feature: rejeição por 403 com feedback específico e estado preservado, sem sucesso/replay/navegação/login. Cobrir todas as famílias de ações, incluindo distinção upload negado (arquivo mantido) e confirmação negada (revisão mantida). Reutilizar MSW/doubles existentes.
5. Atualizar evidência E2E de ao menos uma ação negada com rascunho preservado no fluxo já existente de despesas/importação. README descreve a diferença entre carregamento 403 e ação 403. A task termina com seu próprio teste/validação, sem criar task adicional de testes finais.

## Arquivos

- [alterar] `web/spa/src/features/fixed-expenses/components/fixed-expense-modal.tsx`, `fixed-expenses-page-client.tsx`, `fixed-expenses-page-client.test.tsx` — erros de salvar/desativar e preservação.
- [alterar] `web/spa/src/features/transactions/api/use-transactions.ts`, `components/transactions-page-client.test.tsx` — cancelar/visibilidade (teste criado em T4).
- [alterar] `web/spa/src/features/guilds/api/use-guilds.ts`, `components/guild-page-client.test.tsx` — convite/remoção.
- [alterar] `web/spa/src/features/imports/api/use-imports.ts`, `api/use-imports.test.tsx`, `components/import-page-client.test.tsx` — upload/confirmação e rascunho/revisão.
- [alterar] `web/spa/src/shared/lib/i18n.ts` — mensagens de ação PT/EN.
- [alterar] `web/spa/e2e/fixed-expenses.spec.ts`, `web/spa/e2e/imports.spec.ts` — casos negativos nos fluxos existentes, conforme cobertura necessária.
- [alterar] `web/spa/README.md` — limite de ações versus página.

## Critérios de aceite

- **T5-AC1:** Cada família de mutation indicada recebe mensagem genérica de permissão no canal local já usado, sem tentar inferir causa do 403 → CA11.
- **T5-AC2:** Formulário/arquivo/revisão e rota/sessão são preservados no 403 aplicável, sem sucesso, limpeza de rascunho, replay ou página negada global → CA11.
- **T5-AC3:** 401 segue o fluxo de sessão existente; demais erros e isolamento de guild/logout mantidos; a distinção consulta/ação é documentada e comprovada → CA8/CA11.

## Verificação

Na raiz, gate do README/package.json: `npm --prefix web/spa run api:types:check` → `npm --prefix web/spa run lint` → `npm --prefix web/spa run build` → `npm --prefix web/spa test` → `npm --prefix web/spa run test:e2e` → `npm --prefix web/spa run test:e2e:prod`. Node compatível/dependências, Chromium/portas disponíveis e config pública OIDC no build. Esperado: feedback e estado preservados nos casos 403; fluxos anteriores sem nova falha.

No fechamento da spec, cumprir também os gates integrados da seção **Verificação final** da principal (API, PostgreSQL isolado e login real). Essa execução consolida evidências e não substitui os testes próprios desta task. Não alegar conclusão de CA4–CA6 com PostgreSQL impedido nem login real baseado só em mocks.

Manual: não aplicável como gate separado; evidência de rascunho preservado é automatizada. Não produzir screenshots/trace contendo credenciais/tokens no teste OIDC real.

## Contexto adicional

D6: erro local em ações. T4 já define página apenas para consultas de carregamento; ambas preservam a sessão. `useImports` tem callback de confirmação que invalida transações em `onSettled` para consistência após cancelamento; preservar essa garantia e distinguir a resposta explícita 403 de cancelamento sem resposta. Não apagar o callback indiscriminadamente para passar teste. Seguir organização por feature, efeitos tardios isolados e i18n existente.

## Evidências

<!-- Preenchido pelo dev-mage: matriz ação → mensagem → estado preservado, testes/gates, verificação integrada e commit local. -->
