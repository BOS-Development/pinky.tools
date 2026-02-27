import { test, expect } from '@playwright/test';

test.describe('Navigation', () => {
  test('navbar displays all dropdown triggers and Settings link', async ({ page }) => {
    await page.goto('/');

    const navbar = page.locator('header');
    await expect(navbar.getByText('EVE Industry Tool')).toBeVisible();

    // The navbar has 5 dropdown trigger buttons and a standalone Settings link
    await expect(navbar.getByRole('button', { name: /Account/i })).toBeVisible();
    await expect(navbar.getByRole('button', { name: /Assets/i })).toBeVisible();
    await expect(navbar.getByRole('button', { name: /Trading/i })).toBeVisible();
    await expect(navbar.getByRole('button', { name: /Industry/i })).toBeVisible();
    await expect(navbar.getByRole('button', { name: /Logistics/i })).toBeVisible();
    await expect(navbar.getByRole('link', { name: 'Settings' })).toBeVisible();

    // Open Account dropdown and verify menu items, then close
    await navbar.getByRole('button', { name: /Account/i }).click();
    await expect(page.getByRole('menuitem', { name: 'Characters' })).toBeVisible({ timeout: 3000 });
    await expect(page.getByRole('menuitem', { name: 'Corporations' })).toBeVisible();
    await page.keyboard.press('Escape');

    // Open Assets dropdown and verify menu items, then close
    await navbar.getByRole('button', { name: /Assets/i }).click();
    await expect(page.getByRole('menuitem', { name: 'Inventory' })).toBeVisible({ timeout: 3000 });
    await expect(page.getByRole('menuitem', { name: 'Stockpiles' })).toBeVisible();
    await page.keyboard.press('Escape');

    // Open Trading dropdown and verify menu items, then close
    await navbar.getByRole('button', { name: /Trading/i }).click();
    await expect(page.getByRole('menuitem', { name: /Contacts/ })).toBeVisible({ timeout: 3000 });
    await expect(page.getByRole('menuitem', { name: 'Marketplace' })).toBeVisible();
    await page.keyboard.press('Escape');
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

  test('shows scope warning banner when characters have outdated scopes', async ({ page }) => {
    await page.goto('/characters');
    // Characters and corps added by specs 02 and 03 have a subset of required scopes,
    // so the global navbar banner should be visible on any page
    await expect(page.getByText('Some characters or corporations need to be re-authorized')).toBeVisible({ timeout: 15000 });
  });
});
