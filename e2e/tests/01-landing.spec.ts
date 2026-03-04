import { test, expect } from '@playwright/test';

test.describe('Landing Page', () => {
  test('shows authenticated landing page', async ({ page }) => {
    await page.goto('/');

    await expect(page.getByRole('heading', { name: 'pinky.tools', level: 1 })).toBeVisible();
    await expect(page.getByText('Real-time asset tracking')).toBeVisible();
  });

  test('displays metric cards', async ({ page }) => {
    await page.goto('/');

    await expect(page.getByText('Asset Value')).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('Stockpile Deficit')).toBeVisible();
    await expect(page.getByText('Active Jobs')).toBeVisible();
  });

  test('shows Open Dashboard button for authenticated user', async ({ page }) => {
    await page.goto('/');

    // The new landing page has a single "Open Dashboard" CTA linking to /inventory
    await expect(page.getByRole('link', { name: 'Open Dashboard' })).toBeVisible();
  });

  test('Open Dashboard navigates to inventory page', async ({ page }) => {
    await page.goto('/');

    await page.getByRole('link', { name: 'Open Dashboard' }).click();
    await page.waitForURL('**/inventory');
    await expect(page).toHaveURL(/\/inventory/);
  });
});
