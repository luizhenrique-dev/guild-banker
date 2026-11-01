# ISSUE-2 — T1: Criar Home e navegação inicial

Spec: [ISSUE-2](./ISSUE-2-autenticacao-keycloak.md)

- **Status:** ✅ concluida
- **Depende de:** nenhuma

## O que fazer
Criar `/home` protegida pela sessão existente, contendo somente a saudação “Bem-vindo(a) <nome da pessoa>”, usando interpolação React e i18n pt/en existente. Não condicionar o conteúdo da Home à seleção de guild. Alterar `/` para `/home`, preservando query de negócio, e o item Início para `/home`, com estado ativo correto. Manter shell existente; a proteção integral será conectada na T3.
Atualizar testes unitários e E2E demo que esperam despesas fixas como destino da raiz. Testes planejados: saudação, nome como texto (sem HTML), Home sem guild, clique Início, raiz e histórico/query.

## Arquivos
- [alterar] web/spa/src/app/app.tsx
- [alterar] web/spa/src/shared/components/sidebar.tsx
- [alterar] web/spa/src/shared/lib/i18n.ts
- [alterar] web/spa/src/app/app.test.tsx
- [alterar] web/spa/e2e/navigation.spec.ts
- [alterar] web/spa/e2e/production.spec.ts
- [criar] web/spa/src/features/home/components/home-page.tsx
- [criar] web/spa/src/features/home/components/home-page.test.tsx

## Critérios de aceite
- T1-AC1: Raiz e menu Início levam a /home; saudação contém nome da sessão e funciona sem guild → CA3.
- T1-AC2: Testes de navegação/demo/smoke ajustados ao novo destino passam; guarda existente preservada → CA3.

## Verificação
Na raiz: `npm --prefix web/spa run api:types:check`, `npm --prefix web/spa run lint`, `npm --prefix web/spa run build`, `npm --prefix web/spa test`, `npm --prefix web/spa run test:e2e` e `npm --prefix web/spa run test:e2e:prod`, nessa ordem conforme gate local documentado no README. Pré-requisitos: Node compatível, dependências instaladas (`npm --prefix web/spa ci`), Chromium instalado e portas livres. Esperado: tipos sem drift, compilação/lint/testes passando; E2E demo e smoke offline sem Keycloak real. Para smoke OIDC, compilar com configuração pública de teste e interceptar navegação externa no teste, nunca habilitar mocks no produto.

## Contexto adicional
Seguir AGENTS.md e docs/go/{architecture,coding,naming}.md; preservar organização feature-first do React e contratos públicos da API. Decisões D1–D5 e premissas P1–P7 na spec. Não fazer cadastro automático. Artefatos de autenticação não podem expor tokens/senha. Cada task deve terminar compilável e com testes atualizados; não deixar regressões para a task seguinte.

## Evidências
<!-- preenchido pelo dev-mage; testes acima são planejados, não executados -->

Pre-flight: T1-AC1 → teste de raiz/menu/saudação sem guild (planejado); T1-AC2 → E2E demo e smoke. Baseline: app/sessão/bootstrap, 13 testes passaram.

T1-AC1 aprovado: app.test.tsx, raiz/menu, query e saudação sem guild com HTML escapado. RED observado antes da implementação (saudação ausente). T1-AC2 aprovado: tipos/lint/build, 81 Vitest, 10 E2E demo (PLAYWRIGHT_PORT=5273) e 6 smoke dist passaram. Revisão própria: diff restrito à Home/navegação; porta configurável adicionada pois 5173 já ocupada. Commit local: incluído no commit desta task.
