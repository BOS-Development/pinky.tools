# Auto-Fulfill: Match Buy Orders to For-Sale Listings

## Overview

Automatically creates `purchase_transactions` (status: pending) when active buy orders overlap with active for-sale listings on price, type, and mutual contact permissions. Works for **both manual and auto-buy orders** — any active buy order participates in matching.

## Status

- **Phase**: Implementation
- **Branch**: `feature/auto-fulfill`

## Key Decisions

1. **All buy orders participate** — Both manual buy orders and auto-buy-generated buy orders are matched. The `GetAllActiveBuyOrders()` query doesn't filter by `auto_buy_config_id`.
2. **Price range on buy orders** — Buy orders now have `min_price_per_unit` + `max_price_per_unit` instead of just max. Auto-buy configs have corresponding `min_price_percentage` + `max_price_percentage`.
3. **Mutual permissions required** — Both seller→buyer AND buyer→seller must have `for_sale_browse` contact permissions for a match.
4. **Seller only sees min price** — The `GetDemandForSeller()` endpoint only exposes `min_price_per_unit` to protect the buyer's max willingness-to-pay.
5. **Dedup via unique index** — `idx_auto_fulfill_unique` on `(buy_order_id, for_sale_item_id)` where `is_auto_fulfilled = true AND status = 'pending'` prevents duplicate pending purchases.
6. **Runs last in sync chain** — Order: auto-sell → auto-buy → auto-fulfill. Ensures prices and orders are current before matching.

## Schema Changes

### `auto_buy_configs` (MODIFIED)
- Renamed `price_percentage` → `max_price_percentage`
- Added `min_price_percentage` (default 0.00)

### `buy_orders` (MODIFIED)
- Added `min_price_per_unit` (default 0.00)

### `purchase_transactions` (MODIFIED)
- Added `buy_order_id` (nullable FK → `buy_orders`)
- Added `is_auto_fulfilled` (boolean, default false)
- Added unique index `idx_auto_fulfill_unique` on `(buy_order_id, for_sale_item_id)` where `is_auto_fulfilled = true AND status = 'pending'`

## Sync Logic

### Matching Algorithm (`matchBuyOrder`)
1. Find for-sale items matching `type_id` with `price_per_unit` between `min_price_per_unit` and `max_price_per_unit`, excluding buyer's own items
2. For each match, check mutual `for_sale_browse` permissions (seller→buyer AND buyer→seller)
3. Compute quantity: `min(order.QuantityDesired, item.QuantityAvailable)`
4. Atomically (in a transaction):
   - Reduce for-sale item quantity
   - Create `purchase_transaction` with `buy_order_id` set, `is_auto_fulfilled = true`, status `pending`
5. Send Discord notification (non-blocking)
6. Continue to next match if order still has unfulfilled quantity

### Trigger Points
- **After asset refresh** (`SyncForUser`) — runs after auto-sell + auto-buy
- **After market price update** (`SyncForAllUsers`) — runs after auto-sell + auto-buy
- **After manual buy order create/update** — triggered from buy orders controller
- **After auto-buy config create/update** — triggered from auto-buy configs controller

## File Structure

| File | Change |
|------|--------|
| `internal/database/migrations/*_auto_buy_price_range.*` | New migration |
| `internal/database/migrations/*_auto_fulfill_purchases.*` | New migration |
| `internal/models/models.go` | Updated AutoBuyConfig, BuyOrder, PurchaseTransaction |
| `internal/repositories/autoBuyConfigs.go` | Renamed + added columns |
| `internal/repositories/buyOrders.go` | Added min_price, GetAllActiveBuyOrders, GetActiveBuyOrdersForUser |
| `internal/repositories/purchaseTransactions.go` | Added buy_order_id, is_auto_fulfilled, CreateAutoFulfill |
| `internal/repositories/forSaleItems.go` | Added GetMatchingForSaleItems |
| `internal/updaters/autoFulfill.go` | **New** — matching engine |
| `internal/updaters/autoBuy.go` | Compute min+max prices from config |
| `internal/updaters/assets.go` | Added WithAutoFulfillUpdater hook |
| `internal/updaters/marketPrices.go` | Added WithAutoFulfillUpdater hook |
| `internal/controllers/autoBuyConfigs.go` | Accept min/max percentages, trigger fulfill |
| `internal/controllers/buyOrders.go` | Accept minPricePerUnit, trigger fulfill |
| `cmd/industry-tool/cmd/root.go` | Wire auto-fulfill updater |
| `frontend/.../assets/AssetsList.tsx` | Auto-buy dialog: two percentage inputs (min + max) |
| `frontend/.../marketplace/BuyOrders.tsx` | Buy order form: min/max price range, table shows range |
| `frontend/.../marketplace/DemandViewer.tsx` | Seller view: shows only min price (floor), not max |
| `frontend/.../marketplace/PurchaseHistory.tsx` | "Auto" badge on auto-fulfilled transactions |
| `frontend/.../marketplace/PendingSales.tsx` | "Auto" badge on auto-fulfilled pending sales |
