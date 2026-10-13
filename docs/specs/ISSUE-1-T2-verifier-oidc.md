# ISSUE-1 — T2: Verifier OIDC e configuração

Spec: [ISSUE-1-autenticacao-jwt-rs256.md](./ISSUE-1-autenticacao-jwt-rs256.md)

- **Status:** ⬜ pendente
- **Depende de:** T1

## O que fazer

1. Adicionar release v3 de coreos/go-oidc compatível com o módulo, fixando versão/checksums após verificar manutenção e compatibilidade. Não adicionar jwt/v5, jwx, sqlmock ou biblioteca JWT extra. Referência confirmada: https://pkg.go.dev/github.com/coreos/go-oidc/v3/oidc.
2. Adicionar configuração KEYCLOAK_ISSUER/API_AUDIENCE obrigatórias e reutilizar KEYCLOAK_TIMEOUT_SECONDS positivo (default existente 5s). Atualizar example com issuer local completo e guild-banker-api; nenhum realm/client hardcoded no Go. Retirar obrigatoriedade dos envs antigos BaseURL/Realm/JWKSURI para auth, mantendo campos existentes se não precisarem mudar.
3. Tornar api/config.env opcional somente quando ausente; erros de leitura continuam retornados. Env de processo prevalece, como já ocorre com godotenv.Load. Validar nova configuração sem panic, sem refatorar DBConfig/DSN ou requireEnv dos demais campos. Testes de config controlam ambiente/arquivo e não dependem de config.env pessoal.
4. Criar adapter infra/keycloak que faz discovery com contexto e cliente HTTP limitado, constrói verifier reutilizável com ClientID=audience e SupportedSigningAlgs=[RS256], checks de assinatura/iss/aud/exp habilitados. Retornar auth.Claims após exigir payload typ=Bearer e sub não vazio. preferred_username/realm_access.roles ausentes são opcionais; tipos inválidos recebem rejeição. DTOs JSON ficam privados no adapter.
5. Categorizar rejeições por tipo/checagem própria. TokenExpiredError pode gerar token_expired; falhas não tipadas recebem invalid_token. Não comparar texto de erros nem publicar/logar token, claims ou erro bruto. Não chamar RemoteKeySet.VerifySignature isoladamente para aceitar tokens, não desabilitar checks.
6. Criar testes por tabela com chave RSA exclusiva de teste e instante fixo: assinatura válida/inválida, payload adulterado, issuer errado, exp vencido/ausente, aud ausente/outro/string/array, azp correto sem aud, none/HS256/RS512, ID/refresh token, sub ausente e roles ausentes. Testes de IO/discovery/cache recebem build tag integration.
7. Testes planejados: TestVerifierVerify; TestProviderDiscovery (sucesso, issuer divergente, erro HTTP, timeout); TestVerifierJWKSRotation (aquecer kid A, servir B, aceitar B no mesmo verifier, rejeitar C desconhecido); TestVerifierCancellation. Fixtures e relógio pertencem apenas aos testes; nenhuma flag de bypass/relógio por env. Manter ponto de construção de verifier com relógio/key set de teste privado no pacote infra/keycloak; os testes HTTP com relógio fixo de T4/T6 ficam nesse pacote e importam o middleware, sem nova opção pública.
8. Acrescentar alvo local `api/test/integration` no Makefile com `cd api && go test ./... -race -count=1 -tags=integration`. Trata-se de comando proposto a criar, não pipeline CI. Não ligar o adapter ao middleware placeholder ainda.

## Arquivos

- [alterar] `api/go.mod`, `api/go.sum` — dependência aprovada.
- [alterar] `api/config/config.go`, `api/config.env.example` — env/timeout/carregamento.
- [criar] `api/config/config_test.go` — config isolada.
- [criar] `api/internal/infra/keycloak/verifier.go`, `verifier_test.go`, `verifier_integration_test.go` — adapter/testes/fixtures sintéticas inline ou geradas no próprio teste.
- [alterar] `Makefile` — alvo local de integração.

## Critérios de aceite

- T2-AC1 Env novo é validado, processo pode fornecer config sem arquivo local, cliente HTTP limita IO e discovery não usa overrides inseguros → CA12.
- T2-AC2 Token passa somente com assinatura confiável RS256, issuer exato, exp válido e aud contendo audience configurada; azp não substitui aud → CA3, CA4, CA5, CA8.
- T2-AC3 Adapter devolve apenas claims validados; ID/refresh token e sub ausente são rejeitados; roles ausentes são aceitas → CA2, CA14.
- T2-AC4 Mesmo verifier reutiliza chave conhecida, atualiza JWKS para novo kid e rejeita kid sem chave; erro remoto não causa aceitação → CA7.
- T2-AC5 Categorias são seguras e inspecionáveis, com invalid_token como fallback para erros sem tipo → CA13.
- T2-AC6 Suíte unitária e integração local são reproduzíveis sem Keycloak/PostgreSQL ativos → CA9 local.

## Verificação

- Na raiz: `make api/vet`, `make api/lint`, `make api/test`, `make api/build` — comandos existentes; Go compatível e race necessários; esperado: gates passam.
- Após criar o alvo nesta task: `make api/test/integration` — alvo proposto, esperado: discovery, cancelamento, timeout, cache e rotação passam com httptest e sem serviços externos.
- Manual: não aplicável; não capturar token real ainda nem executar a API contra banco.

## Contexto adicional

D1/D4/D7 aprovadas. API real ainda usa placeholder 401/501 ao fim desta task. go-oidc denomina o verifier IDTokenVerifier; isso não dispensa semântica de access token. API de algoritmo: SupportedSigningAlgs, não AllowedSigningAlgs. Configuração CORS só é adicionada/aplicada em T5. RemoteKeySet administra cache/fetch; timeout de cliente limita fetch compartilhado mesmo após cancelamento individual. Consultar spec para limitações de cache e classificação de erros.

## Evidências

<!-- Preenchido pelo dev-mage: versão escolhida, testes e evidência de cada T2-AC. -->
