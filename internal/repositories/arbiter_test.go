package repositories_test

import (
	"context"
	"testing"

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
	assert.Equal(t, "null", settings.ReactionSecurity)
	assert.Nil(t, settings.ReactionSystemID)

	assert.Equal(t, "raitaru", settings.InventionStructure)
	assert.Equal(t, "t1", settings.InventionRig)
	assert.Equal(t, "high", settings.InventionSecurity)

	assert.Equal(t, "raitaru", settings.ComponentStructure)
	assert.Equal(t, "t2", settings.ComponentRig)
	assert.Equal(t, "null", settings.ComponentSecurity)

	assert.Equal(t, "raitaru", settings.FinalStructure)
	assert.Equal(t, "t2", settings.FinalRig)
	assert.Equal(t, "null", settings.FinalSecurity)
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
		ReactionSecurity:   "null",
		ReactionSystemID:   nil,
		InventionStructure: "raitaru",
		InventionRig:       "t1",
		InventionSecurity:  "high",
		InventionSystemID:  nil,
		ComponentStructure: "azbel",
		ComponentRig:       "t2",
		ComponentSecurity:  "low",
		ComponentSystemID:  nil,
		FinalStructure:     "sotiyo",
		FinalRig:           "t2",
		FinalSecurity:      "null",
		FinalSystemID:      nil,
	}

	err = repo.UpsertArbiterSettings(ctx, settings)
	require.NoError(t, err)

	retrieved, err := repo.GetArbiterSettings(ctx, 9901)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, int64(9901), retrieved.UserID)
	assert.Equal(t, "tatara", retrieved.ReactionStructure)
	assert.Equal(t, "t2", retrieved.ReactionRig)
	assert.Equal(t, "null", retrieved.ReactionSecurity)
	assert.Equal(t, "azbel", retrieved.ComponentStructure)
	assert.Equal(t, "sotiyo", retrieved.FinalStructure)
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
		ReactionSecurity:   "null",
		InventionStructure: "raitaru",
		InventionRig:       "t1",
		InventionSecurity:  "high",
		ComponentStructure: "raitaru",
		ComponentRig:       "t2",
		ComponentSecurity:  "null",
		FinalStructure:     "raitaru",
		FinalRig:           "t2",
		FinalSecurity:      "null",
	}
	require.NoError(t, repo.UpsertArbiterSettings(ctx, first))

	second := &models.ArbiterSettings{
		UserID:             9902,
		ReactionStructure:  "tatara",
		ReactionRig:        "t2",
		ReactionSecurity:   "low",
		InventionStructure: "azbel",
		InventionRig:       "t2",
		InventionSecurity:  "null",
		ComponentStructure: "sotiyo",
		ComponentRig:       "t2",
		ComponentSecurity:  "null",
		FinalStructure:     "station",
		FinalRig:           "none",
		FinalSecurity:      "high",
	}
	require.NoError(t, repo.UpsertArbiterSettings(ctx, second))

	retrieved, err := repo.GetArbiterSettings(ctx, 9902)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, "tatara", retrieved.ReactionStructure)
	assert.Equal(t, "station", retrieved.FinalStructure)
	assert.Equal(t, "none", retrieved.FinalRig)
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
