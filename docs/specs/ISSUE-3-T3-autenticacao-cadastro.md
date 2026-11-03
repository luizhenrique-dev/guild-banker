# ISSUE-3 — T3: Integrar cadastro ao fluxo de autenticação

Spec: [ISSUE-3](./ISSUE-3-cadastro-sincronizacao-usuarios.md)

- **Status:** ⬜ pendente
- **Depende de:** T2

## O que fazer

1. Em `auth.Service.Authenticate`, depois de verificar o token, exigir `Name`/`Email` não brancos e `EmailVerified=true`, antes de qualquer chamada a user. Ausência/vazio/não verificado → `ErrForbidden`; tipo malformado já vem como 401 do verifier. Sem fallback nem preservação parcial para cadastro existente.
2. Substituir a porta consumidora `userLookup` por operação `EnsureAccount(ctx, claims.UserID, claims.Name, claims.Email)` de T2. Não criar dependência user → auth. Formar `Principal.Actor` com ID/e-mail do usuário efetivamente persistido/retornado, nunca username, header ou ID externo convertido. Preservar criação de serviços explícita existente em `cmd/server/main.go:newServer`.
3. Mapear desativado e colisão de e-mail para `RejectionError`/403; motivos controlados novos para perfil incompleto/não verificado/conflito, sem incluir dados. IO/cancelamento/timeout na persistência → erro técnico, 500 no fluxo atual, sem liberar handler. Preservar envelope genérico, um log sanitizado e comportamento de 401/roles/CORS.
4. Atualizar testes/doubles no mesmo commit. Tokens sintéticos de sucesso passam a conter perfil completo/verificado. Fixture Keycloak antiga permanece original e é caso positivo de verificação criptográfica e negativo de autenticação por perfil ausente (403); expiração continua 401. Não enfraquecer o teste criptográfico nem editar os bytes assinados.
5. Testes **planejados**: extensão de `TestServiceAuthenticate`, `TestAuthAcceptance`, `TestAuth`, `TestAuthKeycloakFixture`; novo `TestProvisioningHTTPPostgres` em tag `integration && postgres`. Último compõe serviço auth, user/storage reais e middleware com verifier controlado, cobrindo cadastro/diff/principal/403/500 e não invocação do handler em falha. Os testes de Keycloak existentes provam separadamente assinatura/extração.
6. Atualizar OpenAPI (`bearerAuth`, `Forbidden`) e READMEs operacionais: cadastro não exige seed; nome/e-mail/verificação são obrigatórios; perfil é snapshot do token e pode regredir entre tokens válidos sem ordenação de versão; usuário desativado permanece negado. Não dizer que login em si garante sucesso em toda rota. Regenerar `schema.d.ts` pelo script existente, sem mudar contratos financeiros. Preservar textos não relacionados e evidências históricas das ISSUE-1/2.

## Arquivos

- [alterar] `api/internal/auth/service.go`, `service_test.go`, `errors.go`, `errors_test.go` — nova resolução e recusas.
- [alterar] `api/internal/infra/keycloak/verifier_test.go` — perfil das fixtures sintéticas compartilhadas.
- [alterar] `api/internal/infra/keycloak/verifier_http_integration_test.go` — doubles e cenários HTTP.
- [alterar] `api/internal/infra/keycloak/verifier_keycloak_integration_test.go` — ausência de perfil na fixture real.
- [alterar] `api/internal/infra/webserver/middleware/auth_integration_test.go` — novas recusas e sigilo; `auth.go` somente se mapeamento exigir.
- [criar] `api/internal/infra/webserver/provisioning_integration_test.go` — integração com PostgreSQL isolado.
- [alterar] `api/internal/infra/keycloak/testdata/keycloak/README.md` — replay e limite da fixture antiga.
- [alterar] `infra/keycloak/README.md` — perfil requerido e fim do cadastro prévio obrigatório.
- [alterar] `docs/specs/openapi.yaml` — descrições de autenticação e 403.
- [alterar] `web/spa/src/shared/api/schema.d.ts` — somente geração.
- [alterar] `web/spa/README.md` — novo contrato e pré-requisitos do teste real.

## Critérios de aceite

- **T3-AC1:** Token válido e perfil completo permitem criação/sync e usam ID/e-mail persistidos na mesma chamada; igualdade não altera auditoria → CA1/CA2/CA7.
- **T3-AC2:** Nome/e-mail ausentes/brancos e verificação ausente/falsa resultam em 403 antes de user; token malformado/inválido continua 401 → CA3/CA8.
- **T3-AC3:** Desativado/conflito geram 403 e estado intacto; IO/cancelamento de persistência geram falha sanitizada e não liberam handler → CA4/CA5/CA8.
- **T3-AC4:** Logs/respostas não expõem token/sub/perfil/e-mail/SQL/erro bruto; contexto/timeout e identidade resistente a headers conflitantes preservados → CA7/CA8.
- **T3-AC5:** Documentação e tipos gerados descrevem cadastro automático/perfil obrigatório, fixture antiga incompleta e limite de snapshots, sem mudar envelopes/contratos financeiros → CA1/CA3/CA8.

## Verificação

Na raiz: `make api/vet`, `make api/lint`, `make api/build`, `make api/test`, `make api/test/integration`. Go compatível, C/race; fixtures locais sem serviços externos. Esperado: novos fluxos e regressão criptográfica/CORS/roles passando.

Em `api/`, novo gate de T2: `go test ./... -race -count=1 -tags=integration,postgres`, com `USER_TEST_DATABASE_URL` para PostgreSQL 17+ isolado e schemas próprios. Testes solicitados sem DSN falham, não skip. Esperado: estado persistido e status HTTP reais dos AC1–AC4, com verifier controlado; não alegar teste de assinatura nessa composição.

Na raiz: `npm --prefix web/spa run api:types` para gerar, depois `npm --prefix web/spa run api:types:check`, `npm --prefix web/spa run lint`, `npm --prefix web/spa run build`, `npm --prefix web/spa test`. Scripts confirmados; Node/dependências e config pública OIDC para build conforme README. Esperado: tipos sem drift e consumidor compilável.

Integração operacional: `npm --prefix web/spa run test:e2e:auth` com pré-requisitos do README/config existente (Keycloak 8080, API 8081, Vite 5173 livre, credenciais fictícias por ambiente). Esperado: login e Bearer reais funcionam, sem pressupor seed de user_account. Não alterar realm/banco compartilhado nem expor credenciais em artefatos. Banco da API deve estar preparado em ambiente isolado; startup real executa migrations.

## Contexto adicional

D1–D5, P1/P2/P6: cada chamada usa snapshot do access token; dados novos só chegam após emissão/refresh, sem proteção monotônica entre tokens. Esta task altera intencionalmente o 403 de usuário ausente da ISSUE-1 e cumpre o follow-up da ISSUE-2; preserva desativação e autorização financeira. API não deve distinguir publicamente as causas do 403. Seguir AGENTS/docs/go, sem nova dependência/interface genérica.

## Evidências

<!-- Preenchido pelo dev-mage: tabela entrada → persistência → principal/status, sigilo, tipos/gates e commit. Separar fixtures/verifier controlado de login real. -->
