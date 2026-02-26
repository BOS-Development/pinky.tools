# Consolidate OAuth2 Endpoints

## Context

The app currently uses **two separate CCP applications** and **two callback URLs** for OAuth:
1. **NextAuth EVEOnlineProvider** (`EVE_CLIENT_ID`) — user login at `/api/auth/callback/eveonline` (no ESI scopes, identity only)
2. **openid-client alt flow** (`ALT_EVE_CLIENT_ID`) — adding characters/corporations at `/api/altAuth/callback` (full ESI scopes)

The deployment docs already note "All OAuth variables use the **same** EVE Online application credentials" — but the code requires two sets of env vars and two registered callback URLs. This consolidation merges them into **one CCP application, one callback URL, and simpler code**.

Additionally, tokens are stored without recording which scopes were granted. Adding a `scopes` column lets the system know what each token can do, and in the future lets users choose subsets of scopes.

---

## Approach

- **Keep NextAuth** for session management — it's used in ~35 API routes (`getServerSession`) and ~20 components (`useSession`). Replacing it would be a huge change with no benefit.
- **Remove `EVEOnlineProvider`** from NextAuth. Login now goes through the same openid-client flow as character/corp addition.
- **Single callback** at `/api/auth/callback` handles all three flow types (login, char, corp) based on state.
- **For login**: after OAuth completes, create the NextAuth session by encoding a JWT with `next-auth/jwt`'s `encode()` and setting the session cookie directly.
- **Lazy discovery**: Move `client.discovery()` from module top-level to a lazy getter so the module loads without network calls (fixes E2E where EVE SSO is unreachable).
- **Store scopes**: Add `esi_scopes TEXT` column to `characters` and `player_corporations` tables. Frontend passes the granted scopes alongside the token; backend stores them.

---

## Step 1: Database migration — add scopes column

**New:** `internal/database/migrations/20250217100000_add_esi_scopes.up.sql`

```sql
ALTER TABLE characters ADD COLUMN esi_scopes TEXT NOT NULL DEFAULT '';
ALTER TABLE player_corporations ADD COLUMN esi_scopes TEXT NOT NULL DEFAULT '';
```

**New:** `internal/database/migrations/20250217100000_add_esi_scopes.down.sql`

```sql
ALTER TABLE characters DROP COLUMN esi_scopes;
ALTER TABLE player_corporations DROP COLUMN esi_scopes;
```

Scopes stored as a space-delimited string (same format EVE SSO returns them). Default `''` for existing rows.

## Step 2: Backend — update models and repositories

**`internal/repositories/character.go`**:
- Add `EsiScopes string` field to `Character` struct
- Update `Add()` INSERT/upsert to include `esi_scopes` column
- Update `GetAll()` SELECT to include `esi_scopes` column + Scan

**`internal/repositories/playerCorporations.go`**:
- Add `EsiScopes string` field to `PlayerCorporation` struct
- Update `Upsert()` INSERT to include `esi_scopes` column
- Update `Get()` SELECT to include `esi_scopes` column + Scan

**`internal/controllers/corporations.go`**:
- Pass `character.EsiScopes` to the `Upsert` call (the scopes from the frontend)

No controller changes needed for characters — the `Add` handler already decodes the full `Character` struct from JSON; adding the field to the struct is sufficient.

## Step 3: Frontend — update API client model

**`frontend/packages/client/api.ts`**:
- Add `esiScopes: string` to `FullCharacterData` type

## Step 4: Update `frontend/packages/client/auth/api.ts`

- Replace `ALT_EVE_CLIENT_ID`/`ALT_EVE_CLIENT_SECRET` with `EVE_CLIENT_ID`/`EVE_CLIENT_SECRET`
- Make `client.discovery()` lazy (only runs on first OAuth call, not at module import)
- Change redirect_uri from `/api/altAuth/callback` to `/api/auth/callback`
- Change `getAuthUrl(isCorp: boolean)` to `getAuthUrl(flowType: "login" | "char" | "corp")`
  - `"login"`: scope = `"publicData"` (identity only)
  - `"char"`: scope = current `playerScope` (full character ESI scopes)
  - `"corp"`: scope = current `corpScope` (full corporation ESI scopes)
- `stateMaps` stores `{ flowType, codeVerifier }` per state (generate a fresh PKCE verifier per request)
- `verifyToken` returns `flowType` instead of `redirectType`
- Export the scope strings so the callback can pass them to the backend

## Step 5: Create `frontend/pages/api/auth/callback.ts` (NEW)

Unified callback handler replacing both NextAuth's EVE callback and `altAuth/callback.ts`.

**Login flow** (no existing session required):
1. Call `verifyToken()` to exchange code for tokens
2. Decode the access_token JWT to extract character ID (`sub` field) and name
3. Create/lookup user in backend (reuse `getUser`/`addUser` from `[...nextauth].ts`)
4. Encode a NextAuth-compatible JWT using `encode()` from `next-auth/jwt` with `{ providerAccountId, name }`
5. Set the session cookie (`next-auth.session-token` for HTTP, `__Secure-next-auth.session-token` for HTTPS)
6. Redirect to `/`

**Char flow** (requires existing session):
1. Verify session via `getServerSession(req, res, authOptions)`
2. Call `verifyToken()` to exchange code for tokens
3. Decode JWT, call backend `addCharacter` with `{ id, name, esiToken, esiRefreshToken, esiTokenExpiresOn, esiScopes: playerScope }`
4. Redirect to `/characters`

**Corp flow** (requires existing session):
1. Verify session via `getServerSession(req, res, authOptions)`
2. Call `verifyToken()` to exchange code for tokens
3. Decode JWT, call backend `addCharacterCorporation` with `{ ..., esiScopes: corpScope }`
4. Redirect to `/corporations`

## Step 6: Create `frontend/pages/api/auth/login.ts` (NEW)

Simple entry point for sign-in:
```typescript
export default async function handler(req, res) {
  const redirectTo = await getAuthUrl("login");
  res.redirect(redirectTo);
}
```

## Step 7: Update `frontend/pages/api/auth/[...nextauth].ts`

- Remove `EVEOnlineProvider` from providers array
- Keep `CredentialsProvider` (used by E2E tests)
- Keep all JWT/session callbacks (still needed for session decoding)
- Keep `getUser`/`addUser` functions (used by both CredentialsProvider and the new callback)
- **Export** `getUser` and `addUser` so the new callback.ts can reuse them

## Step 8: Update character/corporation add endpoints

- **`frontend/pages/api/characters/add.ts`**: Change `getAuthUrl(false)` → `getAuthUrl("char")`
- **`frontend/pages/api/corporations/add.ts`**: Change `getAuthUrl(true)` → `getAuthUrl("corp")`

## Step 9: Update sign-in links

Change `href` from `/api/auth/signin` to `/api/auth/login`:
- `frontend/packages/pages/index.tsx` (line 203)
- `frontend/packages/components/unauthorized.tsx` (line 39)

## Step 10: Delete old files

- Delete `frontend/pages/api/altAuth/callback.ts`

## Step 11: Update environment variables

**`frontend/.env.local`**: Remove `ALT_EVE_CLIENT_ID` and `ALT_EVE_CLIENT_SECRET`

**`docker-compose.e2e.yaml`**: No changes needed — `EVE_CLIENT_ID`/`EVE_CLIENT_SECRET` already set for frontend

**`docs/features/infrastructure/railway-deployment.md`**:
- Remove `ALT_EVE_CLIENT_ID`/`ALT_EVE_CLIENT_SECRET` from Frontend env vars
- Update callback URLs section to show single URL: `https://<railway-url>/api/auth/callback`
- Remove troubleshooting entry about `ALT_EVE_CLIENT_ID`

## Step 12: Update E2E seed data

**`e2e/seed.sql`**: Add `esi_scopes` values to any character/corporation INSERT statements that exist in the seed file.

## Step 13: Update tests

- **`frontend/packages/components/__tests__/unauthorized.test.tsx`**: Update expected href from `api/auth/signin` to `api/auth/login`
- E2E tests (`e2e/fixtures/auth.ts`): No changes — still uses CredentialsProvider

---

## Files Summary

| File | Action |
|------|--------|
| `internal/database/migrations/20250217100000_add_esi_scopes.up.sql` | **New** — add esi_scopes column |
| `internal/database/migrations/20250217100000_add_esi_scopes.down.sql` | **New** — drop column |
| `internal/repositories/character.go` | Update — add EsiScopes field, update SQL |
| `internal/repositories/playerCorporations.go` | Update — add EsiScopes field, update SQL |
| `internal/controllers/corporations.go` | Update — pass EsiScopes to Upsert |
| `frontend/packages/client/api.ts` | Update — add esiScopes to FullCharacterData |
| `frontend/packages/client/auth/api.ts` | Update — unified OAuth, lazy discovery, flow types, export scopes |
| `frontend/pages/api/auth/callback.ts` | **New** — unified callback handler |
| `frontend/pages/api/auth/login.ts` | **New** — login entry point |
| `frontend/pages/api/auth/[...nextauth].ts` | Update — remove EVEOnlineProvider, export helpers |
| `frontend/pages/api/characters/add.ts` | Update — `getAuthUrl("char")` |
| `frontend/pages/api/corporations/add.ts` | Update — `getAuthUrl("corp")` |
| `frontend/pages/api/altAuth/callback.ts` | **Delete** |
| `frontend/packages/pages/index.tsx` | Update — sign-in link |
| `frontend/packages/components/unauthorized.tsx` | Update — sign-in link |
| `frontend/.env.local` | Update — remove ALT_ vars |
| `docs/features/infrastructure/railway-deployment.md` | Update — single callback URL, remove ALT_ vars |
| `frontend/packages/components/__tests__/unauthorized.test.tsx` | Update — expected href |
| `e2e/seed.sql` | Update — add esi_scopes to character inserts |

---

## Verification

1. `make test-backend` — migration runs, repository tests pass
2. `yarn build` in frontend — no TypeScript/build errors
3. `make test-frontend` — unauthorized test passes with updated href
4. `make test-e2e-ci` — CredentialsProvider login still works
5. Manual dev test: sign in → EVE SSO → callback creates session → home page
6. Manual dev test: add character → OAuth with ESI scopes → character added with scopes stored
7. Manual dev test: add corporation → OAuth with corp scopes → corporation added with scopes stored
8. Verify database: `SELECT esi_scopes FROM characters` shows space-delimited scopes string
