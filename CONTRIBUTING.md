# Contributing to Tirith

Thanks for your interest in contributing! This document covers how to build, test, and submit changes.

## Development Setup

### Prerequisites

- Go 1.26+
- Node.js 20+ (for the dashboard frontend)
- Python 3.9+ (for the Python SDK)

### Building

```bash
# Build the CLI binary
go build -o tirith ./cmd/tirith

# Build the dashboard frontend (outputs to internal/dashboard/frontend/)
cd frontend && npm ci && npm run build
cp -r out/* ../internal/dashboard/frontend/
```

### Running locally

```bash
go run ./cmd/tirith start
go run ./cmd/tirith report
go run ./cmd/tirith run -- python your_app.py
```

## Testing

```bash
# Go
go test ./...
go test -race ./...
golangci-lint run

# Python SDK
cd sdks/python && pip install -e . && pytest

# TypeScript SDK
cd sdks/typescript && npm ci && npm test
```

## Pull Requests

1. Fork the repo and create a topic branch off `main`.
2. Make your change. Keep PRs focused — one logical change per PR.
3. Add tests where reasonable. Table-driven tests are preferred for Go.
4. Run `go vet ./...`, `golangci-lint run`, and the test suites before pushing.
5. Write a clear PR description: what changed and why.

## Code Style

- **Go:** standard library first, minimize dependencies. Return errors, don't panic. Wrap errors with context.
- **No CGO** — we use `modernc.org/sqlite` for pure-Go SQLite to keep cross-compilation simple.
- **Structured logging** via `slog`.
- **SDK wrappers** are separate packages in `sdks/` — they don't depend on the Go binary at build time.

## Reporting Issues

Use GitHub Issues. For security vulnerabilities, **do not** open a public issue — see [SECURITY.md](./SECURITY.md).

## Code of Conduct

By participating, you agree to uphold the [Code of Conduct](./CODE_OF_CONDUCT.md).
