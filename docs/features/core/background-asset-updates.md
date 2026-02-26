# Background Asset Updates

## Overview

Automatically refresh character and corporation assets from ESI on a background timer, removing the need for users to manually click "Refresh Assets". Also triggers targeted asset updates when a user adds a new character or corporation.

## Status

- **Phase**: Implementation
- **Branch**: `feature/background-asset-updates`

## Key Decisions

- **Interval**: 1 hour (matches ESI asset endpoint cache timer of 3600s)
- **Configurable**: `ASSET_UPDATE_INTERVAL_SEC` env var (default 3600), `ASSET_UPDATE_CONCURRENCY` env var (default 5)
- **Concurrency**: Characters and corporations are updated concurrently within a configurable semaphore limit using `sync.WaitGroup` + buffered channel
- **Error isolation**: Individual character/corporation failures are logged but do not stop other updates
- **On-add trigger**: Adding a character updates only that character's assets; adding a corporation updates only that corporation's assets (not all user assets)
- **Last updated display**: Users see "Last updated: X ago | Next update: in Y" on the inventory page

## Schema

```sql
alter table users add column assets_last_updated_at timestamptz;
```

## File Structure

### Backend
- `cmd/industry-tool/cmd/settings.go` — New settings
- `internal/repositories/user.go` — `GetAllIDs()`, `UpdateAssetsLastUpdated()`, `GetAssetsLastUpdated()`
- `internal/updaters/assets.go` — Concurrency refactor, public per-entity methods, timestamp update
- `internal/runners/assets.go` — New asset runner (follows MarketPricesRunner pattern)
- `internal/controllers/users.go` — Remove RefreshAssets, add GetAssetStatus
- `internal/controllers/characters.go` — Trigger character asset update on add
- `internal/controllers/corporations.go` — Trigger corporation asset update on add

### Frontend
- `frontend/packages/components/assets/AssetsList.tsx` — Display last updated / next update
- `frontend/packages/client/api.ts` — Remove `refreshAssets()`, add `getAssetStatus()`
- `frontend/pages/api/assets/status.ts` — New proxy route
- Characters/corporations list components — Remove Refresh Assets buttons

### Deleted
- `frontend/pages/api/characters/refreshAssets.ts`
- `frontend/pages/api/corporations/refreshAssets.ts`

## Verification

1. `make test-backend` — all tests pass
2. `make test-frontend` — all tests pass (update snapshots)
3. `make test-e2e-ci` — E2E tests pass without "Refresh Assets" buttons
