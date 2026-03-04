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

  test('shows quick-access navigation grid for authenticated user', async ({ page }) => {
    await page.goto('/');

    // Quick-access grid replaces the old "Open Dashboard" CTA
    await expect(page.getByRole('link', { name: /Inventory/ }).filter({ hasText: 'View all assets' })).toBeVisible();
    await expect(page.getByRole('link', { name: /Stockpiles/ }).filter({ hasText: 'Track targets' })).toBeVisible();
    await expect(page.getByRole('link', { name: /Industry/ }).filter({ hasText: 'Manage jobs' })).toBeVisible();
  });

  test('Inventory quick-link navigates to inventory page', async ({ page }) => {
    await page.goto('/');

    await page.getByRole('link', { name: /Inventory/ }).filter({ hasText: 'View all assets' }).click();
    await page.waitForURL('**/inventory');
    await expect(page).toHaveURL(/\/inventory/);
  });
});
