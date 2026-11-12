# ISSUE-5 — T6: Entrada obrigatória após autenticação

Spec: [ISSUE-5](./ISSUE-5-onboarding-guild-apos-login.md)

- Status: pendente
- Depende de: T5

## O que fazer
Integrar onboarding antes de Home/menu/rotas; preservar cabeçalho identidade/idioma/Sair. Liberar rota original com contexto. Adaptar testes de Home e auth real, documentando mudança deliberada. Acrescentar E2E de toda jornada.

## Testes planejados
Testar preferência após novo login, seleção, aceite, recusa do último, criação com ativação; nenhuma consulta financeira antecipada; sessão/links preservados.

## Arquivos
web/spa/src/app/app.tsx, cabeçalho/testes, E2E de contexto/auth, nova suíte onboarding [criar] e README SPA.

## Critérios de aceite
T6-AC1: comportamento descrito em O que fazer coberto por testes → CA1–CA6,CA10.
T6-AC2: testes planejados observados, gates aplicáveis passando e revisão independente do alto risco.

## Verificação
Gates finais API e frontend em ordem da spec. PostgreSQL prova autorização/atomicidade; mocks provam UX.
Comandos exatos estão em Verificação final da spec. API: make api/vet api/lint api/test api/test/integration api/build na raiz. PostgreSQL direcionado: go test ./internal/guild -race -count=1 -tags=integration,postgres em api/, GUILD_TEST_DATABASE_URL exclusiva. Frontend: npm --prefix web/spa run api:types:check; npm --prefix web/spa run lint; npm --prefix web/spa run build; npm --prefix web/spa test; npm --prefix web/spa run test:e2e; npm --prefix web/spa run test:e2e:prod. Auth real: npm --prefix web/spa run test:e2e:auth com pré-requisitos da spec.

## Contexto
Seguir contratos e limites da spec aprovada. Não alterar critérios para acomodar implementação. Todas as evidências futuras permanecem pendentes até execução.

## Evidências
| Critério | Evidência | Esperado | Resultado |
|---|---|---|---|
| T6-AC1 | testes descritos acima (planejados) | comportamento aprovado | pendente |
| T6-AC2 | gates e revisão | sem falhas no escopo | pendente |

Commit: pendente.
