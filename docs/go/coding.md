# Codebase Guidelines

This is the baseline engineering standard for Go services/APIs.

## 0) Non-negotiables

Code MUST:
- compile and tests MUST pass
- keep architecture boundaries intact (see §1)
- avoid global state (except bootstrap wiring)
- use context for request scope + cancellation
- return consistent error responses
- avoid leaking infrastructure models (DB/transport) into domain

## 1) Architecture boundaries (MANDATORY)

Layers (conceptual):
- Transport: HTTP handlers
- Application: services/usecases (orchestrate domain + ports)
- Domain: entities/value objects/business rules
- Infrastructure: DB clients, HTTP clients, queues, external SDKs

Dependency rules:
- Transport → Application → Domain
- Infrastructure implements ports used by Application
- Domain MUST NOT import infrastructure/transport packages
- Transport MUST NOT call repositories directly

Forbidden:
- handler calling repository directly
- repository depending on handler/service
- domain / entity types carrying `json:` or `db:` tags

## 2) Project structure (feature/domain-first)

Prefer domain/feature packages over technical “layers-only” folders.
Keep related things together.

Example (illustrative):
```go
// Example project structure
project/
├── cmd/
│ └── api/
│ └── main.go
├── internal/
│ ├── user/
│ │ ├── handler.go # HTTP transport
│ │ ├── dto.go # request/response DTOs (transport only)
│ │ ├── service.go # usecases/application logic
│ │ ├── domain.go # domain types (no tags)
│ │ ├── repository.go # ports (interfaces) used by service
│ │ └── storage_postgres.go # adapter implementation
│ ├── platform/
│ │ ├── httpmiddleware/
│ │ ├── logging/
│ │ └── persistence/
│ └── app/
│ └── wiring.go # constructors/composition root
├── migrations/
└── go.mod
```

Notes:
- `internal/` is preferred for application code.
- `pkg/` SHOULD be used only for truly reusable libraries across repos.

## 3) Dependency injection (explicit construction)

Rules:
- dependencies MUST be explicit via constructors
- avoid service locators/singletons
- define interfaces where consumed (consumer-side interfaces)

```go
// user/repository.go
type Repository interface {
    FindByID(ctx context.Context, id string) (*User, error)
    // Other methods...
}

type postgresRepository struct {
    db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
    return &postgresRepository{db: db}
}

// user/service.go
type Service struct {
    repo Repository
    logger *log.Logger
}

func NewService(repo Repository, logger *log.Logger) *Service {
    return &Service{
        repo: repo,
        logger: logger,
    }
}
```

## 4) Error Handling (idiomatic + consistent)

Rules:
- MUST wrap errors with context: fmt.Errorf("action: %w", err)
- MUST NOT compare errors by string
- use sentinel/typed errors for domain cases
- transport MUST map errors to HTTP consistently (middleware)

```go
if err != nil {
return nil, fmt.Errorf("find user by id: %w", err)
}
```

### 4.1 API error model (recommended)

Use a single JSON envelope for errors.
Example type:

```
type AppError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Status  int    `json:"-"` // HTTP status code
	Cause   error  `json:"-"` // optional wrapped cause
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Cause }

// Factory helpers:
func NotFound(msg string, cause error) *AppError {
	return &AppError{Type: "not_found", Message: msg, Status: http.StatusNotFound, Cause: cause}
}
```

### 4.2 Error middleware (sketch)

```go
func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		var appErr *AppError
		if errors.As(err, &appErr) {
			c.JSON(appErr.Status, gin.H{"error": gin.H{
				"type":    appErr.Type,
				"message": appErr.Message,
			}})
			return
		}

		logger.Error("unhandled error", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"type":    "internal_error",
			"message": "an unexpected error occurred",
		}})
	}
}
```

### 5) HTTP middleware order (security + operability)

Recommended order:

1. Recovery
2. Request ID
3. Logging (sanitize sensitive data)
4. Security headers
5. CORS
6. Rate limiting (if any)
7. Error handler
8. Routes

Security headers MUST be reviewed for your deployment context.

### 6) Input validation and normalization
Rules:
- handlers MUST validate input before calling application services
- DTOs belong to transport layer
- normalization MAY be applied (trim spaces, lowercasing emails, etc.)
- XSS sanitization SHOULD NOT be applied blindly for JSON APIs; encode at the rendering boundary if you ever render HTML

Example DTO:
```go
type CreateUserRequest struct {
    Username string `json:"username" binding:"required,alphanum,min=3,max=30"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=12"`
}
```

### 7) Authentication and authorization
Rules:
- prefer OIDC/OAuth2 (do not reinvent identity)
- if passwords exist, MUST hash with a strong algorithm (bcrypt/argon2) and enforce policies
- authorization MUST be checked at every secured endpoint

### 8) Persistence (database) guidelines

Rules:
- SQL MUST stay in repositories/adapters (no SQL in services/handlers)
- use context-aware calls with timeouts
- prefer explicit SQL
- keep DB entities separate from domain models when needed

If using `sqlx`, keep it consistent end-to-end.

Example entity vs domain:
```go
// user/entity.go (DB model)
type userEntity struct {
    ID        string    `db:"id"`
    Username  string    `db:"username"`
    Email     string    `db:"email"`
    CreatedAt time.Time `db:"created_at"`
}

// user/model.go (domain model)
type User struct {
    ID       string
    Username string
    Email    string
    CreatedAt time.Time
}

// user/repository.go
type UserRepository interface {
    FindByID(ctx context.Context, id string) (*User, error)
}

type userRepoPostgres struct {
    db *sqlx.DB
}

func NewUserRepoPostgres(db *sqlx.DB) *userRepoPostgres {
    return &userRepoPostgres{db: db}
}

func (r *userRepoPostgres) FindByID(ctx context.Context, id string) (*User, error) {
    const q = `
		SELECT id, username, email, created_at
		FROM users
		WHERE id = \$1
	`

    var e userEntity
    if err := r.db.GetContext(ctx, &e, q, id); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, NotFound("user not found", err)
        }
        return nil, fmt.Errorf("query user by id: %w", err)
    }

    return &User{
        ID:        e.ID,
        Username:  e.Username,
        Email:     e.Email,
        CreatedAt: e.CreatedAt,
	}, nil
}
```

Connection management:
- MUST configure pool settings
- MUST validate connection on startup (with timeout)

### 9) Structured Logging

Standardize on `slog`.

Rules:
- logger SHOULD be injected (`*slog.Logger`)
- logs MUST be structured (no printf formatting)
- MUST NOT log secrets/credentials/PII

Example:

```go
logger.Info("get user by id", "user_id", id, "request_id", requestID)
```

### 10) Context propagation
Rules:
- handlers MUST derive context from the request
- services/repos MUST accept ctx and propagate it
- MUST set timeouts for IO boundaries (DB/HTTP/queues)

Example:
```go
ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
defer cancel()
```

### 11) Concurrency guidelines
Rules:
- MUST NOT start goroutines without lifecycle control
- MUST avoid goroutine leaks
- use context cancellation
- prefer bounded concurrency (worker pools/semaphores) for fan-out

Bad:
- `go doSomething()`

Better:
- fan-out with semaphore and errgroup (when appropriate)

### 12) Testing strategy
Rules:
- domain: pure unit tests (no IO)
- application/service: unit tests with mocks at boundaries
- transport: integration tests for critical paths
- prefer table-driven tests

Tooling:
- `testing` is default
- `testify` MUST be used for assertions/mocks (avoid over-mocking)

### 13) Configuration & secrets
Rules:
- configuration via environment variables (12-factor)
- secrets MUST NOT be committed
- provide sane defaults for local/dev

### 14) Graceful shutdown
Rules:
- MUST implement graceful shutdown
- MUST close resources (DB, consumers, etc.)
- MUST bound shutdown timeouts

### 15) Anti-patterns (must avoid)
- `utils/helpers/common` dumping grounds
- “god services” with hundreds of lines and too many deps
- cyclic dependencies between packages
- leaking DB entities / transport DTOs into domain
- creating interfaces for every struct “by default”
- global state outside bootstrap

