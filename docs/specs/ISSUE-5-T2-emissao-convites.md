# ISSUE-5 — T2: Emissão de convites pendentes

Spec: [ISSUE-5](./ISSUE-5-onboarding-guild-apos-login.md)

- Status: concluida
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
| T2-AC1 | testes descritos acima (planejados) | comportamento aprovado | aprovado |
| T2-AC2 | gates e revisão | sem falhas no escopo | aprovado |

Commit: f65c41b.

### Fechamento T2
RED observado: emissão falhou por query legada users inexistente; APIError descartava code. GREEN: TestPendingInvitationsPostgres comprova convite sem conta/membership, normalização, ErrorIs distinto para duplicata/membro, emissor negado sem persistência e somente uma emissão concorrente vencedora; perdedor ErrInvitationAlreadyPending. TestInvitationErrorCodes observa409/400/500 sanitizado.
Gates API vet/lint(0issues)/unit/integration/build passaram. PostgreSQL guild completo com race passou. Frontend lint/tipos/build passaram; Vitest completo139testes e depois suite de feedback7testes passaram, incluindo duas mensagens localizadas e preservação do rascunho. Demo15/produção8/authreal6 passaram.
Revisão independente review_t1_issue5: sem defeito impeditivo; sensores adicionais pedidos foram implementados e passaram. Retomada após reinício exigiu restaurar serviços exclusivos; fixture de realm agora copiada ao container. Falhas de startup foram corrigidas aguardando prontidão, sem alterar código para acomodar sensor.
