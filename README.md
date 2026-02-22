# Turnstile

Deployment protection for Railway web services: like Vercel or Netlify password protection, but authenticated via your Railway account's access.

## Usecases

- Protect internal tooling like pgAdmin or Metabase
- Ship internal demos to your team without worrying about building auth
- Hide staging / development environments from the public and retain the access control rules already in-place in your Railway workspace

## Getting Started

### Before you deploy

1. Create a Railway OAuth app registration for your team to use

### Setting up Turnstile

1. Create a new service in your project / environment for this repo
2. Either generate or add a custom public domain to the service
3. Add the redirect URL to your OAuth application (should be like `https://super-secret.yourdomain.com/_turnstile/auth/callback`)
4. Add all your environment variables the service [as per this table](#environment-variables)
5. Click the big deploy button at the top of the page

Your `TURNSTILE_PUBLIC_URL` should basically always just be set to `https://${{RAILWAY_PUBLIC_DOMAIN}}`.

> [!TIP]
> Make sure that your `TURNSTILE_BACKEND_URL` includes the proto, domain, and port. The environment variable template should look something like `http://${{super-secret.RAILWAY_PRIVATE_DOMAIN}}:${{super-secret.PORT}}`.

### Testing it out

1. Go to your turnstile service's public domain, you'll be sent down Railway's OAuth flow
2. In the consent screen, make sure you select the project you want to login to
3. You should be forwarded back to one of two places:
    1. the service behind the turnstile proxy: success!
    2. an error page saying you don't have access: this happens if you selected the wrong project, or just don't have access

## Environment variables

| Variable                     | Required | Description                                                                                                   |
|------------------------------|----------|---------------------------------------------------------------------------------------------------------------|
| `RAILWAY_CLIENT_ID`          | Yes      | OAuth app client ID. [Create an OAuth app](https://railway.com/workspace/developers) in your Railway account. |
| `RAILWAY_CLIENT_SECRET`      | Yes      | OAuth app client secret. Generated alongside the client ID when creating your OAuth app.                      |
| `RAILWAY_PROJECT_ID`         | Yes      | The project to gate access to. You must grant the OAuth app access to this project during consent.            |
| `TURNSTILE_BACKEND_URL`      | Yes      | Internal URL of the service to proxy to: i.e. `http://my-service.railway.internal:3000`.                      |
| `TURNSTILE_PUBLIC_URL`       | Yes      | The public URL Turnstile itself is served from, e.g. `https://my-service.example.com`.                        |
| `TURNSTILE_AUTH_PREFIX`      | No       | Default to `/_turnstile`: the prefix under which all turnstile service routes run (Auth, Health, etc)        |
| `PORT`                       | No       | Port to listen on. Railway sets this automatically; defaults to `8080`.                                       |

## Gotchas

### No support for "legacy" private networking (IPv6 only networks)

I *tried* to get this working but I couldn't quite get it. For now, this only supports IPv4 and IPv6 dual stack networking.

Railway has said eventually everyone will get migrated to dual stack so hopefully that day is sooner than later.
