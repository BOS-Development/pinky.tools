import { test, expect } from '@playwright/test';

test.describe('Transport', () => {
  test('navigate to transport page shows heading and tabs', async ({ page }) => {
    await page.goto('/transport');

    // Page heading (h5 Typography component renders as an h5 element)
    await expect(page.getByRole('heading', { name: 'Transport' })).toBeVisible({ timeout: 10000 });

    // All three tabs should be visible
    await expect(page.getByRole('tab', { name: 'Transport Jobs' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'Transport Profiles' })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'JF Routes' })).toBeVisible();
  });

  test('Transport Jobs tab shows empty state', async ({ page }) => {
    await page.goto('/transport');

    // Transport Jobs tab is active by default (index 0)
    await expect(page.getByRole('tab', { name: 'Transport Jobs' })).toBeVisible({ timeout: 10000 });

    // No jobs exist yet — empty state message
    await expect(page.getByText('No transport jobs')).toBeVisible({ timeout: 10000 });
  });

  test('Transport Profiles tab shows empty state', async ({ page }) => {
    await page.goto('/transport');

    // Click the Transport Profiles tab
    await page.getByRole('tab', { name: 'Transport Profiles' }).click();

    // No profiles exist yet — empty state message
    await expect(page.getByText('No transport profiles configured')).toBeVisible({ timeout: 10000 });
  });

  test('JF Routes tab shows empty state', async ({ page }) => {
    await page.goto('/transport');

    // Click the JF Routes tab
    await page.getByRole('tab', { name: 'JF Routes' }).click();

    // No routes exist yet — empty state message
    await expect(page.getByText('No JF routes configured')).toBeVisible({ timeout: 10000 });
  });

  test('open add transport profile dialog', async ({ page }) => {
    await page.goto('/transport');

    // Navigate to the Transport Profiles tab
    await page.getByRole('tab', { name: 'Transport Profiles' }).click();
    await expect(page.getByText('No transport profiles configured')).toBeVisible({ timeout: 10000 });

    // Click "Add Profile"
    await page.getByRole('button', { name: /Add Profile/i }).click();

    // Dialog should appear with "Add Transport Profile" title
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });
    await expect(dialog.getByText('Add Transport Profile')).toBeVisible();

    // Dialog should have the Profile Name field
    await expect(dialog.getByLabel('Profile Name')).toBeVisible();

    // Transport Method dropdown should default to Freighter
    await expect(dialog.getByText('Freighter')).toBeVisible();

    // Cancel and verify dialog closes
    await dialog.getByRole('button', { name: /Cancel/i }).click();
    await expect(dialog).not.toBeVisible({ timeout: 5000 });
  });

  test('create a freighter transport profile', async ({ page }) => {
    await page.goto('/transport');

    // Navigate to the Transport Profiles tab
    await page.getByRole('tab', { name: 'Transport Profiles' }).click();
    // Wait for loading to complete: the "Add Profile" button appears only after loading=false.
    // Then verify the empty state text is present.
    await expect(page.getByRole('button', { name: /Add Profile/i })).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('No transport profiles configured')).toBeVisible({ timeout: 5000 });

    // Open add dialog
    await page.getByRole('button', { name: /Add Profile/i }).click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // Fill in the profile name
    await dialog.getByLabel('Profile Name').fill('Jita Freighter');

    // Transport Method is already "Freighter" — leave as is

    // Set cargo capacity
    const cargoInput = dialog.getByLabel('Cargo Capacity (m3)');
    await cargoInput.clear();
    await cargoInput.fill('860000');

    // Set rate per m3 per jump (visible for non-JF methods)
    const rateInput = dialog.getByLabel('Rate per m3 per Jump (ISK)');
    await rateInput.clear();
    await rateInput.fill('200');

    // Create the profile
    await dialog.getByRole('button', { name: /^Create$/i }).click();

    // Dialog should close
    await expect(dialog).not.toBeVisible({ timeout: 5000 });

    // Profile should appear in the table
    await expect(page.getByRole('cell', { name: 'Jita Freighter' })).toBeVisible({ timeout: 5000 });
    // Use exact: true to match only the transport method chip ("Freighter"),
    // not the profile name cell which contains the substring "Freighter" ("Jita Freighter")
    await expect(page.getByText('Freighter', { exact: true })).toBeVisible();
  });

  test('edit transport profile changes cargo capacity', async ({ page }) => {
    await page.goto('/transport');

    // Navigate to the Transport Profiles tab
    await page.getByRole('tab', { name: 'Transport Profiles' }).click();

    // Wait for the profile row created in the previous test
    await expect(page.getByRole('cell', { name: 'Jita Freighter' })).toBeVisible({ timeout: 10000 });

    // Click the edit icon button on the profile row
    const profileRow = page.getByRole('row').filter({ hasText: 'Jita Freighter' });
    await profileRow.getByRole('button').first().click();

    // Dialog should open as "Edit Transport Profile"
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });
    await expect(dialog.getByText('Edit Transport Profile')).toBeVisible();

    // Update the profile name
    const nameInput = dialog.getByLabel('Profile Name');
    await nameInput.clear();
    await nameInput.fill('Jita Freighter XL');

    // Update cargo capacity
    const cargoInput = dialog.getByLabel('Cargo Capacity (m3)');
    await cargoInput.clear();
    await cargoInput.fill('1200000');

    // Save
    await dialog.getByRole('button', { name: /^Update$/i }).click();

    // Dialog should close
    await expect(dialog).not.toBeVisible({ timeout: 5000 });

    // Verify updated name appears in the table
    await expect(page.getByRole('cell', { name: 'Jita Freighter XL' })).toBeVisible({ timeout: 5000 });
  });

  test('delete transport profile removes it from the table', async ({ page }) => {
    await page.goto('/transport');

    // Navigate to the Transport Profiles tab
    await page.getByRole('tab', { name: 'Transport Profiles' }).click();

    // Wait for the profile row created in previous tests
    await expect(page.getByRole('cell', { name: 'Jita Freighter XL' })).toBeVisible({ timeout: 10000 });

    // Accept the native browser confirm dialog
    page.on('dialog', dialog => dialog.accept());

    // Click the delete icon button on the profile row
    const profileRow = page.getByRole('row').filter({ hasText: 'Jita Freighter XL' });
    await profileRow.getByRole('button').last().click();

    // Profile should be removed and empty state should appear
    await expect(page.getByText('No transport profiles configured')).toBeVisible({ timeout: 5000 });
  });

  test('open add JF route dialog and cancel', async ({ page }) => {
    await page.goto('/transport');

    // Navigate to the JF Routes tab
    await page.getByRole('tab', { name: 'JF Routes' }).click();
    await expect(page.getByText('No JF routes configured')).toBeVisible({ timeout: 10000 });

    // Click "Add JF Route"
    await page.getByRole('button', { name: /Add JF Route/i }).click();

    // Dialog should appear with "Add JF Route" title
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });
    await expect(dialog.getByText('Add JF Route')).toBeVisible();

    // Dialog should have Route Name field and Origin/Destination system inputs
    await expect(dialog.getByLabel('Route Name')).toBeVisible();
    await expect(dialog.getByLabel('Origin System')).toBeVisible();
    await expect(dialog.getByLabel('Destination System')).toBeVisible();

    // Cancel and verify dialog closes
    await dialog.getByRole('button', { name: /Cancel/i }).click();
    await expect(dialog).not.toBeVisible({ timeout: 5000 });
  });

  test('open add transport job dialog and cancel', async ({ page }) => {
    await page.goto('/transport');

    // Transport Jobs tab is the default (index 0)
    await expect(page.getByText('No transport jobs')).toBeVisible({ timeout: 10000 });

    // Click "Create Transport Job"
    await page.getByRole('button', { name: /Create Transport Job/i }).click();

    // Dialog should appear
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // Cancel and verify dialog closes
    await dialog.getByRole('button', { name: /Cancel/i }).click();
    await expect(dialog).not.toBeVisible({ timeout: 5000 });
  });
});
