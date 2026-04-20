# AGENTS.md

Guidance for AI Coding Agents when working with code in this repository.

## About the project

### 1. Overview

**GuildBanker** is a multi-user personal finance management platform designed to give users full visibility and control over their financial life. It allows users to register fixed recurring expenses, import credit card statements and bank extracts, and categorize transactions — either manually, rule-based, or with AI assistance.

The platform is built as a **monorepo** containing a **Go REST API** (backend) and a **React + TypeScript** application (frontend), using **Keycloak** for identity and access management and **PostgreSQL 17+** as the primary database.

---

### 2. Purpose

Provide a centralized, intelligent tool for individuals to:

- Track and manage **fixed monthly expenses** (rent, car payments, utilities, etc.)
- **Import and parse** credit card statements (CSV and PDF formats)
- **Import bank extracts** to track PIX and other transactions (future phase)
- **Categorize transactions** using bank-provided categories, custom user-defined categories, or AI-powered suggestions
- Gain **financial insights** through dashboards and reports

---

### 3. Goals

| Goal | Description |
|------|-------------|
| **Financial Visibility** | Consolidate all expenses in a single platform |
| **Smart Categorization** | Leverage bank categories, custom rules, and AI to auto-categorize transactions |
| **Ease of Import** | Support CSV parsing (algorithm-based) and PDF parsing (LLM-powered) |
| **Multi-User Support** | Each user has isolated data, managed via Keycloak |
| **Extensibility** | Architecture prepared for future features (bank extract import, PIX tracking, multi-currency) |
| **Code Quality** | Clean Architecture, SOLID principles, and Go community best practices |

---

### 4. Tech Stack

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

### 5. Monorepo Structure

```
guildbanker/
├── api/                          # Go backend
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   └── infra/                # Database, Keycloak client, LLM client, config
│   ├── pkg/                      # Shared utilities (logger, errors, pagination)
│   ├── migrations/               # SQL migration files
│   ├── docs/                     # API documentation (Swagger/Guidelines)
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

# Coding Guidelines

Behavioral guidelines to reduce common LLM coding mistakes. These principles bias toward caution over speed—for trivial tasks, use judgment.

## Core Principles
* Simplicity First: Make every change as simple as possible. Minimal code impact.
* No Laziness: Find root causes. No temporary workarounds. Senior developer standards.
* Minimal Impact: Changes should only touch what's necessary. Avoid introducing bugs.

### 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:

- State assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them—don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.
- Disagree honestly. If the user's approach seems wrong, say so—don't be sycophantic.

### 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

### 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:

- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it—don't delete it.

When your changes create orphans:

- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

**The test:** Every changed line should trace directly to the user's request.

### 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:

- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:

```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

## Go Codebase Guidelines

- **IMPORTANT:** Before making any changes to Go files, strictly read and follow the instructions bellow:

### Naming and Architecture Conventions
Always follow the naming and architecture standards defined in [GO_NAMING_CONVENTIONS.md](../api/docs/GO_NAMING_CONVENTIONS.md).

### Coding
Always follow the coding guidelines defined in [GO_GUIDELINES.md](../api/docs/GO_GUIDELINES.md).


