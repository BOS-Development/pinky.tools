# Corporations Feature

## Overview

Corporation management tracks EVE Online corporation assets alongside personal character assets. Corporations are discovered via ESI's character affiliation endpoint — when a user links a character, the backend determines which corporation that character belongs to and stores the corporation with ESI credentials for asset access.

**Key page:** `/corporations` — view linked corporations, add new ones, trigger asset refresh

## Business Context

In EVE Online, corporations are player organizations that own shared assets (hangars, ships, materials). Industrial players need visibility into both personal and corporate inventories. This feature links corporations to the application, enabling the asset system to fetch and display corporate hangar contents alongside personal assets.

### User Stories

**As a corporation director, I want to:**
1. Link my corporation to the application so corp assets are tracked
2. See corporation assets organized by division (Main Hangar, Production Materials, etc.)
3. Set stockpile markers on corporation division assets
4. View corporation and personal assets together in the unified inventory view

## Architecture

### Data Flow

```
EVE OAuth Login (corp scope) → /api/altAuth/callback → decode JWT
    ↓
POST /v1/corporations (character data)
    ↓
Backend calls ESI POST /characters/affiliation → gets corp ID
Backend calls ESI GET /corporations/{corpID} → gets corp name
    ↓
Upsert player_corporations table
    ↓
Asset refresh: GET /corporations/{id}/assets + /divisions
    ↓
/corporations page renders corporation cards
/inventory page shows corp hangars under stations
```

### Key Design Decisions

#### 1. Corporation Discovery via Character Affiliation
**Decision:** Corporations are added by sending character credentials, not corporation IDs directly.

The backend determines the character's corporation via ESI's `/characters/affiliation` endpoint. This ensures the user actually has a character in the corporation and provides valid ESI tokens for corp API calls.

#### 2. Division Support
**Decision:** Corporation hangars are labeled by division names from ESI.

EVE corporations have up to 7 hangar divisions and 7 wallet divisions, each with customizable names. The application fetches these names during asset refresh to display meaningful labels like "Main Hangar" or "Production Materials" instead of generic "Division 1".

#### 3. Shared ESI Token
**Decision:** The corporation record stores the adding character's ESI token.

Corporation API calls require a character token with corp director/member roles. The token from the character used to add the corporation is stored and used for all subsequent corp API calls.

## Database Schema

### `player_corporations` table

| Column | Type | Description |
|--------|------|-------------|
| `id` | `bigint` | EVE corporation ID |
| `user_id` | `bigint` | FK to `users(id)` |
| `name` | `text` | Corporation name |
| `esi_token` | `text` | ESI access token (from adding character) |
| `esi_refresh_token` | `text` | ESI refresh token |
| `esi_token_expires_on` | `timestamp` | Token expiry |

**Primary key:** `(id, user_id)`

### `corporation_divisions` table

| Column | Type | Description |
|--------|------|-------------|
| `corporation_id` | `bigint` | FK to corporation |
| `user_id` | `bigint` | FK to user |
| `division` | `int` | Division number (1-7) |
| `name` | `text` | Division name |
| `type` | `text` | "hangar" or "wallet" |

## Backend Implementation

### Repository (`internal/repositories/playerCorporations.go`)

- **`PlayerCorporation` struct** — ID, UserID, Name, EsiToken, EsiRefreshToken, EsiExpiresOn
- **`Upsert(ctx, corp)`** — inserts or updates corporation record
- **`Get(ctx, userID)`** — fetches all corporations for a user
- **`GetDivisions(ctx, corpID, userID)`** — fetches hangar/wallet division names
- **`UpsertDivisions(ctx, corpID, userID, divisions)`** — updates division names from ESI

### Controller (`internal/controllers/corporations.go`)

- **`POST /v1/corporations`** — Add corporation flow:
  1. Decodes character data from request body
  2. Calls `esiClient.GetCharacterCorporation()` which:
     - POSTs to ESI `/characters/affiliation` with character ID
     - GETs `/corporations/{corpID}` for corporation name
  3. Upserts result into `player_corporations`
- **`GET /v1/corporations`** — Returns all corporations for the authenticated user (ID + name only)

### Response Model

```go
type PlayerCorporation struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
}
```

## Frontend Implementation

### Server-Side Props (`frontend/packages/pages/corporations.tsx`)

Fetches corporations via `api.getCorporations()` in `getServerSideProps`, passes to the `List` component.

### List Component (`frontend/packages/components/corporations/list.tsx`)

- **Empty state:** "No Corporations" message with "Add Corporation" button
- **Populated state:** "Corporations" heading, "Add Corporation" + "Refresh Assets" buttons, card grid

### Card Component (`frontend/packages/components/corporations/item.tsx`)

- Displays corporation logo from `https://images.evetech.net/corporations/{id}/logo`
- Shows "Corporation" chip with business icon
- Shows corporation name in primary color
- Fallback: hides broken logo images gracefully

### Add Flow (`frontend/pages/api/corporations/add.ts` → `frontend/pages/api/altAuth/callback.ts`)

1. `/api/corporations/add` redirects to EVE OAuth with corp scope
2. OAuth callback decodes JWT to extract character ID
3. Calls `api.addCharacterCorporation()` with character data
4. Backend discovers corporation via ESI affiliation
5. Redirects back to `/corporations`

## Testing

### Unit Tests

- **Controller tests** (6 tests in `internal/controllers/corporations_test.go`):
  Get success/error, Add success/invalid JSON/ESI error/repository error

- **Repository tests** (6 tests in `internal/repositories/playerCorporations_test.go`):
  Upsert and get, update existing, upsert divisions, update divisions, nil divisions, empty divisions

### E2E Tests (5 tests in `e2e/tests/03-corporations.spec.ts`)

1. Empty state — "No Corporations" message displayed
2. Add Stargazer Industries via E2E API (mock ESI affiliation discovery)
3. Corporation card displays name and "Corporation" chip
4. "Add Corporation" and "Refresh Assets" buttons visible
5. "Corporations" heading displayed

## Key Files

| File | Purpose |
|------|---------|
| `internal/repositories/playerCorporations.go` | Corporation struct and DB operations |
| `internal/controllers/corporations.go` | REST API handlers + ESI integration |
| `internal/controllers/corporations_test.go` | Controller unit tests |
| `internal/repositories/playerCorporations_test.go` | Repository integration tests |
| `frontend/packages/pages/corporations.tsx` | SSR page |
| `frontend/packages/components/corporations/list.tsx` | List/empty state component |
| `frontend/packages/components/corporations/item.tsx` | Corporation card component |
| `frontend/pages/api/corporations/add.ts` | OAuth redirect route |
| `frontend/pages/api/altAuth/callback.ts` | OAuth callback handler |
| `frontend/pages/api/e2e/add-corporation.ts` | E2E test helper route |
| `e2e/tests/03-corporations.spec.ts` | E2E tests |
| `cmd/mock-esi/main.go` | Mock ESI server (affiliation + corp info endpoints) |
