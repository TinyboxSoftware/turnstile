# Turnstile

Turnstile is a reverse proxy that sits in front of your Railway web services and authenticates requests via Railway's OAuth flow. Think Vercel or Netlify password protection, but gated by your Railway workspace's access controls.

## Usecases

- Protect internal tooling like pgAdmin or Metabase
- Ship internal demos to your team without worrying about building auth
- Hide staging / development environments from the public and retain the access control rules already in-place in your Railway workspace

## Getting Started

### 1. Create a Railway OAuth App

1. Go to [Railway's Developer Settings](https://railway.com/workspace/developers) and create a new OAuth application
2. Note the Client ID and Client Secret
3. You'll add the redirect URL in a later step once you have your Turnstile domain

### 2. Setting up Turnstile

1. Click the Deploy on Railway (coming soom :tm:) button above to create a new service in your project
2. Either generate or add a custom public domain to the service
3. Add all your environment variables to the service [as per this table](#environment-variables)
4. Go back to your OAuth application settings and add the redirect URL: `https://<your-turnstile-domain>/_turnstile/auth/callback`
5. Deploy the service

Your `TURNSTILE_PUBLIC_URL` should basically always just be set to `https://${{RAILWAY_PUBLIC_DOMAIN}}`.

> [!TIP]
> Make sure that your `TURNSTILE_BACKEND_URL` includes the proto, domain, and port. The environment variable template should look something like `http://${{super-secret.RAILWAY_PRIVATE_DOMAIN}}:${{super-secret.PORT}}`.

### 3. Test it out

1. Go to your Turnstile service's public domainn. you'll be dropped into  Railway's OAuth flow
2. In the consent screen, select the project you want to authenticate against. If you select the wrong one, you'll see an error page, but you can re-authenticate from there if needed.
3. You should then be forwarded to the service behind the Turnstile proxy

## Environment variables

| Variable                     | Required | Description                                                                                                   |
|------------------------------|----------|---------------------------------------------------------------------------------------------------------------|
| `RAILWAY_CLIENT_ID`          | Yes      | OAuth app client ID from [Railway Developer Settings](https://railway.com/workspace/developers).              |
| `RAILWAY_CLIENT_SECRET`      | Yes      | OAuth app client secret, generated alongside the client ID.                                                   |
| `RAILWAY_PROJECT_ID`         | Yes      | The project to gate access to. You must grant the OAuth app access to this project during consent.            |
| `TURNSTILE_BACKEND_URL`      | Yes      | Internal URL of the service to proxy to. See the [backend URL tips](#configuring-the-backend-url) below.      |
| `TURNSTILE_PUBLIC_URL`       | Yes      | The public URL Turnstile itself is served from. should be set to `https://${{RAILWAY_PUBLIC_DOMAIN}}`.        |
| `TURNSTILE_AUTH_PREFIX`      | No       | Prefix for all Turnstile service routes (auth, health, etc). Defaults to `/_turnstile`.                       |
| `PORT`                       | No       | Port to listen on. Railway sets this automatically; defaults to `8080`.                                       |

### Configuring the Backend URL

Your `TURNSTILE_BACKEND_URL` needs to reference the HTTP protocol, private domain, and port of the service you're protecting.
Using Railway's variable references, the template should look something like:

```text
http://${{my-service.RAILWAY_PRIVATE_DOMAIN}}:${{my-service.PORT}}
```

## Gotchas

### No Support for Legacy Private Networking (IPv6-Only Networks)

Turnstile supports the "current" IPv4 and IPv6 dual-stack networking Railway uses for it's private networking.
Railway has indicated that all projects will eventually be migrated to dual-stack, but until then, this won't work for you.

If you look at my amatuer proxy handler code and see the issue, feel free to open a PR- I'm not afraid to admit I couldn't get it to work ha.

### WebSockets and SSE

Turnstile proxies WebSocket upgrades and SSE connections without issue in my testing.

Tools like pgAdmin, Metabase, or any service using long-lived connections should work as expected behind the proxy.
