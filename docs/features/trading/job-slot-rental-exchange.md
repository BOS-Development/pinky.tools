# Job Slot Rental Exchange

## Status

- **Phase**: 1 - Matchmaking Board (Implemented)
- **Scope**: Listing creation, interest requests, permission-gated browsing
- **Future**: Phase 2 will add in-game job execution tracking and contract integration

## Overview

A marketplace where players with idle industry job slots can rent them out to other players. Characters earn job slots through skills, but often leave slots unused. This feature allows slot holders to monetize idle capacity while allowing other players to access additional slots without training new characters.

Phase 1 is a pure matchmaking board: users list idle slots with flexible pricing, other users express interest, and coordination happens out-of-band (Discord, in-game). Phase 2 will add direct job execution and ESI contract integration for frictionless workflows.

## Key Decisions

1. **Three independent slot pools** — Manufacturing, Science (shared by ME/TE research, copying, invention), and Reactions each have separate limits:
   - Manufacturing: 1 + Mass Production (3387) + Adv Mass Production (24625)
   - Science: 1 + Laboratory Operation (3406) + Adv Laboratory Operation (24624)
   - Reactions: 1 + Mass Reactions (45748) + Adv Mass Reactions (45749)

2. **Science activities share one pool** — ME research, TE research, copying, and invention all draw from the same slot pool. This matches EVE's actual skill mechanics.

3. **Slot inventory is auto-calculated** — Idle slots are computed as: `total_slots_for_activity - active_jobs_for_activity`. Users cannot list more than available.

4. **Flexible pricing model** — Sellers choose the amount and unit:
   - `per_slot_day`: ISK per slot per day of rental
   - `per_job`: ISK per manufacturing job run
   - `flat_fee`: Single ISK amount for listed rental period

5. **Contact permission gate** — Browsing is permission-gated via the existing contact system using the `job_slot_browse` service type (mirrors `for_sale_browse` pattern). Users can see listings only from contacts who granted that permission.

6. **Phase 1 is matchmaking only** — No ESI job execution, no automatic contracts, no in-game tracking. Renters and sellers coordinate timing and payment out-of-band. Phase 2 will integrate with industry jobs and contracts.

7. **Soft-delete listings** — Listings are deactivated (not deleted) to preserve history and prevent accidental re-publication.

## Schema

### `job_slot_rental_listings` (NEW)

Represents a user's offer to rent job slots.

- `id` (bigint, PK)
- `user_id` (bigint, FK users)
- `character_id` (bigint, FK characters) — which character owns the slots
- `activity_type` (varchar) — enum: `manufacturing`, `reaction`, `copying`, `invention`, `me_research`, `te_research`
- `slots_listed` (int) — number of slots offered for rent (≤ idle slots)
- `price_amount` (numeric(12,2)) — ISK amount
- `pricing_unit` (varchar) — enum: `per_slot_day`, `per_job`, `flat_fee`
- `location_id` (bigint) — EVE location ID where job would run (station, upwell structure, etc.)
- `notes` (text, nullable) — seller's notes (terms, preferences, etc.)
- `is_active` (boolean) — soft-delete flag
- `created_at`, `updated_at` (timestamps)

**Indexes:**
- Unique partial: `(user_id, character_id, activity_type)` WHERE `is_active = true` — ensures one active listing per character per activity type
- Foreign keys: user, character, unique constraint on character ownership

### `job_slot_interest_requests` (NEW)

Represents interest from a renter to contact a listing owner.

- `id` (bigint, PK)
- `listing_id` (bigint, FK job_slot_rental_listings)
- `requester_user_id` (bigint, FK users) — the person interested in renting
- `slots_requested` (int) — how many slots needed (≤ `listing.slots_listed`)
- `duration_days` (int) — intended rental period
- `message` (text, nullable) — renter's notes or questions
- `status` (varchar) — enum: `pending`, `accepted`, `declined`, `withdrawn`
- `created_at`, `updated_at` (timestamps)

**Indexes:**
- Foreign keys: listing, requester user

### Contact Permission Service Type (MODIFIED)

`contact_permissions` table gains new service type:
- `job_slot_browse` — Users who grant this permission allow their contact listings to be browsed in the rental exchange

## API Endpoints

### Slot Inventory

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/job-slots/inventory` | Get auto-calculated idle slots per character per activity type |

**Response:**
```json
{
  "character_slots": [
    {
      "character_id": 1,
      "character_name": "Test Char",
      "total_slots": 5,
      "active_jobs": 2,
      "idle_slots": 3,
      "by_activity": [
        {
          "activity_type": "manufacturing",
          "total": 5,
          "active": 2,
          "idle": 3
        }
      ]
    }
  ]
}
```

### Listing Management

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/job-slots/listings` | Get user's active and inactive listings |
| POST | `/v1/job-slots/listings` | Create new listing |
| PUT | `/v1/job-slots/listings/{id}` | Update listing (price, notes, slots_listed, location) |
| DELETE | `/v1/job-slots/listings/{id}` | Soft-delete listing (set `is_active = false`) |

**POST/PUT body:**
```json
{
  "character_id": 123,
  "activity_type": "manufacturing",
  "slots_listed": 2,
  "price_amount": 5000000,
  "pricing_unit": "per_job",
  "location_id": 60003760,
  "notes": "High-sec only, no pirate BS"
}
```

### Browsing Listings

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/job-slots/listings/browse` | Browse all listings from contacts with `job_slot_browse` permission |

**Query params:**
- `activity_type` (optional) — filter by activity
- `character_name` (optional) — filter by character name
- `min_slots` (optional) — minimum idle slots available

**Response:** Array of listings with `character_name`, `user_name`, `slots_listed`, `price_amount`, `pricing_unit`, `location_id`, `notes`, `created_at`.

### Interest Management

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/job-slots/interest` | Express interest in a listing |
| GET | `/v1/job-slots/interest/sent` | Get sent interest requests (renter view) |
| GET | `/v1/job-slots/interest/received` | Get received interest requests (seller view) |
| PUT | `/v1/job-slots/interest/{id}/status` | Accept, decline, or withdraw interest |

**POST body:**
```json
{
  "listing_id": 42,
  "slots_requested": 1,
  "duration_days": 7,
  "message": "Need 1 slot for 7 days, can pay upfront"
}
```

**PUT body:**
```json
{
  "status": "accepted"
}
```

Allowed transitions:
- `pending` → `accepted`, `declined`
- `pending`, `accepted`, `declined` → `withdrawn` (by requester)

## File Structure

### Backend

**Migrations:**
- `internal/database/migrations/20260227151643_create_job_slot_rental_tables.up.sql` — Creates both tables and indexes
- Corresponding `.down.sql` — Drops tables

**Models:**
- `internal/models/models.go` — Added:
  - `JobSlotRentalListing`
  - `JobSlotInterestRequest`
  - `CharacterSlotInventory` (DTO for inventory calculation)

**Repositories:**
- `internal/repositories/jobSlotRentals.go` — All CRUD operations:
  - `CreateListing`, `UpdateListing`, `GetListingByID`, `GetListingsByUser`
  - `DeleteListing` (soft-delete)
  - `GetAllListingsForBrowse` (with contact permission filtering)
  - `CreateInterestRequest`, `UpdateInterestStatus`, `GetInterestByID`
  - `GetInterestBySender`, `GetInterestByListing` (for received requests)
  - Slot inventory calculation queries

**Controllers:**
- `internal/controllers/jobSlotRentals.go` — HTTP handlers for all endpoints:
  - `GetSlotInventory`, `GetListings`, `CreateListing`, `UpdateListing`, `DeleteListing`
  - `BrowseListings`
  - `ExpressInterest`, `GetSentInterestRequests`, `GetReceivedInterestRequests`, `UpdateInterestStatus`

**Wiring:**
- `cmd/industry-tool/cmd/root.go` — Register `jobSlotRentals` controller in router
- `internal/repositories/contactPermissions.go` — Service type `job_slot_browse` added to constants

**Utilities:**
- `internal/calculator/slots.go` — Slot calculation:
  - `CalculateManufacturingSlots(skills)` — Returns total manufacturing slots
  - `CalculateScienceSlots(skills)` — Returns total science slots (shared pool)
  - `CalculateReactionSlots(skills)` — Returns total reaction slots

### Frontend

**API Proxy Routes:**
- `frontend/pages/api/job-slots/inventory.ts` — GET /v1/job-slots/inventory
- `frontend/pages/api/job-slots/listings.ts` — GET/POST /v1/job-slots/listings
- `frontend/pages/api/job-slots/listing.ts` — PUT/DELETE /v1/job-slots/listings/{id}
- `frontend/pages/api/job-slots/browse.ts` — GET /v1/job-slots/listings/browse
- `frontend/pages/api/job-slots/interest.ts` — POST /v1/job-slots/interest
- `frontend/pages/api/job-slots/sent-interest.ts` — GET /v1/job-slots/interest/sent
- `frontend/pages/api/job-slots/received-interest.ts` — GET /v1/job-slots/interest/received
- `frontend/pages/api/job-slots/interest-status.ts` — PUT /v1/job-slots/interest/{id}/status

**API Client:**
- `frontend/packages/client/api.ts` — Added helper methods:
  - `getSlotInventory()`
  - `getListings()`, `createListing()`, `updateListing()`, `deleteListing()`
  - `browseListings()`
  - `expressInterest()`, `getSentInterestRequests()`, `getReceivedInterestRequests()`, `updateInterestStatus()`

**Components:**
- `frontend/packages/components/job-slots/SlotInventoryPanel.tsx` — Displays auto-calculated idle slots per character/activity
- `frontend/packages/components/job-slots/MyListings.tsx` — View, create, edit, delete user's listings
- `frontend/packages/components/job-slots/ListingsBrowser.tsx` — Browse listings from contacts with permission
- `frontend/packages/components/job-slots/InterestRequests.tsx` — Manage sent and received interest requests

**Page:**
- `frontend/packages/pages/JobSlotExchangePage.tsx` — 4-tab page:
  - Tab 1: "My Slot Inventory" (SlotInventoryPanel)
  - Tab 2: "My Listings" (MyListings)
  - Tab 3: "Browse Listings" (ListingsBrowser)
  - Tab 4: "Interest Requests" (InterestRequests)
- `frontend/pages/job-slots.tsx` — Page router entry point

**Navigation:**
- `frontend/packages/components/Navbar.tsx` — Added "Job Slot Exchange" link to Industry dropdown menu

## Location Resolution

Location IDs are stored in listings. Frontend displays the location ID as-is; future phases will add location name resolution via ESI bulk endpoint (similar to `npc-station-names.md`).

## Testing

### Backend Tests
- Unit tests for slot calculation logic in `internal/calculator/slots_test.go`
- Repository tests in `internal/repositories/jobSlotRentals_test.go`
- Controller tests in `internal/controllers/jobSlotRentals_test.go`

### E2E Tests
- Slot inventory calculation: `e2e/tests/job-slot-inventory.spec.ts`
- Listing creation and management: `e2e/tests/job-slot-listings.spec.ts`
- Interest request workflow: `e2e/tests/job-slot-interest.spec.ts`
- Permission-gated browsing: `e2e/tests/job-slot-browsing.spec.ts`

## Open Questions / Future Work

- **Phase 2**: ESI job execution integration — allow renters to submit jobs directly to seller's character
- **Phase 2**: Auto-contract generation — system creates and manages ESI contracts for payment
- **Phase 2**: Location name resolution — bulk endpoint for station/structure names
- **Phase 3**: Reputation system — renter/seller ratings to combat fraud
- **Phase 3**: Trust collateral — optional escrow or deposit to secure rental terms
