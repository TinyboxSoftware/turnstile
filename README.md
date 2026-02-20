# railway-auth-proxy

Minimal OAuth auth proxy for Railway-protected backends.

## Project structure

- `cmd/server/main.go` — application entrypoint, HTTP routes, middleware wiring.
- `internal/config` — environment config loading and validation.
- `internal/auth` — auth middleware and request session context helpers.
- `internal/oauth` — OAuth login/callback/logout handlers.
- `internal/session` — encrypted cookie session management.
- `internal/proxy` — reverse proxy to backend service.
- `internal/railway` — Railway API client (user info + workspace checks).
- `internal/httpx` — shared HTTP helpers (logging, HTTPS detection, JSON errors).

## Build

```bash
go build ./...
```

## Start the service

```bash
go run ./cmd/server
```

The service listens on `PORT` (default: `8080`).
