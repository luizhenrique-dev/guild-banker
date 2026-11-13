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

## Run the local environment

Install the system prerequisites: Docker with the Compose plugin, GNU Make,
Bash 4.3+, Git, curl, tar, gzip, awk, grep and Python 3. Start Docker and verify
that `docker info` works for your user. Compose is invoked as `docker compose`;
the standalone `docker-compose` command is insufficient. On macOS, install a
recent Bash and ensure `bash` on PATH points to it. Then run:

```sh
make setup
make dev
```

`make setup/tools` first checks these system tools, Python 3, the Compose plugin
and access to the Docker daemon. It logs available prerequisites and reports
missing or unusable ones before downloading toolchains. System prerequisites
are installed through your operating system, not by this target.

It then installs asdf if missing (Linux/macOS on x86_64 or ARM64), adds the
Node.js and Go plugins and installs the versions in `.tool-versions`. npm is
included with Node.js. Existing versions are reused with an `OK` log.

`make setup` runs `setup/tools` first, then
creates missing `api/config.env` and `infra/.env` from their examples, starts
Postgres/Keycloak with realm import and installs API/UI dependencies. At the end,
it shuts down the containers with `infra/down`, preserving volumes and the
imported realm. Cleanup also runs if service startup or dependency installation
fails, or setup is interrupted. Run `make dev` to keep the environment running.
Existing environment files are preserved. Review `infra/.env` to choose your local test
user password before the first import: copy `infra/.env.example` to `infra/.env`
and edit it before running `make setup`. See the
[Keycloak instructions](infra/keycloak/README.md) for details. The application
database name, user and local password are defined in `api/config.env.example`
and `infra/postgres/init/01-init-multiple-dbs.sql`; change `api/config.env` if
your existing database uses different credentials. The initialization SQL only
runs for new volumes; setup does not reset existing databases or passwords.

`make setup/tools` does not start containers or install API/UI dependencies. Make includes
the asdf binary and shims in its PATH, so opening a new terminal is unnecessary.
For direct use of `node`, `npm` and `go` outside Make, add the following to your
shell configuration, following the [asdf setup guide](https://asdf-vm.com/guide/getting-started.html):

```sh
export PATH="$HOME/.local/bin:${ASDF_DATA_DIR:-$HOME/.asdf}/bin:${ASDF_DATA_DIR:-$HOME/.asdf}/shims:$PATH"
```

`make dev` starts Postgres and Keycloak with realm import, waits for their health
checks, builds the API, installs UI dependencies and runs both applications.
UI: `http://localhost:5173`, API: `http://localhost:8081`,
Keycloak: `http://localhost:8080`.
Ctrl+C stops the API and UI; `make infra/down` stops the containers and preserves
their data. If either application exits, the other is stopped as well.

## Keycloak local setup

Realm `guild-banker`, clients `guild-banker-ui`/`guild-banker-api` and the Foo Bar
test user can be prepared and imported with `make keycloak/up`.
See [the setup instructions](infra/keycloak/README.md) for the environment,
initial import, manual client import and API user mapping.

## API – Golang

### Persisted email addresses

Invitation fields validate simple email addresses while typing and convert input to
lowercase. The API applies the same rule to invitations and account writes: quoted
addresses, display names and internal spaces are rejected. Supported addresses are
trimmed and lowercased before persistence.

Before serving requests, a single atomic SQL update lowercases existing account
emails and trims outer spaces, updating only changed addresses. It has a 30-second
timeout. If two accounts collide, startup stops without partial changes; accounts
are never merged automatically. Existing unsupported addresses are neither
reinterpreted nor deleted, and cannot be used in new invitations or authentication.

### Architecture
See [go/architecture.md](./docs/go/architecture.md) for more details.

### Naming Conventions
Always follow the naming defined in [go/naming.md](./docs/go/naming.md).

### Codebase Guidelines
Always follow the coding guidelines defined in [go/coding.md](./docs/go/coding.md).


## 🤖 AI Agents

This repository uses strict engineering standards.

See:
- [Junie](.junie/AGENTS.md)
- [Codex](.codex/AGENTS.md)
