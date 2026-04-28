# Architecture Principles & Boundaries

Purpose: Define the structural rules that ensure maintainability, scalability, and loose coupling.
Audience: Software Engineers and AI Agents.

## 1) Core Philosophy
We follow a variation of **Clean Architecture / Hexagonal Architecture** (Ports & Adapters).
- Business logic is the center of the system.
- External systems (HTTP, DB, Queues) are "details" that adapt to the core.
- Dependency rule: dependencies point inwards (towards the Domain).

## 2) Architectural Layers (MANDATORY)

### 2.1 Domain Layer (Core)
Contains pure business logic, entities, and value objects.
- **Rules:**
    - MUST NOT import any other layer.
    - MUST NOT contain infra-specific tags (e.g., `json:`, `db:`, `xml:`).
    - MUST NOT perform IO operations.
    - MUST be testable with pure unit tests (no mocks needed usually).

### 2.2 Application Layer (Usecases / Services)
Orchestrates domain objects to perform specific actions.
- **Rules:**
    - Logic MUST depend on abstractions (interfaces), not implementations.
    - Defines the "Ports" (Repository or External Service interfaces).
    - Handles transaction boundaries (if applicable) and context propagation.
    - Returns Domain objects or simple Results, never DTOs.

### 2.3 Transport Layer (Primary Adapters)
Handles communication from the outside world (HTTP, CLI, gRPC, Cron).
- **Rules:**
    - Responsible for request parsing, validation, and DTO-to-Domain mapping.
    - MUST NOT contain business rules.
    - MUST call Application Services only.
    - Responsible for mapping errors to the appropriate transport status (e.g., HTTP 404).

### 2.4 Infrastructure Layer (Secondary Adapters)
Implements the interfaces defined by the Application layer.
- **Rules:**
    - Contains all technical details (PostgreSQL, Redis, External APIs, Email).
    - Maps Domain objects to Infrastructure Entities (e.g., DB models).
    - Handles technical failure retries and circuit breaking.

## 3) Boundary Enforcement

### Forbidden Dependencies:
- **Layer Crossing:** A Repository MUST NOT call a Service.
- **Direct Access:** A Handler MUST NOT query the Database directly.
- **DTO Leakage:** Transport-specific DTOs MUST NOT leave the Transport layer.
- **Global State:** Packages MUST NOT rely on mutable global variables (e.g., a global `db` variable).

### Data Flow Pattern:
`Request → DTO → [Transport Mapping] → Domain → [Service Logic] → Port (Interface) → [Infra Implementation] → Entity → Persistence`

## 4) Error Propagation Model

1. **Infrastructure:** Detects technical errors (db timeout, connection lost). Wraps or returns them.
2. **Application:** Interprets errors. If a business rule is violated, returns a **Domain Error**.
3. **Transport:** Catches all errors and translates them into the transport's protocol language (JSON, Status Codes, etc.).

*Rule:* Use `fmt.Errorf("context: %w", err)` to preserve the error chain.

## 5) Dependency Inversion (The Golden Rule)
"High-level modules should not depend on low-level modules. Both should depend on abstractions."

- Interfaces MUST belong to the package that *uses* them (the consumer), not the one that implements them.
- This allows the Infrastructure layer to be swapped without changing the Application layer.

## 6) Concurrency & Lifecycle
- Goroutines MUST NOT be started without a clear lifecycle owner.
- Context cancellation MUST be propagated to all IO-bound operations.
- Graceful shutdown MUST be implemented to ensure in-flight work and resource cleanup (DB, Connections).

## 7) Technical Debt & Anti-patterns
- **God Packages:** Packages with too many responsibilities.
- **Utils/Common:** Dumping grounds for unrelated logic.
- **Circular Dependencies:** Signal of poor boundary definition.
- **Anemic Domain Models:** Entities that are just data bags with no logic.
