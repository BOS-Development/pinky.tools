# Auto-Sell Containers Feature

## Overview

Auto-sell containers allow users to designate an in-game container as an "auto-sell" source. All items inside the container are automatically listed for sale at a configurable percentage of Jita buy price. Listings stay in sync when assets are refreshed or market prices update.

**GitHub Issue**: #24

## Business Context

Industrial players often sell materials from designated containers. Manually creating and maintaining individual sell orders for each item type is tedious and error-prone. Auto-sell containers automate this by:

1. Designating a container as auto-sell with a price percentage (e.g., 90% of Jita buy)
2. Automatically creating for-sale listings for all items in the container
3. Keeping quantities and prices in sync on asset refresh and market price updates
4. Removing listings when items leave the container

## Architecture

### Data Flow

```
User configures auto-sell container (POST /v1/auto-sell)
    ↓
Asset refresh (per-user, 1h) OR Market price update (6h)
    ↓
Auto-sell updater: SyncForUser / SyncForAllUsers
    ↓
For each container:
    Get items from character_assets / corporation_assets
    Get Jita buy prices from market_prices
    Compute price = jitaBuy * percentage / 100
    Upsert for-sale listings with auto_sell_container_id
    Deactivate listings for removed items / missing prices
```

### Key Design Decisions

#### 1. Price Source
**Decision**: Use Jita buy price (best bid) as the base price.
- Jita buy represents what buyers are willing to pay right now
- Percentage allows sellers to price above or below market
- Default 90% provides a slight discount to attract buyers

#### 2. Sync Triggers
**Decision**: Sync on both asset refresh and market price update.
- **Asset refresh** (per-user): Catches quantity changes, new items, removed items
- **Market price update** (all users): Updates prices when Jita market moves
- Immediate sync also triggered when creating/updating a config

#### 3. Soft-delete Pattern
**Decision**: Use `is_active = false` for both auto-sell configs and associated listings.
- Preserves history and avoids orphaned references
- Deactivating a config also deactivates all associated for-sale listings
- Re-creating an auto-sell for the same container upserts (ON CONFLICT)

#### 4. Optional Integration
**Decision**: Use `WithAutoSellUpdater()` setter pattern for assets and market price updaters.
- Avoids changing constructor signatures (breaking existing callers)
- Auto-sell sync is non-critical — errors are logged but don't propagate
- Easy to disable by not calling the setter

## Database Schema

### auto_sell_containers
```sql
create table auto_sell_containers (
    id bigserial primary key,
    user_id bigint not null references users(id),
    owner_type varchar(20) not null,     -- 'character' or 'corporation'
    owner_id bigint not null,
    location_id bigint not null,          -- station/structure ID
    container_id bigint not null,         -- in-game item_id of the container
    division_number int,                  -- corporation hangar division (null for character)
    price_percentage numeric(5, 2) not null default 90.00,
    is_active boolean not null default true,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now(),
    constraint auto_sell_valid_percentage check (price_percentage > 0 and price_percentage <= 200)
);

-- Unique constraint: one active auto-sell per container per user
create unique index idx_auto_sell_unique_container on auto_sell_containers(
    user_id, owner_type, owner_id, location_id, container_id,
    coalesce(division_number, 0)
) where is_active = true;
```

### for_sale_items (modified)
```sql
alter table for_sale_items add column auto_sell_container_id bigint references auto_sell_containers(id);
create index idx_for_sale_auto_sell on for_sale_items(auto_sell_container_id) where auto_sell_container_id is not null;
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/auto-sell` | Get user's active auto-sell configurations |
| POST | `/v1/auto-sell` | Create/upsert an auto-sell configuration |
| PUT | `/v1/auto-sell/{id}` | Update price percentage |
| DELETE | `/v1/auto-sell/{id}` | Soft-delete config and deactivate associated listings |

### POST /v1/auto-sell Request Body
```json
{
    "ownerType": "character",
    "ownerId": 12345,
    "locationId": 60003760,
    "containerId": 9000,
    "divisionNumber": null,
    "pricePercentage": 90.0
}
```

### PUT /v1/auto-sell/{id} Request Body
```json
{
    "pricePercentage": 85.0
}
```

## File Structure

### Backend
| File | Purpose |
|------|---------|
| `internal/database/migrations/20260217215240_auto_sell_containers.{up,down}.sql` | Create table |
| `internal/database/migrations/20260217215243_add_auto_sell_to_for_sale.{up,down}.sql` | Add FK column |
| `internal/models/models.go` | `AutoSellContainer`, `ContainerItem` structs |
| `internal/repositories/autoSellContainers.go` | CRUD + GetItemsInContainer |
| `internal/repositories/forSaleItems.go` | Added `auto_sell_container_id` support |
| `internal/updaters/autoSell.go` | Sync logic (SyncForUser, SyncForAllUsers) |
| `internal/updaters/assets.go` | Auto-sell hook after asset refresh |
| `internal/updaters/marketPrices.go` | Auto-sell hook after market price update |
| `internal/controllers/autoSellContainers.go` | CRUD endpoints |
| `cmd/industry-tool/cmd/root.go` | Wiring |

### Frontend
| File | Purpose |
|------|---------|
| `frontend/pages/api/auto-sell/index.ts` | GET/POST proxy |
| `frontend/pages/api/auto-sell/[id].ts` | PUT/DELETE proxy |
| `frontend/packages/components/assets/AssetsList.tsx` | Auto-sell button + dialog on containers |
| `frontend/packages/components/marketplace/MyListings.tsx` | "Auto" badge on managed listings |

### Tests
| File | Purpose |
|------|---------|
| `internal/updaters/autoSell_test.go` | 16 unit tests for sync logic |
| `internal/controllers/autoSellContainers_test.go` | 18 tests for CRUD endpoints |

## Edge Cases

- **Container removed from assets**: GetItemsInContainer returns empty, all listings deactivated
- **Item partially purchased**: Next asset refresh corrects quantity from ESI
- **No Jita buy price**: Item skipped, existing listing deactivated
- **Zero buy price**: Same as no price — listing deactivated
- **Nil buy price**: Same as no price — listing deactivated
- **Auto-sell config deleted**: All associated for-sale listings deactivated first
- **Duplicate config for same container**: Upserted via ON CONFLICT (updates percentage)

## Frontend UX

### AssetsList
- Container headers show an auto-sell toggle button (AutorenewIcon)
- Auto-sell enabled containers display a blue chip: `Auto-Sell @ 90% JBV`
- Configuration dialog with percentage input (1-200%), save/disable/cancel buttons
- For-sale chips on auto-sell items show "Auto" prefix with blue styling

### MyListings
- Auto-managed listings show a blue "Auto" chip next to the item name
- Tooltip warns: "Auto-managed listing — changes will be overwritten on next sync"
