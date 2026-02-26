import { test, expect } from '@playwright/test';

test.describe('Production Plans', () => {
  test('navigate to production plans page shows heading', async ({ page }) => {
    await page.goto('/production-plans');

    await expect(page.getByRole('heading', { name: 'Production Plans' })).toBeVisible({ timeout: 10000 });
  });

  test('empty state is shown when no plans exist', async ({ page }) => {
    await page.goto('/production-plans');

    await expect(
      page.getByText('No production plans yet. Create one to define how items should be produced.')
    ).toBeVisible({ timeout: 10000 });
  });

  test('New Plan button opens Create Production Plan dialog', async ({ page }) => {
    await page.goto('/production-plans');

    await page.getByRole('button', { name: /New Plan/i }).click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });
    await expect(dialog.getByText('Create Production Plan')).toBeVisible();

    // Blueprint search input should be present
    await expect(dialog.getByLabel('Search for a product')).toBeVisible();

    // Optional station fields should be present
    await expect(dialog.getByLabel('Default Manufacturing Station')).toBeVisible();
    await expect(dialog.getByLabel('Default Reaction Station')).toBeVisible();

    // Close dialog
    await dialog.getByRole('button', { name: /Cancel/i }).click();
    await expect(dialog).not.toBeVisible({ timeout: 5000 });
  });

  test('blueprint search autocomplete returns Rifter result', async ({ page }) => {
    await page.goto('/production-plans');

    await page.getByRole('button', { name: /New Plan/i }).click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // Type "Ri" first then extend — autocomplete fires after 300ms debounce with 2+ chars
    const searchInput = dialog.getByLabel('Search for a product');
    await searchInput.fill('Ri');
    await searchInput.fill('Rifter');

    // Dropdown should show Rifter manufacturing option (may have multiple variants — use first)
    await expect(
      page.getByRole('option', { name: /Rifter.*manufacturing/i }).first()
    ).toBeVisible({ timeout: 10000 });

    // Press Escape to close the autocomplete dropdown before clicking Cancel
    // (the open dropdown intercepts pointer events on the Cancel button)
    await searchInput.press('Escape');
    await dialog.getByRole('button', { name: /Cancel/i }).click();
    await expect(dialog).not.toBeVisible({ timeout: 5000 });
  });

  test('create a Rifter production plan', async ({ page }) => {
    await page.goto('/production-plans');

    // Open create dialog
    await page.getByRole('button', { name: /New Plan/i }).click();
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // Search for Rifter blueprint
    const searchInput = dialog.getByLabel('Search for a product');
    await searchInput.fill('Rifter');

    // Wait for the autocomplete option and select the first match (may have variants)
    const option = page.getByRole('option', { name: /Rifter.*manufacturing/i }).first();
    await expect(option).toBeVisible({ timeout: 10000 });
    await option.click();

    // Create button should now be enabled — click it
    await dialog.getByRole('button', { name: /Create Plan/i }).click();

    // Dialog closes and editor view opens automatically after creation
    await expect(dialog).not.toBeVisible({ timeout: 5000 });

    // The editor shows the plan name (product name is used as default plan name)
    await expect(page.getByRole('heading', { name: /Rifter/i })).toBeVisible({ timeout: 10000 });

    // Editor also shows step count and the Generate Jobs button
    await expect(page.getByText(/production step/i)).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('button', { name: /Generate Jobs/i })).toBeVisible();
  });

  test('plan editor shows Step Tree tab', async ({ page }) => {
    await page.goto('/production-plans');

    // The Rifter plan created in the previous test should be in the list
    await expect(page.getByText('Rifter').first()).toBeVisible({ timeout: 10000 });

    // Click the edit icon on the Rifter plan row to open the editor
    const rifterRow = page.getByRole('row').filter({ hasText: 'Rifter' }).first();
    await rifterRow.getByRole('button').first().click();

    // Editor loads — verify plan name and tabs
    await expect(page.getByRole('heading', { name: /Rifter/i })).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('tab', { name: 'Step Tree' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'Batch Configure' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'Transport' })).toBeVisible();
  });

  test('Back to Plans button returns to list view', async ({ page }) => {
    await page.goto('/production-plans');

    // Wait for the plan row and open the editor
    await expect(page.getByText('Rifter').first()).toBeVisible({ timeout: 10000 });
    const rifterRow = page.getByRole('row').filter({ hasText: 'Rifter' }).first();
    await rifterRow.getByRole('button').first().click();

    // Verify we're in editor view
    await expect(page.getByRole('heading', { name: /Rifter/i })).toBeVisible({ timeout: 10000 });

    // Click "Back to Plans"
    await page.getByRole('button', { name: /Back to Plans/i }).click();

    // Should return to the list view with heading and table
    await expect(page.getByRole('heading', { name: 'Production Plans' })).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Rifter').first()).toBeVisible({ timeout: 5000 });
  });

  test('delete plan removes it from the list', async ({ page }) => {
    await page.goto('/production-plans');

    // Wait for the Rifter plan row
    await expect(page.getByText('Rifter').first()).toBeVisible({ timeout: 10000 });

    // Click the delete icon (last button) on the Rifter row
    const rifterRow = page.getByRole('row').filter({ hasText: 'Rifter' }).first();
    await rifterRow.getByRole('button').last().click();

    // Plan should be removed — empty state message should reappear
    await expect(
      page.getByText('No production plans yet. Create one to define how items should be produced.')
    ).toBeVisible({ timeout: 10000 });
  });
});
