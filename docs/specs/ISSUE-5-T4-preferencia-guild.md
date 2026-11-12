# ISSUE-5 — T4: Contexto e preferência de guild

Spec: [ISSUE-5](./ISSUE-5-onboarding-guild-apos-login.md)

- Status: concluida
- Depende de: T1

## O que fazer
Centralizar resolução em provider abaixo da sessão mantendo useActiveGuild. URL válida > preferência válida > guild única; várias sem escolha aguardam seleção. Persistir ID por identidade, preservar no logout, tratar storage inválido/bloqueado. Preservar URL/histórico/403/cancelamento.

## Testes planejados
Testar novo login, usuário diferente, guild removida, URL inválida, storage bloqueado/corrompido e respostas tardias; adaptar premissas antigas de seleção da primeira guild.

## Arquivos
web/spa/src/features/guilds/ (provider e testes [criar], hook existente), composição de providers e testes de sessão/contexto/E2E afetados.

## Critérios de aceite
T4-AC1: comportamento descrito em O que fazer coberto por testes → CA1–CA2,CA10.
T4-AC2: testes planejados observados, gates aplicáveis passando e revisão independente do alto risco.

## Verificação
Gate frontend completo.
Comandos exatos estão em Verificação final da spec. API: make api/vet api/lint api/test api/test/integration api/build na raiz. PostgreSQL direcionado: go test ./internal/guild -race -count=1 -tags=integration,postgres em api/, GUILD_TEST_DATABASE_URL exclusiva. Frontend: npm --prefix web/spa run api:types:check; npm --prefix web/spa run lint; npm --prefix web/spa run build; npm --prefix web/spa test; npm --prefix web/spa run test:e2e; npm --prefix web/spa run test:e2e:prod. Auth real: npm --prefix web/spa run test:e2e:auth com pré-requisitos da spec.

## Contexto
Seguir contratos e limites da spec aprovada. Não alterar critérios para acomodar implementação. Todas as evidências futuras permanecem pendentes até execução.

## Evidências
| Critério | Evidência | Esperado | Resultado |
|---|---|---|---|
| T4-AC1 | testes descritos acima (planejados) | comportamento aprovado | aprovado |
| T4-AC2 | gates e revisão | sem falhas no escopo | aprovado |

Commit: 397159b.

### Fechamento T4
RED quatro falhas observadas: seleção automática entre várias, restauração ignorada, preferência não persistida e ID corrompido. GREEN provider12testes: URL/prioridade/validação, IDs removidos/corrompidos, storage indisponível, usuário isolado, guild única, logout conserva preferência e resposta tardia bloqueada. Resposta antiga não apaga erro403mais recente, comprovado também diretamente no estado do cache.
Gates frontend lint/build passaram; suite completa152testes antes dos dois sensores finais; provider12passou depois deles. Demo15/prod8/authreal6 passaram. Ajustes de fixtures antigos usam explicitamente guild1/preferência, mantendo asserções de isolamento/autorização/histórico.
Revisão independente aprovou o provider após proteção adicional contra ativação tardia sobre403. Nenhuma API/SQL foi alterada nesta task; gatesAPI não aplicáveis.
