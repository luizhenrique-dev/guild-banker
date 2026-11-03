# ISSUE-4 — T1: Preparar contexto de guild nas despesas da SPA

Spec: [ISSUE-4 — Despesas fixas compartilhadas por guild](./ISSUE-4-despesas-fixas-por-guild.md)

- **Status:** concluida
- **Depende de:** nenhuma task interna; ISSUE-3 concluída, incluindo T4/T5 de página 403 e ações negadas.

## O que fazer

1. Conferir o resultado da ISSUE-3 antes de editar. Aplicar em `/fixed-expenses` a mesma guarda de guild das páginas dependentes em `AppRoutes`, preservando a URL da tela, `?guild=`, estados de carregamento/vazio/seleção inválida e prioridade de 403 sobre cache antigo. Home continua disponível. Isso substitui intencionalmente a exceção para despesas pessoais do CA10 da ISSUE-3.
2. Fazer `fixedExpenseKeys.list` incluir usuário e guild. `useFixedExpenses` usa a guild selecionada e só habilita com sessão e seleção válidas; conservar cancelamento por `AbortSignal`, políticas de retry e comportamento 401/403 existentes. Não reutilizar dados anteriores como placeholder de outra guild.
3. Capturar `guildId` nas variáveis de criação/edição/desativação no momento da ação. Adaptar o wrapper de mutation, formulário e confirmação para invalidar somente a chave da guild de origem. A troca de guild remonta o conteúdo pelo ID e fecha modais/confirmações. Resposta tardia não fecha/limpa formulário, invalida dados nem anuncia sucesso na nova guild; logout continua anulando efeitos tardios da sessão.
4. **Limite deliberado de T1:** manter o HTTP pessoal, schema, resposta, mocks por usuário e textos financeiros atuais até a troca conjunta de T2. A guild já organiza guarda/cache/estado, mas ainda não determina quais registros o servidor entrega. Não anunciar isolamento de registros ou compartilhamento antes de T2; não criar rotas temporárias ou compatibilidade.
5. Adaptar os testes existentes que assumem renderização imediata antes de carregar guilds ou mutations sem guild. O teste de troca de usuário continua usando a listagem pessoal até T2; preservar sua garantia de limpeza de sessão. Testar chaves/contexto com respostas controladas, sem fingir que o endpoint pessoal já filtra guild.
6. Testes **planejados**: nova suíte `use-fixed-expenses.test.tsx` para habilitação/chaves, guild capturada e resposta tardia após troca; ampliar `app.test.tsx` e testes da página para guild vazia/inválida/403, ausência de consulta e reset dos modais; ampliar `guild-context.spec.ts` para navegação/histórico/contexto de despesas. Conservar casos da ISSUE-3: 403 de carregamento mostra página e 403 de ação preserva rascunho/sessão, sem replay.
7. Adaptar desde T1 a preparação do teste de login real: depois de autenticar, consultar as guilds pelas rotas públicas e resolver uma guild de teste; caso necessário, criá-la somente no ambiente dedicado acordado. Abrir `/fixed-expenses?guild=<id>`. Até T2, a consulta financeira esperada ainda é a rota pessoal e o título pessoal. Não exigir que um usuário recém-provisionado pela ISSUE-3 já tenha membership, não depender de ID fixo nem executar SQL/seed/alteração de realm. Essa preparação permite que a nova guarda não quebre a suíte real antes da substituição do HTTP.

## Arquivos

- [alterar] `web/spa/src/app/app.tsx`, `web/spa/src/app/app.test.tsx` — guarda e estados de guild.
- [alterar] `web/spa/src/app/session-provider.test.tsx` — keys, guild nas variáveis e garantias de sessão.
- [alterar] `web/spa/src/features/fixed-expenses/api/use-fixed-expenses.ts` — query/contexto/invalidação.
- [criar] `web/spa/src/features/fixed-expenses/api/use-fixed-expenses.test.tsx` — testes de contexto.
- [alterar] `web/spa/src/features/fixed-expenses/components/fixed-expense-modal.tsx` — guild capturada no envio.
- [alterar] `web/spa/src/features/fixed-expenses/components/fixed-expenses-page-client.tsx`, `fixed-expenses-page-client.test.tsx` — confirmação/contexto e resets.
- [alterar] `web/spa/e2e/guild-context.spec.ts` — contexto e histórico da tela de despesas, sem alegar filtro no servidor pessoal.
- [alterar] `web/spa/e2e/auth.spec.ts` — resolver/criar guild de teste no ambiente dedicado e selecionar seu ID, mantendo endpoint/título pessoais até T2.
- [alterar] `web/spa/README.md` — despesas agora exigem seleção de guild; registrar que a substituição do vínculo/API ocorre em T2.

Paths existentes conferidos no planejamento; reconferir após ISSUE-3. Reutilizar `useActiveGuild`, `GuildSelectionState` e a página 403 criada naquele pré-requisito, sem nova implementação paralela.

## Critérios de aceite

- **T1-AC1:** Rota de despesas usa guarda de guild; sem sessão/seleção válida ou com 403 de guilds não faz consulta de despesas. Home/menu/Sair e distinção vazio/inválido/403 continuam corretos → CA6.
- **T1-AC2:** Keys incluem usuário e guild; troca de contexto não reaproveita entrada de outra guild e remonta/fecha modal e confirmação → parte de CA7. Filtro real de registros será comprovado em T2.
- **T1-AC3:** Mutation captura guild de origem, invalida só sua chave e não afeta UI/feedback de outra guild; logout/troca de usuário mantém limpeza e bloqueio de resposta tardia → CA7/CA11.
- **T1-AC4:** Testes de componente/hook/E2E próprios passam com o contrato HTTP atual e preservam o tratamento 403 implementado na ISSUE-3 → CA6/CA11.

## Verificação

Na raiz, gate confirmado pelo README/package.json: `npm --prefix web/spa run api:types:check` → `npm --prefix web/spa run lint` → `npm --prefix web/spa run build` → `npm --prefix web/spa test` → `npm --prefix web/spa run test:e2e` → `npm --prefix web/spa run test:e2e:prod`.

Pré-requisitos: ISSUE-3 concluída, Node 24.6.0/compatível, dependências existentes instaladas por `npm --prefix web/spa ci`, Chromium/portas livres e variáveis públicas OIDC no build conforme README. Esperado: guarda/cache/estado validados e suítes existentes sem regressão. Nenhum código Go/SQL/contrato muda nesta task; gates de alteração da API/PostgreSQL não se aplicam a T1.

Reexecutar `npm --prefix web/spa run test:e2e:auth`, na raiz, com Keycloak/API da ISSUE-3 disponíveis, Vite 5173 livre, credenciais fictícias por ambiente e guild de teste resolvida/criada no ambiente dedicado. Login real mantém o contrato pessoal até T2; a preparação de guild é a adaptação desta task, sem mudança da política de IAM. Não imprimir tokens nem habilitar screenshots/trace/vídeo de autenticação.

Manual: não é gate adicional; navegação, fechamento de modal e respostas tardias devem ser comprovados automaticamente. Não atribuir aos testes demo evidência de autorização real do backend.

## Contexto adicional

D1/D4: destino final é guild compartilhada após ISSUE-3. Referências existentes: `AppRoutes.guildPage` remonta por ID; `transactionKeys` e mutations de transações exemplificam origem nas variáveis; `SessionProvider.request` controla a vida da sessão. Não alterar `SessionProvider` de produção para acoplar guild ao login. Não mudar totais/textos para compartilhamento antecipadamente, nem incluir owner/visibilidade.

## Evidências

### Checkpoint — 2026-11-03

Working tree inicial limpo. ISSUE-3 concluída conforme seu índice; implementação autorizada pelo pedido da ISSUE-4. Nenhuma mudança Go/SQL/HTTP pessoal nesta task.

| Critério | Evidência observada | Estado |
|---|---|---|
| T1-AC1 | `app.test.tsx`: ausência de sessão e 403 guilds não consultam despesas; shell/Home/Sair preservados. Hook: seleção inválida não consulta. | aprovado |
| T1-AC2 | Keys por usuário/guild; teste de página fecha confirmação na troca; E2E fecha modal e restaura histórico. | aprovado |
| T1-AC3 | Teste de página resolve create G1 com rascunho G2 aberto: mantém dialog/rascunho, sem toast, invalida somente G1. Hook idem após desmontagem; sessão bloqueia resposta após logout e troca de usuário limpa dados. | aprovado |
| T1-AC4 | Vitest 131, E2E demo 15, smoke produção 8 e login real 6 passaram. | aprovado |

Comandos na raiz: `npm --prefix web/spa run api:types:check` passou sem drift; `lint` passou; build com OIDC público passou (aviso de chunk >500 kB); `test` passou, 17 arquivos/131 testes; `test:e2e` passou, 15 testes; `test:e2e:prod` passou, 8 testes. Baseline direcionado: 40 testes passaram. RED observado: dois novos casos 403 guilds falharam antes da guarda. Primeira execução E2E nova falhou porque modal impede interação com fundo; corrigido sensor para navegação via histórico, sem alterar produto.

Login real: Keycloak local disponível; API iniciada exclusivamente no container descartável `guildbanker-issue4-test`, porta 55434, sem volumes. `E2E_ALLOW_GUILD_CREATE=true npm --prefix web/spa run test:e2e:auth`: 5 passaram, 1 falhou no POST público de guild (400). Investigação independente em andamento. Usuário autorizou explicitamente correção mínima com teste do fluxo de guild em 2026-11-03. Não iniciar T2 antes de resolver este gate.

Revisão independente `review_t1`: sem defeito funcional de produção confirmado; lacunas de testes de resposta tardia/confirmação e documentação foram atendidas; rechecagem solicitada. Commit pendente até validar login real e correção autorizada.


Fechamento T1: correção D5 autorizada incorporada ao adapter de guild: `GetContext` consome o RETURNING antes da próxima instrução, preservando SQL parametrizado, contexto e transação. Teste real `TestStorageCreatePostgres` observou RED por erro de protocolo e GREEN após correção; verifica guild/membership e rollback em falha de FK. Gate exclusivo: em `api/`, `GUILD_TEST_DATABASE_URL=<URL de teste> go test ./internal/guild -run TestStorageCreatePostgres -race -v -count=1 -tags=integration,postgres` passou; variável ausente falha explicitamente. Esta variável complementa os gates PostgreSQL globais.

`make api/vet`, `api/lint` (0 issues), `api/test`, `api/test/integration`, `api/build` passaram. Linter no sandbox inicialmente não carregou módulo; reexecução com acesso à toolchain passou. `E2E_ALLOW_GUILD_CREATE=true npm --prefix web/spa run test:e2e:auth` passou, 6 testes, API corrigida no banco exclusivo. Revisão independente de T1/D5 aprovou produção; ajuste de URL do DSN solicitado foi aplicado e seu GREEN observado. Nenhuma falha detectada nas verificações executadas. Implementação validada; commit local integra este fechamento.
