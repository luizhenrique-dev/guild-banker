# GuildBanker — Prototype Assumptions

Documento vivo das decisões e inferências do protótipo front-end.
Quando a OpenAPI oficial sair, atualize este arquivo e o client em `web/src/api`.

Última atualização: 2026-07-14  
Fonte de rotas: handlers Go informados pelo time (`/api/v1/...`)  
Fonte de enums: tipos Go `Visibility`, `Type`, `Category`

---

## 1. Produto e multi-tenancy

| # | Assumption | Status |
|---|------------|--------|
| A1 | **Guild** = workspace compartilhado (família / grupo), não apenas “conta pessoal”. | Confirmado |
| A2 | O usuário pode ser membro de **várias guilds**. | Confirmado |
| A3 | Após login, contexto ativo = `localStorage.guildbanker.activeGuildId` se ainda membro; senão **primeira** guild de `GET /guilds`. | Confirmado |
| A4 | Com 0 guilds → onboarding obrigatório de criação. | Confirmado |
| A5 | Visão **agregada** da família usa transações `visibility = PUBLIC`. | Inferido (UX) |
| A6 | Visão **individual** mostra gastos do usuário autenticado, incluindo `PRIVATE` e as próprias `PUBLIC`. | Inferido (UX) |

---

## 2. Escopo de recursos na API (assimetria importante)

| Recurso | Escopo na rota | Implicação de UX |
|---------|----------------|------------------|
| Guilds / invites / members | Usuário + `:id` da guild | Gestão do workspace |
| **Fixed expenses** | **Usuário autenticado** (sem `:guildID`) | UI: “**Minhas** despesas fixas” — não “fixas da família” |
| Transactions | `:guildID` | Gastos no contexto da guild |
| Imports | `:guildID` | Fatura importada para a guild ativa |
| Categories | **Sem endpoints** | Enum fixo no client (ver §4) |

**Risco de produto:** despesas fixas não são compartilhadas entre membros da guild na API atual. O protótipo **não deve** fingir fixas familiares agregadas. Se no futuro fixas forem por guild, a API e este doc mudam juntos.

---

## 3. Enums canônicos

### 3.1 Visibility
```go
VisibilityPrivate = "PRIVATE"
VisibilityPublic  = "PUBLIC"
```

| Valor | Copy pt-BR | Copy en | Uso no agregado |
|-------|------------|---------|-----------------|
| `PRIVATE` | Só eu vejo | Only me | Não |
| `PUBLIC` | Compartilhado com a guild | Shared with guild | Sim |

Endpoint: `PATCH /api/v1/guilds/:guildID/transactions/:id/visibility`  
Body assumido:
```json
{ "visibility": "PRIVATE" | "PUBLIC" }
```

### 3.2 Type (transação)
```go
TypeExpense = "EXPENSE"
TypeIncome  = "INCOME"
```

### 3.3 Category
Valores exatos (string):

- `GROCERY`
- `HOUSING`
- `UTILITIES`
- `SUBSCRIPTIONS`
- `INSURANCE`
- `EDUCATION`
- `TRANSPORTATION`
- `HEALTH`
- `PERSONAL_CARE`
- `TAXES`
- `OTHER`
- `FOOD_AND_DINING`
- `ENTERTAINMENT`
- `SHOPPING`
- `PETS`
- `TRAVEL`
- `INVESTMENTS`

**Não existem** categories `SYSTEM`/`USER`/`AI` no protótipo atual — o PROJECT.md descreve evolução futura; a API atual usa este enum fechado.

Default ao criar sem escolha: `OTHER`.

---

## 4. Payloads HTTP assumidos (ajustáveis)

> Se o backend retornar 400/422, adaptar **somente** o adapter do client e registrar a diferença aqui.

### 4.1 Guilds
**POST /guilds**
```json
{ "name": "familia-silva", "displayName": "Família Silva" }
```
- `name`: slug/único (inferido do PROJECT)
- Se API aceitar só um campo, enviar `name` e mapear display na UI.

**PUT /guilds/:id** (UpdateName)
```json
{ "name": "novo-nome" }
```
Alternativa possível: `{ "displayName": "..." }` — validar no primeiro integrate.

**POST /guilds/:id/invites**
```json
{ "email": "pessoa@email.com" }
```

**PATCH enable/disable:** body vazio.

### 4.2 Fixed expenses
**POST /fixed-expenses**
```json
{
  "name": "Aluguel",
  "amount": 2500.0,
  "dueDay": 10,
  "category": "HOUSING",
  "currency": "BRL"
}
```

**PATCH /fixed-expenses/:id** — partial dos mesmos campos.

**PATCH /fixed-expenses/:id/deactivate** — body vazio **ou**
```json
{ "status": "PAUSED" | "CANCELLED" }
```
Protótipo tenta body vazio primeiro; se necessário, envia status.

**GET /fixed-expenses** — lista ativas do usuário.  
Query opcional assumida: nenhuma obrigatória.

Campos de resposta assumidos:
```ts
{
  id: string
  name: string
  amount: number
  dueDay: number
  category?: Category
  currency?: "BRL"
  status?: "ACTIVE" | "PAUSED" | "CANCELLED"
  createdAt?: string
  updatedAt?: string
}
```

### 4.3 Transactions
**POST /guilds/:guildID/transactions**
```json
{
  "description": "Mercado",
  "amount": 189.9,
  "type": "EXPENSE",
  "date": "2026-07-01",
  "category": "GROCERY",
  "visibility": "PUBLIC"
}
```

**PATCH /guilds/:guildID/transactions/:id** — partial.

**GET /guilds/:guildID/transactions**  
Query assumida (todas opcionais):
- `from`, `to` (ISO date)
- `type`, `category`, `visibility`
- `page`, `pageSize` (se backend paginar; senão client pagina em memória no mock)

Resposta assumida (uma das duas):
- Array puro `Transaction[]`
- Ou `{ items, page, pageSize, totalItems }`

Client deve aceitar **array** e normalizar para lista interna.

Campos de resposta assumidos:
```ts
{
  id: string
  guildId?: string
  userId?: string        // necessário para “meus gastos”; se ausente → MOCK/assumption
  description: string
  amount: number
  type: "EXPENSE" | "INCOME"
  date: string
  category?: Category
  visibility: "PRIVATE" | "PUBLIC"
  source?: string
  createdAt?: string
  updatedAt?: string
}
```

**POST .../transactions:bulk-categorize**
```json
{
  "ids": ["uuid1", "uuid2"],
  "category": "TRANSPORTATION"
}
```
Alternativa se 400: `transactionIds` no lugar de `ids`.

### 4.4 Imports
**POST /guilds/:guildID/imports**  
`multipart/form-data`:
- `file`: arquivo CSV ou PDF
- `type` ou `source` opcional: `CSV` | `PDF` (se backend ignorar, ok)

**GET .../imports/:importID**
```ts
{
  id: string
  status: "PENDING" | "PROCESSING" | "COMPLETED" | "FAILED"
  items: ImportItem[]
  errorMessage?: string
}
```

**ImportItem** assumido:
```ts
{
  id: string
  description: string
  amount: number
  date: string
  type?: "EXPENSE" | "INCOME"
  category?: Category
  visibility?: "PRIVATE" | "PUBLIC"
}
```

**PATCH .../items/:itemID** — partial do item.  
**DELETE .../items/:itemID** — remove da revisão.  
**POST .../imports/:importID:confirm** — body vazio; materializa transações.

Polling: 2s enquanto `PENDING` | `PROCESSING`.

---

## 5. Auth

| # | Assumption |
|---|------------|
| A10 | Todas as rotas `/api/v1/*` exigem JWT (middleware Auth). |
| A11 | Protótipo pode rodar com **mock auth** + **mock API** (`VITE_USE_MOCKS=true`) para demo de UX. |
| A12 | Não há endpoint `/me` na lista atual → user no protótipo vem do token Keycloak/mock (id, email, name). |
| A13 | Convite por e-mail: UX de sucesso mesmo que o backend só registre o invite (sem deep link no protótipo). |

---

## 6. i18n e formatação

| # | Assumption |
|---|------------|
| A20 | Default `pt-BR`; toggle para `en`. |
| A21 | Moeda sempre BRL no protótipo (`Intl` pt-BR / en-US com currency BRL). |
| A22 | Strings de UI centralizadas; enums de domínio **não** se traduzem no wire (só labels). |

---

## 7. O que o PROJECT.md tem e a API ainda não

| Conceito PROJECT.md | No protótipo |
|---------------------|--------------|
| Category CRUD / rules / AI | Fora — enum fixo |
| Bank extract / PIX | Fora |
| Dashboard charts avançados | Cards simples |
| User profile sync endpoint | Mock/Keycloak claims |
| Generate monthly entries from fixed expenses | Backend/job — front só CRUD de fixas |
| Import review | **Dentro** — coberto pelos endpoints de import |

---

## 8. Riscos / perguntas em aberto para o backend

1. Shape exato de create/update guild (`name` vs `displayName`).
2. Fixed expense: campo `category` usa o mesmo enum `Category`?
3. Transaction GET devolve `userId`/`ownerId`?
4. Paginação real vs lista completa.
5. Bulk categorize: nome do campo `ids` vs `transactionIds` e `category` vs `categoryId`.
6. Visibility default no create e no import confirm.
7. Import: nome do campo multipart e status machine exata.
8. Enable/disable guild: efeito em membros e em listagens.

---

## 9. Regras anti-alucinação para o agente de código

1. Não criar rotas fora da lista Go.
2. Não criar valores de `Category` / `Visibility` / `Type` fora dos enums.
3. Não misturar fixed-expenses com path de guild.
4. Mocks devem espelhar os **mesmos paths**.
5. Toda inferência de body deve aparecer neste arquivo.
