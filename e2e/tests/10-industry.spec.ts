import { test, expect } from '@playwright/test';
import {
  setCharacterIndustryJobs,
  resetMockESI,
  type IndustryJob,
} from '../helpers/mock-esi';

// Alice Alpha's character ID
const ALICE_ALPHA_ID = 2001001;

test.describe('Industry', () => {
  test.afterAll(async () => {
    // Reset mock ESI state if any test mutated it
    await resetMockESI();
  });

  test('navigate to industry page shows heading and tabs', async ({ page }) => {
    await page.goto('/industry');

    // Page heading (use getByRole to avoid matching "No active industry jobs" cell)
    await expect(page.getByRole('heading', { name: 'Industry Jobs' })).toBeVisible({ timeout: 10000 });

    // All three tabs should be present
    await expect(page.getByRole('tab', { name: /Active Jobs/i })).toBeVisible();
    await expect(page.getByRole('tab', { name: /Queue/i })).toBeVisible();
    await expect(page.getByRole('tab', { name: 'Add Job' })).toBeVisible();
  });

  test('active jobs tab shows synced ESI job for Rifter', async ({ page }) => {
    // Clear localStorage so the tab state is fresh
    await page.goto('/industry');
    await page.evaluate(() => localStorage.clear());
    await page.goto('/industry');

    // Active Jobs tab is selected by default (tab index 0)
    await expect(page.getByRole('tab', { name: /Active Jobs/i })).toBeVisible({ timeout: 10000 });

    // The background runner syncs Alice Alpha's active Rifter manufacturing job from mock ESI.
    // This runner fires at startup, so the job should appear within 30s.
    // Wait for "Rifter" product name in the jobs table.
    // Use toPass polling: reload until the product name cell appears (runner fires every 10s).
    await expect(async () => {
      await page.reload();
      await expect(page.getByText('Rifter', { exact: true })).toBeVisible({ timeout: 5000 });
    }).toPass({ timeout: 60000 });

    // The job status chip should show "active" — use exact:true to avoid matching the "Active Jobs (1)" tab label
    await expect(page.getByText('active', { exact: true })).toBeVisible();
  });

  test('queue tab shows empty state', async ({ page }) => {
    await page.goto('/industry');
    await page.evaluate(() => localStorage.clear());
    await page.goto('/industry');

    // Click the Queue tab
    await page.getByRole('tab', { name: /Queue/i }).click();

    // No planned jobs exist yet — the table should show empty state
    await expect(page.getByText('No jobs in queue')).toBeVisible({ timeout: 10000 });
  });

  test('add job tab shows blueprint search input', async ({ page }) => {
    await page.goto('/industry');
    await page.evaluate(() => localStorage.clear());
    await page.goto('/industry');

    // Click the Add Job tab
    await page.getByRole('tab', { name: 'Add Job' }).click();

    // Blueprint search autocomplete should be visible
    await expect(page.getByLabel('Search Blueprint')).toBeVisible({ timeout: 5000 });

    // Activity selector should default to Manufacturing
    await expect(page.getByText('Manufacturing')).toBeVisible();

    // Runs field should be visible
    await expect(page.getByLabel('Runs')).toBeVisible();
  });

  test('blueprint search autocomplete returns Rifter result', async ({ page }) => {
    await page.goto('/industry');
    await page.evaluate(() => localStorage.clear());
    await page.goto('/industry');

    // Navigate to Add Job tab
    await page.getByRole('tab', { name: 'Add Job' }).click();

    // Wait for the search input to appear
    const searchInput = page.getByLabel('Search Blueprint');
    await expect(searchInput).toBeVisible({ timeout: 5000 });

    // Type "Rifter" — the search fires after a 300ms debounce and requires at least 2 chars
    await searchInput.fill('Ri');

    // Wait a moment for debounce then continue typing
    await searchInput.fill('Rifter');

    // The autocomplete dropdown should show Rifter as a result
    // Multiple Rifter variants may appear — pick the first match
    await expect(page.getByRole('option', { name: /Rifter/i }).first()).toBeVisible({ timeout: 10000 });
  });

  test('mock ESI: inject second active job appears after re-sync', async ({ page }) => {
    // Inject a second active job alongside the original Rifter job.
    // The controller only returns status IN ('active', 'paused', 'ready').
    const originalJob: IndustryJob = {
      job_id: 500001,
      installer_id: ALICE_ALPHA_ID,
      facility_id: 60003760,
      station_id: 60003760,
      activity_id: 1,
      blueprint_id: 9876,
      blueprint_type_id: 787,
      blueprint_location_id: 60003760,
      output_location_id: 60003760,
      runs: 10,
      cost: 1500000,
      product_type_id: 587,
      status: 'active',
      duration: 3600,
      start_date: '2026-02-22T00:00:00Z',
      end_date: '2026-02-22T01:00:00Z',
    };

    const secondJob: IndustryJob = {
      job_id: 500002,
      installer_id: ALICE_ALPHA_ID,
      facility_id: 60003760,
      station_id: 60003760,
      activity_id: 1,
      blueprint_id: 9877,
      blueprint_type_id: 787,
      blueprint_location_id: 60003760,
      output_location_id: 60003760,
      runs: 5,
      cost: 750000,
      product_type_id: 587,
      status: 'active',
      duration: 7200,
      start_date: '2026-02-22T00:00:00Z',
      end_date: '2026-02-22T02:00:00Z',
    };

    // Update mock ESI so next poll returns both jobs
    await setCharacterIndustryJobs(ALICE_ALPHA_ID, [originalJob, secondJob]);

    // Wait for the background runner to poll mock ESI (fires every 10s in E2E)
    // and pick up the new job. Reload the page to fetch fresh data.
    await page.goto('/industry');
    await page.evaluate(() => localStorage.clear());
    await page.goto('/industry');

    // Wait for Active Jobs tab
    await expect(page.getByRole('tab', { name: /Active Jobs/i })).toBeVisible({ timeout: 10000 });

    // Both jobs should show "Rifter" — we expect at least 2 rows with Rifter product
    // Use toPass polling: reload until 2 Rifter product name cells appear (runner fires every 10s).
    await expect(async () => {
      await page.reload();
      // Both jobs have product_type_id 587 (Rifter) — expect at least 2 text matches
      const rifterCells = page.getByText('Rifter', { exact: true });
      await expect(rifterCells.first()).toBeVisible({ timeout: 3000 });
      await expect(rifterCells.nth(1)).toBeVisible({ timeout: 3000 });
    }).toPass({ timeout: 35000 });

    // Verify the second job has runs=5 — formatNumber(5) = "5"
    await expect(page.getByText('5', { exact: true }).first()).toBeVisible();
  });
});
