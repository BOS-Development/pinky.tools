# EVE Industry Tool - Project Context

## Project Overview

A full-stack web application for managing EVE Online player inventory and assets. Built with Go backend, Next.js frontend, and PostgreSQL database. Integrates with EVE Online ESI (EVE Swagger Interface) API for real-time game data.

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

## Core Database Tables

**Users & Authentication**
- `users` - User accounts (id, name)
- `characters` - EVE characters with ESI OAuth tokens
- `player_corporations` - Player-owned corporations
- `corporation_divisions` - Corporation hangar/wallet divisions

**Assets**
- `character_assets` - Individual character assets
- `corporation_assets` - Corporate assets
- `asset_item_types` - Item definitions (type_id, name, volume, group_id)

**Key Relationships**
```
users (1) ←→ (N) characters
users (1) ←→ (N) player_corporations
characters (1) ←→ (N) character_assets
player_corporations (1) ←→ (N) corporation_assets
```

For feature-specific tables, see the relevant feature doc below.

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

## External Resources

- **EVE Online ESI**: https://esi.evetech.net/
- **EVE Image Server**: https://image.eveonline.com/
- **EVE Character Images**: `https://image.eveonline.com/Character/{id}_128.jpg`
- **Corporation Logos**: `https://images.evetech.net/corporations/{id}/logo`
- **Item Icons**: `https://images.evetech.net/types/{type_id}/icon`
- **CCP SDE**: https://developers.eveonline.com/static-data/

---

## Feature Index

Read the relevant doc before working on a feature. Each doc contains schema, API endpoints, key decisions, and file paths.

### Core Systems
| Feature | Doc | Summary |
|---------|-----|---------|
| Characters | `docs/features/characters.md` | ESI OAuth, token storage, character asset refresh |
| Corporations | `docs/features/corporations.md` | Corp management, divisions, corp asset refresh |
| SDE Import | `docs/features/sde-import.md` | Static data pipeline, all `sde_*` tables, background runners |
| Authentication | `docs/features/consolidate-oauth.md` | OAuth consolidation, scopes, single callback |
| Background Updates | `docs/features/background-asset-updates.md` | Asset refresh runners (1h), concurrency |
| NPC Station Names | `docs/features/npc-station-names.md` | Station name resolution via ESI bulk endpoint |
| Landing Page | `docs/features/landing-page.md` | Hero section, asset value metrics |

### Market & Pricing
| Feature | Doc | Summary |
|---------|-----|---------|
| Jita Market Pricing | `docs/features/jita-market-pricing.md` | Market orders, asset valuation |
| Stockpile Markers | `docs/features/stockpile-markers.md` | Stockpile targets, deficit tracking, inventory UI |
| Stockpile Multibuy | `docs/features/stockpile-multibuy.md` | Shopping lists, delta calculation, bulk ops |

### Social & Marketplace
| Feature | Doc | Summary |
|---------|-----|---------|
| Contact Marketplace | `docs/features/contact-marketplace.md` | Contacts, permissions, for-sale marketplace |
| Contact Rules | `docs/features/contact-rules.md` | Auto-create contacts, cascade cleanup |
| Discord Notifications | `docs/features/discord-notifications.md` | Discord bot, OAuth linking, event notifications |

### Purchase & Trade
| Feature | Doc | Summary |
|---------|-----|---------|
| Purchases | `docs/features/purchases/` | Purchase transactions, contract workflow |
| Buy Orders | `docs/features/buy-orders/` | Demand tracking, seller demand endpoints |
| Auto-Sell Containers | `docs/features/auto-sell-containers.md` | Auto-sell config, Jita pricing, for-sale sync |
| Auto-Buy | `docs/features/auto-buy.md` | Auto-buy config, buy order management |
| Auto-Fulfill | `docs/features/auto-fulfill.md` | Match buy orders to for-sale listings |
| Contract Sync | `docs/features/contract-sync.md` | ESI contract polling, auto-complete |
| Contract Notifications | `docs/features/contract-created-notification.md` | Discord alerts on contract creation |

### Industry
| Feature | Doc | Summary |
|---------|-----|---------|
| Industry Job Manager | `docs/features/industry-job-manager/` | Skills sync, job tracking, manufacturing calc, job queue |
| Reactions Calculator | `docs/features/reactions-calculator.md` | Moon reactions, batch ME, shopping list |
| Planetary Industry | `docs/features/planetary-industry.md` | PI data, stall detection, profit calc |
| Transportation | `docs/features/transportation.md` | Transport profiles, JF routes, cost calc |

### Infrastructure
| Feature | Doc | Summary |
|---------|-----|---------|
| E2E Testing | `docs/features/e2e-testing.md` | Mock ESI, Playwright, test users |
| Railway Deployment | `docs/features/railway-deployment.md` | PostgreSQL, backend, frontend deployment |
