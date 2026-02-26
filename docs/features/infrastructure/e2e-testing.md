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
- Users (created when Playwright logs in via CredentialsProvider)

Characters and corporations are created during tests via E2E API routes (see below), exercising the full backend creation flow through the mock ESI.

### 4. Dynamic Mock ESI (Runtime Data Manipulation)

The mock ESI server (`cmd/mock-esi/main.go`) uses a thread-safe `State` struct protected by `sync.RWMutex`. All existing handlers read state under `RLock()`. Admin API endpoints (`/_admin/*`) acquire `Lock()` for writes. This allows tests to inject, replace, or reset mock ESI data at runtime without restarting the server.

The `/_admin/reset` endpoint restores all data to the hardcoded defaults, ensuring test isolation. Tests that mutate mock state should call `resetMockESI()` in `afterAll`.

### 5. E2E API Routes (test-only endpoints)

Two Next.js API routes are available only when `E2E_TESTING=true`:

- **`POST /api/e2e/add-character`** — Creates a character by calling `POST /v1/characters/` on the backend. Accepts `{ userId, characterId, characterName }` and generates fake ESI tokens.

- **`POST /api/e2e/add-corporation`** — Creates a corporation by calling `POST /v1/corporations` on the backend, which triggers the full ESI affiliation discovery flow (mock ESI `POST /characters/affiliation` → `GET /corporations/{id}`). Accepts `{ userId, characterId, characterName }`.

Both routes return 404 when `E2E_TESTING` is not set, preventing accidental use in production.

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

1. **Landing + Auth** (`01-landing.spec.ts`) — Login via CredentialsProvider, verify authenticated landing page
2. **Characters** (`02-characters.spec.ts`) — Empty state, add characters via E2E API, verify cards/portraits/buttons
3. **Corporations** (`03-corporations.spec.ts`) — Empty state, add corporation via E2E API (mock ESI affiliation), verify card/chip/buttons
4. **Assets** (`04-assets.spec.ts`) — Refresh Assets triggers mock ESI, assets populate by station (Jita, Amarr), containers, corp divisions
5. **Navigation** (`05-navigation.spec.ts`) — All navbar links resolve to working pages
6. **Stockpile Workflow** (`06-stockpiles.spec.ts`) — Set/edit/delete stockpile markers from assets page, verify deficit calculations on stockpiles page
7. **Contacts Workflow** (`07-contacts.spec.ts`) — Send contact request (Alice → Bob), accept, verify bidirectional connection
8. **Marketplace Workflow** (`08-marketplace.spec.ts`) — Create listing (Bob), browse (Alice), purchase, buy orders
9. **Auto-Sell** (`09-auto-sell.spec.ts`) — Auto-sell container configuration
10. **Industry** (`10-industry.spec.ts`) — Industry jobs page, active jobs sync from mock ESI, queue empty state, add job tab, blueprint search, dynamic mock ESI injection
11. **Stations** (`11-stations.spec.ts`) — Station management CRUD: add/edit/delete preferred stations
12. **Production Plans** (`12-production-plans.spec.ts`) — Create/edit/delete production plans, plan editor with step tree, blueprint search
13. **Reactions** (`13-reactions.spec.ts`) — Reactions calculator tabs, settings toolbar, reaction list, shopping list, plan summary
14. **Planetary Industry** (`14-pi.spec.ts`) — PI overview/profit/supply chain tabs, dynamic mock PI planet injection, stats chips
15. **Transport** (`15-transport.spec.ts`) — Transport profiles CRUD, JF routes dialog, transport jobs
16. **Settings** (`16-settings.spec.ts`) — Settings page, Discord settings section

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
| `GET /characters/{id}/skills` | Character skills (per-character canned data) |
| `GET /characters/{id}/industry/jobs` | Character industry jobs (per-character canned data) |
| `GET /characters/{id}/blueprints` | Character blueprints (per-character canned data) |
| `GET /characters/{id}/contracts` | Character contracts (empty array) |
| `GET /characters/{id}/planets` | Character PI planets (per-character canned data) |
| `GET /characters/{id}/planets/{planetID}` | Planet colony details (per-character canned data) |
| `GET /latest/route/{origin}/{dest}` | Jump route between systems (system ID array) |

**Note:** All character endpoints support versioned paths (`/v1/characters/`, `/v3/characters/`, `/v4/characters/`).

---

## Dynamic Mock ESI (Admin API)

The mock ESI server exposes an admin API to manipulate mock data at runtime. This allows tests to inject, replace, or reset mock ESI state without restarting the server or relying on pre-canned data. The admin API is **only available when `E2E_TESTING=true`**.

### Admin API Endpoints

| Method | Path | Body | Purpose |
|--------|------|------|---------|
| POST | `/_admin/reset` | none | Reset all data to defaults |
| PUT | `/_admin/character-assets/{charID}` | `[]asset` | Replace character assets |
| PUT | `/_admin/character-skills/{charID}` | `skillsResponse` | Replace character skills |
| PUT | `/_admin/character-industry-jobs/{charID}` | `[]industryJob` | Replace industry jobs |
| PUT | `/_admin/character-blueprints/{charID}` | `[]blueprintEntry` | Replace blueprints |
| PUT | `/_admin/corp-assets/{corpID}` | `[]asset` | Replace corp assets |
| PUT | `/_admin/market-orders` | `[]marketOrder` | Replace all market orders |
| PUT | `/_admin/character-planets/{charID}` | `[]piPlanet` | Replace character PI planets |
| PUT | `/_admin/planet-details/{charID}/{planetID}` | `piColony` | Replace planet colony details |

### Playwright Helper

The `e2e/helpers/mock-esi.ts` module wraps these endpoints with typed TypeScript functions. Import and use in tests:

```typescript
import { setCharacterIndustryJobs, resetMockESI, type IndustryJob } from '../helpers/mock-esi';

test.afterAll(async () => {
  await resetMockESI();
});
```

### Rules

- Always call `resetMockESI()` in `afterAll` if any test mutated mock state
- Prefer admin API over adding new canned data for single-test scenarios
- Use `toPass` polling pattern when waiting for background runners to sync changed mock data

---

## File Structure

```
e2e/
  package.json              # Playwright dependency
  tsconfig.json             # TypeScript config
  playwright.config.ts      # Chromium only, workers: 1, sequential
  global-setup.ts           # Authenticates as Alice Stargazer
  seed.sql                  # Bootstrap data (static universe + users only)
  fixtures/
    auth.ts                 # Multi-user auth fixtures (Alice, Bob, Charlie, Diana)
  helpers/
    mock-esi.ts             # Playwright helper for mock ESI admin API
  tests/
    01-landing.spec.ts      # Auth + landing page
    02-characters.spec.ts   # Character creation + display
    03-corporations.spec.ts # Corporation creation via mock ESI
    04-assets.spec.ts       # Asset refresh + display
    05-navigation.spec.ts   # Navbar links
    06-stockpiles.spec.ts   # Stockpile markers + deficits
    07-contacts.spec.ts     # Contact requests + acceptance
    08-marketplace.spec.ts  # Listings, purchases, buy orders
    09-auto-sell.spec.ts    # Auto-sell container configuration
    10-industry.spec.ts     # Industry jobs + dynamic mock ESI
    11-stations.spec.ts     # Station management CRUD
    12-production-plans.spec.ts  # Production plan lifecycle
    13-reactions.spec.ts    # Reactions calculator
    14-pi.spec.ts           # Planetary Industry
    15-transport.spec.ts    # Transport profiles + JF routes
    16-settings.spec.ts     # Settings page

frontend/pages/api/e2e/
  add-character.ts          # E2E-only: create character via backend
  add-corporation.ts        # E2E-only: create corporation via backend + mock ESI

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
