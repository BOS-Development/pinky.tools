# Job Slot Rental Exchange

## Status

- **Phase**: 2 - Discord Notifications (Implemented)
- **Current Scope**: Listing creation, interest requests, permission-gated browsing, Discord notifications for interest events
- **Phase 3 (Planned)**: Rental agreements, ESI job tracking (read-only), Discord job completion notifications

## Overview

A marketplace where players with idle industry job slots can rent them out to other players. Characters earn job slots through skills, but often leave slots unused. This feature allows slot holders to monetize idle capacity while allowing other players to access additional slots without training new characters.

Phase 1 was a pure matchmaking board: users list idle slots with flexible pricing, other users express interest, and coordination happened out-of-band (Discord, in-game). Phase 2 added Discord notifications for interest events, alerting renters and sellers in real-time when interest is received or status changes. Phase 3 will formalize accepted agreements into tracked deals and add read-only ESI job tracking for visibility into running rental jobs.

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

8. **Phase 3: Manual contracts** — ESI cannot create contracts via API, so contracts remain manual in-game. The agreement tracks which interest request was accepted, and the slot owner is responsible for creating a contract in-game (or using cash trade for payment). Phase 3 focuses on job visibility and agreement tracking, not payment automation.

9. **Phase 3: Read-only ESI job tracking** — Phase 3 reads industry jobs from the slot owner's ESI (who is executing the rented-out jobs), but does not submit or modify jobs. Job submission remains the renter's responsibility in-game.

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

### `job_slot_agreements` (NEW — Phase 3)

Represents a formalized rental deal between a slot owner and a renter.

- `id` (bigint, PK)
- `interest_request_id` (bigint, FK job_slot_interest_requests) — the interest request that triggered this agreement
- `listing_id` (bigint, FK job_slot_rental_listings) — the listing being rented
- `seller_user_id` (bigint, FK users) — slot owner (seller)
- `renter_user_id` (bigint, FK users) — person renting slots (renter)
- `slots_agreed` (int) — number of slots committed in this deal
- `price_amount` (numeric(12,2)) — total ISK amount or unit cost
- `pricing_unit` (varchar) — enum: `per_slot_day`, `per_job`, `flat_fee` (copied from listing at time of agreement)
- `agreed_at` (timestamp) — when the agreement was created
- `expected_end_at` (timestamp, nullable) — optional agreed end date for the rental period
- `status` (varchar) — enum: `active`, `completed`, `cancelled` — lifecycle state of the agreement
- `cancellation_reason` (text, nullable) — reason if cancelled
- `created_at`, `updated_at` (timestamps)

**Indexes:**
- Foreign keys: interest_request, listing, seller_user, renter_user
- Composite: `(seller_user_id, status)` — for listing agreements by seller and status
- Composite: `(renter_user_id, status)` — for listing agreements by renter and status

**Lifecycle:**
- Created automatically (atomic transaction) when an interest request status transitions to `accepted`
- Seller can mark `completed` or `cancelled`
- Renter can optionally request cancellation (separate cancel endpoint or status update permission)

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

## Rental Agreements (Phase 3)

### Agreement Creation

When a seller accepts an interest request, a corresponding rental agreement is automatically created in the same database transaction:
1. Interest status transitions from `pending` to `accepted`
2. New row is inserted into `job_slot_agreements` with the agreement details copied from the interest request and listing
3. Agreement starts in `active` status
4. Both events (interest accepted + agreement created) fire Discord notifications

### Agreement Lifecycle

**Seller actions:**
- Mark agreement as `completed` once the rental period ends or all runs are finished
- Mark agreement as `cancelled` (with optional reason) if terms change or renter defaults

**Renter actions:**
- View active and past agreements
- Optionally request cancellation (no automatic action — seller decides)

**Tracking:**
- Sellers see active agreements to know how many slots are committed (against `job_slot_agreements` with status `active`)
- Renters see which deals are active and which have ended
- Both can reference ESI job data to see what's actually running

### API Endpoints (Phase 3)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/job-slots/agreements` | Get user's agreements (as seller or renter) with filter by status |
| PUT | `/v1/job-slots/agreements/{id}/status` | Update agreement status to `completed` or `cancelled` |
| GET | `/v1/job-slots/agreements/{id}/jobs` | Get ESI industry jobs for the listing's character (owner's view of running rental jobs) |

**GET /v1/job-slots/agreements query params:**
- `status` (optional) — filter: `active`, `completed`, `cancelled`
- `role` (optional) — `seller` or `renter` (defaults to both)

**GET /v1/job-slots/agreements response:**
```json
{
  "agreements": [
    {
      "id": 42,
      "interest_request_id": 5,
      "listing_id": 10,
      "seller_user_id": 1,
      "seller_name": "SlotOwner",
      "renter_user_id": 2,
      "renter_name": "RenterName",
      "slots_agreed": 2,
      "price_amount": 5000000,
      "pricing_unit": "per_job",
      "agreed_at": "2026-03-15T10:00:00Z",
      "expected_end_at": "2026-03-22T10:00:00Z",
      "status": "active",
      "created_at": "2026-03-15T10:00:00Z",
      "updated_at": "2026-03-15T10:00:00Z"
    }
  ]
}
```

**PUT /v1/job-slots/agreements/{id}/status body:**
```json
{
  "status": "completed",
  "cancellation_reason": null
}
```
or
```json
{
  "status": "cancelled",
  "cancellation_reason": "Renter no longer needs slots"
}
```

**GET /v1/job-slots/agreements/{id}/jobs response:**
```json
{
  "agreement_id": 42,
  "character_id": 123,
  "jobs": [
    {
      "job_id": 987,
      "activity_id": 1,
      "activity_name": "Manufacturing",
      "blueprint_id": 456,
      "blueprint_name": "Rifter Blueprint",
      "location_id": 60003760,
      "runs": 10,
      "start_date": "2026-03-15T12:00:00Z",
      "end_date": "2026-03-17T14:30:00Z",
      "status": "active"
    }
  ]
}
```

## ESI Industry Job Tracking (Phase 3)

The slot owner can pull their character's industry jobs via ESI to see what rental jobs are running. Each agreement links to the listing's character, and the owner can view that character's job list to confirm slots are in use.

**Data source:** ESI `/characters/{character_id}/industry/jobs/` endpoint

**Caching:** Short TTL (5 minutes) to balance visibility with ESI rate limits

**Permissions:** Only the slot owner (seller) can view jobs for a listing character. Renters see agreement details but not the live job list (to avoid exposing the seller's ESI token or other sensitive job data).

## Discord Notifications (Phase 2)

Two new event types were added to the Discord notification system for real-time interest updates:

### `job_slot_interest_received`

Fires when a renter expresses interest in a listing. Notifies the **listing owner** (seller).

- **Trigger**: New interest request is created (status `pending`)
- **Recipient**: User who created the listing
- **Content**: Requester name, listing details (character, activity type, slots listed), slots requested, duration

### `job_slot_interest_updated`

Fires when a seller accepts or declines an interest request. Notifies the **requester** (renter). Does NOT fire for `withdrawn` status.

- **Trigger**: Interest status changes to `accepted` or `declined`
- **Recipient**: User who created the interest request
- **Content**: Action taken (accepted/declined), listing details, seller name, new status

### `job_slot_job_completed` (Phase 3)

Fires when the system detects (via ESI polling) that a job has completed on a rented-out character.

- **Trigger**: ESI job status transitions to `delivered` or `cancelled` (detected during routine job sync)
- **Recipient**: Renter (optional — seller also receives if configured)
- **Content**: Job summary (item, runs completed, location), agreement details, seller name
- **Polling**: Seller's industry jobs are polled on a background interval (e.g., 15 min) to detect completion
- **Notification**: Sent as a non-blocking goroutine; failure does not impact polling

### Implementation

All event notifications are handled as **non-blocking goroutines** (fire-and-forget):
- Notification is sent in a background goroutine; failure does not affect the request outcome
- The API call succeeds regardless of notification delivery status
- Users configure event types in **Settings → Discord Notifications** like any other event type
- If a user hasn't linked Discord or disabled events, the notification is silently skipped
- Phase 3 adds `job_slot_job_completed` as a new configurable event type

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
  - `CreateInterestRequest`, `UpdateInterestStatus`, `GetInterestByID` (returns enriched interest with requester name and listing details)
  - `GetInterestBySender`, `GetInterestByListing` (for received requests)
  - Slot inventory calculation queries

**Controllers:**
- `internal/controllers/jobSlotRentals.go` — HTTP handlers for all endpoints:
  - `GetSlotInventory`, `GetListings`, `CreateListing`, `UpdateListing`, `DeleteListing`
  - `BrowseListings`
  - `ExpressInterest`, `GetSentInterestRequests`, `GetReceivedInterestRequests`, `UpdateInterestStatus`

**Notifications (Phase 2):**
- `internal/updaters/notifications.go` — Discord event notifiers:
  - `JobSlotInterestNotifier` interface — abstraction for notification delivery
  - `NotifyJobSlotInterestReceived` — builds and sends `job_slot_interest_received` embed; runs in non-blocking goroutine
  - `NotifyJobSlotInterestStatusUpdated` — builds and sends `job_slot_interest_updated` embed for `accepted`/`declined` statuses; runs in non-blocking goroutine
  - Embed builders construct rich Discord messages with relevant context

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

## Phase 3 Key Decisions

- **Agreements are auto-created on acceptance** — when a seller accepts an interest request, the agreement row is created atomically in the same transaction
- **ESI contracts are manual** — ESI API cannot create contracts, so contract creation and payment transfer remain manual in-game activities (or cash trade). Phase 3 does not automate payment
- **Job tracking is read-only** — Phase 3 reads ESI job data to show what's running but does not submit, cancel, or modify jobs. The renter submits jobs in-game on the seller's character
- **Only seller can view ESI jobs** — To prevent renters from seeing the seller's full job list or ESI token leakage, only the slot owner (seller) views ESI job data for a listing character
- **Job polling is background interval** — Job completion detection runs on a scheduled background interval (e.g., 15 min), not real-time. Polling is per-seller, aggregating all their listing characters

## Open Questions / Future Work

- **Phase 3**: Location name resolution — bulk endpoint for station/structure names
- **Phase 4**: Reputation system — renter/seller ratings to combat fraud
- **Phase 4**: Trust collateral — optional escrow or deposit to secure rental terms
- **Phase 4**: In-game job submission — allow renters to submit jobs directly via API (requires ESI account delegation or advanced OAuth scopes)
