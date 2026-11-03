# ISSUE-3 — T1: Extrair claims de perfil validado

Spec: [ISSUE-3](./ISSUE-3-cadastro-sincronizacao-usuarios.md)

- **Status:** concluida
- **Depende de:** nenhuma

## O que fazer

1. Estender `auth.Claims` com `Name`, `Email`, `EmailVerified`, sem tags JSON e sem incorporar DTO de infraestrutura. Preservar `UserID=sub`, username e roles.
2. Estender `keycloak.tokenClaims` e `Verifier.Verify` para extrair `name`, `email`, `email_verified` do payload já validado. Assinatura, issuer, audience, expiração, Bearer e subject continuam obrigatórios. Tipo incompatível ou `null` presente é token malformado; ausência extrai zero value. Política de perfil obrigatório será conectada em T3, sem alterar o contrato do verifier para recusar ausência.
3. Testes antes da implementação: perfil completo, claims ausentes, `email_verified=false`, tipos inválidos/null e preservação dos campos anteriores. Nomes **planejados**: `TestVerifierProfileClaims`, extensão de `TestVerifierVerify` e teste de transporte em `auth/claims_test.go` se necessário.
4. Manter a fixture antiga original e seus testes atuais passando: ela não possui perfil, mas a ausência ainda não altera o resultado de `Authenticate` nesta task. Não inserir claims em token assinado nem relaxar verificação para testar.

## Arquivos

- [alterar] `api/internal/auth/claims.go` — campos internos adicionais.
- [alterar] `api/internal/auth/claims_test.go` — contrato de claims quando pertinente.
- [alterar] `api/internal/infra/keycloak/verifier.go` — adapter JSON e mapeamento.
- [alterar] `api/internal/infra/keycloak/verifier_test.go` — tabela de extração e rejeição de tipos.
- [alterar] `api/internal/infra/keycloak/verifier_keycloak_integration_test.go` — asserções de perfil ausente da fixture, se necessário.

## Critérios de aceite

- **T1-AC1:** Perfil completo/verificado é extraído somente após validar token; UserID/username/roles e proteções criptográficas permanecem → CA3/CA8.
- **T1-AC2:** Ausência é representável para a aplicação negar em T3; falso permanece falso; tipos JSON inválidos/null recebem `ErrInvalidToken`, sem payload no erro → CA3/CA8.
- **T1-AC3:** API compila e suítes existentes passam sem antecipar a mudança de cadastro/autorização da T3 → pré-condição de CA1–CA8.

## Verificação

Na raiz: `make api/vet`, `make api/lint`, `make api/build`, `make api/test`, `make api/test/integration`. Confirmados no Makefile; pré-requisitos: Go compatível com `api/go.mod`, dependências e compilador C para race; integração atual usa fixtures locais, sem banco/Keycloak ativos. Esperado: testes de extração e regressão passam, inclusive replay original sem perfil.

Manual: não aplicável; extração é verificável por testes com JWT assinado e relógio controlado.

## Contexto adicional

D1/D3/D5: fonte é access token validado; obrigatoriedade é da aplicação, não do parser. A spec propõe 401 para tipo inválido e 403 para ausência/perfil insuficiente; não implementar a segunda parte nesta task. Não usar ID token, UserInfo, headers de identidade ou fallback de nomes. Seguir AGENTS e docs/go; não criar interfaces novas para simples extração.

## Evidências

<!-- Preenchido pelo dev-mage: RED/GREEN, comandos/resultados e commit local. Nenhuma execução realizada no planejamento. -->

| Critério | Evidência planejada | Resultado |
|---|---|---|
| T1-AC1 | `go test -v ./internal/infra/keycloak -run 'TestVerifier(ProfileClaims\|Verify)'` em api/: ambas as funções passaram | aprovado |
| T1-AC2 | `TestVerifierProfileClaims`: 10 casos executados e aprovados, incluindo ausência, falso, tipos incompatíveis e null | aprovado |
| T1-AC3 | vet, build, unitários race e integração race passaram; lint falha por incompatibilidade do importer com Go 1.27.3 | aprovado |

Baseline: `go test -v ./internal/auth ./internal/infra/keycloak` em api/: ambas as suítes passaram; árvore inicial limpa.

### Checkpoint — 2026-11-02

- RED: com campos internos adicionados e parser ainda original, `TestVerifierProfileClaims` falhou por perfil vazio e aceitação de tipos incompatíveis/null; ausência e false já passavam. A primeira tentativa na raiz falhou por diretório incorreto; execução efetiva foi em `api/`.
- GREEN: `go test -v ./internal/infra/keycloak -run 'TestVerifier(ProfileClaims|Verify)'` em `api/`: extração e regressão criptográfica passaram. Fixture Keycloak original preservada; autenticação não alterada.
- Gates na raiz: `make api/vet` passou; `make api/build` gerou `api/bin/guildbanker`; `make api/test` passou em todos os pacotes, com race; `make api/test/integration` passou em todos os pacotes com race após escalonamento para permitir listeners locais (primeira execução no sandbox falhou por `socket: operation not permitted`).
- `make api/lint`: primeiro impedido pelo cache Go somente leitura; repetido fora do sandbox, instalou o linter fixado, mas falhou no typecheck. `go tool golangci-lint version` confirmou 2.11.4 compilado com Go 1.27.3. `go tool golangci-lint run ./database` reproduziu a mesma falha num pacote não alterado: `export data version 5 is greater than maximum supported version 2` ao importar biblioteca padrão. Nenhuma mudança de dependência realizada.
- Revisão independente: agente `query_patterns` examinou diff/contrato/segurança e executou os dois testes com `-count=1`; nenhum achado. Segunda revisão própria também não encontrou desvio de escopo. `git diff --check` passou.
- Implementação validada: parcialmente (AC1/AC2); AC3 pendente pelo lint. Commit pendente até resolução do gate; staging vazio.

### Bloqueio — T1

**Problema:** tooling de lint fixado não consegue ler dados de tipos emitidos pelo Go 1.27.3 exigido no projeto.
**Evidência:** erro reproduzido no pacote inalterado `api/database`; versão do linter 2.11.4 e x/tools v0.43.0 em `api/go.mod`.
**Impacto:** T1-AC3 obrigatório não encerrado; T2–T5 dependem sequencialmente de T1.
**Proposta:** investigar e ajustar a compatibilidade do tooling em escopo explicitamente autorizado, mantendo os gates; depois repetir lint e verificações afetadas.
**Confirmação necessária:** autorização solicitada para ampliar escopo de tooling, ainda pendente neste checkpoint.

### Resolução do bloqueio

Usuário autorizou ajuste de tooling. Atualizados golangci-lint v2.15.0 e dependências exigidas, `go mod tidy` sincronizou o grafo/sums. Fonte oficial: https://golangci-lint.run/docs/product/changelog/. A remoção de sums antigos é efeito do tidy; nenhuma dependência de funcionalidade foi introduzida.

Execução após atualização: `make api/lint api/vet api/build api/test api/test/integration` fora do sandbox → lint **0 issues**, vet passou, build gerado, todos os pacotes das duas suítes com race passaram. T1-AC3 aprovado; bloqueio resolvido. Implementação validada: sim. Commit local desta task inclui o ajuste autorizado do tooling; hash será observado após commit.
