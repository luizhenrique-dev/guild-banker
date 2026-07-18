# GuildBanker — Mapa Telas → Endpoints

Referência rápida para o protótipo front-end.  
Enums: ver `PROTOTYPE_ASSUMPTIONS.md`.

Legenda de métodos: ações de UI → HTTP.

---

## Navegação (rotas do front sugeridas)

| Rota front | Tela | Auth |
|------------|------|------|
| `/login` | Login | pública |
| `/onboarding/guild` | Criar primeira guild | privada |
| `/` ou `/dashboard` | Início da guild ativa | privada + guild |
| `/transactions` | Transações | privada + guild |
| `/fixed-expenses` | Minhas despesas fixas | privada |
| `/imports` | Lista/atalho import | privada + guild |
| `/imports/new` | Wizard upload | privada + guild |
| `/imports/:importId` | Processamento + revisão | privada + guild |
| `/guilds` | Minhas guilds | privada |
| `/guilds/:id/members` | Membros e convites | privada |

`activeGuildId` vem do state/localStorage (não precisa estar na URL, mas pode: `?guild=` opcional).

---

## 1. Auth / sessão

| UI | Endpoint | Notas |
|----|----------|-------|
| Entrar (Keycloak) | — (IdP) | Depois API com Bearer |
| Entrar demo | — | Mock local |
| Logout | — | Limpa token + state |
| Bootstrap user | — | Claims token / mock; **sem** `/me` na API atual |

---

## 2. Guild context

| UI | Endpoint | Body / params |
|----|----------|---------------|
| Carregar guilds (bootstrap, switcher) | `GET /api/v1/guilds` | — |
| Criar guild (onboarding ou settings) | `POST /api/v1/guilds` | `{ name, displayName? }` |
| Renomear guild | `PUT /api/v1/guilds/:id` | `{ name }` (assumption) |
| Ativar guild | `PATCH /api/v1/guilds/:id/enable` | — |
| Desativar guild | `PATCH /api/v1/guilds/:id/disable` | — |
| Convidar membro | `POST /api/v1/guilds/:id/invites` | `{ email }` |
| Remover membro | `DELETE /api/v1/guilds/:id/members/:userID` | — |
| Trocar guild ativa | — (client only) | `localStorage` + invalidate queries |

---

## 3. Dashboard (`/` )

| UI / dado | Endpoint | Filtro client |
|-----------|----------|---------------|
| Totais do mês (agregado) | `GET /api/v1/guilds/:guildID/transactions` | `visibility === 'PUBLIC'`, mês corrente, `type` |
| Meus gastos | mesmo GET | dono = user atual; inclui PRIVATE |
| Próximas fixas | `GET /api/v1/fixed-expenses` | ordenar por `dueDay` |
| CTA Nova transação | navega form | — |
| CTA Importar | `/imports/new` | — |
| CTA Fixas | `/fixed-expenses` | — |

> Dashboard **não** tem endpoint próprio na API atual — é composição client-side.

---

## 4. Despesas fixas (`/fixed-expenses`)

| UI | Endpoint |
|----|----------|
| Listar ativas | `GET /api/v1/fixed-expenses` |
| Criar | `POST /api/v1/fixed-expenses` |
| Editar | `PATCH /api/v1/fixed-expenses/:id` |
| Desativar | `PATCH /api/v1/fixed-expenses/:id/deactivate` |

**Não usar** `:guildID` nestas rotas.

Campos de formulário: `name`, `amount`, `dueDay` (1–28), `category` (enum), opcional `currency: BRL`.

---

## 5. Transações (`/transactions`)

| UI | Endpoint |
|----|----------|
| Listar | `GET /api/v1/guilds/:guildID/transactions` |
| Criar | `POST /api/v1/guilds/:guildID/transactions` |
| Editar | `PATCH /api/v1/guilds/:guildID/transactions/:id` |
| Excluir | `DELETE /api/v1/guilds/:guildID/transactions/:id` |
| Alterar visibilidade | `PATCH /api/v1/guilds/:guildID/transactions/:id/visibility` |
| Categorizar em lote | `POST /api/v1/guilds/:guildID/transactions:bulk-categorize` |

### Toggle de visão (client)
| Modo UI | Critério |
|---------|----------|
| Visão da guild | `visibility === 'PUBLIC'` |
| Só eu | transações do usuário logado (PRIVATE + PUBLIC próprias) |

### Formulário create/edit
- `description`, `amount`, `type` (`EXPENSE`\|`INCOME`), `date`
- `category` (enum completo)
- `visibility` (`PRIVATE`\|`PUBLIC`) default sugerido: `PUBLIC` para gastos domésticos comuns; `PRIVATE` para pessoais

### Bulk categorize
```json
{ "ids": ["..."], "category": "GROCERY" }
```

---

## 6. Importações

| UI (passo) | Endpoint |
|------------|----------|
| Enviar arquivo | `POST /api/v1/guilds/:guildID/imports` (`multipart`) |
| Acompanhar status | `GET /api/v1/guilds/:guildID/imports/:importID` |
| Editar item na revisão | `PATCH /api/v1/guilds/:guildID/imports/:importID/items/:itemID` |
| Remover item | `DELETE /api/v1/guilds/:guildID/imports/:importID/items/:itemID` |
| Confirmar importação | `POST /api/v1/guilds/:guildID/imports/:importID:confirm` |

### Wizard
```
/imports/new          → upload → redirect /imports/:id
/imports/:id          → poll GET até COMPLETED/FAILED
                      → tabela de items (patch/delete)
                      → botão Confirmar → POST :confirm
                      → /transactions
```

Não há `GET` de listagem de imports na API informada → protótipo pode guardar `lastImportId` em sessionStorage ou só deep-link após upload.

---

## 7. Categories (UI only)

| UI | Endpoint |
|----|----------|
| Select de categoria | **Nenhum** — enum local `Category` |
| Labels PT/EN | i18n client |

Usado em: fixed expenses, transactions, bulk categorize, import item edit.

---

## 8. Matriz de erros (UX)

| HTTP | Comportamento UI |
|------|------------------|
| 401 | Logout / login |
| 403 | Toast “sem permissão” |
| 404 | Empty ou “não encontrado” |
| 409/422 | Mensagem de validação no form |
| 5xx | Toast + retry |
| Network | Banner offline / retry |

---

## 9. Checklist de integração por tela

### Guilds
- [ ] GET list
- [ ] POST create
- [ ] PUT rename
- [ ] PATCH enable/disable
- [ ] POST invite
- [ ] DELETE member

### Fixed expenses
- [ ] GET list
- [ ] POST create
- [ ] PATCH update
- [ ] PATCH deactivate

### Transactions
- [ ] GET list (+ filtros client)
- [ ] POST create
- [ ] PATCH update
- [ ] DELETE
- [ ] PATCH visibility
- [ ] POST bulk-categorize

### Imports
- [ ] POST upload
- [ ] GET by id (poll)
- [ ] PATCH item
- [ ] DELETE item
- [ ] POST confirm

---

## 10. Fluxos ponta a ponta (smoke)

1. **G1** Login → POST guild → GET guilds → dashboard
2. **G2** POST invite
3. **G3** CRUD fixed-expenses
4. **G4** POST transaction PRIVATE/PUBLIC + PATCH visibility
5. **G5** import upload → get → patch item → confirm → GET transactions
6. **G6** trocar activeGuildId → GET transactions da outra guild
7. **G7** bulk-categorize

Mocks devem cobrir os 7 fluxos offline (`VITE_USE_MOCKS=true`).
