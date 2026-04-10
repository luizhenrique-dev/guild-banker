# Go Naming Conventions (AI-Applicable Style Guide)

This document is a **practical naming guide for Go codebases**, intended to be read and applied by AI coding agents (e.g., Junie, Claude Code) and humans.

## 1) Agent Scope (how this becomes enforceable)

### 1.1 Where to put this

Place this file in a location that agents will reliably see:

- **Recommended (repo root):** `GO_NAMING_CONVENTIONS.md`
- Also recommended:
  - Reference it from `README.md` and/or `CONTRIBUTING.md`
  - If you use Claude Code, add a short pointer in `CLAUDE.md` like: “Follow `GO_NAMING_CONVENTIONS.md` for all Go naming decisions.”

### 1.2 When the agent must apply it

An agent working on this repository **must follow these rules** when it:

- creates/renames Go identifiers (vars, consts, types, funcs, methods, receivers, params, struct fields)
- creates/renames packages or files
- proposes refactors that change public API surfaces

### 1.3 What to do if there is a conflict

1. **Prefer existing local conventions** *only if* they are consistent across the repo and clearly intentional.
2. Otherwise, follow this document.
3. If changing names would be too disruptive, the agent should:
   - keep existing exported names stable (API compatibility)
   - improve names gradually in new code
   - suggest a safe migration plan (with minimal blast radius)

### 1.4 Acceptance criteria for agent output

Any PR/change set produced by an agent should satisfy:

- Go code compiles/tests pass.
- Exported vs unexported naming is intentional.
- No “chattery” public APIs.
- No collisions with builtins or imported package names.
- Receivers are short and consistent.
- Package boundaries remain clear (avoid `utils/helpers/common` dumping grounds).

## 2) The rules (apply these first)

## Rule A — Exporting is a naming decision

In Go, **capitalization changes visibility**:

- `PascalCase` = **exported** (visible outside the package)
- `camelCase` = **unexported** (package-private)

**Guideline:** default to unexported. Export only when there is a real need.

Why this matters:
- Less exported surface = easier refactors, smaller “blast radius”, cleaner APIs.

## Rule B — Use standard casing and consistent initialisms

- Use `camelCase`/`PascalCase`.
- Don’t use `snake_case`, `SCREAMING_SNAKE_CASE`, `ALLUPPERCASE`, etc.

**Initialisms/acronyms** should be consistently cased *inside* identifiers:

- Good: `APIKey`, `apiKey`, `parseURL`, `HTTPClient`
- Bad: `ApiKey`, `HttpClient`

**Special note:** `ID` should be all caps:
- Prefer `userID` over `userId`.

## Rule C — Name length should match scope

Heuristic:

> The farther an identifier is used from where it is declared, the more descriptive it should be.

- Tight scopes (small loops/short blocks): short names are OK (`i`, `p`, `s`).
- Wider scopes: prefer descriptive names (`count`, `sum`, `customer`, `orderTotal`).

Avoid unnecessary verbosity: don’t over-describe when context already exists.

## Rule D — Avoid collisions (builtins + imported packages)

Avoid naming identifiers after:

- builtin types: `int`, `bool`, `any`
- builtin funcs: `len`, `clear`, `min`, `max`
- standard library packages *especially those you import in the same file*: `time`, `url`, `log`, `json`, `regexp`, etc.

Why this matters:
- Prevents shadowing and reader confusion.
- Improves navigation/search and reduces subtle bugs.

## Rule E — Don’t encode types in names (usually)

Avoid: `scoreInt`, `resultSlice`, `fullNameString`.

Acceptable exception:
- When you have an original value and its converted form, type suffixes can disambiguate:
  - `userID` → `userIDStr`

## 3) Package naming (high leverage)

### 3.1 Conventions

- Package names: **lowercase ASCII**, short, easy to type.
- Usually a **single word noun** that reflects contents: `orders`, `customer`, `slug`.
- Multi-word packages: concatenate lowercase words with no separator: `ordermanager` (not `order_manager` or `orderManager`).
- Avoid names that clash with commonly-used stdlib packages.

### 3.2 Avoid “catch-all” packages

Avoid: `util`, `utils`, `helpers`, `common`, `types`, `interfaces`.

Why:
- unclear boundaries, becomes a dumping ground
- increases cross-package coupling and risk of import cycles

Preferred approach:
- create smaller, focused packages: `validation`, `formatting`, `mailer`, `links`, etc.

## 4) Avoid “chatter” in public APIs

When naming exported functions/types, don’t repeat the package name.

Example:

- If package is `customer`:
  - Bad: `customer.NewCustomer()`, `customer.CustomerOrders()`
  - Good: `customer.New()`, `customer.Orders()`

Notes:
- Sometimes repetition is unavoidable/acceptable (e.g., `time.Time`, `context.Context`).

## 5) Method receivers

- Receivers should be short (typically 1–3 chars) and often derived from the type.
- Don’t use generic receiver names like `this`, `self`, `me`.
- Be consistent across all methods on the same type.

Example:

```go
type Order struct {
    Items int
}

func (o *Order) Validate() bool {
    return o.Items > 0
}
```

## 6) Getters and setters

Go usually accesses struct fields directly.

When you must expose an unexported field via methods:

- Getter: **no `Get` prefix**
- Setter: use `SetX` prefix

Example:

```go
type Customer struct {
    address string
}

func (c *Customer) Address() string {
    return c.address
}

func (c *Customer) SetAddress(addr string) {
    c.address = addr
}
```

## 7) Interfaces

For single-method interfaces, prefer method name + `-er` (or similar):

- `io.Reader`, `io.Writer`, `fmt.Stringer`
- `Authenticator`, `Authorizer`

Avoid names like `UserInterface`/`OrderInterface` unless there’s no better alternative.

## 8) File naming (and special suffixes)

- Prefer simple, lowercase filenames that reflect contents: `server.go`, `cookie.go`.
- If multiple words are needed, be consistent within the repo (either concatenated or underscore-separated).

Be aware of special behaviors:

- `*_test.go` is for tests.
- OS-specific files: `*_linux.go`, `*_windows.go`, etc.
- Arch-specific files: `*_amd64.go`, `*_arm64.go`, etc.
- Prefix `.` or `_` makes files ignored by Go tooling.

## 9) Practical checklist (for reviews and agent self-check)

Before finalizing a change:

- [ ] Did I export this identifier only if necessary?
- [ ] Are casing and initialisms consistent (`HTTP`, `URL`, `ID`)?
- [ ] Is the name descriptive enough for its scope (but not overly verbose)?
- [ ] Did I avoid collisions with builtins and imported package names?
- [ ] Did I avoid encoding types in names (unless disambiguating conversions)?
- [ ] Are package names short, lowercase, and specific (not `utils/helpers`)?
- [ ] Did I avoid API “chatter” at the call site?
- [ ] Are method receivers short and consistent on the type?
- [ ] Are getters/setters idiomatic (no `GetX`, setters use `SetX`)?
