# ISSUE-2 — T4: Completar renovação, expiração e logout

Spec: [ISSUE-2](./ISSUE-2-autenticacao-keycloak.md)

- **Status:** ⬜ pendente
- **Depende de:** T3

## O que fazer
Antes de cada request real, renovar token próximo de expirar por updateToken(30) e ler o token atual depois da renovação. Compartilhar renovação em andamento entre requests concorrentes. Em token expirado sem refresh válido, invalidar sessão local e encaminhar ao login uma vez. Evento de expiração do SDK também deve tentar renovar; sem polling próprio ou timers sem cleanup. Não prometer detecção imediata de logout remoto em todos os browsers.
Separar invalidação local de logout voluntário: 401 ou falha definitiva de refresh invalidam e iniciam reautenticação, sem chamar logout global automaticamente. Não repetir mutations após 401. Não tratar 403, timeout/rede ou 5xx como ausência de sessão; apresentar erro existente sem ciclo de login. Falha de rede na renovação bloqueia o envio com token vencido e permite tentativa explícita, sem loop.
No botão Sair: bloquear conteúdo imediatamente, abortar IO, cancelar/limpar cache e callbacks tardios; encerrar SSO por logout do SDK usando id_token_hint/redirect apropriados. Preparar URL de logout antes de descartar dados necessários ao SDK. Retornar à raiz; login-required deve apresentar login, sem reabrir sessão anterior. Proteger transição contra corrida entre guarda e logout. Se Keycloak estiver indisponível, sessão local fica encerrada e UI explica falha de encerramento remoto, oferecendo repetir; não alegar sucesso remoto.
Cada nova vida de sessão tem AbortController novo e cache isolado. Preservar garantias existentes de cancelamento e ausência de efeitos tardios.
Testes planejados: refresh pré-request e concorrente, token renovado, expiração recuperável/definitiva, erro de rede, 401 sem replay, 403 sem logout, logout local/remoto, falha remota, request/mutation tardia e login de outro usuário.

## Arquivos
- [alterar] web/spa/src/app/keycloak-session.ts — proposto, criado na T2
- [alterar] web/spa/src/app/keycloak-session.test.ts — proposto, criado na T2
- [alterar] web/spa/src/app/session-provider.tsx
- [alterar] web/spa/src/app/session-provider.test.tsx
- [alterar] web/spa/src/app/app.tsx
- [alterar] web/spa/src/shared/components/header.tsx
- [alterar] web/spa/src/shared/lib/i18n.ts

## Critérios de aceite
- T4-AC1: Requests concorrentes compartilham refresh e usam token renovado; falha definitiva conduz ao login uma vez, sem replay de mutations → CA5.
- T4-AC2: Logout UI bloqueia shell, cancela IO/cache e solicita fim do SSO; falha remota fica explícita sem restaurar sessão local → CA6.
- T4-AC3: 401 reautentica; 403/rede/5xx não encerram SSO nem geram loop; respostas tardias não afetam sessão seguinte → CA7.

## Verificação
Na raiz: `npm --prefix web/spa run api:types:check`, `npm --prefix web/spa run lint`, `npm --prefix web/spa run build`, `npm --prefix web/spa test`, `npm --prefix web/spa run test:e2e` e `npm --prefix web/spa run test:e2e:prod`, nessa ordem conforme gate local documentado no README. Pré-requisitos: Node compatível, dependências instaladas (`npm --prefix web/spa ci`), Chromium instalado e portas livres. Esperado: tipos sem drift, compilação/lint/testes passando; E2E demo e smoke offline sem Keycloak real. Para smoke OIDC, compilar com configuração pública de teste e interceptar navegação externa no teste, nunca habilitar mocks no produto.

## Contexto adicional
Seguir AGENTS.md e docs/go/{architecture,coding,naming}.md; preservar organização feature-first do React e contratos públicos da API. Decisões D1–D5 e premissas P1–P7 na spec. Não fazer cadastro automático. Artefatos de autenticação não podem expor tokens/senha. Cada task deve terminar compilável e com testes atualizados; não deixar regressões para a task seguinte.

## Evidências
<!-- preenchido pelo dev-mage; testes acima são planejados, não executados -->
