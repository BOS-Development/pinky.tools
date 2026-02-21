# Contract Created Notification

## Overview

Sends a Discord notification to the **buyer** when a seller marks a purchase as "contract created". This lets buyers know a contract is ready for them to accept in-game without having to manually check the UI.

## Status: Implemented

## Event Type

- **Event key**: `contract_created`
- **Recipient**: Buyer (`purchase.BuyerUserID`)
- **Trigger**: Seller calls `POST /v1/purchases/{id}/mark-contract-created`

## Discord Embed

- **Title**: "Contract Created"
- **Description**: "{SellerName} has created a contract for your purchase"
- **Color**: #3b82f6 (primary blue)
- **Fields**: Item, Quantity, Total, Location, Contract Key (if set)

## Key Decisions

1. **Non-blocking**: Notification is sent in a goroutine — never fails the contract creation
2. **Uses existing infrastructure**: Same Discord notification system as `purchase_created` and `pi_stall`
3. **Opt-in per target**: Buyers must enable `contract_created` in their notification preferences
4. **SellerName added to model**: `PurchaseTransaction.SellerName` populated at notification time (not stored in DB)

## File Structure

| File | Role |
|------|------|
| `internal/updaters/notifications.go` | `NotifyContractCreated` method + embed builder |
| `internal/models/models.go` | `SellerName` field on `PurchaseTransaction` |
| `internal/controllers/purchases.go` | Triggers notification in `MarkContractCreated` |
| `cmd/industry-tool/cmd/root.go` | Wires `contractCreatedNotifier` |
| `frontend/packages/components/settings/DiscordSettings.tsx` | `contract_created` event type |
| `internal/updaters/notifications_test.go` | Unit tests for new notification |
