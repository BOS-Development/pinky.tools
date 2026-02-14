package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_ItemTypeShouldUpsertItemTypes(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	itemTypeRepo := repositories.NewItemTypeRepository(db)

	iconID1 := int64(1234)
	iconID2 := int64(5678)

	itemTypes := []models.EveInventoryType{
		{TypeID: 34, TypeName: "Tritanium", Volume: 0.01, IconID: &iconID1},
		{TypeID: 35, TypeName: "Pyerite", Volume: 0.0032, IconID: &iconID2},
		{TypeID: 36, TypeName: "Mexallon", Volume: 0.01, IconID: nil},
	}

	err = itemTypeRepo.UpsertItemTypes(context.Background(), itemTypes)
	assert.NoError(t, err)

	// Update item type
	itemTypes[0].Volume = 0.015

	// Add new item type
	iconID3 := int64(9999)
	itemTypes = append(itemTypes, models.EveInventoryType{
		TypeID:   37,
		TypeName: "Isogen",
		Volume:   0.02,
		IconID:   &iconID3,
	})

	err = itemTypeRepo.UpsertItemTypes(context.Background(), itemTypes)
	assert.NoError(t, err)
}

func Test_ItemTypeShouldHandleNullIconID(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	itemTypeRepo := repositories.NewItemTypeRepository(db)

	// Item types with NULL icon IDs
	itemTypes := []models.EveInventoryType{
		{TypeID: 100, TypeName: "Test Item 1", Volume: 1.0, IconID: nil},
		{TypeID: 101, TypeName: "Test Item 2", Volume: 2.0, IconID: nil},
	}

	err = itemTypeRepo.UpsertItemTypes(context.Background(), itemTypes)
	assert.NoError(t, err)
}

func Test_ItemTypeShouldHandleEmptyUpsert(t *testing.T) {
	db, err := setupDatabase()
	assert.NoError(t, err)

	itemTypeRepo := repositories.NewItemTypeRepository(db)

	err = itemTypeRepo.UpsertItemTypes(context.Background(), []models.EveInventoryType{})
	assert.NoError(t, err)

	err = itemTypeRepo.UpsertItemTypes(context.Background(), nil)
	assert.NoError(t, err)
}
