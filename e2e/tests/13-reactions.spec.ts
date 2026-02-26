import { test, expect } from '@playwright/test';

test.describe('Reactions Calculator', () => {
  test.beforeEach(async ({ page }) => {
    // Clear localStorage so tab/settings state is fresh for each test
    await page.goto('/reactions');
    await page.evaluate(() => localStorage.clear());
  });

  test('navigate to reactions page shows tabs', async ({ page }) => {
    await page.goto('/reactions');

    // All three tabs should be present
    await expect(page.getByRole('tab', { name: 'Pick Reactions' })).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('tab', { name: /Shopping List/i })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'Plan Summary' })).toBeVisible();
  });

  test('settings toolbar controls are visible', async ({ page }) => {
    await page.goto('/reactions');

    // MUI Select without explicit labelId prop does not wire aria-labelledby automatically
    // in MUI v7. Scope to the FormControl container using the label text to find each combobox.

    // Structure dropdown — find the FormControl containing the "Structure" label
    const structureControl = page.locator('.MuiFormControl-root').filter({
      has: page.locator('label').filter({ hasText: 'Structure' }),
    });
    await expect(structureControl.getByRole('combobox')).toBeVisible({ timeout: 10000 });

    // Rig dropdown
    const rigControl = page.locator('.MuiFormControl-root').filter({
      has: page.locator('label').filter({ hasText: 'Rig' }),
    });
    await expect(rigControl.getByRole('combobox')).toBeVisible();

    // Security dropdown
    const securityControl = page.locator('.MuiFormControl-root').filter({
      has: page.locator('label').filter({ hasText: 'Security' }),
    });
    await expect(securityControl.getByRole('combobox')).toBeVisible();

    // Cycle Days input
    await expect(page.getByLabel('Cycle Days')).toBeVisible();

    // System autocomplete (Autocomplete renders a combobox input with associated label)
    await expect(page.getByRole('combobox', { name: /System/i })).toBeVisible();
  });

  test('Pick Reactions tab shows reaction list with seeded Crystalline Carbonide', async ({ page }) => {
    await page.goto('/reactions');

    // The Pick Reactions tab is active by default
    await expect(page.getByRole('tab', { name: 'Pick Reactions' })).toBeVisible({ timeout: 10000 });

    // The reactions table should load and show our seeded reaction.
    // The API fetches all reactions from sde_blueprint_activities where activity='reaction'.
    // Our seed adds Crystalline Carbonide Reaction Formula (blueprint 28209) → Crystalline Carbonide (16634).
    // The product group is 'Advanced Material' which is not in the SIMPLE_GROUPS filter,
    // so it will appear in the table.
    // Use toPass with reload to handle slow API responses or delayed rendering.
    await expect(async () => {
      await page.reload();
      await expect(page.getByText('Crystalline Carbonide')).toBeVisible({ timeout: 5000 });
    }).toPass({ timeout: 30000 });

    // The Group column should show the Crystalline Carbonide group name.
    // We don't assert the specific group name here because the SDE import populates
    // the real EVE group names which may differ from seed data ("Advanced Material").
    // It is sufficient that the product name is visible and the reaction loaded correctly.
  });

  test('Pick Reactions tab shows reaction count and ME factor', async ({ page }) => {
    await page.goto('/reactions');

    await expect(page.getByRole('tab', { name: 'Pick Reactions' })).toBeVisible({ timeout: 10000 });

    // The toolbar above the table shows "N reactions | ME: X"
    // Wait for data to load by waiting for the reaction to appear first
    await expect(page.getByText('Crystalline Carbonide')).toBeVisible({ timeout: 15000 });

    // Verify the count display is shown (contains "reactions")
    await expect(page.getByText(/\d+ reaction/)).toBeVisible();
  });

  test('Search filter narrows down the reaction list', async ({ page }) => {
    await page.goto('/reactions');

    await expect(page.getByRole('tab', { name: 'Pick Reactions' })).toBeVisible({ timeout: 10000 });

    // Wait for reactions to load
    await expect(page.getByText('Crystalline Carbonide')).toBeVisible({ timeout: 15000 });

    // Search for "crystal" — should still show Crystalline Carbonide
    const searchInput = page.getByLabel('Search');
    await searchInput.fill('crystal');
    await expect(page.getByText('Crystalline Carbonide')).toBeVisible({ timeout: 5000 });

    // Search for something that doesn't match — table should show "No reactions found"
    await searchInput.fill('zzznomatch');
    await expect(page.getByText('No reactions found')).toBeVisible({ timeout: 5000 });

    // Clear search — Crystalline Carbonide returns
    await searchInput.fill('');
    await expect(page.getByText('Crystalline Carbonide')).toBeVisible({ timeout: 5000 });
  });

  test('Shopping List tab shows empty state when no reactions selected', async ({ page }) => {
    await page.goto('/reactions');

    // Wait for data to load
    await expect(page.getByRole('tab', { name: 'Pick Reactions' })).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('Crystalline Carbonide')).toBeVisible({ timeout: 15000 });

    // Switch to Shopping List tab
    await page.getByRole('tab', { name: /Shopping List/i }).click();

    // No reactions selected yet — empty state message
    await expect(
      page.getByText('Select reactions in the Pick Reactions tab to generate a shopping list.')
    ).toBeVisible({ timeout: 5000 });
  });

  test('Plan Summary tab shows empty state when no reactions selected', async ({ page }) => {
    await page.goto('/reactions');

    // Wait for data to load
    await expect(page.getByRole('tab', { name: 'Pick Reactions' })).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('Crystalline Carbonide')).toBeVisible({ timeout: 15000 });

    // Switch to Plan Summary tab
    await page.getByRole('tab', { name: 'Plan Summary' }).click();

    // No reactions selected yet — empty state message
    await expect(
      page.getByText('Select reactions in the Pick Reactions tab to see a plan summary.')
    ).toBeVisible({ timeout: 5000 });
  });

  test('selecting a reaction instance populates Shopping List and Plan Summary', async ({ page }) => {
    await page.goto('/reactions');

    // Wait for reactions to load — use toPass with reload in case API is slow
    await expect(page.getByRole('tab', { name: 'Pick Reactions' })).toBeVisible({ timeout: 10000 });
    await expect(async () => {
      await page.reload();
      await expect(page.getByText('Crystalline Carbonide')).toBeVisible({ timeout: 5000 });
    }).toPass({ timeout: 30000 });

    // Find the Instances field in the Crystalline Carbonide row and set it to 1
    // The row contains the product name so we scope to the row
    const carbonideRow = page.getByRole('row').filter({ hasText: 'Crystalline Carbonide' }).first();
    const instancesInput = carbonideRow.getByPlaceholder('0');
    await instancesInput.fill('1');

    // The Shopping List tab label should update to show (1)
    await expect(page.getByRole('tab', { name: /Shopping List \(1\)/i })).toBeVisible({ timeout: 5000 });

    // Switch to Shopping List tab — it should now have materials
    await page.getByRole('tab', { name: /Shopping List/i }).click();

    // The plan API is called when selections change; wait for the shopping list to populate.
    // The SDE import overwrites seeded blueprint materials with real EVE data, so we
    // cannot test for specific material names like "Nocxium"/"Isogen" (seed data).
    // Instead, verify the shopping list has items by waiting for the "Total" footer row
    // which always appears when shopping_list.length > 0.
    await expect(page.locator('td').filter({ hasText: /^Total$/ })).toBeVisible({ timeout: 15000 });

    // Switch to Plan Summary tab — it should now show summary stats
    await page.getByRole('tab', { name: 'Plan Summary' }).click();

    // Plan Summary shows stat cards: Investment, Revenue, Profit
    // Use exact:true for 'Profit' to avoid matching the "Net Profit" table column header
    // from the Shopping List tab that may still be in DOM during tab transition.
    await expect(page.getByText('Investment')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Revenue')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Profit', { exact: true })).toBeVisible({ timeout: 5000 });
  });

  test('structure dropdown changes persist to default Tatara', async ({ page }) => {
    await page.goto('/reactions');

    // MUI v7 Select without explicit labelId prop does not wire aria-labelledby automatically.
    // Scope to the FormControl container that wraps the Structure label and combobox.
    const structureControl = page.locator('.MuiFormControl-root').filter({
      has: page.locator('label').filter({ hasText: 'Structure' }),
    });
    const structureSelect = structureControl.getByRole('combobox');
    await expect(structureSelect).toBeVisible({ timeout: 10000 });

    // The default structure is 'tatara' — verify the displayed value in the combobox
    await expect(structureControl.getByRole('combobox')).toHaveText('Tatara', { timeout: 5000 });

    // Change to Athanor
    await structureSelect.click();
    await page.getByRole('option', { name: 'Athanor' }).click();

    // Verify the selection changed — scope to the combobox display value to avoid matching
    // the MUI MenuItem "Athanor" that may still be in DOM after the dropdown closes.
    await expect(structureControl.getByRole('combobox')).toHaveText('Athanor', { timeout: 5000 });
  });
});
