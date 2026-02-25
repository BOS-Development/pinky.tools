package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_TransportTriggerConfigShouldUpsertAndGet(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	configRepo := repositories.NewTransportTriggerConfig(db)

	user := &repositories.User{ID: 8300, Name: "Trigger Config User"}
	err = userRepo.Add(context.Background(), user)
	require.NoError(t, err)

	config := &models.TransportTriggerConfig{
		UserID:              user.ID,
		TriggerType:         "plan_generation",
		DefaultFulfillment:  "courier_contract",
		AllowedFulfillments: []string{"self_haul", "courier_contract"},
		CourierRatePerM3:    500,
		CourierCollateralRate: 0.02,
	}

	created, err := configRepo.Upsert(context.Background(), config)
	require.NoError(t, err)
	assert.Equal(t, "plan_generation", created.TriggerType)
	assert.Equal(t, "courier_contract", created.DefaultFulfillment)
	assert.Equal(t, []string{"self_haul", "courier_contract"}, created.AllowedFulfillments)
	assert.Equal(t, 500.0, created.CourierRatePerM3)
	assert.Equal(t, 0.02, created.CourierCollateralRate)

	// Get by user
	configs, err := configRepo.GetByUser(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, configs, 1)

	// Upsert again to update
	config.DefaultFulfillment = "self_haul"
	config.CourierRatePerM3 = 750
	updated, err := configRepo.Upsert(context.Background(), config)
	require.NoError(t, err)
	assert.Equal(t, "self_haul", updated.DefaultFulfillment)
	assert.Equal(t, 750.0, updated.CourierRatePerM3)

	// Still only one config
	configs, err = configRepo.GetByUser(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, configs, 1)
}

func Test_TransportTriggerConfigMultipleTriggers(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	configRepo := repositories.NewTransportTriggerConfig(db)

	user := &repositories.User{ID: 8301, Name: "Multi Trigger User"}
	err = userRepo.Add(context.Background(), user)
	require.NoError(t, err)

	// Create configs for different triggers
	_, err = configRepo.Upsert(context.Background(), &models.TransportTriggerConfig{
		UserID:              user.ID,
		TriggerType:         "plan_generation",
		DefaultFulfillment:  "courier_contract",
		AllowedFulfillments: []string{"courier_contract"},
	})
	require.NoError(t, err)

	_, err = configRepo.Upsert(context.Background(), &models.TransportTriggerConfig{
		UserID:              user.ID,
		TriggerType:         "manual",
		DefaultFulfillment:  "self_haul",
		AllowedFulfillments: []string{"self_haul", "courier_contract", "contact_haul"},
	})
	require.NoError(t, err)

	configs, err := configRepo.GetByUser(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, configs, 2)
}

func Test_TransportTriggerConfigShouldReturnEmpty(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	configRepo := repositories.NewTransportTriggerConfig(db)

	configs, err := configRepo.GetByUser(context.Background(), 99999)
	require.NoError(t, err)
	assert.Len(t, configs, 0)
}
