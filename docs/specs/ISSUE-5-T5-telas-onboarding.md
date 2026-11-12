# ISSUE-5 — T5: Componentes do onboarding

Spec: [ISSUE-5](./ISSUE-5-onboarding-guild-apos-login.md)

- Status: concluida
- Depende de: T3,T4

## O que fazer
Criar seleção, cartões de convites e formulário display_name/description. Usar session.request e política de retries existente. Aceite/criação ativam GuildResponse; recusa reconsulta. Implementar estados loading/erro/pending/403, acessibilidade/i18n e mocks com membership mutável/reset.

## Testes planejados
Testar dados exibidos, escolha, aceite/recusa/criação, falhas/conflitos, ausência de duplicação e logout durante ações.

## Arquivos
web/spa/src/features/guilds/ (componentes/schema/testes [criar], API/hooks existentes), mocks e traduções.

## Critérios de aceite
T5-AC1: comportamento descrito em O que fazer coberto por testes → CA2–CA6,CA9–CA10.
T5-AC2: testes planejados observados, gates aplicáveis passando e revisão independente do alto risco.

## Verificação
Tipos, lint, build, Vitest e E2E existente.
Comandos exatos estão em Verificação final da spec. API: make api/vet api/lint api/test api/test/integration api/build na raiz. PostgreSQL direcionado: go test ./internal/guild -race -count=1 -tags=integration,postgres em api/, GUILD_TEST_DATABASE_URL exclusiva. Frontend: npm --prefix web/spa run api:types:check; npm --prefix web/spa run lint; npm --prefix web/spa run build; npm --prefix web/spa test; npm --prefix web/spa run test:e2e; npm --prefix web/spa run test:e2e:prod. Auth real: npm --prefix web/spa run test:e2e:auth com pré-requisitos da spec.

## Contexto
Seguir contratos e limites da spec aprovada. Não alterar critérios para acomodar implementação. Todas as evidências futuras permanecem pendentes até execução.

## Evidências
| Critério | Evidência | Esperado | Resultado |
|---|---|---|---|
| T5-AC1 | testes descritos acima (planejados) | comportamento aprovado | aprovado |
| T5-AC2 | gates e revisão | sem falhas no escopo | aprovado |

Commit: 8c4f5d0.

### Fechamento T5
RED cinco sensores UI falharam pela ausência de convites/formulário. GREEN: aceite com contexto/pref, recusa do último seguida de criação, múltiplos convites/recusa parcial, draft preservado sem retry, validação71, loading403 sem criação, logout e respostas tardias, desambiguação visual/acessível. Sensores de resposta tardia201/200 conservam403 mais recente de convites e não ativam cache/pref.
Lint/build passaram; Vitest19arquivos164testes passaram; demo15/prod8 passaram. Auth real6 passou antes da integração global. Revisão independente sem achados após ajustes de403e labels duplicados. Nenhuma API/SQL muda nesta task. Sensor adicional T4 verifica cache não repovoado após logout.
