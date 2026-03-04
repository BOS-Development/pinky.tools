import { test, expect } from '../fixtures/auth';

test.describe('Marketplace', () => {
  test('marketplace page shows all tabs', async ({ alicePage }) => {
    await alicePage.goto('/marketplace');

    await expect(alicePage.getByRole('tab', { name: 'My Listings' })).toBeVisible();
    await expect(alicePage.getByRole('tab', { name: 'Browse' })).toBeVisible();
    await expect(alicePage.getByRole('tab', { name: 'Pending Sales' })).toBeVisible();
    await expect(alicePage.getByRole('tab', { name: 'History' })).toBeVisible();
    await expect(alicePage.getByRole('tab', { name: 'My Buy Orders' })).toBeVisible();
    await expect(alicePage.getByRole('tab', { name: 'Demand' })).toBeVisible();
    await expect(alicePage.getByRole('tab', { name: 'Analytics' })).toBeVisible();
  });

  test('Bob refreshes his assets and creates a for-sale listing', async ({ bobPage }) => {
    // Clear localStorage so tree expansion state is fresh
    await bobPage.goto('/inventory');
    await bobPage.evaluate(() => localStorage.clear());

    // Assets are automatically updated by the background runner (10s interval in E2E)
    // and also triggered immediately when characters are added
    await bobPage.goto('/inventory');
    await expect(bobPage.getByText('Jita IV - Moon 4')).toBeVisible({ timeout: 30000 });

    // Expand Jita station
    await bobPage.getByText('Jita IV - Moon 4').click();

    // Expand Personal Hangar to see items
    await bobPage.getByText('Personal Hangar').first().click();
    await expect(bobPage.getByText('Rifter')).toBeVisible({ timeout: 5000 });

    // Click sell button on Rifter row (lucide tag icon button)
    const rifterRow = bobPage.getByRole('row').filter({ hasText: 'Rifter' }).first();
    await rifterRow.locator('button:has(.lucide-tag)').click();

    // Fill in listing details
    await expect(bobPage.getByText(/List Item for Sale/i)).toBeVisible({ timeout: 5000 });

    // The listing dialog uses uncontrolled inputs (refs) with setTimeout(fn, 0)
    // to set initial values — wait for quantity to be populated before overwriting
    const qtyInput = bobPage.getByLabel(/Quantity/i).first();
    await expect(qtyInput).not.toHaveValue('', { timeout: 2000 });
    await qtyInput.clear();
    await qtyInput.fill('5');

    const priceInput = bobPage.getByLabel(/Price/i).first();
    await priceInput.clear();
    await priceInput.fill('550000');

    // Blur the price input so onBlur → setListingTotalValue re-render completes
    // before clicking Create Listing (prevents mousedown/mouseup targeting different DOM elements)
    await priceInput.blur();

    // Save listing
    await bobPage.getByRole('button', { name: /Create Listing/i }).click();

    // Verify listing appears on marketplace
    await bobPage.goto('/marketplace');
    await expect(bobPage.getByRole('tab', { name: 'My Listings' })).toBeVisible();
    await expect(bobPage.getByText('Rifter')).toBeVisible({ timeout: 10000 });
  });

  test('Alice can browse marketplace and see Bob listings', async ({ alicePage }) => {
    await alicePage.goto('/marketplace');

    // Click Browse tab
    await alicePage.getByRole('tab', { name: 'Browse' }).click();

    // Should see Bob's Rifter listing
    await expect(alicePage.getByText('Rifter')).toBeVisible({ timeout: 10000 });
    await expect(alicePage.getByText('Bob')).toBeVisible();
  });

  test('Alice purchases from Bob listing', async ({ alicePage }) => {
    await alicePage.goto('/marketplace');

    // Click Browse tab
    await alicePage.getByRole('tab', { name: 'Browse' }).click();
    await expect(alicePage.getByText('Rifter')).toBeVisible({ timeout: 10000 });

    // Click Buy button
    await alicePage.getByRole('button', { name: /Buy/i }).click();

    // Fill in purchase quantity (label not linked via htmlFor, scope to dialog)
    const dialog = alicePage.getByRole('dialog');
    await expect(dialog.getByText(/Quantity to Purchase/i)).toBeVisible();
    const qtyInput = dialog.getByRole('textbox');
    await qtyInput.clear();
    await qtyInput.fill('2');

    // Confirm purchase
    await alicePage.getByRole('button', { name: /Confirm Purchase/i }).click();

    // Check purchase appears in history
    await alicePage.getByRole('tab', { name: 'History' }).click();
    await expect(alicePage.getByText('Rifter')).toBeVisible({ timeout: 5000 });
  });

  test('Bob sees no duplicate items in pending sales', async ({ bobPage }) => {
    // Alice (user 1001) has 2 characters (Alice Alpha + Alice Beta).
    // The bug caused a LEFT JOIN to the characters table to produce one row
    // per character, duplicating every pending sale for multi-character buyers.
    await bobPage.goto('/marketplace');

    // Click Pending Sales tab
    await bobPage.getByRole('tab', { name: 'Pending Sales' }).click();

    // Wait for the pending sale from Alice's Rifter purchase to load
    await expect(bobPage.getByText('Rifter')).toBeVisible({ timeout: 10000 });

    // The header should show exactly 1 item in 1 group — not 2 items (which would mean duplication)
    await expect(bobPage.getByText(/1 items? in 1 groups?/)).toBeVisible();

    // Rifter should appear exactly once in the table body
    const rifterRows = bobPage.getByRole('cell', { name: 'Rifter' });
    await expect(rifterRows).toHaveCount(1);
  });

  test('Alice can create a buy order', async ({ alicePage }) => {
    await alicePage.goto('/marketplace');

    // Click My Buy Orders tab
    await alicePage.getByRole('tab', { name: 'My Buy Orders' }).click();

    // Click Create Buy Order button
    await alicePage.getByRole('button', { name: /Create Buy Order/i }).click();

    // Search for item — click the combobox trigger to open the popover, then fill the search input
    await alicePage.getByRole('combobox').first().click();
    await alicePage.getByPlaceholder('Search items...').fill('Pyerite');

    // Select from dropdown (button accessible name includes image alt + span text)
    await alicePage.getByRole('button', { name: /Pyerite/i }).first().click();

    // Search for station — click the combobox trigger to open the popover
    await alicePage.getByRole('combobox').nth(1).click();
    await alicePage.getByPlaceholder('Search stations...').fill('Jita');

    // Select Jita station from dropdown
    await alicePage.getByRole('button', { name: /Jita/i }).first().click();

    // Fill in quantity and price (labels not linked via htmlFor, use spinbutton role)
    // Three number inputs: Quantity Desired, Min Price (optional), Max Price
    const qtyInput = alicePage.getByRole('spinbutton').first();
    await qtyInput.fill('10000');

    const priceInput = alicePage.getByRole('spinbutton').nth(2);
    await priceInput.fill('10');

    // Create the order
    await alicePage.getByRole('button', { name: /Create/i }).click();

    // Verify buy order appears with location
    await expect(alicePage.getByText('Pyerite').first()).toBeVisible({ timeout: 5000 });
    await expect(alicePage.getByText('Jita').first()).toBeVisible({ timeout: 5000 });
  });

  test('can switch between all marketplace tabs', async ({ alicePage }) => {
    await alicePage.goto('/marketplace');

    const tabs = ['My Listings', 'Browse', 'Pending Sales', 'History', 'My Buy Orders', 'Demand', 'Analytics'];

    for (const tabName of tabs) {
      await alicePage.getByRole('tab', { name: tabName }).click();
      await expect(alicePage.getByRole('tab', { name: tabName })).toHaveAttribute('aria-selected', 'true');
    }
  });
});
