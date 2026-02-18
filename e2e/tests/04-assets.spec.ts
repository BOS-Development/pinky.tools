import { test, expect } from '@playwright/test';

test.describe('Assets Page', () => {
  // Clear localStorage before each test so tree expansion state is fresh
  test.beforeEach(async ({ page }) => {
    await page.goto('/inventory');
    await page.evaluate(() => localStorage.clear());
  });

  test('assets are populated by background refresh', async ({ page }) => {
    // Assets are automatically updated by the background runner (10s interval in E2E)
    // and also triggered immediately when characters/corporations are added
    await page.goto('/inventory');

    // Jita station should appear with Alice Alpha's assets
    await expect(page.getByText('Jita IV - Moon 4')).toBeVisible({ timeout: 30000 });
  });

  test('displays character assets grouped by station', async ({ page }) => {
    await page.goto('/inventory');

    // Jita station
    await expect(page.getByText('Jita IV - Moon 4')).toBeVisible({ timeout: 10000 });

    // Expand Jita to see hangars
    await page.getByText('Jita IV - Moon 4').click();

    // Expand Personal Hangar to see items
    await page.getByText('Personal Hangar').first().click();

    // Check for asset types from mock ESI
    await expect(page.getByText('Tritanium')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Pyerite')).toBeVisible();
    await expect(page.getByText('Mexallon')).toBeVisible();
  });

  test('displays container with nested assets', async ({ page }) => {
    await page.goto('/inventory');

    // Wait for assets to load, then expand Jita
    await expect(page.getByText('Jita IV - Moon 4')).toBeVisible({ timeout: 10000 });
    await page.getByText('Jita IV - Moon 4').click();

    // Container name from mock ESI
    await expect(page.getByText('Minerals Box')).toBeVisible({ timeout: 5000 });
  });

  test('displays corporation assets under station', async ({ page }) => {
    await page.goto('/inventory');

    // Expand Jita to see corp hangars
    await page.getByText('Jita IV - Moon 4').click();

    // Corp hangars appear as sub-nodes (use .first() since multiple hangars match)
    await expect(page.getByText(/Stargazer Industries/).first()).toBeVisible({ timeout: 5000 });
  });

  test('displays Amarr station assets from Alice Beta', async ({ page }) => {
    await page.goto('/inventory');

    // Amarr station should appear with Alice Beta's assets
    await expect(page.getByText('Amarr VIII')).toBeVisible({ timeout: 10000 });
  });

  test('search filters assets', async ({ page }) => {
    await page.goto('/inventory');

    // Wait for assets to load
    await expect(page.getByText('Jita IV - Moon 4')).toBeVisible({ timeout: 10000 });

    // Use search to filter
    const searchInput = page.getByPlaceholder('Search items...');
    await searchInput.fill('Tritanium');

    // Tritanium should be visible in search results (auto-expanded)
    await expect(page.getByText('Tritanium').first()).toBeVisible({ timeout: 10000 });
  });
});
