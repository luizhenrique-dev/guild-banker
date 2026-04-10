# GuildBanker

> Personal finance control platform with smart categorization and AI-powered document parsing.

---

## 1. Overview

**GuildBanker** is a multi-user personal finance management platform designed to give users full visibility and control over their financial life. It allows users to register fixed recurring expenses, import credit card statements and bank extracts, and categorize transactions — either manually, rule-based, or with AI assistance.

The platform is built as a **monorepo** containing a **Go REST API** (backend) and a **React + TypeScript** application (frontend), using **Keycloak** for identity and access management and **PostgreSQL 17+** as the primary database.

---

## 2. Purpose

Provide a centralized, intelligent tool for individuals to:

- Track and manage **fixed monthly expenses** (rent, car payments, utilities, etc.)
- **Import and parse** credit card statements (CSV and PDF formats)
- **Import bank extracts** to track PIX and other transactions (future phase)
- **Categorize transactions** using bank-provided categories, custom user-defined categories, or AI-powered suggestions
- Gain **financial insights** through dashboards and reports

---

## 3. Goals

| Goal | Description |
|------|-------------|
| **Financial Visibility** | Consolidate all expenses in a single platform |
| **Smart Categorization** | Leverage bank categories, custom rules, and AI to auto-categorize transactions |
| **Ease of Import** | Support CSV parsing (algorithm-based) and PDF parsing (LLM-powered) |
| **Multi-User Support** | Each user has isolated data, managed via Keycloak |
| **Extensibility** | Architecture prepared for future features (bank extract import, PIX tracking, multi-currency) |
| **Code Quality** | Clean Architecture, SOLID principles, and Go community best practices |

---

## 4. Tech Stack

| Layer | Technology                                               |
|-------|----------------------------------------------------------|
| **Backend** | Go 1.26+                                                 |
| **Frontend** | React + TypeScript                                       |
| **Database** | PostgreSQL 17+                                           |
| **IAM** | Keycloak 26+                                             |
| **AI / LLM** | LLM integration for PDF parsing and smart categorization |
| **Containerization** | Docker + Docker Compose                                  |
| **Monorepo Structure** | Single repository with `api/` and `web/` directories     |

---

## 5. Monorepo Structure

```
guildbanker/
├── api/                          # Go backend
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── domain/               # Entities, value objects, enums
│   │   ├── usecase/              # Application use cases (business rules)
│   │   ├── adapter/
│   │   │   ├── inbound/          # HTTP handlers, DTOs, middleware
│   │   │   └── outbound/         # Repository implementations, external services
│   │   └── infra/                # Database, Keycloak client, LLM client, config
│   ├── pkg/                      # Shared utilities (logger, errors, pagination)
│   ├── migrations/               # SQL migration files
│   ├── docs/                     # API documentation (Swagger/OpenAPI)
│   ├── go.mod
│   └── go.sum
├── web/                          # React frontend
│   ├── src/
│   ├── public/
│   ├── package.json
│   └── tsconfig.json
├── infra/                        # Infrastructure configs
│   ├── docker-compose.yml
│   ├── keycloak/
│   │   └── realm-config.json
│   └── postgres/
│       └── init.sql
├── docs/                         # Project-level documentation
│   └── SCOPE.md
├── .gitignore
├── Makefile
└── README.md
```

---

## 6. Core Concepts (Domain)

### 6.1 Entities

| Entity | Description |
|--------|-------------|
| **User** | Represents an authenticated user (synced from Keycloak) |
| **FixedExpense** | A recurring monthly expense (e.g., rent, utilities, internet) |
| **Transaction** | A single financial transaction (imported or manually created) |
| **Category** | A classification label for transactions (system-default or user-defined) |
| **CategoryRule** | A rule that auto-assigns a category based on transaction description patterns |
| **CreditCardStatement** | Metadata about an imported credit card statement file |
| **BankExtract** | Metadata about an imported bank extract file (future) |
| **ImportBatch** | Groups transactions from a single import operation for traceability |

### 6.2 Value Objects / Enums

| Name | Values / Description |
|------|----------------------|
| **Currency** | `BRL` (extensible for future multi-currency) |
| **TransactionSource** | `MANUAL`, `CSV_IMPORT`, `PDF_IMPORT`, `BANK_EXTRACT` |
| **TransactionType** | `EXPENSE`, `INCOME` |
| **CategoryOrigin** | `SYSTEM`, `BANK`, `USER`, `AI_SUGGESTED` |
| **ImportStatus** | `PENDING`, `PROCESSING`, `COMPLETED`, `FAILED` |
| **FixedExpenseStatus** | `ACTIVE`, `PAUSED`, `CANCELLED` |

---

## 7. Use Cases

### 7.1 Authentication & Authorization

| ID | Use Case | Description |
|----|----------|-------------|
| UC-01 | **Sign Up / Sign In** | User authenticates via Keycloak (OAuth2 / OpenID Connect) |
| UC-02 | **Token Validation** | API validates JWT tokens issued by Keycloak on every request |
| UC-03 | **User Profile Sync** | On first login, user profile is synced from Keycloak to local DB |

### 7.2 Fixed Expenses Management

| ID | Use Case | Description |
|----|----------|-------------|
| UC-10 | **Create Fixed Expense** | Register a new recurring expense (name, amount, due day, category) |
| UC-11 | **List Fixed Expenses** | View all active fixed expenses for the authenticated user |
| UC-12 | **Update Fixed Expense** | Edit amount, due day, category, or status of a fixed expense |
| UC-13 | **Deactivate Fixed Expense** | Mark a fixed expense as paused or cancelled |
| UC-14 | **Generate Monthly Entries** | Auto-generate transaction entries from active fixed expenses each month |

### 7.3 Transaction Management

| ID | Use Case | Description |
|----|----------|-------------|
| UC-20 | **Create Manual Transaction** | Manually register a one-off expense or income |
| UC-21 | **List Transactions** | List transactions with filters (date range, category, source, type) |
| UC-22 | **Update Transaction** | Edit transaction details (description, amount, category, date) |
| UC-23 | **Delete Transaction** | Remove a transaction |
| UC-24 | **Bulk Categorize** | Assign a category to multiple transactions at once |

### 7.4 Credit Card Statement Import

| ID | Use Case | Description |
|----|----------|-------------|
| UC-30 | **Import CSV Statement** | Upload a CSV file; parse and create transactions using built-in algorithm |
| UC-31 | **Import PDF Statement** | Upload a PDF file; extract transactions using LLM-based parsing |
| UC-32 | **Review Imported Transactions** | After import, user reviews and confirms/edits transactions before saving |
| UC-33 | **Map CSV Columns** | Allow user to map CSV columns to transaction fields (flexible parsing) |

### 7.5 Categorization

| ID | Use Case | Description |
|----|----------|-------------|
| UC-40 | **List Categories** | View system-default and user-created categories |
| UC-41 | **Create Custom Category** | User creates a personal category with name, icon, and color |
| UC-42 | **Edit / Delete Category** | Manage user-created categories |
| UC-43 | **Preserve Bank Category** | Store the original category from the bank/CSV as metadata |
| UC-44 | **Create Category Rule** | Define a pattern-matching rule (e.g., "UBER*" → Transportation) |
| UC-45 | **AI-Powered Categorization** | Send uncategorized transactions to LLM for category suggestions |
| UC-46 | **Accept / Reject AI Suggestion** | User reviews and confirms AI-suggested categories |

### 7.6 Bank Extract Import (Future Phase)

| ID | Use Case | Description |
|----|----------|-------------|
| UC-50 | **Import Bank Extract** | Upload bank extract file (CSV/OFX) to track PIX and other movements |
| UC-51 | **Identify PIX Transactions** | Flag and categorize PIX-specific transactions |
| UC-52 | **Reconcile Transactions** | Match bank extract entries with existing credit card transactions |

### 7.7 Dashboard & Reports (Future Phase)

| ID | Use Case | Description |
|----|----------|-------------|
| UC-60 | **Monthly Summary** | View total expenses, income, and balance for a given month |
| UC-61 | **Category Breakdown** | Pie/bar chart showing spending distribution by category |
| UC-62 | **Expense Trend** | Line chart showing expense evolution over months |
| UC-63 | **Fixed vs Variable** | Compare fixed expenses against variable (imported) expenses |

---

## 8. Non-Functional Requirements

| Requirement | Description |
|-------------|-------------|
| **Security** | All endpoints protected via Keycloak JWT; data isolation per user |
| **Performance** | Pagination on all list endpoints; async processing for imports |
| **Observability** | Structured logging (slog); health check endpoint |
| **Testing** | Unit tests for use cases; integration tests for repositories |
| **Documentation** | OpenAPI/Swagger spec auto-generated from code |
| **CI/CD Ready** | Makefile with targets for build, test, lint, migrate, run |
| **Data Privacy** | Users can only access their own data; soft-delete where applicable |

---

## 9. Phased Delivery Roadmap

### Phase 1 — Foundation
- [ ] Project scaffolding (monorepo, Docker Compose, Makefile)
- [ ] Keycloak integration (authentication + token validation middleware)
- [ ] User profile sync
- [ ] Fixed expenses CRUD
- [ ] Manual transaction CRUD
- [ ] System-default categories + custom categories CRUD

### Phase 2 — Import & Categorization
- [ ] CSV credit card statement import (built-in parser)
- [ ] PDF credit card statement import (LLM-powered)
- [ ] Import review flow (confirm/edit before persisting)
- [ ] Bank category preservation
- [ ] Category rules engine (pattern matching)
- [ ] AI-powered categorization suggestions

### Phase 3 — Bank Extract & Insights
- [ ] Bank extract import (CSV/OFX)
- [ ] PIX transaction identification
- [ ] Transaction reconciliation
- [ ] Monthly summary dashboard
- [ ] Category breakdown charts
- [ ] Expense trend analysis

### Phase 4 — Polish & Scale
- [ ] Notification system (upcoming due dates, budget alerts)
- [ ] Multi-currency support
- [ ] Data export (CSV, PDF reports)
- [ ] Mobile-responsive frontend optimization
- [ ] Rate limiting and advanced security hardening
