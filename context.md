# EVE Industry Tool - Project Context

## Project Overview

A full-stack web application for managing EVE Online player inventory and assets. Built with Go backend, Next.js frontend, and PostgreSQL database. Integrates with EVE Online ESI (EVE Swagger Interface) API for real-time game data.

**Purpose**: Track and manage EVE Online characters, corporations, and their assets across multiple locations.

---

## Architecture

### Tech Stack

**Backend (Go 1.25.5)**
- Framework: Gorilla Mux (HTTP routing)
- Database: PostgreSQL with golang-migrate
- Authentication: Header-based (BACKEND-KEY, USER-ID)
- External APIs: EVE Online ESI, CCP Static Data Export (SDE)
- CLI: Cobra framework

**Frontend (Next.js 16.1.6)**
- React 19.2.3 with TypeScript 5
- UI: Material-UI (MUI) with Emotion
- Auth: NextAuth 4.24.13 with EVE Online OAuth
- State: Server-side props with NextAuth sessions
- Monorepo: Yarn workspaces (Lerna)

**Infrastructure**
- Docker containerized (Go 1.25, Node 24.9.0)
- Docker Compose for orchestration
- Makefile for common tasks

---

## Directory Structure

```
industry-tool/
├── cmd/industry-tool/          # Main entry point
│   ├── main.go                 # Application startup
│   └── cmd/                    # Cobra CLI commands
│       ├── root.go             # Service initialization
│       └── settings.go         # Environment config
├── internal/                   # Backend core logic
│   ├── models/                 # Domain models
│   ├── controllers/            # HTTP handlers
│   ├── repositories/           # Database access layer
│   ├── database/               # PostgreSQL + migrations
│   ├── client/                 # ESI + SDE API clients
│   ├── updaters/               # Business logic (SDE, prices, assets)
│   ├── runners/                # Background refresh runners
│   ├── web/                    # HTTP router
│   └── logging/                # Structured logging
├── frontend/                   # Next.js application
│   ├── app/                    # App router (layout, pages)
│   ├── pages/                  # Pages router + API routes
│   │   └── api/                # Backend API integration
│   ├── packages/               # Monorepo packages
│   │   ├── components/         # Shared UI components
│   │   ├── client/             # API client
│   │   ├── pages/              # Page components
│   │   └── styles/             # Shared styles
│   ├── theme.ts                # MUI theme config
│   └── middleware.ts           # NextAuth middleware
└── docker-compose*.yaml        # Docker orchestration
```

---

## Database Schema

### Core Tables

**Users & Authentication**
- `users` - User accounts (id, name)
- `characters` - EVE characters with ESI OAuth tokens
- `player_corporations` - Player-owned corporations
- `corporation_divisions` - Corporation hangar/wallet divisions

**Assets**
- `character_assets` - Individual character assets
- `character_asset_location_names` - Named container locations
- `corporation_assets` - Corporate assets
- `corporation_asset_location_names` - Named corp locations

**EVE Universe (Static Data — populated from CCP SDE)**
- `asset_item_types` - Item definitions (type_id, name, volume, group_id, mass, published, market_group_id, etc.)
- `regions` - EVE regions
- `constellations` - EVE constellations
- `solar_systems` - EVE solar systems
- `stations` - NPC stations (names come from seed data; SDE has no station names)

**SDE Reference Data** (see `docs/features/sde-import.md` for full schema)
- `sde_categories`, `sde_groups` - Item taxonomy (category → group → type)
- `sde_market_groups` - Market browser hierarchy (self-referencing parent_group_id)
- `sde_meta_groups` - Tech level classification (Tech I, Tech II, Faction, etc.)
- `sde_icons`, `sde_graphics` - Visual asset references

**SDE Blueprints & Industry**
- `sde_blueprints` - Blueprint definitions (max_production_limit)
- `sde_blueprint_activities` - Activity types per blueprint (manufacturing, reaction, invention, copying, etc.) with time
- `sde_blueprint_materials` - Input materials per activity (blueprint + activity + type_id → quantity)
- `sde_blueprint_products` - Output products per activity (includes probability for invention)
- `sde_blueprint_skills` - Required skills per activity
- `industry_cost_indices` - Per-system cost indices by activity (refreshed hourly from ESI)

**SDE Dogma (Item Attributes & Effects)**
- `sde_dogma_attributes`, `sde_dogma_attribute_categories` - Attribute definitions
- `sde_dogma_effects` - Effect definitions
- `sde_type_dogma_attributes` - Per-type attribute values (type_id + attribute_id → value)
- `sde_type_dogma_effects` - Per-type effects

**SDE NPC & Lore**
- `sde_factions`, `sde_npc_corporations`, `sde_races`, `sde_bloodlines`, `sde_ancestries`
- `sde_agents`, `sde_agents_in_space`

**SDE Industry (PI & POS)**
- `sde_planet_schematics`, `sde_planet_schematic_types` - Planetary interaction recipes
- `sde_control_tower_resources` - POS fuel requirements

**SDE Misc**
- `sde_skins`, `sde_skin_licenses`, `sde_skin_materials` - Ship customization
- `sde_certificates` - Certificate definitions
- `sde_metadata` - Tracks SDE checksum for freshness detection

**Key Relationships**
```
users (1) ←→ (N) characters
users (1) ←→ (N) player_corporations
characters (1) ←→ (N) character_assets
player_corporations (1) ←→ (N) corporation_assets
player_corporations (1) ←→ (N) corporation_divisions

sde_categories (1) ←→ (N) sde_groups (1) ←→ (N) asset_item_types
sde_blueprints (1) ←→ (N) sde_blueprint_activities (1) ←→ (N) sde_blueprint_materials
                                                    (1) ←→ (N) sde_blueprint_products
                                                    (1) ←→ (N) sde_blueprint_skills
asset_item_types (1) ←→ (N) sde_type_dogma_attributes
asset_item_types (1) ←→ (N) sde_type_dogma_effects
```

---

## API Endpoints

### Backend REST API (Port 8081)

**Characters**
- `GET /v1/characters/` - List user's characters
- `POST /v1/characters/` - Add character
- `GET /v1/characters/{id}` - Get character details

**Corporations**
- `GET /v1/corporations` - List user's corporations
- `POST /v1/corporations` - Add corporation

**Assets**
- `GET /v1/assets/` - Get aggregated assets by location

**Utilities**
- `GET /v1/users/refreshAssets` - Refresh from ESI
- `GET /v1/static/update` - Trigger SDE refresh

**Authentication Headers**
- `BACKEND-KEY` - Backend auth key (required)
- `USER-ID` - User identifier (for user-scoped endpoints)

### Frontend API Routes

- `/api/auth/[...nextauth]` - NextAuth EVE OAuth
- `/api/characters/add` - Initiate character OAuth flow
- `/api/characters/refreshAssets` - Trigger asset refresh
- `/api/corporations/add` - Initiate corp OAuth flow
- `/api/corporations/refreshAssets` - Trigger corp refresh
- `/api/assets/get` - Fetch user assets
- `/api/static/update` - Trigger SDE refresh

---

## Authentication Flow

### User Login (EVE Online OAuth)
1. User clicks "Sign In" → Redirected to EVE Online OAuth
2. User authorizes → Callback with OAuth token
3. NextAuth extracts `providerAccountId` (EVE character ID)
4. Backend checks if user exists via `/v1/users/{id}`
5. Creates user if new, otherwise loads existing
6. Session created with `providerAccountId` in JWT

### Adding Characters/Corporations
1. User initiates add flow → `/api/characters/add` or `/api/corporations/add`
2. Redirected to EVE OAuth with specific scopes
3. Callback receives access/refresh tokens
4. Tokens stored in database with expiry
5. Backend can now call ESI on behalf of character/corp

### Backend Authentication
- `BACKEND-KEY` header validates all requests
- `USER-ID` header identifies user for scoped operations
- Two middleware levels: `AuthAccessBackend`, `AuthAccessUser`

---

## ESI Integration

**EVE Swagger Interface Client** (`internal/client/esiClient.go`)

### Authenticated Methods (require OAuth tokens)
- `GetCharacterAssets()` - Fetch character assets
- `GetCharacterAssetNames()` - Get container names
- `GetCorporationAssets()` - Fetch corp assets
- `GetCorporationAssetNames()` - Get corp container names
- `GetCorporationDivisions()` - Get hangar/wallet divisions

### Public Methods (no auth required)
- `GetCcpMarketPrices()` - CCP adjusted/average prices per type (for industry job cost calculations)
- `GetIndustryCostIndices()` - Per-system cost indices by activity (manufacturing, reaction, etc.)

### OAuth Token Management
- Tokens stored per character/corporation
- Auto-refresh when expired (using refresh token)
- ESI rate limiting handled

### Location Flags
- `Hangar` - Personal hangar
- `OfficeFolder` - Corporation office
- `CorpSAG1-7` - Corporation hangar divisions
- `Deliveries` - Items in transit
- `AssetSafety` - Items in asset safety

---

## SDE (Static Data Export) Integration

**SDE Client** (`internal/client/sdeClient.go`) — Downloads and parses CCP's official EVE static data.

### Methods
- `GetChecksum(ctx)` - Fetch build number from CCP (for freshness check)
- `DownloadSDE(ctx)` - Download `sde.zip` (~84MB) to temp file
- `ParseSDE(zipPath)` - Parse all YAML files from ZIP into `SdeData` struct

### Data Refresh (Background Runners)
| Runner | Interval | Source | Description |
|--------|----------|--------|-------------|
| SDE Runner | 24h | CCP SDE ZIP | Full static data refresh (types, blueprints, dogma, universe, etc.) |
| CCP Prices Runner | 1h | ESI `/markets/prices/` | Adjusted prices for job cost calculations |
| Cost Indices Runner | 1h | ESI `/industry/systems/` | Per-system manufacturing/reaction cost indices |

All runners execute immediately on startup, then repeat at their interval. The SDE runner skips if the checksum hasn't changed.

### SDE Update Flow
1. Fetch checksum from CCP → compare with `sde_metadata["checksum"]`
2. If unchanged → skip (already current)
3. Download ZIP → parse all YAML files → `SdeData` struct
4. Upsert to existing tables: `asset_item_types`, `regions`, `constellations`, `solar_systems`, `stations`
5. Upsert to all `sde_*` tables
6. Store new checksum in `sde_metadata`

### Gotchas
- `npcStations.yaml` has **no station names** — station upsert preserves existing names when SDE provides empty ones
- Some YAML fields use mixed formats (plain string vs localized `{en: "...", de: "..."}` maps) — handled by `localizedString` custom unmarshaler
- Blueprint activities include: `manufacturing`, `reaction`, `invention`, `copying`, `researching_material_efficiency`, `researching_time_efficiency`

---

## Frontend Architecture

### Routing
- **App Router**: `/app/` (layout, home page)
- **Pages Router**: `/pages/` (characters, inventory, API routes)
- Hybrid approach during migration

### Components Structure

**Packages** (`/packages/components/`)
- `ThemeRegistry.tsx` - MUI theme provider
- `Navbar.tsx` - Top navigation bar
- `Loading.tsx` - Loading state indicator
- `Unauthorized.tsx` - Auth required page
- `characters/` - Character-specific components
  - `item.tsx` - Character card
  - `list.tsx` - Character grid layout

### State Management
- Server-side props via `getServerSideProps`
- NextAuth session state via `useSession` hook
- No global client state (using server props)

### Styling
- MUI theme: Dark mode (#0a0e1a background, #3b82f6 primary)
- Responsive grid layouts
- Emotion for CSS-in-JS
- Custom formatting utilities in `packages/utils/formatting.ts`

---

## Data Flow

### Asset Refresh Flow
```
1. User → /api/characters/refreshAssets
2. Frontend → Backend /v1/users/refreshAssets
3. Backend → ESI API (per character/corp)
4. ESI returns assets → Backend stores in DB
5. Backend aggregates by location → Returns response
```

### Asset Aggregation
Assets organized by location with nested structure:
```typescript
{
  structures: [
    {
      id: station_id,
      name: "Jita IV - Moon 4 - Caldari Navy Assembly Plant",
      solarSystem: "Jita",
      region: "The Forge",
      hangarAssets: [...],        // Character items
      hangarContainers: [...],    // Character containers
      corporationHangers: [       // Corp divisions
        {
          id: 1,
          name: "Corp Hangar",
          assets: [...],
          hangarContainers: [...]
        }
      ]
    }
  ]
}
```

---

## Code Pattern Examples

### Repository Pattern
```go
type CharacterAssets struct {
    db *sql.DB
}

func NewCharacterAssets(db *sql.DB) *CharacterAssets {
    return &CharacterAssets{db: db}
}

func (r *CharacterAssets) UpdateAssets(ctx context.Context, characterID, userID int64, assets []*models.EveAsset) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return errors.Wrap(err, "failed to begin transaction")
    }
    defer tx.Rollback()
    // ...
    return tx.Commit()
}
```

### Frontend Components
```typescript
export const getServerSideProps: GetServerSideProps = async (context) => {
  const session = await getSession(context);
  if (!session) {
    return { props: {} };
  }
  const api = client(process.env.BACKEND_URL, session.providerAccountId);
  const response = await api.getCharacters();
  return { props: { characters: response.data } };
};

export default function Page(props: PageProps) {
  const { data: session, status } = useSession();
  if (status === "loading") return <Loading />;
  if (status !== "authenticated") return <Unauthorized />;
  return <List items={props.items} />;
}
```

### Error Handling
- Backend: Wrap errors with `github.com/pkg/errors`
- Frontend: Return error kinds (`success`, `error`, `not-found`)
- HTTP: Standard status codes (200, 401, 404, 500)

---

## Environment Variables

### Backend
```bash
PORT=8081
BACKEND_KEY=your-backend-key
DATABASE_HOST=localhost
DATABASE_PORT=19236
DATABASE_NAME=app
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres
```

### Frontend
```bash
BACKEND_URL=http://localhost:8081
NEXTAUTH_SECRET=your-secret
EVE_CLIENT_ID=your-eve-client-id
EVE_CLIENT_SECRET=your-eve-client-secret
NEXTAUTH_URL=http://localhost:3000
```

---

## Key Constraints & Design Decisions

1. **Corporation Assets**: Filtered by `OfficeFolder` location flag to identify corp offices
2. **Division Extraction**: `CorpSAG{N}` → Division number from substring position 8
3. **Asset Organization**: By station/structure, then by type (hangar, container, corp division)
4. **Token Storage**: ESI OAuth tokens stored per character/corporation with auto-refresh
5. **Monorepo**: Frontend uses Yarn workspaces for shared packages
6. **Authentication**: Two-level (backend key + user ID) for security
7. **Static Data**: Auto-refreshed from CCP SDE (24h) and ESI public endpoints (1h)

---

## External Resources

- **EVE Online ESI**: https://esi.evetech.net/
- **EVE Image Server**: https://image.eveonline.com/
- **CCP SDE**: https://developers.eveonline.com/static-data/
- **Next.js Docs**: https://nextjs.org/docs
- **MUI Docs**: https://mui.com/material-ui/

---

## Notes

- **EVE Character Images**: `https://image.eveonline.com/Character/{id}_128.jpg`
- **Corporation Logos**: `https://images.evetech.net/corporations/{id}/logo`
- **Item Icons**: `https://images.evetech.net/types/{type_id}/icon`
- **Test Data**: See E2E test users in `docs/features/e2e-testing.md`; Jita station ID: 60003760
- **Division Types**: `hangar` and `wallet` stored separately in `corporation_divisions`
