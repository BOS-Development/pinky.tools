import { test, expect } from '@playwright/test';

test.describe('Landing Page', () => {
  test('shows authenticated landing page', async ({ page }) => {
    await page.goto('/');

    await expect(page.getByText('Master Your EVE Online Assets')).toBeVisible();
    await expect(page.getByText('Real-time asset tracking')).toBeVisible();
  });

  test('displays metric cards', async ({ page }) => {
    await page.goto('/');

    await expect(page.getByText('Total Asset Value')).toBeVisible({ timeout: 10000 });
    // Deficit card label is either "Stockpile Deficit Cost" or "All Stockpiles Met"
    await expect(page.getByText(/Stockpile Deficit Cost|All Stockpiles Met/)).toBeVisible();
  });

  test('shows navigation buttons for authenticated user', async ({ page }) => {
    await page.goto('/');

    // Characters link exists in both navbar and landing page; use .first()
    await expect(page.getByRole('link', { name: 'Characters' }).first()).toBeVisible();
    await expect(page.getByRole('link', { name: 'View Assets' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Manage Stockpiles' })).toBeVisible();
  });

  test('characters link navigates to characters page', async ({ page }) => {
    await page.goto('/');

    await page.getByRole('link', { name: 'Characters' }).first().click();
    await page.waitForURL('**/characters');
    await expect(page).toHaveURL(/\/characters/);
  });
});
