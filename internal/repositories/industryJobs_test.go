package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_IndustryJobsShouldUpsertAndGetActive(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	jobsRepo := repositories.NewIndustryJobs(db)

	user := &repositories.User{ID: 6000, Name: "Jobs Test User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	cost := 1500000.0
	productTypeID := int64(34)
	solarSystemID := int64(30000142) // Jita
	now := time.Now().UTC().Truncate(time.Millisecond)

	jobs := []*models.IndustryJob{
		{
			JobID:               100001,
			InstallerID:         60001,
			FacilityID:          60003760,
			StationID:           60003760,
			ActivityID:          1,
			BlueprintID:         9876,
			BlueprintTypeID:     787,
			BlueprintLocationID: 60003760,
			OutputLocationID:    60003760,
			Runs:                10,
			Cost:                &cost,
			ProductTypeID:       &productTypeID,
			Status:              "active",
			Duration:            3600,
			StartDate:           now,
			EndDate:             now.Add(time.Hour),
			SolarSystemID:       &solarSystemID,
		},
	}

	err = jobsRepo.UpsertJobs(context.Background(), user.ID, jobs)
	assert.NoError(t, err)

	activeJobs, err := jobsRepo.GetActiveJobs(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, activeJobs, 1)
	assert.Equal(t, int64(100001), activeJobs[0].JobID)
	assert.Equal(t, 1, activeJobs[0].ActivityID)
	assert.Equal(t, "active", activeJobs[0].Status)
	assert.Equal(t, 10, activeJobs[0].Runs)
	assert.Equal(t, 1500000.0, *activeJobs[0].Cost)
	assert.Equal(t, "Manufacturing", activeJobs[0].ActivityName)
}

func Test_IndustryJobsShouldUpsertStatusChange(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	jobsRepo := repositories.NewIndustryJobs(db)

	user := &repositories.User{ID: 6010, Name: "Status Change User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	now := time.Now().UTC().Truncate(time.Millisecond)

	// Insert active job
	jobs := []*models.IndustryJob{
		{
			JobID:               100010,
			InstallerID:         60001,
			FacilityID:          60003760,
			StationID:           60003760,
			ActivityID:          9,
			BlueprintID:         1111,
			BlueprintTypeID:     46166,
			BlueprintLocationID: 60003760,
			OutputLocationID:    60003760,
			Runs:                100,
			Status:              "active",
			Duration:            7200,
			StartDate:           now,
			EndDate:             now.Add(2 * time.Hour),
		},
	}

	err = jobsRepo.UpsertJobs(context.Background(), user.ID, jobs)
	assert.NoError(t, err)

	// Mark as delivered
	completedDate := now.Add(2 * time.Hour)
	successfulRuns := 100
	jobs[0].Status = "delivered"
	jobs[0].CompletedDate = &completedDate
	jobs[0].SuccessfulRuns = &successfulRuns

	err = jobsRepo.UpsertJobs(context.Background(), user.ID, jobs)
	assert.NoError(t, err)

	// Active query should return empty
	activeJobs, err := jobsRepo.GetActiveJobs(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, activeJobs, 0)

	// All jobs should show delivered
	allJobs, err := jobsRepo.GetAllJobs(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, allJobs, 1)
	assert.Equal(t, "delivered", allJobs[0].Status)
	assert.Equal(t, "Reaction", allJobs[0].ActivityName)
}

func Test_IndustryJobsShouldGetJobByID(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	jobsRepo := repositories.NewIndustryJobs(db)

	user := &repositories.User{ID: 6020, Name: "GetByID User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	now := time.Now().UTC().Truncate(time.Millisecond)

	jobs := []*models.IndustryJob{
		{
			JobID:               100020,
			InstallerID:         60001,
			FacilityID:          60003760,
			StationID:           60003760,
			ActivityID:          8,
			BlueprintID:         2222,
			BlueprintTypeID:     999,
			BlueprintLocationID: 60003760,
			OutputLocationID:    60003760,
			Runs:                1,
			Status:              "active",
			Duration:            1800,
			StartDate:           now,
			EndDate:             now.Add(30 * time.Minute),
		},
	}

	err = jobsRepo.UpsertJobs(context.Background(), user.ID, jobs)
	assert.NoError(t, err)

	job, err := jobsRepo.GetJobByID(context.Background(), 100020)
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, int64(100020), job.JobID)
	assert.Equal(t, "Invention", job.ActivityName)

	// Non-existent job
	noJob, err := jobsRepo.GetJobByID(context.Background(), 999999)
	assert.NoError(t, err)
	assert.Nil(t, noJob)
}

func Test_IndustryJobsShouldDeleteOldDeliveredJobs(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	jobsRepo := repositories.NewIndustryJobs(db)

	user := &repositories.User{ID: 6030, Name: "Delete Old User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	now := time.Now().UTC().Truncate(time.Millisecond)

	jobs := []*models.IndustryJob{
		{
			JobID:               100030,
			InstallerID:         60001,
			FacilityID:          60003760,
			StationID:           60003760,
			ActivityID:          1,
			BlueprintID:         3333,
			BlueprintTypeID:     787,
			BlueprintLocationID: 60003760,
			OutputLocationID:    60003760,
			Runs:                5,
			Status:              "delivered",
			Duration:            3600,
			StartDate:           now.Add(-48 * time.Hour),
			EndDate:             now.Add(-47 * time.Hour),
		},
		{
			JobID:               100031,
			InstallerID:         60001,
			FacilityID:          60003760,
			StationID:           60003760,
			ActivityID:          1,
			BlueprintID:         4444,
			BlueprintTypeID:     787,
			BlueprintLocationID: 60003760,
			OutputLocationID:    60003760,
			Runs:                10,
			Status:              "active",
			Duration:            3600,
			StartDate:           now,
			EndDate:             now.Add(time.Hour),
		},
	}

	err = jobsRepo.UpsertJobs(context.Background(), user.ID, jobs)
	assert.NoError(t, err)

	deleted, err := jobsRepo.DeleteOldDeliveredJobs(context.Background(), user.ID, now.Add(time.Minute))
	assert.NoError(t, err)
	assert.Equal(t, int64(1), deleted)

	// Active job should remain
	allJobs, err := jobsRepo.GetAllJobs(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, allJobs, 1)
	assert.Equal(t, int64(100031), allJobs[0].JobID)
}

func Test_IndustryJobsShouldHandleEmptyUpsert(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	jobsRepo := repositories.NewIndustryJobs(db)

	err = jobsRepo.UpsertJobs(context.Background(), 99999, []*models.IndustryJob{})
	assert.NoError(t, err)
}

func Test_IndustryJobsShouldGetActiveJobsForMatching(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	jobsRepo := repositories.NewIndustryJobs(db)

	user := &repositories.User{ID: 6040, Name: "Matching User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	now := time.Now().UTC().Truncate(time.Millisecond)

	jobs := []*models.IndustryJob{
		{
			JobID: 100040, InstallerID: 60001, FacilityID: 60003760, StationID: 60003760,
			ActivityID: 1, BlueprintID: 5555, BlueprintTypeID: 787,
			BlueprintLocationID: 60003760, OutputLocationID: 60003760,
			Runs: 10, Status: "active", Duration: 3600,
			StartDate: now, EndDate: now.Add(time.Hour),
		},
		{
			JobID: 100041, InstallerID: 60001, FacilityID: 60003760, StationID: 60003760,
			ActivityID: 1, BlueprintID: 6666, BlueprintTypeID: 787,
			BlueprintLocationID: 60003760, OutputLocationID: 60003760,
			Runs: 5, Status: "delivered", Duration: 3600,
			StartDate: now.Add(-2 * time.Hour), EndDate: now.Add(-time.Hour),
		},
	}

	err = jobsRepo.UpsertJobs(context.Background(), user.ID, jobs)
	assert.NoError(t, err)

	activeJobs, err := jobsRepo.GetActiveJobsForMatching(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, activeJobs, 1)
	assert.Equal(t, int64(100040), activeJobs[0].JobID)
}

func Test_IndustryJobsShouldMapActivityNames(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	jobsRepo := repositories.NewIndustryJobs(db)

	user := &repositories.User{ID: 6050, Name: "Activity Names User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	now := time.Now().UTC().Truncate(time.Millisecond)

	jobs := []*models.IndustryJob{
		{JobID: 100050, InstallerID: 60001, FacilityID: 60003760, StationID: 60003760, ActivityID: 1, BlueprintID: 1, BlueprintTypeID: 787, BlueprintLocationID: 60003760, OutputLocationID: 60003760, Runs: 1, Status: "active", Duration: 3600, StartDate: now, EndDate: now.Add(time.Hour)},
		{JobID: 100051, InstallerID: 60001, FacilityID: 60003760, StationID: 60003760, ActivityID: 3, BlueprintID: 2, BlueprintTypeID: 787, BlueprintLocationID: 60003760, OutputLocationID: 60003760, Runs: 1, Status: "active", Duration: 3600, StartDate: now, EndDate: now.Add(time.Hour)},
		{JobID: 100052, InstallerID: 60001, FacilityID: 60003760, StationID: 60003760, ActivityID: 4, BlueprintID: 3, BlueprintTypeID: 787, BlueprintLocationID: 60003760, OutputLocationID: 60003760, Runs: 1, Status: "active", Duration: 3600, StartDate: now, EndDate: now.Add(time.Hour)},
		{JobID: 100053, InstallerID: 60001, FacilityID: 60003760, StationID: 60003760, ActivityID: 5, BlueprintID: 4, BlueprintTypeID: 787, BlueprintLocationID: 60003760, OutputLocationID: 60003760, Runs: 1, Status: "active", Duration: 3600, StartDate: now, EndDate: now.Add(time.Hour)},
		{JobID: 100054, InstallerID: 60001, FacilityID: 60003760, StationID: 60003760, ActivityID: 8, BlueprintID: 5, BlueprintTypeID: 787, BlueprintLocationID: 60003760, OutputLocationID: 60003760, Runs: 1, Status: "active", Duration: 3600, StartDate: now, EndDate: now.Add(time.Hour)},
		{JobID: 100055, InstallerID: 60001, FacilityID: 60003760, StationID: 60003760, ActivityID: 9, BlueprintID: 6, BlueprintTypeID: 787, BlueprintLocationID: 60003760, OutputLocationID: 60003760, Runs: 1, Status: "active", Duration: 3600, StartDate: now, EndDate: now.Add(time.Hour)},
	}

	err = jobsRepo.UpsertJobs(context.Background(), user.ID, jobs)
	assert.NoError(t, err)

	allJobs, err := jobsRepo.GetActiveJobs(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, allJobs, 6)

	nameMap := map[int64]string{}
	for _, j := range allJobs {
		nameMap[j.JobID] = j.ActivityName
	}

	assert.Equal(t, "Manufacturing", nameMap[100050])
	assert.Equal(t, "TE Research", nameMap[100051])
	assert.Equal(t, "ME Research", nameMap[100052])
	assert.Equal(t, "Copying", nameMap[100053])
	assert.Equal(t, "Invention", nameMap[100054])
	assert.Equal(t, "Reaction", nameMap[100055])
}
