# End-to-End Testing

## Overview

Full-stack browser-based end-to-end testing using Playwright. Tests exercise the entire application through a real browser: Next.js frontend, Go backend, PostgreSQL database, and a mock ESI service. All external dependencies (EVE Online ESI API, EVE OAuth) are mocked for deterministic, self-contained tests.

**Status:** Implemented

---

## Architecture

```
Playwright (browser) → Next.js Frontend → Go Backend → Mock ESI Server
                                              ↓
                                         PostgreSQL
```

All services run via `docker-compose.e2e.yaml`. Auth is mocked via a NextAuth CredentialsProvider (no EVE OAuth needed). The ESI client points at a mock Go HTTP server instead of the real `esi.evetech.net`.

---

## Key Design Decisions

### 1. Mock ESI Service (not real ESI)

The backend's ESI client (`internal/client/esiClient.go`) has a configurable `baseURL` field (via `ESI_BASE_URL` env var). In E2E mode, this points to `cmd/mock-esi/main.go` — a lightweight Go HTTP server that returns deterministic canned responses for all ESI endpoints.

When `ESI_BASE_URL` is set, the backend uses a plain `http.Client` instead of the OAuth token client, bypassing ESI token refresh entirely.

### 2. Mock Auth (not real EVE OAuth)

When `E2E_TESTING=true`, the NextAuth config (`frontend/pages/api/auth/[...nextauth].ts`) adds a `CredentialsProvider` that accepts `userId` + `userName` fields. This creates a real NextAuth JWT session without going through EVE OAuth. The JWT callback sets `token.providerAccountId` from `user.id`, so all downstream API routes work identically.

### 3. Mock API First (not pre-seeded game data)

Game data (assets, market prices, corp divisions) is NOT pre-seeded into the database. Instead, tests trigger app actions (Refresh Assets, Update Market Prices) which hit the mock ESI to populate data. This tests the full data pipeline.

Only bootstrap data that can't be created through the app is seeded via SQL:
- Static universe data (regions, systems, stations, item types)
- Characters and corporations with fake ESI tokens
- Users (created when Playwright logs in via CredentialsProvider)

---

## Running E2E Tests

### Prerequisites

```bash
cd e2e
npm install
npx playwright install chromium
```

### Local (headless)

```bash
make test-e2e
```

### Local (interactive UI)

```bash
make test-e2e-ui
```

### Cleanup

```bash
make test-e2e-clean
```

### CI

Runs automatically via the `e2e-tests` job in `.github/workflows/ci.yml`. Uploads Playwright HTML report and failure screenshots as artifacts.

---

## Test Users

| User ID | Name | Characters | Role |
|---|---|---|---|
| 1001 | Alice Stargazer | Alice Alpha (2001001), Alice Beta (2001002) | Primary test user, logged in by default |
| 1002 | Bob Miner | Bob Bravo (2002001) | Contact/marketplace partner |
| 1003 | Charlie Trader | Charlie Charlie (2003001) | Pending contact tests |
| 1004 | Diana Scout | Diana Delta (2004001) | Isolation tests (no relationships) |

**Corporation:** Stargazer Industries (3001001), owned by Alice (user 1001)

---

## Test Flows

1. **Landing + Auth** — Login via CredentialsProvider, verify authenticated landing page
2. **Navigation** — All navbar links resolve to working pages
3. **Characters** — Character list shows seeded characters
4. **Asset Refresh** — Refresh Assets triggers mock ESI, assets populate by station (Jita, Amarr), containers, corp divisions
5. **Stockpile Workflow** — Set/edit/delete stockpile markers from assets page, verify deficit calculations on stockpiles page
6. **Contacts Workflow** — Send contact request (Alice → Bob), accept, verify bidirectional connection
7. **Marketplace Workflow** — Create listing (Bob), browse (Alice), purchase, buy orders

---

## Mock ESI Endpoints

| Endpoint | Data Returned |
|---|---|
| `GET /characters/{id}/assets` | Character assets (per-character canned data) |
| `POST /characters/{id}/assets/names` | Container name mappings |
| `POST /characters/affiliation` | Character → corporation mapping |
| `GET /corporations/{id}` | Corporation name/info |
| `GET /corporations/{id}/assets` | Corporation assets |
| `POST /corporations/{id}/assets/names` | Corp container names |
| `GET /corporations/{id}/divisions` | Hangar/wallet divisions |
| `GET /universe/structures/{id}` | Returns 403 (no player-owned structures in test data) |
| `GET /latest/markets/{regionID}/orders/` | Market buy/sell orders for The Forge |

---

## File Structure

```
e2e/
  package.json              # Playwright dependency
  tsconfig.json             # TypeScript config
  playwright.config.ts      # Chromium only, workers: 1, sequential
  global-setup.ts           # Authenticates as Alice Stargazer
  seed.sql                  # Bootstrap data (static universe, characters, corps)
  fixtures/
    auth.ts                 # Multi-user auth fixtures (Alice, Bob, Charlie, Diana)
  tests/
    landing.spec.ts
    navigation.spec.ts
    characters.spec.ts
    assets.spec.ts
    stockpiles.spec.ts
    contacts.spec.ts
    marketplace.spec.ts

cmd/mock-esi/
  main.go                   # Mock ESI HTTP server

Dockerfile.mock-esi         # Docker build for mock ESI
docker-compose.e2e.yaml     # Full stack E2E environment
```

---

## Modified Existing Files

| File | Change |
|---|---|
| `internal/client/esiClient.go` | Added configurable `baseURL` field (default: `https://esi.evetech.net`) |
| `cmd/industry-tool/cmd/settings.go` | Added `EsiBaseURL` from `ESI_BASE_URL` env var |
| `cmd/industry-tool/cmd/root.go` | When `ESI_BASE_URL` is set, uses plain HTTP client (bypasses OAuth) |
| `frontend/pages/api/auth/[...nextauth].ts` | Added CredentialsProvider when `E2E_TESTING=true` |
| `makefile` | Added `test-e2e`, `test-e2e-ui`, `test-e2e-clean` targets |
| `.github/workflows/ci.yml` | Added `e2e-tests` job, updated `test-summary` |

---

## Adding New Tests

1. Create a new `.spec.ts` file in `e2e/tests/`
2. Import from `@playwright/test` for single-user tests, or from `../fixtures/auth` for multi-user tests
3. Tests run sequentially (shared database state) — design tests to be additive
4. If new mock ESI data is needed, update `cmd/mock-esi/main.go`
5. If new seed data is needed, update `e2e/seed.sql`
