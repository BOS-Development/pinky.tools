# Arbiter — T2 Industry Opportunity Scanner

## Status
Implemented (Phase 2 complete)

## Overview

Arbiter is a comprehensive T2 (Tech 2) industry opportunity scanner that helps players identify profitable manufacturing chains. It combines real-time market data with player-owned structure costs to analyze full production chains (reactions → moon goo → components → final modules), including invention requirements and job costs. Phase 2 adds multi-scope asset tracking, BOM tree visualization with delta calculation, customizable tax profiles, and item black/whitelisting.

## Key Decisions

- **Scope-based asset aggregation**: Player assets are grouped by scope (character + corp divisions). The BOM tree loads only assets within a selected scope, ensuring accurate delta (needed - available) calculations. Scopes can include multiple characters and corporation divisions.

- **Dual price types for materials**: The `input_price_type` (buy/sell order pricing) controls cost calculations for input materials. For output products, `output_price_type` controls revenue calculations. Each defaults to the opposite of the other (buy for inputs, sell for outputs) but can be customized per tax profile. If the preferred type is unavailable, the system falls back to the other type.

- **Full batch costs for reactions**: Reaction jobs always charge the full batch cost regardless of quantity needed. There is no pro-rating across multiple output units — if you run 1 batch, you pay for all 16 units of output. This matches how reactions work in EVE.

- **Blacklist/whitelist enforcement**: Items on the blacklist are never considered for building (treated as "buy only"); whitelist items are forcibly built if a blueprint exists. Both defaults are true (use both lists), configurable per user.

- **Two-hour market price cache**: Market prices update every 2 hours (UpdateInterval in `updaters/marketPrices.go`). Arbiter results reflect Jita prices at the last update time.

- **Full production chains**: The BOM tree recursively walks the blueprint tree up to 10 levels deep, alternating between manufacturing and reactions. Sub-components can be purchased from market if cheaper than building, or built if a blueprint exists and the cost is lower.

## Database Schema

### Core Tables

```sql
-- Arbiter settings per user (structure tiers, facility taxes, defaults)
CREATE TABLE arbiter_settings (
  user_id BIGINT PRIMARY KEY,
  -- Reaction (moon goo) structure/rig/location
  reaction_structure VARCHAR(50),
  reaction_rig VARCHAR(50),
  reaction_system_id BIGINT,
  reaction_facility_tax NUMERIC(5,2),
  -- Invention structure/rig/location
  invention_structure VARCHAR(50),
  invention_rig VARCHAR(50),
  invention_system_id BIGINT,
  invention_facility_tax NUMERIC(5,2),
  -- Component manufacturing structure/rig/location
  component_structure VARCHAR(50),
  component_rig VARCHAR(50),
  component_system_id BIGINT,
  component_facility_tax NUMERIC(5,2),
  -- Final manufacturing structure/rig/location
  final_structure VARCHAR(50),
  final_rig VARCHAR(50),
  final_system_id BIGINT,
  final_facility_tax NUMERIC(5,2),
  -- Item filtering
  use_whitelist BOOLEAN DEFAULT true,
  use_blacklist BOOLEAN DEFAULT true,
  decryptor_type_id BIGINT,
  default_scope_id BIGINT,
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Tax profiles: customizable price types and broker/facility fees
CREATE TABLE arbiter_tax_profiles (
  user_id BIGINT PRIMARY KEY,
  broker_fee_rate NUMERIC(5,4),       -- Player broker fee %
  structure_broker_fee NUMERIC(5,4),  -- Market order fee in player structure
  input_price_type VARCHAR(10),       -- "buy" or "sell" for material costs
  output_price_type VARCHAR(10),      -- "buy" or "sell" for revenue
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Scopes: groupings of assets for analysis (chars + corp divisions)
CREATE TABLE arbiter_scopes (
  id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  user_id BIGINT NOT NULL,
  name VARCHAR(256) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Scope members: characters and corp divisions within a scope
CREATE TABLE arbiter_scope_members (
  id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  scope_id BIGINT NOT NULL REFERENCES arbiter_scopes(id) ON DELETE CASCADE,
  character_id BIGINT NOT NULL,
  corporation_division_id BIGINT,  -- NULL means character assets, non-NULL means corp division
  created_at TIMESTAMP DEFAULT NOW()
);

-- Blacklist: item type IDs to never build
CREATE TABLE arbiter_blacklist (
  id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  user_id BIGINT NOT NULL,
  type_id BIGINT NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  UNIQUE(user_id, type_id)
);

-- Whitelist: item type IDs to always build if possible
CREATE TABLE arbiter_whitelist (
  id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  user_id BIGINT NOT NULL,
  type_id BIGINT NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  UNIQUE(user_id, type_id)
);
```

### Related Tables

- `asset_item_types` — Item definitions (name, volume, group_id) used by BOM tree
- `market_prices` — Jita market data (buy/sell orders) cached via market price updater
- `character_assets` / `corporation_assets` — Asset inventory loaded by scope membership
- `sde_blueprints` — Blueprint definitions and material requirements
- `sde_reactions` — Moon reaction product definitions

## API Endpoints

### Settings & Configuration

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/arbiter/settings` | Get user's Arbiter settings (structure tiers, taxes, defaults) |
| PUT | `/v1/arbiter/settings` | Update Arbiter settings |

### Opportunities Scan

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/arbiter/opportunities?buildAll=true&systemID=30000142` | Scan T2 manufacturing opportunities (all builds or market buys, cost index filtered by system) |

### Scopes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/arbiter/scopes` | List all scopes for user |
| GET | `/v1/arbiter/scopes/:id` | Get scope details (name, members) |
| POST | `/v1/arbiter/scopes` | Create new scope |
| PUT | `/v1/arbiter/scopes/:id` | Update scope name |
| DELETE | `/v1/arbiter/scopes/:id` | Delete scope |
| GET | `/v1/arbiter/scopes/:id/members` | Get members (characters/divisions) in scope |
| POST | `/v1/arbiter/scopes/:id/members` | Add character/division to scope |
| DELETE | `/v1/arbiter/scopes/:id/members/:memberID` | Remove member from scope |

### Tax Profiles

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/arbiter/tax-profile` | Get user's tax profile (input/output price type, broker/facility fees) |
| PUT | `/v1/arbiter/tax-profile` | Update tax profile |

### Black/Whitelist

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/arbiter/blacklist` | Get items never to build |
| POST | `/v1/arbiter/blacklist/:typeID` | Add item to blacklist |
| DELETE | `/v1/arbiter/blacklist/:typeID` | Remove item from blacklist |
| GET | `/v1/arbiter/whitelist` | Get items to always build |
| POST | `/v1/arbiter/whitelist/:typeID` | Add item to whitelist |
| DELETE | `/v1/arbiter/whitelist/:typeID` | Remove item from whitelist |

### BOM Tree (Phase 2)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/arbiter/[typeID]/bom?inputPriceType=buy&scopeID=1` | Build full BOM tree for product (delta calculation, recursive materials) |
| GET | `/v1/arbiter/decryptors` | Get list of available decryptors for invention |

## Frontend Routes

- `/arbiter` — Main Arbiter page (opportunities scan, settings, scopes, tax profiles)
- `api/arbiter/[typeID]/bom.ts` — BOM tree endpoint (Next.js API route)

## File Paths

### Backend

- Controller: `internal/controllers/arbiter.go`
- Controller Tests: `internal/controllers/arbiter_test.go`
- Services: `internal/services/arbiter.go`
- Service Tests: `internal/services/arbiter_test.go`
- Repository: `internal/repositories/arbiter.go`
- Repository Tests: `internal/repositories/arbiter_test.go`
- Models: `internal/models/arbiter.go`
- Updater (Market Prices): `internal/updaters/marketPrices.go`
- Calculator (Manufacturing): `internal/calculator/manufacturing.go` — manufacturing job cost calculation
- Calculator (Cost, Time): `internal/calculator/` — various cost/time computation helpers

### Frontend

- Main Page: `frontend/packages/pages/arbiter.tsx`
- API Route: `frontend/pages/api/arbiter/[typeID]/bom.ts`

## Phase 2 Fixes

### BOM Tree Delta Calculation
- **Issue**: Leaf nodes (buyable items with no blueprint) always showed delta = 0
- **Root Cause**: The `Delta` field was not being set on leaf node creation
- **Fix**: Now explicitly computed as `delta := needed - available`, with a floor of 0
- **Impact**: Shopping lists now correctly show true deficit quantities

### Input Price Type Wiring
- **Issue**: `inputPriceType` query parameter was ignored; costs always used sell order prices
- **Root Cause**: Parameter was read by frontend but never passed to backend, and getBuyPrice was hardcoded to use sell orders
- **Fix**:
  - Frontend: `/api/arbiter/[typeID]/bom.ts` now reads `inputPriceType` from query and passes to service
  - Service: `getInputPrice()` respects the `inputPriceType` field in `bomTreeContext`
  - Fallback: If preferred type is nil, the system falls back to the other type
- **Impact**: Cost calculations now respect player's configured pricing strategy (buy vs. sell order)

### Shopping List Total Value
- **Issue**: Shopping list displayed `delta_cost` instead of `total_value`
- **Root Cause**: Cost field was calculated as `unitPrice * delta` (only the deficit), not full inventory cost
- **Fix**: Shopping list now uses `total_value = unitPrice * quantity` (full Jita cost for entire needed quantity)
- **Impact**: Players see true cost to procure all materials, not just the gap

### Market Price Update Throttle
- **Changed**: UpdateInterval in `internal/updaters/marketPrices.go` changed from 6 hours to 2 hours
- **Rationale**: More frequent updates keep Arbiter results closer to real market conditions
- **Impact**: Opportunity analysis refreshes every 2 hours instead of 6

### Reaction Batch Cost Pro-Ration Removal
- **Issue**: Reaction jobs were pro-rated cost across qty/productQtyPerRun
- **Root Cause**: Cost calculation tried to amortize batch fee across multiple output units
- **Fix**: Reactions now always charge the full batch cost, regardless of how many units are produced
- **Rationale**: EVE reactions always run full batches; there's no such thing as a partial reaction
- **Code**: `calcChainCost()` no longer applies pro-ration for reaction activities
- **Impact**: Reaction opportunity costs are now accurate

### Input Price Fallback
- **Behavior**: If `input_price_type` is "buy" but no buy orders exist for an item, the system falls back to sell order price
- **Implementation**: `getInputPrice()` and `getBuyPrice()` in bomTreeContext check the preferred type first, then try the other if nil
- **Impact**: BOM trees always have a cost, even if market data is incomplete

## Job Cost Calculation (Phase 2)

### Manufacturing Job Cost Formula

The corrected manufacturing job cost formula applies facility tax to the job fee, not directly to EIV:

```
jobCost = (eiv × costIndex × (1 - structBonus) × (1 + facilityTaxRate) + eiv × SCC) × runs
```

**Components**:
- **eiv** — Estimated Industry Value, calculated from base blueprint quantities (not ME-adjusted). EVE computes job costs using blueprint base quantities, regardless of ME level.
- **costIndex** — System cost index (0.5–2.0 in 0.5-1.5 range for most empire systems)
- **structBonus** — Structure rig bonus (0–0.2; higher rigs = lower cost). Applied as multiplication: `(1 - structBonus)`
- **facilityTaxRate** — User-configured facility tax % (0–0.2 typical). Applied to the job fee component only
- **SCC** — Sales tax + customs (0.02)
- **runs** — Number of manufacturing runs

**Old formula (incorrect)**: `(eiv × costIndex × (1 - structBonus) + eiv × facilityTaxRate + eiv × SCC) × runs` applied tax directly to EIV.

**New formula (correct)**: Tax applies to the job fee `costIndex × (1 - structBonus)` before adding SCC.

**Impact**: Manufacturing job costs are now accurate. Facility tax reduction is more meaningful with higher-tier rigs.

### BOM Node Cost Fields

`BOMNode` in `internal/models/arbiter.go` now includes:
- **job_cost** (`int64`) — Per-run manufacturing or reaction cost in ISK
- **runs** (`int`) — Number of runs needed to produce required quantity

These fields enable the frontend to display per-step manufacturing costs in the Job Costs tab.

### Job Costs Tab in BOM Panel

The BOM panel in `frontend/packages/pages/arbiter.tsx` now includes a "Job Costs" tab showing:
- **Step** — Item name being built
- **Qty** — Total quantity needed
- **Runs** — Number of runs required (at max manufacturing slots, or max reaction batch)
- **Job Fee** — Total manufacturing cost per run (in ISK)
- **Total Job Cost** — Sum of all steps' per-run costs

This gives players immediate visibility into labor costs across the full production chain.

## Configuration

### Environment Variables

None specific to Arbiter — uses standard `BACKEND_KEY`, `DATABASE_*`, and `PORT` variables.

### Feature Flags

- `use_whitelist`, `use_blacklist` in `arbiter_settings` — Per-user toggles for item filtering

## Testing

### E2E Tests
- `e2e/tests/arbiter.spec.ts` — Scan opportunities, build BOM trees, manage scopes

### Unit Tests
- `internal/controllers/arbiter_test.go` — Route handlers
- `internal/services/arbiter_test.go` — BOM tree calculation, delta logic, price type logic
- `internal/repositories/arbiter_test.go` — Settings/scope/list CRUD

## Known Limitations

1. **BOM depth limit**: Recursion stops at 10 levels. Most T2 chains fit within this; some exotic chains may not.
2. **No job slot validation**: Arbiter assumes unlimited job slots. Real players must verify they have slots available.
3. **Single-system cost index**: Each structure uses one system's cost index. Multi-system chains (e.g., reaction in A, invention in B) compute each stage separately.
4. **Static decryptor**: The decryptor type is fixed per user (via `decryptor_type_id`). Invention chains always assume this decryptor.

## Related Features

- **Job Manager** — Executes the builds Arbiter recommends; tracks actual job outcomes
- **Stockpile Tracking** — Tracks material inventories that feed into delta calculations
- **Market Pricing** — Provides the Jita market data Arbiter uses for cost/revenue calculations
- **Background Asset Updates** — Refreshes scope assets hourly so BOM trees have current inventory

## Open Questions

- [ ] Should sub-buildings (materials with blueprints) display in BOM tree as toggles for "build vs. buy"?
- [ ] Should Arbiter export shopping lists to Excel or other formats?
- [ ] Should cost index variations across regions be supported (not just primary system)?
