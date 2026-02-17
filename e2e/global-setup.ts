import { chromium, FullConfig } from '@playwright/test';

const BASE_URL = process.env.BASE_URL || 'http://localhost:3000';
// In CI Docker, BASE_URL is http://frontend:3000 — derive backend from that network
const BACKEND_URL = process.env.CI ? 'http://backend:80' : 'http://localhost:8080';
const BACKEND_KEY = process.env.BACKEND_KEY || 'e2e-test-key';

async function waitForDatabase(timeoutMs = 60_000) {
  const start = Date.now();
  const healthUrl = `${BACKEND_URL}/v1/characters/`;
  while (Date.now() - start < timeoutMs) {
    try {
      const res = await fetch(healthUrl, {
        headers: { 'BACKEND-KEY': BACKEND_KEY, 'USER-ID': '1' },
      });
      if (res.ok) {
        console.log('Database and backend are ready');
        return;
      }
    } catch {
      // not ready yet
    }
    await new Promise((r) => setTimeout(r, 2_000));
  }
  throw new Error(`Database/backend not ready after ${timeoutMs}ms`);
}

async function globalSetup(config: FullConfig) {
  await waitForDatabase();

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
