package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_HaulingRuns_CreateRun(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	userID := int64(9000)
	user := &repositories.User{ID: userID, Name: "Hauling Test User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	repo := repositories.NewHaulingRuns(db)

	fromRegion := int64(10000002)
	toRegion := int64(10000043)
	run := &models.HaulingRun{
		UserID:       userID,
		Name:         "Test Run",
		Status:       "PLANNING",
		FromRegionID: fromRegion,
		ToRegionID:   toRegion,
	}

	created, err := repo.CreateRun(context.Background(), run)
	assert.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, "Test Run", created.Name)
	assert.Equal(t, "PLANNING", created.Status)
	assert.Equal(t, fromRegion, created.FromRegionID)
	assert.Equal(t, toRegion, created.ToRegionID)
	assert.NotEmpty(t, created.CreatedAt)
	assert.NotEmpty(t, created.UpdatedAt)
}

func Test_HaulingRuns_CreateRunWithOptionalFields(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	userID := int64(9010)
	user := &repositories.User{ID: userID, Name: "Hauling Test User 2"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	repo := repositories.NewHaulingRuns(db)

	fromSystem := int64(30000142)
	maxVol := 350000.0
	threshold := 5000000.0
	notes := "My notes"

	run := &models.HaulingRun{
		UserID:           userID,
		Name:             "Optional Fields Run",
		Status:           "PLANNING",
		FromRegionID:     int64(10000002),
		FromSystemID:     &fromSystem,
		ToRegionID:       int64(10000043),
		MaxVolumeM3:      &maxVol,
		HaulThresholdISK: &threshold,
		NotifyTier2:      true,
		NotifyTier3:      false,
		DailyDigest:      true,
		Notes:            &notes,
	}

	created, err := repo.CreateRun(context.Background(), run)
	assert.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, &fromSystem, created.FromSystemID)
	assert.Equal(t, &maxVol, created.MaxVolumeM3)
	assert.Equal(t, &threshold, created.HaulThresholdISK)
	assert.True(t, created.NotifyTier2)
	assert.True(t, created.DailyDigest)
	assert.Equal(t, &notes, created.Notes)
}

func Test_HaulingRuns_GetRunByID(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	userID := int64(9020)
	user := &repositories.User{ID: userID, Name: "Hauling Test User 3"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	repo := repositories.NewHaulingRuns(db)

	run := &models.HaulingRun{
		UserID:       userID,
		Name:         "Get By ID Run",
		Status:       "PLANNING",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
	}
	created, err := repo.CreateRun(context.Background(), run)
	assert.NoError(t, err)

	// Get by correct user
	found, err := repo.GetRunByID(context.Background(), created.ID, userID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "Get By ID Run", found.Name)
	assert.NotNil(t, found.Items)

	// Get by wrong user returns nil
	notFound, err := repo.GetRunByID(context.Background(), created.ID, int64(99999))
	assert.NoError(t, err)
	assert.Nil(t, notFound)
}

func Test_HaulingRuns_GetRunByID_NotFound(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingRuns(db)

	found, err := repo.GetRunByID(context.Background(), int64(999999999), int64(1))
	assert.NoError(t, err)
	assert.Nil(t, found)
}

func Test_HaulingRuns_ListRunsByUser(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	userID := int64(9030)
	user := &repositories.User{ID: userID, Name: "Hauling Test User 4"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	repo := repositories.NewHaulingRuns(db)

	// Create two runs
	for _, name := range []string{"Run A", "Run B"} {
		_, err := repo.CreateRun(context.Background(), &models.HaulingRun{
			UserID:       userID,
			Name:         name,
			Status:       "PLANNING",
			FromRegionID: int64(10000002),
			ToRegionID:   int64(10000043),
		})
		assert.NoError(t, err)
	}

	runs, err := repo.ListRunsByUser(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, runs, 2)
	// Each run should have an empty items slice, not nil
	for _, r := range runs {
		assert.NotNil(t, r.Items)
	}
}

func Test_HaulingRuns_ListRunsByUser_Empty(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingRuns(db)

	runs, err := repo.ListRunsByUser(context.Background(), int64(9999999))
	assert.NoError(t, err)
	assert.NotNil(t, runs)
	assert.Len(t, runs, 0)
}

func Test_HaulingRuns_UpdateRun(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	userID := int64(9040)
	user := &repositories.User{ID: userID, Name: "Hauling Test User 5"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	repo := repositories.NewHaulingRuns(db)

	created, err := repo.CreateRun(context.Background(), &models.HaulingRun{
		UserID:       userID,
		Name:         "Before Update",
		Status:       "PLANNING",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
	})
	assert.NoError(t, err)

	// Update
	created.Name = "After Update"
	created.ToRegionID = int64(10000030)
	err = repo.UpdateRun(context.Background(), created)
	assert.NoError(t, err)

	// Verify
	found, err := repo.GetRunByID(context.Background(), created.ID, userID)
	assert.NoError(t, err)
	assert.Equal(t, "After Update", found.Name)
	assert.Equal(t, int64(10000030), found.ToRegionID)
}

func Test_HaulingRuns_UpdateRun_NotFound(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingRuns(db)

	err = repo.UpdateRun(context.Background(), &models.HaulingRun{
		ID:           int64(999999999),
		UserID:       int64(1),
		Name:         "Ghost Run",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func Test_HaulingRuns_UpdateRunStatus(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	userID := int64(9050)
	user := &repositories.User{ID: userID, Name: "Hauling Test User 6"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	repo := repositories.NewHaulingRuns(db)

	created, err := repo.CreateRun(context.Background(), &models.HaulingRun{
		UserID:       userID,
		Name:         "Status Test Run",
		Status:       "PLANNING",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
	})
	assert.NoError(t, err)

	err = repo.UpdateRunStatus(context.Background(), created.ID, userID, "ACCUMULATING")
	assert.NoError(t, err)

	found, err := repo.GetRunByID(context.Background(), created.ID, userID)
	assert.NoError(t, err)
	assert.Equal(t, "ACCUMULATING", found.Status)
}

func Test_HaulingRuns_UpdateRunStatus_NotFound(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingRuns(db)

	err = repo.UpdateRunStatus(context.Background(), int64(999999999), int64(1), "READY")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func Test_HaulingRuns_DeleteRun(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	userID := int64(9060)
	user := &repositories.User{ID: userID, Name: "Hauling Test User 7"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	repo := repositories.NewHaulingRuns(db)

	created, err := repo.CreateRun(context.Background(), &models.HaulingRun{
		UserID:       userID,
		Name:         "Delete Test Run",
		Status:       "PLANNING",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
	})
	assert.NoError(t, err)

	err = repo.DeleteRun(context.Background(), created.ID, userID)
	assert.NoError(t, err)

	found, err := repo.GetRunByID(context.Background(), created.ID, userID)
	assert.NoError(t, err)
	assert.Nil(t, found)
}

func Test_HaulingRuns_DeleteRun_NotFound(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingRuns(db)

	err = repo.DeleteRun(context.Background(), int64(999999999), int64(1))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func Test_HaulingRuns_ListAccumulatingByUser(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	userID := int64(9070)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "Accumulating Test User"})
	assert.NoError(t, err)

	repo := repositories.NewHaulingRuns(db)

	// Create runs with various statuses
	statuses := []string{"PLANNING", "ACCUMULATING", "ACCUMULATING", "READY"}
	for i, status := range statuses {
		_, err := repo.CreateRun(context.Background(), &models.HaulingRun{
			UserID:       userID,
			Name:         "Run " + string(rune('A'+i)),
			Status:       status,
			FromRegionID: int64(10000002),
			ToRegionID:   int64(10000043),
		})
		assert.NoError(t, err)
	}

	runs, err := repo.ListAccumulatingByUser(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, runs, 2)
	for _, r := range runs {
		assert.Equal(t, "ACCUMULATING", r.Status)
		assert.NotNil(t, r.Items)
	}
}

func Test_HaulingRuns_ListAccumulatingByUser_Empty(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingRuns(db)
	runs, err := repo.ListAccumulatingByUser(context.Background(), int64(99999999))
	assert.NoError(t, err)
	assert.NotNil(t, runs)
	assert.Len(t, runs, 0)
}

func Test_HaulingRuns_ListDigestRunsByUser(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	userID := int64(9080)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "Digest Test User"})
	assert.NoError(t, err)

	repo := repositories.NewHaulingRuns(db)

	// Create runs: digest=true active, digest=false, digest=true COMPLETE
	digestActive := true
	digestFalse := false

	_, err = repo.CreateRun(context.Background(), &models.HaulingRun{
		UserID:       userID,
		Name:         "Digest Active Run",
		Status:       "PLANNING",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
		DailyDigest:  digestActive,
	})
	assert.NoError(t, err)

	_, err = repo.CreateRun(context.Background(), &models.HaulingRun{
		UserID:       userID,
		Name:         "No Digest Run",
		Status:       "ACCUMULATING",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
		DailyDigest:  digestFalse,
	})
	assert.NoError(t, err)

	completeRun, err := repo.CreateRun(context.Background(), &models.HaulingRun{
		UserID:       userID,
		Name:         "Complete Digest Run",
		Status:       "PLANNING",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
		DailyDigest:  digestActive,
	})
	assert.NoError(t, err)
	// Mark it complete
	err = repo.UpdateRunStatus(context.Background(), completeRun.ID, userID, "COMPLETE")
	assert.NoError(t, err)

	runs, err := repo.ListDigestRunsByUser(context.Background(), userID)
	assert.NoError(t, err)
	// Should only return the PLANNING run with daily_digest=true
	assert.Len(t, runs, 1)
	assert.Equal(t, "Digest Active Run", runs[0].Name)
	assert.True(t, runs[0].DailyDigest)
}

func Test_HaulingRuns_ListDigestRunsByUser_ExcludesCancelled(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	userID := int64(9090)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "Digest Test User 2"})
	assert.NoError(t, err)

	repo := repositories.NewHaulingRuns(db)

	// Create a cancelled run with daily_digest=true
	cancelledRun, err := repo.CreateRun(context.Background(), &models.HaulingRun{
		UserID:       userID,
		Name:         "Cancelled Digest Run",
		Status:       "PLANNING",
		FromRegionID: int64(10000002),
		ToRegionID:   int64(10000043),
		DailyDigest:  true,
	})
	assert.NoError(t, err)
	err = repo.UpdateRunStatus(context.Background(), cancelledRun.ID, userID, "CANCELLED")
	assert.NoError(t, err)

	runs, err := repo.ListDigestRunsByUser(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, runs, 0)
}

func Test_HaulingRuns_ListSellingByUser(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	userID := int64(9095)
	err = userRepo.Add(context.Background(), &repositories.User{ID: userID, Name: "Selling Test User"})
	assert.NoError(t, err)

	repo := repositories.NewHaulingRuns(db)

	// Create runs with various statuses
	statuses := []string{"PLANNING", "SELLING", "SELLING", "COMPLETE"}
	for i, status := range statuses {
		_, err := repo.CreateRun(context.Background(), &models.HaulingRun{
			UserID:       userID,
			Name:         "Run " + string(rune('A'+i)),
			Status:       status,
			FromRegionID: int64(10000002),
			ToRegionID:   int64(10000043),
		})
		assert.NoError(t, err)
	}

	runs, err := repo.ListSellingByUser(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, runs, 2)
	for _, r := range runs {
		assert.Equal(t, "SELLING", r.Status)
		assert.NotNil(t, r.Items)
	}
}

func Test_HaulingRuns_ListSellingByUser_Empty(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	repo := repositories.NewHaulingRuns(db)
	runs, err := repo.ListSellingByUser(context.Background(), int64(99999998))
	assert.NoError(t, err)
	assert.NotNil(t, runs)
	assert.Len(t, runs, 0)
}
