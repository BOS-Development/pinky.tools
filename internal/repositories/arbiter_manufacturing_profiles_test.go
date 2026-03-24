package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ManufacturingProfiles_ListEmpty(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)
	ctx := context.Background()

	profiles, err := repo.ListManufacturingProfiles(ctx, 8100)
	require.NoError(t, err)
	assert.Equal(t, []*models.ArbiterManufacturingProfile{}, profiles)
}

func Test_ManufacturingProfiles_CreateAndList(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	_, err = db.ExecContext(context.Background(), `INSERT INTO users (id, name) VALUES (8101, 'MfgProfile Test') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)
	ctx := context.Background()

	p := &models.ArbiterManufacturingProfile{
		UserID:               8101,
		Name:                 "My Raitaru Setup",
		ReactionStructure:    "athanor",
		ReactionRig:          "t1",
		ReactionFacilityTax:  0.3,
		InventionStructure:   "raitaru",
		InventionRig:         "t1",
		InventionFacilityTax: 0.0,
		ComponentStructure:   "raitaru",
		ComponentRig:         "t2",
		ComponentFacilityTax: 0.5,
		FinalStructure:       "raitaru",
		FinalRig:             "t2",
		FinalFacilityTax:     0.5,
	}

	created, err := repo.CreateManufacturingProfile(ctx, p)
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Greater(t, created.ID, int64(0))
	assert.Equal(t, int64(8101), created.UserID)
	assert.Equal(t, "My Raitaru Setup", created.Name)
	assert.Equal(t, "athanor", created.ReactionStructure)
	assert.Equal(t, "t1", created.ReactionRig)
	assert.Nil(t, created.ReactionSystemID)
	assert.Equal(t, 0.3, created.ReactionFacilityTax)
	assert.Equal(t, "raitaru", created.FinalStructure)
	assert.Equal(t, "t2", created.FinalRig)
	assert.Equal(t, 0.5, created.FinalFacilityTax)

	profiles, err := repo.ListManufacturingProfiles(ctx, 8101)
	require.NoError(t, err)
	require.Len(t, profiles, 1)
	assert.Equal(t, created.ID, profiles[0].ID)
	assert.Equal(t, "My Raitaru Setup", profiles[0].Name)
}

func Test_ManufacturingProfiles_ListOrderedByName(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	_, err = db.ExecContext(context.Background(), `INSERT INTO users (id, name) VALUES (8102, 'MfgProfile Test 2') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)
	ctx := context.Background()

	for _, name := range []string{"Zeta Setup", "Alpha Setup", "Beta Setup"} {
		_, err := repo.CreateManufacturingProfile(ctx, &models.ArbiterManufacturingProfile{
			UserID: 8102,
			Name:   name,
		})
		require.NoError(t, err)
	}

	profiles, err := repo.ListManufacturingProfiles(ctx, 8102)
	require.NoError(t, err)
	require.Len(t, profiles, 3)
	assert.Equal(t, "Alpha Setup", profiles[0].Name)
	assert.Equal(t, "Beta Setup", profiles[1].Name)
	assert.Equal(t, "Zeta Setup", profiles[2].Name)
}

func Test_ManufacturingProfiles_Get(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	_, err = db.ExecContext(context.Background(), `INSERT INTO users (id, name) VALUES (8103, 'MfgProfile Test 3') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)
	ctx := context.Background()

	created, err := repo.CreateManufacturingProfile(ctx, &models.ArbiterManufacturingProfile{
		UserID:             8103,
		Name:               "Get Test",
		ReactionStructure:  "tatara",
		ReactionRig:        "t2",
		FinalFacilityTax:   1.5,
	})
	require.NoError(t, err)

	// Get by ID and userID
	fetched, err := repo.GetManufacturingProfile(ctx, created.ID, 8103)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, "Get Test", fetched.Name)
	assert.Equal(t, "tatara", fetched.ReactionStructure)
	assert.Equal(t, "t2", fetched.ReactionRig)
	assert.Equal(t, 1.5, fetched.FinalFacilityTax)

	// Get with wrong userID returns nil
	notFound, err := repo.GetManufacturingProfile(ctx, created.ID, 9999)
	require.NoError(t, err)
	assert.Nil(t, notFound)
}

func Test_ManufacturingProfiles_Update(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	_, err = db.ExecContext(context.Background(), `INSERT INTO users (id, name) VALUES (8104, 'MfgProfile Test 4') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)
	ctx := context.Background()

	created, err := repo.CreateManufacturingProfile(ctx, &models.ArbiterManufacturingProfile{
		UserID:             8104,
		Name:               "Original Name",
		ReactionStructure:  "athanor",
		FinalFacilityTax:   0.0,
	})
	require.NoError(t, err)

	created.Name = "Updated Name"
	created.FinalStructure = "sotiyo"
	created.FinalRig = "t2"
	created.FinalFacilityTax = 2.0

	updated, err := repo.UpdateManufacturingProfile(ctx, created)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "sotiyo", updated.FinalStructure)
	assert.Equal(t, "t2", updated.FinalRig)
	assert.Equal(t, 2.0, updated.FinalFacilityTax)

	// Update with wrong userID returns nil
	created.UserID = 9999
	notFound, err := repo.UpdateManufacturingProfile(ctx, created)
	require.NoError(t, err)
	assert.Nil(t, notFound)
}

func Test_ManufacturingProfiles_Delete(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	_, err = db.ExecContext(context.Background(), `INSERT INTO users (id, name) VALUES (8105, 'MfgProfile Test 5') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)
	ctx := context.Background()

	created, err := repo.CreateManufacturingProfile(ctx, &models.ArbiterManufacturingProfile{
		UserID: 8105,
		Name:   "To Delete",
	})
	require.NoError(t, err)

	err = repo.DeleteManufacturingProfile(ctx, created.ID, 8105)
	require.NoError(t, err)

	// Verify it's gone
	profiles, err := repo.ListManufacturingProfiles(ctx, 8105)
	require.NoError(t, err)
	assert.Empty(t, profiles)
}

func Test_ManufacturingProfiles_DeleteWithWrongUser_DoesNothing(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	_, err = db.ExecContext(context.Background(), `INSERT INTO users (id, name) VALUES (8106, 'MfgProfile Test 6') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)
	ctx := context.Background()

	created, err := repo.CreateManufacturingProfile(ctx, &models.ArbiterManufacturingProfile{
		UserID: 8106,
		Name:   "Still Here",
	})
	require.NoError(t, err)

	// Delete with wrong userID - should not error, just delete nothing
	err = repo.DeleteManufacturingProfile(ctx, created.ID, 9999)
	require.NoError(t, err)

	// Row should still be there
	profiles, err := repo.ListManufacturingProfiles(ctx, 8106)
	require.NoError(t, err)
	require.Len(t, profiles, 1)
}

func Test_ManufacturingProfiles_CreateWithSystemID(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	_, err = db.ExecContext(context.Background(), `INSERT INTO users (id, name) VALUES (8107, 'MfgProfile Test 7') ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)

	// Seed a solar system for the FK reference
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO regions (region_id, name) VALUES (88000001, 'Test Region 8107') ON CONFLICT (region_id) DO NOTHING;
		INSERT INTO constellations (constellation_id, name, region_id) VALUES (88000001, 'Test Constellation 8107', 88000001) ON CONFLICT (constellation_id) DO NOTHING;
		INSERT INTO solar_systems (solar_system_id, name, constellation_id, security) VALUES (88000001, 'Test System 8107', 88000001, 0.5) ON CONFLICT (solar_system_id) DO NOTHING;
	`)
	require.NoError(t, err)

	repo := repositories.NewArbiterRepository(db)
	ctx := context.Background()

	systemID := int64(88000001)
	created, err := repo.CreateManufacturingProfile(ctx, &models.ArbiterManufacturingProfile{
		UserID:            8107,
		Name:              "With System",
		ReactionSystemID:  &systemID,
		FinalSystemID:     &systemID,
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	require.NotNil(t, created.ReactionSystemID)
	assert.Equal(t, systemID, *created.ReactionSystemID)
	require.NotNil(t, created.FinalSystemID)
	assert.Equal(t, systemID, *created.FinalSystemID)
}
