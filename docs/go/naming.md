# Go Naming Conventions (Agent-Enforceable)

Purpose: provide *deterministic* naming rules for Go identifiers, packages, and files.
Audience: humans and AI coding agents working in this repository.

## 0) Enforcement Scope

Agents MUST apply this guide when they:
- create/rename Go identifiers (vars, consts, types, funcs, methods, receivers, params, struct fields)
- create/rename packages or files
- propose refactors that change exported/public API surface

### Conflicts
If there is a conflict:
1. Prefer **established local conventions** only if they are consistent and intentional across the repo.
2. Otherwise follow this document.
3. If renaming is too disruptive, keep exported names stable and propose a migration plan (minimal blast radius).

### Output acceptance criteria
Changes MUST:
- compile and tests MUST pass
- keep exported surface intentional and minimal
- avoid shadowing builtins and commonly imported packages
- keep package boundaries clear (no dumping grounds like `utils/helpers/common`)

## 1) Visibility is a naming decision (Exported vs Unexported)

- `PascalCase` identifiers are **exported**
- `camelCase` identifiers are **unexported**

Rules:
- Default to **unexported**.
- Export only when required by cross-package use or stable API contracts.

## 2) Casing, initialisms, and acronyms

Rules:
- Use `camelCase` / `PascalCase` only (no `snake_case`, no `SCREAMING_SNAKE_CASE`).
- Initialisms MUST be consistently cased:
    - Good: `apiKey`, `APIKey`, `parseURL`, `HTTPClient`, `userID`
    - Bad: `ApiKey`, `HttpClient`, `userId`

## 3) Name length MUST match scope

Heuristic:
> The farther an identifier is used from its declaration, the more descriptive it must be.

- Small local scopes: short names are OK (`i`, `n`, `s`).
- Wider scopes / exported API: names MUST be descriptive.

Avoid verbosity when context already carries meaning.

## 4) Avoid collisions and shadowing

Identifiers MUST NOT shadow:
- builtins (`len`, `clear`, `any`, etc.)
- imported packages used in the file (`time`, `url`, `json`, `log`, `regexp`, etc.)

If unavoidable, rename the local identifier (not the import).

## 5) Do not encode types in names (with rare exception)

Avoid:
- `scoreInt`, `resultSlice`, `fullNameString`

Allowed:
- when disambiguating an original value vs a converted value:
    - `userID` → `userIDStr`

## 6) Package naming

Rules:
- package names MUST be lowercase ASCII
- prefer short, single-word nouns (`orders`, `customer`, `slug`)
- avoid names that clash with stdlib packages

Avoid “catch-all” packages:
- `util`, `utils`, `helpers`, `common`, `types`, `interfaces`

Prefer small, focused packages:
- `validation`, `formatting`, `mailer`, `links`, `clock`, `uuid`

## 7) Avoid “API chatter” at call sites

Exported identifiers SHOULD NOT repeat the package name.

Example (`customer` package):
- Bad: `customer.NewCustomer()`, `customer.CustomerOrders()`
- Good: `customer.New()`, `customer.Orders()`

Exceptions exist (e.g., `time.Time`, `context.Context`).

## 8) Method receivers

Rules:
- receivers MUST be short (1–3 chars) and consistent across the type’s methods
- MUST NOT use `this`, `self`, `me`

Example:
```
go
type Order struct{ Items int }

func (o *Order) Validate() bool { return o.Items > 0 }
```

## 9) Getters and setters (idiomatic Go)
Rules:
- getter MUST NOT use Get prefix
- setter SHOULD use SetX prefix

Example:
```go
type Customer struct{ address string }

func (c *Customer) Address() string        { return c.address }
func (c *Customer) SetAddress(addr string) { c.address = addr }
```

## 10) Context naming
Rules:
- context variables MUST be named `ctx`
- `ctx context.Context` MUST be the first parameter
Good:
```go
func (s *Service) CreateUser(ctx context.Context, input CreateUserInput) error
```

Bad:
```go
func (s *Service) CreateUser(c context.Context, input CreateUserInput) error
```

## 11) Error naming
Rules:
- error variables MUST be named `err`
- exported sentinel errors MUST start with `Err`
- error messages MUST be lowercase and MUST NOT end with punctuation

Good:
```go
var errNotFound = errors.New("user not found")
```

Bad:
```go
var errNotFound = errors.New("User not found")
```

## 12) Boolean naming

Rules:
- booleans SHOULD use `is/has/can/should` prefixes

Good: isActive, hasPermission, canRetry

Bad: active, permission, retryFlag

## 13) Collections naming (slices, maps)

Rules:
- slices/maps MUST be plural

Good: users, orders, productsByID

Bad: userList, orderSlice, mapUsers

## 14) Interfaces: naming and placement
Rules:
- interfaces SHOULD be defined where they are consumed, not where they are implemented
- avoid premature interfaces; create them only when needed (multiple impls or test boundary)

Bad: repository interface defined in repository package

Good: repository interface defined near the service/usecase that depends on it

## 15) File naming
Rules:

- filenames SHOULD be lowercase and reflect contents (`server.go`, `cookie.go`)
- if multiword files exist, use ONE consistent convention repo-wide

Go special suffixes:
- *_test.go
- *_linux.go, *_windows.go
- *_amd64.go, *_arm64.go

## 16) Review checklist
- Exported only when needed?
- Initialisms consistent (HTTP, URL, ID)?
- Name length matches scope?
- No shadowing of builtins/imports?
- No type-encoding in names (unless conversion disambiguation)?
- Packages are specific (no utils/helpers/common)?
- No API chatter at call site?
- Receiver short + consistent?
- Getter/setter idiomatic?
- Interfaces created/placed correctly?
