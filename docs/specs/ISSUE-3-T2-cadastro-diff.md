# ISSUE-3 — T2: Resolver cadastro e diff com persistência segura

Spec: [ISSUE-3](./ISSUE-3-cadastro-sincronizacao-usuarios.md)

- **Status:** concluida
- **Depende de:** T1

## O que fazer

1. Adicionar `user.Service.EnsureAccount(ctx, externalID, name, email) (*User, error)` (**proposto**), validando strings obrigatórias/não brancas, sem conhecer JWT, HTTP ou Keycloak. Buscar por external_id; se ausente, criar habilitado com `New` e autor igual ao e-mail recebido. Sem fallback/lookup por e-mail para vincular conta.
2. Estender porta `Storage` e mocks com `CreateIfAbsent(ctx, u) (*User, error)` e `UpdateProfile(ctx, u) (*User, error)` (**propostos**). Primeiro método retorna cadastro inserido ou vencedor concorrente do mesmo external_id. Registro desativado retorna erro de domínio sem sync. Serviço compara nome/e-mail; se iguais, retorna sem `User.Sync`/gravação; se diferentes, usa `User.Sync` e a operação restrita de perfil.
3. No adapter, proteger criação pelo UNIQUE de external_id; distinguir colisão de e-mail usando erro estruturado do driver. Propor erros de domínio `ErrDisabled` e `ErrEmailConflict` em `user/errors.go`. Não comparar texto de erro, esconder todo conflito ou prosseguir em transação abortada.
4. Update de perfil escreve apenas nome/e-mail e `updated_at/by`, condicionado a usuário habilitado e diff atual no banco. Se não alterar linha, reler e distinguir desativado/perfil já aplicado; retornar estado persistido. Evitar UPDATE físico redundante entre chamadas concorrentes e preservar desativação concluída antes da escrita.
5. Corrigir `updated_by` no `userEntity`, INSERT/SELECT/UPDATE e conversões do adapter tocado; criação sem alteração deixa auditoria de update nula. Testar round-trip com autores de alteração/desativação distintos. Preservar APIs e comportamentos não relacionados de Create/Update/Enable/Disable; não usar o UPDATE completo atual no caminho novo.
6. Testes **planejados**: `TestServiceEnsureAccount` (ausente, igual, alterações isoladas/conjuntas, desativado, conflito, erro/cancelamento), `TestStorageProfileAudit`, `TestStorageEnsureAccountPostgres`, `TestStorageEnsureAccountConcurrent`, `TestStorageProfileDisableRace`. Asserções derivadas dos CAs, com testify; concorrência com barreiras/context e goroutines aguardadas, não sleeps.
7. Adicionar teste real PostgreSQL opt-in (`integration && postgres`), variável **proposta** `USER_TEST_DATABASE_URL`. Isolar um schema exclusivo com SQL da migration existente e search_path aplicado a todas as conexões. Credenciais ausentes/serviço indisponível no gate opt-in são falha clara, sem skip. Limpar só schema criado pelo teste; nunca dados compartilhados. Verificar conflitos reais, disputa de criação, diff físico e corrida de desativação, com conexões independentes. Não adicionar dependência de testcontainers.

## Arquivos

- [alterar] `api/internal/user/service.go` — novo caso de uso e diff.
- [alterar] `api/internal/user/errors.go` — erros distinguíveis.
- [alterar] `api/internal/user/storage.go` — porta, adapter parametrizado/atômico e auditoria.
- [alterar] `api/internal/user/service_test.go` — mocks e comportamento.
- [alterar] `api/internal/user/storage_test.go` — classificação e round-trip.
- [criar] `api/internal/user/storage_postgres_integration_test.go` — PostgreSQL isolado e concorrência.

## Critérios de aceite

- **T2-AC1:** Ausente cria habilitado com `external_id` intacto e ID local; criação concorrente do mesmo sub/perfil converge para um registro/ID → CA1/CA6/CA7.
- **T2-AC2:** Perfil igual não chama update; mudanças de nome/e-mail atualizam apenas perfil/auditoria, mantendo id/external_id/criação. Igualdade é textual, sem normalização implícita → CA2/CA7.
- **T2-AC3:** Colisão de e-mail na criação e no update preserva cadastros e retorna erro distinguível; falha técnica não vira conflito → CA4/CA8.
- **T2-AC4:** Desativado e desativação antes da escrita não recebem sync/reativação; sync concorrente igual não reescreve linha nem auditoria → CA5/CA6.
- **T2-AC5:** `updated_by` faz round-trip correto, inclusive com `disabled_by` diferente; contexto/cancelamento chegam a todos os IOs → CA7/CA8.

## Verificação

Na raiz: `make api/vet`, `make api/lint`, `make api/build`, `make api/test`, `make api/test/integration`. Gates confirmados, Go compatível e C para race. Esperado: capacidade nova coberta sem alterar `Authenticate` ainda.

**Gate novo proposto:** em `api/`, `go test ./... -race -count=1 -tags=integration,postgres`; fornecer `USER_TEST_DATABASE_URL` por ambiente para PostgreSQL 17+ isolado, com permissão para schema de teste. Variável/tag não existem antes desta task. Esperado: provas reais dos AC1–AC5; sem skip e sem gravação fora do schema exclusivo. Sem ambiente, registrar gate impedido; mocks não encerram os critérios de concorrência/constraints.

Manual: não aplicável; a inspeção do estado/auditoria e da contagem/versão da linha deve estar nos testes reais.

## Contexto adicional

D2/D3/D4: diff de nome/e-mail, vínculo apenas por sub e nenhum histórico. Validação de `email_verified` pertence a auth (T3), não a user. Auditoria usa e-mail novo validado. Não renomear a API existente nem alterar migrations. As garantias valem para o estado observado/escrita; não prometem revogar operação já em voo ou ordenar snapshots de tokens diferentes (P1/P6 na spec). SQL só no adapter, regras no serviço/domínio, sem user → auth e sem locks globais.

## Evidências

<!-- Preenchido pelo dev-mage: testes unitários e PostgreSQL reais separados; concorrência, auditoria, ambiente isolado, comandos/resultados e commit. -->

### Preparação — 2026-11-02

Baseline: `go test -v ./internal/user` em api/ passou. PostgreSQL 17.9 isolado em container próprio, loopback 55433; migration será aplicada exclusivamente em schemas exclusivos. Implementação aguarda conclusão dos gates T1.

| Critério | Evidência (execuções abaixo) | Resultado |
|---|---|---|
| T2-AC1 | Serviço ausente/vencedor + PostgreSQL concorrência de criação e ID único | aprovado — evidências observadas abaixo |
| T2-AC2 | Mock sem update em igualdade e diffs isolados/conjuntos; PostgreSQL xmin/auditoria | aprovado — evidências observadas abaixo |
| T2-AC3 | UNIQUE real em criação/update; classificação pq.Error; perfis intactos | aprovado — evidências observadas abaixo |
| T2-AC4 | Desativado; candidatos concorrentes do mesmo snapshot e corrida de desativação | aprovado — evidências observadas abaixo |
| T2-AC5 | Round-trip autores distintos; contexto/cancelamento nos IOs | aprovado — evidências observadas abaixo |

### Evidências observadas

- RED: `TestServiceEnsureAccount` com simples lookup falhou por usuário ausente, diff não aplicado e desativado não recusado; validação whitespace chamou indevidamente o mock. APIs ausentes inicialmente também impediram compilação, antes do RED comportamental.
- GREEN: `go test -v ./internal/user` passou, incluindo igualdade sem UpdateProfile, diffs isolados/conjunto/case, erros e propagação de contexto.
- T2-AC1: `TestStorageEnsureAccountConcurrent` passou com 8 chamadas simultâneas, ID único e uma linha; `TestStorageEnsureAccountPostgres` provou cadastro/auditoria inicial.
- T2-AC2: tabela de serviço e PostgreSQL provaram diff textual e preservação de ID/vínculo/criação; igualdade preservou xmin/auditoria.
- T2-AC3: PostgreSQL real provou UNIQUE em criação/update com dois cadastros intactos; `TestStorageProfileErrors` distingue constraint/código de falha técnica.
- T2-AC4: corrida controlada por bloqueio real de linha e pg_stat_activity passou; usuário desativado não recebe perfil. Candidatos concorrentes usam timestamps deliberadamente diferentes; retornos/auditoria iguais provam supressão do segundo UPDATE; terceira chamada também preserva xmin.
- T2-AC5: autores updated_by/disabled_by distintos fizeram round-trip; contextos observados no mock e cancelamento real antes do IO.
- `make api/lint api/vet api/build api/test api/test/integration` → lint 0 issues; vet/build e ambas as suítes race passaram. Dois avisos de teste corrigidos (chave de contexto e alias deprecated do driver), sem supressões.
- Gate em api/: `USER_TEST_DATABASE_URL=<container isolado> go test ./... -race -count=1 -tags=integration,postgres` → todos os pacotes passaram, inclusive os três testes reais. Cada suíte criou/removeu somente seu schema, migration original e search_path no DSN do pool.
- Revisão independente `query_patterns`: SQL/serviço sem defeito funcional; apontou timestamps potencialmente iguais na prova concorrente. Ajustado para 60 segundos de diferença e reexecutado gate completo com sucesso.
- Todos os cinco critérios aprovados. Implementação validada: sim. Commit local preparado; T1: `3a9b07e`.

Estado final de commit: confirmado `0e19eb4`. Registros anteriores de bloqueio/preparação são históricos e estão resolvidos pelas evidências de fechamento.
