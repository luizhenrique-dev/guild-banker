# OPENAPI_GAPS.md — Pendências da especificação `openapi.yaml`

Este documento lista as informações que não puderam ser confirmadas com
certeza a partir do código, testes ou documentação existente durante a
geração de `docs/specs/openapi.yaml`, junto com a evidência disponível e a
recomendação de ação.

---

## 1. Autenticação JWT (Keycloak) não está implementada

- **Endpoint/componente afetado**: todos os endpoints sob `/api/v1/*`
  (security scheme `bearerAuth`).
- **Informação faltante**: regras reais de validação do token (claims,
  audience, issuer, expiração, mapeamento de roles/escopos).
- **Motivo da incerteza**: `api/internal/infra/webserver/middleware/auth.go`
  apenas extrai o Bearer token do header `Authorization` e imediatamente
  retorna `501 Not Implemented` com o comentário
  `// TODO: implement Keycloak JWT validation.`. Nenhuma validação de
  assinatura, claims ou expiração é executada.
- **Arquivo(s) a revisar**: `api/internal/infra/webserver/middleware/auth.go`,
  `api/config/config.go` (campos `Keycloak.*`).
- **Recomendação**: implementar a validação JWT via JWKS do Keycloak antes de
  considerar o endpoint utilizável em produção; atualizar a especificação
  com os claims/roles relevantes assim que a implementação existir.

## 2. Identidade efetiva do requisitante via headers, não via JWT

- **Endpoint/componente afetado**: todos os handlers de `guild`,
  `fixed_expense`, `transaction` e `importer`.
- **Informação faltante**: se os headers `X-User-ID` / `X-User-Email` são um
  mecanismo definitivo (ex.: para uso interno/gateway) ou um placeholder
  temporário até a validação JWT estar pronta.
- **Motivo da incerteza**: como a validação JWT retorna 501 antes de chegar
  aos handlers, é impossível, hoje, executar qualquer chamada autenticada
  com sucesso fim-a-fim; os headers são a única fonte de identidade lida
  pelo código dos handlers.
- **Arquivo(s) a revisar**: `api/internal/guild/handler.go`,
  `api/internal/fixed_expense/handler.go`,
  `api/internal/transaction/handler.go`, `api/internal/importer/handler.go`.
- **Recomendação**: definir se os headers serão substituídos por claims do
  JWT (ex.: `sub`, `email`) quando a validação for implementada, e atualizar
  a especificação removendo os parâmetros de header caso a identidade passe
  a vir exclusivamente do token.

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
