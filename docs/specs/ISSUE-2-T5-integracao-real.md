# ISSUE-2 — T5: Validar jornada real com Keycloak

Spec: [ISSUE-2](./ISSUE-2-autenticacao-keycloak.md)

- **Status:** ⬜ pendente
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
