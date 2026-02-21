# Auto-Complete Purchases via ESI Contract Sync

## Overview

Automatically detects when an EVE Online contract has been accepted by scanning buyer characters' contracts via ESI. Matches contracts to purchases using the `contract_key` in the contract title, then auto-completes the purchase.

The manual "Complete" button remains available as a fallback.

## Status

- **Phase**: Implementation
- **Branch**: `feature/contract-sync`

## How It Works

1. Seller marks purchase as "contract created" → app auto-generates a `contract_key` (e.g., `PT-123`)
2. Seller copies the key into the EVE in-game contract title when creating the contract
3. Background runner (every 15 minutes) scans buyer characters' ESI contracts
4. Finds `finished` item_exchange contracts whose title contains a known `contract_key`
5. Auto-completes all purchases sharing that key

## Key Decisions

- **Matching via title**: Uses the contract title field from ESI to match against `contract_key`. No need for the contract items sub-endpoint.
- **Auto-generated keys**: Format `PT-{purchaseID}`. Custom keys still supported for grouping multiple purchases.
- **Manual fallback**: Buyers can still click "Complete" manually if the seller forgot to include the key.
- **Scope check**: Only processes characters with `esi-contracts.read_character_contracts.v1` scope.
- **15-minute interval**: Balances responsiveness with ESI rate limits.

## File Structure

| File | Purpose |
|------|---------|
| `internal/client/esiClient.go` | `GetCharacterContracts` + `EsiContract` type |
| `internal/controllers/purchases.go` | Auto-generate `contract_key` in `MarkContractCreated` |
| `internal/repositories/purchaseTransactions.go` | `GetContractCreatedWithKeys`, `CompleteWithContractID` |
| `internal/updaters/contractSync.go` | Matching logic + auto-completion |
| `internal/updaters/contractSync_test.go` | Unit tests |
| `internal/runners/contractSync.go` | Background runner |
| `cmd/industry-tool/cmd/root.go` | Wiring |

## ESI Endpoints Used

- `GET /v1/characters/{character_id}/contracts/` — paginated contract list
- Scope: `esi-contracts.read_character_contracts.v1` (already in player scopes)
