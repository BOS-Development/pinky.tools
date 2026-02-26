import { test, expect } from '@playwright/test';

test.describe('Stations', () => {
  test('navigate to stations page shows Preferred Stations heading', async ({ page }) => {
    await page.goto('/stations');

    await expect(page.getByRole('heading', { name: 'Preferred Stations' })).toBeVisible({ timeout: 10000 });
  });

  test('stations page shows empty state initially', async ({ page }) => {
    await page.goto('/stations');

    // Table renders but no stations are configured yet
    await expect(
      page.getByText('No preferred stations configured. Click "Add Station" to get started.')
    ).toBeVisible({ timeout: 10000 });
  });

  test('open add station dialog', async ({ page }) => {
    await page.goto('/stations');

    // Click "Add Station" button
    await page.getByRole('button', { name: /Add Station/i }).click();

    // Dialog should appear with title "Add Station"
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });
    await expect(dialog.getByText('Add Station')).toBeVisible();

    // Dialog should have Station autocomplete field
    await expect(dialog.getByLabel('Station')).toBeVisible();

    // Close dialog
    await dialog.getByRole('button', { name: /Cancel/i }).click();
    await expect(dialog).not.toBeVisible({ timeout: 5000 });
  });

  test('station autocomplete searches and shows results for Jita', async ({ page }) => {
    await page.goto('/stations');

    // Open Add Station dialog
    await page.getByRole('button', { name: /Add Station/i }).click();
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // Type "Jita" into the station search autocomplete
    const stationInput = dialog.getByLabel('Station');
    await stationInput.fill('Jita');

    // Wait for autocomplete dropdown to appear with Jita station
    // The search has a 300ms debounce, so wait for the option to appear
    await expect(
      page.getByRole('option', { name: /Jita IV - Moon 4/i })
    ).toBeVisible({ timeout: 5000 });

    // Close dialog
    await dialog.getByRole('button', { name: /Cancel/i }).click();
    await expect(dialog).not.toBeVisible({ timeout: 5000 });
  });

  test('create station with Jita and Raitaru structure', async ({ page }) => {
    await page.goto('/stations');

    // Open Add Station dialog
    await page.getByRole('button', { name: /Add Station/i }).click();
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // Search for and select Jita station
    const stationInput = dialog.getByLabel('Station');
    await stationInput.fill('Jita');
    await expect(
      page.getByRole('option', { name: /Jita IV - Moon 4/i })
    ).toBeVisible({ timeout: 5000 });
    await page.getByRole('option', { name: /Jita IV - Moon 4/i }).click();

    // Verify structure is already set to Raitaru (default)
    // Set facility tax to 1%
    const taxInput = dialog.getByLabel('Facility Tax %');
    await taxInput.clear();
    await taxInput.fill('1');

    // Save the station (button says "Add" for new stations)
    await dialog.getByRole('button', { name: /^Add$/i }).click();

    // Dialog should close
    await expect(dialog).not.toBeVisible({ timeout: 5000 });

    // The station should now appear in the table
    await expect(page.getByRole('cell', { name: /Jita IV - Moon 4/i })).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('cell', { name: /raitaru/i })).toBeVisible();
    await expect(page.getByRole('cell', { name: '1%' })).toBeVisible();
  });

  test('edit station changes facility tax', async ({ page }) => {
    await page.goto('/stations');

    // Wait for the Jita station row to be visible (created in previous test)
    await expect(page.getByRole('cell', { name: /Jita IV - Moon 4/i })).toBeVisible({ timeout: 10000 });

    // Click the edit icon button on the Jita station row
    const jitaRow = page.getByRole('row').filter({ hasText: /Jita IV - Moon 4/ });
    await jitaRow.getByRole('button').first().click();

    // Dialog should open as "Edit Station" with station search disabled
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });
    await expect(dialog.getByText('Edit Station')).toBeVisible();

    // Station name should be pre-filled and disabled
    const stationInput = dialog.getByLabel('Station');
    await expect(stationInput).toBeDisabled();

    // Change facility tax to 2%
    const taxInput = dialog.getByLabel('Facility Tax %');
    await taxInput.clear();
    await taxInput.fill('2');

    // Save (button says "Update" for edits)
    await dialog.getByRole('button', { name: /^Update$/i }).click();

    // Dialog should close
    await expect(dialog).not.toBeVisible({ timeout: 5000 });

    // Verify updated tax value in the table
    await expect(page.getByRole('cell', { name: '2%' })).toBeVisible({ timeout: 5000 });
  });

  test('delete station removes it from the table', async ({ page }) => {
    await page.goto('/stations');

    // Wait for the Jita station row
    await expect(page.getByRole('cell', { name: /Jita IV - Moon 4/i })).toBeVisible({ timeout: 10000 });

    // Accept the native browser confirm dialog
    page.on('dialog', dialog => dialog.accept());

    // Click the delete icon button on the Jita station row
    const jitaRow = page.getByRole('row').filter({ hasText: /Jita IV - Moon 4/ });
    await jitaRow.getByRole('button').last().click();

    // Station should be removed and empty state should appear
    await expect(
      page.getByText('No preferred stations configured. Click "Add Station" to get started.')
    ).toBeVisible({ timeout: 5000 });
  });
});
