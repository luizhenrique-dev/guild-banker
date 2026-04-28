# AGENTS.md

Purpose: This file defines how AI agents must behave when working in this repository.

## About the project

### Overview

**GuildBanker** is a multi-user personal finance management platform designed to give users full visibility and control over their financial life. It allows users to register fixed recurring expenses, import credit card statements and bank extracts, and categorize transactions — either manually, rule-based, or with AI assistance.

The platform is built as a **monorepo** containing a **Go REST API** (backend) and a **React + TypeScript** application (frontend), using **Keycloak** for identity and access management and **PostgreSQL 17+** as the primary database.

---

## 1) Mandatory Reading

Before making ANY change, the agent MUST read and follow:

- [go/architecture.md](../docs/go/architecture.md)
- [go/coding.md](../docs/go/coding.md)
- [go/naming.md](../docs/go/naming.md)

These documents are the **source of truth** for:
- architecture decisions
- coding practices
- naming conventions

If there is any conflict:
1. `architecture.md` takes precedence
2. then `coding.md`
3. then `naming.md`

---

## 2) Non-Negotiable Rules

The agent MUST:

- respect architectural boundaries (no layer violations)
- follow dependency direction (inward toward domain)
- avoid introducing global state
- use context propagation correctly
- keep domain free of infrastructure concerns
- keep transport logic free of business rules

The agent MUST NOT:

- create `utils`, `helpers`, or `common` packages
- introduce cyclic dependencies
- leak DTOs or database models into domain
- create unnecessary abstractions or interfaces
- export identifiers without clear need

---

## 3) Code Generation Rules

When generating code, the agent MUST:

### Core Principles
* Simplicity First: Make every change as simple as possible. Minimal code impact.
* No Laziness: Find root causes. No temporary workarounds. Senior developer standards.
* Minimal Impact: Changes should only touch what's necessary. Avoid introducing bugs.

### Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:

- State assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them—don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.
- Disagree honestly. If the user's approach seems wrong, say so—don't be sycophantic.

### Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

### Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:

- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it—don't delete it.

When your changes create orphans:

- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

### Structure
- follow feature/domain-based organization under `internal/`
- keep related files grouped (handler, service, repository, domain)

### Interfaces
- define interfaces at the **consumer side**
- avoid creating interfaces prematurely

### Errors
- wrap errors using `fmt.Errorf("context: %w", err)`
- use domain-level errors for business rules
- never compare errors by string

### Logging
- use structured logging (`slog`)
- do not log sensitive data

### Concurrency
- do not start goroutines without lifecycle control
- always propagate context

---

## 4) Naming Enforcement

All identifiers MUST follow:

- `.standards/go/naming.md`

Critical rules:
- correct use of exported vs unexported names
- consistent initialisms (`ID`, `HTTP`, `URL`)
- no type encoding in names
- plural naming for collections
- `ctx` as context variable name
- `err` as error variable name

---

## 5) Architecture Enforcement

All changes MUST comply with:

- `.standards/go/architecture.md`

Critical checks:
- handlers → services → domain → repositories
- repositories implement interfaces (never the reverse)
- domain has zero external dependencies

---

## 6) When Modifying Existing Code

The agent MUST:

- preserve public API stability unless explicitly instructed
- avoid large refactors without justification
- prefer incremental improvements
- suggest refactors separately when needed

---

## 7) Pull Request Expectations

Any generated change MUST:

- compile successfully
- pass tests (if present)
- maintain or improve readability
- not increase architectural coupling
- follow all standards referenced above

---

## 8) If Uncertain

If any rule is unclear, the agent MUST:

- default to simplicity
- prefer idiomatic Go solutions
- avoid adding abstractions
- follow existing patterns in the codebase (if consistent)

---

## 9) Summary

This repository enforces:

- Clean Architecture (Ports & Adapters)
- SOLID principles
- Idiomatic Go
- Explicit dependencies
- Low coupling, high cohesion

The agent’s goal is to **produce maintainable, predictable, and scalable code**, not just working code.
