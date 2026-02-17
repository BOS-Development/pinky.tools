import { test, expect } from '@playwright/test';

test.describe('Navigation', () => {
  test('navbar displays all navigation links', async ({ page }) => {
    await page.goto('/');

    const navbar = page.locator('header');
    await expect(navbar.getByText('EVE Industry Tool')).toBeVisible();
    await expect(navbar.getByRole('link', { name: 'Characters' })).toBeVisible();
    await expect(navbar.getByRole('link', { name: 'Corporations' })).toBeVisible();
    await expect(navbar.getByRole('link', { name: 'Inventory' })).toBeVisible();
    await expect(navbar.getByRole('link', { name: 'Stockpiles' })).toBeVisible();
    await expect(navbar.getByRole('link', { name: /Contacts/ })).toBeVisible();
    await expect(navbar.getByRole('link', { name: 'Marketplace' })).toBeVisible();
  });

  test('characters page loads', async ({ page }) => {
    await page.goto('/characters');
    await expect(page.getByRole('heading', { name: /Characters/ })).toBeVisible({ timeout: 10000 });
  });

  test('corporations page loads', async ({ page }) => {
    await page.goto('/corporations');
    await expect(page.getByRole('heading', { name: /Corporations/ })).toBeVisible({ timeout: 10000 });
  });

  test('inventory page loads', async ({ page }) => {
    await page.goto('/inventory');
    // Shows "Asset Inventory" when assets exist, "No Assets Found" when empty
    await expect(page.getByText(/Asset Inventory|No Assets Found/)).toBeVisible({ timeout: 10000 });
  });

  test('stockpiles page loads', async ({ page }) => {
    await page.goto('/stockpiles');
    await expect(page.getByText('Stockpiles Needing Replenishment')).toBeVisible({ timeout: 10000 });
  });

  test('contacts page loads', async ({ page }) => {
    await page.goto('/contacts');
    await expect(page.getByRole('heading', { name: /Contacts/ })).toBeVisible({ timeout: 10000 });
  });

  test('marketplace page loads', async ({ page }) => {
    await page.goto('/marketplace');
    await expect(page.getByRole('tab', { name: 'My Listings' })).toBeVisible({ timeout: 10000 });
  });
});
