import { test, expect } from '../fixtures/auth';

test.describe('Auto-Sell Containers', () => {
  test('Alice enables auto-sell on Minerals Box container', async ({ alicePage }) => {
    // Clear localStorage so tree expansion state is fresh
    await alicePage.goto('/inventory');
    await alicePage.evaluate(() => localStorage.clear());
    await alicePage.goto('/inventory');

    // Wait for assets to load
    await expect(alicePage.getByText('Jita IV - Moon 4')).toBeVisible({ timeout: 30000 });

    // Expand Jita station
    await alicePage.getByText('Jita IV - Moon 4').click();

    // "Minerals Box" container should be visible after expanding
    await expect(alicePage.getByText('Minerals Box')).toBeVisible({ timeout: 10000 });

    // Click the auto-sell toggle button on the Minerals Box container
    // Use CSS selector to target only <button> elements (not div[role="button"] from ListItemButton)
    await alicePage.locator('button[aria-label="Enable Auto-Sell"]').click();

    // Auto-sell dialog should appear with container name
    const dialog = alicePage.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });
    await expect(dialog.getByText('Minerals Box')).toBeVisible();

    // Set percentage to 90%
    const percentageInput = dialog.getByLabel(/Price Percentage/i);
    await percentageInput.clear();
    await percentageInput.fill('90');

    // Save
    await dialog.getByRole('button', { name: /Save/i }).click();

    // Dialog should close and chip should appear on container
    await expect(dialog).not.toBeVisible({ timeout: 5000 });
    await expect(alicePage.getByText(/Auto-Sell @ 90% JBV/)).toBeVisible({ timeout: 5000 });
  });

  test('auto-sell listing appears on My Listings with Auto badge', async ({ alicePage }) => {
    await alicePage.goto('/marketplace');

    // My Listings tab should be active by default
    await expect(alicePage.getByRole('tab', { name: 'My Listings' })).toBeVisible();

    // Wait for the auto-created Isogen listing to appear
    // The controller triggers immediate sync when auto-sell is configured
    // Isogen is in the Minerals Box container, Jita buy = 50 ISK, 90% = 45 ISK
    await expect(alicePage.getByText('Isogen')).toBeVisible({ timeout: 15000 });

    // Verify the "Auto" badge is visible on the listing
    await expect(alicePage.getByText('Auto').first()).toBeVisible();
  });

  test('Alice updates auto-sell percentage', async ({ alicePage }) => {
    await alicePage.goto('/inventory');
    await alicePage.evaluate(() => localStorage.clear());
    await alicePage.goto('/inventory');

    // Wait for assets and expand Jita
    await expect(alicePage.getByText('Jita IV - Moon 4')).toBeVisible({ timeout: 30000 });
    await alicePage.getByText('Jita IV - Moon 4').click();

    // The container should show auto-sell chip
    await expect(alicePage.getByText(/Auto-Sell @ 90% JBV/)).toBeVisible({ timeout: 10000 });

    // Click the auto-sell button (aria-label now says "Edit Auto-Sell")
    await alicePage.locator('button[aria-label="Edit Auto-Sell"]').click();

    // Dialog should show "Edit Auto-Sell"
    const dialog = alicePage.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // Change percentage to 80%
    const percentageInput = dialog.getByLabel(/Price Percentage/i);
    await percentageInput.clear();
    await percentageInput.fill('80');

    // Save
    await dialog.getByRole('button', { name: /Save/i }).click();

    // Chip should update
    await expect(dialog).not.toBeVisible({ timeout: 5000 });
    await expect(alicePage.getByText(/Auto-Sell @ 80% JBV/)).toBeVisible({ timeout: 5000 });
  });

  test('Alice disables auto-sell and listing is removed', async ({ alicePage }) => {
    await alicePage.goto('/inventory');
    await alicePage.evaluate(() => localStorage.clear());
    await alicePage.goto('/inventory');

    // Wait for assets and expand Jita
    await expect(alicePage.getByText('Jita IV - Moon 4')).toBeVisible({ timeout: 30000 });
    await alicePage.getByText('Jita IV - Moon 4').click();

    // The container should show auto-sell chip
    await expect(alicePage.getByText(/Auto-Sell @ 80% JBV/)).toBeVisible({ timeout: 10000 });

    // Click the auto-sell button
    await alicePage.locator('button[aria-label="Edit Auto-Sell"]').click();

    // Dialog should show with "Disable" button
    const dialog = alicePage.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // Click "Disable" to remove auto-sell
    await dialog.getByRole('button', { name: /Disable/i }).click();

    // Dialog should close, auto-sell chip should disappear
    await expect(dialog).not.toBeVisible({ timeout: 5000 });
    await expect(alicePage.getByText(/Auto-Sell @/)).not.toBeVisible({ timeout: 5000 });

    // Verify listing is gone from marketplace
    await alicePage.goto('/marketplace');
    await expect(alicePage.getByRole('tab', { name: 'My Listings' })).toBeVisible();

    // Wait for the page to load, then verify Isogen is NOT listed
    await alicePage.waitForTimeout(2000);
    await expect(alicePage.getByRole('cell', { name: 'Isogen' })).not.toBeVisible();
  });
});
