# Turnstile

Deployment protection for Railway web services: like Vercel or Netlify password protection, but authenticated via your Railway account and it's access.

You can drop this in front of any WIP app, staging environment, or PR preview to gate access to your Railway workspace members with viewer access on your project. This saves you from building auth code into your app that you'll have to rip out later, or just using HTTP basic auth.

Just set this up as sidecar proxy, point it at the Railway internal networking URL of the service you want to protect, then give the Turnstile service a domain, and remove the domain from your proxied service.

## Environment variables

| Variable                     | Required | Description                                                                                                   |
|------------------------------|----------|---------------------------------------------------------------------------------------------------------------|
| `RAILWAY_CLIENT_ID`          | Y        | OAuth app client ID. [Create an OAuth app](https://railway.com/workspace/developers) in your Railway account. |
| `RAILWAY_CLIENT_SECRET`      | Y        | OAuth app client secret. Generated alongside the client ID when creating your OAuth app.                      |
| `RAILWAY_WORKSPACE_ID`       | Y        | The workspace whose members are allowed through. Grab with `ctrl+k` and "copy workspace id" in Railway        |
| `BACKEND_URL`                | Y        | Internal URL of the service to proxy to: i.e. `http://my-service.railway.internal:3000`.                      |
| `PUBLIC_URL`                 | Y        | The public URL Turnstile itself is served from, e.g. `https://my-service.example.com`.                        |
| `SESSION_SECRET`             | Y        | Random string (min 32 chars) used to encrypt session cookies. Generate one with `openssl rand -base64 32`.    |
| `PORT`                       | —        | Port to listen on. Railway sets this automatically; defaults to `8080`.                                       |

## Project structure

- `cmd/server/main.go` application entrypoint, HTTP routes, middleware wiring.
- `internal/config` environment config loading and validation.
- `internal/auth` auth middleware and request session context helpers.
- `internal/oauth` OAuth login/callback/logout handlers.
- `internal/session` encrypted cookie session management.
- `internal/proxy` reverse proxy to backend service.
- `internal/railway` Railway API client (user info + workspace checks).
- `internal/httpx` shared HTTP helpers (logging, HTTPS detection, JSON errors).

## Build

```bash
go build ./...
```

## Start the service

```bash
go run ./cmd/server
```

The service listens on `PORT` (default: `8080`).
