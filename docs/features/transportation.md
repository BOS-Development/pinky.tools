# Transportation System

## Status: Phase 1 — Implemented

## Overview

The transportation system tracks, costs, and manages item movement between stations. Production plans span multiple stations and stockpile targets exist at different locations — this system bridges the gap by providing transport profiles, JF route management, transport job tracking, and cost calculation.

## Phase 1 Scope

- **Transport Profiles**: Per-ship/character transport configurations (freighter, JF, DST, blockade runner)
- **JF Routes**: User-defined jump freighter cyno routes with LY distance calculation
- **Transport Jobs**: Manual transport job creation with cost calculation and status tracking
- **ESI Route Integration**: Gate route calculation via ESI for jump counts
- **Job Queue Integration**: Transport jobs create corresponding industry job queue entries
- **Trigger Config**: Per-trigger fulfillment preferences (plan_generation, manual)

## Transport Methods

| Method | Route Type | Cost Basis |
|--------|-----------|------------|
| Freighter | Gate-to-gate | Rate/m3/jump + collateral |
| Jump Freighter | Cyno waypoints | Fuel cost + collateral |
| DST | Gate-to-gate | Rate/m3/jump + collateral |
| Blockade Runner | Gate-to-gate | Rate/m3/jump + collateral |

## Fulfillment Types

- **Self Haul**: Player moves items using their own ships (detailed cost calculation)
- **Courier Contract**: Public courier contract (flat rate per m3 + collateral %)
- **Contact Haul**: Trusted hauler via contacts (flat rate per m3 + collateral %)

## Data Model

### Tables (6 new + 1 altered)

1. **transport_profiles** — Per-ship transport configurations
   - cargo_m3, rate_per_m3_per_jump, collateral_rate, collateral_price_basis
   - fuel_type_id, fuel_per_ly, fuel_conservation_level (JF only)
   - route_preference (shortest/secure/insecure), is_default per method

2. **jf_routes** — Jump freighter routes
   - origin_system_id, destination_system_id, total_distance_ly

3. **jf_route_waypoints** — Ordered cyno stops per route
   - sequence, system_id, distance_ly (calculated from SDE coordinates)

4. **transport_jobs** — Transport job instances
   - origin/destination station+system, transport_method, route_preference
   - total_volume_m3, total_collateral, estimated_cost, jumps, distance_ly
   - fulfillment_type, status (planned → in_transit → delivered | cancelled)

5. **transport_job_items** — Items in a transport job
   - type_id, quantity, volume_m3, estimated_value

6. **transport_trigger_config** — Per-trigger fulfillment preferences
   - trigger_type PK, default_fulfillment, allowed_fulfillments[], courier rates

7. **industry_job_queue** — Added transport_job_id FK column

### Solar System Coordinates

Added x, y, z DOUBLE PRECISION columns to `solar_systems` table. Populated from CCP SDE `position` data (coordinates in meters). Used for JF light-year distance calculation:

```
distance_ly = sqrt((x2-x1)² + (y2-y1)² + (z2-z1)²) / 9.461e+15
```

## Cost Formulas

### Gate Transport (Freighter/DST/Blockade Runner)
```
trips = ceil(totalVolume / cargoM3)
cost = ((volume × ratePerM3PerJump × jumps) + (collateral × collateralRate)) × trips
```

### Jump Freighter
```
per leg: fuel_units = ceil(fuelPerLY × distanceLY × (1 - fuelConservationLevel × 0.10))
totalFuel = sum(fuel_units per leg)
fuelCost = totalFuel × isotope_price
trips = ceil(totalVolume / cargoM3)
cost = (fuelCost + (collateral × collateralRate)) × trips
```

### Courier/Contact (flat rate)
```
cost = (volume × ratePerM3) + (collateral × collateralRate)
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /v1/transport/profiles | List user's transport profiles |
| POST | /v1/transport/profiles | Create transport profile |
| PUT | /v1/transport/profiles/{id} | Update transport profile |
| DELETE | /v1/transport/profiles/{id} | Delete transport profile |
| GET | /v1/transport/jf-routes | List user's JF routes |
| POST | /v1/transport/jf-routes | Create JF route with waypoints |
| PUT | /v1/transport/jf-routes/{id} | Update JF route |
| DELETE | /v1/transport/jf-routes/{id} | Delete JF route |
| GET | /v1/transport/jobs | List user's transport jobs |
| POST | /v1/transport/jobs | Create transport job (calculates cost) |
| POST | /v1/transport/jobs/{id}/status | Update job status |
| GET | /v1/transport/route | ESI gate route proxy |
| GET | /v1/transport/trigger-config | Get trigger configs |
| PUT | /v1/transport/trigger-config | Upsert trigger config |

## File Structure

### Backend
- `internal/models/models.go` — TransportProfile, JFRoute, JFRouteWaypoint, TransportJob, TransportJobItem, TransportTriggerConfig
- `internal/repositories/transportProfiles.go` — CRUD with default management
- `internal/repositories/jfRoutes.go` — CRUD with waypoint/distance calculation
- `internal/repositories/transportJobs.go` — CRUD with status transitions
- `internal/repositories/transportTriggerConfig.go` — Upsert on trigger_type
- `internal/calculator/transport.go` — Cost calculation functions
- `internal/controllers/transportation.go` — HTTP handlers (14 routes)
- `internal/client/esiClient.go` — GetRoute method for ESI route API

### Frontend
- `frontend/pages/transport.tsx` — Page router entry
- `frontend/packages/pages/transport.tsx` — Page with tabs (Jobs, Profiles, JF Routes)
- `frontend/packages/components/transport/` — TransportProfilesList, TransportProfileDialog, JFRoutesList, JFRouteDialog, TransportJobsList, TransportJobDialog
- `frontend/pages/api/transport/` — API proxy routes (8 files)

## Key Decisions

1. **Pre-designated JF routes**: Users configure cyno waypoints upfront; distance calculated from SDE coordinates at create time
2. **Dual cost model**: Self-haul uses detailed profile-based calculation; courier/contact uses flat rates
3. **Multi-profile support**: Multiple profiles per transport method, one default per method
4. **Job queue integration**: Transport jobs create queue entries with activity='transport'
5. **Status machine**: planned → in_transit → delivered, or planned/in_transit → cancelled
6. **Collateral price basis**: buy, sell, or split — same pattern as reactions calculator

## Future Phases

- Phase 2: Production plan integration (auto-generate transport jobs from plans)
- Phase 3: Contract tracking and cost reconciliation
- Phase 4: Multi-leg route optimization
- Phase 5: Fleet coordination and scheduling
