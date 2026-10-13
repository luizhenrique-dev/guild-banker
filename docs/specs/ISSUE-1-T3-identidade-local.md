# ISSUE-1 — T3: Vínculo com usuário local

Spec: [ISSUE-1-autenticacao-jwt-rs256.md](./ISSUE-1-autenticacao-jwt-rs256.md)

- **Status:** ⬜ pendente
- **Depende de:** T1, T2

## O que fazer

1. Criar user.ErrNotFound, traduzindo sql.ErrNoRows no adapter user/storage.go sem expor SQL ao consumidor auth. Preservar wrapping e inspeção já suportada por consumidores existentes, sem alterar assinatura de Storage/Service ou mensagens de outros erros. Não reutilizar guild.ErrUserNotFound.
2. Criar auth.Service.Authenticate com portas estreitas no consumidor: verifier de token e lookup GetByExternalID, satisfeito por *user.Service existente. Verificar token antes de lookup; sub externo é passado intacto, nunca convertido em número nem substituído por email/username.
3. Resolver usuário local e formar Principal com Actor{UserID: user.ID, Email: user.Email}. Inexistente ou Enabled=false recebe acesso negado; falha técnica retorna erro interno distinguível. Não cadastrar, sincronizar, habilitar ou atualizar usuário. Registro válido previamente cadastrado é a premissa operacional.
4. Propagar ctx para ambas as portas e preservar erro via %w/errors.Is/As. Serviço não faz logging, não depende de Gin/Keycloak/SQL e não altera regras de membership/ownership.
5. Testes planejados: TestServiceAuthenticate (ativo, inexistente, inativo, erro técnico, token inválido sem lookup, ID externo não numérico, dados locais diferindo do JWT), TestServiceAuthenticateContext e extensão de TestService_GetByExternalID para ErrNotFound. Testar tradução do storage por driver de teste database/sql, sem nova dependência externa ou banco compartilhado.

## Arquivos

- [criar] `api/internal/auth/service.go`, `service_test.go` — orquestração/portas.
- [criar] `api/internal/user/errors.go` — sentinel da feature.
- [alterar] `api/internal/user/storage.go` — ausência de linha traduzida.
- [alterar] `api/internal/user/service_test.go` — propagação de ErrNotFound.
- [criar] `api/internal/user/storage_test.go` — comportamento de ausência/erro técnico com driver de teste stdlib.

## Critérios de aceite

- T3-AC1 Usuário é localizado exclusivamente pelo sub/ExternalID e principal usa ID/Email locais → CA2, CA10.
- T3-AC2 Inexistente/inativo produz acesso negado; erro técnico não é mascarado como ausência/token inválido; nenhuma porta de criação/update é chamada → CA10, CA13.
- T3-AC3 Token inválido não consulta usuário e ctx chega às duas fronteiras → CA1, CA3–CA5, CA8, CA10.
- T3-AC4 Ausência é distinguível por user.ErrNotFound sem importar database/sql no serviço auth; falha técnica mantém wrapping → CA10.

## Verificação

- Na raiz: `make api/vet`, `make api/lint`, `make api/test`, `make api/test/integration`, `make api/build` — alvo integration introduzido por T2; Go compatível/race/dependências necessários. Esperado: novos testes de portas/lookup e gates passam sem PostgreSQL ativo.
- Manual: não aplicável; não executar migration nem preparar usuário no banco compartilhado.

## Contexto adicional

D3/D6 aprovadas: cadastro prévio obrigatório, 403 para inexistente/inativo, sem onboarding. Service.GetByExternalID existe em api/internal/user/service.go; User.Enabled é promovido de audit.DisableEntry; Email é fonte de auditoria existente. O serviço retorna erros, middleware mapeará HTTP em T4. Os handlers ainda não mudam e API permanece 401/501 ao fim desta task. Não criar ciclo user → auth.

## Evidências

<!-- Preenchido pelo dev-mage: testes observados e evidência de cada T3-AC. -->
