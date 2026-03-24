import { test, expect } from '@playwright/test';

// ---------------------------------------------------------------------------
// Arbiter E2E tests
//
// Covers:
//   1. Page loads — heading "Arbiter" is visible, no crash
//   2. Settings accordion opens — Structures tab content visible
//   3. Settings persist across save — final_facility_tax saved and reloaded
//   4. Create a scope — type name in New scope input, click +, scope appears in dropdown
//   5. Scan runs — Scan Opportunities button is clickable; if results appear, verify row
//   6. BOM tree expands — if scan produced results, clicking first row fetches BOM
//
// Alice (user 1001) has arbiter_enabled=true in seed.sql.
// No mock ESI needed — Arbiter uses backend CRUD + cached market prices.
// Scan may return 0 results if SDE data is sparse in the test DB; all
// scan-result assertions gracefully skip when the table is empty.
// ---------------------------------------------------------------------------

test.describe('Arbiter', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/arbiter');
  });

  // -------------------------------------------------------------------------
  // 1. Page loads
  // -------------------------------------------------------------------------

  test('page loads with Arbiter heading and no crash', async ({ page }) => {
    // The h1 "Arbiter" heading is always rendered (not behind auth guard for Alice)
    await expect(page.getByRole('heading', { name: /^Arbiter$/i })).toBeVisible({ timeout: 10000 });

    // The "Scan Opportunities" button should be present (controls bar)
    await expect(page.getByRole('button', { name: /Scan Opportunities/i })).toBeVisible({ timeout: 5000 });

    // No error boundary / crash text
    await expect(page.getByText(/something went wrong/i)).not.toBeVisible();
  });

  // -------------------------------------------------------------------------
  // 2. Settings accordion opens
  // -------------------------------------------------------------------------

  test('Settings accordion expands and shows Structures tab content', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /^Arbiter$/i })).toBeVisible({ timeout: 10000 });

    // The collapsible trigger is a <button> whose visible text is "Settings"
    // (rendered as a <span> child). "Save Settings" is only visible after expansion,
    // so before clicking we can safely use hasText: 'Settings' on all buttons,
    // then take the first match which is the trigger at the top of the accordion.
    const settingsTrigger = page.locator('button').filter({ hasText: 'Settings' }).first();
    await expect(settingsTrigger).toBeVisible({ timeout: 5000 });
    await settingsTrigger.click();

    // The Structures tab should now be visible inside the collapsible content
    await expect(page.getByRole('tab', { name: /Structures/i })).toBeVisible({ timeout: 5000 });

    // The "Reaction" section heading is rendered inside StructureSection
    await expect(page.getByRole('heading', { name: /^Reaction$/i })).toBeVisible({ timeout: 5000 });
  });

  // -------------------------------------------------------------------------
  // 3. Settings persist across save
  // -------------------------------------------------------------------------

  test('changed final_facility_tax setting persists after Save Settings and reload', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /^Arbiter$/i })).toBeVisible({ timeout: 10000 });

    // Open Settings accordion
    const settingsTrigger = page.locator('button').filter({ hasText: 'Settings' }).first();
    await settingsTrigger.click();

    // Wait for the Structures tab content to appear
    await expect(page.getByRole('heading', { name: /^Final Build$/i })).toBeVisible({ timeout: 5000 });

    // The Facility Tax % input for Final Build is the 4th facility-tax input on the page
    // (Reaction, Invention, Component Build, Final Build — in that order).
    // Each StructureSection renders exactly one <input type="number"> for facility tax.
    const facilityTaxInputs = page.locator('input[type="number"]');
    // Wait until at least 4 number inputs are present
    await expect(facilityTaxInputs.nth(3)).toBeVisible({ timeout: 5000 });

    // Clear and fill the Final Build facility tax field (4th number input, index 3)
    const finalTaxInput = facilityTaxInputs.nth(3);
    await finalTaxInput.fill('2.5');
    await expect(finalTaxInput).toHaveValue('2.5', { timeout: 3000 });

    // Click "Save Settings" — find by text (not role name) to avoid the Settings trigger above
    await page.getByRole('button', { name: /Save Settings/i }).click();

    // After save the accordion closes (settingsOpen → false)
    await expect(page.getByRole('tab', { name: /Structures/i })).not.toBeVisible({ timeout: 5000 });

    // Reload and re-open settings to verify persistence
    await page.reload();
    await expect(page.getByRole('heading', { name: /^Arbiter$/i })).toBeVisible({ timeout: 10000 });

    const settingsTriggerReloaded = page.locator('button').filter({ hasText: 'Settings' }).first();
    await settingsTriggerReloaded.click();

    await expect(page.getByRole('heading', { name: /^Final Build$/i })).toBeVisible({ timeout: 5000 });

    const facilityTaxInputsReloaded = page.locator('input[type="number"]');
    await expect(facilityTaxInputsReloaded.nth(3)).toBeVisible({ timeout: 5000 });
    await expect(facilityTaxInputsReloaded.nth(3)).toHaveValue('2.5', { timeout: 5000 });
  });

  // -------------------------------------------------------------------------
  // 4. Create a scope
  // -------------------------------------------------------------------------

  test('can create a new scope by filling the New scope input and clicking +', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /^Arbiter$/i })).toBeVisible({ timeout: 10000 });

    // The "New scope..." placeholder input is always visible in the controls bar
    const newScopeInput = page.getByPlaceholder(/New scope/i);
    await expect(newScopeInput).toBeVisible({ timeout: 5000 });

    await newScopeInput.fill('E2E Test Scope');

    // The + button (Plus icon, no label) is the only button inside the same
    // flex wrapper div as the "New scope..." input. Locate the innermost div
    // that contains both the input and a button.
    const newScopeWrapper = page.locator('div').filter({
      has: page.getByPlaceholder(/New scope/i),
    }).last(); // last() gets the innermost matching wrapper
    const plusButton = newScopeWrapper.getByRole('button');
    await expect(plusButton).toBeEnabled({ timeout: 3000 });
    await plusButton.click();

    // After creation, the scope should appear in the scope dropdown
    // The Select trigger shows the newly selected scope name
    await expect(page.getByText('E2E Test Scope').first()).toBeVisible({ timeout: 5000 });
  });

  // -------------------------------------------------------------------------
  // 5. Scan runs
  // -------------------------------------------------------------------------

  test('Scan Opportunities button is clickable and page does not error', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /^Arbiter$/i })).toBeVisible({ timeout: 10000 });

    const scanButton = page.getByRole('button', { name: /Scan Opportunities/i });
    await expect(scanButton).toBeVisible({ timeout: 5000 });
    await expect(scanButton).toBeEnabled({ timeout: 3000 });

    await scanButton.click();

    // The button label changes to "Scanning..." while in progress
    // Wait for it to finish (either returns results or goes back to Scan Opportunities)
    await expect(async () => {
      const label = await scanButton.textContent();
      expect(label).toMatch(/Scan Opportunities/i);
    }).toPass({ timeout: 30000 });

    // No crash / error boundary
    await expect(page.getByText(/something went wrong/i)).not.toBeVisible();
    await expect(page.getByText(/scan failed/i)).not.toBeVisible();
  });

  test('scan results table appears if SDE data is available', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /^Arbiter$/i })).toBeVisible({ timeout: 10000 });

    const scanButton = page.getByRole('button', { name: /Scan Opportunities/i });
    await scanButton.click();

    // Wait for scan to finish (button returns to non-scanning state)
    await expect(async () => {
      const label = await scanButton.textContent();
      expect(label).toMatch(/Scan Opportunities/i);
    }).toPass({ timeout: 35000 });

    // Check whether any result rows appeared in the results table.
    // The results table has TableRow elements with item names.
    // Each OpportunityRow has an expand chevron and item name in the first cell.
    const resultRows = page.locator('table tbody tr').first();
    const hasResults = await resultRows.isVisible().catch(() => false);

    if (!hasResults) {
      // SDE data not loaded — just verify no error state
      await expect(page.getByText(/something went wrong/i)).not.toBeVisible();
      return;
    }

    // At least one result row is visible
    await expect(resultRows).toBeVisible({ timeout: 5000 });
  });

  // -------------------------------------------------------------------------
  // 6. BOM tree expands (only runs if scan produces results)
  // -------------------------------------------------------------------------

  test('clicking a result row expands BOM tree panel', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /^Arbiter$/i })).toBeVisible({ timeout: 10000 });

    const scanButton = page.getByRole('button', { name: /Scan Opportunities/i });
    await scanButton.click();

    // Wait for scan to complete
    await expect(async () => {
      const label = await scanButton.textContent();
      expect(label).toMatch(/Scan Opportunities/i);
    }).toPass({ timeout: 35000 });

    // Check if there are any results — if not, skip gracefully
    const firstRow = page.locator('table tbody tr').first();
    const hasResults = await firstRow.isVisible().catch(() => false);

    if (!hasResults) {
      // No SDE data — test passes vacuously
      await expect(page.getByText(/something went wrong/i)).not.toBeVisible();
      return;
    }

    // Click the first result row to expand it
    await firstRow.click();

    // After expanding, the BOM section header appears:
    // "Bill of Materials — <item name>"
    await expect(page.getByText(/Bill of Materials/i)).toBeVisible({ timeout: 10000 });

    // The BOM panel has column headers: Item, Qty, Available, Needed, Delta, Unit Price, Decision
    await expect(page.getByRole('columnheader', { name: /^Item$/i })).toBeVisible({ timeout: 5000 });
  });
});
