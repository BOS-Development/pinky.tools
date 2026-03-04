/**
 * 00-design-review-seed.spec.ts
 *
 * Seeds the E2E environment with rich, realistic fixture data for design reviews.
 * Skipped in normal CI — run explicitly with DESIGN_REVIEW=true.
 *
 * Usage:
 *   BASE_URL=http://localhost:3000 DESIGN_REVIEW=true npx playwright test tests/00-design-review-seed.spec.ts --project=chromium
 */
import { test, expect } from '@playwright/test';
import {
  setCharacterAssets,
  setCharacterNames,
  setCharacterIndustryJobs,
  setCharacterBlueprints,
  setCorpAssets,
  setMarketOrders,
  type Asset,
  type NameEntry,
  type IndustryJob,
  type BlueprintEntry,
  type MarketOrder,
} from '../helpers/mock-esi';

// ---------------------------------------------------------------------------
// Skip guard — only runs when DESIGN_REVIEW=true
// ---------------------------------------------------------------------------
test.skip(process.env.DESIGN_REVIEW !== 'true', 'Design review seed — run with DESIGN_REVIEW=true');

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

// Characters
const ALICE_ALPHA_ID = 2001001;
const ALICE_BETA_ID = 2001002;
const BOB_BRAVO_ID = 2002001;
const CHARLIE_ID = 2003001;

// Stations
const JITA = 60003760;
const AMARR = 60008494;
const DODIXIE = 60011866;

// Item type IDs
const TRITANIUM = 34;
const PYERITE = 35;
const MEXALLON = 36;
const ISOGEN = 37;
const NOCXIUM = 38;
const ZYDRINE = 39;
const MEGACYTE = 40;
const RIFTER = 587;
const THORAX = 24692;
const PLEX = 44992;
const SKILL_INJECTOR = 40520;
const CONTAINER = 9999001; // Medium Standard Container
const OFFICE = 27;
const RAVEN_NAVY = 11399;

// Blueprint type IDs
const RIFTER_BP = 787;
const THORAX_BP = 24694;

// Corp
const STARGAZER_CORP = 3001001;

// ---------------------------------------------------------------------------
// Asset fixtures
// ---------------------------------------------------------------------------

const aliceAlphaAssets: Asset[] = [
  // === Jita — Personal Hangar ===
  // Minerals
  { item_id: 800001, is_blueprint_copy: false, is_singleton: false, location_flag: 'Hangar', location_id: JITA, location_type: 'station', quantity: 50000, type_id: TRITANIUM },
  { item_id: 800002, is_blueprint_copy: false, is_singleton: false, location_flag: 'Hangar', location_id: JITA, location_type: 'station', quantity: 30000, type_id: PYERITE },
  { item_id: 800003, is_blueprint_copy: false, is_singleton: false, location_flag: 'Hangar', location_id: JITA, location_type: 'station', quantity: 10000, type_id: MEXALLON },
  { item_id: 800004, is_blueprint_copy: false, is_singleton: false, location_flag: 'Hangar', location_id: JITA, location_type: 'station', quantity: 5000, type_id: ISOGEN },
  // Container: "Minerals Container" (named via character-names API)
  { item_id: 800010, is_blueprint_copy: false, is_singleton: true, location_flag: 'Hangar', location_id: JITA, location_type: 'station', quantity: 1, type_id: CONTAINER },
  { item_id: 800011, is_blueprint_copy: false, is_singleton: false, location_flag: 'Hangar', location_id: 800010, location_type: 'item', quantity: 2000, type_id: NOCXIUM },
  { item_id: 800012, is_blueprint_copy: false, is_singleton: false, location_flag: 'Hangar', location_id: 800010, location_type: 'item', quantity: 500, type_id: ZYDRINE },
  { item_id: 800013, is_blueprint_copy: false, is_singleton: false, location_flag: 'Hangar', location_id: 800010, location_type: 'item', quantity: 100, type_id: MEGACYTE },
  // Ships
  { item_id: 800020, is_blueprint_copy: false, is_singleton: true, location_flag: 'Hangar', location_id: JITA, location_type: 'station', quantity: 1, type_id: RIFTER },
  { item_id: 800021, is_blueprint_copy: false, is_singleton: true, location_flag: 'Hangar', location_id: JITA, location_type: 'station', quantity: 1, type_id: THORAX },
  { item_id: 800022, is_blueprint_copy: false, is_singleton: true, location_flag: 'Hangar', location_id: JITA, location_type: 'station', quantity: 1, type_id: RAVEN_NAVY },
  // High-value items
  { item_id: 800030, is_blueprint_copy: false, is_singleton: false, location_flag: 'Hangar', location_id: JITA, location_type: 'station', quantity: 10, type_id: PLEX },
  { item_id: 800031, is_blueprint_copy: false, is_singleton: false, location_flag: 'Hangar', location_id: JITA, location_type: 'station', quantity: 3, type_id: SKILL_INJECTOR },

  // === Amarr — Personal Hangar ===
  { item_id: 800100, is_blueprint_copy: false, is_singleton: false, location_flag: 'Hangar', location_id: AMARR, location_type: 'station', quantity: 100000, type_id: TRITANIUM },
  // Container: "Trade Goods Box" (named via character-names API)
  { item_id: 800110, is_blueprint_copy: false, is_singleton: true, location_flag: 'Hangar', location_id: AMARR, location_type: 'station', quantity: 1, type_id: CONTAINER },
  { item_id: 800111, is_blueprint_copy: false, is_singleton: false, location_flag: 'Hangar', location_id: 800110, location_type: 'item', quantity: 200, type_id: NOCXIUM },
  { item_id: 800112, is_blueprint_copy: false, is_singleton: false, location_flag: 'Hangar', location_id: 800110, location_type: 'item', quantity: 150, type_id: ISOGEN },

  // === Dodixie — Personal Hangar ===
  { item_id: 800200, is_blueprint_copy: false, is_singleton: false, location_flag: 'Hangar', location_id: DODIXIE, location_type: 'station', quantity: 25000, type_id: MEXALLON },
  { item_id: 800201, is_blueprint_copy: false, is_singleton: false, location_flag: 'Hangar', location_id: DODIXIE, location_type: 'station', quantity: 1000, type_id: ZYDRINE },
];

const aliceAlphaNames: NameEntry[] = [
  { item_id: 800010, name: 'Minerals Container' },
  { item_id: 800110, name: 'Trade Goods Box' },
];

const aliceBetaAssets: Asset[] = [
  // Amarr — Personal Hangar
  { item_id: 810001, is_blueprint_copy: false, is_singleton: true, location_flag: 'Hangar', location_id: AMARR, location_type: 'station', quantity: 3, type_id: RIFTER },
  { item_id: 810002, is_blueprint_copy: false, is_singleton: false, location_flag: 'Hangar', location_id: AMARR, location_type: 'station', quantity: 5000, type_id: NOCXIUM },
  { item_id: 810003, is_blueprint_copy: false, is_singleton: false, location_flag: 'Hangar', location_id: AMARR, location_type: 'station', quantity: 20000, type_id: TRITANIUM },
];

const bobBravoAssets: Asset[] = [
  // Jita — Personal Hangar
  { item_id: 820001, is_blueprint_copy: false, is_singleton: false, location_flag: 'Hangar', location_id: JITA, location_type: 'station', quantity: 30000, type_id: TRITANIUM },
  { item_id: 820002, is_blueprint_copy: false, is_singleton: true, location_flag: 'Hangar', location_id: JITA, location_type: 'station', quantity: 10, type_id: RIFTER },
  { item_id: 820003, is_blueprint_copy: false, is_singleton: false, location_flag: 'Hangar', location_id: JITA, location_type: 'station', quantity: 15000, type_id: PYERITE },
];

// ---------------------------------------------------------------------------
// Corp asset fixtures
// ---------------------------------------------------------------------------

const stargazerCorpAssets: Asset[] = [
  // Office at Jita (required parent for corp hangar items)
  { item_id: 850000, is_blueprint_copy: false, is_singleton: true, location_flag: 'OfficeFolder', location_id: JITA, location_type: 'item', quantity: 1, type_id: OFFICE },
  // Division 1 (Main Hangar)
  { item_id: 850001, is_blueprint_copy: false, is_singleton: false, location_flag: 'CorpSAG1', location_id: 850000, location_type: 'item', quantity: 1000, type_id: MEGACYTE },
  { item_id: 850002, is_blueprint_copy: false, is_singleton: false, location_flag: 'CorpSAG1', location_id: 850000, location_type: 'item', quantity: 2000, type_id: ZYDRINE },
  // Division 2 (Production Materials)
  { item_id: 850003, is_blueprint_copy: false, is_singleton: false, location_flag: 'CorpSAG2', location_id: 850000, location_type: 'item', quantity: 500000, type_id: TRITANIUM },
  { item_id: 850004, is_blueprint_copy: false, is_singleton: false, location_flag: 'CorpSAG2', location_id: 850000, location_type: 'item', quantity: 200000, type_id: PYERITE },
];

// ---------------------------------------------------------------------------
// Industry job fixtures — varied states
// ---------------------------------------------------------------------------

const aliceAlphaJobs: IndustryJob[] = [
  // Active manufacturing: Rifter, 10 runs
  {
    job_id: 600001, installer_id: ALICE_ALPHA_ID,
    facility_id: JITA, station_id: JITA,
    activity_id: 1, blueprint_id: 700001, blueprint_type_id: RIFTER_BP,
    blueprint_location_id: JITA, output_location_id: JITA,
    runs: 10, cost: 1500000, product_type_id: RIFTER,
    status: 'active', duration: 3600,
    start_date: '2026-03-01T00:00:00Z', end_date: '2026-03-01T01:00:00Z',
  },
  // Active manufacturing: Thorax, 5 runs
  {
    job_id: 600002, installer_id: ALICE_ALPHA_ID,
    facility_id: JITA, station_id: JITA,
    activity_id: 1, blueprint_id: 700010, blueprint_type_id: THORAX_BP,
    blueprint_location_id: JITA, output_location_id: JITA,
    runs: 5, cost: 8000000, product_type_id: THORAX,
    status: 'active', duration: 7200,
    start_date: '2026-03-01T02:00:00Z', end_date: '2026-03-01T04:00:00Z',
  },
  // Delivered manufacturing: Rifter, 20 runs (completed)
  {
    job_id: 600003, installer_id: ALICE_ALPHA_ID,
    facility_id: JITA, station_id: JITA,
    activity_id: 1, blueprint_id: 700001, blueprint_type_id: RIFTER_BP,
    blueprint_location_id: JITA, output_location_id: JITA,
    runs: 20, cost: 3000000, product_type_id: RIFTER,
    status: 'delivered', duration: 3600,
    start_date: '2026-02-28T00:00:00Z', end_date: '2026-02-28T01:00:00Z',
  },
  // Cancelled manufacturing: Rifter, 3 runs
  {
    job_id: 600004, installer_id: ALICE_ALPHA_ID,
    facility_id: JITA, station_id: JITA,
    activity_id: 1, blueprint_id: 700001, blueprint_type_id: RIFTER_BP,
    blueprint_location_id: JITA, output_location_id: JITA,
    runs: 3, cost: 450000, product_type_id: RIFTER,
    status: 'cancelled', duration: 3600,
    start_date: '2026-02-27T00:00:00Z', end_date: '2026-02-27T01:00:00Z',
  },
  // Active ME research on Rifter BP
  {
    job_id: 600005, installer_id: ALICE_ALPHA_ID,
    facility_id: JITA, station_id: JITA,
    activity_id: 4, blueprint_id: 700001, blueprint_type_id: RIFTER_BP,
    blueprint_location_id: JITA, output_location_id: JITA,
    runs: 1, cost: 500000, product_type_id: RIFTER_BP,
    status: 'active', duration: 10800,
    start_date: '2026-03-01T00:00:00Z', end_date: '2026-03-01T03:00:00Z',
  },
  // Delivered TE research on Rifter BP
  {
    job_id: 600006, installer_id: ALICE_ALPHA_ID,
    facility_id: JITA, station_id: JITA,
    activity_id: 3, blueprint_id: 700001, blueprint_type_id: RIFTER_BP,
    blueprint_location_id: JITA, output_location_id: JITA,
    runs: 1, cost: 250000, product_type_id: RIFTER_BP,
    status: 'delivered', duration: 7200,
    start_date: '2026-02-28T00:00:00Z', end_date: '2026-02-28T02:00:00Z',
  },
];

// ---------------------------------------------------------------------------
// Blueprint fixtures
// ---------------------------------------------------------------------------

const aliceAlphaBlueprints: BlueprintEntry[] = [
  // Rifter BPO — ME 10, TE 20, unlimited runs
  { item_id: 700001, type_id: RIFTER_BP, location_id: JITA, location_flag: 'Hangar', quantity: -1, material_efficiency: 10, time_efficiency: 20, runs: -1 },
  // Thorax BPC — ME 5, TE 8, 3 runs remaining
  { item_id: 700010, type_id: THORAX_BP, location_id: JITA, location_flag: 'Hangar', quantity: -2, material_efficiency: 5, time_efficiency: 8, runs: 3 },
];

// ---------------------------------------------------------------------------
// Market order fixtures (extends default set with more variety)
// ---------------------------------------------------------------------------

const richMarketOrders: MarketOrder[] = [
  // Tritanium — high volume
  { order_id: 1, type_id: TRITANIUM, location_id: JITA, volume_total: 10000000, volume_remain: 5000000, min_volume: 1, price: 6.00, is_buy_order: false, duration: 90, issued: '2026-02-01T00:00:00Z', range: 'station' },
  { order_id: 2, type_id: TRITANIUM, location_id: JITA, volume_total: 10000000, volume_remain: 5000000, min_volume: 1, price: 5.50, is_buy_order: true, duration: 90, issued: '2026-02-01T00:00:00Z', range: 'station' },
  // Pyerite
  { order_id: 3, type_id: PYERITE, location_id: JITA, volume_total: 5000000, volume_remain: 2000000, min_volume: 1, price: 11.50, is_buy_order: false, duration: 90, issued: '2026-02-01T00:00:00Z', range: 'station' },
  { order_id: 4, type_id: PYERITE, location_id: JITA, volume_total: 5000000, volume_remain: 2000000, min_volume: 1, price: 10.00, is_buy_order: true, duration: 90, issued: '2026-02-01T00:00:00Z', range: 'station' },
  // Mexallon
  { order_id: 5, type_id: MEXALLON, location_id: JITA, volume_total: 1000000, volume_remain: 500000, min_volume: 1, price: 75.00, is_buy_order: false, duration: 90, issued: '2026-02-01T00:00:00Z', range: 'station' },
  { order_id: 6, type_id: MEXALLON, location_id: JITA, volume_total: 1000000, volume_remain: 500000, min_volume: 1, price: 70.00, is_buy_order: true, duration: 90, issued: '2026-02-01T00:00:00Z', range: 'station' },
  // Isogen
  { order_id: 11, type_id: ISOGEN, location_id: JITA, volume_total: 500000, volume_remain: 200000, min_volume: 1, price: 55.00, is_buy_order: false, duration: 90, issued: '2026-02-01T00:00:00Z', range: 'station' },
  { order_id: 12, type_id: ISOGEN, location_id: JITA, volume_total: 500000, volume_remain: 200000, min_volume: 1, price: 50.00, is_buy_order: true, duration: 90, issued: '2026-02-01T00:00:00Z', range: 'station' },
  // Nocxium
  { order_id: 13, type_id: NOCXIUM, location_id: JITA, volume_total: 200000, volume_remain: 80000, min_volume: 1, price: 900.00, is_buy_order: false, duration: 90, issued: '2026-02-01T00:00:00Z', range: 'station' },
  { order_id: 14, type_id: NOCXIUM, location_id: JITA, volume_total: 200000, volume_remain: 80000, min_volume: 1, price: 800.00, is_buy_order: true, duration: 90, issued: '2026-02-01T00:00:00Z', range: 'station' },
  // Zydrine
  { order_id: 15, type_id: ZYDRINE, location_id: JITA, volume_total: 50000, volume_remain: 20000, min_volume: 1, price: 1200.00, is_buy_order: false, duration: 90, issued: '2026-02-01T00:00:00Z', range: 'station' },
  { order_id: 16, type_id: ZYDRINE, location_id: JITA, volume_total: 50000, volume_remain: 20000, min_volume: 1, price: 1000.00, is_buy_order: true, duration: 90, issued: '2026-02-01T00:00:00Z', range: 'station' },
  // Megacyte
  { order_id: 17, type_id: MEGACYTE, location_id: JITA, volume_total: 10000, volume_remain: 4000, min_volume: 1, price: 55000.00, is_buy_order: false, duration: 90, issued: '2026-02-01T00:00:00Z', range: 'station' },
  { order_id: 18, type_id: MEGACYTE, location_id: JITA, volume_total: 10000, volume_remain: 4000, min_volume: 1, price: 50000.00, is_buy_order: true, duration: 90, issued: '2026-02-01T00:00:00Z', range: 'station' },
  // Rifter
  { order_id: 7, type_id: RIFTER, location_id: JITA, volume_total: 100, volume_remain: 50, min_volume: 1, price: 600000, is_buy_order: false, duration: 90, issued: '2026-02-01T00:00:00Z', range: 'region' },
  { order_id: 8, type_id: RIFTER, location_id: JITA, volume_total: 100, volume_remain: 50, min_volume: 1, price: 500000, is_buy_order: true, duration: 90, issued: '2026-02-01T00:00:00Z', range: 'region' },
  // Thorax
  { order_id: 19, type_id: THORAX, location_id: JITA, volume_total: 50, volume_remain: 25, min_volume: 1, price: 11000000, is_buy_order: false, duration: 90, issued: '2026-02-01T00:00:00Z', range: 'region' },
  { order_id: 20, type_id: THORAX, location_id: JITA, volume_total: 50, volume_remain: 25, min_volume: 1, price: 9500000, is_buy_order: true, duration: 90, issued: '2026-02-01T00:00:00Z', range: 'region' },
  // Raven Navy Issue
  { order_id: 9, type_id: RAVEN_NAVY, location_id: JITA, volume_total: 10, volume_remain: 5, min_volume: 1, price: 520000000, is_buy_order: false, duration: 90, issued: '2026-02-01T00:00:00Z', range: 'region' },
  { order_id: 10, type_id: RAVEN_NAVY, location_id: JITA, volume_total: 10, volume_remain: 5, min_volume: 1, price: 500000000, is_buy_order: true, duration: 90, issued: '2026-02-01T00:00:00Z', range: 'region' },
  // PLEX
  { order_id: 21, type_id: PLEX, location_id: JITA, volume_total: 1000, volume_remain: 500, min_volume: 1, price: 3700000, is_buy_order: false, duration: 90, issued: '2026-02-01T00:00:00Z', range: 'region' },
  { order_id: 22, type_id: PLEX, location_id: JITA, volume_total: 1000, volume_remain: 500, min_volume: 1, price: 3500000, is_buy_order: true, duration: 90, issued: '2026-02-01T00:00:00Z', range: 'region' },
];

// ---------------------------------------------------------------------------
// Tests — each test seeds a category and asserts success
// ---------------------------------------------------------------------------

test.describe('Design Review Seed', () => {
  test('seed mock ESI data before adding characters', async () => {
    // Set rich asset data in mock ESI BEFORE adding characters,
    // so the immediate asset sync triggered by add-character picks up our data.
    await setCharacterAssets(ALICE_ALPHA_ID, aliceAlphaAssets);
    await setCharacterNames(ALICE_ALPHA_ID, aliceAlphaNames);
    await setCharacterAssets(ALICE_BETA_ID, aliceBetaAssets);
    await setCharacterAssets(BOB_BRAVO_ID, bobBravoAssets);
    await setCorpAssets(STARGAZER_CORP, stargazerCorpAssets);
    await setCharacterIndustryJobs(ALICE_ALPHA_ID, aliceAlphaJobs);
    await setCharacterBlueprints(ALICE_ALPHA_ID, aliceAlphaBlueprints);
    await setMarketOrders(richMarketOrders);
    // If any of the above threw, the test fails — no explicit assertion needed
    // since the helpers throw on non-2xx responses
  });

  test('add all characters', async ({ page }) => {
    const characters = [
      { userId: 1001, characterId: ALICE_ALPHA_ID, characterName: 'Alice Alpha' },
      { userId: 1001, characterId: ALICE_BETA_ID, characterName: 'Alice Beta' },
      { userId: 1002, characterId: BOB_BRAVO_ID, characterName: 'Bob Bravo' },
      { userId: 1003, characterId: CHARLIE_ID, characterName: 'Charlie Charlie' },
    ];

    for (const char of characters) {
      const response = await page.request.post('/api/e2e/add-character', { data: char });
      expect(response.status()).toBe(200);
    }
  });

  test('add Stargazer Industries corporation', async ({ page }) => {
    const response = await page.request.post('/api/e2e/add-corporation', {
      data: { userId: 1001, characterId: ALICE_ALPHA_ID, characterName: 'Alice Alpha' },
    });
    expect(response.status()).toBe(200);
  });

  test('verify assets loaded at Jita', async ({ page }) => {
    // Wait for the background asset runner to sync (fires every 10s in E2E).
    // Navigate to inventory and confirm Jita station appears.
    await page.goto('/inventory');
    await expect(page.getByText('Jita IV - Moon 4')).toBeVisible({ timeout: 30000 });
  });

  test('verify assets loaded at Amarr', async ({ page }) => {
    await page.goto('/inventory');
    await expect(page.getByText('Amarr VIII')).toBeVisible({ timeout: 30000 });
  });

  test('verify assets loaded at Dodixie', async ({ page }) => {
    await page.goto('/inventory');
    await expect(page.getByText('Dodixie IX')).toBeVisible({ timeout: 30000 });
  });

  test('verify industry jobs synced', async ({ page }) => {
    await page.goto('/industry');
    // Wait for the active Rifter job to appear (background runner syncs every 10s)
    await expect(async () => {
      await page.reload();
      await expect(page.getByText('Rifter', { exact: true })).toBeVisible({ timeout: 5000 });
    }).toPass({ timeout: 60000 });
  });
});
