# Hauling Runs — JF Logistics Feature Spec

## Overview

Hauling Runs is the connective tissue that unifies the market scanner, transportation, stockpile, and sell-order tracking features into a single end-to-end JF logistics workflow.

The goal: **identify what to haul, plan the run, acquire the cargo, haul it, sell it, and track the actual profit** — all without leaving pinky.tools.

---

## Problem Statement

Currently pinky.tools has separate tools for:
- Transportation cost calculation
- Jita market pricing
- Stockpile multibuy
- Auto-sell containers / contract sync

These are valuable individually but disconnected. A JF operator has to mentally stitch them together: check prices somewhere, calculate margins in a spreadsheet, generate a buy list, haul, then manually track what sold and what didn't.

Hauling Runs replaces that workflow with a single orchestrated flow.

---

## Core Concepts

### Market Scanner

Before a run exists, the operator scans the market to find opportunities. Two modes:

**Hub-to-Hub Arbitrage**
Find items with a price spread between trade hubs (Jita, Amarr, Dodixie, Rens, Hek) that exceed haul cost + fees.

**System Supply Scanner**
Given a null/low-sec target system (e.g. BWF-ZZ), identify:
- **Supply gaps** — items with buy orders but zero or thin sell orders
- **Markup opportunities** — items where local sell price / Jita price exceeds threshold

ESI supports system-level market filtering: `/markets/{region_id}/orders/?system_id={system_id}`

### Hauling Run

The central object. Tracks a haul from acquisition through final sale.

**Statuses:**
```
PLANNING     → items selected, math done, not yet acquiring
ACCUMULATING → buy orders placed at source, waiting on fills
READY        → cargo threshold met, waiting on operator to load up
IN_TRANSIT   → JF is moving
SELLING      → sell orders up at destination, tracking fills
COMPLETE     → all items sold, final P&L recorded
CANCELLED    → abandoned
```

### Cargo Items

Each item in a run has:
- `type_id`, `quantity`, `m3`
- `acquisition_mode`: `instant_buy` | `buy_order`
- `buy_price` (target or actual)
- `target_sell_price`
- `source`: `scanner` | `stockpile` | `manual`
- Linked ESI order IDs (buy-side and sell-side)

---

## Feature Specification

### 1. Market Scanner

**Inputs (filter panel):**
- Source hub / system
- Destination system (supports null-sec system IDs, not just trade hubs)
- Transport route (links to existing TransportProfile)
- Min markup multiplier (e.g. 1.3x)
- Min net profit per unit
- Max total m³ (JF capacity)
- Min daily volume at destination (liquidity filter — critical)
- Item category filter (ships, modules, ammo, fuel blocks, etc.)
- Acquisition mode: instant buy vs. buy order pricing

**Output (results table):**

| Item | Source Price | Dest Sell | Dest Buy Orders | Spread | m³/slot | Haul Cost | Net Profit | Daily Vol | Days to Sell |
|------|-------------|-----------|-----------------|--------|---------|-----------|------------|-----------|--------------|
| Void M | 90 | 450 | 15.2M ISK | 5.0x | 0.005 | — | 175M | 4.2M/day | <1 |
| CN Raven | 410M | — | 3 @ 480M | gap 🔴 | 900 | 12M | 58M | 0.4/day | ~8 |

**Gap indicator:**
- 🔴 = buy orders exist, zero sell orders → sale guaranteed at buy order price
- ✅ = sell orders exist but marked up → competitive opportunity
- ⚠️ = thin sell orders → partial fill risk

**Actions from results:**
- Select items → **Add to Hauling Run** (new or existing planned run)
- **Fill remaining capacity** — re-run scanner filtered to remaining m³ in current run

**Data freshness:**
- ESI market data polled in background, ~1 hour refresh cycle (same pattern as asset refresh)
- Market history (`/markets/{region_id}/history/`) used for daily volume / days-to-sell estimates
- Scanner shows data age; operator can trigger manual refresh

---

### 2. Hauling Run Planner

**Capacity view:**
```
Route: Jita IV-4 → BWF-ZZ  (Delve)
JF Capacity: 300,000 m³   Used: 5,240 m³ (2%)  ██░░░░░░░░

 Item                 │  Qty    │   m³   │ Acq Mode    │ Buy Cost  │ Sell Est  │ Profit
 ─────────────────────┼─────────┼────────┼─────────────┼───────────┼───────────┼────────
 Void M               │ 500,000 │  2,500 │ Buy Order   │  44M      │  220M     │  176M
 CN Raven             │       3 │  2,700 │ Instant Buy │ 1.23B     │  1.44B    │  210M
 Pithum C-Type Adap   │       2 │     40 │ Buy Order   │  2.4B     │  3.8B     │  1.4B
 ─────────────────────┼─────────┼────────┼─────────────┼───────────┼───────────┼────────
 Totals               │         │  5,240 │             │  3.67B    │  5.46B    │  1.79B

 Haul Cost (est):   180M    (from TransportProfile: Jita → BWF @ 600 ISK/m³ + 0.5% collateral)
 Net Profit:        1.61B
 Margin:            30.5%
 ISK/m³:            307,000
```

**Per-item acquisition settings (buy order mode):**
- Target buy price (default: current Jita buy order - configurable % below)
- Min haul threshold (don't haul until X% of this item is filled)
- Max wait days (auto-upgrade to instant buy after N days)

**Run-level settings:**
- Min overall fill % before haul (e.g. 80%)
- Alert me when ready

---

### 3. Acquisition Tracking (Buy Orders in Jita)

When items are set to `buy_order` mode:

1. Operator places buy orders in Jita in-game
2. Next ESI poll (`/characters/{id}/orders/`) picks them up
3. Tool matches by type_id + location_id + timestamp → links to cargo item
4. Auto-match confidence shown; operator confirms linkage
5. `volume_remain` tracked on each poll; fill progress displayed

**Run status during accumulation:**
```
Run #47 — BWF Haul               Status: ACCUMULATING  ◐
Haul readiness: 71%  ███████░░░  (threshold: 80%)

 Item              │ Mode        │ Filled      │ Wait   │ Status
 ─────────────────────────────────────────────────────────────────
 Void M            │ Buy Order   │ 312k/500k   │ day 2  │ ⏳ filling
 CN Raven          │ Instant Buy │ 3/3         │ —      │ ✅ ready
 Pithum C-Type     │ Buy Order   │ 2/2         │ day 1  │ ✅ filled

[ Haul Now (71%) ]  [ Upgrade Stragglers to Instant Buy ]  [ Wait ]
```

Discord alert when haul threshold is met.

---

### 4. Multibuy Export

At any point in PLANNING or ACCUMULATING status:

**"Generate Multibuy"** exports:
- Instant buy items → standard EVE multibuy format (same as existing stockpile multibuy)
- Buy order items → separate list with target prices annotated

---

### 5. In-Transit Tracking

Operator marks run **IN_TRANSIT** when loading up.

Discord notification to configured logistics channel:
> 🚀 **Hauling Run #47** is underway — Jita → BWF
> Cargo: 5,240 m³ | Projected profit: 1.61B ISK

---

### 6. Sell-Side Order Tracking

Operator arrives in BWF, places sell orders in-game.

Next ESI poll picks up new orders at BWF. Tool matches to run by:
- `type_id` matches cargo item
- `location_id` matches destination system
- `issued` timestamp after run start

Operator confirms matches (or manually links order IDs).

**Active selling view:**
```
Run #47 — BWF Haul               Status: SELLING  ●
────────────────────────────────────────────────────
 Item              │ Qty     │ Filled      │ Revenue
 ──────────────────┼─────────┼─────────────┼──────────
 Void M            │ 500,000 │ ████████░░  │ +140M
 CN Raven          │ 3       │ ██████████  │ +210M ✓
 Pithum C-Type     │ 2       │ ░░░░░░░░░░  │ listed

Realized so far:  350M ISK
Pending (est):    ~1.26B ISK
```

Discord ping per item when sold. Final ping when run completes.

---

### 7. Run Completion & P&L

When all orders fill (or operator manually closes), run moves to COMPLETE.

**Post-run summary:**
```
Run #47 — BWF Haul                              COMPLETE ✓
──────────────────────────────────────────────────────────
Route:          Jita → BWF-ZZ
Duration:       Feb 28 → Mar 4 (5 days)
Cargo:          5,240 m³

                   Projected      Actual
Buy Cost           3.67B          3.61B  (buy orders saved 60M)
Haul Cost          180M           180M
Revenue            5.46B          5.22B  (Pithum moved slower)
Net Profit         1.61B          1.43B
Margin             30.5%          27.4%
ISK/m³             307,000        273,000
──────────────────────────────────────────────────────────
Buy order savings vs instant buy:  +60M ISK
Market movement vs scan estimate:  -240M ISK
```

Actual revenue pulled from wallet journal transactions for accuracy.

---

### 8. Run History & Analytics

- Profit per route over time
- Best-performing item categories
- Buy order wait time vs. savings tradeoff
- Average fill time per item type
- Total ISK earned by the logistics operation

---

## Data Model

### `hauling_runs`
```sql
id                  SERIAL PRIMARY KEY
name                TEXT
route_id            INT REFERENCES transport_profiles(id)
status              TEXT  -- planning|accumulating|ready|in_transit|selling|complete|cancelled
haul_threshold_pct  INT   -- min fill % before operator hauls (default 80)
created_at          TIMESTAMPTZ
updated_at          TIMESTAMPTZ
in_transit_at       TIMESTAMPTZ
completed_at        TIMESTAMPTZ
notes               TEXT
```

### `hauling_run_items`
```sql
id                  SERIAL PRIMARY KEY
run_id              INT REFERENCES hauling_runs(id)
type_id             INT
quantity            BIGINT
m3_per_unit         NUMERIC
acquisition_mode    TEXT  -- instant_buy|buy_order
target_buy_price    NUMERIC
actual_buy_price    NUMERIC
target_sell_price   NUMERIC
actual_sell_price   NUMERIC
source              TEXT  -- scanner|stockpile|manual
buy_order_id        BIGINT  -- ESI order ID (buy side)
sell_order_id       BIGINT  -- ESI order ID (sell side)
qty_acquired        BIGINT
qty_sold            BIGINT
max_wait_days       INT
status              TEXT  -- pending|acquiring|acquired|listed|sold
```

### Market cache tables (new)
```sql
market_orders_cache   -- regional ESI order snapshots (type_id, region_id, system_id, price, volume, is_buy_order, ts)
market_history_cache  -- daily volume history per type per region (type_id, region_id, date, volume, avg_price)
```

---

## ESI Endpoints Required

| Endpoint | Purpose | Already Used? |
|----------|---------|--------------|
| `/markets/{region_id}/orders/` | Regional market orders (scanner) | No |
| `/markets/{region_id}/orders/?system_id={id}` | System-filtered orders (null-sec scanner) | No |
| `/markets/{region_id}/history/` | Daily volume history | No |
| `/characters/{id}/orders/` | Active buy/sell orders (acquisition + sell tracking) | No |
| `/characters/{id}/orders/history/` | Completed orders | No |
| `/characters/{id}/wallet/journal/` | Actual revenue for P&L | No |

---

## Integration Points (Existing Features)

| Feature | Integration |
|---------|------------|
| TransportProfile | Route selection, haul cost calculation |
| Jita market pricing | Source price baseline for scanner |
| Stockpile multibuy | Multibuy export format reused |
| Auto-sell containers | Optional: auto-list items on arrival |
| Contract sync | Alternative sell tracking for contract-based sales |
| Discord notifications | Accumulation ready, transit start, item sold, run complete |

---

## MVP Scope

### Phase 1 — Core Loop
- [ ] ESI market order cache (regional + system-filtered polling, ~1h refresh)
- [ ] ESI market history cache (daily volume for liquidity filtering)
- [ ] System Supply Scanner UI (gap 🔴 + markup ✅ analysis)
- [ ] Hub-to-Hub Arbitrage Scanner
- [ ] Hauling Run object (CRUD, status machine)
- [ ] Add items from scanner to run
- [ ] Capacity tracker (m³ used / available)
- [ ] Multibuy export (instant buy items)
- [ ] Manual status transitions

### Phase 2 — Acquisition Tracking
- [ ] Buy order ESI polling + run linkage (auto-match by type + location + time)
- [ ] Acquisition progress UI per item
- [ ] Haul threshold Discord alerts
- [ ] Upgrade-to-instant-buy action
- [ ] Multibuy export with buy order target prices

### Phase 3 — Sell Tracking & P&L
- [ ] Sell order ESI polling + run linkage
- [ ] Sell progress UI (volume_remain tracking)
- [ ] Wallet journal integration (actual revenue)
- [ ] Post-run P&L summary (projected vs. actual)
- [ ] Discord sell notifications (per item + run complete)

### Phase 4 — Analytics
- [ ] Run history view
- [ ] Per-route and per-item performance analytics
- [ ] Buy order wait time vs. savings analysis
- [ ] Scanner recommendations informed by historical run data

---

## Open Questions

1. **Corporation vs. character orders** — Should hauling runs support corp-wallet buy orders placed by a designated logistics character?
2. **Multiple JF operators** — Can a run be split across operators hauling different portions?
3. **Sell order undercut alerts** — Should the tool notify when a competitor lists below your sell price at destination?
4. **Route security** — For null/low-sec routes, surface Dotlan traffic/kill data as a safety indicator?
5. **Third-party JF services** — Integration with PushX/Red Frog collateral calculation for outsourced hauling?
