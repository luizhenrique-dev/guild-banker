# OPENAPI_GAPS.md — Pendências da especificação `openapi.yaml`

Este documento lista as informações que não puderam ser confirmadas com
certeza a partir do código, testes ou documentação existente durante a
geração de `docs/specs/openapi.yaml`, junto com a evidência disponível e a
recomendação de ação.

---

## 1. Autenticação JWT — resolvida por ISSUE-1

Access token Keycloak validado por go-oidc/v3 com RS256, JWKS, issuer,
expiração e audience configurada. ID/refresh tokens rejeitados. `/api` e
`/api/*` protegidos, inclusive NoRoute; health público. Ver spec ISSUE-1 e
os testes em `api/internal/infra/keycloak` e `webserver/middleware`.

## 2. Identidade efetiva — resolvida por ISSUE-1

`sub` resolve ExternalID do usuário local previamente cadastrado e ativo.
ID/e-mail locais formam o Actor no request.Context dos quatro handlers.
Headers de identidade não são aceitos como fonte de autenticação; parâmetros
removidos do OpenAPI. Ausente/inativo recebe403; falha técnica500 sanitizado.
Integração OIDC real da SPA e CI permanecem em issues separadas.

## 3. Inconsistência de nomenclatura em `PUT /api/v1/guilds/{id}`

- **Endpoint/componente afetado**: `updateGuildName` /
  `UpdateGuildRequest`.
- **Informação faltante**: se o campo deveria se chamar `display_name` (para
  consistência com `CreateGuildRequest`) ou se `name` é intencional.
- **Motivo da incerteza**: o handler (`guild/handler.go`) lê `req.Name` e
  chama `service.UpdateName(ctx, guildID, req.Name, actor)`, mas o método
  `Service.UpdateName` (`guild/service.go`) trata esse valor como
  `displayName` e o repassa para `Guild.Rename`, que altera o campo
  `DisplayName` (`guild/guild.go`). O nome de negócio (`Name`) do guild nunca
  é alterado por esse endpoint, apesar do nome do campo JSON sugerir o
  contrário.
- **Arquivo(s) a revisar**: `api/internal/guild/dto.go`,
  `api/internal/guild/handler.go`, `api/internal/guild/service.go`,
  `api/internal/guild/guild.go`.
- **Recomendação**: renomear o campo do DTO para `display_name` (alinhado ao
  `CreateGuildRequest`) ou renomear as funções internas para refletir que o
  endpoint realmente atualiza o `name`. Até a decisão, a especificação
  documenta o comportamento real observado no código.

## 4. Semântica de `UpdateTransactionRequest` com campos opcionais

- **Endpoint/componente afetado**: `PATCH /api/v1/guilds/{guildID}/transactions/{id}`.
- **Informação faltante**: comportamento exato quando um campo do corpo é
  omitido (JSON ausente) versus enviado como `null`, dado que
  `UpdateInput`/`ApplyUpdate` usam ponteiros mas a validação
  (`ValidateForUpdate`) exige valores não vazios/zerados.
- **Motivo da incerteza**: não foi encontrado teste de unidade/integração
  cobrindo explicitamente o envio parcial do corpo de atualização; o texto
  do binding do Gin (`ShouldBindJSON`) preenche a struct com ponteiros, mas o
  service (`transaction/service.go`) não foi inspecionado a fundo quanto ao
  tratamento de `nil` antes da chamada a `ApplyUpdate` (que recebe valores
  concretos, não ponteiros).
- **Arquivo(s) a revisar**: `api/internal/transaction/service.go`,
  `api/internal/transaction/command.go`, `api/internal/transaction/transaction.go`,
  `api/internal/transaction/service_test.go`.
- **Recomendação**: adicionar testes cobrindo atualização parcial e, se o
  comportamento for "campo omitido preserva valor atual", documentar
  claramente esse contrato na especificação (hoje descrito apenas como
  `TODO` na descrição do schema).

## 5. Versão da API (`info.version`)

- **Endpoint/componente afetado**: documento como um todo (`info.version`).
- **Informação faltante**: número de versão oficial da API/projeto.
- **Motivo da incerteza**: não há tag de release, `CHANGELOG`, campo de
  versão em `go.mod` (apenas a versão da linguagem Go) ou constante de
  versão no código-fonte.
- **Arquivo(s) a revisar**: `api/go.mod`, `PROJECT.md`, `README.md`, tags Git
  (não inspecionadas nesta análise).
- **Recomendação**: adotar versionamento semântico explícito (ex.: via tag
  Git ou constante no código) e atualizar `info.version` de acordo.
  Enquanto isso, o valor `0.1.0` foi usado como padrão, conforme instruído.

## 6. Formato de erro `500 Internal Server Error` em `transaction`

- **Endpoint/componente afetado**: todos os endpoints de `Transactions`
  (`handleError`, caso `default`).
- **Informação faltante**: se erros internos (500) sempre retornam a mesma
  estrutura `{"error": "..."}` ou se, em produção, algum middleware global de
  recuperação (`recover`) altera essa resposta.
- **Motivo da incerteza**: não foi localizado middleware de tratamento
  global de exceções/erros além do padrão do Gin (`gin.Default()` já inclui
  `Recovery()`), e o comportamento do `Recovery()` padrão do Gin em caso de
  panic não foi verificado neste projeto (normalmente retorna 500 sem corpo
  JSON estruturado).
- **Arquivo(s) a revisar**: `api/internal/infra/webserver/server.go`.
- **Recomendação**: confirmar/padronizar o formato de resposta para panics
  não tratados, e considerar um middleware de erro global consistente com o
  envelope `{"error": "..."}` já usado pelos handlers.

## 7. Upload de importação restrito ao banco C6

- **Endpoint/componente afetado**: `POST /api/v1/guilds/{guildID}/imports`.
- **Informação faltante**: se outros bancos serão suportados no futuro e se
  o parâmetro `SourceBank` será exposto como campo do formulário.
- **Motivo da incerteza**: o handler fixa `SourceBank: SourceBankC6` sem ler
  nenhum campo de formulário para essa informação
  (`api/internal/importer/handler.go`); o enum `SourceBank` só possui o valor
  `C6` implementado (`api/internal/importer/importer.go`).
- **Arquivo(s) a revisar**: `api/internal/importer/handler.go`,
  `api/internal/importer/importer.go`, `api/internal/importer/parser.go`.
- **Recomendação**: quando outro parser for adicionado, expor o campo
  `sourceBank` no `multipart/form-data` e atualizar o `requestBody` do
  endpoint na especificação.

---

## Ferramentas de validação

- Validação sintática: `lint` interno do ambiente (sem erros).
- Validação semântica adicional executada com `@redocly/cli lint` via
  `npx @redocly/cli@latest lint docs/specs/openapi.yaml`
  → resultado: **válido**, apenas 4 avisos (recomendações de estilo, não
  bloqueantes): ausência de `info.license`, uso de `localhost` em `servers`
  (intencional para ambiente local), ausência de resposta `4xx` em
  `/health` (não aplicável, endpoint não tem entrada de usuário) e um
  componente não utilizado (corrigido, removido).
- Não foi executado nenhum linter/validador previamente configurado no
  repositório, pois não foi encontrado nenhum (`Makefile`, `package.json`
  ou scripts de CI dedicados a OpenAPI) no momento da análise.

## Correções GB-20 (2026-10-04)
Conferidos routes.go, middleware/auth.go e DTOs/handlers de fixed_expense, transaction, guild e importer. Corrigidos servidor local 8081, /transactions/bulk-categorize, /imports/{importID}/confirm, enum de deactivate PAUSED/CANCELLED e resposta protegida 501. Os mocks demonstram handlers; não simulam autenticação real nem provam persistência/autorização. Sem endpoint GET de membros.
