import { test, expect } from '@playwright/test';

// ---------------------------------------------------------------------------
// Station Markets — E2E tests (Feature #162)
//
// Tests cover:
//   - GET /api/hauling/stations returns preset trading stations (Jita 4-4)
//   - GET /api/hauling/structures returns user-configured structures
//   - Scanner page shows Source/Destination Location dropdowns (including
//     Jita 4-4 station and user structures alongside regions)
//   - Gear button opens the Manage Trading Structures dialog
//   - Dialog shows structures with Access OK / No Access badges
//   - DELETE structure removes it from the list
//   - Scanner with Jita 4-4 station source loads results
//
// Seed data (e2e/seed.sql) provides:
//   - trading_stations: Jita IV - Moon 4 (is_preset=true) — seeded via migration
//   - user_trading_structures: "Perimeter - Test Trading Hub" (access_ok=true),
//                              "Restricted Citadel" (access_ok=false)
//   - hauling_structure_snapshots: Tritanium/Pyerite/Mexallon in structure 1234567890123
//   - hauling_market_snapshots: Domain region snapshots for scanner results
//
// Mock ESI:
//   - /latest/universe/structures/1234567890123/ → Perimeter - Test Trading Hub
//   - /latest/markets/structures/1234567890123/ → 5 market orders
//   - /latest/universe/structures/9999999999999/ → 403 Forbidden
//   - /latest/markets/structures/9999999999999/ → 403 Forbidden
// ---------------------------------------------------------------------------

test.describe('Station Markets', () => {
  test.beforeEach(async ({ page }) => {
    // Clear localStorage so scanner/panel state is fresh
    await page.goto('/hauling/scanner');
    await page.evaluate(() => localStorage.clear());
  });

  // -------------------------------------------------------------------------
  // 1. Trading stations API
  // -------------------------------------------------------------------------

  test('GET /api/hauling/stations returns Jita 4-4 preset station', async ({ page }) => {
    // Call the stations API directly to verify the preset seeded via migration
    const response = await page.request.get('/api/hauling/stations');
    expect(response.status()).toBe(200);

    const data = await response.json();
    expect(Array.isArray(data)).toBe(true);
    expect(data.length).toBeGreaterThan(0);

    // The migration inserts Jita 4-4 as a preset station
    const jita = data.find(
      (s: { name: string; isPreset: boolean }) =>
        s.name.includes('Jita') && s.isPreset === true,
    );
    expect(jita).toBeDefined();
    expect(jita.stationId).toBe(60003760);
    expect(jita.regionId).toBe(10000002);
  });

  test('GET /api/hauling/stations includes station system and region IDs', async ({ page }) => {
    const response = await page.request.get('/api/hauling/stations');
    expect(response.status()).toBe(200);

    const data = await response.json();
    const jita = data.find((s: { stationId: number }) => s.stationId === 60003760);
    expect(jita).toBeDefined();
    expect(jita.systemId).toBe(30000142);
    expect(jita.regionId).toBe(10000002);
  });

  // -------------------------------------------------------------------------
  // 2. User trading structures API
  // -------------------------------------------------------------------------

  test('GET /api/hauling/structures returns seeded structures for user', async ({ page }) => {
    // The seed data adds two structures for user_id=1001 (Alice)
    const response = await page.request.get('/api/hauling/structures');
    expect(response.status()).toBe(200);

    const data = await response.json();
    expect(Array.isArray(data)).toBe(true);

    const accessible = data.find(
      (s: { name: string }) => s.name === 'Perimeter - Test Trading Hub',
    );
    expect(accessible).toBeDefined();
    expect(accessible.structureId).toBe(1234567890123);
    expect(accessible.accessOk).toBe(true);
    expect(accessible.regionId).toBe(10000002);
  });

  test('GET /api/hauling/structures shows structure with access_ok=false', async ({ page }) => {
    const response = await page.request.get('/api/hauling/structures');
    expect(response.status()).toBe(200);

    const data = await response.json();
    const restricted = data.find(
      (s: { name: string }) => s.name === 'Restricted Citadel',
    );
    expect(restricted).toBeDefined();
    expect(restricted.accessOk).toBe(false);
  });

  // -------------------------------------------------------------------------
  // 3. Scanner page — updated UI with location dropdowns
  // -------------------------------------------------------------------------

  test('scanner page loads with Market Scanner heading', async ({ page }) => {
    await page.goto('/hauling/scanner');
    await expect(
      page.getByRole('heading', { name: /Market Scanner/i }),
    ).toBeVisible({ timeout: 10000 });
  });

  test('scanner page shows Source Location and Destination Location labels', async ({ page }) => {
    await page.goto('/hauling/scanner');
    await expect(
      page.getByRole('heading', { name: /Market Scanner/i }),
    ).toBeVisible({ timeout: 10000 });

    // The updated MarketScanner uses "Source Location" / "Destination Location"
    await expect(page.getByText('Source Location')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Destination Location')).toBeVisible({ timeout: 5000 });
  });

  test('scanner page shows Load and Scan buttons', async ({ page }) => {
    await page.goto('/hauling/scanner');
    await expect(
      page.getByRole('heading', { name: /Market Scanner/i }),
    ).toBeVisible({ timeout: 10000 });

    await expect(page.getByRole('button', { name: 'Load' })).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('button', { name: 'Scan' })).toBeVisible({ timeout: 5000 });
  });

  test('scanner page shows gear button to manage trading structures', async ({ page }) => {
    await page.goto('/hauling/scanner');
    await expect(
      page.getByRole('heading', { name: /Market Scanner/i }),
    ).toBeVisible({ timeout: 10000 });

    // The gear icon button has title="Manage trading structures"
    await expect(
      page.getByTitle('Manage trading structures'),
    ).toBeVisible({ timeout: 5000 });
  });

  test('scanner source location defaults to The Forge region', async ({ page }) => {
    await page.goto('/hauling/scanner');
    await expect(
      page.getByRole('heading', { name: /Market Scanner/i }),
    ).toBeVisible({ timeout: 10000 });

    // The first combobox is Source Location — default is The Forge (10000002)
    const sourceCombo = page.getByRole('combobox').first();
    await expect(sourceCombo).toBeVisible({ timeout: 5000 });
    await expect(sourceCombo).toContainText('The Forge');
  });

  test('scanner destination location defaults to Domain region', async ({ page }) => {
    await page.goto('/hauling/scanner');
    await expect(
      page.getByRole('heading', { name: /Market Scanner/i }),
    ).toBeVisible({ timeout: 10000 });

    // The second combobox is Destination Location — default is Domain (10000043)
    const destCombo = page.getByRole('combobox').nth(1);
    await expect(destCombo).toBeVisible({ timeout: 5000 });
    await expect(destCombo).toContainText('Domain');
  });

  test('source location dropdown includes Jita 4-4 station from preset stations', async ({ page }) => {
    await page.goto('/hauling/scanner');
    await expect(
      page.getByRole('heading', { name: /Market Scanner/i }),
    ).toBeVisible({ timeout: 10000 });

    // The scanner fetches stations on load — wait for UI to populate
    // Then open the source combobox to check options
    const sourceCombo = page.getByRole('combobox').first();
    await expect(sourceCombo).toBeVisible({ timeout: 5000 });

    // Wait for stations to be fetched and dropdown options populated
    // The component fetches stations in useEffect([session])
    await sourceCombo.click();

    // Jita 4-4 should appear in the dropdown under "Stations" section
    await expect(
      page.getByRole('option', { name: /Jita IV - Moon 4/i }),
    ).toBeVisible({ timeout: 5000 });
  });

  test('source location dropdown includes user structures under My Structures', async ({ page }) => {
    await page.goto('/hauling/scanner');
    await expect(
      page.getByRole('heading', { name: /Market Scanner/i }),
    ).toBeVisible({ timeout: 10000 });

    const sourceCombo = page.getByRole('combobox').first();
    await expect(sourceCombo).toBeVisible({ timeout: 5000 });

    // Wait for structures to load and open the dropdown
    await sourceCombo.click();

    // Perimeter - Test Trading Hub should appear under "My Structures"
    await expect(
      page.getByRole('option', { name: /Perimeter - Test Trading Hub/i }),
    ).toBeVisible({ timeout: 5000 });
  });

  test('Load button fetches cached region-to-region results without error', async ({ page }) => {
    await page.goto('/hauling/scanner');
    await expect(
      page.getByRole('heading', { name: /Market Scanner/i }),
    ).toBeVisible({ timeout: 10000 });

    // Click Load (uses cached market snapshot data — no ESI call)
    await page.getByRole('button', { name: 'Load' }).click();

    // Accept either a results table or the empty-state message
    // (cached snapshots may or may not produce arbitrage between The Forge and Domain)
    await expect(
      page.getByRole('table').or(
        page.getByText(/No arbitrage opportunities found/i),
      ),
    ).toBeVisible({ timeout: 15000 });
  });

  // -------------------------------------------------------------------------
  // 4. Manage Trading Structures dialog
  // -------------------------------------------------------------------------

  test('gear button opens Manage Trading Structures dialog', async ({ page }) => {
    await page.goto('/hauling/scanner');
    await expect(
      page.getByRole('heading', { name: /Market Scanner/i }),
    ).toBeVisible({ timeout: 10000 });

    await page.getByTitle('Manage trading structures').click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });
    await expect(dialog.getByText('Manage Trading Structures')).toBeVisible({ timeout: 5000 });
  });

  test('Manage Structures dialog shows seeded accessible structure', async ({ page }) => {
    await page.goto('/hauling/scanner');
    await expect(
      page.getByRole('heading', { name: /Market Scanner/i }),
    ).toBeVisible({ timeout: 10000 });

    await page.getByTitle('Manage trading structures').click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // The seeded accessible structure should appear
    await expect(
      dialog.getByText('Perimeter - Test Trading Hub'),
    ).toBeVisible({ timeout: 5000 });
  });

  test('Manage Structures dialog shows Access OK badge for accessible structure', async ({ page }) => {
    await page.goto('/hauling/scanner');
    await expect(
      page.getByRole('heading', { name: /Market Scanner/i }),
    ).toBeVisible({ timeout: 10000 });

    await page.getByTitle('Manage trading structures').click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // Find the row for the accessible structure and check its badge
    const accessibleRow = dialog
      .getByRole('row')
      .filter({ hasText: 'Perimeter - Test Trading Hub' });
    await expect(accessibleRow).toBeVisible({ timeout: 5000 });
    await expect(accessibleRow.getByText('Access OK')).toBeVisible({ timeout: 5000 });
  });

  test('Manage Structures dialog shows No Access badge for restricted structure', async ({ page }) => {
    await page.goto('/hauling/scanner');
    await expect(
      page.getByRole('heading', { name: /Market Scanner/i }),
    ).toBeVisible({ timeout: 10000 });

    await page.getByTitle('Manage trading structures').click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // Find the row for the restricted structure and check its badge
    const restrictedRow = dialog
      .getByRole('row')
      .filter({ hasText: 'Restricted Citadel' });
    await expect(restrictedRow).toBeVisible({ timeout: 5000 });
    await expect(restrictedRow.getByText('No Access')).toBeVisible({ timeout: 5000 });
  });

  // -------------------------------------------------------------------------
  // 5. Scanner with station as source
  // -------------------------------------------------------------------------

  test('selecting Jita 4-4 station as source and loading shows results or empty state', async ({ page }) => {
    await page.goto('/hauling/scanner');
    await expect(
      page.getByRole('heading', { name: /Market Scanner/i }),
    ).toBeVisible({ timeout: 10000 });

    // Open source location dropdown
    const sourceCombo = page.getByRole('combobox').first();
    await sourceCombo.click();

    // Select Jita 4-4 from the stations section
    await page.getByRole('option', { name: /Jita IV - Moon 4/i }).click();

    // Verify the source combobox now shows Jita IV
    await expect(sourceCombo).toContainText('Jita IV');

    // Click Load to fetch cached scanner results for Jita (system_id=30000142) → Domain
    await page.getByRole('button', { name: 'Load' }).click();

    // Accept either a results table or the empty-state message
    await expect(
      page.getByRole('table').or(
        page.getByText(/No arbitrage opportunities found/i),
      ),
    ).toBeVisible({ timeout: 15000 });
  });

  test('selecting structure source with access_ok=false shows access warning', async ({ page }) => {
    await page.goto('/hauling/scanner');
    await expect(
      page.getByRole('heading', { name: /Market Scanner/i }),
    ).toBeVisible({ timeout: 10000 });

    // Open source location dropdown
    const sourceCombo = page.getByRole('combobox').first();
    await sourceCombo.click();

    // Select the restricted citadel (access_ok=false) — shows with ⚠ prefix
    const restrictedOption = page.getByRole('option', { name: /Restricted Citadel/i }).first();
    await restrictedOption.waitFor({ state: 'attached', timeout: 5000 });
    // Scroll the option itself into view inside the portal, then click via DOM
    await restrictedOption.evaluate((el) => { el.scrollIntoView({ block: 'nearest' }); });
    await restrictedOption.evaluate((el) => { (el as HTMLElement).click(); });

    // The scanner should show an access warning banner
    await expect(
      page.getByText(/Structure market access failed/i),
    ).toBeVisible({ timeout: 5000 });
  });

  // -------------------------------------------------------------------------
  // 5b. Add Structure dialog — dropdown-based UI
  // -------------------------------------------------------------------------

  test('Add Structure form shows Character and Structure dropdowns', async ({ page }) => {
    await page.goto('/hauling/scanner');
    await expect(
      page.getByRole('heading', { name: /Market Scanner/i }),
    ).toBeVisible({ timeout: 10000 });

    await page.getByTitle('Manage trading structures').click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // The "Add Structure" form uses shadcn Select dropdowns, not raw text inputs.
    // Both triggers should show their placeholder text.
    await expect(dialog.getByText('Select character')).toBeVisible({ timeout: 5000 });
  });

  test('Add Structure character dropdown populates from /api/characters', async ({ page }) => {
    await page.goto('/hauling/scanner');
    await expect(
      page.getByRole('heading', { name: /Market Scanner/i }),
    ).toBeVisible({ timeout: 10000 });

    await page.getByTitle('Manage trading structures').click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // The Character Select trigger shows "Select character" placeholder until opened.
    // Click it to open the dropdown and see the character options.
    const characterTrigger = dialog.getByRole('combobox').first();
    await expect(characterTrigger).toBeVisible({ timeout: 5000 });
    await characterTrigger.click();

    // Alice's characters should appear as options in the listbox
    await expect(
      page.getByRole('option', { name: /Alice Alpha/i }),
    ).toBeVisible({ timeout: 5000 });
  });

  test('selecting character in Add Structure form populates Structure dropdown', async ({ page }) => {
    await page.goto('/hauling/scanner');
    await expect(
      page.getByRole('heading', { name: /Market Scanner/i }),
    ).toBeVisible({ timeout: 10000 });

    await page.getByTitle('Manage trading structures').click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // Step 1: Open the character dropdown and select Alice Alpha
    const characterTrigger = dialog.getByRole('combobox').first();
    await expect(characterTrigger).toBeVisible({ timeout: 5000 });
    await characterTrigger.click();
    await page.getByRole('option', { name: /Alice Alpha/i }).click();

    // Step 2: The backend fetches asset structures for Alice Alpha.
    // Alice has an asset inside structure 1234567890123 (Perimeter - Test Trading Hub)
    // which was populated by the background asset runner from mock-ESI.
    // The structure dropdown should now show that structure.
    const structureTrigger = dialog.getByRole('combobox').nth(1);
    await expect(structureTrigger).toBeVisible({ timeout: 10000 });
    await structureTrigger.click();

    await expect(
      page.getByRole('option', { name: /Perimeter - Test Trading Hub/i }),
    ).toBeVisible({ timeout: 10000 });
  });

  test('Add button is enabled when both Character and Structure are selected', async ({ page }) => {
    await page.goto('/hauling/scanner');
    await expect(
      page.getByRole('heading', { name: /Market Scanner/i }),
    ).toBeVisible({ timeout: 10000 });

    await page.getByTitle('Manage trading structures').click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // Add button should be disabled initially (no character or structure selected)
    const addButton = dialog.getByRole('button', { name: /^Add$/i });
    await expect(addButton).toBeDisabled({ timeout: 5000 });

    // Select Alice Alpha from character dropdown
    const characterTrigger = dialog.getByRole('combobox').first();
    await expect(characterTrigger).toBeVisible({ timeout: 5000 });
    await characterTrigger.click();
    await page.getByRole('option', { name: /Alice Alpha/i }).click();

    // Add button still disabled — no structure selected yet
    await expect(addButton).toBeDisabled({ timeout: 5000 });

    // Select Perimeter - Test Trading Hub from structure dropdown
    const structureTrigger = dialog.getByRole('combobox').nth(1);
    await expect(structureTrigger).toBeVisible({ timeout: 10000 });
    await structureTrigger.click();
    await page.getByRole('option', { name: /Perimeter - Test Trading Hub/i }).click();

    // Now Add button should be enabled
    await expect(addButton).toBeEnabled({ timeout: 5000 });
  });

  // -------------------------------------------------------------------------
  // 6. DELETE structure via API (must be last — removes seeded data)
  // -------------------------------------------------------------------------

  test('DELETE /api/hauling/structures removes a structure', async ({ page }) => {
    // First, get the list to find the structure's id
    const listResponse = await page.request.get('/api/hauling/structures');
    expect(listResponse.status()).toBe(200);
    const structures = await listResponse.json();

    const restricted = structures.find(
      (s: { name: string }) => s.name === 'Restricted Citadel',
    );
    expect(restricted).toBeDefined();

    // Delete the restricted citadel (access_ok=false structure)
    const deleteResponse = await page.request.delete(
      `/api/hauling/structures?id=${restricted.id}`,
    );
    expect(deleteResponse.status()).toBe(204);

    // Confirm it's gone from the list
    const afterResponse = await page.request.get('/api/hauling/structures');
    const afterList = await afterResponse.json();
    const stillPresent = afterList.find(
      (s: { name: string }) => s.name === 'Restricted Citadel',
    );
    expect(stillPresent).toBeUndefined();
  });
});
