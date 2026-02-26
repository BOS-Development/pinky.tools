# Auto-Buy for Understocked Stockpiles

## Overview

Automatically creates and maintains buy orders for stockpile items that are below their desired quantity. Uses Jita market pricing (same options as auto-sell: jita_buy, jita_sell, jita_split) with a configurable percentage.

## Status

- **Phase**: Implementation
- **Branch**: `feature/auto-buy`

## Key Decisions

1. **Container-level config with per-item override** — An `auto_buy_configs` entry targets a container (same key structure as auto-sell). Individual stockpile markers can override the pricing.
2. **Mirrors auto-sell pattern** — Same table structure, sync triggers, and pricing logic.
3. **Buy orders deactivate on filled deficit** — When current quantity >= desired quantity, the linked buy order is deactivated. Reactivated when deficit returns.
4. **Same trigger points as auto-sell** — After asset refresh (`SyncForUser`) and after market price update (`SyncForAllUsers`).

## Schema

### `auto_buy_configs` (NEW)
Same shape as `auto_sell_containers`, with nullable `container_id`:
- `id`, `user_id`, `owner_type`, `owner_id`, `location_id`, `container_id` (nullable), `division_number` (nullable)
- `price_percentage` (default 100), `price_source` (default 'jita_sell')
- `is_active`, `created_at`, `updated_at`

### `buy_orders` (MODIFIED)
- Added `auto_buy_config_id` — links auto-generated orders to the config

### `stockpile_markers` (MODIFIED)
- Added `price_source`, `price_percentage` — nullable per-item overrides

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/auto-buy` | List user's auto-buy configs |
| POST | `/v1/auto-buy` | Create config (triggers sync) |
| PUT | `/v1/auto-buy/{id}` | Update config (triggers sync) |
| DELETE | `/v1/auto-buy/{id}` | Delete config (deactivates linked orders) |

## Sync Logic

1. Get stockpile markers matching the config's container context
2. Compare desired quantities with current asset quantities → compute deficits
3. Look up Jita prices for deficit item types
4. For each deficit: compute `maxPrice = basePrice * percentage / 100`, upsert buy order
5. Deactivate buy orders for items no longer in deficit

## File Structure

### Backend
- `internal/database/migrations/*_auto_buy*` — 3 migration pairs
- `internal/models/models.go` — AutoBuyConfig model + extensions
- `internal/repositories/autoBuyConfigs.go` — CRUD + deficit query
- `internal/repositories/buyOrders.go` — Auto-buy methods added
- `internal/repositories/stockpileMarkers.go` — Pricing columns added
- `internal/updaters/autoBuy.go` — Sync logic
- `internal/updaters/assets.go` — WithAutoBuyUpdater hook
- `internal/updaters/marketPrices.go` — WithAutoBuyUpdater hook
- `internal/controllers/autoBuyConfigs.go` — REST endpoints
- `cmd/industry-tool/cmd/root.go` — Wiring

### Frontend
- `frontend/pages/api/auto-buy/` — API proxy routes
- `frontend/packages/components/assets/AssetsList.tsx` — Auto-buy dialog + indicators
- `frontend/packages/components/marketplace/BuyOrders.tsx` — Auto badge on orders
