# ISSUE-3 — T1: Extrair claims de perfil validado

Spec: [ISSUE-3](./ISSUE-3-cadastro-sincronizacao-usuarios.md)

- **Status:** ⬜ pendente
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
