# GuildBanker SPA

Aplicação frontend moderna para gestão financeira pessoal colaborativa.
Projetada como material de referência/estudo para engenheiros aprendendo React + TypeScript moderno.

## Stack Técnica

| Tecnologia | Versão | Propósito |
|---|---|---|
| Next.js | 14+ (App Router) | Framework React com SSR/SSG |
| TypeScript | 5.2 (strict + noUncheckedIndexedAccess) | Tipagem forte |
| Tailwind CSS | 3 | Estilização utility-first |
| TanStack Query | 5 | Data fetching + cache |
| Zustand | 5 | Estado global minimal |
| React Hook Form | 7 | Gerenciamento de formulários |
| Zod | 3 | Validação de schemas |
| MSW | 2 | Mock Service Worker (API mock) |
| react-i18next | - | Internacionalização PT-BR / EN |
| Axios | - | HTTP client com interceptors |

## Estrutura de Pastas (Feature-Based Architecture)

```
src/
  app/                          # Next.js App Router (páginas)
  features/
    fixed-expenses/             # Feature: Despesas fixas
      api/                      # Hooks TanStack Query
      components/               # Componentes React da feature
      schemas/                  # Schemas Zod (validação)
      types.ts                  # Tipos TypeScript
    transactions/               # Feature: Transações
    guilds/                     # Feature: Guilds/Membros
    imports/                    # Feature: Importação CSV
  shared/
    components/                 # Componentes reutilizáveis (Sidebar, Header, Modal...)
    hooks/                      # Hooks utilitários (useDebounce)
    lib/
      api-client.ts             # Axios com interceptors
      query-client.ts           # Configuração TanStack Query
      i18n.ts                   # Setup i18n (PT-BR/EN)
    store/
      auth.store.ts             # Zustand: usuário logado
      guild.store.ts            # Zustand: guild selecionado
    types/                      # Tipos compartilhados
  mocks/
    handlers/                   # MSW handlers por recurso
    fixtures/                   # Dados mock realistas
    browser.ts                  # MSW browser setup
```

## Como Rodar

```bash
# Instalar dependências
yarn install

# Rodar em desenvolvimento (MSW ativo por padrão)
yarn dev
```

Acesse http://localhost:3000 — a aplicação redireciona para `/fixed-expenses`.

## Como Conectar na API Real

### Passo a passo:

1. **Defina as variáveis de ambiente** no `.env.local`:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_USE_MOCK=false
```

2. **Autenticação atual (headers):**

A API usa `X-User-ID` e `X-User-Email` como forma temporária de autenticação
(validação JWT via Keycloak ainda não implementada no backend).

O `api-client.ts` já injeta esses headers automaticamente a partir do Zustand store.

3. **Quando o JWT estiver pronto:**

- Integre o Keycloak JS adapter
- Armazene o `access_token` no Zustand
- O interceptor Axios já envia `Authorization: Bearer <token>`
- Remova os headers `X-User-ID` / `X-User-Email` do interceptor

4. **Desabilitar MSW:**

Basta definir `NEXT_PUBLIC_USE_MOCK=false`. O MSW Provider não inicializará o worker.

## Decisões de Arquitetura

### Por que Feature-Based Architecture?
- Cada feature é autossuficiente (types, api, components, schemas)
- Fácil de navegar e escalar
- Sem dependências circulares entre features

### Por que TanStack Query em vez de useEffect + fetch?
- Cache automático e deduplicação de requests
- Loading/error states consistentes sem boilerplate
- Invalidação inteligente após mutations

### Por que Zustand em vez de Context API?
- Funciona fora de componentes React (ex: interceptors Axios)
- API simples, sem Provider hell
- Performance: re-renders granulares via selectors

### Por que MSW?
- Intercepta requests no nível de rede (fetch/XHR)
- Handlers espelham a API real (fácil migração)
- Estado in-memory permite testar mutations

### Por que Zod + React Hook Form?
- Validação declarativa com type inference
- Schemas reutilizáveis entre client e server
- Integração nativa via zodResolver

## Padrões Educacionais

Procure por estes comentários no código:
- `💡 REAL API:` — Como seria a integração com o backend real
- `📌 PADRÃO:` — Explicação do padrão utilizado
- `⚠️ ATENÇÃO:` — Cuidados e tradeoffs

## Usuário Mock

- **Nome:** Ana Silva
- **Email:** ana@familia.com
- **Guild:** Família Silva

## API Reference

A especificação OpenAPI completa está em `openapi.yaml`.
