import { test, expect } from '@playwright/test';

test.describe('Stockpiles', () => {
  // Clear localStorage before each test so tree expansion state is fresh
  test.beforeEach(async ({ page }) => {
    await page.goto('/inventory');
    await page.evaluate(() => localStorage.clear());
  });

  test('stockpiles page shows empty state initially', async ({ page }) => {
    await page.goto('/stockpiles');

    await expect(page.getByText('Stockpiles Needing Replenishment')).toBeVisible({ timeout: 10000 });
    // No stockpile markers set yet
    await expect(page.getByText('No stockpiles need replenishment')).toBeVisible();
  });

  test('set stockpile marker on character asset', async ({ page }) => {
    await page.goto('/inventory');

    // Expand Jita station to see hangars
    await expect(page.getByText('Jita IV - Moon 4')).toBeVisible({ timeout: 10000 });
    await page.getByText('Jita IV - Moon 4').click();

    // Expand Personal Hangar to see items
    await page.getByText('Personal Hangar').first().click();
    await expect(page.getByText('Tritanium')).toBeVisible({ timeout: 5000 });

    // Find the Tritanium row and click the first button (set stockpile)
    const tritaniumRow = page.getByRole('row').filter({ hasText: 'Tritanium' }).first();
    await tritaniumRow.getByRole('button').first().click();

    // Dialog should appear with "Set Stockpile Marker"
    await expect(page.getByText('Set Stockpile Marker')).toBeVisible({ timeout: 5000 });

    // Dialog shows item name and current quantity (scope to dialog to avoid table matches)
    const dialog = page.getByRole('dialog');
    await expect(dialog.getByText(/Tritanium/)).toBeVisible();
    await expect(dialog.getByText(/50,000/)).toBeVisible();

    // Enter desired quantity (higher than current 50000 to create a deficit)
    const desiredQtyInput = page.getByLabel(/Desired Quantity/i);
    await desiredQtyInput.clear();
    await desiredQtyInput.fill('100000');

    // Click Save
    await page.getByRole('button', { name: /Save/i }).click();

    // Dialog should close
    await expect(page.getByText('Set Stockpile Marker')).not.toBeVisible({ timeout: 5000 });

    // Verify inline delta is displayed on the Tritanium row
    // The stockpile column shows delta / desired (e.g., "-50,000 / 100,000")
    await expect(tritaniumRow.getByText('100,000')).toBeVisible({ timeout: 5000 });
  });

  test('stockpiles page shows deficit with correct values', async ({ page }) => {
    await page.goto('/stockpiles');

    // Should show the Tritanium deficit
    await expect(page.getByText('Tritanium')).toBeVisible({ timeout: 10000 });

    // Summary cards should show values
    await expect(page.getByText('Items Below Target')).toBeVisible();
    await expect(page.getByText('Total Deficit')).toBeVisible();
    await expect(page.getByText('Total Cost (ISK)')).toBeVisible();

    // Verify the deficit table row has current and target quantities
    const tritaniumRow = page.getByRole('row').filter({ hasText: 'Tritanium' });
    await expect(tritaniumRow).toBeVisible();
    await expect(tritaniumRow.getByText('100,000')).toBeVisible();

    // Verify location info appears (use specific text to avoid matching both structure and system columns)
    await expect(tritaniumRow.getByText(/Jita IV/)).toBeVisible();
  });

  test('set stockpile marker on corporation asset', async ({ page }) => {
    await page.goto('/inventory');

    // Expand Jita station
    await expect(page.getByText('Jita IV - Moon 4')).toBeVisible({ timeout: 10000 });
    await page.getByText('Jita IV - Moon 4').click();

    // Expand Stargazer Industries - Main Hangar (corp division 1 has Tritanium x100,000)
    await expect(page.getByText(/Stargazer Industries - Main Hangar/).first()).toBeVisible({ timeout: 5000 });
    await page.getByText(/Stargazer Industries - Main Hangar/).first().click();

    // Find the Tritanium row in the corp hangar and set a stockpile marker
    // Use last() since the personal hangar Tritanium row may also be in the DOM
    const corpTritaniumRow = page.getByRole('row').filter({ hasText: 'Tritanium' }).last();
    await corpTritaniumRow.getByRole('button').first().click();

    // Dialog should appear
    await expect(page.getByText('Set Stockpile Marker')).toBeVisible({ timeout: 5000 });

    // Should show corp Tritanium's current quantity (100,000) - scope to dialog
    const dialog = page.getByRole('dialog');
    await expect(dialog.getByText(/100,000/)).toBeVisible();

    // Set desired quantity higher
    const desiredQtyInput = page.getByLabel(/Desired Quantity/i);
    await desiredQtyInput.clear();
    await desiredQtyInput.fill('200000');

    await page.getByRole('button', { name: /Save/i }).click();
    await expect(page.getByText('Set Stockpile Marker')).not.toBeVisible({ timeout: 5000 });

    // Verify both deficits appear on the stockpiles page
    await page.goto('/stockpiles');
    await expect(page.getByText('Tritanium').first()).toBeVisible({ timeout: 10000 });

    // Should show 2 deficit rows for Tritanium (personal + corp)
    const rows = page.getByRole('row').filter({ hasText: 'Tritanium' });
    await expect(rows).toHaveCount(2, { timeout: 5000 });

    // One row should show Stargazer Industries as owner
    await expect(page.getByText('Stargazer Industries')).toBeVisible();
  });

  test('edit stockpile marker changes desired quantity', async ({ page }) => {
    await page.goto('/inventory');

    // Expand Jita station
    await expect(page.getByText('Jita IV - Moon 4')).toBeVisible({ timeout: 10000 });
    await page.getByText('Jita IV - Moon 4').click();

    // Expand Personal Hangar to see items
    await page.getByText('Personal Hangar').first().click();
    await expect(page.getByText('Tritanium')).toBeVisible({ timeout: 5000 });

    // Find Tritanium row and click the edit stockpile button (first button in row)
    const tritaniumRow = page.getByRole('row').filter({ hasText: 'Tritanium' }).first();
    await tritaniumRow.getByRole('button').first().click();

    // Dialog should show "Edit Stockpile Marker" (marker already exists)
    await expect(page.getByText('Edit Stockpile Marker')).toBeVisible({ timeout: 5000 });

    // Change desired quantity
    const desiredQtyInput = page.getByLabel(/Desired Quantity/i);
    await desiredQtyInput.clear();
    await desiredQtyInput.fill('75000');

    await page.getByRole('button', { name: /Save/i }).click();
    await expect(page.getByText('Edit Stockpile Marker')).not.toBeVisible({ timeout: 5000 });

    // Verify updated value on stockpiles page
    await page.goto('/stockpiles');
    await expect(page.getByText('Tritanium').first()).toBeVisible({ timeout: 10000 });

    // Character Tritanium row should now show target of 75,000
    const personalRow = page.getByRole('row').filter({ hasText: 'Alice' });
    await expect(personalRow.getByText('75,000')).toBeVisible();
  });

  test('below target filter shows only deficit items on assets page', async ({ page }) => {
    await page.goto('/inventory');

    // Wait for assets to load
    await expect(page.getByText('Jita IV - Moon 4')).toBeVisible({ timeout: 10000 });

    // Toggle the "Below target only" switch
    await page.getByText('Below target only').click();

    // Expand Jita to see filtered results
    await page.getByText('Jita IV - Moon 4').click();

    // Personal Hangar should appear (Tritanium has a marker with deficit)
    await page.getByText('Personal Hangar').first().click();
    await expect(page.getByText('Tritanium')).toBeVisible({ timeout: 5000 });

    // Pyerite should NOT be visible (no stockpile marker set on it)
    await expect(page.getByText('Pyerite')).not.toBeVisible();
    await expect(page.getByText('Mexallon')).not.toBeVisible();
  });

  test('stockpiles page search filters deficit items', async ({ page }) => {
    await page.goto('/stockpiles');

    // Wait for deficits to load
    await expect(page.getByText('Tritanium').first()).toBeVisible({ timeout: 10000 });

    // Search filters by item name, structure, solar system, region, and container
    const searchInput = page.getByPlaceholder(/Search items, structures/i);

    // Search for "Tritanium" to verify filter shows matching items
    await searchInput.fill('Tritanium');
    await expect(page.getByText('Tritanium').first()).toBeVisible({ timeout: 5000 });

    // Search for something that won't match
    await searchInput.clear();
    await searchInput.fill('Nonexistent Item');
    await expect(page.getByText('No items match your search')).toBeVisible({ timeout: 5000 });

    // Search by solar system name
    await searchInput.clear();
    await searchInput.fill('Jita');
    await expect(page.getByText('Tritanium').first()).toBeVisible({ timeout: 5000 });
  });

  test('delete stockpile markers', async ({ page }) => {
    await page.goto('/inventory');

    // Accept all confirm dialogs in this test
    page.on('dialog', dialog => dialog.accept());

    // Expand Jita station
    await expect(page.getByText('Jita IV - Moon 4')).toBeVisible({ timeout: 10000 });
    await page.getByText('Jita IV - Moon 4').click();

    // Delete the character Tritanium marker
    await page.getByText('Personal Hangar').first().click();
    await expect(page.getByText('Tritanium')).toBeVisible({ timeout: 5000 });

    const tritaniumRow = page.getByRole('row').filter({ hasText: 'Tritanium' }).first();
    await tritaniumRow.getByTitle('Remove stockpile target').click();
    await expect(tritaniumRow.getByTitle('Remove stockpile target')).not.toBeVisible({ timeout: 10000 });

    // Delete the corp Tritanium marker
    await page.getByText(/Stargazer Industries - Main Hangar/).first().click();

    const corpTritaniumRow = page.getByRole('row').filter({ hasText: 'Tritanium' }).last();
    await corpTritaniumRow.getByTitle('Remove stockpile target').click();
    await expect(corpTritaniumRow.getByTitle('Remove stockpile target')).not.toBeVisible({ timeout: 10000 });

    // Verify on stockpiles page - should be empty again
    await page.goto('/stockpiles');
    await expect(page.getByText('No stockpiles need replenishment')).toBeVisible({ timeout: 10000 });
  });
});
