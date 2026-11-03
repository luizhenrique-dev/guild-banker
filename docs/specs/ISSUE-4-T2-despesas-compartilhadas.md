# ISSUE-4 — T2: Substituir despesas pessoais por despesas da guild

Spec: [ISSUE-4 — Despesas fixas compartilhadas por guild](./ISSUE-4-despesas-fixas-por-guild.md)

- **Status:** ⬜ pendente
- **Depende de:** T1; ISSUE-3 concluída conforme o pré-requisito geral.

## O que fazer

1. **Troca coesa:** implementar domínio, persistência, aplicação, HTTP, contrato, cliente e mocks no mesmo commit. Não deixar cliente chamando rota removida nem manter implementação pessoal paralela. Esta task é maior que T1 porque a mudança aprovada atravessa contrato e schema; separar seus trechos em entregas quebradas não é permitido.
2. Adicionar `GuildID int64` à entidade `FixedExpense` e ao constructor `New`, exigindo guild positiva. Preservar validações financeiras, status e auditoria; não adicionar tags de infraestrutura no domínio. Atualizar testes/factories consumidores no mesmo commit.
3. Em `000001_create_users_table.up.sql`, substituir `fixed_expense.user_id` por `guild_id BIGINT NOT NULL REFERENCES guild(id)` e o índice pessoal por `idx_fixed_expense_guild_id_status`. Preservar os demais campos/constraints. Não criar backfill ou nova migration de compatibilidade; não executar force/down/limpeza para fazer a edição reaplicar. Testar a partir de schema novo; down atual já remove a tabela completa.
4. Na porta consumidora `Storage`, propor `IsGuildMember`, `Create(ctx, expense)`, `GetByIDAndGuild`, `ListActiveByGuild` e `Update(ctx, expense)`, removendo os métodos/argumentos pessoais substituídos. Cada método da aplicação recebe guild e ator; verifica identidade/IDs/membership antes do acesso às despesas. Não membro é erro de domínio; falha técnica da verificação continua técnica e não prossegue. Sem restrição de escrita pelo criador e sem copiar privacidade de transações.
5. Adapter usa `guild_member` para membership. INSERT/mapeamentos levam `GuildID`; consultas por ID e UPDATE incluem `id` e `guild_id`, listagem inclui guild e ACTIVE e ordenação por `due_day/name`. UPDATE preserva guild/criação e atualiza autoria do ator. Nenhuma linha no escopo retorna erro de despesa não encontrada, ainda que o ID exista em outra guild. Remoção do criador de `guild_member` não apaga a despesa.
6. Registrar exclusivamente POST/GET `/api/v1/guilds/:guildID/fixed-expenses`, PATCH `.../:id` e PATCH `.../:id/deactivate`. Todos os handlers obtêm ator do principal, validam IDs positivos da URL e passam contexto da requisição. Bodies não escolhem guild/owner; remover extrator pessoal que ficar órfão. DTO de resposta inclui `guild_id` obrigatório. Respostas: 201/200/200/204; 400 validação, 403 não membro, 404 recurso fora do escopo/ausente, 500 técnico genérico. Distinguir validação de IO com classificação mínima de erros do próprio módulo; sem comparar textos nem expor erro bruto.
7. Alterar OpenAPI: paths, parâmetro guild em cada operação, descrição de compartilhamento/autorização, `FixedExpenseResponse.guild_id` obrigatório e erros coerentes. Manter operationIds, corpos financeiros e status existentes. Regenerar tipos pelo script `api:types`; ajustar fixtures tipadas simultaneamente.
8. Conectar o contexto preparado em T1 ao cliente: todas as chamadas recebem a guild capturada e constroem novas URLs com IDs seguros pelo padrão existente `safeID`. Create/update/deactivate não leem a guild selecionada posteriormente para escolher alvo/invalidação. Manter bodies sem `guild_id`, owner ou rename; PATCH existente aceita status no backend, sem inventar novo controle na UI.
9. Trocar mocks de despesas por coleção por guild, validar `mockGuildMember` em todas as operações e localizar ID dentro do escopo. Usar fixtures diferentes para guilds 1/2, garantindo que usuários 1/2 da guild 2 vejam o mesmo registro. Adaptar o teste de troca de usuário: limpar cache continua obrigatório, mas membros da mesma guild podem legitimamente ver a mesma despesa após consulta própria. Atualizar interceptadores e objetos tipados usados por página/hooks/sessão; nunca preservar teste baseado na regra pessoal antiga.
10. Atualizar título/badge/banner PT/EN e README para compartilhamento por membros, sem somar despesas nos totais nem gerar transações. Ajustar nomes acessíveis e expectativas de navegação/CRUD/foco. Manter rascunho no 403 de ação e página no 403 de consulta, inclusive refetch com cache anterior, conforme ISSUE-3/T1.
11. **Testes Go planejados:** tabelas de domínio/aplicação para guild inválida, duas identidades da mesma guild, membership falso/falha/cancelamento, quatro operações negadas, guild correta e guild errada; ampliar HTTP com principal confiável e headers conflitantes, novos paths/status/JSON, 401 e inexistência das rotas antigas autenticadas. Atualizar `TestAPIPrefixProtection` sem perder a proteção de rotas desconhecidas.
12. **Teste PostgreSQL planejado:** criar `storage_postgres_integration_test.go`, com `//go:build integration && postgres`, usando sqlx/libpq/testify já declarados. Nova env `FIXED_EXPENSE_TEST_DATABASE_URL` é exclusiva de teste. Criar schema exclusivo, configurar todas as conexões, aplicar migration inicial atualizada e inserir usuários/guilds/members fictícios. Ausência/invalidez da env no gate explícito falha, sem skip. Provar FK/NOT NULL, round-trip de `GuildID`/auditoria, mesma linha compartilhada A/B, isolamento G1/G2 e nenhum UPDATE em guild errada; incluir caso do usuário membro de ambas. Exercitar serviço e handlers reais com esse adapter e principals controlados para confirmar a matriz 403/404, sem chamar servidor com migrations automáticas. Remover A de membership, preservar a despesa, negar novas operações de A e permitir B. Limpar apenas o schema criado pelo teste.
13. **Testes SPA planejados:** `contracts.test.ts` cobre compartilhamento por membros/não membro, ID de outra guild, payloads/status e ausência de rename. Testes de página/hook cobrem respostas tardias, troca de guild, cache, 403 com dados antigos e feedback local. E2E de despesas/contrato/contexto/navegação cobre novas URLs, listas distintas, histórico, modal antigo fechado, CRUD e foco a 375/1280 px.
14. Adaptar login real em `auth.spec.ts`: após login, resolver guild de teste pela lista autenticada e, no ambiente dedicado acordado, criar pelas rotas públicas existentes caso não exista. Não depender de ID fixo nem de dados prévios; não promover automaticamente usuário a guild alheia. Abrir `/fixed-expenses?guild=<id>` e exigir 200 da rota nova/título atualizado. Documentar ambiente novo, membership e preparação pelo contrato público; sem SQL/seed do navegador, alteração de realm ou captura de tokens/credenciais. Não modificar ambiente compartilhado para satisfazer o gate.

## Arquivos

- [alterar] `api/migrations/000001_create_users_table.up.sql` — guild/FK/índice.
- [alterar] `api/internal/fixed_expense/fixed_expense.go`, `service.go`, `storage.go`, `handler.go`, `dto.go`, `errors.go` — vínculo, autorização e HTTP.
- [alterar] `api/internal/fixed_expense/fixed_expense_test.go`, `service_test.go`, `handler_test.go` — domínio/aplicação/HTTP e identidade.
- [criar] `api/internal/fixed_expense/storage_postgres_integration_test.go` — PostgreSQL real isolado e integração com serviço/handlers.
- [alterar] `api/internal/infra/webserver/routes.go`, `routes_integration_test.go` — contrato e proteção.
- [alterar] `docs/specs/openapi.yaml`, `web/spa/src/shared/api/schema.d.ts` — OpenAPI e regeneração.
- [alterar] `web/spa/src/features/fixed-expenses/api/fixed-expenses.api.ts`, `use-fixed-expenses.ts` — transporte por guild usando contexto de T1.
- [alterar] `web/spa/src/features/fixed-expenses/api/use-fixed-expenses.test.tsx` — arquivo criado em T1, adaptar ao contrato novo.
- [alterar] `web/spa/src/features/fixed-expenses/components/fixed-expense-modal.tsx`, `fixed-expenses-page-client.tsx`, `fixed-expenses-page-client.test.tsx` — captura de alvo, feedback e contrato.
- [alterar] `web/spa/src/app/app.test.tsx`, `session-provider.test.tsx` — regressões da guarda/cache/compartilhamento.
- [alterar] `web/spa/src/mocks/fixtures/fixed-expenses.ts`, `web/spa/src/mocks/handlers/fixed-expenses.ts`, `contracts.test.ts` — dados e acesso por guild.
- [alterar] `web/spa/src/shared/lib/i18n.ts`, `web/spa/README.md` — linguagem PT/EN e operação com schema novo.
- [alterar] `web/spa/e2e/fixed-expenses.spec.ts`, `fixed-expense-contract.spec.ts`, `guild-context.spec.ts`, `navigation.spec.ts`, `auth.spec.ts` — fluxos demo e login real adaptados.

Paths existentes foram conferidos no planejamento; `use-fixed-expenses.test.tsx` será criado em T1. Revalidar também testes PostgreSQL da ISSUE-3 depois do pré-requisito, sem reescrever sua regra histórica pessoal nos documentos anteriores.

## Critérios de aceite

- **T2-AC1:** Entidade, SQL e DTO exigem e preservam guild única; schema novo sem owner pessoal nem backfill → CA1/CA8.
- **T2-AC2:** A/B membros da mesma guild operam a mesma linha; não membro/removido recebe 403 e membro da guild errada recebe 404 sem alteração, mesmo pertencendo às duas → CA2/CA3.
- **T2-AC3:** Rotas pessoais removidas e novas operações coerentes em API/OpenAPI/tipos/cliente/mocks, sem alias; testes autenticados/sem token distinguem 404/401 → CA4/CA11.
- **T2-AC4:** Autoria/criação estáveis, alteração atribuída ao ator correto, headers conflitantes ignorados e remoção do criador não remove despesa → CA5.
- **T2-AC5:** Guarda/cache/contexto de T1 operam sobre registros realmente filtrados por guild, com 403 sem dados antigos, ações sem replay e resposta tardia sem efeitos cruzados → CA6/CA7/CA10.
- **T2-AC6:** Validações, decimal, nome read-only, ACTIVE, PAUSED/CANCELLED, payloads/foco e totais preservados; linguagem PT/EN atualizada → CA2/CA9.
- **T2-AC7:** IO não autoriza; validação/403/404/500 distinguíveis e genéricos quando necessário; contexto propagado sem segredos em saída → CA10.
- **T2-AC8:** PostgreSQL/HTTP real controlado prova isolamento e auditoria; demo prova UX/contrato; suíte real de login usa guild e rota novas, com premissas de ambiente documentadas → CA8/CA11.

## Verificação

Na raiz: `make api/vet` → `make api/lint` → `make api/test`; complementar com `make api/build` e `make api/test/integration`. Go conforme `api/go.mod`, dependências/ferramentas e compilador C disponíveis. Esperado: domínio, aplicação, HTTP e rotas passando, sem dependência de PostgreSQL no gate offline.

**Novo gate proposto:** em `api/`, `go test ./internal/fixed_expense -race -count=1 -tags=integration,postgres`, com `FIXED_EXPENSE_TEST_DATABASE_URL` por ambiente. Executar também o gate amplo previsto na ISSUE-3: `go test ./... -race -count=1 -tags=integration,postgres`, exigindo a configuração `USER_TEST_DATABASE_URL` daquele pré-requisito. Presença de PostgreSQL não garante DSN/schema: falhar explicitamente no gate solicitado quando indisponível, sem substituir prova real por mock. Não imprimir DSNs; não limpar tabelas compartilhadas.

Na raiz, regenerar por `npm --prefix web/spa run api:types`; então `npm --prefix web/spa run api:types:check` → `npm --prefix web/spa run lint` → `npm --prefix web/spa run build` → `npm --prefix web/spa test` → `npm --prefix web/spa run test:e2e` → `npm --prefix web/spa run test:e2e:prod`. Node compatível, dependências já declaradas instaladas, Chromium/portas e configuração pública OIDC no build.

Fechar com `npm --prefix web/spa run test:e2e:auth`, conforme README/config existente: Keycloak local, API de T2/schema novo, usuário de teste e guild criada/resolvida no ambiente dedicado, senha via ambiente, sem artefatos de autenticação. O gate prova login e consulta real; compartilhamento por dois membros é comprovado pela integração da feature com banco/principals distintos.

Manual: opcionalmente consultar a mesma guild em duas sessões de teste após refetch; não substitui CA3/CA8 automatizados. Não iniciar servidor real em banco desconhecido: `main` aplica migrations no startup. Editar a migration inicial não migra schema já versionado.

## Contexto adicional

D1–D4: todos os membros operam, sem dados a preservar, quebra de rotas aprovada e ISSUE-3 anterior. A spec principal contém contrato, erros e matriz completa CA → evidência. Reusar `transaction.Service.authorizeGuildMember` como referência de direção de dependências, sem importar serviço de guild ou copiar owner/visibilidade. Não expandir para totais, geração recorrente, UI de roles ou infraestrutura nova.

## Evidências

<!-- Preenchido pelo dev-mage: matriz ator/guild/operação/status, round-trip PostgreSQL/auditoria, contratos gerados, contextos UI, gates reais/demo, limitações e commit local. -->
