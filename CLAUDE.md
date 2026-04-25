# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build ./cmd/onboarding/... ./internal/... ./pkg/...

# Run tests
go test ./...

# Run a single test
go test ./internal/onboarding/domain/entity/ -run TestName -v

# Run with race detector
go test -race ./...

# Format and vet
go fmt ./...
go vet ./...
```

There is no Makefile — use standard Go tooling directly.

## Architecture

This is a DDD (Domain-Driven Design) core banking system in Go. The current bounded context is **Onboarding** (client account creation and approval flow).

### Layer Structure

```
cmd/onboarding/main.go               # Entry point — wires dependencies and starts HTTP server + outbox worker
internal/onboarding/
  domain/                            # Core business logic — no external dependencies
    entity/                          # Aggregate root: Client
    event/                           # Domain events: AccountCreated, AccountApproved, AccountRejected
    repository/                      # Repository interfaces (ClientRepository, OutboxRepository)
    errors.go                        # All domain sentinel errors
  application/usecases/              # Use cases: CreateAccount, ApproveAccount, RejectAccount
  infrastructure/
    outbox/                          # Outbox worker — LISTEN/NOTIFY based, dedicated pgx.Conn
    persistence/postgres/            # Postgres implementations of repository interfaces
  interfaces/http/                   # HTTP handlers (package handler, not http)
pkg/
  events/                            # DomainEvent and EventBus interfaces (shared kernel)
  outbox/                            # OutboxRecord struct
  http/                              # Shared ApiResponse struct
  valueobjects/                      # Document (CPF/CNPJ), Email, ID
```

### Key Design Decisions

**Aggregate root (`Client`)**: Holds a private `events []DomainEvent` slice. `PullEvents()` returns and clears it. State transitions (`ApproveAccount`, `RejectAccount`) only fire from `PENDING` status and append a domain event. `CreateAccount()` also appends `AccountCreated` on construction.

**`Document()` returns `valueobjects.Document`** (not `string`). Use `.Number()` for the raw digits and `.Category()` for CPF/CNPJ. `Email` value object exposes `.Value()`.

**Domain errors**: All sentinel errors live in `internal/onboarding/domain/errors.go` (package `domain`). The `entity` package imports `domain` to return these sentinels. Use `errors.Is(err, domain.ErrXxx)` throughout — never compare error strings. Wrapping with `fmt.Errorf("%w", ...)` preserves sentinel identity through the call stack.

**Outbox pattern**: Use cases do not publish events directly. `ClientRepository.Save(client, events)` persists the client and domain events atomically in one DB transaction, then calls `pg_notify('outbox_event', '')`. `infrastructure/outbox.Worker` holds a dedicated `*pgx.Conn` (not pool) and blocks on `WaitForNotification`, then calls `processAll` to publish via `MessagePublisher` and mark records processed. The pool (`pgxpool`) is separate and used only by the HTTP handlers.

**HTTP handler package**: The package at `interfaces/http/` is declared as `package handler` (not `package http`) to avoid colliding with `net/http`. Import it with an explicit alias: `handler "ddd-core-banking/internal/onboarding/interfaces/http"`.

**Error → HTTP status mapping** in handlers:
- `ErrEmailAlreadyInUse` → 409
- `ErrInvalidDocument`, `ErrInvalidEmail`, `ErrFullNameRequired`, `ErrPhoneRequired`, `ErrAccountNotPending`, `ErrRejectionReasonRequired` → 422
- `ErrNotFound` → 404
- anything else → 500

**`main.go` wiring**: Two goroutines started — HTTP server and outbox worker — each sending fatal errors to a channel. A `signal.NotifyContext` cancels both on `SIGINT`/`SIGTERM`. On shutdown, `server.Shutdown` drains in-flight requests; the worker exits because `WaitForNotification(ctx)` returns when `ctx` is cancelled.

### Onboarding Flow

```
CreateAccount() → status: PENDING, emits AccountCreated
  ↓
ApproveAccount() → status: APPROVED, emits AccountApproved
RejectAccount(reason) → status: REJECTED, emits AccountRejected
```

Follow the dependency rule: inner layers (domain) must not import outer layers. The `entity` package importing `domain/errors.go` is the one intentional exception — errors are part of the domain and both live inside the domain layer.
