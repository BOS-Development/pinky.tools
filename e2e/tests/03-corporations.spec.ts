import { test, expect } from '@playwright/test';

test.describe('Corporations', () => {
  test('shows empty state with no corporations', async ({ page }) => {
    await page.goto('/corporations');

    await expect(page.getByText('No Corporations')).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('link', { name: /Add Corporation/i })).toBeVisible();
  });

  test('add Stargazer Industries via E2E API', async ({ page }) => {
    // Sends Alice Alpha's character data — mock ESI affiliation maps char 2001001 to corp 3001001
    const response = await page.request.post('/api/e2e/add-corporation', {
      data: { userId: 1001, characterId: 2001001, characterName: 'Alice Alpha' },
    });
    expect(response.ok()).toBeTruthy();

    await page.goto('/corporations');
    await expect(page.getByText('Stargazer Industries')).toBeVisible({ timeout: 10000 });
  });

  test('corporation card displays name and Corporation chip', async ({ page }) => {
    await page.goto('/corporations');

    await expect(page.getByText('Stargazer Industries')).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('Corporation', { exact: true })).toBeVisible();
  });

  test('shows Add Corporation button', async ({ page }) => {
    await page.goto('/corporations');

    await expect(page.getByText('Stargazer Industries')).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('link', { name: /Add Corporation/i })).toBeVisible();
  });

  test('displays Corporations heading', async ({ page }) => {
    await page.goto('/corporations');

    await expect(page.getByRole('heading', { name: 'Corporations' })).toBeVisible({ timeout: 10000 });
  });
});
