# EVE Online Reactions Calculator — Implementation Specification

This document fully specifies the Reactions Calculator tool for EVE Online moon reaction chains. It is intended as instructions for recreating the tool in a new codebase.

---

## 1. Overview

The Reactions Calculator helps players plan and optimize moon reaction chains. It calculates:

- Which intermediate (simple) reactions to run, and how many slots/runs each needs
- Shopping lists of raw moon goo and fuel blocks needed
- Production quantities, investment costs, revenue, and profit per cycle
- Multibuy-compatible text export for in-game purchasing

### Key Design Principles

- **Only complex reactions are selectable** — intermediates are ALWAYS built from raw materials, never bought
- **Everything runs simultaneously** — intermediates and complex reactions run in parallel for the full cycle
- **Shared intermediates are consolidated** — when multiple complex reactions use the same intermediate, demand is aggregated into shared slots
- **Batch ME** — EVE applies material efficiency to the total batch, not per-run. This affects all quantity calculations

---

## 2. EVE Reaction Chain

```
Raw Moon Goo → Simple/Intermediate Reactions → Complex Reactions
```

### Reaction Groups (from `sde_groups`)

| Group Name | Role | Selectable? |
|---|---|---|
| `Intermediate Materials` | Simple reactions, produce intermediates | No (auto-calculated) |
| `Unrefined Mineral` | Simple reactions (alchemy) | No (auto-calculated) |
| `Composite` | Complex composite reactions | Yes |
| `Biochemical Material` | Complex biochem reactions | Yes |
| `Hybrid Polymers` | Complex polymer reactions | Yes |
| `Molecular-Forged Materials` | Complex molecular reactions | Yes |

### Reaction Data (from SDE)

- **Base time**: All standard reactions have `time = 10800` seconds (3 hours) base
- **Product quantity**: Varies by reaction (e.g., intermediates produce 200/run, composites produce varying amounts like 400, 10000, etc.)
- **Materials**: Each reaction has input materials with base quantities. Some inputs are themselves reaction products (intermediates)

---

## 3. Core Math

### 3.1 ME Factor (Material Efficiency)

```
me_factor = 1.0 - rig_me_value × security_multiplier
```

Constants:
```
RIG_ME:        none=0, t1=0.02, t2=0.024
SECURITY_MULT: null=1.1, low=1.0, high=1.0
```

Example (T2 rig in nullsec): `me_factor = 1.0 - 0.024 × 1.1 = 0.9736`

### 3.2 TE Factor (Time Efficiency)

```
te_factor = (1.0 - skill × 0.04) × (1.0 - structure_te) × (1.0 - rig_te_value × security_multiplier)
```

Constants:
```
RIG_TE:       none=0, t1=0.20, t2=0.24
STRUCTURE_TE: athanor=0, tatara=0.25
```

Example (Reactions 5, T2 Tatara in null):
```
te_factor = (1.0 - 5×0.04) × (1.0 - 0.25) × (1.0 - 0.24×1.1)
          = 0.8 × 0.75 × 0.736
          = 0.4416
```

### 3.3 Time Per Run and Runs Per Cycle

```
secs_per_run   = floor(base_time × te_factor)
runs_per_cycle = floor(cycle_seconds / secs_per_run)
```

Where `cycle_seconds = cycle_days × 86400`.

Example (7-day cycle, base_time=10800):
```
secs_per_run   = floor(10800 × 0.4416) = floor(4769.28) = 4769
runs_per_cycle = floor(604800 / 4769) = floor(126.82) = 126
```

### 3.4 Per-Run ME (for API display fields)

Per-run adjusted quantity of each input material:
```
adj_qty = ceil(base_qty × me_factor)
```

Example: `ceil(100 × 0.9736) = ceil(97.36) = 98`

### 3.5 Batch ME (for actual quantities — shopping list and demand calculations)

**CRITICAL**: EVE applies ME to the total batch across all runs, not per-run then multiply. This gives slightly different (lower) totals than per-run ME.

```
batch_qty(runs, base_qty) = max(runs, ceil(runs × base_qty × me_factor))
```

Example: `batch_qty(126, 100) = max(126, ceil(126 × 100 × 0.9736)) = max(126, ceil(12267.36)) = 12268`

Compare per-run: `adj_qty × runs = 98 × 126 = 12348` (higher by 80)

**Batch ME is used for:**
- Intermediate demand calculation (how much of each intermediate is consumed)
- Shopping list quantities (raw materials needed)

**Per-run ME is used for:**
- Display in the picker table (`adj_qty`)
- Per-run cost calculations in the API (`input_cost_per_run`)

### 3.6 Complex Instances (Slot Ratio)

Each complex reaction has a `complex_instances` value: how many parallel complex reaction lines one cycle of intermediate production can feed.

```python
for each intermediate material of the complex reaction:
    ratio = simple_reaction.product_quantity / complex_reaction.adj_qty_of_that_intermediate

complex_instances = floor(min(all ratios))
# Minimum 1
```

Example: Simple produces 200/run, complex needs 98/run → `floor(200/98) = floor(2.04) = 2`

This means **1 "instance"** of a complex reaction = `complex_instances` parallel complex reaction lines + the shared intermediate slots to feed them.

### 3.7 Intermediate Demand Aggregation

When calculating the plan, intermediate demand is aggregated across ALL selected complex reactions:

```javascript
for each selected complex reaction:
    complexLines = instances × complex_instances
    for each intermediate material:
        demand[material].total_qty += batchQty(runs_per_cycle, base_qty) × complexLines
```

Note: Uses `batchQty()` (batch ME), NOT `adj_qty × runs` (per-run ME). This is critical for getting the correct intermediate run counts.

### 3.8 Intermediate Slots and Runs

```
supply_per_slot = product_qty_per_run × runs_per_cycle
slots_needed    = ceil(total_demand / supply_per_slot)
runs_needed     = ceil(total_demand / (slots_needed × product_qty_per_run))
```

The `runs_needed` formula naturally produces the correct number for any cycle length:
- 7-day cycle: 123 runs for standard intermediates (vs 126 max)
- 14-day cycle: 247 runs (vs 253 max)
- The gap scales proportionally — no hardcoded buffer needed

**DO NOT** use `runs_per_cycle - 3` or any fixed buffer. The demand-driven `ceil()` formula handles all cycle lengths correctly because batch ME produces slightly less demand than full-cycle supply.

### 3.9 Job Cost

```
base_job_cost = sum(adj_qty × adjusted_price) for each input material
job_cost      = base_job_cost × system_cost_index × (1.0 + facility_tax / 100)
```

- `adjusted_price` comes from CCP's `market_prices` table (ESI endpoint), NOT from market orders
- `system_cost_index` comes from `industry_cost_indices` table for the `'reaction'` activity
- `facility_tax` is a user setting (make default 0%)

### 3.10 Output Value and Fees

```
output_value = product_qty_per_run × unit_price
output_fees  = output_value × (broker_fee + sales_tax) / 100
```

### 3.11 Shipping

```
shipping_in  = input_volume × shipping_per_m3 + input_value × shipping_collateral
shipping_out = output_volume × shipping_per_m3 + output_value × shipping_collateral
```

Both are optional (toggled by user settings).

### 3.12 Profit (per-run, API level)

```
total_cost     = input_cost + job_cost + shipping_in + shipping_out
profit_per_run = output_value - output_fees - total_cost
```

### 3.13 Plan Summary Financials

**KNOWN ISSUE**: The current plan summary calculates investment using `input_cost_per_run` from the API, which prices intermediate materials at their MARKET value. Since intermediates are built from raw moon goo, the real investment cost is lower. A more accurate approach would compute investment from the shopping list (raw material costs) + all job costs (both intermediate and complex) + shipping + output fees. This is a TODO.

Current implementation:
```javascript
// Per complex reaction, per cycle:
costPerRun = input_cost_per_run + job_cost_per_run + shipping_in + shipping_out + output_fees
investment = costPerRun × runs_per_cycle × complexLines
revenue    = output_value_per_run × runs_per_cycle × complexLines
profit     = revenue - investment
```

---

## 4. Shopping List Calculation

The shopping list resolves ALL intermediates down to raw moon goo + fuel blocks.

```javascript
// 1. Raw materials from intermediate reactions
for each intermediate:
    for each input material of the simple reaction:
        qty_per_slot = batchQty(intermediate.runs, material.base_qty)
        total_qty    = qty_per_slot × intermediate.slots
        add to shopping list

// 2. Non-intermediate materials from complex reactions (e.g., fuel blocks)
for each selected complex reaction:
    complexLines = instances × complex_instances
    for each non-intermediate material:
        qty_per_line = batchQty(runs_per_cycle, material.base_qty)
        total_qty    = qty_per_line × complexLines
        add to shopping list
```

### Multibuy Export

Format: `Material Name QTY` (one per line), sorted alphabetically. Copied to clipboard for pasting into EVE's multibuy window.

---

## 5. Database Schema (relevant tables)

```sql
-- Static data from EVE SDE
sde_reactions (reaction_type_id PK, product_type_id, product_quantity, time)
sde_reaction_materials (reaction_type_id + material_type_id PK, quantity)
sde_types (type_id PK, name, group_id, volume, packaged_volume)
sde_groups (group_id PK, name, category_id)

-- Market data (refreshed periodically)
market_orders (order_id PK, type_id, location_id, is_buy_order, price, ...)
market_prices (type_id PK, adjusted_price, average_price)  -- CCP global prices

-- Industry indices
industry_cost_indices (system_id + activity PK, cost_index)
```

### Key Queries

**Reactions + products:**
```sql
SELECT r.reaction_type_id, r.product_type_id, r.product_quantity, r.time,
       pt.name AS product_name,
       COALESCE(pt.packaged_volume, pt.volume, 0) AS product_volume,
       g.name AS group_name
FROM sde_reactions r
JOIN sde_types pt ON pt.type_id = r.product_type_id
JOIN sde_groups g ON g.group_id = pt.group_id
WHERE r.time >= 3600  -- exclude deprecated short-cycle reactions
ORDER BY g.name, pt.name
```

**Jita prices (from specific station):**
```sql
SELECT type_id,
       MIN(price) FILTER (WHERE NOT is_buy_order) AS sell_min,
       MAX(price) FILTER (WHERE is_buy_order) AS buy_max
FROM market_orders
WHERE location_id = 60003760  -- Jita 4-4
GROUP BY type_id
```

**CCP adjusted prices (for job cost base):**
```sql
SELECT type_id, adjusted_price FROM market_prices
```

**System cost indices:**
```sql
SELECT cost_index FROM industry_cost_indices
WHERE system_id = ? AND activity = 'reaction'
```

**Systems with reaction indices (for picker):**
```sql
SELECT i.system_id, s.name, s.security_status, i.cost_index
FROM industry_cost_indices i
JOIN sde_systems s USING (system_id)
WHERE i.activity = 'reaction'
ORDER BY s.name
```

### Identifying Intermediates

A material is an "intermediate" if its `type_id` appears as a `product_type_id` in the `sde_reactions` table:
```python
reaction_product_ids = {r.product_type_id for r in all_reactions}
is_intermediate = (material_type_id in reaction_product_ids)
```

---

## 6. API Endpoint

`GET /api/reactions`

### Query Parameters

| Parameter | Default | Description |
|---|---|---|
| `system_id` | (none) | Solar system for cost index |
| `structure` | `tatara` | `tatara` or `athanor` |
| `rig` | `t2` | `none`, `t1`, or `t2` |
| `security` | `null` | `null`, `low`, or `high` |
| `reactions_skill` | `5` | 0-5 |
| `facility_tax` | `0.25` | % |
| `cycle_days` | `7` | 1-30 |
| `broker_fee` | `3.5` | % |
| `sales_tax` | `2.25` | % |
| `shipping_m3` | `0` | ISK per m³ |
| `shipping_collateral` | `0` | fraction (0-1) |
| `input_price` | `sell` | `sell`, `buy`, or `split` |
| `output_price` | `sell` | `sell` or `buy` |
| `ship_inputs` | `1` | `1` or `0` |
| `ship_outputs` | `1` | `1` or `0` |

### Response

```json
{
  "reactions": [
    {
      "reaction_type_id": 12345,
      "product_type_id": 67890,
      "product_name": "Fermionic Condensates",
      "group_name": "Composite",
      "product_qty_per_run": 400,
      "runs_per_cycle": 126,
      "secs_per_run": 4769,
      "complex_instances": 2,
      "num_intermediates": 4,
      "input_cost_per_run": 13047980.00,
      "job_cost_per_run": 301342.19,
      "output_value_per_run": 12616000.00,
      "output_fees_per_run": 725420.00,
      "shipping_in_per_run": 0.00,
      "shipping_out_per_run": 0.00,
      "profit_per_run": -1458742.19,
      "profit_per_cycle": -183801515.94,
      "margin": -11.18,
      "materials": [
        {
          "type_id": 11111,
          "name": "Caesarium Cadmide",
          "base_qty": 100,
          "adj_qty": 98,
          "price": 132160.0,
          "cost": 12951680.00,
          "volume": 980.00,
          "is_intermediate": true
        }
      ]
    }
  ],
  "count": 119,
  "cost_index": 0.0638,
  "me_factor": 0.9736,
  "te_factor": 0.4416,
  "runs_per_cycle_base": 126
}
```

The response includes ALL reactions (simple + complex). The frontend filters which are selectable.

---

## 7. Frontend Architecture

### Tabs

1. **Pick Reactions** — Table of all complex reactions with instance count inputs, sortable columns, expandable material details
2. **Shopping List** — Aggregated raw materials with quantities, prices, costs, volumes. Includes "Copy Multibuy" button
3. **Plan Summary** — Overview cards (slots, investment, revenue, profit) + detailed table of intermediate and complex reactions with slots, runs, produced quantities, and financials

### State

```javascript
reactions = [];           // All reactions from API (simple + complex)
meFactor = 1.0;           // From API response
selections = {};          // { reaction_type_id: instance_count }
```

Settings and selections are persisted in `localStorage`.

### Core Functions

- **`computePlan()`** — Shared by plan, shopping, and multibuy. Aggregates intermediate demand, calculates slots/runs, returns the full plan object
- **`batchQty(runs, baseQty)`** — `max(runs, ceil(runs × baseQty × meFactor))` — EVE batch ME formula
- **`collectShoppingMats()`** — Uses `computePlan()` to derive raw material quantities with batch ME
- **`renderPlan()`** — Renders plan summary cards and table
- **`renderShopping()`** — Renders shopping list table
- **`renderPicker()`** — Renders reaction picker with instance inputs

### Filtering

- `SIMPLE_GROUPS = ['Intermediate Materials', 'Unrefined Mineral']` — excluded from picker
- Group filter dropdown (Composite, Biochemical, Hybrid Polymers, Molecular-Forged)
- Text search on product name
- These filters only affect the picker display, NOT the plan or shopping calculations

### Settings Modal

Additional settings in a popup modal:
- Reactions Skill Level (0-5)
- Facility Tax %
- Broker Fee %
- Sales Tax %
- Input Price (Jita Sell / Jita Buy / Split)
- Output Price (Jita Sell / Jita Buy)
- Ship Inputs? (Yes/No)
- Ship Outputs? (Yes/No)
- Shipping ISK/m³
- Collateral %
- Total Reaction Slots (for the slot usage indicator)

### Toolbar Settings (always visible)

- System (dropdown populated from `/api/reaction-systems`)
- Structure (Tatara / Athanor)
- Rig (T2 / T1 / None)
- Security (Null/WH / Lowsec / Highsec)
- Cycle Days (number input)
- Group filter
- Search

All settings changes trigger an API reload. Filter/search changes are client-side only.

---

## 8. Verification Data

For **1 instance of each of the 17 Composite reactions** in a **T2-rigged Tatara in nullsec (BWF-ZZ)**, **7-day cycle**:

- **ME factor**: 0.9736
- **TE factor**: 0.4416
- **Runs per cycle**: 126 (complex), 123 (standard intermediates), 26 (Oxy-Organic Solvents)
- **Total slots**: 74 (42 intermediate + 32 complex)
- **24 unique intermediates**, **20 raw moon goo types**

### Expected Intermediate Slots and Runs (7-day)

| Intermediate | Slots | Runs |
|---|---|---|
| Caesarium Cadmide | 2 | 123 |
| Carbon Fiber | 1 | 123 |
| Carbon Polymers | 3 | 123 |
| Ceramic Powder | 2 | 123 |
| Crystallite Alloy | 2 | 123 |
| Dysporite | 2 | 123 |
| Fernite Alloy | 2 | 123 |
| Ferrofluid | 2 | 123 |
| Fluxed Condensates | 1 | 123 |
| Hexite | 2 | 123 |
| Hyperflurite | 1 | 123 |
| Neo Mercurite | 2 | 123 |
| Oxy-Organic Solvents | 1 | 26 |
| Platinum Technite | 2 | 123 |
| Promethium Mercurite | 1 | 123 |
| Prometium | 2 | 123 |
| Rolled Tungsten Alloy | 2 | 123 |
| Silicon Diborite | 2 | 123 |
| Solerium | 1 | 123 |
| Sulfuric Acid | 3 | 123 |
| Thermosetting Polymer | 1 | 123 |
| Thulium Hafnite | 1 | 123 |
| Titanium Chromide | 2 | 123 |
| Vanadium Hafnite | 2 | 123 |

### 14-day Cycle

Same setup: intermediate runs = **247**, complex runs = **253**. The demand-driven formula produces the correct values for any cycle length.

---

## 9. Known Issues / TODOs

1. **Plan summary investment is overstated**: `input_cost_per_run` prices intermediates at market value, but they're built from raw materials. The true investment = shopping list cost (raw mats) + all job costs (intermediate + complex) + shipping + output fees. This should be rewritten to derive investment from `collectShoppingMats()` + aggregated job costs.

2. **Per-run vs batch ME in API fields**: The API's `input_cost_per_run` and `adj_qty` use per-run ceiling (`ceil(base_qty × me_factor)`), while actual EVE consumption uses batch ME. This is fine for display but should not be used for totaling across runs.

3. **Price methods**: `sell` = Jita sell minimum, `buy` = Jita buy maximum, `split` = average of sell and buy. Prices come from the `market_orders` table filtered to Jita station (location_id = 60003760).