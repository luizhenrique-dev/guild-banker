# ISSUE-1 — T1: Claims e contexto tipados

Spec: [ISSUE-1-autenticacao-jwt-rs256.md](./ISSUE-1-autenticacao-jwt-rs256.md)

- **Status:** ✅ concluida
- **Depende de:** nenhuma

## O que fazer

1. Criar feature `internal/auth` conforme docs/go: Claims com UserID string igual ao sub externo, PreferredUsername e RealmRoles; Principal reúne Claims e audit.Actor local. Não adicionar tags JSON nem dependência Gin/OIDC/SQL nesses valores.
2. Criar acesso tipado por request.Context: WithPrincipal, PrincipalFromContext e ClaimsFromContext. Chave privada, sem string/global mutável; contexto derivado preserva valores, cancelamento e deadline. Consulta ausente retorna bool falso, sem panic.
3. Implementar Claims.HasRole com igualdade exata em RealmRoles; member/admin sem hierarquia. Ausência de roles retorna false; nenhuma autorização de negócio nesta task.
4. Definir erros/categorias mínimos consumidos nas próximas tasks: token inválido, acesso negado e motivo de rejeição controlado. Permitir errors.Is/As e wrapping; não depender de texto de erro nem armazenar token/claims no erro público.
5. Testes planejados: TestPrincipalFromContext (ausente/presente, IDs externos e locais distintos), TestWithPrincipalPreservesContext, TestClaimsHasRole e TestRejectionError. Testar papéis ausentes e inexistentes; usar testing/testify existentes.

## Arquivos

- [criar] `api/internal/auth/claims.go`, `claims_test.go` — valores e HasRole.
- [criar] `api/internal/auth/context.go`, `context_test.go` — acesso tipado.
- [criar] `api/internal/auth/errors.go`, `errors_test.go` — categorias e inspeção de erros.

## Critérios de aceite

- T1-AC1 Claims externos e Actor local mantêm seus tipos/identificadores distintos → CA2, CA10.
- T1-AC2 Principal/claims são recuperados somente pelo contrato tipado; ausência não causa panic e cancelamento/deadline original permanece observável → CA2.
- T1-AC3 HasRole retorna resultados exatos para member, admin, desconhecida e lista ausente, sem admin implicar member → CA6.
- T1-AC4 Rejeições são inspecionáveis sem comparar strings e não carregam dados sensíveis para log → CA13.

## Verificação

- Na raiz, `make api/vet`, `make api/lint`, `make api/test`, `make api/build` — pré-requisitos: Go compatível com api/go.mod, race e dependências; esperado: novos testes e gates passam.
- Manual: não aplicável; não ativar autenticação, inicializar serviço ou acessar banco nesta task.

## Contexto adicional

D2/D3 aprovadas: roles somente em teste; identidade confiável substituirá headers em T4. `audit.Actor` existente em api/internal/audit/audit.go usa ID int64 e Email; Claims.UserID é sub string, não o ID local. Não criar interfaces antes de haver consumidor concreto; portas do serviço serão definidas em T3. Middleware atual permanece 401/501.

## Evidências

| Critério | Evidência observada | Resultado |
|---|---|---|
| T1-AC1 | TestPrincipalFromContext: sub externo e Actor local distintos | aprovado |
| T1-AC2 | TestPrincipalFromContext e TestWithPrincipalPreservesContext | aprovado |
| T1-AC3 | TestClaimsHasRole: igualdade exata e roles ausentes | aprovado |
| T1-AC4 | TestRejectionError e TestReasonString: errors.Is/As, vocabulário fixo e fallback desconhecido | aprovado |

RED observado: `go test ./internal/auth` falhou pelos tipos/funções ainda ausentes. GREEN observado: `go test -v ./internal/auth -race -count=1` executou os cinco testes e subcasos, todos PASS. Baseline `audit/user` PASS. Gates da task: `go vet ./internal/auth`, `go tool golangci-lint run ./internal/auth` (0 issues), `go build -o bin/guildbanker ./cmd/server` passaram. Gates restritos ao componente conforme dev-mage; suíte completa no fechamento da spec.

Go 1.26.2, GCC 15 extraído em `/tmp/issue1-tools/gcc`, CGO_ENABLED=1; caches temporários. Revisão independente por infra_ci_facts: sem bloqueador; sugestão de teste do vocabulário atendida e revalidada. Implementação validada; commit local preparado com esta task e os documentos do planejamento aprovado.
