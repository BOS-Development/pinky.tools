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
- External APIs: EVE Online ESI, FuzzWorks
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
│   ├── client/                 # ESI API client
│   ├── updaters/               # Business logic
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

**EVE Universe (Static Data)**
- `asset_item_types` - Item definitions (type_id, name, volume)
- `regions` - EVE regions
- `constellations` - EVE constellations
- `solar_systems` - EVE solar systems
- `stations` - NPC stations

**Key Relationships**
```
users (1) ←→ (N) characters
users (1) ←→ (N) player_corporations
characters (1) ←→ (N) character_assets
player_corporations (1) ←→ (N) corporation_assets
player_corporations (1) ←→ (N) corporation_divisions
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
- `GET /v1/static/update` - Update static EVE data

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
- `/api/static/update` - Update static data

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

### Methods
- `GetCharacterAssets()` - Fetch character assets
- `GetCharacterAssetNames()` - Get container names
- `GetCorporationAssets()` - Fetch corp assets
- `GetCorporationAssetNames()` - Get corp container names
- `GetCorporationDivisions()` - Get hangar/wallet divisions

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

## Development Workflow

### Backend Development
```bash
# Run backend server
go run cmd/industry-tool/main.go

# Run tests
go test ./internal/repositories -v

# Run integration tests (requires database)
make integration-test-clean
docker-compose -f docker-compose.ci.yaml up -d database
go test ./internal/repositories -run Test_Assets
```

### Frontend Development
```bash
cd frontend
npm install
npm run dev  # Starts on localhost:3000
```

### Database Migrations
```bash
# Migrations located in internal/database/migrations/
# Format: {version}_{name}.up.sql

# Auto-applied on server startup via golang-migrate
```

### Docker Development
```bash
# Full stack
make dev

# Clean up
make dev-clean

# Integration tests
make integration-test
```

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

## Code Patterns

### Repository Pattern
```go
type CharacterAssets struct {
    db *sql.DB
}

func NewCharacterAssets(db *sql.DB) *CharacterAssets {
    return &CharacterAssets{db: db}
}

func (r *CharacterAssets) UpdateAssets(ctx context.Context, characterID, userID int64, assets []*models.EveAsset) error {
    // Use transactions with deferred Rollback
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return errors.Wrap(err, "failed to begin transaction")
    }
    defer tx.Rollback()

    // Execute queries
    // ...

    return tx.Commit()
}
```

### Frontend Components
```typescript
// Server-side props
export const getServerSideProps: GetServerSideProps = async (context) => {
  const session = await getSession(context);
  if (!session) {
    return { props: {} };
  }

  const api = client(process.env.BACKEND_URL, session.providerAccountId);
  const response = await api.getCharacters();

  return {
    props: { characters: response.data }
  };
};

// Component
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

## Testing

**IMPORTANT**: Always prefer using the Makefile targets to run all tests. Do not run test commands directly — the Makefile handles Docker Compose orchestration, environment setup, and cleanup.

```bash
# Unit/integration tests
make test-backend        # Backend Go tests with coverage
make test-frontend       # Frontend Jest tests with coverage
make test-all            # Run both backend and frontend tests
make test-clean          # Clean up test containers

# End-to-end tests (Playwright)
make test-e2e            # Run E2E tests headless
make test-e2e-ui         # Run E2E tests with Playwright UI
make test-e2e-clean      # Clean up E2E containers
```

### Integration Tests
- Located in `internal/repositories/*_test.go`
- Use testcontainers pattern (random DB names)
- Setup test data via repositories
- Assert expected vs actual structs

### Test Database
- PostgreSQL on port 19236
- Started via `docker-compose.ci.yaml`
- Migrations auto-applied
- Cleaned up after tests

---

## Key Constraints & Design Decisions

1. **Corporation Assets**: Filtered by `OfficeFolder` location flag to identify corp offices
2. **Division Extraction**: `CorpSAG{N}` → Division number from substring position 8
3. **Asset Organization**: By station/structure, then by type (hangar, container, corp division)
4. **Token Storage**: ESI OAuth tokens stored per character/corporation with auto-refresh
5. **Monorepo**: Frontend uses Yarn workspaces for shared packages
6. **Authentication**: Two-level (backend key + user ID) for security
7. **Static Data**: Periodic updates from ESI/FuzzWorks APIs

---

## Common Tasks

### Add New Repository
1. Create `internal/repositories/myrepo.go`
2. Implement struct with `*sql.DB`
3. Add methods with transactions
4. Create test file `myrepo_test.go`
5. Wire up in `cmd/cmd/root.go`

### Add New API Endpoint
1. Create handler in `internal/controllers/`
2. Register route in `internal/web/router.go`
3. Add client method in `frontend/packages/client/api.ts`
4. Create frontend API route in `frontend/pages/api/`
5. Call from component via `getServerSideProps`

### Add New Component
1. Create in `frontend/packages/components/`
2. Use MUI components for consistency
3. Follow naming: `Item` (card), `List` (grid)
4. Export from component directory

### Update Database Schema
1. Create migration file `{version}_{name}.up.sql`
2. Write SQL with tab indentation, lowercase keywords
3. Restart server to auto-apply
4. Update repository queries as needed

---

## External Resources

- **EVE Online ESI**: https://esi.evetech.net/
- **EVE Image Server**: https://image.eveonline.com/
- **FuzzWorks API**: https://www.fuzzwork.co.uk/
- **Next.js Docs**: https://nextjs.org/docs
- **MUI Docs**: https://mui.com/material-ui/

---

## AI Agent Guidelines

### Working on This Project

**General Approach**
- Always read files before editing them
- Follow existing patterns from similar components/files
- Test changes before marking work complete
- When unsure, ask the user rather than guessing

**Git Workflow**
- **NEVER commit directly to main branch** - Always use feature branches
- Branch naming: `feature/{feature-name}` or `fix/{bug-name}`
  - Examples: `feature/add-buy-orders`, `fix/null-contacts-error`
- Before starting work:
  ```bash
  git checkout main
  git pull origin main
  git checkout -b feature/your-feature-name
  ```
- Commit frequently with clear messages
- Push branch and create PR when ready for review
- Delete branch after PR is merged

**Planning Complex Features**
- For multi-file changes or architectural decisions, use plan mode first
- Present options to the user when multiple approaches are valid
- Break down large features into phases with clear verification steps
- **Feature Documentation**: All feature plans MUST be documented in `docs/features/`
  - Always check for existing feature plans before starting work
  - Create a feature doc for every new feature or significant change (include overview, design decisions, schema, file structure, verification steps)
  - Update existing docs when modifying a feature
  - Examples: `contact-marketplace.md`, `jita-market-pricing.md`, `e2e-testing.md`
  - Store in repo (`docs/features/`), not in local `.claude/plans/` directory

**Code Quality Standards**

*Backend (Go):*
- **CRITICAL**: Initialize slices as `items := []*Type{}` NOT `var items []*Type`
  - Prevents nil JSON marshaling (nil → `null` instead of `[]`)
  - Example: `contacts := []*models.Contact{}` for rows.Next() loops
- Use transactions with deferred Rollback for multi-statement operations
- Wrap errors with `github.com/pkg/errors` for context
- Follow repository → controller → router pattern

*Frontend (React/Next.js):*
- **MUI SSR**: ThemeRegistry must use Emotion cache (see `ThemeRegistry.tsx`)
- **Formatting**: Use utilities from `packages/utils/formatting.ts` for ISK/numbers
- **Authentication**: Check session status before rendering protected content
- **API Routes**: Proxy to backend, don't implement business logic
- Read existing components before creating similar ones

**Testing Requirements**

*Backend:*
- Write integration tests in `*_test.go` files
- Test repository methods with real database (testcontainers)
- Cover success cases, edge cases, and error scenarios
- Use table-driven tests for multiple scenarios

*Frontend:*
- **Snapshot testing**: Create snapshots for all new components
  - Run `npm test -- -u` to update snapshots after intentional changes
  - Location: `__tests__/{ComponentName}.test.tsx`
  - Test loading, error, and success states
  - Example:
    ```typescript
    import { render } from '@testing-library/react';
    import MyComponent from '../MyComponent';

    describe('MyComponent', () => {
      it('matches snapshot with data', () => {
        const { container } = render(<MyComponent data={mockData} />);
        expect(container).toMatchSnapshot();
      });

      it('matches snapshot when loading', () => {
        const { container } = render(<MyComponent loading={true} />);
        expect(container).toMatchSnapshot();
      });
    });
    ```
- Manually test in browser before marking complete
- Verify edge cases (empty data, errors, null values)
- Test both character and corporation flows if applicable
- Check responsive layouts (mobile, tablet, desktop)

**Common Pitfalls to Avoid**

1. **Go nil slices → JSON null**: Always initialize `items := []*Type{}`
2. **MUI FOUC**: Ensure ThemeRegistry has Emotion cache setup
3. **Missing auth headers**: Backend needs `BACKEND-KEY` and `USER-ID`
4. **Incomplete transactions**: Always defer `tx.Rollback()` before operations
5. **Hardcoded IDs**: Use session providerAccountId, not hardcoded values

**UI/UX Standards**

*Design System:*
- Dark theme: Background `#0a0e1a`, Cards `#12151f`, Primary `#3b82f6`
- Use gradients for important metrics: `linear-gradient(135deg, rgba(59, 130, 246, 0.1) 0%, ...)`
- Color coding: Green `#10b981` for revenue/success, Red `#ef4444` for costs/errors
- Icons: Use MUI icons from `@mui/icons-material`

*Component Patterns:*
- Stat cards: Gradient background + colored border + icon + formatted value
- Tables: Dark header (`#0f1219`), alternating row colors, right-align numbers
- Loading states: Use `<Loading />` component, not custom spinners
- Empty states: Centered message in table cell with `colSpan`

*Formatting:*
```typescript
import { formatISK, formatNumber, formatCompact } from '@industry-tool/utils/formatting';

// Currency: "1.5M ISK", "842.3K ISK"
formatISK(value)

// Large numbers: "1,234,567"
formatNumber(value)

// Compact: "1.5M", "842K"
formatCompact(value)
```

**File Organization**
- Backend: `internal/repositories/` → `internal/controllers/` → `cmd/cmd/root.go`
- Frontend components: `packages/components/{feature}/{ComponentName}.tsx`
- API routes: `pages/api/{feature}/{action}.ts`
- Types/interfaces: Define in component file or `internal/models/models.go`

**Migration & Schema Changes**
- **Creating new migrations**: Use the helper script
  - Command: `./scripts/new-migration.sh migration_name`
  - Example: `./scripts/new-migration.sh add_user_preferences`
  - Auto-generates both `.up.sql` and `.down.sql` files with timestamp
- **Naming**: Timestamp-based versions (prevents merge conflicts)
  - Format: `{YYYYMMDDHHMMSS}_{name}.up.sql` (no underscore in timestamp)
  - Example: `20250215143022_create_contacts.up.sql`
  - Manual timestamp: `date +%Y%m%d%H%M%S`
- Location: `internal/database/migrations/`
- Style: Lowercase SQL keywords, tab indentation
- Auto-applied on server startup (no manual commands needed)
- Update repository methods after schema changes
- Always create both `.up.sql` and `.down.sql` for rollback support

**Debugging Tips**
- Backend logs: Check container output via `docker logs`
- Frontend: Use browser DevTools Network tab for API calls
- Database: Connect via `psql -h localhost -p 19236 -U postgres -d app`
- Authentication issues: Verify session in browser DevTools Application tab

**When in Doubt**
- Check similar existing features for patterns
- Read the full file before editing
- Ask user for clarification on requirements
- Prefer simple solutions over complex abstractions

---

## Notes

- **EVE Character Images**: `https://image.eveonline.com/Character/{id}_128.jpg`
- **Corporation Logos**: `https://images.evetech.net/corporations/{id}/logo`
- **Item Icons**: `https://images.evetech.net/types/{type_id}/icon`
- **Test Data**: See E2E test users in `docs/features/e2e-testing.md`; Jita station ID: 60003760
- **Division Types**: `hangar` and `wallet` stored separately in `corporation_divisions`
