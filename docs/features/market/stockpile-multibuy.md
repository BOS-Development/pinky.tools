# Stockpile-Aware Multibuy

## Overview

Enhances the reactions calculator shopping list with stockpile awareness. When logged in, users can select a location where they store moon goo. The shopping list then shows what they already have in stock and computes the delta (what still needs to be purchased). Users can also set stockpile targets for all shopping list materials at the selected location.

## Features

- **Location picker** — dropdown populated from user's asset locations (only visible when logged in)
- **In Stock column** — shows quantity of each material at the selected location
- **Delta column** — shows `max(0, needed - in_stock)` with green checkmark when fully stocked
- **Copy Multibuy** — copies the full shopping list (unchanged behavior)
- **Copy Delta** — copies only items where delta > 0, in EVE multibuy format
- **Delta Cost** — shows total cost of remaining materials to purchase
- **Set Stockpile** — opens a dialog to set stockpile targets for all shopping list materials
  - User picks an owner (character or corporation) at the location
  - Quantities pre-filled from shopping list, editable before saving
  - Bulk-saves stockpile markers via existing `/api/stockpiles/upsert` endpoint
- **Persistence** — selected location saved to localStorage across sessions

## Design Decisions

- **Frontend-only**: Uses existing `GET /api/assets/get` and `POST /api/stockpiles/upsert` endpoints. No backend changes needed.
- **Auth-gated**: Location picker only appears when user is logged in. Unauthenticated users see only the standard "Copy Multibuy" button.
- **Aggregation**: Assets are summed across all sources at a structure (personal hangar, containers, deliveries, asset safety, corp hangars, corp containers).
- **Owner selection**: Owners are extracted from assets at the selected structure. This avoids extra API calls for characters/corporations.

## Files

| File | Description |
|------|-------------|
| `frontend/packages/pages/reactions.tsx` | Session awareness, assets fetch, location state |
| `frontend/packages/components/reactions/ShoppingList.tsx` | Location dropdown, delta columns, copy delta, set stockpile button |
| `frontend/packages/components/reactions/StockpileDialog.tsx` | Dialog for setting stockpile targets with editable quantities |
| `frontend/packages/utils/assetAggregation.ts` | Aggregates assets by type_id, extracts unique owners |
| `frontend/packages/utils/__tests__/assetAggregation.test.ts` | Unit tests |
