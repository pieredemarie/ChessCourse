# ChessMaster API server

## Scope and stack
- Work only on the Go HTTP API in this directory. Client UI is outside this repository's scope.
- Go, Gin, PostgreSQL through GORM. SQLite is used by tests.
- `cmd/main.go` wires routes; `internal/handlers` handles HTTP; `internal/service` owns business rules; `internal/storage` owns persistence; `internal/models` and `internal/dto` define stored data and API payloads.
- Keep dependencies one way: handlers -> services -> storage. Put interfaces where they are consumed.

## Requirements
- `ChessMaster_правки_ЛР1-ЛР4 (1).md` contains the final rules for courses and applications. Where laboratory reports conflict, its final rules take precedence.
- The API enforces authorization and business invariants itself.
- Preserve existing JSON fields and routes where possible. Explain required API contract changes.

## Database and changes
- The current project uses GORM `AutoMigrate` in `internal/storage/migrate.go`. New data changes must preserve existing records and be repeatable.
- Do not edit an existing migration without agreement. Add a new migration when needed.
- Do not add dependencies without explaining why.

## Code rules
- Pass `context.Context` first to methods that access the database.
- Wrap errors with `fmt.Errorf("...: %w", err)`; do not discard errors.
- Store passwords only as bcrypt hashes. Never log passwords or tokens. Read secrets and database settings from environment variables.
- Return API errors as `{"error":"text"}` with appropriate HTTP status codes.
- Use table driven tests for service rules and integration tests for cross record invariants.

## Verification
- Run `go test ./...` and `golangci-lint run` before considering a code task complete. Report unavailable tools or failing checks.
- First state a short plan; after implementation report changes, checks, and remaining risks.
