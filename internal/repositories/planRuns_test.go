package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_PlanRunsShouldCreateAndGetByPlan(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)
	runsRepo := repositories.NewPlanRuns(db)

	user := &repositories.User{ID: 8500, Name: "Plan Runs Test User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "Rifter Plan",
	})
	assert.NoError(t, err)

	run, err := runsRepo.Create(context.Background(), &models.ProductionPlanRun{
		PlanID:   plan.ID,
		UserID:   user.ID,
		Quantity: 10,
	})
	assert.NoError(t, err)
	assert.NotNil(t, run)
	assert.NotZero(t, run.ID)
	assert.Equal(t, plan.ID, run.PlanID)
	assert.Equal(t, user.ID, run.UserID)
	assert.Equal(t, 10, run.Quantity)
	assert.NotZero(t, run.CreatedAt)

	runs, err := runsRepo.GetByPlan(context.Background(), plan.ID, user.ID)
	assert.NoError(t, err)
	assert.Len(t, runs, 1)
	assert.Equal(t, run.ID, runs[0].ID)
	assert.Equal(t, 10, runs[0].Quantity)
	assert.Equal(t, "Rifter Plan", runs[0].PlanName)
	assert.Equal(t, "pending", runs[0].Status)
	assert.NotNil(t, runs[0].JobSummary)
	assert.Equal(t, 0, runs[0].JobSummary.Total)
}

func Test_PlanRunsShouldGetByIDWithJobs(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)
	runsRepo := repositories.NewPlanRuns(db)
	queueRepo := repositories.NewJobQueue(db)

	user := &repositories.User{ID: 8510, Name: "Plan Run GetByID User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "Test Plan",
	})
	assert.NoError(t, err)

	run, err := runsRepo.Create(context.Background(), &models.ProductionPlanRun{
		PlanID:   plan.ID,
		UserID:   user.ID,
		Quantity: 5,
	})
	assert.NoError(t, err)

	// Create a job linked to this run
	stepID := int64(999)
	_, err = queueRepo.Create(context.Background(), &models.IndustryJobQueueEntry{
		UserID:          user.ID,
		BlueprintTypeID: 787,
		Activity:        "manufacturing",
		Runs:            5,
		FacilityTax:     1.0,
		PlanRunID:       &run.ID,
		PlanStepID:      &stepID,
	})
	assert.NoError(t, err)

	fetched, err := runsRepo.GetByID(context.Background(), run.ID, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, fetched)
	assert.Equal(t, run.ID, fetched.ID)
	assert.Equal(t, 5, fetched.Quantity)
	assert.Equal(t, "Test Plan", fetched.PlanName)
	assert.Equal(t, "pending", fetched.Status)
	assert.NotNil(t, fetched.Jobs)
	assert.Len(t, fetched.Jobs, 1)
	assert.Equal(t, "manufacturing", fetched.Jobs[0].Activity)
	assert.Equal(t, &run.ID, fetched.Jobs[0].PlanRunID)
	assert.Equal(t, &stepID, fetched.Jobs[0].PlanStepID)
	assert.NotNil(t, fetched.JobSummary)
	assert.Equal(t, 1, fetched.JobSummary.Total)
	assert.Equal(t, 1, fetched.JobSummary.Planned)
}

func Test_PlanRunsShouldDeriveStatusPending(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)
	runsRepo := repositories.NewPlanRuns(db)
	queueRepo := repositories.NewJobQueue(db)

	user := &repositories.User{ID: 8520, Name: "Status Pending User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "Pending Plan",
	})
	assert.NoError(t, err)

	run, err := runsRepo.Create(context.Background(), &models.ProductionPlanRun{
		PlanID:   plan.ID,
		UserID:   user.ID,
		Quantity: 1,
	})
	assert.NoError(t, err)

	// Create planned jobs
	for i := 0; i < 3; i++ {
		_, err = queueRepo.Create(context.Background(), &models.IndustryJobQueueEntry{
			UserID:          user.ID,
			BlueprintTypeID: 787,
			Activity:        "manufacturing",
			Runs:            1,
			FacilityTax:     1.0,
			PlanRunID:       &run.ID,
		})
		assert.NoError(t, err)
	}

	runs, err := runsRepo.GetByPlan(context.Background(), plan.ID, user.ID)
	assert.NoError(t, err)
	assert.Len(t, runs, 1)
	assert.Equal(t, "pending", runs[0].Status)
	assert.Equal(t, 3, runs[0].JobSummary.Total)
	assert.Equal(t, 3, runs[0].JobSummary.Planned)
}

func Test_PlanRunsShouldDeriveStatusInProgress(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)
	runsRepo := repositories.NewPlanRuns(db)
	queueRepo := repositories.NewJobQueue(db)

	user := &repositories.User{ID: 8530, Name: "Status InProgress User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "InProgress Plan",
	})
	assert.NoError(t, err)

	run, err := runsRepo.Create(context.Background(), &models.ProductionPlanRun{
		PlanID:   plan.ID,
		UserID:   user.ID,
		Quantity: 1,
	})
	assert.NoError(t, err)

	// Create a planned job and an active job
	job1, err := queueRepo.Create(context.Background(), &models.IndustryJobQueueEntry{
		UserID:          user.ID,
		BlueprintTypeID: 787,
		Activity:        "manufacturing",
		Runs:            1,
		FacilityTax:     1.0,
		PlanRunID:       &run.ID,
	})
	assert.NoError(t, err)

	_, err = queueRepo.Create(context.Background(), &models.IndustryJobQueueEntry{
		UserID:          user.ID,
		BlueprintTypeID: 788,
		Activity:        "manufacturing",
		Runs:            1,
		FacilityTax:     1.0,
		PlanRunID:       &run.ID,
	})
	assert.NoError(t, err)

	// Mark first job as active via LinkToEsiJob
	err = queueRepo.LinkToEsiJob(context.Background(), job1.ID, 99999)
	assert.NoError(t, err)

	runs, err := runsRepo.GetByPlan(context.Background(), plan.ID, user.ID)
	assert.NoError(t, err)
	assert.Len(t, runs, 1)
	assert.Equal(t, "in_progress", runs[0].Status)
	assert.Equal(t, 2, runs[0].JobSummary.Total)
	assert.Equal(t, 1, runs[0].JobSummary.Planned)
	assert.Equal(t, 1, runs[0].JobSummary.Active)
}

func Test_PlanRunsShouldDeriveStatusCompleted(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)
	runsRepo := repositories.NewPlanRuns(db)
	queueRepo := repositories.NewJobQueue(db)

	user := &repositories.User{ID: 8540, Name: "Status Completed User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "Completed Plan",
	})
	assert.NoError(t, err)

	run, err := runsRepo.Create(context.Background(), &models.ProductionPlanRun{
		PlanID:   plan.ID,
		UserID:   user.ID,
		Quantity: 1,
	})
	assert.NoError(t, err)

	// Create two jobs and complete both
	job1, err := queueRepo.Create(context.Background(), &models.IndustryJobQueueEntry{
		UserID:          user.ID,
		BlueprintTypeID: 787,
		Activity:        "manufacturing",
		Runs:            1,
		FacilityTax:     1.0,
		PlanRunID:       &run.ID,
	})
	assert.NoError(t, err)

	job2, err := queueRepo.Create(context.Background(), &models.IndustryJobQueueEntry{
		UserID:          user.ID,
		BlueprintTypeID: 788,
		Activity:        "manufacturing",
		Runs:            1,
		FacilityTax:     1.0,
		PlanRunID:       &run.ID,
	})
	assert.NoError(t, err)

	err = queueRepo.CompleteJob(context.Background(), job1.ID)
	assert.NoError(t, err)
	err = queueRepo.CompleteJob(context.Background(), job2.ID)
	assert.NoError(t, err)

	runs, err := runsRepo.GetByPlan(context.Background(), plan.ID, user.ID)
	assert.NoError(t, err)
	assert.Len(t, runs, 1)
	assert.Equal(t, "completed", runs[0].Status)
	assert.Equal(t, 2, runs[0].JobSummary.Total)
	assert.Equal(t, 2, runs[0].JobSummary.Completed)
}

func Test_PlanRunsShouldNotReturnOtherUsersRuns(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)
	runsRepo := repositories.NewPlanRuns(db)

	user1 := &repositories.User{ID: 8550, Name: "User 1"}
	err = userRepo.Add(context.Background(), user1)
	assert.NoError(t, err)

	user2 := &repositories.User{ID: 8551, Name: "User 2"}
	err = userRepo.Add(context.Background(), user2)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user1.ID,
		ProductTypeID: 587,
		Name:          "User 1 Plan",
	})
	assert.NoError(t, err)

	_, err = runsRepo.Create(context.Background(), &models.ProductionPlanRun{
		PlanID:   plan.ID,
		UserID:   user1.ID,
		Quantity: 10,
	})
	assert.NoError(t, err)

	// User 2 should see no runs
	runs, err := runsRepo.GetByPlan(context.Background(), plan.ID, user2.ID)
	assert.NoError(t, err)
	assert.Len(t, runs, 0)
}

func Test_PlanRunsShouldReturnEmptyForNoRuns(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)
	runsRepo := repositories.NewPlanRuns(db)

	user := &repositories.User{ID: 8560, Name: "No Runs User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "Empty Plan",
	})
	assert.NoError(t, err)

	runs, err := runsRepo.GetByPlan(context.Background(), plan.ID, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, runs)
	assert.Len(t, runs, 0)
}

func Test_PlanRunsShouldDelete(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)
	runsRepo := repositories.NewPlanRuns(db)
	queueRepo := repositories.NewJobQueue(db)

	user := &repositories.User{ID: 8570, Name: "Delete Run User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "Delete Test Plan",
	})
	assert.NoError(t, err)

	run, err := runsRepo.Create(context.Background(), &models.ProductionPlanRun{
		PlanID:   plan.ID,
		UserID:   user.ID,
		Quantity: 5,
	})
	assert.NoError(t, err)

	// Create a job linked to this run
	job, err := queueRepo.Create(context.Background(), &models.IndustryJobQueueEntry{
		UserID:          user.ID,
		BlueprintTypeID: 787,
		Activity:        "manufacturing",
		Runs:            5,
		FacilityTax:     1.0,
		PlanRunID:       &run.ID,
	})
	assert.NoError(t, err)

	// Delete the run
	err = runsRepo.Delete(context.Background(), run.ID, user.ID)
	assert.NoError(t, err)

	// Run should be gone
	fetched, err := runsRepo.GetByID(context.Background(), run.ID, user.ID)
	assert.NoError(t, err)
	assert.Nil(t, fetched)

	// Job should still exist but with null plan_run_id (ON DELETE SET NULL)
	entries, err := queueRepo.GetByUser(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, job.ID, entries[0].ID)
	assert.Nil(t, entries[0].PlanRunID)
}

func Test_PlanRunsShouldGetByUser(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)
	runsRepo := repositories.NewPlanRuns(db)

	user := &repositories.User{ID: 8580, Name: "GetByUser Test"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan1, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "Rifter Plan",
	})
	assert.NoError(t, err)

	plan2, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 588,
		Name:          "Slasher Plan",
	})
	assert.NoError(t, err)

	_, err = runsRepo.Create(context.Background(), &models.ProductionPlanRun{
		PlanID: plan1.ID, UserID: user.ID, Quantity: 10,
	})
	assert.NoError(t, err)

	_, err = runsRepo.Create(context.Background(), &models.ProductionPlanRun{
		PlanID: plan2.ID, UserID: user.ID, Quantity: 5,
	})
	assert.NoError(t, err)

	runs, err := runsRepo.GetByUser(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, runs, 2)
	// Most recent first
	assert.Equal(t, 5, runs[0].Quantity)
	assert.Equal(t, "Slasher Plan", runs[0].PlanName)
	assert.Equal(t, 10, runs[1].Quantity)
	assert.Equal(t, "Rifter Plan", runs[1].PlanName)
	assert.NotNil(t, runs[0].JobSummary)
	assert.NotNil(t, runs[1].JobSummary)
}

func Test_PlanRunsShouldGetByUserEmpty(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	runsRepo := repositories.NewPlanRuns(db)

	runs, err := runsRepo.GetByUser(context.Background(), 99999)
	assert.NoError(t, err)
	assert.NotNil(t, runs)
	assert.Len(t, runs, 0)
}

func Test_PlanRunsShouldCancelPlannedJobs(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)
	runsRepo := repositories.NewPlanRuns(db)
	queueRepo := repositories.NewJobQueue(db)

	user := &repositories.User{ID: 8590, Name: "Cancel Planned User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "Cancel Test Plan",
	})
	assert.NoError(t, err)

	run, err := runsRepo.Create(context.Background(), &models.ProductionPlanRun{
		PlanID: plan.ID, UserID: user.ID, Quantity: 1,
	})
	assert.NoError(t, err)

	// Create 3 planned jobs and 1 active job
	for i := 0; i < 3; i++ {
		_, err = queueRepo.Create(context.Background(), &models.IndustryJobQueueEntry{
			UserID:          user.ID,
			BlueprintTypeID: 787,
			Activity:        "manufacturing",
			Runs:            1,
			FacilityTax:     1.0,
			PlanRunID:       &run.ID,
		})
		assert.NoError(t, err)
	}
	activeJob, err := queueRepo.Create(context.Background(), &models.IndustryJobQueueEntry{
		UserID:          user.ID,
		BlueprintTypeID: 788,
		Activity:        "manufacturing",
		Runs:            1,
		FacilityTax:     1.0,
		PlanRunID:       &run.ID,
	})
	assert.NoError(t, err)
	err = queueRepo.LinkToEsiJob(context.Background(), activeJob.ID, 99999)
	assert.NoError(t, err)

	// Cancel planned jobs
	cancelled, err := runsRepo.CancelPlannedJobs(context.Background(), run.ID, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), cancelled)

	// Verify: active job still active, 3 planned now cancelled
	fetched, err := runsRepo.GetByID(context.Background(), run.ID, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, fetched)
	assert.Equal(t, 4, fetched.JobSummary.Total)
	assert.Equal(t, 0, fetched.JobSummary.Planned)
	assert.Equal(t, 1, fetched.JobSummary.Active)
	assert.Equal(t, 3, fetched.JobSummary.Cancelled)
}

func Test_PlanRunsGetPendingOutputForPlan_WithPlannedJobs(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)
	runsRepo := repositories.NewPlanRuns(db)
	queueRepo := repositories.NewJobQueue(db)

	user := &repositories.User{ID: 8610, Name: "Pending Output User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "Pending Output Plan",
	})
	assert.NoError(t, err)

	run, err := runsRepo.Create(context.Background(), &models.ProductionPlanRun{
		PlanID:   plan.ID,
		UserID:   user.ID,
		Quantity: 50,
	})
	assert.NoError(t, err)

	// Create a planned job (status = 'planned')
	_, err = queueRepo.Create(context.Background(), &models.IndustryJobQueueEntry{
		UserID:          user.ID,
		BlueprintTypeID: 787,
		Activity:        "manufacturing",
		Runs:            1,
		FacilityTax:     1.0,
		PlanRunID:       &run.ID,
	})
	assert.NoError(t, err)

	pending, err := runsRepo.GetPendingOutputForPlan(context.Background(), plan.ID, user.ID)
	assert.NoError(t, err)
	// run has quantity=50 and it has a 'planned' job, so pending = 50
	assert.Equal(t, int64(50), pending)
}

func Test_PlanRunsGetPendingOutputForPlan_WithActiveJob(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)
	runsRepo := repositories.NewPlanRuns(db)
	queueRepo := repositories.NewJobQueue(db)

	user := &repositories.User{ID: 8611, Name: "Active Job User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "Active Job Plan",
	})
	assert.NoError(t, err)

	run, err := runsRepo.Create(context.Background(), &models.ProductionPlanRun{
		PlanID:   plan.ID,
		UserID:   user.ID,
		Quantity: 30,
	})
	assert.NoError(t, err)

	// Create a job and mark it active
	job, err := queueRepo.Create(context.Background(), &models.IndustryJobQueueEntry{
		UserID:          user.ID,
		BlueprintTypeID: 787,
		Activity:        "manufacturing",
		Runs:            1,
		FacilityTax:     1.0,
		PlanRunID:       &run.ID,
	})
	assert.NoError(t, err)
	err = queueRepo.LinkToEsiJob(context.Background(), job.ID, 12345)
	assert.NoError(t, err)

	pending, err := runsRepo.GetPendingOutputForPlan(context.Background(), plan.ID, user.ID)
	assert.NoError(t, err)
	// run has quantity=30 and an 'active' job, so pending = 30
	assert.Equal(t, int64(30), pending)
}

func Test_PlanRunsGetPendingOutputForPlan_AllCompleted(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)
	runsRepo := repositories.NewPlanRuns(db)
	queueRepo := repositories.NewJobQueue(db)

	user := &repositories.User{ID: 8612, Name: "All Completed User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "All Completed Plan",
	})
	assert.NoError(t, err)

	run, err := runsRepo.Create(context.Background(), &models.ProductionPlanRun{
		PlanID:   plan.ID,
		UserID:   user.ID,
		Quantity: 20,
	})
	assert.NoError(t, err)

	// Create a job and complete it
	job, err := queueRepo.Create(context.Background(), &models.IndustryJobQueueEntry{
		UserID:          user.ID,
		BlueprintTypeID: 787,
		Activity:        "manufacturing",
		Runs:            1,
		FacilityTax:     1.0,
		PlanRunID:       &run.ID,
	})
	assert.NoError(t, err)
	err = queueRepo.CompleteJob(context.Background(), job.ID)
	assert.NoError(t, err)

	pending, err := runsRepo.GetPendingOutputForPlan(context.Background(), plan.ID, user.ID)
	assert.NoError(t, err)
	// All jobs are completed, so pending = 0
	assert.Equal(t, int64(0), pending)
}

func Test_PlanRunsGetPendingOutputForPlan_NoRuns(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)
	runsRepo := repositories.NewPlanRuns(db)

	user := &repositories.User{ID: 8613, Name: "No Runs Output User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "No Runs Output Plan",
	})
	assert.NoError(t, err)

	pending, err := runsRepo.GetPendingOutputForPlan(context.Background(), plan.ID, user.ID)
	assert.NoError(t, err)
	// No runs at all, so pending = 0
	assert.Equal(t, int64(0), pending)
}

func Test_PlanRunsGetPendingOutputForPlan_MultipleRuns(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)
	runsRepo := repositories.NewPlanRuns(db)
	queueRepo := repositories.NewJobQueue(db)

	user := &repositories.User{ID: 8614, Name: "Multi Run Output User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "Multi Run Plan",
	})
	assert.NoError(t, err)

	// Run 1: quantity=10 with a planned job (pending)
	run1, err := runsRepo.Create(context.Background(), &models.ProductionPlanRun{
		PlanID:   plan.ID,
		UserID:   user.ID,
		Quantity: 10,
	})
	assert.NoError(t, err)
	_, err = queueRepo.Create(context.Background(), &models.IndustryJobQueueEntry{
		UserID:          user.ID,
		BlueprintTypeID: 787,
		Activity:        "manufacturing",
		Runs:            1,
		FacilityTax:     1.0,
		PlanRunID:       &run1.ID,
	})
	assert.NoError(t, err)

	// Run 2: quantity=15 with a planned job (pending)
	run2, err := runsRepo.Create(context.Background(), &models.ProductionPlanRun{
		PlanID:   plan.ID,
		UserID:   user.ID,
		Quantity: 15,
	})
	assert.NoError(t, err)
	_, err = queueRepo.Create(context.Background(), &models.IndustryJobQueueEntry{
		UserID:          user.ID,
		BlueprintTypeID: 787,
		Activity:        "manufacturing",
		Runs:            1,
		FacilityTax:     1.0,
		PlanRunID:       &run2.ID,
	})
	assert.NoError(t, err)

	// Run 3: quantity=20 with a completed job (not pending)
	run3, err := runsRepo.Create(context.Background(), &models.ProductionPlanRun{
		PlanID:   plan.ID,
		UserID:   user.ID,
		Quantity: 20,
	})
	assert.NoError(t, err)
	job3, err := queueRepo.Create(context.Background(), &models.IndustryJobQueueEntry{
		UserID:          user.ID,
		BlueprintTypeID: 787,
		Activity:        "manufacturing",
		Runs:            1,
		FacilityTax:     1.0,
		PlanRunID:       &run3.ID,
	})
	assert.NoError(t, err)
	err = queueRepo.CompleteJob(context.Background(), job3.ID)
	assert.NoError(t, err)

	pending, err := runsRepo.GetPendingOutputForPlan(context.Background(), plan.ID, user.ID)
	assert.NoError(t, err)
	// Runs 1 and 2 are pending: 10 + 15 = 25
	assert.Equal(t, int64(25), pending)
}

func Test_PlanRunsShouldCancelPlannedJobsNoneToCancel(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)
	runsRepo := repositories.NewPlanRuns(db)
	queueRepo := repositories.NewJobQueue(db)

	user := &repositories.User{ID: 8600, Name: "No Cancel User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "No Cancel Plan",
	})
	assert.NoError(t, err)

	run, err := runsRepo.Create(context.Background(), &models.ProductionPlanRun{
		PlanID: plan.ID, UserID: user.ID, Quantity: 1,
	})
	assert.NoError(t, err)

	// Create only a completed job
	job, err := queueRepo.Create(context.Background(), &models.IndustryJobQueueEntry{
		UserID:          user.ID,
		BlueprintTypeID: 787,
		Activity:        "manufacturing",
		Runs:            1,
		FacilityTax:     1.0,
		PlanRunID:       &run.ID,
	})
	assert.NoError(t, err)
	err = queueRepo.CompleteJob(context.Background(), job.ID)
	assert.NoError(t, err)

	// Cancel should return 0
	cancelled, err := runsRepo.CancelPlannedJobs(context.Background(), run.ID, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), cancelled)
}
