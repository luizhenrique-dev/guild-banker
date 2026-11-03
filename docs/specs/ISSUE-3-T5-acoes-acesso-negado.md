# ISSUE-3 — T5: Tratar permissão negada nas ações existentes

Spec: [ISSUE-3](./ISSUE-3-cadastro-sincronizacao-usuarios.md)

- **Status:** concluida
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

| Critério | Evidência (execuções abaixo) | Resultado |
|---|---|---|
| T5-AC1 | Despesas criar/editar/desativar, cancelar/visibilidade, convite/remoção, upload/confirmação: mensagem local | aprovado — evidências observadas abaixo |
| T5-AC2 | Draft/arquivo/revisão, cache, sessão e rota preservados sem sucesso/replay | aprovado — evidências observadas abaixo |
| T5-AC3 | Regressões 401/cancelamento/troca de guild; README, E2E e gates integrados | aprovado — evidências observadas abaixo |

### Evidências observadas — fechamento 2026-11-03

- RED por contrato: todas as famílias apresentavam mensagem de erro genérica no lugar de permissão; o teste do hook de confirmação constatou cache financeiro invalidado no 403 explícito. Falhas de preparação dos novos testes foram corrigidas antes de observar RED comportamental: arquivo de teste composto incorretamente por script (restaurado de HEAD mantendo os testes anteriores), opções exact próprias de Playwright removidas de Testing Library (name string já é exato), timeout de leitura de transações passou a respeitar delays sequenciais das fixtures de guild+transações. Nenhuma asserção de comportamento foi relaxada.
- T5-AC1: casos create/edit/deactivate de despesas, cancel/visibility de transações, invite/remove de guilds e upload/confirm de importações passaram com mensagem genérica local. Modal também comprova tradução PT→EN; erro bruto não exibido.
- T5-AC2: formularios, valor alterado, diálogo de confirmação, arquivo e revisão permanecem; testes observam uma chamada por ação, ausência de sucesso e página/estado de sessão preservados. Hook confirma cache/revisão intactos após 403 explícito, sem invalidation/replay.
- T5-AC3: testes existentes de 401/403/isolamento/logout e cancelamento/troca de guild passaram. onSettled continua invalidando a guild original quando não há resposta definitiva; não foi apagado. Revisão independente pediu conservar cenário anterior upload500/confirm400; mantido junto ao novo upload403/confirm403, ambos passaram.
- GREEN focado: **25 testes/5 arquivos** passaram. Revisão independente `query_patterns` executou **40 testes/6 arquivos** incluindo sessão e não encontrou defeito funcional; recomendação de cobertura atendida.
- Gates web finais: api:types:check passou; lint sem warnings; build OIDC/TypeScript passou; **125 testes/16 arquivos** passaram. Primeiro build identificou opções exact inválidas nos testes, corrigidas, depois build/lint/suíte repetidos com sucesso. Avisos preexistentes de shim asdf e tamanho de bundle não impediram execução.
- E2E demo completo: **14 passaram**, incluindo rascunho de despesa e revisão de importação preservados em 403; smoke compilado **8 passaram**; OIDC real final **6 passaram**, sem artefatos de autenticação. API temporária apontou exclusivamente para PostgreSQL descartável, sem seed prévio de usuário local.
- Gates integrados finais API: `make api/vet api/lint api/build api/test api/test/integration` passou, lint **0 issues**, duas suítes race aprovadas. `go test ./... -race -count=1 -tags=integration,postgres` com URL do banco isolado passou em todos os pacotes, incluindo constraints/concorrência, composição HTTP e lifecycle de migrations.
- Revisão própria final de escopo, erros, dependências e preservação do preexistente: nenhuma falha detectada nas verificações executadas. Critérios AC1–AC3 aprovados; sem bloqueios ou critérios manuais pendentes. Commit local deste fechamento preparado; push/PR a cargo do usuário.

Ambiente final: API temporária encerrada graciosamente e container PostgreSQL exclusivo parado; ambiente preexistente preservado.

Revisão final organizou imports dos testes tocados; lint e 21 testes/4 arquivos afetados passaram novamente. Sem mudança de comportamento após os gates integrados.
