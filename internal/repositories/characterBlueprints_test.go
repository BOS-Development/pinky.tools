package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_CharacterBlueprints_ReplaceBlueprints_InsertsData(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)
	bpRepo := repositories.NewCharacterBlueprints(db)

	user := &repositories.User{ID: 7000, Name: "BP Test User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 70001, Name: "BP Char", UserID: user.ID}
	err = charRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	blueprints := []*models.CharacterBlueprint{
		{
			ItemID:             80001,
			TypeID:             787,
			LocationID:         60003760,
			LocationFlag:       "Hangar",
			Quantity:           -1,
			MaterialEfficiency: 10,
			TimeEfficiency:     20,
			Runs:               -1,
		},
		{
			ItemID:             80002,
			TypeID:             46166,
			LocationID:         60003760,
			LocationFlag:       "Hangar",
			Quantity:           -2,
			MaterialEfficiency: 8,
			TimeEfficiency:     16,
			Runs:               50,
		},
	}

	err = bpRepo.ReplaceBlueprints(context.Background(), char.ID, "character", user.ID, blueprints)
	assert.NoError(t, err)

	// Verify data was inserted by querying blueprint levels
	levels, err := bpRepo.GetBlueprintLevels(context.Background(), user.ID, []int64{787, 46166})
	assert.NoError(t, err)
	assert.Len(t, levels, 2)

	bp787 := levels[787]
	assert.NotNil(t, bp787)
	assert.Equal(t, 10, bp787.MaterialEfficiency)
	assert.Equal(t, 20, bp787.TimeEfficiency)
	assert.False(t, bp787.IsCopy)

	bp46166 := levels[46166]
	assert.NotNil(t, bp46166)
	assert.Equal(t, 8, bp46166.MaterialEfficiency)
	assert.True(t, bp46166.IsCopy)
	assert.Equal(t, 50, bp46166.Runs)
}

func Test_CharacterBlueprints_ReplaceBlueprints_ReplacesOnReCall(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)
	bpRepo := repositories.NewCharacterBlueprints(db)

	user := &repositories.User{ID: 7010, Name: "BP Replace User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 70101, Name: "BP Replace Char", UserID: user.ID}
	err = charRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	initial := []*models.CharacterBlueprint{
		{
			ItemID:             81001,
			TypeID:             787,
			LocationID:         60003760,
			LocationFlag:       "Hangar",
			Quantity:           -1,
			MaterialEfficiency: 5,
			TimeEfficiency:     10,
			Runs:               -1,
		},
	}
	err = bpRepo.ReplaceBlueprints(context.Background(), char.ID, "character", user.ID, initial)
	assert.NoError(t, err)

	// Replace with new data — different item_id, updated ME
	updated := []*models.CharacterBlueprint{
		{
			ItemID:             81002,
			TypeID:             787,
			LocationID:         60003760,
			LocationFlag:       "Hangar",
			Quantity:           -1,
			MaterialEfficiency: 10,
			TimeEfficiency:     20,
			Runs:               -1,
		},
	}
	err = bpRepo.ReplaceBlueprints(context.Background(), char.ID, "character", user.ID, updated)
	assert.NoError(t, err)

	levels, err := bpRepo.GetBlueprintLevels(context.Background(), user.ID, []int64{787})
	assert.NoError(t, err)
	assert.Len(t, levels, 1)
	assert.Equal(t, 10, levels[787].MaterialEfficiency)
}

func Test_CharacterBlueprints_ReplaceBlueprints_HandlesEmptySlice(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)
	bpRepo := repositories.NewCharacterBlueprints(db)

	user := &repositories.User{ID: 7020, Name: "BP Empty User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 70201, Name: "BP Empty Char", UserID: user.ID}
	err = charRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	// Insert something first
	initial := []*models.CharacterBlueprint{
		{
			ItemID:     82001,
			TypeID:     787,
			LocationID: 60003760,
			Quantity:   -1,
			Runs:       -1,
		},
	}
	err = bpRepo.ReplaceBlueprints(context.Background(), char.ID, "character", user.ID, initial)
	assert.NoError(t, err)

	// Now replace with empty slice — should delete everything
	err = bpRepo.ReplaceBlueprints(context.Background(), char.ID, "character", user.ID, []*models.CharacterBlueprint{})
	assert.NoError(t, err)

	levels, err := bpRepo.GetBlueprintLevels(context.Background(), user.ID, []int64{787})
	assert.NoError(t, err)
	assert.Len(t, levels, 0)
}

func Test_CharacterBlueprints_GetBlueprintLevels_PrefersBPCOverBPO(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)
	bpRepo := repositories.NewCharacterBlueprints(db)

	user := &repositories.User{ID: 7030, Name: "BP Prefer User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	// Two characters — one with BPO, one with BPC
	charA := &repositories.Character{ID: 70301, Name: "BPO Owner", UserID: user.ID}
	err = charRepo.Add(context.Background(), charA)
	assert.NoError(t, err)

	charB := &repositories.Character{ID: 70302, Name: "BPC Owner", UserID: user.ID}
	err = charRepo.Add(context.Background(), charB)
	assert.NoError(t, err)

	// charA has BPO ME10
	bpoBlueprints := []*models.CharacterBlueprint{
		{
			ItemID:             83001,
			TypeID:             787,
			LocationID:         60003760,
			LocationFlag:       "Hangar",
			Quantity:           -1, // BPO
			MaterialEfficiency: 10,
			TimeEfficiency:     20,
			Runs:               -1,
		},
	}
	err = bpRepo.ReplaceBlueprints(context.Background(), charA.ID, "character", user.ID, bpoBlueprints)
	assert.NoError(t, err)

	// charB has BPC ME8
	bpcBlueprints := []*models.CharacterBlueprint{
		{
			ItemID:             83002,
			TypeID:             787,
			LocationID:         60003760,
			LocationFlag:       "Hangar",
			Quantity:           -2, // BPC
			MaterialEfficiency: 8,
			TimeEfficiency:     16,
			Runs:               10,
		},
	}
	err = bpRepo.ReplaceBlueprints(context.Background(), charB.ID, "character", user.ID, bpcBlueprints)
	assert.NoError(t, err)

	levels, err := bpRepo.GetBlueprintLevels(context.Background(), user.ID, []int64{787})
	assert.NoError(t, err)

	bp := levels[787]
	assert.NotNil(t, bp)
	// BPC should be preferred even though ME is lower
	assert.True(t, bp.IsCopy)
	assert.Equal(t, 8, bp.MaterialEfficiency)
}

func Test_CharacterBlueprints_GetBlueprintLevels_PrefersHighestMEWhenSameType(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)
	bpRepo := repositories.NewCharacterBlueprints(db)

	user := &repositories.User{ID: 7040, Name: "BP ME Prefer User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	charA := &repositories.Character{ID: 70401, Name: "Low ME Char", UserID: user.ID}
	err = charRepo.Add(context.Background(), charA)
	assert.NoError(t, err)

	charB := &repositories.Character{ID: 70402, Name: "High ME Char", UserID: user.ID}
	err = charRepo.Add(context.Background(), charB)
	assert.NoError(t, err)

	// charA has BPO ME5
	err = bpRepo.ReplaceBlueprints(context.Background(), charA.ID, "character", user.ID, []*models.CharacterBlueprint{
		{ItemID: 84001, TypeID: 788, LocationID: 60003760, Quantity: -1, MaterialEfficiency: 5, TimeEfficiency: 10, Runs: -1},
	})
	assert.NoError(t, err)

	// charB has BPO ME10
	err = bpRepo.ReplaceBlueprints(context.Background(), charB.ID, "character", user.ID, []*models.CharacterBlueprint{
		{ItemID: 84002, TypeID: 788, LocationID: 60003760, Quantity: -1, MaterialEfficiency: 10, TimeEfficiency: 20, Runs: -1},
	})
	assert.NoError(t, err)

	levels, err := bpRepo.GetBlueprintLevels(context.Background(), user.ID, []int64{788})
	assert.NoError(t, err)

	bp := levels[788]
	assert.NotNil(t, bp)
	// Highest ME should win among BPOs
	assert.Equal(t, 10, bp.MaterialEfficiency)
	assert.False(t, bp.IsCopy)
}

func Test_CharacterBlueprints_GetBlueprintLevels_ReturnsEmptyForUnknownTypes(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	bpRepo := repositories.NewCharacterBlueprints(db)

	levels, err := bpRepo.GetBlueprintLevels(context.Background(), 9999999, []int64{99999, 88888})
	assert.NoError(t, err)
	assert.Len(t, levels, 0)
}

func Test_CharacterBlueprints_GetBlueprintLevels_EmptyTypeIDs(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	bpRepo := repositories.NewCharacterBlueprints(db)

	levels, err := bpRepo.GetBlueprintLevels(context.Background(), 1, []int64{})
	assert.NoError(t, err)
	assert.Len(t, levels, 0)
}

func Test_CharacterBlueprints_DeleteByOwner_RemovesOnlyMatchingOwner(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	charRepo := repositories.NewCharacterRepository(db)
	bpRepo := repositories.NewCharacterBlueprints(db)

	user := &repositories.User{ID: 7050, Name: "BP Delete User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	charA := &repositories.Character{ID: 70501, Name: "Delete Char A", UserID: user.ID}
	err = charRepo.Add(context.Background(), charA)
	assert.NoError(t, err)

	charB := &repositories.Character{ID: 70502, Name: "Delete Char B", UserID: user.ID}
	err = charRepo.Add(context.Background(), charB)
	assert.NoError(t, err)

	// Insert blueprints for both characters
	err = bpRepo.ReplaceBlueprints(context.Background(), charA.ID, "character", user.ID, []*models.CharacterBlueprint{
		{ItemID: 85001, TypeID: 790, LocationID: 60003760, Quantity: -1, Runs: -1},
	})
	assert.NoError(t, err)

	err = bpRepo.ReplaceBlueprints(context.Background(), charB.ID, "character", user.ID, []*models.CharacterBlueprint{
		{ItemID: 85002, TypeID: 791, LocationID: 60003760, Quantity: -1, Runs: -1},
	})
	assert.NoError(t, err)

	// Delete only charA's blueprints
	err = bpRepo.DeleteByOwner(context.Background(), charA.ID, "character")
	assert.NoError(t, err)

	// charA's type 790 should be gone, charB's type 791 should remain
	levels, err := bpRepo.GetBlueprintLevels(context.Background(), user.ID, []int64{790, 791})
	assert.NoError(t, err)
	assert.Nil(t, levels[790])
	assert.NotNil(t, levels[791])
}
