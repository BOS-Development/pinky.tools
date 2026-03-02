# Characters Feature

## Overview

Character management links EVE Online characters to user accounts, storing ESI (EVE Swagger Interface) tokens for API access. Each user can have multiple characters, and each character's assets are fetched independently from EVE's ESI API during asset refresh.

**Key page:** `/characters` — view all linked characters, add new ones, trigger asset refresh

## Business Context

EVE Online players commonly have multiple characters ("alts") spread across different accounts. To get a unified view of assets across all characters, each must be linked to the application with valid ESI credentials. The character management feature handles this linking and token storage.

### User Stories

**As an EVE Online player, I want to:**
1. Link my EVE characters to the application so their assets can be tracked
2. See all my linked characters at a glance
3. Refresh assets for all characters with one click
4. Add additional characters (alts) over time

## Architecture

### Data Flow

```
EVE OAuth Login → /api/altAuth/callback → decode JWT → POST /v1/characters/
    ↓
Backend upserts character with ESI tokens → characters table
    ↓
GET /v1/characters/ → returns all characters for user
    ↓
/characters page renders character cards with portraits
```

### Key Design Decisions

#### 1. ESI Token Storage
**Decision:** Store raw ESI access and refresh tokens per character.

The backend refreshes tokens as needed when making ESI API calls. Tokens are stored alongside the character record for simplicity.

#### 2. UPSERT on Add
**Decision:** Character add uses `INSERT ... ON CONFLICT DO UPDATE`.

If a character already exists for the user, re-adding it simply refreshes the ESI tokens. This handles token renewal gracefully. On re-auth, the UPSERT automatically resets `esi_needs_reauth` to `false`.

#### 3. User Isolation
**Decision:** Characters are scoped to users via `(id, user_id)` composite key.

The same EVE character ID could theoretically appear under different users. All queries filter by `user_id` from the authenticated session.

#### 4. ESI Hard 401 Detection
**Decision:** When ESI returns 401 Unauthorized or token refresh fails, set `esi_needs_reauth = true` and pause asset sync for that character.

The asset sync updater catches two scenarios:
1. **Refresh token failure** — if `RefreshAccessToken()` fails, the token is stale; mark for reauth
2. **ESI 401 response** — when fetching assets or location names returns 401, the access token is invalid; mark for reauth

Characters with `esi_needs_reauth = true` are skipped by the asset updater to avoid repeated 401s. The user sees an error alert on the characters page and must click "Re-authorize" to trigger a new OAuth flow, which resets the flag via UPSERT.

## Database Schema

### `characters` table

| Column | Type | Description |
|--------|------|-------------|
| `id` | `bigint` | EVE character ID |
| `user_id` | `bigint` | FK to `users(id)` |
| `name` | `text` | Character name |
| `esi_token` | `text` | ESI access token |
| `esi_refresh_token` | `text` | ESI refresh token |
| `esi_token_expires_on` | `timestamp` | Token expiry |
| `esi_scopes` | `text` | Comma-separated list of granted ESI scopes |
| `esi_needs_reauth` | `boolean` | Flag: set to `true` when ESI returns 401 or token refresh fails; pauses asset sync until user re-authorizes |

**Primary key:** `(id, user_id)`
**Conflict target:** `(id, user_id)` — upserts update name, tokens, scopes, and reset `esi_needs_reauth` to `false`

## Backend Implementation

### Repository (`internal/repositories/character.go`)

- **`Character` struct** — ID, Name, EsiToken, EsiRefreshToken, EsiTokenExpiresOn, UserID, EsiScopes, EsiNeedsReauth
- **`GetAll(ctx, userID)`** — fetches all characters for a user
- **`Get(ctx, id)`** — fetches a single character by ID
- **`Add(ctx, character)`** — upserts character with conflict on `(id, user_id)`; resets `esi_needs_reauth` to `false` on conflict
- **`SetNeedsReauth(ctx, id, userID, value)`** — sets `esi_needs_reauth` flag for a character

### Controller (`internal/controllers/characters.go`)

- **`GET /v1/characters/`** — returns all characters for the authenticated user (name + ID only)
- **`GET /v1/characters/{id}`** — returns a single character
- **`POST /v1/characters/`** — adds/updates a character; sets `UserID` from auth header

### Response Model

```go
type CharacterModel struct {
    ID          int64  `json:"id"`
    Name        string `json:"name"`
    EsiScopes   string `json:"esiScopes"`
    NeedsReauth bool   `json:"needsReauth"`
}
```

The API returns character ID, name, granted ESI scopes, and the re-auth flag. ESI tokens are never exposed to the frontend.

## Frontend Implementation

### Server-Side Props (`frontend/packages/pages/characters.tsx`)

Fetches characters via `api.getCharacters()` in `getServerSideProps`, passes to the `List` component.

### List Component (`frontend/packages/components/characters/list.tsx`)

- **Empty state:** "No Characters" message with "Add Character" button
- **Populated state:** "Characters" heading, "Add Character" button, card grid
- **Re-auth alert:** Red error `Alert` component displays for each character with `needsReauth: true`, with a "Re-authorize" button linking to `/api/characters/add`

### Card Component (`frontend/packages/components/characters/item.tsx`)

- Displays character portrait from `https://image.eveonline.com/Character/{id}_128.jpg`
- Shows character name below the portrait
- Hover animation (translateY + shadow increase)
- **Red error icon & border** if `needsReauth: true` — indicates ESI authorization revoked
- **Amber warning icon & border** if scopes are outdated — indicates scope update needed
- Each card includes a "Re-authorize" button if either condition is true

### Add Flow (`frontend/pages/api/characters/add.ts` → `frontend/pages/api/altAuth/callback.ts`)

1. `/api/characters/add` redirects to EVE OAuth
2. OAuth callback decodes JWT to extract character ID and name
3. Calls `api.addCharacter()` with ID, name, and ESI tokens
4. Redirects back to `/characters`

## Testing

### Unit Tests

- **Controller tests** (`internal/controllers/characters_test.go`):
  GetAll with mixed `needsReauth` states, Get success/missing ID/not found, Add success/invalid JSON/repository error

- **Repository tests** (`internal/repositories/character_test.go`):
  `SetNeedsReauth` sets and gets flag, `Add` UPSERT resets flag to false, user isolation on GetAll

- **Asset Updater tests** (`internal/updaters/assets_test.go`):
  Character with `EsiNeedsReauth: true` is skipped, refresh token failure sets flag, ESI 401 response sets flag

### E2E Tests (`e2e/tests/02-characters.spec.ts`)

1. Empty state — "No Characters" message displayed
2. Add Alice Alpha via E2E API — character appears on page
3. Add Alice Beta via E2E API — both characters visible
4. Add remaining characters (Bob, Charlie, Diana) for downstream tests
5. Character cards display portrait images with correct src
6. "Add Character" button visible
7. "Characters" heading displayed
8. Scope warning displayed for outdated scopes
9. **Re-auth banner displayed** when character has `needsReauth: true` — uses mock ESI 401 response

## Key Files

| File | Purpose |
|------|---------|
| `internal/repositories/character.go` | Character struct, DB operations, `SetNeedsReauth` method |
| `internal/client/esiClient.go` | `EsiUnauthorizedError` type for 401 detection |
| `internal/updaters/assets.go` | Asset sync logic, 401 detection and reauth flag setting |
| `internal/updaters/assets_test.go` | Asset updater unit tests, 401 and token failure tests |
| `internal/controllers/characters.go` | REST API handlers, returns `needsReauth` in response |
| `internal/controllers/characters_test.go` | Controller unit tests |
| `internal/repositories/character_test.go` | Repository integration tests, `SetNeedsReauth` test |
| `frontend/packages/pages/characters.tsx` | SSR page |
| `frontend/packages/components/characters/list.tsx` | List/empty state component, re-auth alert banner |
| `frontend/packages/components/characters/item.tsx` | Character card component, error/warning icons |
| `frontend/pages/api/characters/add.ts` | OAuth redirect route |
| `frontend/pages/api/altAuth/callback.ts` | OAuth callback handler |
| `frontend/pages/api/e2e/add-character.ts` | E2E test helper route |
| `e2e/tests/02-characters.spec.ts` | E2E tests, reauth banner test |
