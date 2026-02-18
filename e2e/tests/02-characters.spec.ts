import { test, expect } from '@playwright/test';

test.describe('Characters', () => {
  test('shows empty state with no characters', async ({ page }) => {
    await page.goto('/characters');

    await expect(page.getByText('No Characters')).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('link', { name: /Add Character/i })).toBeVisible();
  });

  test('add Alice Alpha via E2E API', async ({ page }) => {
    const response = await page.request.post('/api/e2e/add-character', {
      data: { userId: 1001, characterId: 2001001, characterName: 'Alice Alpha' },
    });
    expect(response.ok()).toBeTruthy();

    await page.goto('/characters');
    await expect(page.getByText('Alice Alpha')).toBeVisible({ timeout: 10000 });
  });

  test('add Alice Beta via E2E API', async ({ page }) => {
    const response = await page.request.post('/api/e2e/add-character', {
      data: { userId: 1001, characterId: 2001002, characterName: 'Alice Beta' },
    });
    expect(response.ok()).toBeTruthy();

    await page.goto('/characters');
    await expect(page.getByText('Alice Alpha')).toBeVisible({ timeout: 10000 });
    await expect(page.getByText('Alice Beta')).toBeVisible();
  });

  test('add remaining characters for downstream tests', async ({ page }) => {
    // Bob Miner's character
    let response = await page.request.post('/api/e2e/add-character', {
      data: { userId: 1002, characterId: 2002001, characterName: 'Bob Bravo' },
    });
    expect(response.ok()).toBeTruthy();

    // Charlie Trader's character
    response = await page.request.post('/api/e2e/add-character', {
      data: { userId: 1003, characterId: 2003001, characterName: 'Charlie Charlie' },
    });
    expect(response.ok()).toBeTruthy();

    // Diana Scout's character
    response = await page.request.post('/api/e2e/add-character', {
      data: { userId: 1004, characterId: 2004001, characterName: 'Diana Delta' },
    });
    expect(response.ok()).toBeTruthy();
  });

  test('character cards display portrait images', async ({ page }) => {
    await page.goto('/characters');

    await expect(page.getByText('Alice Alpha')).toBeVisible({ timeout: 10000 });

    // Character portraits use EVE Online image server with character ID
    const aliceImg = page.getByRole('img', { name: 'Alice Alpha' });
    await expect(aliceImg).toBeVisible();
    await expect(aliceImg).toHaveAttribute('src', /2001001/);
  });

  test('shows Add Character button', async ({ page }) => {
    await page.goto('/characters');

    await expect(page.getByText('Alice Alpha')).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole('link', { name: /Add Character/i })).toBeVisible();
  });

  test('displays Characters heading', async ({ page }) => {
    await page.goto('/characters');

    await expect(page.getByRole('heading', { name: 'Characters' })).toBeVisible({ timeout: 10000 });
  });
});
