import { test } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

const PAGES: Record<string, string> = {
  landing: '/',
  characters: '/characters',
  corporations: '/corporations',
  assets: '/assets',
  inventory: '/inventory',
  stockpiles: '/stockpiles',
  marketplace: '/marketplace',
  reactions: '/reactions',
  industry: '/industry',
  'production-plans': '/production-plans',
  'plan-runs': '/plan-runs',
  'job-slots': '/job-slots',
  pi: '/pi',
  transport: '/transport',
  hauling: '/hauling',
  contacts: '/contacts',
  stations: '/stations',
  settings: '/settings',
};

const SCREENSHOT_DIR = path.resolve(__dirname, '..', 'screenshots');

test.describe('Visual Review', () => {
  test.beforeAll(() => {
    fs.mkdirSync(SCREENSHOT_DIR, { recursive: true });
  });

  const pagesToCapture = process.env.PAGES
    ? process.env.PAGES.split(',').map((p) => p.trim())
    : Object.keys(PAGES);

  for (const pageName of pagesToCapture) {
    const route = PAGES[pageName];
    if (!route) continue;

    test(`capture ${pageName}`, async ({ page }) => {
      await page.goto(route);
      await page.waitForLoadState('networkidle');

      // Give dynamic content a moment to render
      await page.waitForTimeout(500);

      await page.screenshot({
        path: path.join(SCREENSHOT_DIR, `${pageName}.png`),
        fullPage: true,
      });
    });
  }
});
