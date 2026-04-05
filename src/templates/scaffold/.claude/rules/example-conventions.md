---
description: "Code conventions — scoped to specific directories"
paths:
  - "src/**/*.go"
  - "cmd/**/*.go"
---

## Go Conventions

- Error handling: always wrap errors with `fmt.Errorf("context: %w", err)`
- **Why:** Unwrapped errors lose context in stack traces

- Logging: use structured logging (`slog` or `zerolog`)
- **Why:** Structured logs are searchable; printf-style logs are not

- Testing: table-driven tests in `_test.go` files alongside source
- **Why:** Go convention; keeps tests close to implementation
