# GuildBanker SPA

React 18 + TypeScript estrito, Vite, React Router, Tailwind 3, TanStack Query e RHF/Zod. A Home `/home` exibe a saudação depois de resolver uma guild no onboarding obrigatório. As demais telas são `/fixed-expenses`, `/transactions`, `/import` e `/guild`; `/` redireciona para `/home` após autenticação e resolução da guild. Idiomas: português e inglês. Fontes Inter/JetBrains Mono são servidas localmente.

## Instalação e comandos

Execute na raiz do monorepo com Node 24.6.0 (ou versão compatível >=22.12), npm e, para E2E, Chromium:

```sh
npm --prefix web/spa ci
npm --prefix web/spa exec -- playwright install chromium
npm --prefix web/spa run dev
```

O lockfile npm é a fonte da instalação reproduzível; o frontend não utiliza workflow Yarn, geração Prisma ou seed. `make web/install` usa `npm ci`. Não é necessário iniciar API, banco ou Keycloak para validar a demonstração.

| Comando na raiz | Uso |
|---|---|
| `npm --prefix web/spa run dev` / `make web/dev` | Vite em localhost:5173, sem mocks, com login Keycloak |
| `npm --prefix web/spa run dev:demo` / `make web/demo` | Demonstração explicitamente habilitada com MSW e identidade local |
| `npm --prefix web/spa run api:types` | Gerar tipos de `docs/specs/openapi.yaml` em `src/shared/api/schema.d.ts` |
| `npm --prefix web/spa run api:types:check` / `make web/types` | Verificar drift dos tipos versionados |
| `npm --prefix web/spa run lint` / `make web/lint` | ESLint, sem warnings |
| `npm --prefix web/spa test` / `make web/test` | Vitest/Testing Library, HTTP e comportamento com MSW |
| `npm --prefix web/spa run build` / `make web/build` | TypeScript + build estático em `web/spa/dist/` |
| `npm --prefix web/spa run test:e2e` / `make web/e2e` | Fluxos demo em Chromium; inicia Vite com modo test |
| `npm --prefix web/spa run test:e2e:prod` | Smoke de um dist já compilado com vite preview e configuração separada |
| `make web/e2e/prod` | Build e smoke do dist |
| `npm --prefix web/spa run preview` / `make web/preview` | Inspeção local de um dist já compilado |

`vite preview` é ferramenta local de verificação. A produção serve arquivos estáticos; não exige servidor Node frontend, Next, BFF ou rotas API locais.

## Modos, estado e contrato

`VITE_USE_MOCKS=true` só habilita MSW quando `import.meta.env.DEV` também é verdadeiro. `.env.demo` e `.env.test` ativam esse opt-in. O bootstrap aguarda o worker antes de montar; falha de startup aparece na tela. Requests `/api` não reconhecidas falham nos mocks. Produção ignora o opt-in, não inicializa worker nem identidade demo.

A sessão usa Context restrito, dados remotos usam Query, seleção de guild e filtros usam URL; formulários/arquivo selecionado ficam locais. Logout/401 aborta requests e limpa cache; logout voluntário encerra também o SSO, enquanto 401 inicia recuperação de login limitada. Keys incluem usuário/guild/filtros pertinentes; despesas fixas exigem seleção de guild. Todos os membros compartilham as despesas da guild e podem criar, editar e desativar. Consultas e mutations usam `/api/v1/guilds/{guildID}/fixed-expenses`; cache inclui usuário e guild e ações capturam a guild de origem. A troca de guild fecha modais e confirmações; respostas antigas não afetam o novo formulário. O cadastro não altera totais nem gera transações. GET repete no máximo uma vez falhas de rede/5xx (exceto 501); mutations não têm retry automático nem atualização financeira otimista. A revisão de importação é descartada na troca de guild. Os membros exibidos são fixtures explicitamente demonstrativas, pois não existe GET de membros no backend.

Os tipos são gerados do OpenAPI corrigido contra rotas/DTOs Go; a geração não substitui a conferência do contrato. [OPENAPI_GAPS](../../docs/specs/OPENAPI_GAPS.md) registra limites e divergências. Valores monetários são strings decimais no payload, com normalização de vírgula para ponto; não há cálculos financeiros no cliente. Nome da despesa é somente leitura na edição porque PATCH não aceita renomear. A política local de nome tem limite de 100 caracteres.

## API real e variáveis públicas

`VITE_API_URL` vazio usa `/api` na mesma origem. Se definido, deve ser uma URL base pública da API, sem duplicar `/api/v1`; o cliente acrescenta o caminho. Exemplo: `https://api.example.com`. A porta local da API é **8081**; Keycloak usa **8080**. O proxy Vite encaminha `/api` para `http://localhost:8081` apenas no desenvolvimento.

Variáveis `VITE_*` são públicas e substituídas no build. Não coloque tokens, senhas, chaves privadas ou outros segredos nelas; alterar a configuração de build exige recompilar o dist. Use `.env.example` como referência local.

A API valida Bearer pelo Keycloak e resolve o `sub` em `user_account`; cadastro ausente é criado automaticamente, perfil diferente é sincronizado e usuário inativo permanece em **403**. Exige `name`/`email` não brancos e `email_verified=true` no access token; perfil insuficiente e colisão de e-mail também geram 403 genérico. O ator usa ID/e-mail persistidos. Perfil é snapshot sem ordenação entre tokens válidos; login não garante autorização da rota. Headers `X-User-ID`/`X-User-Email` são exclusivos da demonstração. O proxy de desenvolvimento não resolve CORS de produção nem implementa IAM. O contrato de leitura de membros permanece fora desta integração; fixtures não são usadas no modo real. O smoke mock não comprova operação com API Go.

## Entrega estática e hosting futuro

Artefato: **`web/spa/dist/`**, com `index.html`, JS/CSS com hash e fontes locais. A premissa é publicação sob a **raiz `/` do domínio**. Um subpath futuro exige ajustar `base` do Vite e basename do Router. Nenhum hosting/provedor ou pipeline remoto foi configurado ou publicado.

O hosting escolhido deverá:

1. Servir os arquivos existentes do dist e encaminhar `/api` ao backend/gateway, sem fallback SPA nesse prefixo.
2. Retornar `index.html` para navegação direta nas rotas de UI, incluindo refresh das quatro URLs. Rotas de UI desconhecidas mostram a página 404 do Router.
3. Retornar 404 para assets ausentes; nunca devolver HTML para JS/CSS/fontes inexistentes. Servir MIME correto.
4. Usar cache curto/revalidação para `index.html` (por exemplo `Cache-Control: no-cache`) e cache longo imutável para assets com hash (`public, max-age=31536000, immutable`). Não aplicar cache de assets às respostas API.
5. Validar HTTPS, roteamento `/api` e as decisões de IAM/CORS do trabalho de integração.

## Evidências e checklist de publicação

A configuração `playwright.production.config.ts` serve o build com vite preview. O smoke intercepta a navegação OIDC e verifica guarda das rotas, assets, ausência de worker/identidade demo/requests financeiros, PKCE e callback inválido sem loop. A página 404 só aparece após autenticação. Os E2E demo verificam histórico/filtros/guild, importação, membros e CRUD de despesas fixas, incluindo teclado/foco/rolagem local a 375 e 1280 px. Vitest cobre HTTP/cancelamento, sessões/cache, validação e falhas/pending/refetch.

Gate local reproduzível: `npm ci` → tipos sem drift → lint → build → Vitest → E2E demo → smoke do dist. Consultar as evidências observadas nas [tasks GB-20](../../docs/specs/GB-20-reestruturacao-frontend.md). O build pode avisar que o chunk principal tem aproximadamente 500 kB; isso não impede a geração do artefato.

Checklist **futuro**, a executar no provedor escolhido pelo responsável pela publicação:

- [ ] Abrir e recarregar diretamente cada uma das quatro URLs no domínio publicado.
- [ ] Confirmar que `/api` chega ao backend/gateway e não retorna HTML do SPA.
- [ ] Solicitar JS/CSS/fonte inexistente e confirmar 404/MIME, sem fallback HTML.
- [ ] Conferir cache de index e assets com hash, HTTPS e ausência de worker/identidade demo.
- [ ] Validar IAM/CORS/persistência com a API real quando essa integração existir.

O preview local não comprova as regras de fallback/cache de um provedor ainda não escolhido. Publicação e validação de hosting permanecem futuras.

A instalação final ainda reporta **10 avisos npm audit** (2 low, 2 moderate, 6 high), todos em dependências de desenvolvimento no lockfile: ESLint/plugin-kit, Vitest/mocker, PostCSS e a cadeia de glob/watch do Tailwind 3. Esse relatório não é um gate da GB-20; atualizações de segurança da toolchain exigem trabalho próprio, especialmente quando alteram majors. Não foi aplicado `npm audit fix --force`.

## Configuração OIDC
Antes de `dev` ou do build, copie `.env.example` para `.env.local` em `web/spa`.
As variáveis públicas `VITE_KEYCLOAK_URL`, `VITE_KEYCLOAK_REALM` e
`VITE_KEYCLOAK_CLIENT_ID` são obrigatórias fora da demo DEV. Valores locais:
`http://localhost:8080`, `guild-banker`, `guild-banker-ui`.
Configuração ausente ou callback inválido mostra erro sem liberar o shell.
A SPA inicializa OIDC antes do Router com Code + PKCE S256; o login ocorre
na página do Keycloak. `/` leva a `/home` após o onboarding; um link interno conserva sua rota/query.
Nome vem de `name`, depois `preferred_username`, com saudação sem nome como fallback.
O ID de sessão real é o `sub` textual e não o ID numérico interno da API.
Tokens permanecem na memória do SDK; recarga restaura sessão pelo SSO.
O smoke de produção deve receber essas variáveis no build e intercepta apenas
as requisições OIDC no navegador do teste. Ele não comprova login real.
Para usar outra porta no E2E demo: `PLAYWRIGHT_PORT=5273 npm --prefix web/spa run test:e2e`.

### Sessão, expiração e saída
O SDK `keycloak-js` 26.2.5 é usado diretamente, sem wrapper React. Antes de
cada request, `updateToken(30)` renova o token se necessário. O evento de
expiração também tenta renovar. Uma falha definitiva encerra a sessão local e
inicia login; não se repetem mutations. 403, rede e 5xx não provocam logout.
Um marcador sem token em sessionStorage limita recuperação automática entre
recargas; é apagado após uma chamada API bem sucedida. Storage bloqueado impede
recuperação automática e mantém tentativa explícita. Tokens não são persistidos.

Sair limpa IO/cache e bloqueia o conteúdo antes de encerrar SSO. A URL de logout
com ID token hint é preparada antes de limpar tokens. Uma checagem de disponibilidade
do Keycloak (até 5s) permite exibir falha e repetir na UI; falhas depois da navegação
são exibidas pelo próprio navegador. O iframe de sessão fica desabilitado: logout
em outra aba é percebido na renovação/expiração, não necessariamente de imediato.
Configurar retornos de logout conforme [Keycloak local](../../infra/keycloak/README.md).

### Testes reais de autenticação
Na raiz: `npm --prefix web/spa run test:e2e:auth`. Requer UI 5173 livre,
Keycloak local 8080 configurado, API atual na 8081 com banco isolado preparado e access token com nome/e-mail
e verificação verdadeiros; não exige cadastro local prévio. A suíte inicia/encerra apenas o Vite e nunca faz seed, migrations
ou mudanças de política no realm. Pré-requisitos ausentes resultam em falha, sem skip.
A consulta de despesas requer membership: a suíte resolve a guild pela listagem pública.
Em ambiente exclusivo de testes, `E2E_ALLOW_GUILD_CREATE=true` permite criar uma guild
pelo POST público se a lista estiver vazia; não habilite em ambiente compartilhado.
A API deve ser iniciada pelo responsável pelo ambiente, considerando que o startup
atual executa automaticamente sua rotina de migrações. A ISSUE-4 altera a migration
inicial: API e SPA devem ser usadas juntas com banco/schema novo. Um banco que já
executou a versão 1 não recebe essa alteração automaticamente. Não apagar volumes
ou executar force/down em bancos compartilhados para reaplicá-la.

Integração PostgreSQL real: no diretório `api/`,
`go test ./... -race -count=1 -tags=integration,postgres` exige as URLs de teste
`USER_TEST_DATABASE_URL`, `FIXED_EXPENSE_TEST_DATABASE_URL` e `GUILD_TEST_DATABASE_URL`.
Cada suíte cria e limpa apenas seu próprio schema e falha se faltar configuração;
mocks demonstram a interface e não substituem a prova de autorização do backend.

Credenciais: `E2E_USERNAME` (padrão `foo.bar@email.com`), `E2E_NAME` (padrão `Foo Bar`)
e `E2E_PASSWORD` ou `KC_TEST_USER_PASSWORD`. O config lê também `infra/.env` local
se existir, sem substituir variáveis já exportadas. Não grave senha no repositório.
Trace, screenshots e vídeo ficam desabilitados nessa suíte para não registrar
credenciais/códigos/tokens. A demo e o smoke compilado são suítes separadas.

Os testes reais cobrem login/Home, deep link/query, reload/SSO, Bearer na API,
logout/histórico, erro de transporte no logout e renovação/expiração. O relógio do
browser avança até `expires_in` para disparar o timer do SDK; o servidor Keycloak
continua real, sem alterar duração de sessão do realm compartilhado. Outra aba
encerra SSO para provar rejeição real do refresh e retorno ao login.

Gate reproduzível (raiz), com configuração pública OIDC no build:
```sh
npm --prefix web/spa ci
npm --prefix web/spa run api:types:check
npm --prefix web/spa run lint
VITE_KEYCLOAK_URL=http://localhost:8080 VITE_KEYCLOAK_REALM=guild-banker VITE_KEYCLOAK_CLIENT_ID=guild-banker-ui npm --prefix web/spa run build
npm --prefix web/spa test
npm --prefix web/spa run test:e2e
npm --prefix web/spa run test:e2e:prod
npm --prefix web/spa run test:e2e:auth
```

Cadastro automático e sincronização de `user_account` acontecem nas chamadas
autenticadas à API, usando apenas o access token validado. O onboarding precede Home e navegação: loading/erro das consultas necessárias mantém o conteúdo fechado, com identidade, idioma e Sair disponíveis.

Com contexto de guild autorizado, 403 nas consultas de despesas, transações ou importação
mostra acesso negado no conteúdo da rota, com menu, Sair e link Home. O erro tem
prioridade sobre cache anterior; guilds recusadas não ficam no seletor nem iniciam
consultas dependentes. Falha de guilds bloqueia Home e rotas dependentes; 403 da listagem oculta também o menu e não oferece link de volta para o mesmo bloqueio.
403 não inicia login, logout, retry automático ou redirect em loop.

403 de uma ação (criar/editar/desativar despesa, cancelar/alterar visibilidade,
convidar/remover membro, upload/confirmar importação) mostra mensagem de permissão
no feedback local existente, sem navegar ou encerrar sessão. Formulário, arquivo
e revisão ficam preservados; não há sucesso, replay ou limpeza de rascunho.
Confirmação explicitamente negada não invalida transações; cancelamento sem
resposta mantém a invalidação do contexto original para consistência.

## Onboarding de guild (ISSUE-5)

Após autenticar, a SPA consulta as guilds. Uma guild é selecionada automaticamente; com várias, URL válida tem prioridade, seguida da preferência válida salva por identidade. Sem escolha válida, o usuário escolhe explicitamente. URL inválida não é substituída silenciosamente. Rotas e filtros são preservados.

Somente o ID é salvo em localStorage (`guildbanker:last-guild:demo|oidc:<identidade>`). A preferência sobrevive ao logout e é sempre validada contra memberships retornadas pela API. Storage bloqueado ou inválido não impede uso. Tokens continuam somente em memória.

Sem guilds, a SPA consulta convites pendentes. Os cartões mostram guild/descrição/convidador (nome/e-mail), com Aceitar e Negar. Recusar aguarda nova consulta antes de oferecer criação. Aceitar ou criar associa o usuário no servidor e seleciona imediatamente a guild retornada. Erros não são interpretados como lista vazia; callbacks antigos não sobrescrevem 403 nem sessão encerrada.

Criação usa apenas `display_name` (1–70 caracteres Unicode) e descrição opcional. A API gera `name` único e imutável, sem acentos, lowercase, com espaços convertidos em traços, somente letras ASCII/números/traços e máximo 70; colisões recebem sufixo. DisplayName pode repetir; seleção/convites mostram o identificador para distinguir homônimos. PUT recebe `display_name` e descrição (omitir preserva; string vazia limpa), nunca altera `name`.

Convites são persistidos para qualquer e-mail antes do primeiro login, sem expiração. Não concedem membership até o aceite. Duplicata pendente e membro existente retornam 409 com códigos distintos e feedback localizado; após recusa é possível reenviar. Não há envio de e-mail nesta issue.

Migrações 000002/000003 são incrementais e preservam memberships/dados anteriores. Não executar force/down em banco compartilhado. Os testes reais podem criar a primeira guild pela UI apenas com `E2E_ALLOW_GUILD_CREATE=true` em ambiente exclusivo. A suíte demo usa fixtures para a jornada; os testes PostgreSQL reais provam autorização, concorrência e rollback.
