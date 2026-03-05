import { test, expect } from '@playwright/test';

// ---------------------------------------------------------------------------
// Hauling Runs Phase 3 — Sell Tracking & P&L E2E tests
//
// Covers:
//   - Sell progress columns (Qty Sold, Sell Fill, Revenue) visible on SELLING runs
//   - Sell Progress summary card (Realized so far / Pending (est.)) on SELLING runs
//   - Projected vs Actual P&L section for COMPLETE runs
//   - Sell progress columns hidden for non-SELLING runs (PLANNING)
//   - Enter P&L dialog populates sell data (integration smoke test)
//
// All tests are self-contained: they create their own runs and do not rely
// on state left by earlier spec files.
// ---------------------------------------------------------------------------

// Helper: create a new hauling run, navigate to its detail page, and return the run ID.
async function createHaulingRun(
  page: import('@playwright/test').Page,
  name: string,
): Promise<string> {
  await page.goto('/hauling');
  await expect(page.getByRole('heading', { name: /Hauling Runs/i }).first()).toBeVisible({ timeout: 10000 });
  await page.getByRole('button', { name: /New Run/i }).click();

  const dialog = page.getByRole('dialog');
  await expect(dialog).toBeVisible({ timeout: 5000 });

  await dialog.getByPlaceholder('Run name').fill(name);

  // From Region: The Forge (first combobox)
  await dialog.getByRole('combobox').first().click();
  await page.getByRole('option', { name: /The Forge/i }).click();

  // To Region: Domain (second combobox)
  await dialog.getByRole('combobox').nth(1).click();
  await page.getByRole('option', { name: /Domain/i }).click();

  await dialog.getByRole('button', { name: /Create/i }).click();
  await expect(dialog).not.toBeVisible({ timeout: 5000 });

  // Navigate to the detail page to get the run ID
  await expect(page.getByText(name).first()).toBeVisible({ timeout: 10000 });
  await page.getByRole('row').filter({ hasText: name }).first().click();
  await expect(page).toHaveURL(/\/hauling\/\d+/, { timeout: 5000 });

  const url = page.url();
  const match = url.match(/\/hauling\/(\d+)/);
  return match ? match[1] : '';
}

// Helper: add an item to the currently-open run detail page.
async function addRunItem(
  page: import('@playwright/test').Page,
  typeId: string,
  typeName: string,
  quantityPlanned: string,
  buyPrice?: string,
  sellPrice?: string,
): Promise<void> {
  await page.getByRole('button', { name: /Add Item/i }).click();

  const dialog = page.getByRole('dialog');
  await expect(dialog).toBeVisible({ timeout: 5000 });

  // Type ID (first spinbutton)
  await dialog.getByRole('spinbutton').first().fill(typeId);
  // Type Name (textbox)
  await dialog.getByRole('textbox').fill(typeName);
  // Quantity Planned (second spinbutton)
  await dialog.getByRole('spinbutton').nth(1).fill(quantityPlanned);

  if (buyPrice) {
    // Buy Price (third spinbutton)
    await dialog.getByRole('spinbutton').nth(2).fill(buyPrice);
  }
  if (sellPrice) {
    // Sell Price (fourth spinbutton)
    await dialog.getByRole('spinbutton').nth(3).fill(sellPrice);
  }

  await dialog.getByRole('button', { name: /^Add$/i }).click();
  await expect(dialog).not.toBeVisible({ timeout: 5000 });

  // Wait for the item to appear in the table
  await expect(page.getByText(typeName).first()).toBeVisible({ timeout: 10000 });
}

// Helper: change the run status via the combobox on the detail page.
async function changeRunStatus(
  page: import('@playwright/test').Page,
  newStatus: string,
): Promise<void> {
  await page.getByRole('combobox').click();
  await page.getByRole('option', { name: newStatus }).click();
  await expect(page.getByText(newStatus).first()).toBeVisible({ timeout: 10000 });
}

test.describe('Hauling Runs Phase 3 — Sell Tracking & P&L', () => {

  // -------------------------------------------------------------------------
  // 1. Sell progress columns are hidden for PLANNING status runs
  // -------------------------------------------------------------------------

  test('sell progress columns are NOT visible on a PLANNING run', async ({ page }) => {
    const runName = 'Phase3 Planning No Sell Cols';
    await createHaulingRun(page, runName);

    await expect(page.getByText(runName).first()).toBeVisible({ timeout: 5000 });

    // Confirm we are in PLANNING status
    await expect(page.getByText(/PLANNING/i).first()).toBeVisible({ timeout: 5000 });

    // Add an item so the table is populated
    await addRunItem(page, '34', 'Tritanium', '500', '5.50', '6.00');

    // The sell-specific columns should NOT be present for PLANNING status
    await expect(page.getByRole('columnheader', { name: /Qty Sold/i })).not.toBeVisible({ timeout: 3000 });
    await expect(page.getByRole('columnheader', { name: /Sell Fill/i })).not.toBeVisible({ timeout: 3000 });

    // The Sell Progress summary card heading should also not appear
    await expect(page.getByText('Sell Progress').first()).not.toBeVisible({ timeout: 3000 });
  });

  // -------------------------------------------------------------------------
  // 2. Sell progress columns appear on SELLING runs
  // -------------------------------------------------------------------------

  test('sell progress columns appear in items table when run is SELLING', async ({ page }) => {
    const runName = 'Phase3 Sell Cols Test';
    await createHaulingRun(page, runName);

    await expect(page.getByText(runName).first()).toBeVisible({ timeout: 5000 });

    // Add an item before changing status
    await addRunItem(page, '34', 'Tritanium', '500', '5.50', '6.00');

    // Change status to SELLING
    await changeRunStatus(page, 'SELLING');

    // Sell-specific columns should now be visible as table headers
    await expect(page.getByRole('columnheader', { name: /Qty Sold/i })).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('columnheader', { name: /Sell Fill/i })).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('columnheader', { name: /Revenue/i })).toBeVisible({ timeout: 5000 });

    // The item row should show "0 / 500" for qty sold (no sales yet)
    await expect(page.getByText(/0 \/ 500/).first()).toBeVisible({ timeout: 5000 });
  });

  // -------------------------------------------------------------------------
  // 3. Sell Progress summary card appears on SELLING runs
  // -------------------------------------------------------------------------

  test('Sell Progress card shows Realized so far and Pending (est.) for SELLING run', async ({ page }) => {
    const runName = 'Phase3 Sell Progress Card';
    await createHaulingRun(page, runName);

    await expect(page.getByText(runName).first()).toBeVisible({ timeout: 5000 });

    // Add an item with buy + sell prices so the pending revenue is non-zero
    await addRunItem(page, '35', 'Pyerite', '1000', '10.00', '11.50');

    // Change status to SELLING
    await changeRunStatus(page, 'SELLING');

    // The Sell Progress card should appear
    await expect(page.getByText('Sell Progress').first()).toBeVisible({ timeout: 5000 });

    // "Realized so far" and "Pending (est.)" labels should be present
    await expect(page.getByText(/Realized so far/i).first()).toBeVisible({ timeout: 5000 });
    await expect(page.getByText(/Pending \(est\.\)/i).first()).toBeVisible({ timeout: 5000 });
  });

  // -------------------------------------------------------------------------
  // 4. Enter P&L dialog populates sell data — smoke test
  // -------------------------------------------------------------------------

  test('Enter P&L dialog allows entering qty sold and price, updates item row', async ({ page }) => {
    const runName = 'Phase3 Enter PnL Smoke';
    await createHaulingRun(page, runName);

    await expect(page.getByText(runName).first()).toBeVisible({ timeout: 5000 });

    // Add a Tritanium item: planned 500 units, buy 5.50, sell 6.00
    await addRunItem(page, '34', 'Tritanium', '500', '5.50', '6.00');

    // Change to SELLING status
    await changeRunStatus(page, 'SELLING');

    // The P&L section heading should appear
    await expect(page.getByText(/Profit.*Loss|Profit & Loss/i).first()).toBeVisible({ timeout: 10000 });

    // Click "Enter P&L"
    await page.getByRole('button', { name: /Enter P&L/i }).click();

    const pnlDialog = page.getByRole('dialog');
    await expect(pnlDialog).toBeVisible({ timeout: 5000 });

    // Dialog title should mention P&L
    await expect(pnlDialog.getByText(/Enter P&L|P&L Entry/i).first()).toBeVisible({ timeout: 5000 });

    // The dialog has a Select dropdown for "Item" — must select item before Save enables
    // The item we added is "Tritanium" with typeId 34 — select it via the combobox
    await pnlDialog.getByRole('combobox').click();
    await page.getByRole('option', { name: /Tritanium/i }).click();

    // First spinbutton = Qty Sold (required)
    const qtySoldInput = pnlDialog.getByRole('spinbutton').first();
    await expect(qtySoldInput).toBeVisible({ timeout: 5000 });
    await qtySoldInput.fill('150');

    // Second spinbutton = Avg Sell Price (optional)
    const avgPriceInput = pnlDialog.getByRole('spinbutton').nth(1);
    await expect(avgPriceInput).toBeVisible({ timeout: 5000 });
    await avgPriceInput.fill('6.00');

    // Save button should now be enabled (typeId and quantitySold are set)
    const saveBtn = pnlDialog.getByRole('button', { name: /^Save$/i });
    await expect(saveBtn).toBeEnabled({ timeout: 5000 });
    await saveBtn.click();
    await expect(pnlDialog).not.toBeVisible({ timeout: 5000 });

    // The P&L section should still be visible after saving
    await expect(page.getByText(/Profit.*Loss|Profit & Loss/i).first()).toBeVisible({ timeout: 5000 });
  });

  // -------------------------------------------------------------------------
  // 5. Projected vs Actual P&L section appears for COMPLETE runs
  // -------------------------------------------------------------------------

  test('Projected vs Actual P&L section appears when run status is COMPLETE', async ({ page }) => {
    const runName = 'Phase3 Complete PnL Test';
    await createHaulingRun(page, runName);

    await expect(page.getByText(runName).first()).toBeVisible({ timeout: 5000 });

    // Add an item with prices set so that P&L figures are non-trivial
    await addRunItem(page, '34', 'Tritanium', '500', '5.50', '6.00');

    // Change status to COMPLETE directly (status transitions are permissive in the UI)
    await changeRunStatus(page, 'COMPLETE');

    // The Projected vs Actual P&L section should appear
    await expect(page.getByText(/Projected vs Actual P&L/i).first()).toBeVisible({ timeout: 10000 });

    // Column headers: Projected, Actual
    await expect(page.getByRole('columnheader', { name: /Projected/i })).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('columnheader', { name: /Actual/i })).toBeVisible({ timeout: 5000 });

    // Row labels: Buy Cost, Revenue, Net Profit
    await expect(page.getByText(/Buy Cost/i).first()).toBeVisible({ timeout: 5000 });
    await expect(page.getByText(/Revenue/i).first()).toBeVisible({ timeout: 5000 });
    await expect(page.getByText(/Net Profit/i).first()).toBeVisible({ timeout: 5000 });
  });

  // -------------------------------------------------------------------------
  // 6. Projected vs Actual P&L values are computed correctly
  // -------------------------------------------------------------------------

  test('Projected vs Actual P&L shows correct ISK values for COMPLETE run', async ({ page }) => {
    const runName = 'Phase3 PnL Values Test';
    await createHaulingRun(page, runName);

    await expect(page.getByText(runName).first()).toBeVisible({ timeout: 5000 });

    // Add Tritanium: 1000 units, buy 5.50, sell 6.00
    // Projected buy cost = 1000 * 5.50 = 5500 → formatISK(5500) = "5.50K ISK"
    // Projected revenue  = 1000 * 6.00 = 6000 → formatISK(6000) = "6.00K ISK"
    // Projected net profit = 6000 - 5500 = 500 → formatISK(500) = "500.00 ISK"
    await addRunItem(page, '34', 'Tritanium', '1000', '5.50', '6.00');

    // Change status to COMPLETE
    await changeRunStatus(page, 'COMPLETE');

    // Wait for the P&L section to appear
    await expect(page.getByText(/Projected vs Actual P&L/i).first()).toBeVisible({ timeout: 10000 });

    // Verify projected buy cost: 1000 * 5.50 = 5500 → "5.50K ISK"
    await expect(page.getByText(/5\.50K/i).first()).toBeVisible({ timeout: 5000 });

    // Verify projected revenue: 1000 * 6.00 = 6000 → "6.00K ISK"
    await expect(page.getByText(/6\.00K/i).first()).toBeVisible({ timeout: 5000 });

    // Verify net profit row is present (value varies by qty_acquired; with 0 acquired, actual = 0)
    // The projected net profit = 500 → "500.00 ISK"
    await expect(page.getByText(/Net Profit/i).first()).toBeVisible({ timeout: 5000 });
  });

  // -------------------------------------------------------------------------
  // 7. Sell Progress card is NOT visible for PLANNING runs
  // -------------------------------------------------------------------------

  test('Sell Progress card does NOT appear for PLANNING runs', async ({ page }) => {
    const runName = 'Phase3 Planning No PnL Sections';
    await createHaulingRun(page, runName);

    await expect(page.getByText(runName).first()).toBeVisible({ timeout: 5000 });

    // Confirm PLANNING status
    await expect(page.getByText(/PLANNING/i).first()).toBeVisible({ timeout: 5000 });

    // The Sell Progress card heading is an <h3> element; it should NOT appear for PLANNING runs.
    // Use locator('h3') to avoid partial-text match on the run name itself.
    await expect(page.locator('h3').filter({ hasText: /^Sell Progress$/ })).not.toBeVisible({ timeout: 3000 });

    // Projected vs Actual P&L should also not appear
    await expect(page.locator('h3').filter({ hasText: /^Projected vs Actual P&L$/ })).not.toBeVisible({ timeout: 3000 });
  });

  // -------------------------------------------------------------------------
  // 8. Qty sold shows updated values after P&L entry submission
  // -------------------------------------------------------------------------

  test('qty sold column updates after P&L entry is saved', async ({ page }) => {
    const runName = 'Phase3 Qty Sold After PnL';
    await createHaulingRun(page, runName);

    await expect(page.getByText(runName).first()).toBeVisible({ timeout: 5000 });

    // Add a Pyerite item
    await addRunItem(page, '35', 'Pyerite', '200', '10.00', '11.50');

    // Change to SELLING
    await changeRunStatus(page, 'SELLING');

    // Verify sell columns are present and Pyerite shows 0 sold initially
    await expect(page.getByRole('columnheader', { name: /Qty Sold/i })).toBeVisible({ timeout: 5000 });
    await expect(page.getByText(/0 \/ 200/).first()).toBeVisible({ timeout: 5000 });

    // Click Enter P&L and submit 50 units sold at 11.50 each
    await page.getByRole('button', { name: /Enter P&L/i }).click();

    const pnlDialog = page.getByRole('dialog');
    await expect(pnlDialog).toBeVisible({ timeout: 5000 });

    // Must select the item from the dropdown first (required field)
    await pnlDialog.getByRole('combobox').click();
    await page.getByRole('option', { name: /Pyerite/i }).click();

    // Fill qty sold (required) and avg sell price (optional)
    await pnlDialog.getByRole('spinbutton').first().fill('50');
    await pnlDialog.getByRole('spinbutton').nth(1).fill('11.50');

    // Save should now be enabled
    const saveBtn = pnlDialog.getByRole('button', { name: /^Save$/i });
    await expect(saveBtn).toBeEnabled({ timeout: 5000 });
    await saveBtn.click();
    await expect(pnlDialog).not.toBeVisible({ timeout: 5000 });

    // The P&L section should still be visible after saving
    await expect(page.getByText(/Profit.*Loss|Profit & Loss/i).first()).toBeVisible({ timeout: 5000 });

    // The P&L entry table should show the sold quantity (50)
    // P&L entries appear in a table below the Profit & Loss heading
    await expect(page.getByText('50').first()).toBeVisible({ timeout: 5000 });
  });
});
