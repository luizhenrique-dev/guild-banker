# ISSUE-5 — T3: Consulta e resposta de convites

Spec: [ISSUE-5](./ISSUE-5-onboarding-guild-apos-login.md)

- Status: pendente
- Depende de: T2

## O que fazer
Implementar GET guild-invites e POST accept/decline definidos na spec. Autorizar pelo principal e carregar dados atuais do convidador. Transacionar resposta/membership com locks guild→convite; primeira decisão vence. Atualizar contratos/tipos/mocks.

## Testes planejados
Testar convite antes de login, recurso alheio, duplo aceite, aceite×recusa, emissão×aceite, rollback de membership e reenvio após recusa.

## Arquivos
api/internal/guild/ (novos arquivos de convites e testes), api/internal/infra/webserver/routes.go, integração de provisionamento quando necessária, OpenAPI/tipos/mocks.

## Critérios de aceite
T3-AC1: comportamento descrito em O que fazer coberto por testes → CA3–CA5,CA9–CA10.
T3-AC2: testes planejados observados, gates aplicáveis passando e revisão independente do alto risco.

## Verificação
API completa e integração PostgreSQL completa; tipos sem drift e mocks compatíveis.
Comandos exatos estão em Verificação final da spec. API: make api/vet api/lint api/test api/test/integration api/build na raiz. PostgreSQL direcionado: go test ./internal/guild -race -count=1 -tags=integration,postgres em api/, GUILD_TEST_DATABASE_URL exclusiva. Frontend: npm --prefix web/spa run api:types:check; npm --prefix web/spa run lint; npm --prefix web/spa run build; npm --prefix web/spa test; npm --prefix web/spa run test:e2e; npm --prefix web/spa run test:e2e:prod. Auth real: npm --prefix web/spa run test:e2e:auth com pré-requisitos da spec.

## Contexto
Seguir contratos e limites da spec aprovada. Não alterar critérios para acomodar implementação. Todas as evidências futuras permanecem pendentes até execução.

## Evidências
| Critério | Evidência | Esperado | Resultado |
|---|---|---|---|
| T3-AC1 | testes descritos acima (planejados) | comportamento aprovado | pendente |
| T3-AC2 | gates e revisão | sem falhas no escopo | pendente |

Commit: pendente.

