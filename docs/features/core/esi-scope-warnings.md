# ESI Scope Warnings

## Status

Implemented

## Overview

When new ESI scopes are added to the application, existing characters and corporations have stale scope tokens stored in the database. The backend silently skips ESI calls for entities missing required scopes, but users receive no indication of this limitation. The ESI Scope Warnings feature adds visible warnings so users know when they need to re-authorize a character or corporation to enable new functionality.

## Business Context

EVE Online's ESI API requires specific OAuth scopes to access different data endpoints. As the application gains new features (e.g., industry jobs, planetary industry data), new scopes are added to the required scope list. Characters and corporations added before these scope additions retain their original scopes, preventing the new features from functioning.

Without warnings, users experience silent failures: a new feature appears in the UI but returns no data, leading to confusion. This feature bridges that gap by surfacing scope mismatches prominently on character/corporation cards and in a global navbar alert.

## How It Works

### Scope Storage and Comparison

1. **Database storage**: ESI scopes are stored as a space-delimited string in the `esi_scopes TEXT` column on both `characters` and `player_corporations` tables (added by the [consolidate-oauth.md](consolidate-oauth.md) migration)
2. **Required scopes definition**: The frontend defines two scope arrays in `frontend/packages/client/scope-definitions.ts`:
   - `CHARACTER_REQUIRED_SCOPES` — array of scope strings required for character endpoints
   - `CORPORATION_REQUIRED_SCOPES` — array of scope strings required for corporation endpoints
3. **Scope comparison happens in the frontend** — the UI compares stored scopes against the current required scope arrays using a utility function
4. **Scope utility function** (`frontend/packages/client/scopes.ts`):
   - `areCharacterScopesOutdated(storedScopes: string): boolean` — returns true if stored scopes don't match CHARACTER_REQUIRED_SCOPES
   - `areCorporationScopesOutdated(storedScopes: string): boolean` — returns true if stored scopes don't match CORPORATION_REQUIRED_SCOPES

### User-Facing Warnings

#### Character/Corporation Cards
- Character and corporation item cards (`frontend/packages/components/characters/item.tsx` and `frontend/packages/components/corporations/item.tsx`) display a red warning border and "Re-authorize" button when `areCharacterScopesOutdated()` or `areCorporationScopesOutdated()` returns true
- Clicking "Re-authorize" triggers the existing OAuth flow for that entity (same as "Add Character" / "Add Corporation")
- Re-adding the entity through OAuth updates the scopes via UPSERT on the backend

#### Global Navbar Alert
- The navbar component (`frontend/packages/components/Navbar.tsx`) fetches scope status once on mount via a new frontend API route
- If any character or corporation has outdated scopes, a persistent alert banner displays: "You have characters/corporations with outdated scopes. Please re-authorize them to access new features."
- The banner includes a "Show" button that navigates to `/characters` or `/corporations` depending on which has outdated entities

### Re-Authorization Flow

Re-authorization follows the existing OAuth flow:
1. User clicks "Re-authorize" on a character/corporation card
2. App redirects to `/api/characters/add` or `/api/corporations/add`
3. OAuth flow proceeds as normal (user approves permissions at EVE SSO)
4. Backend receives the token with current scopes
5. UPSERT updates the `esi_scopes` column with the new scopes
6. Page redirects back to `/characters` or `/corporations`
7. Next page load compares stored scopes against required scopes — warning disappears if scopes now match

## Key Decisions

### 1. **Scope Comparison Happens in Frontend**
**Decision**: The frontend is the single source of truth for required scopes. The backend only stores the raw scope string.

**Rationale**: Scope requirements are UI-level concerns tied to feature availability. Defining them in the frontend keeps all feature config in one place. The backend has no logic to compare scopes — it simply stores what it receives from OAuth.

### 2. **Scope Definitions in Separate File**
**Decision**: Scope arrays live in `frontend/packages/client/scope-definitions.ts`, not in auth module.

**Rationale**: The auth module (`frontend/packages/client/auth/api.ts`) is server-side only and uses environment variables. Feature components and client-side utilities need to import scope definitions, so they must be in a shared, non-SSR file. The definitions export raw scope strings that both the auth code (for requesting from EVE SSO) and the client-side UI (for comparison) can use.

### 3. **Navbar Fetches Scope Status Once on Mount**
**Decision**: The navbar calls `/api/scope-status` once when it loads, not on every page navigation.

**Rationale**: Scope status is stable between page loads (scopes only change via re-authorization). A single fetch on navbar mount keeps UI responsive. If users re-authorize, the next full page load includes the updated character/corporation list, so the navbar sees the new scopes automatically.

### 4. **No Polling or Real-Time Updates**
**Decision**: Scope status is not polled or updated in real-time.

**Rationale**: Scope changes only occur when users explicitly re-authorize. A single fetch on mount is sufficient; users will see the updated status on the next full page load or manual refresh.

## Schema

No new tables or migrations. The feature uses existing columns added by the [consolidate-oauth.md](consolidate-oauth.md) feature:

| Table | Column | Type | Description |
|-------|--------|------|-------------|
| `characters` | `esi_scopes` | `TEXT` | Space-delimited scope string (e.g., `"esi-assets.read_assets.v1 esi-skills.read_skills.v1"`) |
| `player_corporations` | `esi_scopes` | `TEXT` | Space-delimited scope string |

## API Endpoints

### Backend (Existing Endpoints, Enhanced)

| Method | Path | Changes |
|--------|------|---------|
| `GET` | `/v1/characters/` | Response now includes `esiScopes: string` field |
| `GET` | `/v1/characters/{id}` | Response now includes `esiScopes: string` field |
| `GET` | `/v1/corporations` | Response now includes `esiScopes: string` field |
| `GET` | `/v1/corporations/{id}` | Response now includes `esiScopes: string` field |

Example character response:
```json
{
  "id": 2150003001,
  "name": "Alice Alpha",
  "esiScopes": "esi-assets.read_assets.v1 esi-skills.read_skills.v1"
}
```

### Frontend API (New)

| Method | Path | Description | Response |
|--------|------|-------------|----------|
| `GET` | `/api/scope-status` | Returns scope status for all characters and corporations | `{ outdatedCharacters: CharacterData[], outdatedCorporations: CorporationData[], hasOutdated: boolean }` |

## File Paths

### Scope Definitions and Utilities
- **Scope definitions**: `frontend/packages/client/scope-definitions.ts`
- **Scope utility functions**: `frontend/packages/client/scopes.ts`

### Backend (Controllers)
- **Character controller**: `internal/controllers/characters.go` — includes `esiScopes` in response model
- **Corporation controller**: `internal/controllers/corporations.go` — includes `esiScopes` in response model

### Frontend (Components)
- **Character card**: `frontend/packages/components/characters/item.tsx` — displays warning border and "Re-authorize" button if scopes outdated
- **Corporation card**: `frontend/packages/components/corporations/item.tsx` — displays warning border and "Re-authorize" button if scopes outdated
- **Navbar**: `frontend/packages/components/Navbar.tsx` — fetches scope status and displays global alert
- **Scope status API route**: `frontend/pages/api/scope-status.ts` — new route that aggregates scope status across all characters and corporations

### Models and Types
- **Frontend API client**: `frontend/packages/client/api.ts` — `FullCharacterData` and `CorporationData` types include `esiScopes: string`

## Testing

### Unit Tests
- **Scope utility tests**: `frontend/packages/client/__tests__/scopes.test.ts` — tests for `areCharacterScopesOutdated()` and `areCorporationScopesOutdated()`

### Integration Tests
- **Navbar tests**: `frontend/packages/components/__tests__/Navbar.test.tsx` — mocks `/api/scope-status`, verifies alert displays when scopes outdated
- **Character card tests**: `frontend/packages/components/__tests__/characters/item.test.tsx` — verifies warning border and "Re-authorize" button appear
- **Corporation card tests**: `frontend/packages/components/__tests__/corporations/item.test.tsx` — verifies warning border and "Re-authorize" button appear

### E2E Tests
- **Character scope warning flow** (`e2e/tests/15-scope-warnings.spec.ts`):
  1. Add a character with minimal scopes
  2. Verify warning border on character card
  3. Verify global navbar alert displays
  4. Re-authorize character with full scopes
  5. Verify warning disappears from both card and navbar
- **Corporation scope warning flow**:
  1. Add a corporation with minimal scopes
  2. Verify warning border on corporation card
  3. Re-authorize corporation with full scopes
  4. Verify warning disappears

## Cross-References

- [consolidate-oauth.md](consolidate-oauth.md) — Adds `esi_scopes` column to characters and corporations tables
- [characters.md](characters.md) — Character management and OAuth flow
- [corporations.md](corporations.md) — Corporation management and OAuth flow

## Open Questions

- None at this time
