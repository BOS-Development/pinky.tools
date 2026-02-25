package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_JobQueueShouldCreateAndGetByUser(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	queueRepo := repositories.NewJobQueue(db)

	user := &repositories.User{ID: 7000, Name: "Queue Test User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	productTypeID := int64(34)
	estimatedCost := 5000000.0
	estimatedDuration := 3600
	notes := "Test job"
	systemID := int64(30000142)

	entry := &models.IndustryJobQueueEntry{
		UserID:            user.ID,
		BlueprintTypeID:   787,
		Activity:          "manufacturing",
		Runs:              10,
		MELevel:           10,
		TELevel:           20,
		SystemID:          &systemID,
		FacilityTax:       5.0,
		ProductTypeID:     &productTypeID,
		EstimatedCost:     &estimatedCost,
		EstimatedDuration: &estimatedDuration,
		Notes:             &notes,
	}

	created, err := queueRepo.Create(context.Background(), entry)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, user.ID, created.UserID)
	assert.Equal(t, int64(787), created.BlueprintTypeID)
	assert.Equal(t, "manufacturing", created.Activity)
	assert.Equal(t, 10, created.Runs)
	assert.Equal(t, 10, created.MELevel)
	assert.Equal(t, 20, created.TELevel)
	assert.Equal(t, "planned", created.Status)
	assert.Nil(t, created.EsiJobID)
	assert.NotZero(t, created.ID)

	entries, err := queueRepo.GetByUser(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, created.ID, entries[0].ID)
}

func Test_JobQueueShouldUpdatePlannedEntry(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	queueRepo := repositories.NewJobQueue(db)

	user := &repositories.User{ID: 7010, Name: "Update Queue User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	entry := &models.IndustryJobQueueEntry{
		UserID:          user.ID,
		BlueprintTypeID: 787,
		Activity:        "manufacturing",
		Runs:            10,
		MELevel:         0,
		TELevel:         0,
		FacilityTax:     0,
	}

	created, err := queueRepo.Create(context.Background(), entry)
	assert.NoError(t, err)

	// Update the entry
	updateEntry := &models.IndustryJobQueueEntry{
		BlueprintTypeID: 787,
		Activity:        "manufacturing",
		Runs:            20,
		MELevel:         10,
		TELevel:         20,
		FacilityTax:     3.5,
	}

	updated, err := queueRepo.Update(context.Background(), created.ID, user.ID, updateEntry)
	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, 20, updated.Runs)
	assert.Equal(t, 10, updated.MELevel)
	assert.Equal(t, 20, updated.TELevel)
	assert.Equal(t, 3.5, updated.FacilityTax)
}

func Test_JobQueueShouldNotUpdateNonPlannedEntry(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	queueRepo := repositories.NewJobQueue(db)

	user := &repositories.User{ID: 7020, Name: "Non-Planned User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	entry := &models.IndustryJobQueueEntry{
		UserID:          user.ID,
		BlueprintTypeID: 787,
		Activity:        "manufacturing",
		Runs:            10,
		FacilityTax:     0,
	}

	created, err := queueRepo.Create(context.Background(), entry)
	assert.NoError(t, err)

	// Link to ESI job (makes it 'active')
	err = queueRepo.LinkToEsiJob(context.Background(), created.ID, 999999)
	assert.NoError(t, err)

	// Update should return nil (no rows affected)
	updated, err := queueRepo.Update(context.Background(), created.ID, user.ID, entry)
	assert.NoError(t, err)
	assert.Nil(t, updated)
}

func Test_JobQueueShouldCancelEntry(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	queueRepo := repositories.NewJobQueue(db)

	user := &repositories.User{ID: 7030, Name: "Cancel Queue User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	entry := &models.IndustryJobQueueEntry{
		UserID:          user.ID,
		BlueprintTypeID: 787,
		Activity:        "manufacturing",
		Runs:            10,
		FacilityTax:     0,
	}

	created, err := queueRepo.Create(context.Background(), entry)
	assert.NoError(t, err)

	err = queueRepo.Cancel(context.Background(), created.ID, user.ID)
	assert.NoError(t, err)

	// Cancelled entries should not appear in GetByUser
	entries, err := queueRepo.GetByUser(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, entries, 0)
}

func Test_JobQueueShouldNotCancelOtherUsersEntry(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	queueRepo := repositories.NewJobQueue(db)

	user1 := &repositories.User{ID: 7040, Name: "Owner"}
	user2 := &repositories.User{ID: 7041, Name: "Other"}
	err = userRepo.Add(context.Background(), user1)
	assert.NoError(t, err)
	err = userRepo.Add(context.Background(), user2)
	assert.NoError(t, err)

	entry := &models.IndustryJobQueueEntry{
		UserID:          user1.ID,
		BlueprintTypeID: 787,
		Activity:        "manufacturing",
		Runs:            10,
		FacilityTax:     0,
	}

	created, err := queueRepo.Create(context.Background(), entry)
	assert.NoError(t, err)

	// User2 should not be able to cancel user1's entry
	err = queueRepo.Cancel(context.Background(), created.ID, user2.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found or not cancellable")
}

func Test_JobQueueShouldLinkAndComplete(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	queueRepo := repositories.NewJobQueue(db)

	user := &repositories.User{ID: 7050, Name: "Link Test User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	entry := &models.IndustryJobQueueEntry{
		UserID:          user.ID,
		BlueprintTypeID: 787,
		Activity:        "manufacturing",
		Runs:            10,
		FacilityTax:     0,
	}

	created, err := queueRepo.Create(context.Background(), entry)
	assert.NoError(t, err)

	// Link to ESI job
	esiJobID := int64(12345)
	err = queueRepo.LinkToEsiJob(context.Background(), created.ID, esiJobID)
	assert.NoError(t, err)

	// Verify it's active
	linked, err := queueRepo.GetLinkedActiveJobs(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, linked, 1)
	assert.Equal(t, "active", linked[0].Status)
	assert.Equal(t, &esiJobID, linked[0].EsiJobID)

	// Complete the job
	err = queueRepo.CompleteJob(context.Background(), created.ID)
	assert.NoError(t, err)

	// No more linked active jobs
	linked, err = queueRepo.GetLinkedActiveJobs(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, linked, 0)
}

func Test_JobQueueShouldGetPlannedJobs(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	queueRepo := repositories.NewJobQueue(db)

	user := &repositories.User{ID: 7060, Name: "Planned Jobs User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	// Create two planned entries
	entry1 := &models.IndustryJobQueueEntry{
		UserID:          user.ID,
		BlueprintTypeID: 787,
		Activity:        "manufacturing",
		Runs:            10,
		FacilityTax:     0,
	}
	entry2 := &models.IndustryJobQueueEntry{
		UserID:          user.ID,
		BlueprintTypeID: 46166,
		Activity:        "reaction",
		Runs:            100,
		FacilityTax:     0,
	}

	created1, err := queueRepo.Create(context.Background(), entry1)
	assert.NoError(t, err)
	created2, err := queueRepo.Create(context.Background(), entry2)
	assert.NoError(t, err)

	// Link first one to ESI job (makes it active)
	err = queueRepo.LinkToEsiJob(context.Background(), created1.ID, 55555)
	assert.NoError(t, err)

	// Only second entry should be planned
	planned, err := queueRepo.GetPlannedJobs(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, planned, 1)
	assert.Equal(t, created2.ID, planned[0].ID)
	assert.Equal(t, "reaction", planned[0].Activity)
}

func Test_JobQueueShouldEnrichTransportFields(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	queueRepo := repositories.NewJobQueue(db)
	transportJobsRepo := repositories.NewTransportJobs(db)

	user := &repositories.User{ID: 7080, Name: "Transport Enrich User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	// Create a transport job
	tj, err := transportJobsRepo.Create(context.Background(), &models.TransportJob{
		UserID:               user.ID,
		OriginStationID:      60003760,
		DestinationStationID: 60008494,
		OriginSystemID:       30000142,
		DestinationSystemID:  30002187,
		TransportMethod:      "freighter",
		RoutePreference:      "shortest",
		Status:               "planned",
		TotalVolumeM3:        50000,
		TotalCollateral:      1000000000,
		EstimatedCost:        25000000,
		Jumps:                15,
		FulfillmentType:      "self_haul",
	})
	assert.NoError(t, err)

	// Create a queue entry linked to the transport job
	entry := &models.IndustryJobQueueEntry{
		UserID:          user.ID,
		BlueprintTypeID: 0,
		Activity:        "transport",
		Runs:            0,
		FacilityTax:     0,
		TransportJobID:  &tj.ID,
	}

	created, err := queueRepo.Create(context.Background(), entry)
	assert.NoError(t, err)
	assert.NotNil(t, created)

	// Fetch via GetByUser and verify enriched transport fields
	entries, err := queueRepo.GetByUser(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "transport", entries[0].Activity)
	assert.Equal(t, &tj.ID, entries[0].TransportJobID)
	assert.Equal(t, "freighter", entries[0].TransportMethod)
	assert.Equal(t, "self_haul", entries[0].TransportFulfillment)
	assert.Equal(t, 50000.0, entries[0].TransportVolumeM3)
	assert.Equal(t, 15, entries[0].TransportJumps)
	// Station names come from stations table — if seeded, they'd be non-empty
	// but in test DB they may be empty if no station seed data exists.
	// Just verify the query runs without errors.
	_ = entries[0].TransportOriginName
	_ = entries[0].TransportDestName
	_ = created
}

func Test_JobQueueShouldReturnEmptyForNoEntries(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	queueRepo := repositories.NewJobQueue(db)

	entries, err := queueRepo.GetByUser(context.Background(), 99999)
	assert.NoError(t, err)
	assert.Len(t, entries, 0)
}

func Test_JobQueueShouldCreateWithPlanRunAndStep(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)
	runsRepo := repositories.NewPlanRuns(db)
	queueRepo := repositories.NewJobQueue(db)

	user := &repositories.User{ID: 7070, Name: "Plan Run Queue User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "Queue Plan",
	})
	assert.NoError(t, err)

	run, err := runsRepo.Create(context.Background(), &models.ProductionPlanRun{
		PlanID:   plan.ID,
		UserID:   user.ID,
		Quantity: 10,
	})
	assert.NoError(t, err)

	stepID := int64(42)
	entry := &models.IndustryJobQueueEntry{
		UserID:          user.ID,
		BlueprintTypeID: 787,
		Activity:        "manufacturing",
		Runs:            10,
		FacilityTax:     1.0,
		PlanRunID:       &run.ID,
		PlanStepID:      &stepID,
	}

	created, err := queueRepo.Create(context.Background(), entry)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, &run.ID, created.PlanRunID)
	assert.Equal(t, &stepID, created.PlanStepID)

	// Verify via GetByUser
	entries, err := queueRepo.GetByUser(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, &run.ID, entries[0].PlanRunID)
	assert.Equal(t, &stepID, entries[0].PlanStepID)
}
