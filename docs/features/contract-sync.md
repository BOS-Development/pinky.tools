# Auto-Complete Purchases via ESI Contract Sync

## Overview

Automatically detects when an EVE Online contract has been accepted by scanning buyer characters' and corporations' contracts via ESI. Matches contracts to purchases using the `contract_key` in the contract title, then auto-completes the purchase.

The manual "Complete" button remains available as a fallback.

## Status

- **Phase**: Complete

## How It Works

1. Seller marks purchase as "contract created" → app auto-generates a `contract_key` (e.g., `PT-123`)
2. Seller copies the key into the EVE in-game contract title when creating the contract
3. Background runner (every 15 minutes) scans buyer characters' and corporations' ESI contracts
4. Finds `finished` item_exchange contracts whose title contains a known `contract_key`
5. Auto-completes all purchases sharing that key

## Key Decisions

- **Matching via title**: Uses the contract title field from ESI to match against `contract_key`. No need for the contract items sub-endpoint.
- **Auto-generated keys**: Format `PT-{purchaseID}`. Custom keys still supported for grouping multiple purchases.
- **Manual fallback**: Buyers can still click "Complete" manually if the seller forgot to include the key.
- **Character + corporation search**: Searches both character contracts (personal) and corporation contracts. A seller can assign the contract to either the buyer's character or their corporation.
- **Scope check**: Only processes characters with `esi-contracts.read_character_contracts.v1` and corporations with `esi-contracts.read_corporation_contracts.v1`.
- **15-minute interval**: Balances responsiveness with ESI rate limits.

## File Structure

| File | Purpose |
|------|---------|
| `internal/client/esiClient.go` | `GetCharacterContracts`, `GetCorporationContracts` + `EsiContract` type |
| `internal/controllers/purchases.go` | Auto-generate `contract_key` in `MarkContractCreated` |
| `internal/repositories/purchaseTransactions.go` | `GetContractCreatedWithKeys`, `CompleteWithContractID` |
| `internal/repositories/playerCorporations.go` | `Get` (fetch corps for user), `UpdateTokens` |
| `internal/updaters/contractSync.go` | Matching logic + auto-completion for characters and corporations |
| `internal/updaters/contractSync_test.go` | Unit tests (13 tests: 8 character, 5 corporation) |
| `internal/runners/contractSync.go` | Background runner |
| `cmd/industry-tool/cmd/root.go` | Wiring |

## ESI Endpoints Used

- `GET /v1/characters/{character_id}/contracts/` — paginated character contract list
  - Scope: `esi-contracts.read_character_contracts.v1`
- `GET /v1/corporations/{corporation_id}/contracts/` — paginated corporation contract list
  - Scope: `esi-contracts.read_corporation_contracts.v1`
