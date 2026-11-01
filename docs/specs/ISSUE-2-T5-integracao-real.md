# ISSUE-2 — T5: Validar jornada real com Keycloak

Spec: [ISSUE-2](./ISSUE-2-autenticacao-keycloak.md)

- **Status:** ⛔ bloqueada
- **Depende de:** T4

## O que fazer
Adicionar suíte Playwright real separada da demo e do smoke offline, com mocks desativados, UI em localhost:5173, Keycloak local e API real. Criar configuração e script test:e2e:auth (propostos), excluir a suíte auth dos configs demo/prod para não contaminar gates offline. Script deve falhar claramente quando pré-requisitos/credencial não estiverem presentes; não produzir sucesso por skip silencioso.
Usar usuário local existente e senha de ambiente, sem versionar segredos ou artefatos contendo tokens/senha (rever traces/screenshots de login). Não automatizar seed ou reset. Para API real, requerer user_account ativo vinculado ao sub; ausência produz 403 e é pré-requisito operacional, não autorização para provisionar.
Cobrir login partindo da raiz → /home e nome, rota interna direta → login → mesma rota com query, recarga autenticada/SSO sem nova digitação, Bearer aceito em endpoint real, logout via UI → Keycloak encerrado → novo acesso exige login, histórico voltar sem conteúdo privado. Validar renovação/expiração com configuração de curta duração em ambiente isolado ou procedimento manual documentado; não alterar duração de sessão no realm compartilhado. Manter prova determinística de falhas/races na T4.
Atualizar README da SPA, removendo afirmação desatualizada de API 501 e documentando modos, variáveis de build, callbacks/logout, execução e limites. Registrar cadastro automático como follow-up sem inventar ID. Rodar gates e registrar evidências por CA, distinguindo testes com substituição do SDK, smoke offline e servidor real.

## Arquivos
- [criar] web/spa/playwright.auth.config.ts
- [criar] web/spa/e2e/auth.spec.ts
- [alterar] web/spa/playwright.config.ts
- [alterar] web/spa/playwright.production.config.ts
- [alterar] web/spa/package.json
- [alterar] web/spa/README.md
- [alterar] infra/keycloak/README.md

## Critérios de aceite
- T5-AC1: Login real pela raiz exibe Home e nome; deep link com query, recarga e SSO funcionam sem mocks → CA1/CA2/CA3/CA4.
- T5-AC2: API aceita Bearer do login e trata ausência de vínculo por 403 sem loop; nenhum header de demo é enviado → CA7/CA8.
- T5-AC3: Sair encerra SSO real e acesso seguinte pede login; prova de refresh/expiração real é registrada por teste ou procedimento isolado → CA5/CA6.
- T5-AC4: Suítes demo/prod/auth isoladas, comandos e pré-requisitos documentados; gates e evidências por CA registrados → CA8.

## Verificação
Na raiz: `npm --prefix web/spa run api:types:check`, `npm --prefix web/spa run lint`, `npm --prefix web/spa run build`, `npm --prefix web/spa test`, `npm --prefix web/spa run test:e2e` e `npm --prefix web/spa run test:e2e:prod`, nessa ordem conforme gate local documentado no README. Pré-requisitos: Node compatível, dependências instaladas (`npm --prefix web/spa ci`), Chromium instalado e portas livres. Esperado: tipos sem drift, compilação/lint/testes passando; E2E demo e smoke offline sem Keycloak real. Para smoke OIDC, compilar com configuração pública de teste e interceptar navegação externa no teste, nunca habilitar mocks no produto.
Suíte proposta após criação: `npm --prefix web/spa run test:e2e:auth` na raiz, com API local 8081, Keycloak local 8080 e UI 5173, usuário vinculado ativo e senha via ambiente; variáveis exatas devem ser documentadas na implementação. É comando novo planejado, ainda inexistente. Manual: provocar expiração/refresh e encerramento SSO em ambiente isolado, registrar resultado e limitações. Não executar seed, sobrescrita de realm ou limpeza de volumes.

## Contexto adicional
Seguir AGENTS.md e docs/go/{architecture,coding,naming}.md; preservar organização feature-first do React e contratos públicos da API. Decisões D1–D5 e premissas P1–P7 na spec. Não fazer cadastro automático. Artefatos de autenticação não podem expor tokens/senha. Cada task deve terminar compilável e com testes atualizados; não deixar regressões para a task seguinte.

## Evidências
<!-- preenchido pelo dev-mage; testes acima são planejados, não executados -->

Pre-flight: AC1/AC3 → Playwright real login/SSO/deep link/logout/refresh; AC2 → API real com Bearer, requer vínculo ativo; AC4 → isolamento suítes e documentação. Fato verificado: Keycloak/API disponíveis fora sandbox; count de vínculo ativo do sub fixture na API = 0. Solicitado recurso de teste ao usuário, sem seed.

Implementação dos testes e documentação pronta; T5 não concluída por impedimento operacional em AC2. Commit local registra entrega parcial, não sucesso da integração completa.

### Evidências observadas — 2026-11-01
| Critério | Comando / sensor | Resultado |
|---|---|---|
| T5-AC1 | test:e2e:auth, cenários Home/SSO e deep link | aprovado: login real, nome Foo Bar, /home, reload sem nova senha, query preservada e Bearer sem headers demo |
| T5-AC2 | test:e2e:auth, real API authorizes the linked identity | impedido: esperado 200, recebido 500; log API reason=lookup_failed. 403 real por vínculo ausente também não comprovado |
| T5-AC3 | test:e2e:auth, cenários logout, token renewal, expired token e logout transport failure | aprovado: logout encerra SSO; refresh real 200; logout noutra aba seguido da expiração exige login; falha de transporte bloqueia shell e retry funciona |
| T5-AC4 | npm ci → tipos → lint → build → Vitest → demo → smoke | aprovado no frontend: 104 Vitest, 10 E2E demo, 8 smoke compilados; suíte real separada: 5 passaram e 1 falhou por AC2 |

Comandos executados na raiz conforme Verificação, com VITE_KEYCLOAK_URL=http://localhost:8080, VITE_KEYCLOAK_REALM=guild-banker e VITE_KEYCLOAK_CLIENT_ID=guild-banker-ui no build. `make keycloak/validate` passou na T2. Na suíte real, a única interceptação de falha de rede está no caso explicitamente denominado logout transport failure; identidade, token/refresh e SSO usam Keycloak real. O relógio virtual existe só no browser, sem alterar o realm. Nenhum artefato auth de trace/screenshot/vídeo é gerado.

Revisão independente: auth_review aprovou T4 após correções e inspecionou isolamento/segurança T5. Observou que 500 não pode ser atribuído apenas ao vínculo ausente: o fonte mapeia ausência para 403. Sem alteração de Go ou API nesta issue.

### Bloqueio — T5-AC2
**Problema:** API local não comprova autorização real; devolve 500/lookup_failed. Além disso, consulta somente leitura confirmou zero user_account ativos vinculados ao sub fixture `00000000-0000-4000-8000-000000000001`.
**Evidência:** `web/spa/e2e/auth.spec.ts`, teste real API authorizes the linked identity, falhou com esperado 200/recebido 500. API em execução logou reason=lookup_failed; não houve ciclo de login. Banco consultado: guild-banker; schema_migrations versão 1, dirty=false. Binário existente tem build vcs.revision=8f54206, vcs.modified=true; isso não prova a causa do 500.
**Impacto:** faltam prova positiva API 200 e negativa real 403; T5 e conclusão integral da spec ficam bloqueadas.
**Proposta:** responsável pelo ambiente deve disponibilizar API correspondente ao código/configuração/schema atual e identidade local já vinculada/ativa, ou autorizar separadamente o trabalho de correção do backend/dados. Reexecutar `npm --prefix web/spa run test:e2e:auth`. Não reduzir a asserção para aceitar 500.
**Confirmação/recurso necessário:** ambiente da API saudável e fixture de identidade vinculada; solicitação assíncrona enviada ao usuário. Cadastro automático continua follow-up, sem seed nesta entrega.

### Ocorrência operacional
A API deixou de responder durante os testes. Foi iniciado o binário existente api/bin/guildbanker para validar a integração. O startup executou automaticamente a rotina de migrações antes de o agente conferir esse efeito; deveria ter sido verificado antes de iniciar. Não foram criadas migrations nem executados comandos de seed/reset. A versão observada após startup foi 1, dirty=false; não há evidência anterior suficiente para afirmar ausência de efeitos. O processo iniciado para teste foi encerrado ao finalizar as verificações.
