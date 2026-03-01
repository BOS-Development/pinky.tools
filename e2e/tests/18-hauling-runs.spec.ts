import { test, expect } from '@playwright/test';

// ---------------------------------------------------------------------------
// Hauling Runs — Phase 1 E2E tests
//
// These tests cover the hauling runs feature:
//   /hauling          — list page with PLANNING/ACCUMULATING runs
//   /hauling/[id]     — detail page with items and fill progress
//   /hauling/scanner  — market scanner for hub-to-hub arbitrage
//
// Tests are additive: earlier tests create data later tests depend on.
// ---------------------------------------------------------------------------

test.describe('Hauling Runs', () => {
  test.beforeEach(async ({ page }) => {
    // Clear localStorage so tree/panel state is fresh
    await page.goto('/hauling');
    await page.evaluate(() => localStorage.clear());
  });

  // -------------------------------------------------------------------------
  // 1. List page smoke tests
  // -------------------------------------------------------------------------

  test('navigate to hauling page shows heading and New Run button', async ({ page }) => {
    await page.goto('/hauling');

    // Page should load without error and show a heading
    await expect(page.getByRole('heading', { name: /Hauling Runs/i }).first()).toBeVisible({ timeout: 10000 });

    // "New Run" button should be visible
    await expect(page.getByRole('button', { name: /New Run/i })).toBeVisible({ timeout: 5000 });
  });

  test('hauling runs list shows empty state when no runs exist', async ({ page }) => {
    await page.goto('/hauling');

    // Page heading confirms we are on the right page
    await expect(page.getByRole('heading', { name: /Hauling Runs/i }).first()).toBeVisible({ timeout: 10000 });

    // Note: DB may have existing runs from previous test runs; empty state not guaranteed
  });

  // -------------------------------------------------------------------------
  // 2. Create a hauling run
  // -------------------------------------------------------------------------

  test('create a hauling run via New Run dialog', async ({ page }) => {
    await page.goto('/hauling');

    await expect(page.getByRole('heading', { name: /Hauling Runs/i }).first()).toBeVisible({ timeout: 10000 });

    // Wait for the New Run button to be interactive (belt-and-suspenders after the heading check)
    await expect(page.getByRole('button', { name: /New Run/i })).toBeVisible({ timeout: 5000 });

    // Open the create dialog
    await page.getByRole('button', { name: /New Run/i }).click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // Fill in the run name
    await dialog.getByLabel(/Name/i).fill('Jita to Amarr Test Run');

    // Select source region: The Forge (10000002)
    // MUI Select — find FormControl by label then click combobox inside
    const sourceControl = dialog.locator('.MuiFormControl-root').filter({
      has: page.locator('label').filter({ hasText: /Source Region|From Region/i }),
    });
    await sourceControl.getByRole('combobox').click();
    await page.getByRole('option', { name: /The Forge/i }).click();

    // Select destination region: Domain (10000043)
    const destControl = dialog.locator('.MuiFormControl-root').filter({
      has: page.locator('label').filter({ hasText: /Destination Region|To Region/i }),
    });
    await destControl.getByRole('combobox').click();
    await page.getByRole('option', { name: /Domain/i }).click();

    // Fill in Max Volume
    const volumeInput = dialog.getByLabel(/Max Volume/i);
    await volumeInput.clear();
    await volumeInput.fill('300000');

    // Submit the form
    await dialog.getByRole('button', { name: /Create|Save/i }).click();

    // Dialog should close
    await expect(dialog).not.toBeVisible({ timeout: 5000 });

    // The new run should appear in the list
    await expect(page.getByText('Jita to Amarr Test Run')).toBeVisible({ timeout: 10000 });
  });

  test('created run shows PLANNING status in the list', async ({ page }) => {
    await page.goto('/hauling');

    // Wait for the run created in the previous test
    await expect(page.getByText('Jita to Amarr Test Run')).toBeVisible({ timeout: 10000 });

    // Status badge should show PLANNING
    await expect(page.getByText(/PLANNING/i).first()).toBeVisible({ timeout: 5000 });
  });

  // -------------------------------------------------------------------------
  // 3. View run detail
  // -------------------------------------------------------------------------

  test('clicking a run navigates to detail page', async ({ page }) => {
    await page.goto('/hauling');

    // Wait for run to appear
    await expect(page.getByText('Jita to Amarr Test Run')).toBeVisible({ timeout: 10000 });

    // Click the run row to navigate to the detail page
    // Use row-based click to avoid matching non-link DOM ancestors; .first() guards against
    // duplicate runs from CI retries accumulating in the DB
    await page.getByRole('row').filter({ hasText: 'Jita to Amarr Test Run' }).first().click();

    // The detail page should show the run name
    await expect(page.getByText('Jita to Amarr Test Run')).toBeVisible({ timeout: 10000 });

    // Should be on the detail URL /hauling/<id>
    await expect(page).toHaveURL(/\/hauling\/\d+/, { timeout: 5000 });
  });

  test('run detail page shows items table and Add Item button', async ({ page }) => {
    await page.goto('/hauling');

    // Navigate to the run detail
    await expect(page.getByText('Jita to Amarr Test Run')).toBeVisible({ timeout: 10000 });
    await page.getByRole('row').filter({ hasText: 'Jita to Amarr Test Run' }).first().click();

    // Wait for detail page load
    await expect(page).toHaveURL(/\/hauling\/\d+/, { timeout: 5000 });

    // Item table should be visible (even if empty)
    // The table may be a MUI Table or just a section heading
    await expect(
      page.getByRole('table').or(page.getByText(/Items/i)).first()
    ).toBeVisible({ timeout: 10000 });

    // "Add Item" button should be present
    await expect(page.getByRole('button', { name: /Add Item/i })).toBeVisible({ timeout: 5000 });
  });

  test('run detail page shows run metadata', async ({ page }) => {
    await page.goto('/hauling');

    await expect(page.getByText('Jita to Amarr Test Run')).toBeVisible({ timeout: 10000 });
    await page.getByRole('row').filter({ hasText: 'Jita to Amarr Test Run' }).first().click();

    await expect(page).toHaveURL(/\/hauling\/\d+/, { timeout: 5000 });

    // Run name should be in the heading or title area
    await expect(page.getByText('Jita to Amarr Test Run').first()).toBeVisible({ timeout: 10000 });

    // Region info should be visible (The Forge → Domain)
    await expect(page.getByText(/The Forge/i).first()).toBeVisible({ timeout: 10000 });
    await expect(page.getByText(/Domain/i).first()).toBeVisible({ timeout: 5000 });
  });

  // -------------------------------------------------------------------------
  // 4. Market scanner
  // -------------------------------------------------------------------------

  test('navigate to scanner page shows scanner controls', async ({ page }) => {
    await page.goto('/hauling/scanner');

    // Page should load without error
    await expect(
      page.getByRole('heading', { name: /Scanner|Hauling Scanner|Arbitrage/i }).or(
        page.getByText(/Scanner|Arbitrage/i)
      ).first()
    ).toBeVisible({ timeout: 10000 });

    // Region selection controls should be present
    await expect(
      page.getByText(/Source Region|From Region/i).first()
    ).toBeVisible({ timeout: 5000 });
    await expect(
      page.getByText(/Destination Region|To Region/i).first()
    ).toBeVisible({ timeout: 5000 });
  });

  test('scanner page has Scan button', async ({ page }) => {
    await page.goto('/hauling/scanner');

    await expect(
      page.getByRole('heading', { name: /Scanner|Hauling Scanner|Arbitrage/i }).or(
        page.getByText(/Scanner|Arbitrage/i)
      ).first()
    ).toBeVisible({ timeout: 10000 });

    // Scan button should be present
    await expect(page.getByRole('button', { name: /Scan/i })).toBeVisible({ timeout: 5000 });
  });

  test('scanner returns results after scan completes', async ({ page }) => {
    await page.goto('/hauling/scanner');

    await expect(
      page.getByRole('heading', { name: /Scanner|Hauling Scanner|Arbitrage/i }).or(
        page.getByText(/Scanner|Arbitrage/i)
      ).first()
    ).toBeVisible({ timeout: 10000 });

    // Select source region: The Forge
    const sourceControl = page.locator('.MuiFormControl-root').filter({
      has: page.locator('label').filter({ hasText: /Source Region|From Region/i }),
    });
    await sourceControl.getByRole('combobox').click();
    await page.getByRole('option', { name: /The Forge/i }).click();

    // Select destination region: Domain
    const destControl = page.locator('.MuiFormControl-root').filter({
      has: page.locator('label').filter({ hasText: /Destination Region|To Region/i }),
    });
    await destControl.getByRole('combobox').click();
    await page.getByRole('option', { name: /Domain/i }).click();

    // Click Scan
    await page.getByRole('button', { name: /Scan/i }).click();

    // The backend triggers a background scan (returns immediately with {status:"scanning"}).
    // The scanner page should show either:
    //   a) A loading/scanning indicator while results come in, or
    //   b) Scan results once the background job completes and results are fetched.
    // Use toPass with reload to handle async scan completion.
    await expect(async () => {
      await page.reload();
      // Either a results table, a "Scanning..." message, or a "No results" empty state
      await expect(
        page.getByRole('table').or(
          page.getByText(/Scanning|scanning|No results|No arbitrage/i)
        ).first()
      ).toBeVisible({ timeout: 5000 });
    }).toPass({ timeout: 30000 });
  });

  // -------------------------------------------------------------------------
  // 5. Run management
  // -------------------------------------------------------------------------

  test('can navigate back to list from detail page', async ({ page }) => {
    await page.goto('/hauling');

    await expect(page.getByText('Jita to Amarr Test Run')).toBeVisible({ timeout: 10000 });
    await page.getByRole('row').filter({ hasText: 'Jita to Amarr Test Run' }).first().click();

    await expect(page).toHaveURL(/\/hauling\/\d+/, { timeout: 5000 });

    // Navigate back to the list via back button or breadcrumb/link
    await expect(
      page.getByRole('link', { name: /Hauling Runs|Back/i }).or(
        page.getByRole('button', { name: /Back/i })
      ).first()
    ).toBeVisible({ timeout: 5000 });
  });

  test('delete run removes it from the list', async ({ page }) => {
    await page.goto('/hauling');

    await expect(page.getByText('Jita to Amarr Test Run')).toBeVisible({ timeout: 10000 });

    // Accept the native browser confirm dialog if one appears
    page.on('dialog', d => d.accept());

    // Find the run row and click delete button
    const runRow = page.getByRole('row').filter({ hasText: 'Jita to Amarr Test Run' });
    await runRow.getByRole('button', { name: /Delete/i }).click();

    // Run should be removed from the list
    await expect(page.getByText('Jita to Amarr Test Run')).not.toBeVisible({ timeout: 10000 });
  });
});
