package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_ProductionPlansShouldCreateAndGetByUser(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)

	user := &repositories.User{ID: 8000, Name: "Plans Test User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "Rifter Plan",
	})
	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.NotZero(t, plan.ID)
	assert.Equal(t, user.ID, plan.UserID)
	assert.Equal(t, int64(587), plan.ProductTypeID)
	assert.Equal(t, "Rifter Plan", plan.Name)

	plans, err := plansRepo.GetByUser(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, plans, 1)
	assert.Equal(t, plan.ID, plans[0].ID)
	assert.Equal(t, "Rifter Plan", plans[0].Name)
}

func Test_ProductionPlansShouldGetByID(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)

	user := &repositories.User{ID: 8010, Name: "GetByID Test User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "Test Plan",
	})
	assert.NoError(t, err)

	fetched, err := plansRepo.GetByID(context.Background(), plan.ID, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, fetched)
	assert.Equal(t, plan.ID, fetched.ID)
	assert.Equal(t, "Test Plan", fetched.Name)
	assert.NotNil(t, fetched.Steps)
	assert.Len(t, fetched.Steps, 0) // No steps yet
}

func Test_ProductionPlansShouldReturnNilForWrongUser(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)

	user := &repositories.User{ID: 8020, Name: "Owner"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "My Plan",
	})
	assert.NoError(t, err)

	// Different user should not see this plan
	fetched, err := plansRepo.GetByID(context.Background(), plan.ID, 9999)
	assert.NoError(t, err)
	assert.Nil(t, fetched)
}

func Test_ProductionPlansShouldUpdate(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)

	user := &repositories.User{ID: 8030, Name: "Update Test User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "Old Name",
	})
	assert.NoError(t, err)

	notes := "Some notes"
	err = plansRepo.Update(context.Background(), plan.ID, user.ID, "New Name", &notes, nil, nil)
	assert.NoError(t, err)

	fetched, err := plansRepo.GetByID(context.Background(), plan.ID, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "New Name", fetched.Name)
	assert.NotNil(t, fetched.Notes)
	assert.Equal(t, "Some notes", *fetched.Notes)
}

func Test_ProductionPlansShouldDelete(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)

	user := &repositories.User{ID: 8040, Name: "Delete Test User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "To Delete",
	})
	assert.NoError(t, err)

	err = plansRepo.Delete(context.Background(), plan.ID, user.ID)
	assert.NoError(t, err)

	fetched, err := plansRepo.GetByID(context.Background(), plan.ID, user.ID)
	assert.NoError(t, err)
	assert.Nil(t, fetched)
}

func Test_ProductionPlansShouldCreateAndGetSteps(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)

	user := &repositories.User{ID: 8050, Name: "Steps Test User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "Steps Test",
	})
	assert.NoError(t, err)

	// Create root step
	rootStep, err := plansRepo.CreateStep(context.Background(), &models.ProductionPlanStep{
		PlanID:           plan.ID,
		ProductTypeID:    587,
		BlueprintTypeID:  787,
		Activity:         "manufacturing",
		MELevel:          10,
		TELevel:          20,
		IndustrySkill:    5,
		AdvIndustrySkill: 5,
		Structure:        "raitaru",
		Rig:              "t2",
		Security:         "high",
		FacilityTax:      1.0,
	})
	assert.NoError(t, err)
	assert.NotZero(t, rootStep.ID)

	// Create child step
	childStep, err := plansRepo.CreateStep(context.Background(), &models.ProductionPlanStep{
		PlanID:           plan.ID,
		ParentStepID:     &rootStep.ID,
		ProductTypeID:    34,
		BlueprintTypeID:  100,
		Activity:         "manufacturing",
		MELevel:          10,
		TELevel:          20,
		IndustrySkill:    5,
		AdvIndustrySkill: 5,
		Structure:        "raitaru",
		Rig:              "t2",
		Security:         "high",
		FacilityTax:      1.0,
	})
	assert.NoError(t, err)
	assert.NotZero(t, childStep.ID)
	assert.Equal(t, &rootStep.ID, childStep.ParentStepID)

	// Get full plan with steps
	fetched, err := plansRepo.GetByID(context.Background(), plan.ID, user.ID)
	assert.NoError(t, err)
	assert.Len(t, fetched.Steps, 2)
	// Root step first (parent_step_id IS NULL)
	assert.Nil(t, fetched.Steps[0].ParentStepID)
	assert.NotNil(t, fetched.Steps[1].ParentStepID)
}

func Test_ProductionPlansShouldUpdateStep(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)

	user := &repositories.User{ID: 8060, Name: "Update Step User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "Update Step Test",
	})
	assert.NoError(t, err)

	step, err := plansRepo.CreateStep(context.Background(), &models.ProductionPlanStep{
		PlanID:           plan.ID,
		ProductTypeID:    587,
		BlueprintTypeID:  787,
		Activity:         "manufacturing",
		MELevel:          10,
		TELevel:          20,
		IndustrySkill:    5,
		AdvIndustrySkill: 5,
		Structure:        "raitaru",
		Rig:              "t2",
		Security:         "high",
		FacilityTax:      1.0,
	})
	assert.NoError(t, err)

	// Update step params
	locationID := int64(60003760)
	ownerType := "character"
	ownerID := int64(1001)
	err = plansRepo.UpdateStep(context.Background(), step.ID, plan.ID, user.ID, &models.ProductionPlanStep{
		MELevel:          8,
		TELevel:          16,
		IndustrySkill:    4,
		AdvIndustrySkill: 3,
		Structure:        "azbel",
		Rig:              "t1",
		Security:         "low",
		FacilityTax:      2.5,
		SourceLocationID: &locationID,
		SourceOwnerType:  &ownerType,
		SourceOwnerID:    &ownerID,
	})
	assert.NoError(t, err)

	fetched, err := plansRepo.GetByID(context.Background(), plan.ID, user.ID)
	assert.NoError(t, err)
	assert.Len(t, fetched.Steps, 1)
	assert.Equal(t, 8, fetched.Steps[0].MELevel)
	assert.Equal(t, "azbel", fetched.Steps[0].Structure)
	assert.Equal(t, &locationID, fetched.Steps[0].SourceLocationID)
}

func Test_ProductionPlansShouldUpdateStepWithOutputFields(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)

	user := &repositories.User{ID: 8065, Name: "Output Fields User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "Output Fields Test",
	})
	assert.NoError(t, err)

	step, err := plansRepo.CreateStep(context.Background(), &models.ProductionPlanStep{
		PlanID:           plan.ID,
		ProductTypeID:    587,
		BlueprintTypeID:  787,
		Activity:         "manufacturing",
		MELevel:          10,
		TELevel:          20,
		IndustrySkill:    5,
		AdvIndustrySkill: 5,
		Structure:        "raitaru",
		Rig:              "t2",
		Security:         "high",
		FacilityTax:      1.0,
	})
	assert.NoError(t, err)

	// Update with output location fields
	outOwnerType := "corporation"
	outOwnerID := int64(5000)
	outDivNum := 3
	outContainerID := int64(9999)
	err = plansRepo.UpdateStep(context.Background(), step.ID, plan.ID, user.ID, &models.ProductionPlanStep{
		MELevel:              10,
		TELevel:              20,
		IndustrySkill:        5,
		AdvIndustrySkill:     5,
		Structure:            "raitaru",
		Rig:                  "t2",
		Security:             "high",
		FacilityTax:          1.0,
		OutputOwnerType:      &outOwnerType,
		OutputOwnerID:        &outOwnerID,
		OutputDivisionNumber: &outDivNum,
		OutputContainerID:    &outContainerID,
	})
	assert.NoError(t, err)

	fetched, err := plansRepo.GetByID(context.Background(), plan.ID, user.ID)
	assert.NoError(t, err)
	assert.Len(t, fetched.Steps, 1)
	assert.NotNil(t, fetched.Steps[0].OutputOwnerType)
	assert.Equal(t, "corporation", *fetched.Steps[0].OutputOwnerType)
	assert.NotNil(t, fetched.Steps[0].OutputOwnerID)
	assert.Equal(t, int64(5000), *fetched.Steps[0].OutputOwnerID)
	assert.NotNil(t, fetched.Steps[0].OutputDivisionNumber)
	assert.Equal(t, 3, *fetched.Steps[0].OutputDivisionNumber)
	assert.NotNil(t, fetched.Steps[0].OutputContainerID)
	assert.Equal(t, int64(9999), *fetched.Steps[0].OutputContainerID)
}

func Test_ProductionPlansShouldBatchUpdateSteps(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)

	user := &repositories.User{ID: 8090, Name: "Batch Update User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "Batch Update Test",
	})
	assert.NoError(t, err)

	rootStep, err := plansRepo.CreateStep(context.Background(), &models.ProductionPlanStep{
		PlanID:           plan.ID,
		ProductTypeID:    587,
		BlueprintTypeID:  787,
		Activity:         "manufacturing",
		MELevel:          10,
		TELevel:          20,
		IndustrySkill:    5,
		AdvIndustrySkill: 5,
		Structure:        "raitaru",
		Rig:              "t2",
		Security:         "high",
		FacilityTax:      1.0,
	})
	assert.NoError(t, err)

	// Create 3 child steps with same product type
	var stepIDs []int64
	for i := 0; i < 3; i++ {
		step, err := plansRepo.CreateStep(context.Background(), &models.ProductionPlanStep{
			PlanID:           plan.ID,
			ParentStepID:     &rootStep.ID,
			ProductTypeID:    34,
			BlueprintTypeID:  100,
			Activity:         "manufacturing",
			MELevel:          10,
			TELevel:          20,
			IndustrySkill:    5,
			AdvIndustrySkill: 5,
			Structure:        "raitaru",
			Rig:              "t2",
			Security:         "high",
			FacilityTax:      1.0,
		})
		assert.NoError(t, err)
		stepIDs = append(stepIDs, step.ID)
	}

	// Batch update all 3 steps
	rowsAffected, err := plansRepo.BatchUpdateSteps(context.Background(), stepIDs, plan.ID, user.ID, &models.ProductionPlanStep{
		MELevel:          8,
		TELevel:          16,
		IndustrySkill:    4,
		AdvIndustrySkill: 3,
		Structure:        "azbel",
		Rig:              "t1",
		Security:         "low",
		FacilityTax:      2.5,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), rowsAffected)

	// Verify all steps were updated
	fetched, err := plansRepo.GetByID(context.Background(), plan.ID, user.ID)
	assert.NoError(t, err)
	for _, step := range fetched.Steps {
		if step.ID == rootStep.ID {
			continue // skip root
		}
		assert.Equal(t, 8, step.MELevel)
		assert.Equal(t, 16, step.TELevel)
		assert.Equal(t, 4, step.IndustrySkill)
		assert.Equal(t, 3, step.AdvIndustrySkill)
		assert.Equal(t, "azbel", step.Structure)
		assert.Equal(t, "t1", step.Rig)
		assert.Equal(t, "low", step.Security)
		assert.Equal(t, 2.5, step.FacilityTax)
	}
}

func Test_ProductionPlansShouldBatchUpdateOnlySelectedSteps(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)

	user := &repositories.User{ID: 8095, Name: "Partial Batch User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "Partial Batch Test",
	})
	assert.NoError(t, err)

	rootStep, err := plansRepo.CreateStep(context.Background(), &models.ProductionPlanStep{
		PlanID:           plan.ID,
		ProductTypeID:    587,
		BlueprintTypeID:  787,
		Activity:         "manufacturing",
		MELevel:          10,
		TELevel:          20,
		IndustrySkill:    5,
		AdvIndustrySkill: 5,
		Structure:        "raitaru",
		Rig:              "t2",
		Security:         "high",
		FacilityTax:      1.0,
	})
	assert.NoError(t, err)

	// Create 3 child steps
	var stepIDs []int64
	for i := 0; i < 3; i++ {
		step, err := plansRepo.CreateStep(context.Background(), &models.ProductionPlanStep{
			PlanID:           plan.ID,
			ParentStepID:     &rootStep.ID,
			ProductTypeID:    34,
			BlueprintTypeID:  100,
			Activity:         "manufacturing",
			MELevel:          10,
			TELevel:          20,
			IndustrySkill:    5,
			AdvIndustrySkill: 5,
			Structure:        "raitaru",
			Rig:              "t2",
			Security:         "high",
			FacilityTax:      1.0,
		})
		assert.NoError(t, err)
		stepIDs = append(stepIDs, step.ID)
	}

	// Batch update only first 2 steps
	rowsAffected, err := plansRepo.BatchUpdateSteps(context.Background(), stepIDs[:2], plan.ID, user.ID, &models.ProductionPlanStep{
		MELevel:          5,
		TELevel:          10,
		IndustrySkill:    5,
		AdvIndustrySkill: 5,
		Structure:        "sotiyo",
		Rig:              "t2",
		Security:         "null",
		FacilityTax:      0.5,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), rowsAffected)

	// Verify the third step was NOT updated
	fetched, err := plansRepo.GetByID(context.Background(), plan.ID, user.ID)
	assert.NoError(t, err)

	for _, step := range fetched.Steps {
		if step.ID == stepIDs[2] {
			// Third step should remain unchanged
			assert.Equal(t, 10, step.MELevel)
			assert.Equal(t, 20, step.TELevel)
			assert.Equal(t, "raitaru", step.Structure)
		} else if step.ID == stepIDs[0] || step.ID == stepIDs[1] {
			// First two should be updated
			assert.Equal(t, 5, step.MELevel)
			assert.Equal(t, 10, step.TELevel)
			assert.Equal(t, "sotiyo", step.Structure)
		}
	}
}

func Test_ProductionPlansShouldDeleteStepAndCascadeChildren(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)

	user := &repositories.User{ID: 8070, Name: "Cascade Test User"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "Cascade Test",
	})
	assert.NoError(t, err)

	rootStep, err := plansRepo.CreateStep(context.Background(), &models.ProductionPlanStep{
		PlanID:           plan.ID,
		ProductTypeID:    587,
		BlueprintTypeID:  787,
		Activity:         "manufacturing",
		MELevel:          10,
		TELevel:          20,
		IndustrySkill:    5,
		AdvIndustrySkill: 5,
		Structure:        "raitaru",
		Rig:              "t2",
		Security:         "high",
		FacilityTax:      1.0,
	})
	assert.NoError(t, err)

	// Create child step
	_, err = plansRepo.CreateStep(context.Background(), &models.ProductionPlanStep{
		PlanID:           plan.ID,
		ParentStepID:     &rootStep.ID,
		ProductTypeID:    34,
		BlueprintTypeID:  100,
		Activity:         "manufacturing",
		MELevel:          10,
		TELevel:          20,
		IndustrySkill:    5,
		AdvIndustrySkill: 5,
		Structure:        "raitaru",
		Rig:              "t2",
		Security:         "high",
		FacilityTax:      1.0,
	})
	assert.NoError(t, err)

	// Verify 2 steps
	fetched, err := plansRepo.GetByID(context.Background(), plan.ID, user.ID)
	assert.NoError(t, err)
	assert.Len(t, fetched.Steps, 2)

	// Delete root step — should cascade delete the child
	err = plansRepo.DeleteStep(context.Background(), rootStep.ID, plan.ID, user.ID)
	assert.NoError(t, err)

	fetched, err = plansRepo.GetByID(context.Background(), plan.ID, user.ID)
	assert.NoError(t, err)
	assert.Len(t, fetched.Steps, 0)
}

func Test_ProductionPlansShouldDeletePlanAndCascadeSteps(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	plansRepo := repositories.NewProductionPlans(db)

	user := &repositories.User{ID: 8080, Name: "Plan Cascade Test"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	plan, err := plansRepo.Create(context.Background(), &models.ProductionPlan{
		UserID:        user.ID,
		ProductTypeID: 587,
		Name:          "Plan Cascade",
	})
	assert.NoError(t, err)

	_, err = plansRepo.CreateStep(context.Background(), &models.ProductionPlanStep{
		PlanID:           plan.ID,
		ProductTypeID:    587,
		BlueprintTypeID:  787,
		Activity:         "manufacturing",
		MELevel:          10,
		TELevel:          20,
		IndustrySkill:    5,
		AdvIndustrySkill: 5,
		Structure:        "raitaru",
		Rig:              "t2",
		Security:         "high",
		FacilityTax:      1.0,
	})
	assert.NoError(t, err)

	// Delete plan — should cascade delete steps
	err = plansRepo.Delete(context.Background(), plan.ID, user.ID)
	assert.NoError(t, err)

	fetched, err := plansRepo.GetByID(context.Background(), plan.ID, user.ID)
	assert.NoError(t, err)
	assert.Nil(t, fetched)
}
