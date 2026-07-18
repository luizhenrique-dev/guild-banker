# Prompt: Protótipo Front-end — GuildBanker (UX / validação de ideia)

## Papel
Você é um product engineer / front-end focado em **protótipo de alta fidelidade de UX**.
Objetivo: entregar uma aplicação web **navegável** que valide fluxos, layout e linguagem do produto **GuildBanker**, consumindo (ou mockando de forma explícita) a API Go já esboçada.

Prioridades (nesta ordem):
1. Clareza de UX e fluxos completos do protótipo
2. Fidelidade ao domínio e aos endpoints reais
3. Código simples e legível
4. Stack “bonita de produção” **não** é prioridade

**Anti-alucinação:** não invente endpoints além da lista em `MAPA_TELAS_ENDPOINTS.md`. Se faltar API, use mock local marcado como `MOCK` ou tela “em breve”. Declare e mantenha assumptions em `PROTOTYPE_ASSUMPTIONS.md`.
O design system base do protótipo é o presente no arquivo [DESIGN.md](DESIGN.md) (não invente design).

Leia `PROJECT.md` e `AGENTS.md` se existirem. Em conflito de escopo, **este prompt manda**.

---

## Produto (contexto de UX)

**GuildBanker** — controle financeiro pessoal e familiar.

- **Guild** = workspace compartilhado (ex.: família).
- Membros veem gastos **agregados da guild** e também o recorte **individual**.
- Transações têm **visibilidade**:
    - `PRIVATE` — só o dono (visão individual)
    - `PUBLIC` — compartilhado com a guild (entra no agregado familiar)
- Moeda inicial: **BRL**. Locale default: **pt-BR**, com estrutura extensível para **en**.
- Auth: Keycloak (no protótipo pode ser **login mock** se Keycloak não estiver no ar; documentar como ligar depois).

### Contexto de guild após login
1. Carregar guilds do usuário (`GET /api/v1/guilds`).
2. Ativar: `localStorage.guildbanker.activeGuildId` se ainda for membro; senão a **primeira** da lista.
3. Se zero guilds → onboarding **Criar guild**.
4. Se 2+ guilds → **Guild switcher** sempre acessível (header).
5. Trocar guild recarrega dados dependentes de `guildID`.

---

## Escopo do protótipo (tudo que deve ser clicável)

### Deve existir (fluxo feliz + estados vazios)
1. **Auth** — login / logout (Keycloak real **ou** mock).
2. **Guilds** — listar, criar, renomear, enable/disable, convidar por e-mail, remover membro.
3. **Home / Dashboard da guild** — resumo:
    - total do mês (**PUBLIC** / agregado)
    - meu total individual (inclui `PRIVATE` + minhas `PUBLIC`)
    - próximas despesas fixas (do usuário)
    - atalhos: nova transação, importar fatura, despesas fixas
4. **Despesas fixas (usuário)** — criar, listar ativas, editar, desativar.
5. **Transações (por guild)** — listar, criar, editar, excluir, bulk categorize, alterar **visibility**.
6. **Import de fatura (por guild)** — upload → status → revisar itens → confirm.
7. **Visões Agregado vs Individual** — tabs/toggle em dashboard e transações.
8. **Switcher de guild** + empty/loading/error.
9. **i18n mínimo** — dicionário `pt-BR` / `en` + toggle no header.

### Fora de escopo / superficial ok
- Design system enterprise, E2E, CI, PWA, SSR.
- Charts elaborados (cards + barras CSS bastam).
- Parsing de PDF/CSV no browser (backend processa).
- Categorias customizadas via API (usar enum fixo do backend).

---

## Stack (pragmática)

**Recomendado:** Vite + React + TypeScript leve + Tailwind + React Router.

**Aceitável:** HTML multi-page + Tailwind CDN + JS modules.

Sem over-engineering (sem clean architecture pesada, Storybook, etc.).

Estrutura sugerida (React):
```
web/
src/
app/
pages/
api/
mocks/
i18n/
domain/          # enums: Visibility, Type, Category + labels i18n
state/
styles/
.env.example
PROTOTYPE_ASSUMPTIONS.md
MAPA_TELAS_ENDPOINTS.md
README.md
```

---

## Enums de domínio (fonte da verdade — backend Go)

### Visibility
```ts
type Visibility = 'PRIVATE' | 'PUBLIC';
```
- `PRIVATE` → “Só eu vejo”
- `PUBLIC` → “Compartilhado com a guild” (agregado familiar)

### Transaction Type
```ts
type TransactionType = 'EXPENSE' | 'INCOME';
```

### Category (string enum fixa — não há CRUD de categories na API atual)
```ts
type Category =
  | 'GROCERY'
  | 'HOUSING'
  | 'UTILITIES'
  | 'SUBSCRIPTIONS'
  | 'INSURANCE'
  | 'EDUCATION'
  | 'TRANSPORTATION'
  | 'HEALTH'
  | 'PERSONAL_CARE'
  | 'TAXES'
  | 'OTHER'
  | 'FOOD_AND_DINING'
  | 'ENTERTAINMENT'
  | 'SHOPPING'
  | 'PETS'
  | 'TRAVEL'
  | 'INVESTMENTS';
```

UI: select com **labels pt-BR** (e en) mapeados a partir desses valores. Nunca inventar category fora da lista.

Labels sugeridos (pt-BR):
| Enum | Label |
|------|--------|
| GROCERY | Mercado / supermercado |
| HOUSING | Moradia |
| UTILITIES | Contas (água, luz, gás) |
| SUBSCRIPTIONS | Assinaturas |
| INSURANCE | Seguros |
| EDUCATION | Educação |
| TRANSPORTATION | Transporte |
| HEALTH | Saúde |
| PERSONAL_CARE | Cuidados pessoais |
| TAXES | Impostos |
| OTHER | Outros |
| FOOD_AND_DINING | Alimentação e restaurantes |
| ENTERTAINMENT | Lazer |
| SHOPPING | Compras |
| PETS | Pets |
| TRAVEL | Viagens |
| INVESTMENTS | Investimentos |

---

## Contrato de API (não inventar rotas)

Base: `VITE_API_BASE_URL`  
Header: `Authorization: Bearer <token>`

```
# Guilds
POST   /api/v1/guilds
PUT    /api/v1/guilds/:id
PATCH  /api/v1/guilds/:id/enable
PATCH  /api/v1/guilds/:id/disable
GET    /api/v1/guilds
POST   /api/v1/guilds/:id/invites
DELETE /api/v1/guilds/:id/members/:userID

# Fixed expenses (escopo: USUÁRIO — sem guild na rota)
POST   /api/v1/fixed-expenses
GET    /api/v1/fixed-expenses
PATCH  /api/v1/fixed-expenses/:id
PATCH  /api/v1/fixed-expenses/:id/deactivate

# Transactions (escopo: GUILD)
POST   /api/v1/guilds/:guildID/transactions
GET    /api/v1/guilds/:guildID/transactions
PATCH  /api/v1/guilds/:guildID/transactions/:id
DELETE /api/v1/guilds/:guildID/transactions/:id
POST   /api/v1/guilds/:guildID/transactions:bulk-categorize
PATCH  /api/v1/guilds/:guildID/transactions/:id/visibility

# Imports (escopo: GUILD)
POST   /api/v1/guilds/:guildID/imports
GET    /api/v1/guilds/:guildID/imports/:importID
PATCH  /api/v1/guilds/:guildID/imports/:importID/items/:itemID
DELETE /api/v1/guilds/:guildID/imports/:importID/items/:itemID
POST   /api/v1/guilds/:guildID/imports/:importID:confirm
```

Detalhes de payload, query params e mapa tela→endpoint: ver `MAPA_TELAS_ENDPOINTS.md` e `PROTOTYPE_ASSUMPTIONS.md`.

### Inconsistência de domínio a respeitar na UX
- **Despesas fixas = por usuário** (“Minhas despesas fixas”).
- **Transações/imports = por guild**.
- Agregado familiar = transações com `visibility === 'PUBLIC'`.
- “Meus gastos” = transações do usuário autenticado (PRIVATE + PUBLIC próprias). Se a API não devolver `userId`/`ownerId`, documentar e mockar no protótipo.

---

## Telas e fluxos obrigatórios

### Layout (logado)
- Topbar: logo, **Guild switcher**, idioma PT/EN, usuário, logout
- Nav: Início | Transações | Despesas fixas | Importar | Guild / Membros
- Toggle **Visão da guild (PUBLIC)** vs **Só eu** onde couber

### User stories (aceitação do protótipo)

**G1 — Primeiro acesso**  
Login → sem guild → criar guild → dashboard.

**G2 — Convidar família**  
Membros → convidar e-mail → toast sucesso.

**G3 — Despesa fixa**  
Criar “Aluguel” R$ 2500 dueDay 10 → editar → desativar.

**G4 — Individual vs compartilhado**  
Transação `PRIVATE` só na visão individual.  
Transação `PUBLIC` no agregado.  
Alterar via `PATCH .../visibility`.

**G5 — Import**  
Upload → poll `GET import` → revisar items → confirm → lista de transações.

**G6 — Multi-guild**  
2+ guilds → switcher → persistir `activeGuildId`.

**G7 — Bulk categorize**  
Selecionar N → category enum → `transactions:bulk-categorize`.

---

## Auth no protótipo

**Modo A — Mock** (default se Keycloak vazio ou `VITE_USE_MOCKS=true`)  
“Entrar como demo” com token fake + fixtures nos mesmos paths.

**Modo B — Keycloak**
```
VITE_KEYCLOAK_URL=
VITE_KEYCLOAK_REALM=
VITE_KEYCLOAK_CLIENT_ID=
VITE_API_BASE_URL=
VITE_USE_MOCKS=true|false
```

---

## Import wizard
1. Upload (`multipart` `file`)
2. Processando (poll 2s até status final)
3. Revisão (edit/delete items)
4. Confirmar → Transações + toast

---

## Dashboard
Cards: total PUBLIC do mês, meus gastos, qtd transações, próximas fixas.  
Seletor de mês. Sem lib de chart obrigatória.

---

## Qualidade mínima
- Mobile usable
- loading / empty / error + retry
- confirm em delete/disable/remove
- null-safe
- README + assumptions + mapa de telas
- Zero endpoints inventados

---

## Ordem de implementação
1. Shell + router + i18n + guild context
2. API client + mocks
3. Auth mock/Keycloak stub
4. Guilds + onboarding + switcher
5. Fixed expenses
6. Transactions + visibility + bulk
7. Import wizard
8. Dashboard
9. Polish pt-BR/en
10. README + assumptions

## Critérios de aceite
- [ ] Fluxos G1–G7 com mocks
- [ ] Guild switcher + localStorage
- [ ] PUBLIC vs PRIVATE compreensível na UI
- [ ] Categories apenas do enum Go
- [ ] Fixed expenses sem guildId na URL
- [ ] Transactions/imports com :guildID
- [ ] pt-BR default + en
- [ ] Assumptions documentadas

## Não fazer
- Reescrever backend Go
- OpenAPI completo
- Categories CRUD
- App mobile nativa
