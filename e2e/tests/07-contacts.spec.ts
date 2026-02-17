import { test, expect } from '../fixtures/auth';

test.describe('Contacts Workflow', () => {
  test('Alice sends contact request to Bob', async ({ alicePage }) => {
    await alicePage.goto('/contacts');

    // Click "Add Contact" button
    await expect(alicePage.getByRole('button', { name: /Add Contact/i })).toBeVisible({ timeout: 10000 });
    await alicePage.getByRole('button', { name: /Add Contact/i }).click();

    // Fill in Bob's character name
    await expect(alicePage.getByLabel(/Character Name/i)).toBeVisible();
    await alicePage.getByLabel(/Character Name/i).fill('Bob Bravo');

    // Click Send Request
    await alicePage.getByRole('button', { name: /Send Request/i }).click();

    // Switch to Sent Requests tab
    await alicePage.getByRole('tab', { name: /Sent/i }).click();
    await expect(alicePage.getByText('Bob')).toBeVisible({ timeout: 5000 });
  });

  test('Bob sees pending request from Alice', async ({ bobPage }) => {
    await bobPage.goto('/contacts');

    // Switch to Pending Requests tab
    await bobPage.getByRole('tab', { name: /Pending/i }).click();

    // Should see Alice's request
    await expect(bobPage.getByText('Alice')).toBeVisible({ timeout: 5000 });
  });

  test('Bob accepts contact request from Alice', async ({ bobPage }) => {
    await bobPage.goto('/contacts');

    // Switch to Pending Requests tab
    await bobPage.getByRole('tab', { name: /Pending/i }).click();
    await expect(bobPage.getByText('Alice')).toBeVisible({ timeout: 10000 });

    // Click accept button (first button in the row — CheckIcon)
    const aliceRow = bobPage.getByRole('row').filter({ hasText: 'Alice' });
    await aliceRow.getByRole('button').first().click();

    // Verify on My Contacts tab - Alice should now appear as connected
    await bobPage.getByRole('tab', { name: /My Contacts/i }).click();
    await expect(bobPage.getByText('Alice')).toBeVisible({ timeout: 10000 });
  });

  test('Bob grants Alice browse permission for marketplace', async ({ bobPage }) => {
    await bobPage.goto('/contacts');

    // Go to My Contacts tab
    await bobPage.getByRole('tab', { name: /My Contacts/i }).click();
    await expect(bobPage.getByText('Alice')).toBeVisible({ timeout: 10000 });

    // Click Manage Permissions (settings icon) on Alice's row
    const aliceRow = bobPage.getByRole('row').filter({ hasText: 'Alice' });
    await aliceRow.getByTitle('Manage Permissions').click();

    // Wait for permissions dialog to load
    await expect(bobPage.getByText(/Manage Permissions/)).toBeVisible({ timeout: 5000 });

    // Toggle "Browse For-Sale Items" switch under "Permissions I Grant to Alice"
    // The first switch in the dialog is the editable one (grants to Alice)
    const dialog = bobPage.getByRole('dialog');
    await dialog.getByRole('switch').first().click();
    await expect(dialog.getByRole('switch').first()).toBeChecked({ timeout: 5000 });

    // Close dialog
    await bobPage.getByRole('button', { name: /Close/i }).click();
  });

  test('Alice sees Bob as accepted contact', async ({ alicePage }) => {
    await alicePage.goto('/contacts');

    // My Contacts tab should show Bob
    await expect(alicePage.getByText('Bob')).toBeVisible({ timeout: 10000 });
  });
});
