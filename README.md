# Deploy and Host Turnstile on Railway

Turnstile is a reverse proxy that authenticates requests via Railway's OAuth flow. It sits in front of your Railway web services and gates access using your Railway workspace's access controls.

This is similar to Vercel or Netlify password protection, but integrated with Railway.

## About Hosting Turnstile

Turnstile is a Go service that deploys to Railway. When a request comes in, Turnstile intercepts it and redirects unauthenticated users through Railway's OAuth flow. Once authenticated, Turnstile ensures they have `project:viewer` access to your project. If they do, the user is proxied through to your backend service using Railway's private networking.

The service handles session management, WebSocket and SSE upgrades, and all OAuth token exchange automatically.

## Common Use Cases

- Ship internal demos to your team without building authentication
- Protect internal tooling and dashboards that are convenient online but not public
- Auth gate staging/development environments using existing Railway workspace access controls

## Dependencies for Turnstile Hosting

- An existing Railway project with a service to protect
- A Railway OAuth application (created in Railway's developer settings)
- A custom domain or Railway-generated domain for Turnstile

### Setting Up the OAuth App

1. Go to [Railway's Developer Settings](https://railway.com/workspace/developers) and create a new OAuth application
2. Note the Client ID and Client Secret
3. You'll add the redirect URL in a later step once you have your Turnstile domain

### Configuring Turnstile

- Deploy the Turnstile template to your Railway project
- Ensure you've setup all the required environment variables correctly:

| Variable | Required | Description |
|----------|----------|-------------|
| `RAILWAY_CLIENT_ID` | Yes | OAuth app client ID from Railway Developer Settings |
| `RAILWAY_CLIENT_SECRET` | Yes | OAuth app client secret |
| `RAILWAY_PROJECT_ID` | Yes | The project to gate access to |
| `TURNSTILE_BACKEND_URL` | Yes | Internal URL of the service to proxy (e.g., `http://${{my-service.RAILWAY_PRIVATE_DOMAIN}}:${{my-service.PORT}}`) |
| `TURNSTILE_PUBLIC_URL` | Yes | Public URL Turnstile is served from (`https://${{RAILWAY_PUBLIC_DOMAIN}}`) |
| `TURNSTILE_AUTH_PREFIX` | No | Prefix for auth routes (defaults to `/_turnstile`) |
| `PORT` | No | Port to listen on (defaults to `8080`) |

- Add the OAuth redirect URL to your OAuth application registration: `https://<your-turnstile-domain>/_turnstile/auth/callback`
- Redeploy your turnstile service

### Testing

1. Visit Turnstile's public domain
2. Complete Railway's OAuth flow, selecting the project to authenticate against
3. You'll be forwarded to your protected service OR shown an error page if you don't have access.

## Implementation Details

The source code is available on [GitHub](https://github.com/mykal/railway-auth-proxy).

Turnstile is open source: If you're interested in how the OAuth flow, proxying, or session management works, check out the repo.

## Why Deploy Turnstile on Railway?

Railway is a singular platform to deploy your infrastructure stack. Railway will host your infrastructure so you don't have to deal with configuration, while allowing you to vertically and horizontally scale it.

By deploying Turnstile on Railway, you are one step closer to supporting a complete full-stack application with minimal burden. Host your servers, databases, AI agents, and more on Railway.
