import { chromium, FullConfig } from '@playwright/test';

const BASE_URL = process.env.BASE_URL || 'http://localhost:3000';

async function globalSetup(config: FullConfig) {
  const browser = await chromium.launch();
  const context = await browser.newContext();
  const page = await context.newPage();

  // Get CSRF token
  const csrfResponse = await page.request.get(`${BASE_URL}/api/auth/csrf`);
  const { csrfToken } = await csrfResponse.json();

  // Sign in as Alice Stargazer (primary test user)
  // maxRedirects: 0 prevents following NextAuth's redirect to NEXTAUTH_URL,
  // which may differ from BASE_URL when running in Docker
  await page.request.post(`${BASE_URL}/api/auth/callback/credentials`, {
    form: {
      userId: '1001',
      userName: 'Alice Stargazer',
      csrfToken,
    },
    maxRedirects: 0,
  });

  // Load the app to verify sign-in and populate cookies
  await page.goto(`${BASE_URL}/`);

  // Save authenticated state
  await context.storageState({ path: './auth-state.json' });

  await browser.close();
}

export default globalSetup;
