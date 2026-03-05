package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

// userIDs 9700–9790 reserved for userTradingStructures tests

func Test_UserTradingStructures_Upsert_Creates(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	userID := int64(9700)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "Structures Test User 1"})
	assert.NoError(t, err)

	repo := repositories.NewUserTradingStructures(db)

	s := &models.UserTradingStructure{
		UserID:      userID,
		StructureID: int64(1000000000001),
		Name:        "Test Structure Alpha",
		SystemID:    int64(30000142),
		RegionID:    int64(10000002),
		CharacterID: int64(200001001),
		AccessOK:    true,
	}

	created, err := repo.Upsert(context.Background(), s)
	assert.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, int64(1000000000001), created.StructureID)
	assert.Equal(t, "Test Structure Alpha", created.Name)
	assert.Equal(t, true, created.AccessOK)
	assert.NotEmpty(t, created.CreatedAt)
}

func Test_UserTradingStructures_Upsert_UpdatesExisting(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	userID := int64(9710)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "Structures Test User 2"})
	assert.NoError(t, err)

	repo := repositories.NewUserTradingStructures(db)

	s := &models.UserTradingStructure{
		UserID:      userID,
		StructureID: int64(1000000000002),
		Name:        "Original Name",
		SystemID:    int64(30000142),
		RegionID:    int64(10000002),
		CharacterID: int64(200001002),
		AccessOK:    true,
	}

	first, err := repo.Upsert(context.Background(), s)
	assert.NoError(t, err)

	// Update the name and access status
	s.Name = "Updated Name"
	s.AccessOK = false
	second, err := repo.Upsert(context.Background(), s)
	assert.NoError(t, err)

	// ID should remain the same after upsert
	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, "Updated Name", second.Name)
	assert.Equal(t, false, second.AccessOK)
}

func Test_UserTradingStructures_List_ReturnsUserStructures(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	userID := int64(9720)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "Structures Test User 3"})
	assert.NoError(t, err)

	repo := repositories.NewUserTradingStructures(db)

	// Create two structures with distinct structure IDs
	structureDefs := []struct {
		id   int64
		name string
	}{
		{int64(1000000000011), "Structure B"},
		{int64(1000000000012), "Structure A"},
	}
	for _, def := range structureDefs {
		_, err := repo.Upsert(context.Background(), &models.UserTradingStructure{
			UserID:      userID,
			StructureID: def.id,
			Name:        def.name,
			SystemID:    int64(30000142),
			RegionID:    int64(10000002),
			CharacterID: int64(200001003),
			AccessOK:    true,
		})
		assert.NoError(t, err)
	}

	structures, err := repo.List(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, structures, 2)
	// Should be ordered by name ASC: "Structure A" before "Structure B"
	assert.Equal(t, "Structure A", structures[0].Name)
	assert.Equal(t, "Structure B", structures[1].Name)
}

func Test_UserTradingStructures_List_EmptyForUnknownUser(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewUserTradingStructures(db)

	structures, err := repo.List(context.Background(), int64(99999999))
	assert.NoError(t, err)
	assert.NotNil(t, structures)
	assert.Len(t, structures, 0)
}

func Test_UserTradingStructures_Delete_RemovesRecord(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	userID := int64(9730)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "Structures Test User 4"})
	assert.NoError(t, err)

	repo := repositories.NewUserTradingStructures(db)

	created, err := repo.Upsert(context.Background(), &models.UserTradingStructure{
		UserID:      userID,
		StructureID: int64(1000000000020),
		Name:        "To Be Deleted",
		SystemID:    int64(30000142),
		RegionID:    int64(10000002),
		CharacterID: int64(200001004),
		AccessOK:    true,
	})
	assert.NoError(t, err)

	err = repo.Delete(context.Background(), created.ID, userID)
	assert.NoError(t, err)

	// Should be gone
	structures, err := repo.List(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, structures, 0)
}

func Test_UserTradingStructures_Delete_NotFoundError(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewUserTradingStructures(db)

	err = repo.Delete(context.Background(), int64(999999), int64(9730))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func Test_UserTradingStructures_UpdateAccessStatus(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	userID := int64(9740)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "Structures Test User 5"})
	assert.NoError(t, err)

	repo := repositories.NewUserTradingStructures(db)

	structureID := int64(1000000000030)
	_, err = repo.Upsert(context.Background(), &models.UserTradingStructure{
		UserID:      userID,
		StructureID: structureID,
		Name:        "Access Test Structure",
		SystemID:    int64(30000142),
		RegionID:    int64(10000002),
		CharacterID: int64(200001005),
		AccessOK:    true,
	})
	assert.NoError(t, err)

	// Mark access as denied
	err = repo.UpdateAccessStatus(context.Background(), userID, structureID, false)
	assert.NoError(t, err)

	// Verify via List
	structures, err := repo.List(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, structures, 1)
	assert.Equal(t, false, structures[0].AccessOK)
}
