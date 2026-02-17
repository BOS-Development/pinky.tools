package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_CharacterAssetsShouldUpdateData(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepository := repositories.NewUserRepository(db)
	characterRepository := repositories.NewCharacterRepository(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Ibn Kabab",
	}
	testCharacter := &repositories.Character{
		ID:     1337,
		Name:   "Crushim deez nuts",
		UserID: 42,
	}

	err = userRepository.Add(context.Background(), testUser)
	assert.NoError(t, err)

	err = characterRepository.Add(context.Background(), testCharacter)
	assert.NoError(t, err)

	repository := repositories.NewCharacterAssets(db)

	firstInventory := []*models.EveAsset{
		{
			ItemID:          1001,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      123,
			LocationType:    "OVER",
			Quantity:        551,
			TypeID:          56,
			LocationFlag:    "IDK",
		},
	}

	err = repository.UpdateAssets(context.Background(), 1337, 42, firstInventory)
	assert.NoError(t, err)

	assets, err := repository.Get(context.Background(), testUser.ID, testCharacter.ID)
	assert.NoError(t, err)

	expectedAssets := []*repositories.CharacterAsset{
		{
			CharacterID:     1337,
			UserID:          42,
			ItemID:          1001,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      123,
			LocationType:    "OVER",
			Quantity:        551,
			TypeID:          56,
			LocationFlag:    "IDK",
		},
	}

	assert.Equal(t, expectedAssets, assets)

	newInventory := []*models.EveAsset{
		{
			ItemID:          1002,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      123,
			LocationType:    "OVER",
			Quantity:        551,
			TypeID:          56,
			LocationFlag:    "IDK",
		},
		{
			ItemID:          1003,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      321,
			LocationType:    "UNDER",
			Quantity:        2,
			TypeID:          32,
			LocationFlag:    "SOMEWHERE",
		},
	}

	err = repository.UpdateAssets(context.Background(), 1337, 42, newInventory)
	assert.NoError(t, err)

	assets, err = repository.Get(context.Background(), testUser.ID, testCharacter.ID)
	assert.NoError(t, err)

	expectedUpdatedAssets := []*repositories.CharacterAsset{
		{
			CharacterID:     1337,
			UserID:          42,
			ItemID:          1002,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      123,
			LocationType:    "OVER",
			Quantity:        551,
			TypeID:          56,
			LocationFlag:    "IDK",
		},
		{
			CharacterID:     1337,
			UserID:          42,
			ItemID:          1003,
			IsBlueprintCopy: false,
			IsSingleton:     false,
			LocationID:      321,
			LocationType:    "UNDER",
			Quantity:        2,
			TypeID:          32,
			LocationFlag:    "SOMEWHERE",
		},
	}

	assert.Equal(t, expectedUpdatedAssets, assets)
}
