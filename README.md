# Turnstile

Deployment protection for Railway web services: like Vercel or Netlify password protection, but authenticated via your Railway account's access.

You can drop this in front of any WIP app, staging environment, or PR preview to gate access to other developers who have access to the project in Railway. This saves you from building auth code into your app that you'll have to rip out later, or just using HTTP basic auth.

Just set this up as sidecar proxy, point it at the Railway internal networking URL of the service you want to protect, then give the Turnstile service a domain, and remove the domain from your newly proxied service.

## Environment variables

| Variable                     | Required | Description                                                                                                   |
|------------------------------|----------|---------------------------------------------------------------------------------------------------------------|
| `RAILWAY_CLIENT_ID`          | Yes      | OAuth app client ID. [Create an OAuth app](https://railway.com/workspace/developers) in your Railway account. |
| `RAILWAY_CLIENT_SECRET`      | Yes      | OAuth app client secret. Generated alongside the client ID when creating your OAuth app.                      |
| `RAILWAY_PROJECT_ID`         | Yes      | The project to gate access to. You must grant the OAuth app access to this project during consent.            |
| `TURNSTILE_BACKEND_URL`      | Yes      | Internal URL of the service to proxy to: i.e. `http://my-service.railway.internal:3000`.                      |
| `TURNSTILE_PUBLIC_URL`       | Yes      | The public URL Turnstile itself is served from, e.g. `https://my-service.example.com`.                        |
| `TURNSTILE_AUTH_PREFIX`      | No       | Default to `/_turnstile/`: the prefix under which all turnstile service routes run (Auth, Health, etc)        |
| `PORT`                       | No       | Port to listen on. Railway sets this automatically; defaults to `8080`.                                       |

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
