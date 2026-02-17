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
```

### Frontend (Service name: `Frontend`)

```bash
BACKEND_URL=http://backend.production.railway.internal:8081/
BACKEND_KEY=${{Backend.BACKEND_KEY}}
NEXTAUTH_URL=https://$RAILWAY_PUBLIC_DOMAIN/
NEXTAUTH_SECRET=<generated-secret>
EVE_CLIENT_ID=<eve-client-id>
EVE_CLIENT_SECRET=<eve-client-secret>
```

**Note:** `EVE_CLIENT_ID` and `OAUTH_CLIENT_ID` use the **same** EVE Online application credentials.

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

### 4. Update EVE OAuth Callbacks

Update your EVE Online application at https://developers.eveonline.com/ with:
- `https://<railway-url>/api/auth/callback`

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

---

**Auto-deploys on push to main** • Railway free tier includes PostgreSQL with 7-day backups
