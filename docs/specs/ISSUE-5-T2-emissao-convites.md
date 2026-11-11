# ISSUE-5 — T2: Emissão de convites pendentes

Spec: [ISSUE-5](./ISSUE-5-onboarding-guild-apos-login.md)

- Status: pendente
- Depende de: T1

## O que fazer
Criar guild_invitation/status/índices e migration 000003. Emissão 204 salva pendência para qualquer e-mail, sem conceder acesso. Normalizar e-mail e rejeitar duplicata pendente/membro com códigos distintos. Acrescentar code opcional ao erro HTTP/APIError e traduzir feedback.

## Testes planejados
Testar destinatário sem cadastro, normalização de e-mail, autorização, duas emissões simultâneas e nenhuma membership ao convidar.

## Arquivos
api/internal/guild/invitation.go [criar], serviço/storage/handler/erros/testes, api/migrations/000003_create_guild_invitation.{up,down}.sql [criar], contratos/cliente HTTP/mocks/traduções.

## Critérios de aceite
T2-AC1: comportamento descrito em O que fazer coberto por testes → CA9–CA10.
T2-AC2: testes planejados observados, gates aplicáveis passando e revisão independente do alto risco.

## Verificação
API completa, PostgreSQL de guild e gate frontend.
Comandos exatos estão em Verificação final da spec. API: make api/vet api/lint api/test api/test/integration api/build na raiz. PostgreSQL direcionado: go test ./internal/guild -race -count=1 -tags=integration,postgres em api/, GUILD_TEST_DATABASE_URL exclusiva. Frontend: npm --prefix web/spa run api:types:check; npm --prefix web/spa run lint; npm --prefix web/spa run build; npm --prefix web/spa test; npm --prefix web/spa run test:e2e; npm --prefix web/spa run test:e2e:prod. Auth real: npm --prefix web/spa run test:e2e:auth com pré-requisitos da spec.

## Contexto
Seguir contratos e limites da spec aprovada. Não alterar critérios para acomodar implementação. Todas as evidências futuras permanecem pendentes até execução.

## Evidências
| Critério | Evidência | Esperado | Resultado |
|---|---|---|---|
| T2-AC1 | testes descritos acima (planejados) | comportamento aprovado | pendente |
| T2-AC2 | gates e revisão | sem falhas no escopo | pendente |

Commit: pendente.

