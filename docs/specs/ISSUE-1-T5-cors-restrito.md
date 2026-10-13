# ISSUE-1 — T5: CORS restrito e preflight

Spec: [ISSUE-1-autenticacao-jwt-rs256.md](./ISSUE-1-autenticacao-jwt-rs256.md)

- **Status:** ✅ concluida
- **Depende de:** T4

## O que fazer

1. Acrescentar CORS_ALLOWED_ORIGINS ao config.Load, lista por vírgulas de origens completas. Remover espaços externos, validar scheme/host/port e ausência de path/query/fragment, wildcard e origem vazia. Configuração obrigatória na versão final; retornar erro sanitizado quando inválida.
2. Example local contém http://localhost:5173,http://127.0.0.1:5173. Domínio de produção é operacional e não foi informado; não inventá-lo nem incluir wildcard. Preview4173 é permitido somente se explicitamente acrescentado por quem configura ambiente.
3. Alterar Server para usar allowlist, AllowCredentials=false e headers Origin/Content-Type/Authorization, mantendo métodos de negócio existentes e OPTIONS de preflight. Não permitir X-User-ID/X-User-Email. CORS permanece global e anterior à autenticação de prefixo.
4. Testes planejados: TestLoadCORSOrigins (válida, ausente, wildcard, origem inválida); TestServerCORS (origens permitidas/externas, preflight com Authorization sem token, ausência de Allow-Credentials, request efetiva sem token401). Testes de config unitários; HTTP com tag integration, fake autenticador quando necessário, sem DB/listener de produção.
5. Provar que preflight permitido não chama autenticador e que GET/POST efetivo ainda exige JWT. Origem não autorizada não recebe Access-Control-Allow-Origin liberando acesso, mesmo com token ausente: CORS pode rejeitá-la com 403 antes de Auth. Não confundir esse status com rejeição de usuário/role ou esperar 401 nesse caso. Request sem Origin é tratada pela autenticação normal.

## Arquivos

- [alterar] `api/config/config.go`, `api/config.env.example` — allowlist.
- [alterar] `api/config/config_test.go` — arquivo criado em T2.
- [alterar] `api/internal/infra/webserver/server.go`; [criar] `server_integration_test.go` — teste de CORS separado do teste de prefixo de T4.

## Critérios de aceite

- T5-AC1 Origem da UI vem apenas da allowlist/env; wildcard/config vazia/inválida não inicia API → CA11, CA12.
- T5-AC2 Preflight autorizado aceita Authorization sem JWT e não consulta auth/usuário; efetiva sem token continua 401 → CA1, CA11.
- T5-AC3 Origem externa não obtém liberação CORS, credenciais/cookies e headers antigos de identidade não são habilitados → CA11.
- T5-AC4 Health público e CA1–CA8 permanecem passando após mudar a ordem/configuração do middleware → CA1, CA9 local.

## Verificação

- Na raiz: `make api/vet`, `make api/lint`, `make api/test`, `make api/test/integration`, `make api/build` — Go compatível/race necessários; integration criado em T2. Esperado: config/CORS/preflight e cenários de auth passam sem serviços externos.
- Manual: não aplicável para domínio real da UI, ainda não informado. Configurá-lo explicitamente no ambiente será pré-requisito operacional, não evidência presumida.

## Contexto adicional

D7 aprovou allowlist e preflight antes de Auth. Evidência UI local: web/spa/README.md documenta 5173 e Playwright usa 127.0.0.1:5173. Backend utiliza Bearer, não cookies. SessionProvider demo continua enviando headers antigos para MSW; integração frontend cross-origin real é outra issue e precisará removê-los. Esta task não modifica UI/mocks nem interpreta CORS como substituto de autenticação.

## Evidências

T4 concluída no commit `95c81aa`.

| Critério | Evidência observada | Resultado |
|---|---|---|
| T5-AC1 | TestLoadCORSOrigins: origens válidas, trimming/normalização, ausente/wildcard/path/query/fragment/credenciais/portas inválidas com erro sanitizado; TestServerCORSLoadedOrigins | PASS com race |
| T5-AC2 | TestServerCORS: preflight GET/POST204 com Authorization, zero chamadas ao autenticador; efetivas401 | PASS com race |
| T5-AC3 | Origem externa/preview403 sem ACAO; nenhuma credencial ou X-User nos headers permitidos | PASS com race |
| T5-AC4 | TestAPIPrefixProtection, TestAuthAcceptance CA1–CA8 e TestAuthHTTPRotation; health público | PASS com race |

RED observado: Load aceitava origens ausentes/inválidas; CORS retornava wildcard e credentials=true, liberando origem externa. Após implementar allowlist, revisão independente encontrou scheme maiúsculo causando panic em cors.New: regressão RED do valor normalizado, correção via URL.String e teste Load→NewServer sem panic passaram. Revisão independente final aprovada.

Gates incrementais: `go test -v -race -count=1 -tags=integration` de config, webserver/middleware, keycloak e cmd/server; vet e golangci-lint integration em config/webserver (0 issues); build cmd/server em /tmp. Nenhuma falha detectada nas verificações executadas. Gates globais permanecem para T6. Domínio de produção e preview devem ser acrescentados explicitamente no env operacional.
