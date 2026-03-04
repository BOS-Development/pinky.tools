# Design Review Fixtures

A dedicated E2E seed spec that populates the application with rich, realistic data for visual design reviews.

## How to Run

Start the full E2E stack (backend, frontend, mock ESI, database with seed data), then:

```bash
BASE_URL=http://localhost:3000 DESIGN_REVIEW=true npx playwright test tests/00-design-review-seed.spec.ts --project=chromium
```

The spec is skipped in normal CI runs (`DESIGN_REVIEW=true` env var required).

After seeding, open `http://localhost:3000` in a browser and navigate pages for visual review.

## What Gets Seeded

### Characters

| Character       | ID      | User ID | Role          |
|-----------------|---------|---------|---------------|
| Alice Alpha     | 2001001 | 1001    | Main          |
| Alice Beta      | 2001002 | 1001    | Alt           |
| Bob Bravo       | 2002001 | 1002    | Second user   |
| Charlie Charlie | 2003001 | 1003    | Third user    |

### Corporation

- **Stargazer Industries** (3001001) — linked to Alice Alpha

### Alice Alpha Assets (multi-station, nested containers)

**Jita IV - Moon 4 (60003760):**
- Tritanium x50,000, Pyerite x30,000, Mexallon x10,000, Isogen x5,000
- "Minerals Container" (nested): Nocxium x2,000, Zydrine x500, Megacyte x100
- Ships: Rifter, Thorax, Raven Navy Issue
- PLEX x10, Large Skill Injector x3

**Amarr VIII (60008494):**
- Tritanium x100,000
- "Trade Goods Box" (nested): Nocxium x200, Isogen x150

**Dodixie IX (60011866):**
- Mexallon x25,000, Zydrine x1,000

### Alice Beta Assets

**Amarr VIII:** Rifter x3, Nocxium x5,000, Tritanium x20,000

### Bob Bravo Assets

**Jita IV:** Tritanium x30,000, Rifter x10, Pyerite x15,000

### Corp Assets (Stargazer Industries at Jita)

- Division 1 (Main Hangar): Megacyte x1,000, Zydrine x2,000
- Division 2 (Production Materials): Tritanium x500,000, Pyerite x200,000

### Industry Jobs (Alice Alpha)

| Job  | Activity        | Product           | Runs | Status    |
|------|-----------------|-------------------|------|-----------|
| 1    | Manufacturing   | Rifter            | 10   | active    |
| 2    | Manufacturing   | Thorax            | 5    | active    |
| 3    | Manufacturing   | Rifter            | 20   | delivered |
| 4    | Manufacturing   | Rifter            | 3    | cancelled |
| 5    | ME Research     | Rifter Blueprint  | 1    | active    |
| 6    | TE Research     | Rifter Blueprint  | 1    | delivered |

### Blueprints (Alice Alpha)

- Rifter BPO: ME 10, TE 20, unlimited runs
- Thorax BPC: ME 5, TE 8, 3 runs

### Market Orders

Full set of buy/sell pairs for all 7 minerals, Rifter, Thorax, Raven Navy Issue, and PLEX.

## How to Add More Data

Edit `e2e/tests/00-design-review-seed.spec.ts`:

1. Add new item type IDs to `e2e/seed.sql` (for name resolution)
2. Add assets to the relevant character's array (e.g., `aliceAlphaAssets`)
3. Add station data to `seed.sql` + mock ESI `knownNames` if using a new station
4. PR the changes

The mock ESI admin API supports: `setCharacterAssets`, `setCharacterNames`, `setCharacterIndustryJobs`, `setCharacterBlueprints`, `setCorpAssets`, `setMarketOrders`.
