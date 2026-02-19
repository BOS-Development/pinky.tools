# Contact Rules

## Overview

Contact rules allow users to automatically create connections with all members of a corporation, alliance, or everyone in the system. When a rule is created, auto-accepted contacts are generated with all matching users, granting configurable permissions. When new users join the matching entity, contacts are auto-created for them too.

## Status

- **Issue**: #31
- **Branch**: `feature/contact-rules`
- **Status**: Implemented

## Key Decisions

- **Auto-create real contacts**: Rules produce actual `contacts` + `contact_permissions` rows rather than a separate permission path. Existing marketplace permission checks (`GetUserPermissionsForService`, `CheckPermission`, `GetBrowsableItems`) work unchanged.
- **Configurable permissions**: The rule creator selects which permissions to grant (e.g., `for_sale_browse`). These are stored on the rule as a `permissions` jsonb array and applied to all auto-created contacts. Matched users' permissions default to `false` (they can grant back manually).
- **No accept/reject flow**: Auto-created contacts are immediately `accepted`.
- **Cascade cleanup**: Contacts created by a rule have `contact_rule_id` FK with `ON DELETE CASCADE`. Deleting a rule removes all its auto-created contacts and their permissions.
- **Duplicate handling**: Before creating a contact, check both directions (A→B and B→A). Skip if any contact already exists between the pair.

## Rule Types

| Type | Matching Logic | Entity Required |
|------|---------------|-----------------|
| `corporation` | All users with a `player_corporations` row matching the corp ID | Yes (corporation ID) |
| `alliance` | All users with a `player_corporations` row matching the alliance ID | Yes (alliance ID) |
| `everyone` | All users in the system | No |

## Schema

### `contact_rules` table

| Column | Type | Description |
|--------|------|-------------|
| `id` | bigserial PK | |
| `user_id` | bigint FK → users | Rule creator |
| `rule_type` | varchar(20) | 'corporation', 'alliance', 'everyone' |
| `entity_id` | bigint nullable | Corporation/alliance ID (null for everyone) |
| `entity_name` | varchar(500) nullable | Display name |
| `is_active` | boolean | Soft-delete flag |
| `created_at` | timestamp | |
| `permissions` | jsonb | Array of service types to grant (default `["for_sale_browse"]`) |
| `updated_at` | timestamp | |

**Constraints:**
- `contact_rules_valid_type`: rule_type must be one of the three values
- `contact_rules_entity_check`: entity_id required for corp/alliance, null for everyone
- Unique partial index on `(user_id, rule_type, coalesce(entity_id, 0))` where `is_active = true`

### `contacts` table additions

| Column | Type | Description |
|--------|------|-------------|
| `contact_rule_id` | bigint FK → contact_rules(id) ON DELETE CASCADE | Links auto-created contacts to their rule |

### `player_corporations` table additions

| Column | Type | Description |
|--------|------|-------------|
| `alliance_id` | bigint nullable | EVE alliance ID from ESI affiliation |
| `alliance_name` | varchar(500) nullable | Alliance name from ESI |

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/contact-rules` | List user's active rules |
| POST | `/v1/contact-rules` | Create rule (triggers async ApplyRule) |
| DELETE | `/v1/contact-rules/{id}` | Delete rule + cascade delete auto-contacts |
| GET | `/v1/contact-rules/corporations?q=name` | Search corporations |
| GET | `/v1/contact-rules/alliances?q=name` | Search alliances |

## Auto-Contact Hooks

1. **Rule created** → `ApplyRule()` finds all matching users and creates contacts
2. **Corporation added** → `ApplyRulesForNewCorporation()` checks all active rules and creates contacts for the new user
3. **Character added** → Looks up the character's corporation from ESI, then calls `ApplyRulesForNewCorporation()` to check all active rules

## File Structure

### Backend
- `internal/database/migrations/20260219132354_contact_rules.*` — Table creation
- `internal/database/migrations/20260219132355_add_contact_rule_id_to_contacts.*` — FK on contacts
- `internal/database/migrations/20260219132356_add_alliance_to_player_corporations.*` — Alliance columns
- `internal/models/models.go` — `ContactRule` struct, updated `Contact` and `Corporation`
- `internal/repositories/contactRules.go` — CRUD, search, user matching
- `internal/repositories/contacts.go` — `CreateAutoContact`, `contact_rule_id` scanning
- `internal/repositories/contactPermissions.go` — `UpsertInTx` for transaction-aware permission upsert
- `internal/repositories/playerCorporations.go` — Alliance fields in struct and upsert
- `internal/updaters/contactRules.go` — `ApplyRule`, `ApplyRulesForNewCorporation`
- `internal/controllers/contactRules.go` — HTTP endpoints
- `internal/controllers/characters.go` — Hook for auto-contacts on character add (looks up corp from ESI)
- `internal/controllers/corporations.go` — Hook for auto-contacts on corp add
- `internal/client/esiClient.go` — Alliance info from ESI affiliation
- `cmd/industry-tool/cmd/root.go` — Wiring

### Frontend
- `frontend/pages/api/contact-rules/index.ts` — GET/POST proxy
- `frontend/pages/api/contact-rules/[id].ts` — DELETE proxy
- `frontend/pages/api/contact-rules/corporations.ts` — Corporation search proxy
- `frontend/pages/api/contact-rules/alliances.ts` — Alliance search proxy
- `frontend/packages/components/contacts/ContactsList.tsx` — Contact Rules tab + Auto chips
