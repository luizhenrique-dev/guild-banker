# GuildBanker SPA

React 18 + TypeScript estrito, Vite, React Router, Tailwind 3, TanStack Query e RHF/Zod. As quatro telas existentes são `/fixed-expenses`, `/transactions`, `/import` e `/guild`; `/` redireciona para despesas fixas. Idiomas: português e inglês. Fontes Inter/JetBrains Mono são servidas localmente.

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
| `npm --prefix web/spa run dev` / `make web/dev` | Vite em localhost:5173, sem mocks e sem sessão integrada |
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

A sessão usa Context restrito, dados remotos usam Query, seleção de guild e filtros usam URL; formulários/arquivo selecionado ficam locais. Logout/401 aborta requests e limpa cache. Keys incluem usuário/guild/filtros pertinentes; despesas fixas são pessoais. GET repete no máximo uma vez falhas de rede/5xx (exceto 501); mutations não têm retry automático nem atualização financeira otimista. A revisão de importação é descartada na troca de guild. Os membros exibidos são fixtures explicitamente demonstrativas, pois não existe GET de membros no backend.

Os tipos são gerados do OpenAPI corrigido contra rotas/DTOs Go; a geração não substitui a conferência do contrato. [OPENAPI_GAPS](../../docs/specs/OPENAPI_GAPS.md) registra limites e divergências. Valores monetários são strings decimais no payload, com normalização de vírgula para ponto; não há cálculos financeiros no cliente. Nome da despesa é somente leitura na edição porque PATCH não aceita renomear. A política local de nome tem limite de 100 caracteres.

## API real e variáveis públicas

`VITE_API_URL` vazio usa `/api` na mesma origem. Se definido, deve ser uma URL base pública da API, sem duplicar `/api/v1`; o cliente acrescenta o caminho. Exemplo: `https://api.example.com`. A porta local da API é **8081**; Keycloak usa **8080**. O proxy Vite encaminha `/api` para `http://localhost:8081` apenas no desenvolvimento.

Variáveis `VITE_*` são públicas e substituídas no build. Não coloque tokens, senhas, chaves privadas ou outros segredos nelas; alterar a configuração de build exige recompilar o dist. Use `.env.example` como referência local.

A integração real está pendente de outra issue: endpoints protegidos retornam **401 sem Bearer** e **501 com Bearer** no middleware atual. Headers `X-User-ID`/`X-User-Email` não liberam autenticação e não estão liberados no CORS atual. O proxy de desenvolvimento não resolve CORS de produção nem implementa IAM. OIDC/sessão real, autorização, CORS, persistência real e contrato de leitura de membros precisam de implementação/validação própria. O smoke mock não comprova operação com API Go.

## Entrega estática e hosting futuro

Artefato: **`web/spa/dist/`**, com `index.html`, JS/CSS com hash e fontes locais. A premissa é publicação sob a **raiz `/` do domínio**. Um subpath futuro exige ajustar `base` do Vite e basename do Router. Nenhum hosting/provedor ou pipeline remoto foi configurado ou publicado nesta issue.

O hosting escolhido deverá:

1. Servir os arquivos existentes do dist e encaminhar `/api` ao backend/gateway, sem fallback SPA nesse prefixo.
2. Retornar `index.html` para navegação direta nas rotas de UI, incluindo refresh das quatro URLs. Rotas de UI desconhecidas mostram a página 404 do Router.
3. Retornar 404 para assets ausentes; nunca devolver HTML para JS/CSS/fontes inexistentes. Servir MIME correto.
4. Usar cache curto/revalidação para `index.html` (por exemplo `Cache-Control: no-cache`) e cache longo imutável para assets com hash (`public, max-age=31536000, immutable`). Não aplicar cache de assets às respostas API.
5. Validar HTTPS, roteamento `/api` e as decisões de IAM/CORS do trabalho de integração.

## Evidências e checklist de publicação

A configuração `playwright.production.config.ts` serve o build com vite preview. O smoke verifica acesso direto às quatro telas, assets sem erro, ausência de worker/identidade demo/requests financeiros, estado sem sessão, redirect e 404 de UI. Os E2E demo verificam histórico/filtros/guild, importação, membros e CRUD de despesas fixas, incluindo teclado/foco/rolagem local a 375 e 1280 px. Vitest cobre HTTP/cancelamento, sessões/cache, validação e falhas/pending/refetch.

Gate local reproduzível: `npm ci` → tipos sem drift → lint → build → Vitest → E2E demo → smoke do dist. Consultar as evidências observadas nas [tasks GB-20](../../docs/specs/GB-20-reestruturacao-frontend.md). O build pode avisar que o chunk principal tem aproximadamente 500 kB; isso não impede a geração do artefato.

Checklist **futuro**, a executar no provedor escolhido pelo responsável pela publicação:

- [ ] Abrir e recarregar diretamente cada uma das quatro URLs no domínio publicado.
- [ ] Confirmar que `/api` chega ao backend/gateway e não retorna HTML do SPA.
- [ ] Solicitar JS/CSS/fonte inexistente e confirmar 404/MIME, sem fallback HTML.
- [ ] Conferir cache de index e assets com hash, HTTPS e ausência de worker/identidade demo.
- [ ] Validar IAM/CORS/persistência com a API real quando essa integração existir.

O preview local não comprova as regras de fallback/cache de um provedor ainda não escolhido. Publicação e validação de hosting permanecem futuras.

A instalação final ainda reporta **10 avisos npm audit** (2 low, 2 moderate, 6 high), todos em dependências de desenvolvimento no lockfile: ESLint/plugin-kit, Vitest/mocker, PostCSS e a cadeia de glob/watch do Tailwind 3. Esse relatório não é um gate da GB-20; atualizações de segurança da toolchain exigem trabalho próprio, especialmente quando alteram majors. Não foi aplicado `npm audit fix --force`.
