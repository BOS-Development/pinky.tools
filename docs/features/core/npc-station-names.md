# NPC Station Name Resolution

## Status: Implemented

## Overview

NPC station names (e.g., "Jita IV - Moon 4 - Caldari Navy Assembly Plant") were not displaying on the inventory page. The SDE's `npcStations.yaml` file does not include station names, so stations were inserted into the database with empty names.

## Solution

After the SDE upserts station records, the system queries for any NPC stations with empty names and resolves them using ESI's public `POST /universe/names/` bulk endpoint. This endpoint accepts up to 1000 IDs per call and returns names for stations, solar systems, and other universe entities.

## How It Works

1. SDE updater imports stations from `npcStations.yaml` (IDs, solar system, corporation — but no names)
2. After station upsert, queries `stations` table for rows where `name = '' AND is_npc_station = true`
3. Sends station IDs to ESI `POST /universe/names/` in batches of 1000
4. Updates station names in the database
5. Subsequent asset queries now return proper station names via INNER JOIN

## Key Details

- **Frequency**: Runs during SDE refresh (every 24 hours), not on every asset refresh
- **Efficiency**: Only fetches names for stations that are missing them; once resolved, they persist
- **Authentication**: The `/universe/names/` endpoint is public (no OAuth required)
- **Batching**: IDs are sent in chunks of 1000 (ESI limit per request)

## Key Files

| File | Description |
|------|-------------|
| `internal/client/esiClient.go` | `GetUniverseNames()` — bulk name resolution via ESI |
| `internal/repositories/stations.go` | `GetStationsWithEmptyNames()`, `UpdateNames()` |
| `internal/updaters/sde.go` | Name resolution logic after station upsert |
| `cmd/mock-esi/main.go` | Mock `POST /universe/names/` for E2E tests |
