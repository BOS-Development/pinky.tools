package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_PlayerCorporationsShouldUpsertAndGet(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepository := repositories.NewUserRepository(db)
	repository := repositories.NewPlayerCorporations(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Ibn Kabab",
	}

	err = userRepository.Add(context.Background(), testUser)
	assert.NoError(t, err)

	testCorp := repositories.PlayerCorporation{
		ID:              1001,
		UserID:          42,
		Name:            "Test Corporation",
		EsiToken:        "token123",
		EsiRefreshToken: "refresh456",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}

	err = repository.Upsert(context.Background(), testCorp)
	assert.NoError(t, err)

	corps, err := repository.Get(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.Len(t, corps, 1)
	assert.Equal(t, testCorp.ID, corps[0].ID)
	assert.Equal(t, testCorp.Name, corps[0].Name)
	assert.Equal(t, testCorp.UserID, corps[0].UserID)
	assert.Equal(t, testCorp.EsiToken, corps[0].EsiToken)
	assert.Equal(t, testCorp.EsiRefreshToken, corps[0].EsiRefreshToken)
}

func Test_PlayerCorporationsShouldUpdateExistingCorporation(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepository := repositories.NewUserRepository(db)
	repository := repositories.NewPlayerCorporations(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Ibn Kabab",
	}

	err = userRepository.Add(context.Background(), testUser)
	assert.NoError(t, err)

	testCorp := repositories.PlayerCorporation{
		ID:              1001,
		UserID:          42,
		Name:            "Original Name",
		EsiToken:        "token123",
		EsiRefreshToken: "refresh456",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}

	err = repository.Upsert(context.Background(), testCorp)
	assert.NoError(t, err)

	updatedCorp := repositories.PlayerCorporation{
		ID:              1001,
		UserID:          42,
		Name:            "Updated Name",
		EsiToken:        "newtoken789",
		EsiRefreshToken: "newrefresh012",
		EsiExpiresOn:    time.Now().Add(2 * time.Hour),
	}

	err = repository.Upsert(context.Background(), updatedCorp)
	assert.NoError(t, err)

	corps, err := repository.Get(context.Background(), testUser.ID)
	assert.NoError(t, err)
	assert.Len(t, corps, 1)
	assert.Equal(t, updatedCorp.Name, corps[0].Name)
	assert.Equal(t, updatedCorp.EsiToken, corps[0].EsiToken)
	assert.Equal(t, updatedCorp.EsiRefreshToken, corps[0].EsiRefreshToken)
}

func Test_PlayerCorporationsShouldUpsertDivisions(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepository := repositories.NewUserRepository(db)
	repository := repositories.NewPlayerCorporations(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Ibn Kabab",
	}

	err = userRepository.Add(context.Background(), testUser)
	assert.NoError(t, err)

	testCorp := repositories.PlayerCorporation{
		ID:              1001,
		UserID:          42,
		Name:            "Test Corporation",
		EsiToken:        "token123",
		EsiRefreshToken: "refresh456",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}

	err = repository.Upsert(context.Background(), testCorp)
	assert.NoError(t, err)

	divisions := &models.CorporationDivisions{
		Hanger: map[int]string{
			1: "Main Hangar",
			2: "Secondary Hangar",
		},
		Wallet: map[int]string{
			1: "Main Wallet",
			2: "Secondary Wallet",
		},
	}

	err = repository.UpsertDivisions(context.Background(), testCorp.ID, testUser.ID, divisions)
	assert.NoError(t, err)

	retrievedDivisions, err := repository.GetDivisions(context.Background(), testCorp.ID, testUser.ID)
	assert.NoError(t, err)
	assert.Equal(t, divisions.Hanger, retrievedDivisions.Hanger)
	assert.Equal(t, divisions.Wallet, retrievedDivisions.Wallet)
}

func Test_PlayerCorporationsShouldUpdateDivisions(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepository := repositories.NewUserRepository(db)
	repository := repositories.NewPlayerCorporations(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Ibn Kabab",
	}

	err = userRepository.Add(context.Background(), testUser)
	assert.NoError(t, err)

	testCorp := repositories.PlayerCorporation{
		ID:              1001,
		UserID:          42,
		Name:            "Test Corporation",
		EsiToken:        "token123",
		EsiRefreshToken: "refresh456",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}

	err = repository.Upsert(context.Background(), testCorp)
	assert.NoError(t, err)

	firstDivisions := &models.CorporationDivisions{
		Hanger: map[int]string{
			1: "Original Hangar",
		},
		Wallet: map[int]string{
			1: "Original Wallet",
		},
	}

	err = repository.UpsertDivisions(context.Background(), testCorp.ID, testUser.ID, firstDivisions)
	assert.NoError(t, err)

	retrievedFirst, err := repository.GetDivisions(context.Background(), testCorp.ID, testUser.ID)
	assert.NoError(t, err)
	assert.Equal(t, firstDivisions.Hanger, retrievedFirst.Hanger)
	assert.Equal(t, firstDivisions.Wallet, retrievedFirst.Wallet)

	updatedDivisions := &models.CorporationDivisions{
		Hanger: map[int]string{
			1: "Updated Hangar",
			2: "New Hangar",
		},
		Wallet: map[int]string{
			1: "Updated Wallet",
		},
	}

	err = repository.UpsertDivisions(context.Background(), testCorp.ID, testUser.ID, updatedDivisions)
	assert.NoError(t, err)

	retrievedUpdated, err := repository.GetDivisions(context.Background(), testCorp.ID, testUser.ID)
	assert.NoError(t, err)
	assert.Equal(t, updatedDivisions.Hanger, retrievedUpdated.Hanger)
	assert.Equal(t, updatedDivisions.Wallet, retrievedUpdated.Wallet)
}

func Test_PlayerCorporationsShouldHandleNilDivisions(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepository := repositories.NewUserRepository(db)
	repository := repositories.NewPlayerCorporations(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Ibn Kabab",
	}

	err = userRepository.Add(context.Background(), testUser)
	assert.NoError(t, err)

	testCorp := repositories.PlayerCorporation{
		ID:              1001,
		UserID:          42,
		Name:            "Test Corporation",
		EsiToken:        "token123",
		EsiRefreshToken: "refresh456",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}

	err = repository.Upsert(context.Background(), testCorp)
	assert.NoError(t, err)

	err = repository.UpsertDivisions(context.Background(), testCorp.ID, testUser.ID, nil)
	assert.NoError(t, err)

	retrievedDivisions, err := repository.GetDivisions(context.Background(), testCorp.ID, testUser.ID)
	assert.NoError(t, err)
	assert.Empty(t, retrievedDivisions.Hanger)
	assert.Empty(t, retrievedDivisions.Wallet)
}

func Test_PlayerCorporationsShouldHandleEmptyDivisions(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepository := repositories.NewUserRepository(db)
	repository := repositories.NewPlayerCorporations(db)

	testUser := &repositories.User{
		ID:   42,
		Name: "Ibn Kabab",
	}

	err = userRepository.Add(context.Background(), testUser)
	assert.NoError(t, err)

	testCorp := repositories.PlayerCorporation{
		ID:              1001,
		UserID:          42,
		Name:            "Test Corporation",
		EsiToken:        "token123",
		EsiRefreshToken: "refresh456",
		EsiExpiresOn:    time.Now().Add(time.Hour),
	}

	err = repository.Upsert(context.Background(), testCorp)
	assert.NoError(t, err)

	emptyDivisions := &models.CorporationDivisions{
		Hanger: map[int]string{},
		Wallet: map[int]string{},
	}

	err = repository.UpsertDivisions(context.Background(), testCorp.ID, testUser.ID, emptyDivisions)
	assert.NoError(t, err)

	retrievedDivisions, err := repository.GetDivisions(context.Background(), testCorp.ID, testUser.ID)
	assert.NoError(t, err)
	assert.Empty(t, retrievedDivisions.Hanger)
	assert.Empty(t, retrievedDivisions.Wallet)
}
