# ISSUE-2 — T2: Preparar adaptador OIDC e configuração

Spec: [ISSUE-2](./ISSUE-2-autenticacao-keycloak.md)

- **Status:** ⬜ pendente
- **Depende de:** T1

## O que fazer
Adicionar `keycloak-js` diretamente, com versão compatível com Keycloak 26.7.5 verificada na implementação e lockfile atualizado; sem wrapper React adicional. Criar fábrica com instância possuída pelo bootstrap (sem singleton global mutável), fronteira mínima para inicialização, obtenção/renovação de token e logout. Nesta task não conectar ao bootstrap: a aplicação existente deve continuar funcionando.
Configuração pública proposta: VITE_KEYCLOAK_URL, VITE_KEYCLOAK_REALM e VITE_KEYCLOAK_CLIENT_ID. Documentar valores locais http://localhost:8080, guild-banker e guild-banker-ui. Validar configuração ausente/inválida antes de redirecionar; nenhuma credencial no build.
Usar Authorization Code, PKCE S256, login-required e tokens apenas em memória. Inicialização única antes do Router; retorno somente à própria origem, conservando pathname/query de negócio/hash de navegação e retirando parâmetros técnicos OIDC após callback. Raiz retorna à Home pela regra T1. Nome: claim name, depois preferred_username, depois saudação sem nome; não renderizar undefined. Exigir sub não vazio.
Preparar URL de logout de volta à raiz da UI. Verificar se o client existente aceita post_logout_redirect_uri: declarar explicitamente os retornos locais permitidos no JSON se necessário, sem ampliar origens. Ajustar validação em prepare.py apenas se a alteração exigir. Não sobrescrever realm existente nem executar importação/migração; documentar atualização manual direcionada no console.
Testes planejados do adaptador com SDK substituído: configuração, PKCE, login-required, redirect local, init único, claims, refresh concorrente e URL de logout. Testes aqui verificam contrato do adaptador, não substituem integração real T5.

## Arquivos
- [criar] web/spa/src/app/keycloak-session.ts
- [criar] web/spa/src/app/keycloak-session.test.ts
- [alterar] web/spa/package.json
- [alterar] web/spa/package-lock.json
- [alterar] web/spa/.env.example
- [alterar] web/spa/src/vite-env.d.ts
- [alterar] infra/keycloak/guild-banker-clients.json
- [alterar] infra/keycloak/prepare.py — somente se necessário à validação
- [alterar] infra/keycloak/README.md

## Critérios de aceite
- T2-AC1: Adaptador testado inicializa Code + PKCE S256 e login-required no client configurado, sem armazenamento persistente de tokens → CA1/CA5.
- T2-AC2: Configuração pública validada e retorno de login/logout restrito à origem; client local documentado e JSON válido → CA6/CA8.
- T2-AC3: Aplicação ainda compila e mantém fluxo atual, pois adaptador só será conectado na T3 → CA8.

## Verificação
Na raiz: `npm --prefix web/spa run api:types:check`, `npm --prefix web/spa run lint`, `npm --prefix web/spa run build`, `npm --prefix web/spa test`, `npm --prefix web/spa run test:e2e` e `npm --prefix web/spa run test:e2e:prod`, nessa ordem conforme gate local documentado no README. Pré-requisitos: Node compatível, dependências instaladas (`npm --prefix web/spa ci`), Chromium instalado e portas livres. Esperado: tipos sem drift, compilação/lint/testes passando; E2E demo e smoke offline sem Keycloak real. Para smoke OIDC, compilar com configuração pública de teste e interceptar navegação externa no teste, nunca habilitar mocks no produto.
Também `make keycloak/validate` na raiz, com Python 3; esperado: JSON válido sem iniciar serviços. Validar dependência/lockfile por instalação limpa.

## Contexto adicional
Seguir AGENTS.md e docs/go/{architecture,coding,naming}.md; preservar organização feature-first do React e contratos públicos da API. Decisões D1–D5 e premissas P1–P7 na spec. Não fazer cadastro automático. Artefatos de autenticação não podem expor tokens/senha. Cada task deve terminar compilável e com testes atualizados; não deixar regressões para a task seguinte.

## Evidências
<!-- preenchido pelo dev-mage; testes acima são planejados, não executados -->
