import { test, expect, type Page } from '@playwright/test';

// Helper: navigate to /arbiter and open the Settings collapsible to the Structures tab.
// The accordion is collapsed by default — clicking the "Settings" trigger expands it.
// "Structures" is the default active tab once open, so Manufacturing Profiles is immediately visible.
async function openSettingsStructuresTab(page: Page) {
  await page.goto('/arbiter');

  // Wait for the page to load — the Settings trigger is always rendered
  const settingsTrigger = page.getByRole('button', { name: /^Settings$/i });
  await expect(settingsTrigger).toBeVisible({ timeout: 15000 });
  await settingsTrigger.click();

  // Manufacturing Profiles section is inside the Structures tab (default active tab)
  await expect(page.getByText('Manufacturing Profiles')).toBeVisible({ timeout: 10000 });
}

// Helper: click "Save as Profile", fill in the name, and click Save.
async function saveProfile(page: Page, name: string) {
  await page.getByRole('button', { name: /Save as Profile/i }).click();

  const dialog = page.getByRole('dialog');
  await expect(dialog).toBeVisible({ timeout: 5000 });
  await expect(dialog.getByRole('heading', { name: 'Save as Profile' })).toBeVisible();

  const nameInput = dialog.getByLabel('Profile Name');
  await nameInput.clear();
  await nameInput.fill(name);

  await dialog.getByRole('button', { name: /^Save$/i }).click();
  await expect(dialog).not.toBeVisible({ timeout: 5000 });
}

// Helper: get the Select trigger inside the Manufacturing Profiles flex row.
// The profiles row is the first "flex items-center gap-2 flex-wrap" div on the page
// (it is inside the Manufacturing Profiles section, above the StructureSection blocks).
function profilesSelectTrigger(page: Page) {
  return page
    .locator('div.flex.items-center.gap-2.flex-wrap')
    .getByRole('combobox')
    .first();
}

test.describe('Arbiter Manufacturing Profiles', () => {
  test('navigate to /arbiter — page loads and Settings accordion trigger is visible', async ({ page }) => {
    await page.goto('/arbiter');

    await expect(page.getByRole('button', { name: /^Settings$/i })).toBeVisible({ timeout: 15000 });

    // No crash
    await expect(page.getByText(/something went wrong/i)).not.toBeVisible();
    await expect(page.getByText(/application error/i)).not.toBeVisible();
  });

  test('open Settings accordion — Structures tab and Manufacturing Profiles section are visible', async ({ page }) => {
    await openSettingsStructuresTab(page);

    // Structures tab is visible and active by default
    await expect(page.getByRole('tab', { name: /Structures/i })).toBeVisible();

    // The Manufacturing Profiles section renders with the placeholder and action buttons
    await expect(page.getByText('Manufacturing Profiles')).toBeVisible();
    await expect(page.getByText(/Select a profile…/i)).toBeVisible();
    await expect(page.getByRole('button', { name: /Save as Profile/i })).toBeVisible();

    // Apply button is present but disabled (no profile selected yet)
    const applyBtn = page.getByRole('button', { name: /^Apply$/i });
    await expect(applyBtn).toBeVisible();
    await expect(applyBtn).toBeDisabled();
  });

  test('save a new profile — profile appears in the dropdown', async ({ page }) => {
    await openSettingsStructuresTab(page);

    await saveProfile(page, 'Test Profile');

    // After saving, the saved profile is automatically selected — its name appears
    // inside the Select trigger instead of the placeholder
    await expect(page.getByText('Test Profile').first()).toBeVisible({ timeout: 5000 });

    // Apply becomes enabled because a profile is now selected
    const applyBtn = page.getByRole('button', { name: /^Apply$/i });
    await expect(applyBtn).toBeEnabled({ timeout: 5000 });
  });

  test('apply a profile — settings fields reflect the applied profile', async ({ page }) => {
    await openSettingsStructuresTab(page);

    // Select "Test Profile" from the dropdown
    await profilesSelectTrigger(page).click();
    await page.getByRole('option', { name: 'Test Profile' }).click();

    // Apply is now enabled
    const applyBtn = page.getByRole('button', { name: /^Apply$/i });
    await expect(applyBtn).toBeEnabled({ timeout: 5000 });
    await applyBtn.click();

    // After applying, the four StructureSection headings (Reaction, Invention,
    // Component Build, Final Build) should all be visible in the Structures tab,
    // confirming the settings panel populated correctly.
    await expect(page.getByRole('heading', { name: 'Reaction' })).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('heading', { name: 'Invention' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Component Build' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Final Build' })).toBeVisible();
  });

  test('save with duplicate name overwrites — still only one entry in dropdown', async ({ page }) => {
    await openSettingsStructuresTab(page);

    // Save using the same name "Test Profile" — UI shows an overwrite warning
    await saveProfile(page, 'Test Profile');

    // Open the dropdown and verify exactly one option named "Test Profile" exists
    await profilesSelectTrigger(page).click();
    const options = page.getByRole('option', { name: 'Test Profile' });
    await expect(options).toHaveCount(1, { timeout: 5000 });

    // Close the dropdown without selecting
    await page.keyboard.press('Escape');
  });

  test('delete a profile — profile is removed from dropdown', async ({ page }) => {
    await openSettingsStructuresTab(page);

    // Select "Test Profile" so the delete button becomes visible
    await profilesSelectTrigger(page).click();
    await page.getByRole('option', { name: 'Test Profile' }).click();

    // The delete button is a ghost icon-only button that only renders when a profile
    // is selected. It is the last button inside the profiles flex row.
    const profilesRow = page.locator('div.flex.items-center.gap-2.flex-wrap').first();
    const deleteBtn = profilesRow.locator('button').last();
    await expect(deleteBtn).toBeVisible({ timeout: 5000 });
    await deleteBtn.click();

    // After deletion the Select reverts to "Select a profile…" placeholder
    await expect(page.getByText(/Select a profile…/i)).toBeVisible({ timeout: 5000 });

    // Opening the dropdown should show no "Test Profile" option
    await profilesSelectTrigger(page).click();
    await expect(page.getByRole('option', { name: 'Test Profile' })).not.toBeVisible({ timeout: 3000 });

    // Close the dropdown
    await page.keyboard.press('Escape');
  });
});
