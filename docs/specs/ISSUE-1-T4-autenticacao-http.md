# ISSUE-1 — T4: Ativação HTTP e abandono dos headers

Spec: [ISSUE-1-autenticacao-jwt-rs256.md](./ISSUE-1-autenticacao-jwt-rs256.md)

- **Status:** ✅ concluida
- **Depende de:** T1, T2, T3

## O que fazer

1. Substituir Auth placeholder por middleware com autenticador/logger injetados. Parsing Bearer rejeita ausência, token vazio, scheme indevido e credenciais malformadas/duplicadas; não aceitar token em query/cookie. Derivar ctx da request com timeout e sempre liberar cancel.
2. Mapear erros: token ausente/inválido/expirado ou chave indisponível → 401 unauthorized e WWW-Authenticate Bearer genérico; usuário inexistente/inativo → 403 forbidden; lookup técnico → 500 internal server error. Todos usam {"error":"texto"} sem causa/claims. Emitir um log slog categorizado por rejeição, sem erro bruto nem PII.
3. Implementar RequireRole separado: principal ausente 401, role ausente 403, role exata presente segue. Aplicar somente em rota registrada nos testes; rotas de negócio não exigem member/admin.
4. Fazer bootstrap criar provider/verifier e auth.Service antes de escutar, reaproveitando user.Service/NewStorage existentes e cliente HTTP limitado. Injetar serviço/logger/timeout no webserver por construção explícita. Discovery/config inválidos propagam erro e impedem startup. Nenhum cache/verifier por request.
5. Proteger a fronteira /api ou /api/ também no NoRoute, depois do CORS existente; /health público e /apix não correspondem. Evitar aplicar Auth duas vezes ao grupo /api/v1. Preflight continua tratado pelo CORS antes de Auth; OPTIONS sem preflight não recebe bypass indiscriminado.
6. No MESMO commit de ativação, trocar getRequesterActor/getRequesterUserID nos quatro handlers para principal do request.Context. Nunca fallback para X-User-ID/X-User-Email nem sobrescrita desses headers pelo middleware. Ausência de principal →401 sanitizado. Preservar serviços/DTOs de negócio e suas políticas já existentes.
7. Atualizar OpenAPI: bearer real, 401/403 de autenticação/acesso local, remover placeholder 501 e headers de identidade exigidos. Preservar ErrorResponse.error string e demais contratos. Atualizar somente as lacunas de auth/identidade resolvidas em OPENAPI_GAPS; regenerar schema.d.ts por script existente. SPA/demo e mocks não mudam.
8. Testes planejados: TestAuth (parsing/abort/envelope/logs), TestRequireRole; TestAuthAcceptance (HTTP CA1–CA8 com fixtures controladas), TestAPIPrefixProtection; TestAuthenticatedHandlerIdentity em cada família de handler. TestAuthAcceptance fica em infra/keycloak/verifier_http_integration_test.go, para usar relógio fixo privado de T2 e importar middleware/auth.Service sem duplicar validação; testes de prefixo ficam no webserver. Request com token A e headers ID/Email de B continua usando Actor local A. Sem principal, headers não liberam handler. Token inválido não chama lookup/handler; válido/ativo disponibiliza claims por request.Context.
9. Capturar logs em buffer nos testes com token/username/email fictícios reconhecíveis; provar ausência desses valores e do erro bruto. Testar falha de bootstrap sem iniciar listener e propagação de cancelamento. Integrações HTTP/JWKS recebem build tag integration e não dependem de banco ativo.

## Arquivos

- [alterar] `api/internal/infra/webserver/middleware/auth.go`.
- [criar] `api/internal/infra/webserver/middleware/auth_test.go`, `auth_integration_test.go`, `roles.go`, `roles_test.go`.
- [criar] `api/internal/infra/keycloak/verifier_http_integration_test.go` — aceitação HTTP com verifier real e relógio privado de teste.
- [alterar] `api/internal/infra/webserver/server.go`, `routes.go`, `api/cmd/server/main.go`.
- [criar] `api/internal/infra/webserver/server_integration_test.go`, `routes_integration_test.go`.
- [alterar] `api/internal/guild/handler.go`, `api/internal/fixed_expense/handler.go`, `api/internal/transaction/handler.go`, `api/internal/importer/handler.go`.
- [criar] `handler_test.go` em cada um dos quatro diretórios acima.
- [alterar] `docs/specs/openapi.yaml`, `docs/specs/OPENAPI_GAPS.md`, `web/spa/src/shared/api/schema.d.ts` (gerado).

## Critérios de aceite

- T4-AC1 HTTP sem token/invalidade intercepta /api/* em requests sem Origin ou com origem autorizada, inclusive NoRoute; health público e /apix fora da fronteira → CA1, CA3, CA4, CA5, CA8, CA14.
- T4-AC2 Token válido/usuário ativo recebe 200 no handler de prova, com claims e principal em request.Context → CA2.
- T4-AC3 HTTP cobre rotação A→B no mesmo verifier e rejeição de kid desconhecido → CA7.
- T4-AC4 Role guard prova401/403/200 em rota exclusiva dos testes; nenhuma nova role nas rotas de negócio → CA6.
- T4-AC5 As quatro famílias usam ID/Email locais e ignoram headers conflitantes; token inválido não consulta usuário; inexistente/inativo 403 e lookup técnico 500 → CA10.
- T4-AC6 Erros são genéricos/consistentes; logs têm reason controlado e nenhum token/PII/erro bruto; cancelamento/timeout preservados → CA13.
- T4-AC7 API não inicia com discovery/config inválidos; principal/adapter são injetados e verifier reutilizado → CA12.
- T4-AC8 OpenAPI/tipos deixam de exigir headers e mencionar 501; scripts de drift/compilação passam; HTTP CA1–CA8 entra na suíte local → CA9 local.

## Verificação

- Na raiz: `make api/vet`, `make api/lint`, `make api/test`, `make api/test/integration`, `make api/build` — integration criado em T2; Go compatível/race necessários. Esperado: todos os cenários HTTP, regressões dos quatro handlers e bootstrap passam sem banco/Keycloak ativos.
- `npm --prefix web/spa run api:types`, `make web/types`, `make web/build` — scripts existentes; Node compatível >=22.12/dependências instaladas. Esperado: tipos regenerados sem drift e consumidores compilam.
- Manual: inspeção das rotas garante ausência de guards de role em produção. Não executar binário que migra banco no startup para comprovar testes de auth.

## Contexto adicional

D2/D3/D6/D7 aprovadas. Esta é a troca atômica: antes T4 a API permanece 401/501; depois T4 não existe handler autenticado usando identidade fornecida pelo cliente. CORS ainda recebe a restrição de origens em T5; testes desta task não alegam allowlist já restrita. Construtor/assinaturas internas podem receber dependências necessárias; não criar service locator global ou refatorar serviços de negócio. Ausência de integração OIDC na SPA e headers ainda usados na demo são limitações registradas, não tarefa de UI aqui.

## Evidências

T3 concluída no commit `cd4be4f`.

| Critério | Evidência observada | Resultado |
|---|---|---|
| T4-AC1 | TestAuth, TestAPIPrefixProtection, TestAuthAcceptance; inclui NoRoute, OPTIONS sem preflight e barras finais GET/POST | PASS com race |
| T4-AC2 | TestAuthAcceptance: claims e Actor no handler de prova | PASS com race |
| T4-AC3 | TestAuthHTTPRotation: requests A/B/C no mesmo verifier | PASS com race |
| T4-AC4 | TestRequireRole: 401/403/200; routes.go sem RequireRole | PASS com race + inspeção |
| T4-AC5 | TestAuthenticatedHandlerIdentity e TestHandlerUsesTrustedActor nas quatro famílias; TestAuthAcceptance lookup/403/500 | PASS com race |
| T4-AC6 | TestAuth: logs JSON/envelope/contexto ativo/deadline original; TestAuthTimeoutAndCancellation | PASS com race |
| T4-AC7 | TestBootstrapDiscoveryFailure sem listener/banco | PASS com race |
| T4-AC8 | api:types, api:types:check, build SPA; 22 operações protegidas com401/403/500 e sem501/headers; TestAuthAcceptance | PASS |

TDD observado: helpers antigos confiavam nos headers e falharam nas quatro famílias; middleware placeholder falhou nos cenários válidos; OPTIONS comum retornava204; Gin redirecionava barras finais com301/307. Após as correções, os cenários passam. RedirectTrailingSlash foi desabilitado; CORS só trata OPTIONS como preflight quando Access-Control-Request-Method está presente. gin.New + Recovery evita o access log padrão com URL/IP, mantendo o log sanitizado de rejeição.

Gates incrementais executados em `api`: testes com `-race -count=1 -tags=integration` de keycloak, webserver/middleware, cmd/server e quatro famílias de handlers; vet e golangci-lint com integration nos mesmos componentes (0 issues); build de cmd/server em /tmp. SPA: regeneração, drift e build passaram; Vite emitiu aviso de bundle >500kB. Gates globais serão executados no fechamento T6. Revisão independente aprovou T4 após corrigir redirect e adicionar regressão do contexto. Nenhuma falha detectada nas verificações executadas.

`server_integration_test.go` será criado em T5 para CORS; prefixo está em `routes_integration_test.go` nesta task.

Fechamento global reabriu o gate de lint: sem a tag integration, dois getRequesterUserID ficaram órfãos em importer/transaction após a troca de getRequesterActor. Foram removidos somente esses helpers criados como órfãos pela mudança, mantendo as assertions do Actor/ID local e os testes que capturam storage. `make api/vet api/lint api/test api/build` passou em toda a API (0 issues); testes de identidade dos dois handlers passaram com race; lint integration global passou (0 issues). Revisão independente da correção aprovada. Correção registrada em commit local separado, sem reescrever T4.
