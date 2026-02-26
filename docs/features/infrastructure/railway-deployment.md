# Railway Deployment Guide

## Overview

Deploy three Railway services: **PostgreSQL** (database), **Backend** (Go), **Frontend** (Next.js). Backend is private, frontend proxies requests via Railway's internal DNS.

## Prerequisites

### 1. Generate Secrets

```bash
# BACKEND_KEY
openssl rand -hex 32

# NEXTAUTH_SECRET
openssl rand -base64 32
```

### 2. EVE Online OAuth Application

Create at https://developers.eveonline.com/ with:

**Callback URL** (update `<railway-url>` after deployment):
- `https://<railway-url>/api/auth/callback`

**Required Scopes:** See [frontend/packages/client/auth/api.ts](../../frontend/packages/client/auth/api.ts) for complete list. Minimum:
- `esi-assets.read_assets.v1`, `esi-assets.read_corporation_assets.v1`
- `esi-characters.read_corporation_roles.v1`, `esi-corporations.read_divisions.v1`
- `esi-universe.read_structures.v1`, `esi-wallet.*`, `esi-industry.*`, `esi-markets.*`

Save your **Client ID** and **Client Secret**.

### 3. Discord Bot & OAuth Application (Optional)

To enable Discord purchase notifications, create a Discord application at https://discord.com/developers/applications:

1. **Create Application** → Name it (e.g., "Pinky.Tools")
2. **Bot** tab → Reset Token → Copy **Bot Token** (this is `DISCORD_BOT_TOKEN`)
3. **Bot** tab → Enable "Message Content Intent" under Privileged Gateway Intents
4. **OAuth2** tab → Copy **Client ID** and **Client Secret**
5. **OAuth2** tab → Add redirect URL: `https://<railway-url>/api/discord/callback`

**Bot Permissions** (for invite link): `Send Messages` (2048)

**Bot Invite URL:**
```
https://discord.com/api/oauth2/authorize?client_id=<discord-client-id>&permissions=2048&scope=bot
```

Users must invite the bot to their Discord server before they can configure channel notifications. DM notifications work without a server invite.

## Environment Variables

### Backend (Service name: `Backend`)

```bash
PORT=8081
DATABASE_HOST=${{Postgres.PGHOST}}
DATABASE_PORT=${{Postgres.PGPORT}}
DATABASE_USER=${{Postgres.PGUSER}}
DATABASE_PASSWORD=${{Postgres.PGPASSWORD}}
DATABASE_NAME=${{Postgres.PGDATABASE}}
BACKEND_KEY=<generated-secret>
OAUTH_CLIENT_ID=<eve-client-id>
OAUTH_CLIENT_SECRET=<eve-client-secret>

# Discord notifications (optional)
DISCORD_BOT_TOKEN=<discord-bot-token>
```

### Frontend (Service name: `Frontend`)

```bash
BACKEND_URL=http://backend.production.railway.internal:8081/
BACKEND_KEY=${{Backend.BACKEND_KEY}}
NEXTAUTH_URL=https://$RAILWAY_PUBLIC_DOMAIN/
NEXTAUTH_SECRET=<generated-secret>
EVE_CLIENT_ID=<eve-client-id>
EVE_CLIENT_SECRET=<eve-client-secret>

# Discord notifications (optional)
DISCORD_CLIENT_ID=<discord-client-id>
DISCORD_CLIENT_SECRET=<discord-client-secret>
```

**Note:** `EVE_CLIENT_ID` and `OAUTH_CLIENT_ID` use the **same** EVE Online application credentials.

**Note:** Discord variables are optional. If omitted, Discord notifications are disabled and the Settings page won't show Discord linking.

## Deployment Steps

### 1. Create Railway Project

1. Sign up at https://railway.app/
2. Create New Project → Deploy from GitHub repo → Select `industry-tool`
3. Add PostgreSQL: "+ New" → Database → PostgreSQL (name: `Postgres`)

### 2. Deploy Backend

1. "+ New" → GitHub Repo → `industry-tool` (name: `Backend`)
2. Settings → Build:
   - Builder: `Dockerfile`
   - Dockerfile Path: `Dockerfile`
   - Docker Build Args: `--target final-backend`
3. Add environment variables from Backend section above
4. Wait for deployment (logs should show "starting services")

### 3. Deploy Frontend

1. "+ New" → GitHub Repo → `industry-tool` (name: `Frontend`)
2. Settings → Build:
   - Builder: `Dockerfile`
   - Dockerfile Path: `Dockerfile`
   - Docker Build Args: `--target publish-ui`
3. Settings → Networking → Generate Domain (copy URL)
4. Add environment variables (use generated URL for `NEXTAUTH_URL`)
5. Wait for deployment

### 4. Update OAuth Callbacks

Update your EVE Online application at https://developers.eveonline.com/ with:
- `https://<railway-url>/api/auth/callback`

If using Discord, update your Discord application at https://discord.com/developers/applications with:
- `https://<railway-url>/api/discord/callback`

### 5. Verify

- Visit Railway URL → Sign in with EVE → Test "Add Character" button

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Backend can't connect to database | Check service name is `Postgres` (case-sensitive). Verify `DATABASE_*` variables use `${{Postgres.PGHOST}}` syntax |
| Frontend gets 401 from backend | `BACKEND_KEY` must match in both services. Use `${{Backend.BACKEND_KEY}}` in Frontend |
| OAuth callback fails | Check `NEXTAUTH_URL` matches Railway domain exactly (with `https://` and trailing `/`). Verify callback URL in EVE app: `https://<url>/api/auth/callback` |
| Frontend can't reach backend | `BACKEND_URL` must be `http://backend.production.railway.internal:8081/`. Service name must be exactly `Backend` |
| Docker build fails | Verify build args: Backend `--target final-backend`, Frontend `--target publish-ui` |
| Discord linking fails | Check `DISCORD_CLIENT_ID` and `DISCORD_CLIENT_SECRET` are set on Frontend. Verify callback URL in Discord app: `https://<url>/api/discord/callback` |
| Discord notifications not sending | Check `DISCORD_BOT_TOKEN` is set on Backend. Backend logs "discord notifications enabled" on startup if configured correctly |
| Channel notifications fail | Bot must be invited to the Discord server. Use the bot invite URL with `Send Messages` permission |

---

**Auto-deploys on push to main** • Railway free tier includes PostgreSQL with 7-day backups
