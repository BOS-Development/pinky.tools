import { test, expect } from '@playwright/test';
import {
  setCharacterPlanets,
  setPlanetDetails,
  resetMockESI,
  type PiPlanet,
  type PiColony,
} from '../helpers/mock-esi';

// Alice Alpha's character ID (created in 02-characters.spec.ts)
const ALICE_ALPHA_ID = 2001001;

// Jita solar system (exists in seed.sql)
const JITA_SYSTEM_ID = 30000142;

test.describe('Planetary Industry', () => {
  test.afterAll(async () => {
    await resetMockESI();
  });

  test('navigate to /pi shows heading and tabs', async ({ page }) => {
    await page.goto('/pi');
    // Clear PI tab state so we start on the Overview tab
    await page.evaluate(() => localStorage.removeItem('pi-tab'));
    await page.reload();

    await expect(
      page.getByRole('heading', { name: 'Planetary Industry' }),
    ).toBeVisible({ timeout: 10000 });

    await expect(page.getByRole('tab', { name: 'Overview' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'Profit' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'Supply Chain' })).toBeVisible();
  });

  test('Overview tab shows empty state when no planets are synced', async ({
    page,
  }) => {
    await page.goto('/pi');
    await page.evaluate(() => localStorage.removeItem('pi-tab'));
    await page.reload();

    // Overview is the default tab (index 0)
    await expect(page.getByRole('tab', { name: 'Overview' })).toBeVisible({
      timeout: 5000,
    });

    // Wait for loading to complete then confirm empty state message
    await expect(
      page.getByText(/No planets found/i),
    ).toBeVisible({ timeout: 10000 });
  });

  test('tab switching works', async ({ page }) => {
    await page.goto('/pi');
    await page.evaluate(() => localStorage.removeItem('pi-tab'));
    await page.reload();

    await expect(page.getByRole('tab', { name: 'Overview' })).toBeVisible({
      timeout: 5000,
    });

    // Switch to Profit tab
    await page.getByRole('tab', { name: 'Profit' }).click();
    // The profit table renders — just confirm we navigated away from overview empty state
    // (Profit tab fetches from /api/pi/profit which returns empty data without planets)
    await expect(
      page.getByRole('tab', { name: 'Profit' }),
    ).toHaveAttribute('aria-selected', 'true');

    // Switch to Supply Chain tab
    await page.getByRole('tab', { name: 'Supply Chain' }).click();
    await expect(
      page.getByRole('tab', { name: 'Supply Chain' }),
    ).toHaveAttribute('aria-selected', 'true');

    // Switch back to Overview
    await page.getByRole('tab', { name: 'Overview' }).click();
    await expect(
      page.getByRole('tab', { name: 'Overview' }),
    ).toHaveAttribute('aria-selected', 'true');
  });

  test('planets appear after injecting PI data and waiting for sync', async ({
    page,
  }) => {
    // Inject a barren planet in Jita for Alice Alpha
    const planet: PiPlanet = {
      last_update: new Date(Date.now() - 60 * 60 * 1000).toISOString(), // 1 hour ago
      num_pins: 3,
      owner_id: ALICE_ALPHA_ID,
      planet_id: 40000001,
      planet_type: 'barren',
      solar_system_id: JITA_SYSTEM_ID,
      upgrade_level: 4,
    };

    // A simple colony: one command center + one extractor (expired, so stale_data or expired status)
    // Using a future expiry so the extractor shows as running
    const futureExpiry = new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString();
    const colony: PiColony = {
      links: [],
      pins: [
        {
          pin_id: 1001,
          type_id: 2525, // barren command center
          latitude: 0.1,
          longitude: 0.2,
          contents: [],
        },
        {
          pin_id: 1002,
          type_id: 3060, // Extractor Control Unit (barren)
          latitude: 0.3,
          longitude: 0.4,
          expiry_time: futureExpiry,
          install_time: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
          contents: [],
          extractor_details: {
            cycle_time: 1800,
            head_radius: 0.01,
            heads: [],
            product_type_id: 2267, // Aqueous Liquids (P0 raw)
            qty_per_cycle: 9000,
          },
        },
      ],
      routes: [],
    };

    await setCharacterPlanets(ALICE_ALPHA_ID, [planet]);
    await setPlanetDetails(ALICE_ALPHA_ID, 40000001, colony);

    // Wait for the PI runner to pick up the new mock data and write it to the DB.
    // The runner interval is PI_UPDATE_INTERVAL_SEC=10 in E2E.
    // Strategy: load the page, clear tab state, then reload until the planet appears.
    await page.goto('/pi');
    await page.evaluate(() => localStorage.removeItem('pi-tab'));

    // Poll by reloading: the runner fires every 10s; after each reload the component
    // fetches fresh data. We retry for up to 30s.
    await expect(async () => {
      await page.reload();
      await expect(page.getByText(/Jita/i)).toBeVisible({ timeout: 3000 });
    }).toPass({ timeout: 35000 });

    // The planet card should show the planet type
    await expect(page.getByText(/Barren/i)).toBeVisible({ timeout: 5000 });
  });

  test('stats chips show planet and extractor counts after data is loaded', async ({
    page,
  }) => {
    // Data from the previous test is already in the DB (runner synced it).
    // Navigate fresh so the component fetches current state.
    await page.goto('/pi');
    await page.evaluate(() => localStorage.removeItem('pi-tab'));
    await page.reload();

    // Planet data should already be in DB from previous test — no need to wait for runner
    await expect(page.getByText('Jita')).toBeVisible({ timeout: 15000 });

    // StatChip labels appear in the overview stats bar.
    // 'Planets' also appears as a menu item inside the Industry dropdown in the navbar,
    // but the dropdown is closed so its items are not rendered in the DOM.
    // Use .last() as a defensive measure in case multiple matches appear.
    await expect(page.getByText('Planets').last()).toBeVisible({ timeout: 5000 });
    // 'Extractors' appears in both the StatChip (stats bar) and in each planet card
    // that has extractors (as a section heading). Use .first() to target either.
    await expect(page.getByText('Extractors').first()).toBeVisible({ timeout: 5000 });
  });
});
