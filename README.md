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


## 5. Business Logic and Roadmap
The project use case and roadmap are defined in [PROJECT.md](./PROJECT.md) (work in progress). 


# Repository Instructions

## API – Golang

### Naming and Architecture Conventions
Always follow the naming and architecture standards defined in [GO_NAMING_CONVENTIONS.md](./api/docs/GO_NAMING_CONVENTIONS.md).

### Codebase Guidelines
Always follow the coding guidelines defined in [GO_GUIDELINES.md](./api/docs/GO_GUIDELINES.md).
