# AGENTS.md

## Project

`win-powerctl` — Windows service that exposes an HTTP endpoint to trigger system shutdown. Go, Windows-only.

## Build & Test

```bash
go build ./cmd/win-powerctl    # build
go test ./...                  # run all tests
go vet ./...                   # static analysis
```

No linter, no CI pipeline, no typecheck beyond `go vet`.

## Architecture

```
cmd/win-powerctl/main.go      — entrypoint, HTTP server, CLI (cobra), Windows service, firewall rules
internal/
  admin/admin.go              — Windows elevation check
  auth/auth.go                — password.txt reader
  logger/logger.go            — console logger
  ratelimit/ratelimit.go      — per-IP sliding window rate limiter
  shutdown/win.go             — Windows ExitWindowsEx syscall wrappers
  timeout/timeout.go          — force-shutdown escalation timer
```

## Key Facts

- **Windows-only**: most files use `//go:build windows` build constraint. `go vet` and `go test` run on Windows only.
- **No TLS, no HTTPS**: server binds `0.0.0.0:10125` over plain HTTP.
- **Auth**: `?auth=` query parameter compared via `crypto/subtle.ConstantTimeCompare`. Password stored in plaintext `password.txt` (intentional for local network use).
- **Rate limiter**: `internal/ratelimit` — sliding window, 5 req/60s per IP. Has `sync.Mutex` for thread safety.
- **Dependencies**: chi (router), go-ole (COM/firewall), cobra (CLI), testify (tests), x/sys (Windows syscalls).
- **`password.txt`**: plaintext, committed to repo by design. Do NOT add to `.gitignore`.

## Conventions

- Package-per-concern under `internal/`.
- Tests use `github.com/stretchr/testify/assert`.
- Logger: `logger.Info/Warn/Error/Fatal(component, msg, kv...)` — structured key-value pairs.
- Error messages to HTTP clients are hardcoded strings (no user input reflected).
