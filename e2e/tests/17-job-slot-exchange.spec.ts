import { test, expect } from '../fixtures/auth';

test.describe('Job Slot Rental Exchange', () => {
  test('navigate to job slots page shows all tabs', async ({ alicePage }) => {
    await alicePage.goto('/job-slots');

    await expect(alicePage.getByRole('tab', { name: 'Slot Inventory' })).toBeVisible({ timeout: 10000 });
    await expect(alicePage.getByRole('tab', { name: 'My Listings' })).toBeVisible();
    await expect(alicePage.getByRole('tab', { name: 'Browse Listings' })).toBeVisible();
    await expect(alicePage.getByRole('tab', { name: 'Interest Requests' })).toBeVisible();
  });

  test('slot inventory tab shows character slot data', async ({ alicePage }) => {
    await alicePage.goto('/job-slots');
    await alicePage.evaluate(() => localStorage.clear());
    await alicePage.goto('/job-slots');

    // Slot Inventory tab should be active by default
    await expect(alicePage.getByRole('tab', { name: 'Slot Inventory' })).toBeVisible({ timeout: 10000 });

    // Wait for slot data to load - Alice Alpha should appear with calculated slot capacity
    // based on her Industry 5 and Advanced Industry 5 skills from mock ESI
    await expect(alicePage.getByText('Alice Alpha').first()).toBeVisible({ timeout: 15000 });

    // Should show activity types where Alice has skills (Manufacturing, Reactions)
    await expect(alicePage.getByText(/Manufacturing|manufacturing/i).first()).toBeVisible({ timeout: 5000 });
  });

  test('Alice creates a manufacturing slot rental listing', async ({ alicePage }) => {
    await alicePage.goto('/job-slots');
    await alicePage.evaluate(() => localStorage.clear());
    await alicePage.goto('/job-slots');

    // Navigate to My Listings tab
    await alicePage.getByRole('tab', { name: 'My Listings' }).click();
    await expect(alicePage.getByRole('tab', { name: 'My Listings' })).toHaveAttribute('aria-selected', 'true');

    // Click Create Listing button
    await alicePage.getByRole('button', { name: /Create Listing/i }).click();

    // Wait for dialog to appear
    const dialog = alicePage.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // Select character (Alice Alpha)
    const characterControl = dialog.locator('.MuiFormControl-root').filter({
      has: alicePage.locator('label').filter({ hasText: 'Character' }),
    });
    await characterControl.getByRole('combobox').click();
    await alicePage.getByRole('option', { name: /Alice Alpha/i }).click();

    // Select activity type (Manufacturing)
    const activityControl = dialog.locator('.MuiFormControl-root').filter({
      has: alicePage.locator('label').filter({ hasText: 'Activity Type' }),
    });
    await activityControl.getByRole('combobox').click();
    await alicePage.getByRole('option', { name: /Manufacturing/i }).click();

    // Enter slots to list
    const slotsInput = dialog.getByLabel(/Slots to List/i);
    await slotsInput.fill('3');

    // Enter price amount
    const priceInput = dialog.getByLabel(/Price Amount/i);
    await priceInput.fill('100000');

    // Select pricing unit (per_slot_day)
    const pricingControl = dialog.locator('.MuiFormControl-root').filter({
      has: alicePage.locator('label').filter({ hasText: 'Pricing Unit' }),
    });
    await pricingControl.getByRole('combobox').click();
    await alicePage.getByRole('option', { name: /per.*slot.*day/i }).click();

    // Enter notes
    const notesInput = dialog.getByLabel(/Notes/i);
    await notesInput.fill('Manufacturing slots available in Jita');

    // Save listing
    await dialog.getByRole('button', { name: /Create|Save/i }).click();

    // Verify dialog closes
    await expect(dialog).not.toBeVisible({ timeout: 5000 });

    // Verify listing appears in My Listings table
    await expect(alicePage.getByText('Alice Alpha').first()).toBeVisible({ timeout: 10000 });
    await expect(alicePage.getByText(/Manufacturing/i).first()).toBeVisible();
    await expect(alicePage.getByText(/100\.00K/).first()).toBeVisible();
  });

  test('Bob grants Alice job slot browse permission', async ({ bobPage }) => {
    await bobPage.goto('/contacts');

    // Go to My Contacts tab
    await bobPage.getByRole('tab', { name: /My Contacts/i }).click();
    await expect(bobPage.getByText('Alice')).toBeVisible({ timeout: 10000 });

    // Click Manage Permissions on Alice's row
    const aliceRow = bobPage.getByRole('row').filter({ hasText: 'Alice' });
    await aliceRow.getByTitle('Manage Permissions').click();

    // Wait for permissions dialog
    const dialog = bobPage.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // Find the "Browse Job Slot Listings" switch in the "Permissions I Grant" section
    // The heading contains the other user's name, so use a partial text match
    const grantSection = dialog.getByText(/Permissions I Grant/i).locator('..');
    // MUI FormControlLabel wraps the Switch + label text in a <label> element
    // Find the label containing our text, then click its input checkbox with force
    const jobSlotLabel = grantSection.locator('label').filter({ hasText: /Browse Job Slot Listings/i });
    await expect(jobSlotLabel).toBeVisible({ timeout: 5000 });
    await jobSlotLabel.locator('input[type="checkbox"]').click({ force: true });

    // Wait for the API call to complete
    await bobPage.waitForTimeout(1000);
    // Verify the switch is now checked
    await expect(jobSlotLabel.locator('input[type="checkbox"]')).toBeChecked({ timeout: 5000 });

    // Close dialog
    await bobPage.getByRole('button', { name: /Close/i }).click();
  });

  test('Bob creates a manufacturing slot listing', async ({ bobPage }) => {
    await bobPage.goto('/job-slots');
    await bobPage.evaluate(() => localStorage.clear());
    await bobPage.goto('/job-slots');

    // Navigate to My Listings tab
    await bobPage.getByRole('tab', { name: 'My Listings' }).click();

    // Click Create Listing button
    await bobPage.getByRole('button', { name: /Create Listing/i }).click();

    const dialog = bobPage.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // Select character (Bob Bravo)
    const characterControl = dialog.locator('.MuiFormControl-root').filter({
      has: bobPage.locator('label').filter({ hasText: 'Character' }),
    });
    await characterControl.getByRole('combobox').click();
    await bobPage.getByRole('option', { name: /Bob Bravo/i }).click();

    // Select activity type (Manufacturing)
    const activityControl = dialog.locator('.MuiFormControl-root').filter({
      has: bobPage.locator('label').filter({ hasText: 'Activity Type' }),
    });
    await activityControl.getByRole('combobox').click();
    await bobPage.getByRole('option', { name: /Manufacturing/i }).click();

    // Enter slots to list (Bob has 4 total slots: 1 + Mass Production 3)
    await dialog.getByLabel(/Slots to List/i).fill('3');

    // Enter price amount
    await dialog.getByLabel(/Price Amount/i).fill('75000');

    // Select pricing unit
    const pricingControl = dialog.locator('.MuiFormControl-root').filter({
      has: bobPage.locator('label').filter({ hasText: 'Pricing Unit' }),
    });
    await pricingControl.getByRole('combobox').click();
    await bobPage.getByRole('option', { name: /per.*slot.*day/i }).click();

    // Enter notes
    await dialog.getByLabel(/Notes/i).fill('Manufacturing slots in Jita, fast turnaround');

    // Save listing
    await dialog.getByRole('button', { name: /Create|Save/i }).click();

    // Verify listing appears
    await expect(bobPage.getByText('Bob Bravo').first()).toBeVisible({ timeout: 10000 });
    await expect(bobPage.getByText(/Manufacturing/i).first()).toBeVisible();
    await expect(bobPage.getByText(/75\.00K/).first()).toBeVisible();
  });

  test('Alice can browse Bob listings with permission', async ({ alicePage }) => {
    await alicePage.goto('/job-slots');
    await alicePage.evaluate(() => localStorage.clear());
    await alicePage.goto('/job-slots');

    // Navigate to Browse Listings tab
    await alicePage.getByRole('tab', { name: 'Browse Listings' }).click();
    await expect(alicePage.getByRole('tab', { name: 'Browse Listings' })).toHaveAttribute('aria-selected', 'true');

    // Should see Bob's listing (permission-filtered)
    await expect(alicePage.getByText('Bob')).toBeVisible({ timeout: 10000 });
    await expect(alicePage.getByText('Bob Bravo')).toBeVisible();
    await expect(alicePage.getByText(/Manufacturing/i).first()).toBeVisible();
    await expect(alicePage.getByText(/75\.00K/).first()).toBeVisible();
  });

  test('Alice expresses interest in Bob listing', async ({ alicePage }) => {
    await alicePage.goto('/job-slots');
    await alicePage.evaluate(() => localStorage.clear());
    await alicePage.goto('/job-slots');

    // Navigate to Browse Listings tab
    await alicePage.getByRole('tab', { name: 'Browse Listings' }).click();
    await expect(alicePage.getByText('Bob')).toBeVisible({ timeout: 10000 });

    // Click Express Interest button on Bob's listing
    const bobRow = alicePage.getByRole('row').filter({ hasText: 'Bob Bravo' });
    await bobRow.getByRole('button', { name: /Express Interest|Interest/i }).click();

    // Fill in interest request dialog
    const dialog = alicePage.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // Enter slots requested
    await dialog.getByLabel(/Slots Requested/i).fill('2');

    // Enter duration (optional, but let's test it)
    await dialog.getByLabel(/Duration.*Days/i).fill('30');

    // Enter message
    await dialog.getByLabel(/Message/i).fill('Looking to use 2 slots for Rifter manufacturing, 30 day contract');

    // Submit interest request
    await dialog.getByRole('button', { name: /Submit|Send/i }).click();

    // Verify dialog closes
    await expect(dialog).not.toBeVisible({ timeout: 5000 });
  });

  test('Alice sees her sent interest request', async ({ alicePage }) => {
    await alicePage.goto('/job-slots');
    await alicePage.evaluate(() => localStorage.clear());
    await alicePage.goto('/job-slots');

    // Navigate to Interest Requests tab
    await alicePage.getByRole('tab', { name: 'Interest Requests' }).click();

    // Switch to Sent subtab (wait for it to appear)
    const sentTab = alicePage.getByRole('tab', { name: /Sent/i });
    await expect(sentTab).toBeVisible({ timeout: 10000 });
    await sentTab.click();

    // Verify Alice's interest request to Bob appears
    await expect(alicePage.getByText('Bob Bravo')).toBeVisible({ timeout: 15000 });
    await expect(alicePage.getByText(/30.*day|Duration.*30/i).first()).toBeVisible({ timeout: 5000 });
    await expect(alicePage.getByText(/pending/i).first()).toBeVisible();
  });

  test('Bob sees received interest request from Alice', async ({ bobPage }) => {
    await bobPage.goto('/job-slots');
    await bobPage.evaluate(() => localStorage.clear());
    await bobPage.goto('/job-slots');

    // Navigate to Interest Requests tab
    await bobPage.getByRole('tab', { name: 'Interest Requests' }).click();

    // Switch to Received subtab (wait for it to appear)
    const receivedTab = bobPage.getByRole('tab', { name: /Received/i });
    await expect(receivedTab).toBeVisible({ timeout: 10000 });
    await receivedTab.click();

    // Verify Bob sees Alice's interest request
    await expect(bobPage.getByText('Alice Stargazer').first()).toBeVisible({ timeout: 15000 });
    await expect(bobPage.getByText(/2.*slot|Slots.*2/i).first()).toBeVisible({ timeout: 5000 });
    await expect(bobPage.getByText(/30.*day|Duration.*30/i).first()).toBeVisible();
    await expect(bobPage.getByText(/Rifter manufacturing/i)).toBeVisible();
  });

  test('Bob accepts Alice interest request', async ({ bobPage }) => {
    await bobPage.goto('/job-slots');
    await bobPage.evaluate(() => localStorage.clear());
    await bobPage.goto('/job-slots');

    // Navigate to Interest Requests tab, Received subtab
    await bobPage.getByRole('tab', { name: 'Interest Requests' }).click();
    const receivedTab = bobPage.getByRole('tab', { name: /Received/i });
    await expect(receivedTab).toBeVisible({ timeout: 10000 });
    await receivedTab.click();

    await expect(bobPage.getByText('Alice Stargazer').first()).toBeVisible({ timeout: 15000 });

    // Click Accept button on Alice's request
    const aliceRow = bobPage.getByRole('row').filter({ hasText: 'Alice' });
    await aliceRow.getByRole('button', { name: /Accept/i }).click();

    // Verify status changes to accepted (use exact match to avoid snackbar "Interest accepted")
    await expect(aliceRow.getByText('accepted')).toBeVisible({ timeout: 5000 });
  });

  test('Alice creates another interest request for decline test', async ({ alicePage }) => {
    await alicePage.goto('/job-slots');
    await alicePage.evaluate(() => localStorage.clear());
    await alicePage.goto('/job-slots');

    // Navigate to Browse Listings tab
    await alicePage.getByRole('tab', { name: 'Browse Listings' }).click();
    await expect(alicePage.getByText('Bob Bravo')).toBeVisible({ timeout: 10000 });

    // Click Express Interest button
    const bobRow = alicePage.getByRole('row').filter({ hasText: 'Bob Bravo' });
    await bobRow.getByRole('button', { name: /Express Interest|Interest/i }).click();

    const dialog = alicePage.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // Enter minimal data for decline test
    await dialog.getByLabel(/Slots Requested/i).fill('1');
    await dialog.getByLabel(/Message/i).fill('Test request for decline');

    // Submit
    await dialog.getByRole('button', { name: /Submit|Send/i }).click();
    await expect(dialog).not.toBeVisible({ timeout: 5000 });
  });

  test('Bob declines second interest request', async ({ bobPage }) => {
    await bobPage.goto('/job-slots');
    await bobPage.evaluate(() => localStorage.clear());
    await bobPage.goto('/job-slots');

    // Navigate to Interest Requests tab, Received subtab
    await bobPage.getByRole('tab', { name: 'Interest Requests' }).click();
    const receivedTab = bobPage.getByRole('tab', { name: /Received/i });
    await expect(receivedTab).toBeVisible({ timeout: 10000 });
    await receivedTab.click();

    // Wait for requests to load
    await expect(bobPage.getByText('Alice Stargazer').first()).toBeVisible({ timeout: 15000 });

    // Find the row with "Test request for decline" message
    const testRow = bobPage.getByRole('row').filter({ hasText: 'Test request for decline' });
    await expect(testRow).toBeVisible({ timeout: 5000 });

    // Register dialog handler before clicking Decline
    bobPage.on('dialog', dialog => dialog.accept());

    // Click Decline button
    await testRow.getByRole('button', { name: /Decline/i }).click();

    // Verify status changes to declined
    await expect(testRow.getByText(/declined/i)).toBeVisible({ timeout: 5000 });
  });

  test('Alice withdraws her own interest request', async ({ alicePage }) => {
    // First, create a new interest request that Alice will withdraw
    await alicePage.goto('/job-slots');
    await alicePage.evaluate(() => localStorage.clear());
    await alicePage.goto('/job-slots');

    // Navigate to Browse Listings tab
    await alicePage.getByRole('tab', { name: 'Browse Listings' }).click();
    await expect(alicePage.getByText('Bob Bravo')).toBeVisible({ timeout: 10000 });

    // Create interest request
    const bobRow = alicePage.getByRole('row').filter({ hasText: 'Bob Bravo' });
    await bobRow.getByRole('button', { name: /Express Interest|Interest/i }).click();

    const dialog = alicePage.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });
    await dialog.getByLabel(/Slots Requested/i).fill('3');
    await dialog.getByLabel(/Message/i).fill('Request for withdrawal test');
    await dialog.getByRole('button', { name: /Submit|Send/i }).click();
    await expect(dialog).not.toBeVisible({ timeout: 5000 });

    // Now navigate to Sent requests to withdraw it
    await alicePage.getByRole('tab', { name: 'Interest Requests' }).click();
    const sentTab = alicePage.getByRole('tab', { name: /Sent/i });
    await expect(sentTab).toBeVisible({ timeout: 10000 });
    await sentTab.click();

    // Find the request for withdrawal
    const withdrawRow = alicePage.getByRole('row').filter({ hasText: 'Request for withdrawal test' });
    await expect(withdrawRow).toBeVisible({ timeout: 15000 });

    // Register dialog handler before clicking Withdraw
    alicePage.on('dialog', dialog => dialog.accept());

    // Click Withdraw button
    await withdrawRow.getByRole('button', { name: /Withdraw/i }).click();

    // Verify status changes to withdrawn
    await expect(withdrawRow.getByText(/withdrawn/i)).toBeVisible({ timeout: 5000 });
  });

  test('Alice updates her listing price', async ({ alicePage }) => {
    await alicePage.goto('/job-slots');
    await alicePage.evaluate(() => localStorage.clear());
    await alicePage.goto('/job-slots');

    // Navigate to My Listings tab
    await alicePage.getByRole('tab', { name: 'My Listings' }).click();
    await expect(alicePage.getByText('Alice Alpha').first()).toBeVisible({ timeout: 10000 });

    // Click Edit button on Alice's listing
    const aliceRow = alicePage.getByRole('row').filter({ hasText: 'Alice Alpha' });
    await aliceRow.getByRole('button', { name: /Edit/i }).click();

    // Edit dialog should appear
    const dialog = alicePage.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // Change price amount
    const priceInput = dialog.getByLabel(/Price Amount/i);
    await priceInput.clear();
    await priceInput.fill('125000');

    // Save changes
    await dialog.getByRole('button', { name: /Save|Update/i }).click();
    await expect(dialog).not.toBeVisible({ timeout: 5000 });

    // Verify updated price displays
    await expect(alicePage.getByText(/125\.00K/)).toBeVisible({ timeout: 10000 });
  });

  test('Alice deletes a listing', async ({ alicePage }) => {
    await alicePage.goto('/job-slots');
    await alicePage.evaluate(() => localStorage.clear());
    await alicePage.goto('/job-slots');

    // Navigate to My Listings tab
    await alicePage.getByRole('tab', { name: 'My Listings' }).click();
    await expect(alicePage.getByText('Alice Alpha').first()).toBeVisible({ timeout: 10000 });

    // Check if a Reaction listing already exists (from a previous test run or retry)
    const reactionExists = await alicePage.getByRole('row').filter({ hasText: /Reaction/i }).count() > 0;

    if (!reactionExists) {
      // Create a new listing for deletion test
      await alicePage.getByRole('button', { name: /Create Listing/i }).click();
      const createDialog = alicePage.getByRole('dialog');
      await expect(createDialog).toBeVisible({ timeout: 5000 });

      const characterControl = createDialog.locator('.MuiFormControl-root').filter({
        has: alicePage.locator('label').filter({ hasText: 'Character' }),
      });
      await characterControl.getByRole('combobox').click();
      await alicePage.getByRole('option', { name: /Alice Alpha/i }).click();

      const activityControl = createDialog.locator('.MuiFormControl-root').filter({
        has: alicePage.locator('label').filter({ hasText: 'Activity Type' }),
      });
      await activityControl.getByRole('combobox').click();
      await alicePage.getByRole('option', { name: /Reaction/i }).click();

      await createDialog.getByLabel(/Slots to List/i).fill('2');
      await createDialog.getByLabel(/Price Amount/i).fill('50000');

      const pricingControl = createDialog.locator('.MuiFormControl-root').filter({
        has: alicePage.locator('label').filter({ hasText: 'Pricing Unit' }),
      });
      await pricingControl.getByRole('combobox').click();
      await alicePage.getByRole('option', { name: /per.*job/i }).click();

      await createDialog.getByLabel(/Notes/i).fill('Listing to be deleted');
      await createDialog.getByRole('button', { name: /Create|Save/i }).click();
      await expect(createDialog).not.toBeVisible({ timeout: 5000 });
    }

    // Wait for Reaction listing to appear
    await expect(alicePage.getByText(/Reaction/i).first()).toBeVisible({ timeout: 10000 });

    // Find and delete the reaction listing
    const reactionRow = alicePage.getByRole('row').filter({ hasText: /Reaction/i });
    await expect(reactionRow).toBeVisible({ timeout: 5000 });

    // Register dialog handler before clicking Delete
    alicePage.on('dialog', dialog => dialog.accept());

    // Click Delete button
    await reactionRow.getByRole('button', { name: /Delete/i }).click();

    // Verify listing is removed
    await expect(reactionRow).not.toBeVisible({ timeout: 10000 });
  });

  test('can switch between all job slot tabs', async ({ alicePage }) => {
    await alicePage.goto('/job-slots');
    await alicePage.evaluate(() => localStorage.clear());
    await alicePage.goto('/job-slots');

    const tabs = ['Slot Inventory', 'My Listings', 'Browse Listings', 'Interest Requests'];

    for (const tabName of tabs) {
      await alicePage.getByRole('tab', { name: tabName }).click();
      await expect(alicePage.getByRole('tab', { name: tabName })).toHaveAttribute('aria-selected', 'true');
    }
  });
});
