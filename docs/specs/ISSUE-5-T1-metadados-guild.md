# ISSUE-5 — T1: Metadados e nome de guild

Spec: [ISSUE-5](./ISSUE-5-onboarding-guild-apos-login.md)

- Status: concluida
- Depende de: nenhuma

## O que fazer
Adicionar descrição em todas as camadas e migration 000002. Gerar name exclusivamente na aplicação, normalizando e resolvendo colisão conforme spec. Corrigir update display_name/description, sem alterar name ou exigir unicidade visual. Atualizar OpenAPI/tipos/mocks/fixtures/OPENAPI_GAPS.

## Testes planejados
Testar acentos compostos/decompostos, whitespace, pontuação, base sem ASCII, limites 70/71, colisão e concorrência; descrição omitida/vazia; criação atômica e update imutável.

## Arquivos
api/internal/guild/ (domínio/DTO/service/storage/handler/testes), api/migrations/000002_add_guild_description.{up,down}.sql [criar], docs/specs/openapi.yaml, OPENAPI_GAPS.md e tipos/mocks/consumidores web.

## Critérios de aceite
T1-AC1: comportamento descrito em O que fazer coberto por testes → CA6–CA8.
T1-AC2: testes planejados observados, gates aplicáveis passando e revisão independente do alto risco.

## Verificação
API completa, PostgreSQL de guild e gate frontend.
Comandos exatos estão em Verificação final da spec. API: make api/vet api/lint api/test api/test/integration api/build na raiz. PostgreSQL direcionado: go test ./internal/guild -race -count=1 -tags=integration,postgres em api/, GUILD_TEST_DATABASE_URL exclusiva. Frontend: npm --prefix web/spa run api:types:check; npm --prefix web/spa run lint; npm --prefix web/spa run build; npm --prefix web/spa test; npm --prefix web/spa run test:e2e; npm --prefix web/spa run test:e2e:prod. Auth real: npm --prefix web/spa run test:e2e:auth com pré-requisitos da spec.

## Contexto
Seguir contratos e limites da spec aprovada. Não alterar critérios para acomodar implementação. Todas as evidências futuras permanecem pendentes até execução.

## Evidências
| Critério | Evidência | Esperado | Resultado |
|---|---|---|---|
| T1-AC1 | testes descritos acima (planejados) | comportamento aprovado | aprovado |
| T1-AC2 | gates e revisão | sem falhas no escopo | aprovado |

Commit: pendente.


### Fechamento T1 — 2026-11-11
- RED observado: TestCreateGeneratesGuildName, TestGuildDisplayNameLimit e TestCreateRetriesOnlyNameConflicts falharam pelo comportamento ausente antes do código.
- GREEN: go test ./internal/guild passou; classificação de constraints e limite de seis persistências passaram com race. TestNameGenerationExhaustionReturnsUnavailable observou503.
- PostgreSQL exclusivo guildbanker-issue5-postgres/porta55435: TestStorageCreatePostgres, TestConcurrentGuildCreationPostgres (seis nomes visuais iguais, seis nomes únicos/memberships) e TestGuildMetadataHTTPPostgres passaram com race. HTTP prova descrição, campo name ignorado, update imutável, omissão/limpeza e limite71.
- make api/vet api/lint api/test api/test/integration api/build passou; lint0issues.
- Frontend lint/build/tipos passaram. Vitest completo17arquivos136testes passou antes da adição do sensor de mocks; suíte contracts posteriormente7testes passou.
- E2E demo15, produção8 e auth real6 passaram. Serviços usados exclusivamente para teste; sem alteração em dados compartilhados.
- Primeiro smoke foi executado com build sem VITE_KEYCLOAK_* e falhou na configuração, corrigido recompilando com as variáveis públicas documentadas. Chromium ausente foi instalado; falha inicial era de infraestrutura.
- Revisão independente review_t1_issue5: sem defeito impeditivo de produção. Ajustes de mock/contrato e sensores solicitados foram aplicados e verificados.
- Avisos existentes: npm config python, chunk de build >500kB. Sem falha nas verificações finais executadas.
