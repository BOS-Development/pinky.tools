import { test, expect } from '@playwright/test';

test.describe('Settings', () => {
  test('navigate to settings page shows Settings heading', async ({ page }) => {
    await page.goto('/settings');

    await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible({ timeout: 10000 });
  });

  test('Discord settings section is visible in unlinked state', async ({ page }) => {
    await page.goto('/settings');

    // Discord Notifications card heading is visible
    await expect(page.getByText('Discord Notifications')).toBeVisible({ timeout: 10000 });

    // Unlinked state: descriptive text and link button are shown
    await expect(
      page.getByText('Link your Discord account to receive notifications when marketplace events occur.')
    ).toBeVisible();

    await expect(page.getByRole('link', { name: /Link Discord Account/i })).toBeVisible();
  });

  test('page renders without errors', async ({ page }) => {
    await page.goto('/settings');

    // Page loads and the main container content is visible
    await expect(page.getByRole('heading', { name: 'Settings' })).toBeVisible({ timeout: 10000 });

    // No error boundary or crash message
    await expect(page.getByText(/something went wrong/i)).not.toBeVisible();
    await expect(page.getByText(/application error/i)).not.toBeVisible();

    // Notification Targets section is NOT shown when Discord is unlinked
    await expect(page.getByText('Notification Targets')).not.toBeVisible();
  });
});
