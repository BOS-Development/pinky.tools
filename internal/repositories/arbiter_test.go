package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetArbiterSettings_ReturnsDefaults_WhenNoRow(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)
	ctx := context.Background()

	// Use a user ID that has no settings row
	settings, err := repo.GetArbiterSettings(ctx, 9900)
	require.NoError(t, err)
	require.NotNil(t, settings)

	assert.Equal(t, int64(9900), settings.UserID)
	assert.Equal(t, "athanor", settings.ReactionStructure)
	assert.Equal(t, "t1", settings.ReactionRig)
	assert.Nil(t, settings.ReactionSystemID)

	assert.Equal(t, "raitaru", settings.InventionStructure)
	assert.Equal(t, "t1", settings.InventionRig)

	assert.Equal(t, "raitaru", settings.ComponentStructure)
	assert.Equal(t, "t2", settings.ComponentRig)

	assert.Equal(t, "raitaru", settings.FinalStructure)
	assert.Equal(t, "t2", settings.FinalRig)

	assert.True(t, settings.UseWhitelist)
	assert.True(t, settings.UseBlacklist)
}

func Test_UpsertArbiterSettings_SavesAndRetrieves(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	// Create user first (arbiter_settings references users)
	_, err = db.ExecContext(context.Background(), `INSERT INTO users (id, name) VALUES (9901, 'Arbiter Test User') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)
	ctx := context.Background()

	settings := &models.ArbiterSettings{
		UserID:             9901,
		ReactionStructure:  "tatara",
		ReactionRig:        "t2",
		ReactionSystemID:   nil,
		InventionStructure: "raitaru",
		InventionRig:       "t1",
		InventionSystemID:  nil,
		ComponentStructure: "azbel",
		ComponentRig:       "t2",
		ComponentSystemID:  nil,
		FinalStructure:     "sotiyo",
		FinalRig:           "t2",
		FinalSystemID:      nil,
		UseWhitelist:       true,
		UseBlacklist:       false,
	}

	err = repo.UpsertArbiterSettings(ctx, settings)
	require.NoError(t, err)

	retrieved, err := repo.GetArbiterSettings(ctx, 9901)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, int64(9901), retrieved.UserID)
	assert.Equal(t, "tatara", retrieved.ReactionStructure)
	assert.Equal(t, "t2", retrieved.ReactionRig)
	assert.Equal(t, "azbel", retrieved.ComponentStructure)
	assert.Equal(t, "sotiyo", retrieved.FinalStructure)
	assert.True(t, retrieved.UseWhitelist)
	assert.False(t, retrieved.UseBlacklist)
}

func Test_UpsertArbiterSettings_UpdatesExistingRow(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	_, err = db.ExecContext(context.Background(), `INSERT INTO users (id, name) VALUES (9902, 'Arbiter Update Test') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)
	ctx := context.Background()

	first := &models.ArbiterSettings{
		UserID:             9902,
		ReactionStructure:  "athanor",
		ReactionRig:        "t1",
		InventionStructure: "raitaru",
		InventionRig:       "t1",
		ComponentStructure: "raitaru",
		ComponentRig:       "t2",
		FinalStructure:     "raitaru",
		FinalRig:           "t2",
		UseWhitelist:       true,
		UseBlacklist:       true,
	}
	require.NoError(t, repo.UpsertArbiterSettings(ctx, first))

	second := &models.ArbiterSettings{
		UserID:             9902,
		ReactionStructure:  "tatara",
		ReactionRig:        "t2",
		InventionStructure: "azbel",
		InventionRig:       "t2",
		ComponentStructure: "sotiyo",
		ComponentRig:       "t2",
		FinalStructure:     "station",
		FinalRig:           "none",
		UseWhitelist:       false,
		UseBlacklist:       false,
	}
	require.NoError(t, repo.UpsertArbiterSettings(ctx, second))

	retrieved, err := repo.GetArbiterSettings(ctx, 9902)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, "tatara", retrieved.ReactionStructure)
	assert.Equal(t, "station", retrieved.FinalStructure)
	assert.Equal(t, "none", retrieved.FinalRig)
	assert.False(t, retrieved.UseWhitelist)
	assert.False(t, retrieved.UseBlacklist)
}

func Test_GetDecryptors_ReturnsEmpty_WhenNonePopulated(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)
	ctx := context.Background()

	decryptors, err := repo.GetDecryptors(ctx)
	require.NoError(t, err)
	assert.NotNil(t, decryptors)
	// May be empty if sde_decryptors not populated — just ensure no error
}

func Test_GetT2BlueprintsForScan_ReturnsEmpty_WhenNoT2Data(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)
	ctx := context.Background()

	items, err := repo.GetT2BlueprintsForScan(ctx)
	require.NoError(t, err)
	assert.NotNil(t, items)
	// Fresh DB has no SDE data, so should be empty
	assert.Empty(t, items)
}

func Test_GetArbiterEnabled_ReturnsFalse_ForNewUser(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	_, err = db.ExecContext(context.Background(), `INSERT INTO users (id, name) VALUES (9903, 'Arbiter Enabled Test') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)
	ctx := context.Background()

	enabled, err := repo.GetArbiterEnabled(ctx, 9903)
	require.NoError(t, err)
	assert.False(t, enabled)
}

func Test_GetArbiterEnabled_ReturnsTrue_WhenEnabled(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	_, err = db.ExecContext(context.Background(), `INSERT INTO users (id, name) VALUES (9904, 'Arbiter Enabled True') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	_, err = db.ExecContext(context.Background(), `UPDATE users SET arbiter_enabled = true WHERE id = 9904`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)
	ctx := context.Background()

	enabled, err := repo.GetArbiterEnabled(ctx, 9904)
	require.NoError(t, err)
	assert.True(t, enabled)
}

func Test_UpsertDecryptors_Succeeds_WithEmptyDogmaData(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)
	ctx := context.Background()

	// Should succeed even with no dogma data — just inserts nothing
	err = repo.UpsertDecryptors(ctx)
	require.NoError(t, err)
}

// --- Scope tests ---

func Test_CreateAndGetScope(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)
	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO users (id, name) VALUES (9910, 'Scope Test User') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)

	scope := &models.ArbiterScope{
		UserID:    9910,
		Name:      "My Test Scope",
		IsDefault: false,
	}
	id, err := repo.CreateScope(ctx, scope)
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	retrieved, err := repo.GetScope(ctx, id, 9910)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, "My Test Scope", retrieved.Name)
	assert.Equal(t, int64(9910), retrieved.UserID)
	assert.False(t, retrieved.IsDefault)
}

func Test_GetScope_ReturnsNil_ForWrongUser(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)
	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO users (id, name) VALUES (9911, 'Scope User A') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO users (id, name) VALUES (9912, 'Scope User B') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)

	id, err := repo.CreateScope(ctx, &models.ArbiterScope{UserID: 9911, Name: "Private Scope"})
	require.NoError(t, err)

	// User 9912 should not be able to see scope owned by 9911
	retrieved, err := repo.GetScope(ctx, id, 9912)
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}

func Test_UpdateScope(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)
	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO users (id, name) VALUES (9913, 'Scope Update User') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)

	id, err := repo.CreateScope(ctx, &models.ArbiterScope{UserID: 9913, Name: "Old Name"})
	require.NoError(t, err)

	err = repo.UpdateScope(ctx, &models.ArbiterScope{ID: id, UserID: 9913, Name: "New Name", IsDefault: true})
	require.NoError(t, err)

	retrieved, err := repo.GetScope(ctx, id, 9913)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, "New Name", retrieved.Name)
	assert.True(t, retrieved.IsDefault)
}

func Test_DeleteScope(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)
	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO users (id, name) VALUES (9914, 'Scope Delete User') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)

	id, err := repo.CreateScope(ctx, &models.ArbiterScope{UserID: 9914, Name: "To Delete"})
	require.NoError(t, err)

	err = repo.DeleteScope(ctx, id, 9914)
	require.NoError(t, err)

	retrieved, err := repo.GetScope(ctx, id, 9914)
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}

func Test_GetScopes_ReturnsList(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)
	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO users (id, name) VALUES (9915, 'Scope List User') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)

	_, err = repo.CreateScope(ctx, &models.ArbiterScope{UserID: 9915, Name: "Alpha"})
	require.NoError(t, err)
	_, err = repo.CreateScope(ctx, &models.ArbiterScope{UserID: 9915, Name: "Beta"})
	require.NoError(t, err)

	scopes, err := repo.GetScopes(ctx, 9915)
	require.NoError(t, err)
	assert.Len(t, scopes, 2)
	// Ordered by name
	assert.Equal(t, "Alpha", scopes[0].Name)
	assert.Equal(t, "Beta", scopes[1].Name)
}

func Test_AddAndGetScopeMembers(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)
	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO users (id, name) VALUES (9916, 'Member Test User') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)

	scopeID, err := repo.CreateScope(ctx, &models.ArbiterScope{UserID: 9916, Name: "Member Scope"})
	require.NoError(t, err)

	// Add a member (no FK on member_id — just a bare ID)
	err = repo.AddScopeMember(ctx, &models.ArbiterScopeMember{
		ScopeID:    scopeID,
		MemberType: "character",
		MemberID:   99999,
	})
	require.NoError(t, err)

	// Adding same member twice should be a no-op
	err = repo.AddScopeMember(ctx, &models.ArbiterScopeMember{
		ScopeID:    scopeID,
		MemberType: "character",
		MemberID:   99999,
	})
	require.NoError(t, err)

	members, err := repo.GetScopeMembers(ctx, scopeID)
	require.NoError(t, err)
	assert.Len(t, members, 1)
	assert.Equal(t, "character", members[0].MemberType)
	assert.Equal(t, int64(99999), members[0].MemberID)
}

func Test_RemoveScopeMember(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)
	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO users (id, name) VALUES (9917, 'Remove Member User') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)

	scopeID, err := repo.CreateScope(ctx, &models.ArbiterScope{UserID: 9917, Name: "Remove Scope"})
	require.NoError(t, err)

	err = repo.AddScopeMember(ctx, &models.ArbiterScopeMember{
		ScopeID:    scopeID,
		MemberType: "character",
		MemberID:   88888,
	})
	require.NoError(t, err)

	members, err := repo.GetScopeMembers(ctx, scopeID)
	require.NoError(t, err)
	require.Len(t, members, 1)

	err = repo.RemoveScopeMember(ctx, members[0].ID, scopeID)
	require.NoError(t, err)

	members, err = repo.GetScopeMembers(ctx, scopeID)
	require.NoError(t, err)
	assert.Empty(t, members)
}

// --- Tax Profile tests ---

func Test_GetTaxProfile_ReturnsDefaults_WhenNoRow(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)
	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO users (id, name) VALUES (9920, 'Tax Profile User') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)

	profile, err := repo.GetTaxProfile(ctx, 9920)
	require.NoError(t, err)
	require.NotNil(t, profile)
	assert.Equal(t, int64(9920), profile.UserID)
	assert.InDelta(t, 0.036, profile.SalesTaxRate, 0.001)
	assert.InDelta(t, 0.03, profile.BrokerFeeRate, 0.001)
	assert.Equal(t, "sell", profile.InputPriceType)
	assert.Equal(t, "buy", profile.OutputPriceType)
}

func Test_UpsertTaxProfile(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)
	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO users (id, name) VALUES (9921, 'Tax Profile Upsert User') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)

	profile := &models.ArbiterTaxProfile{
		UserID:             9921,
		SalesTaxRate:       0.02,
		BrokerFeeRate:      0.01,
		StructureBrokerFee: 0.005,
		InputPriceType:     "buy",
		OutputPriceType:    "sell",
	}
	err = repo.UpsertTaxProfile(ctx, profile)
	require.NoError(t, err)

	retrieved, err := repo.GetTaxProfile(ctx, 9921)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.InDelta(t, 0.02, retrieved.SalesTaxRate, 0.0001)
	assert.Equal(t, "buy", retrieved.InputPriceType)
	assert.Equal(t, "sell", retrieved.OutputPriceType)
}

// --- Blacklist / Whitelist tests ---

func Test_BlacklistAddAndGet(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)
	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO users (id, name) VALUES (9930, 'BL Test User') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)

	err = repo.AddToBlacklist(ctx, 9930, 100)
	require.NoError(t, err)

	// Adding same item twice should be a no-op
	err = repo.AddToBlacklist(ctx, 9930, 100)
	require.NoError(t, err)

	items, err := repo.GetBlacklist(ctx, 9930)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, int64(100), items[0].TypeID)
	assert.Equal(t, int64(9930), items[0].UserID)
	assert.False(t, items[0].AddedAt.IsZero())
}

func Test_BlacklistRemove(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)
	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO users (id, name) VALUES (9931, 'BL Remove User') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)

	err = repo.AddToBlacklist(ctx, 9931, 200)
	require.NoError(t, err)

	err = repo.RemoveFromBlacklist(ctx, 9931, 200)
	require.NoError(t, err)

	items, err := repo.GetBlacklist(ctx, 9931)
	require.NoError(t, err)
	assert.Empty(t, items)
}

func Test_WhitelistAddAndGet(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)
	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO users (id, name) VALUES (9932, 'WL Test User') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)

	err = repo.AddToWhitelist(ctx, 9932, 300)
	require.NoError(t, err)

	items, err := repo.GetWhitelist(ctx, 9932)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, int64(300), items[0].TypeID)
}

func Test_WhitelistRemove(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)
	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO users (id, name) VALUES (9933, 'WL Remove User') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)

	err = repo.AddToWhitelist(ctx, 9933, 400)
	require.NoError(t, err)

	err = repo.RemoveFromWhitelist(ctx, 9933, 400)
	require.NoError(t, err)

	items, err := repo.GetWhitelist(ctx, 9933)
	require.NoError(t, err)
	assert.Empty(t, items)
}

// --- GetScopeAssets tests ---

func Test_GetScopeAssets_ReturnsEmpty_WhenNoMembers(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)
	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO users (id, name) VALUES (9940, 'Assets Test User') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)

	scopeID, err := repo.CreateScope(ctx, &models.ArbiterScope{UserID: 9940, Name: "Empty Scope"})
	require.NoError(t, err)

	assets, err := repo.GetScopeAssets(ctx, scopeID, 9940)
	require.NoError(t, err)
	assert.NotNil(t, assets)
	assert.Empty(t, assets)
}

func Test_GetScopeAssets_ReturnsEmpty_ForUnknownScope(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)
	ctx := context.Background()

	repo := repositories.NewArbiterRepository(db)

	// Non-existent scope — should return empty map, not error
	assets, err := repo.GetScopeAssets(ctx, 99999999, 9941)
	require.NoError(t, err)
	assert.NotNil(t, assets)
	assert.Empty(t, assets)
}

// --- GetDemandStats tests ---

func Test_GetDemandStats_ReturnsEmpty_WhenNoHistory(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)
	ctx := context.Background()

	repo := repositories.NewArbiterRepository(db)

	stats, err := repo.GetDemandStats(ctx, []int64{99999})
	require.NoError(t, err)
	assert.NotNil(t, stats)
}

func Test_GetDemandStats_ReturnsEmpty_ForEmptyTypeIDs(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)
	ctx := context.Background()

	repo := repositories.NewArbiterRepository(db)

	stats, err := repo.GetDemandStats(ctx, []int64{})
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Empty(t, stats)
}

// --- SearchSolarSystems tests ---

func Test_SearchSolarSystems_ReturnsEmpty_WhenNoMatch(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)
	ctx := context.Background()

	repo := repositories.NewArbiterRepository(db)

	results, err := repo.SearchSolarSystems(ctx, "zzzzz_no_match", 10)
	require.NoError(t, err)
	assert.NotNil(t, results)
	assert.Empty(t, results)
}

// --- InsertPriceHistorySnapshot tests ---

func Test_InsertPriceHistorySnapshot_Succeeds_WithEmptyPrices(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)
	ctx := context.Background()

	// Direct test of the market prices repo's snapshot method
	mpRepo := repositories.NewMarketPrices(db)

	// Should succeed even with no Jita prices
	err = mpRepo.InsertPriceHistorySnapshot(ctx, time.Now())
	require.NoError(t, err)

	// Second call same day should be a no-op (ON CONFLICT DO NOTHING)
	err = mpRepo.InsertPriceHistorySnapshot(ctx, time.Now())
	require.NoError(t, err)
}

// --- GetSecurityClassForSystem tests ---

func Test_GetSecurityClassForSystem_ReturnsHigh_ForHighSecSystem(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)
	ctx := context.Background()

	// Insert minimal universe hierarchy for a high-sec test system (security = 0.9)
	_, err = db.ExecContext(ctx, `INSERT INTO regions (region_id, name) VALUES (99010000, 'Test Region Sec') ON CONFLICT (region_id) DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO constellations (constellation_id, name, region_id) VALUES (99020000, 'Test Constellation Sec', 99010000) ON CONFLICT (constellation_id) DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES (99030001, 'Test High Sec System', 99020000, 0.9) ON CONFLICT (solar_system_id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)

	sec, err := repo.GetSecurityClassForSystem(ctx, 99030001)
	require.NoError(t, err)
	assert.Equal(t, "high", sec)
}

func Test_GetSecurityClassForSystem_ReturnsLow_ForLowSecSystem(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)
	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO regions (region_id, name) VALUES (99010001, 'Test Region Low') ON CONFLICT (region_id) DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO constellations (constellation_id, name, region_id) VALUES (99020001, 'Test Constellation Low', 99010001) ON CONFLICT (constellation_id) DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES (99030002, 'Test Low Sec System', 99020001, 0.2) ON CONFLICT (solar_system_id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)

	sec, err := repo.GetSecurityClassForSystem(ctx, 99030002)
	require.NoError(t, err)
	assert.Equal(t, "low", sec)
}

func Test_GetSecurityClassForSystem_ReturnsNull_ForNullSecSystem(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)
	ctx := context.Background()

	_, err = db.ExecContext(ctx, `INSERT INTO regions (region_id, name) VALUES (99010002, 'Test Region Null') ON CONFLICT (region_id) DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO constellations (constellation_id, name, region_id) VALUES (99020002, 'Test Constellation Null', 99010002) ON CONFLICT (constellation_id) DO NOTHING`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES (99030003, 'Test Null Sec System', 99020002, -0.5) ON CONFLICT (solar_system_id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)

	sec, err := repo.GetSecurityClassForSystem(ctx, 99030003)
	require.NoError(t, err)
	assert.Equal(t, "null", sec)
}

func Test_GetSecurityClassForSystem_ReturnsNull_WhenSystemNotFound(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)
	ctx := context.Background()

	repo := repositories.NewArbiterRepository(db)

	// Use a system ID that does not exist
	sec, err := repo.GetSecurityClassForSystem(ctx, 99999998)
	require.NoError(t, err)
	assert.Equal(t, "null", sec)
}
