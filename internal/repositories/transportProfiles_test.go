package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_TransportProfilesShouldCreateAndGetByUser(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	profilesRepo := repositories.NewTransportProfiles(db)

	user := &repositories.User{ID: 8000, Name: "Transport Profile User"}
	err = userRepo.Add(context.Background(), user)
	require.NoError(t, err)

	profile := &models.TransportProfile{
		UserID:               user.ID,
		Name:                 "Charon Main",
		TransportMethod:      "freighter",
		CargoM3:              435000,
		RatePerM3PerJump:     2.5,
		CollateralRate:       0.02,
		CollateralPriceBasis: "sell",
		RoutePreference:      "secure",
		IsDefault:            true,
	}

	created, err := profilesRepo.Create(context.Background(), profile)
	require.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, "Charon Main", created.Name)
	assert.Equal(t, "freighter", created.TransportMethod)
	assert.Equal(t, 435000.0, created.CargoM3)
	assert.Equal(t, 2.5, created.RatePerM3PerJump)
	assert.Equal(t, 0.02, created.CollateralRate)
	assert.Equal(t, "sell", created.CollateralPriceBasis)
	assert.Equal(t, "secure", created.RoutePreference)
	assert.True(t, created.IsDefault)

	profiles, err := profilesRepo.GetByUser(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, profiles, 1)
	assert.Equal(t, created.ID, profiles[0].ID)
}

func Test_TransportProfilesShouldGetByID(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	profilesRepo := repositories.NewTransportProfiles(db)

	user := &repositories.User{ID: 8001, Name: "Profile GetByID User"}
	err = userRepo.Add(context.Background(), user)
	require.NoError(t, err)

	fuelPerLY := 500.0
	fuelTypeID := int64(16274) // Nitrogen Isotopes

	profile := &models.TransportProfile{
		UserID:                user.ID,
		Name:                  "Rhea Alt - JFC 5",
		TransportMethod:       "jump_freighter",
		CargoM3:               34246,
		CollateralRate:        0.03,
		CollateralPriceBasis:  "split",
		FuelTypeID:            &fuelTypeID,
		FuelPerLY:             &fuelPerLY,
		FuelConservationLevel: 5,
		RoutePreference:       "shortest",
		IsDefault:             false,
	}

	created, err := profilesRepo.Create(context.Background(), profile)
	require.NoError(t, err)

	fetched, err := profilesRepo.GetByID(context.Background(), created.ID, user.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, "Rhea Alt - JFC 5", fetched.Name)
	assert.Equal(t, "jump_freighter", fetched.TransportMethod)
	assert.Equal(t, 34246.0, fetched.CargoM3)
	assert.Equal(t, &fuelTypeID, fetched.FuelTypeID)
	assert.Equal(t, &fuelPerLY, fetched.FuelPerLY)
	assert.Equal(t, 5, fetched.FuelConservationLevel)

	// Should not find with wrong user ID
	notFound, err := profilesRepo.GetByID(context.Background(), created.ID, 99999)
	require.NoError(t, err)
	assert.Nil(t, notFound)
}

func Test_TransportProfilesShouldUpdateAndClearDefaults(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	profilesRepo := repositories.NewTransportProfiles(db)

	user := &repositories.User{ID: 8002, Name: "Profile Update User"}
	err = userRepo.Add(context.Background(), user)
	require.NoError(t, err)

	// Create first profile as default
	p1, err := profilesRepo.Create(context.Background(), &models.TransportProfile{
		UserID:               user.ID,
		Name:                 "Charon #1",
		TransportMethod:      "freighter",
		CargoM3:              435000,
		CollateralPriceBasis: "sell",
		RoutePreference:      "shortest",
		IsDefault:            true,
	})
	require.NoError(t, err)
	assert.True(t, p1.IsDefault)

	// Create second profile as default — should clear first
	p2, err := profilesRepo.Create(context.Background(), &models.TransportProfile{
		UserID:               user.ID,
		Name:                 "Charon #2",
		TransportMethod:      "freighter",
		CargoM3:              500000,
		CollateralPriceBasis: "sell",
		RoutePreference:      "shortest",
		IsDefault:            true,
	})
	require.NoError(t, err)
	assert.True(t, p2.IsDefault)

	// First profile should no longer be default
	p1After, err := profilesRepo.GetByID(context.Background(), p1.ID, user.ID)
	require.NoError(t, err)
	assert.False(t, p1After.IsDefault)

	// Update first profile name
	p1After.Name = "Charon #1 Updated"
	updated, err := profilesRepo.Update(context.Background(), p1After)
	require.NoError(t, err)
	assert.Equal(t, "Charon #1 Updated", updated.Name)
}

func Test_TransportProfilesShouldGetDefaultByMethod(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	profilesRepo := repositories.NewTransportProfiles(db)

	user := &repositories.User{ID: 8003, Name: "Default Method User"}
	err = userRepo.Add(context.Background(), user)
	require.NoError(t, err)

	_, err = profilesRepo.Create(context.Background(), &models.TransportProfile{
		UserID:               user.ID,
		Name:                 "My Freighter",
		TransportMethod:      "freighter",
		CargoM3:              435000,
		CollateralPriceBasis: "sell",
		RoutePreference:      "shortest",
		IsDefault:            true,
	})
	require.NoError(t, err)

	def, err := profilesRepo.GetDefaultByMethod(context.Background(), user.ID, "freighter")
	require.NoError(t, err)
	require.NotNil(t, def)
	assert.Equal(t, "My Freighter", def.Name)

	// No default for jump_freighter
	noDef, err := profilesRepo.GetDefaultByMethod(context.Background(), user.ID, "jump_freighter")
	require.NoError(t, err)
	assert.Nil(t, noDef)
}

func Test_TransportProfilesShouldDelete(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	profilesRepo := repositories.NewTransportProfiles(db)

	user := &repositories.User{ID: 8004, Name: "Profile Delete User"}
	err = userRepo.Add(context.Background(), user)
	require.NoError(t, err)

	created, err := profilesRepo.Create(context.Background(), &models.TransportProfile{
		UserID:               user.ID,
		Name:                 "Disposable",
		TransportMethod:      "blockade_runner",
		CargoM3:              10000,
		CollateralPriceBasis: "sell",
		RoutePreference:      "shortest",
		IsDefault:            false,
	})
	require.NoError(t, err)

	err = profilesRepo.Delete(context.Background(), created.ID, user.ID)
	require.NoError(t, err)

	profiles, err := profilesRepo.GetByUser(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, profiles, 0)

	// Delete again should fail
	err = profilesRepo.Delete(context.Background(), created.ID, user.ID)
	assert.Error(t, err)
}

func Test_TransportProfilesShouldReturnEmptySlice(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	profilesRepo := repositories.NewTransportProfiles(db)

	profiles, err := profilesRepo.GetByUser(context.Background(), 99999)
	require.NoError(t, err)
	assert.Len(t, profiles, 0)
}
