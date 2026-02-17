import { test as base, Page, BrowserContext } from '@playwright/test';

const BASE_URL = process.env.BASE_URL || 'http://localhost:3000';

async function loginAs(context: BrowserContext, userId: string, userName: string): Promise<Page> {
  const page = await context.newPage();

  // Get CSRF token
  const csrfResponse = await page.request.get(`${BASE_URL}/api/auth/csrf`);
  const { csrfToken } = await csrfResponse.json();

  // Sign in via CredentialsProvider
  // maxRedirects: 0 prevents following NextAuth's redirect to NEXTAUTH_URL,
  // which may differ from BASE_URL when running in Docker
  await page.request.post(`${BASE_URL}/api/auth/callback/credentials`, {
    form: {
      userId,
      userName,
      csrfToken,
    },
    maxRedirects: 0,
  });

  return page;
}

type AuthFixtures = {
  alicePage: Page;
  bobPage: Page;
  charliePage: Page;
  dianaPage: Page;
};

export const test = base.extend<AuthFixtures>({
  // Alice Stargazer (user 1001) — uses default storageState
  alicePage: async ({ page }, use) => {
    await use(page);
  },

  // Bob Miner (user 1002) — fresh context with separate login
  bobPage: async ({ browser }, use) => {
    const context = await browser.newContext();
    const page = await loginAs(context, '1002', 'Bob Miner');
    await use(page);
    await context.close();
  },

  // Charlie Trader (user 1003)
  charliePage: async ({ browser }, use) => {
    const context = await browser.newContext();
    const page = await loginAs(context, '1003', 'Charlie Trader');
    await use(page);
    await context.close();
  },

  // Diana Scout (user 1004)
  dianaPage: async ({ browser }, use) => {
    const context = await browser.newContext();
    const page = await loginAs(context, '1004', 'Diana Scout');
    await use(page);
    await context.close();
  },
});

export { expect } from '@playwright/test';
