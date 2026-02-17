import { test, expect } from '@playwright/test';

test.describe('Characters Page', () => {
  test('displays seeded characters for Alice', async ({ page }) => {
    await page.goto('/characters');

    await expect(page.getByText('Alice Alpha')).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('Alice Beta')).toBeVisible();
  });

  test('shows Add Character and Refresh Assets buttons', async ({ page }) => {
    await page.goto('/characters');

    await expect(page.getByRole('link', { name: /Add Character/i })).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('link', { name: /Refresh Assets/i })).toBeVisible();
  });
});
