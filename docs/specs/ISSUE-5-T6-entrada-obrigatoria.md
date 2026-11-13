# ISSUE-5 — T6: Entrada obrigatória após autenticação

Spec: [ISSUE-5](./ISSUE-5-onboarding-guild-apos-login.md)

- Status: concluida
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
| T6-AC1 | testes descritos acima (planejados) | comportamento aprovado | aprovado |
| T6-AC2 | gates e revisão | sem falhas no escopo | aprovado |

Commit: este documento integra o commit local de fechamento da T6; identificar o hash pelo Git.

### Fechamento T6 — 2026-11-12
- RED: três testes observaram Home/menu antecipados e Home liberada no403. GREEN: onboarding-entry4testes incluindo contexto obrigatório, loading sem IO financeiro,403sem link em loop e Header distinguindo homônimos.
- Gate completo frontend passou em ordem: tipos sem drift, lint, build com OIDC público,20 arquivos / 168 testes,19 E2E demo,8 smoke de produção,6 de autenticação real. Avisos existentes: npm config python e chunk>500kB; nenhuma falha nas verificações finais executadas.
- Novos E2E: escolha e retomada após logout entrando em /home SEM ?guild= (preferência comprovada), múltiplos convites e aceite com rota/filtro conservados sem request financeiro antecipado, recusa do último seguida de criação com retry/draft/teclado/layout375 e 1280.
- Auth real executado inicialmente com banco NOVO exclusivo issue5_auth_20261112, sem conta/membership prévia e E2E_ALLOW_GUILD_CREATE=true. Primeiro teste criou pela UI e entrou em Home; consulta posterior observou1 usuário / 1 guild / 1 membership e name válido. Repetição final também passou.
- API final: make api/vet api/lint api/test api/test/integration api/build passou (lint0issues). Integração PostgreSQL de todos os pacotes com race e três URLs exclusivas passou. Sensor adicional TestAcceptancePreservesOtherInvitationsPostgres passou e observou o outro convite PENDING após aceitar um de dois.
- Revisão independente final: nenhum achado remanescente. Ajustes: Header compartilha guildLabel com cartões; preferência E2E provada sem URL; documentação distingue403 de memberships de 403 financeiro.
- Nenhuma migration ou operação destrutiva em ambiente compartilhado; serviços descartáveis são encerrados após a validação. Sem push/PR.
