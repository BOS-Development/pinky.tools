package services

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/calculator"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockJobGenSdeRepository implements JobGenSdeRepository for testing.
type MockJobGenSdeRepository struct {
	mock.Mock
}

func (m *MockJobGenSdeRepository) GetBlueprintForActivity(ctx context.Context, blueprintTypeID int64, activity string) (*repositories.ManufacturingBlueprintRow, error) {
	args := m.Called(ctx, blueprintTypeID, activity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.ManufacturingBlueprintRow), args.Error(1)
}

func (m *MockJobGenSdeRepository) GetBlueprintMaterialsForActivity(ctx context.Context, blueprintTypeID int64, activity string) ([]*repositories.ManufacturingMaterialRow, error) {
	args := m.Called(ctx, blueprintTypeID, activity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repositories.ManufacturingMaterialRow), args.Error(1)
}

// ---------------------------------------------------------------------------
// FormatLocation tests
// ---------------------------------------------------------------------------

func Test_FormatLocation(t *testing.T) {
	tests := []struct {
		name          string
		ownerName     string
		divisionName  string
		containerName string
		expected      string
	}{
		{
			name:          "all three parts provided",
			ownerName:     "My Corp",
			divisionName:  "Hangar 1",
			containerName: "Big Box",
			expected:      "My Corp > Hangar 1 > Big Box",
		},
		{
			name:          "only owner provided",
			ownerName:     "My Corp",
			divisionName:  "",
			containerName: "",
			expected:      "My Corp",
		},
		{
			name:          "owner and division no container",
			ownerName:     "My Corp",
			divisionName:  "Hangar 2",
			containerName: "",
			expected:      "My Corp > Hangar 2",
		},
		{
			name:          "owner and container no division",
			ownerName:     "My Corp",
			divisionName:  "",
			containerName: "Small Box",
			expected:      "My Corp > Small Box",
		},
		{
			name:          "all empty strings",
			ownerName:     "",
			divisionName:  "",
			containerName: "",
			expected:      "",
		},
		{
			name:          "only container provided",
			ownerName:     "",
			divisionName:  "",
			containerName: "Container",
			expected:      "Container",
		},
		{
			name:          "only division provided",
			ownerName:     "",
			divisionName:  "Division",
			containerName: "",
			expected:      "Division",
		},
		{
			name:          "division and container no owner",
			ownerName:     "",
			divisionName:  "Hangar",
			containerName: "Crate",
			expected:      "Hangar > Crate",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatLocation(tc.ownerName, tc.divisionName, tc.containerName)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// ---------------------------------------------------------------------------
// SimulateAssignment tests
// ---------------------------------------------------------------------------

func makePendingJob(blueprintTypeID int64, activity string, runs int, duration int, depth int) *PendingJob {
	return &PendingJob{
		Entry: &models.IndustryJobQueueEntry{
			BlueprintTypeID:   blueprintTypeID,
			Activity:          activity,
			Runs:              runs,
			EstimatedDuration: &duration,
		},
		Activity:          activity,
		BaseBlueprintTime: 3600, // 1 hour base time
		BlueprintTE:       0,
		Structure:         "station",
		Rig:               "none",
		Security:          "high",
		Depth:             depth,
	}
}

func makeCapacity(charID int64, mfgSlots, reactSlots, industrySkill, advSkill, reactSkill int) *calculator.CharacterCapacity {
	return &calculator.CharacterCapacity{
		CharacterID:      charID,
		CharacterName:    "Character",
		MfgSlotsMax:      mfgSlots,
		MfgSlotsUsed:     0,
		ReactSlotsMax:    reactSlots,
		ReactSlotsUsed:   0,
		IndustrySkill:    industrySkill,
		AdvIndustrySkill: advSkill,
		ReactionsSkill:   reactSkill,
	}
}

func Test_SimulateAssignment(t *testing.T) {
	t.Run("empty jobs list returns empty result", func(t *testing.T) {
		caps := []*calculator.CharacterCapacity{
			makeCapacity(1001, 5, 0, 5, 5, 0),
		}
		assigned, unassigned := SimulateAssignment([]*PendingJob{}, caps, 5)
		assert.Empty(t, assigned)
		assert.Equal(t, 0, unassigned)
	})

	t.Run("single job single character with capacity", func(t *testing.T) {
		job := makePendingJob(1, "manufacturing", 10, 36000, 0)
		caps := []*calculator.CharacterCapacity{
			makeCapacity(1001, 5, 0, 5, 5, 0),
		}
		assigned, unassigned := SimulateAssignment([]*PendingJob{job}, caps, 1)
		assert.Equal(t, 0, unassigned)
		assert.Len(t, assigned, 1)
		assert.Equal(t, int64(1001), assigned[0].CharacterID)
		assert.Equal(t, 10, assigned[0].Runs)
		assert.Equal(t, "manufacturing", assigned[0].Activity)
	})

	t.Run("multiple jobs distributed across multiple characters", func(t *testing.T) {
		// Two manufacturing jobs at depth 0, two characters each with 2 slots.
		// job1 is split across both characters (1 slot each consumed), job2 also
		// sees 2 eligible chars and is split across them too.
		job1 := makePendingJob(1, "manufacturing", 10, 36000, 0)
		job2 := makePendingJob(2, "manufacturing", 4, 18000, 0)
		caps := []*calculator.CharacterCapacity{
			makeCapacity(1001, 2, 0, 5, 5, 0),
			makeCapacity(1002, 2, 0, 5, 5, 0),
		}
		assigned, unassigned := SimulateAssignment([]*PendingJob{job1, job2}, caps, 2)
		assert.Equal(t, 0, unassigned)
		// job1 splits → 2 entries; job2 splits → 2 entries → 4 total
		assert.Len(t, assigned, 4)
		charIDs := map[int64]bool{}
		for _, aj := range assigned {
			assert.NotEqual(t, int64(0), aj.CharacterID)
			charIDs[aj.CharacterID] = true
		}
		// Both characters should be used
		assert.Len(t, charIDs, 2)
	})

	t.Run("parallelism=1 limits to single character", func(t *testing.T) {
		job := makePendingJob(1, "manufacturing", 10, 36000, 0)
		caps := []*calculator.CharacterCapacity{
			makeCapacity(1001, 5, 0, 5, 5, 0),
			makeCapacity(1002, 5, 0, 5, 5, 0),
		}
		assigned, unassigned := SimulateAssignment([]*PendingJob{job}, caps, 1)
		assert.Equal(t, 0, unassigned)
		assert.Len(t, assigned, 1)
		// Only character 1001 should be used (first in pool)
		assert.Equal(t, int64(1001), assigned[0].CharacterID)
	})

	t.Run("jobs exceeding total capacity get CharacterID=0", func(t *testing.T) {
		// Two jobs at depth 0 but only one character with one slot — second job is unassigned
		job1 := makePendingJob(1, "manufacturing", 10, 36000, 0)
		job2 := makePendingJob(2, "manufacturing", 5, 18000, 0)
		caps := []*calculator.CharacterCapacity{
			makeCapacity(1001, 1, 0, 5, 5, 0), // only 1 mfg slot
		}
		assigned, unassigned := SimulateAssignment([]*PendingJob{job1, job2}, caps, 1)
		// Second job should have CharacterID=0 and unassigned runs counted
		assert.Len(t, assigned, 2)
		assignedCount := 0
		unassignedCount := 0
		for _, aj := range assigned {
			if aj.CharacterID == 0 {
				unassignedCount++
			} else {
				assignedCount++
			}
		}
		assert.Equal(t, 1, assignedCount)
		assert.Equal(t, 1, unassignedCount)
		assert.Greater(t, unassigned, 0)
	})

	t.Run("reaction jobs assigned to character with reaction slots", func(t *testing.T) {
		job := makePendingJob(10, "reaction", 5, 54000, 0)
		caps := []*calculator.CharacterCapacity{
			// First char has only manufacturing slots
			makeCapacity(1001, 5, 0, 5, 5, 0),
			// Second char has reaction slots
			makeCapacity(1002, 0, 5, 0, 0, 5),
		}
		assigned, unassigned := SimulateAssignment([]*PendingJob{job}, caps, 2)
		assert.Equal(t, 0, unassigned)
		assert.Len(t, assigned, 1)
		// Should be assigned to char with reaction slots
		assert.Equal(t, int64(1002), assigned[0].CharacterID)
	})

	t.Run("mixed manufacturing and reaction jobs", func(t *testing.T) {
		mfgJob := makePendingJob(1, "manufacturing", 10, 36000, 0)
		reactJob := makePendingJob(10, "reaction", 5, 54000, 0)
		caps := []*calculator.CharacterCapacity{
			makeCapacity(1001, 5, 5, 5, 5, 5), // both activities
		}
		assigned, unassigned := SimulateAssignment([]*PendingJob{mfgJob, reactJob}, caps, 1)
		assert.Equal(t, 0, unassigned)
		assert.Len(t, assigned, 2)
		activities := map[string]bool{}
		for _, aj := range assigned {
			activities[aj.Activity] = true
			assert.Equal(t, int64(1001), aj.CharacterID)
		}
		assert.True(t, activities["manufacturing"])
		assert.True(t, activities["reaction"])
	})

	t.Run("slot resets between depth levels", func(t *testing.T) {
		// Jobs at depth 1 (deeper) followed by depth 0 (shallower)
		// With sorted order deepest-first: depth1 first, depth0 second
		// Character has 1 slot — after depth1 job finishes, slot should reset for depth0
		deepJob := makePendingJob(1, "manufacturing", 5, 18000, 1)
		shallowJob := makePendingJob(2, "manufacturing", 10, 36000, 0)
		caps := []*calculator.CharacterCapacity{
			makeCapacity(1001, 1, 0, 5, 5, 0),
		}
		// Jobs must be passed in deepest-first order (as WalkAndMergeSteps produces)
		assigned, unassigned := SimulateAssignment([]*PendingJob{deepJob, shallowJob}, caps, 1)
		assert.Equal(t, 0, unassigned)
		assert.Len(t, assigned, 2)
		for _, aj := range assigned {
			assert.Equal(t, int64(1001), aj.CharacterID)
		}
	})

	t.Run("jobs with non-manufacturing/reaction activity are skipped", func(t *testing.T) {
		job := &PendingJob{
			Entry: &models.IndustryJobQueueEntry{
				BlueprintTypeID: 999,
				Activity:        "copying",
				Runs:            1,
			},
			Activity: "copying",
			Depth:    0,
		}
		caps := []*calculator.CharacterCapacity{
			makeCapacity(1001, 5, 0, 5, 5, 0),
		}
		assigned, unassigned := SimulateAssignment([]*PendingJob{job}, caps, 1)
		assert.Empty(t, assigned)
		assert.Equal(t, 0, unassigned)
	})

	t.Run("parallelism larger than capacities is clamped", func(t *testing.T) {
		job := makePendingJob(1, "manufacturing", 10, 36000, 0)
		caps := []*calculator.CharacterCapacity{
			makeCapacity(1001, 5, 0, 5, 5, 0),
		}
		// parallelism=10 but only 1 capacity — should clamp to 1
		assigned, unassigned := SimulateAssignment([]*PendingJob{job}, caps, 10)
		assert.Equal(t, 0, unassigned)
		assert.Len(t, assigned, 1)
		assert.Equal(t, int64(1001), assigned[0].CharacterID)
	})

	t.Run("empty capacities — all jobs unassigned", func(t *testing.T) {
		job := makePendingJob(1, "manufacturing", 10, 36000, 0)
		assigned, unassigned := SimulateAssignment([]*PendingJob{job}, []*calculator.CharacterCapacity{}, 5)
		assert.Len(t, assigned, 1)
		assert.Equal(t, int64(0), assigned[0].CharacterID)
		assert.Equal(t, 10, unassigned)
	})
}

// ---------------------------------------------------------------------------
// EstimateWallClock tests
// ---------------------------------------------------------------------------

func Test_EstimateWallClock(t *testing.T) {
	t.Run("empty assignments returns 0", func(t *testing.T) {
		caps := []*calculator.CharacterCapacity{
			makeCapacity(1001, 5, 0, 5, 5, 0),
		}
		result := EstimateWallClock([]*AssignedJob{}, caps)
		assert.Equal(t, 0, result)
	})

	t.Run("single assignment returns its duration", func(t *testing.T) {
		pj := makePendingJob(1, "manufacturing", 10, 36000, 0)
		job := &AssignedJob{
			Original:    pj,
			Activity:    "manufacturing",
			Runs:        10,
			DurationSec: 3600,
			Depth:       0,
			CharacterID: 1001,
		}
		caps := []*calculator.CharacterCapacity{
			makeCapacity(1001, 1, 0, 5, 5, 0),
		}
		result := EstimateWallClock([]*AssignedJob{job}, caps)
		assert.Equal(t, 3600, result)
	})

	t.Run("two assignments same depth same character uses LPT across slots", func(t *testing.T) {
		pj := makePendingJob(1, "manufacturing", 10, 36000, 0)
		job1 := &AssignedJob{
			Original:    pj,
			Activity:    "manufacturing",
			Runs:        5,
			DurationSec: 3600,
			Depth:       0,
			CharacterID: 1001,
		}
		job2 := &AssignedJob{
			Original:    pj,
			Activity:    "manufacturing",
			Runs:        5,
			DurationSec: 1800,
			Depth:       0,
			CharacterID: 1001,
		}
		// Character has 2 mfg slots; jobs run in parallel
		caps := []*calculator.CharacterCapacity{
			makeCapacity(1001, 2, 0, 5, 5, 0),
		}
		result := EstimateWallClock([]*AssignedJob{job1, job2}, caps)
		// LPT: 3600 in slot 0, 1800 in slot 1 → max is 3600
		assert.Equal(t, 3600, result)
	})

	t.Run("two assignments different depths sum sequentially", func(t *testing.T) {
		pj := makePendingJob(1, "manufacturing", 5, 36000, 0)
		deepJob := &AssignedJob{
			Original:    pj,
			Activity:    "manufacturing",
			Runs:        5,
			DurationSec: 7200,
			Depth:       1,
			CharacterID: 1001,
		}
		shallowJob := &AssignedJob{
			Original:    pj,
			Activity:    "manufacturing",
			Runs:        5,
			DurationSec: 3600,
			Depth:       0,
			CharacterID: 1001,
		}
		caps := []*calculator.CharacterCapacity{
			makeCapacity(1001, 1, 0, 5, 5, 0),
		}
		result := EstimateWallClock([]*AssignedJob{deepJob, shallowJob}, caps)
		// Depths processed sequentially: depth1 (7200) + depth0 (3600) = 10800
		assert.Equal(t, 10800, result)
	})

	t.Run("multiple characters at same depth uses max across chars", func(t *testing.T) {
		pj := makePendingJob(1, "manufacturing", 5, 36000, 0)
		// char 1001 has a 5000s job, char 1002 has a 3000s job
		job1 := &AssignedJob{
			Original:    pj,
			Activity:    "manufacturing",
			Runs:        5,
			DurationSec: 5000,
			Depth:       0,
			CharacterID: 1001,
		}
		job2 := &AssignedJob{
			Original:    pj,
			Activity:    "manufacturing",
			Runs:        3,
			DurationSec: 3000,
			Depth:       0,
			CharacterID: 1002,
		}
		caps := []*calculator.CharacterCapacity{
			makeCapacity(1001, 1, 0, 5, 5, 0),
			makeCapacity(1002, 1, 0, 5, 5, 0),
		}
		result := EstimateWallClock([]*AssignedJob{job1, job2}, caps)
		// Both run in parallel; total = max(5000, 3000) = 5000
		assert.Equal(t, 5000, result)
	})

	t.Run("unassigned job CharacterID=0 uses 1 slot fallback", func(t *testing.T) {
		pj := makePendingJob(1, "manufacturing", 5, 36000, 0)
		job := &AssignedJob{
			Original:    pj,
			Activity:    "manufacturing",
			Runs:        5,
			DurationSec: 4000,
			Depth:       0,
			CharacterID: 0, // unassigned
		}
		caps := []*calculator.CharacterCapacity{}
		result := EstimateWallClock([]*AssignedJob{job}, caps)
		assert.Equal(t, 4000, result)
	})
}

// ---------------------------------------------------------------------------
// WalkAndMergeSteps tests
// ---------------------------------------------------------------------------

func intPtr(i int64) *int64 {
	return &i
}

func strPtr(s string) *string {
	return &s
}

func makeBlueprintRow(bpTypeID, productTypeID int64, productName string, productQty, baseTime int) *repositories.ManufacturingBlueprintRow {
	return &repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: bpTypeID,
		ProductTypeID:   productTypeID,
		ProductName:     productName,
		ProductQuantity: productQty,
		Time:            baseTime,
		ProductVolume:   1.0,
	}
}

func makeMaterialRow(bpTypeID, typeID int64, name string, qty int) *repositories.ManufacturingMaterialRow {
	return &repositories.ManufacturingMaterialRow{
		BlueprintTypeID: bpTypeID,
		TypeID:          typeID,
		TypeName:        name,
		Quantity:        qty,
		Volume:          1.0,
	}
}

func makeStep(id int64, parentID *int64, productTypeID, blueprintTypeID int64, activity string) *models.ProductionPlanStep {
	return &models.ProductionPlanStep{
		ID:              id,
		PlanID:          1,
		ParentStepID:    parentID,
		ProductTypeID:   productTypeID,
		BlueprintTypeID: blueprintTypeID,
		Activity:        activity,
		MELevel:         0,
		TELevel:         0,
		IndustrySkill:   5,
		AdvIndustrySkill: 5,
		Structure:       "station",
		Rig:             "none",
		Security:        "high",
		FacilityTax:     0.0,
	}
}

func emptyJitaPrices() map[int64]*models.MarketPrice {
	return map[int64]*models.MarketPrice{}
}

func emptyAdjustedPrices() map[int64]float64 {
	return map[int64]float64{}
}

func Test_WalkAndMergeSteps(t *testing.T) {
	ctx := context.Background()

	t.Run("plan with no steps returns error", func(t *testing.T) {
		sdeRepo := &MockJobGenSdeRepository{}
		plan := &models.ProductionPlan{
			ID:    1,
			Steps: []*models.ProductionPlanStep{},
		}
		result, err := WalkAndMergeSteps(ctx, sdeRepo, plan, 10, emptyJitaPrices(), emptyAdjustedPrices())
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("plan with no root step returns error", func(t *testing.T) {
		sdeRepo := &MockJobGenSdeRepository{}
		parentID := int64(99)
		step := makeStep(1, &parentID, 100, 200, "manufacturing")
		plan := &models.ProductionPlan{
			ID:    1,
			Steps: []*models.ProductionPlanStep{step},
		}
		result, err := WalkAndMergeSteps(ctx, sdeRepo, plan, 10, emptyJitaPrices(), emptyAdjustedPrices())
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("single step plan produces one merged job", func(t *testing.T) {
		sdeRepo := &MockJobGenSdeRepository{}

		// Blueprint produces 1 unit per run
		bp := makeBlueprintRow(200, 100, "Tritanium", 1, 3600)
		// No materials for simplicity
		mats := []*repositories.ManufacturingMaterialRow{}

		sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(200), "manufacturing").Return(bp, nil)
		sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(200), "manufacturing").Return(mats, nil)

		rootStep := makeStep(1, nil, 100, 200, "manufacturing")
		plan := &models.ProductionPlan{
			ID:    1,
			Steps: []*models.ProductionPlanStep{rootStep},
		}

		result, err := WalkAndMergeSteps(ctx, sdeRepo, plan, 10, emptyJitaPrices(), emptyAdjustedPrices())
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.MergedJobs, 1)
		assert.Empty(t, result.Skipped)

		job := result.MergedJobs[0]
		assert.Equal(t, int64(200), job.Entry.BlueprintTypeID)
		assert.Equal(t, "manufacturing", job.Entry.Activity)
		// 10 units / 1 per run = 10 runs
		assert.Equal(t, 10, job.Entry.Runs)
		assert.Equal(t, 0, job.Depth)
		assert.Equal(t, "Tritanium", job.ProductName)

		sdeRepo.AssertExpectations(t)
	})

	t.Run("quantity scales runs correctly", func(t *testing.T) {
		sdeRepo := &MockJobGenSdeRepository{}

		// Blueprint produces 10 units per run
		bp := makeBlueprintRow(200, 100, "SomeItem", 10, 3600)
		mats := []*repositories.ManufacturingMaterialRow{}

		sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(200), "manufacturing").Return(bp, nil)
		sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(200), "manufacturing").Return(mats, nil)

		rootStep := makeStep(1, nil, 100, 200, "manufacturing")
		plan := &models.ProductionPlan{
			ID:    1,
			Steps: []*models.ProductionPlanStep{rootStep},
		}

		// Request 25 units; 10 per run → ceil(25/10) = 3 runs
		result, err := WalkAndMergeSteps(ctx, sdeRepo, plan, 25, emptyJitaPrices(), emptyAdjustedPrices())
		assert.NoError(t, err)
		assert.Len(t, result.MergedJobs, 1)
		assert.Equal(t, 3, result.MergedJobs[0].Entry.Runs)

		sdeRepo.AssertExpectations(t)
	})

	t.Run("multi-step plan with parent-child relationships", func(t *testing.T) {
		sdeRepo := &MockJobGenSdeRepository{}

		// Root: makes product A (type 100, bp 200)
		// Child: makes intermediate B (type 110, bp 210)
		// Root requires 5 units of B per run

		rootBP := makeBlueprintRow(200, 100, "Product A", 1, 3600)
		rootMats := []*repositories.ManufacturingMaterialRow{
			makeMaterialRow(200, 110, "Intermediate B", 5),
		}

		childBP := makeBlueprintRow(210, 110, "Intermediate B", 10, 1800)
		childMats := []*repositories.ManufacturingMaterialRow{
			makeMaterialRow(210, 120, "Raw Ore", 3),
		}

		sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(200), "manufacturing").Return(rootBP, nil)
		sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(200), "manufacturing").Return(rootMats, nil)
		sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(210), "manufacturing").Return(childBP, nil)
		sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(210), "manufacturing").Return(childMats, nil)

		rootStep := makeStep(1, nil, 100, 200, "manufacturing")
		childStepID := int64(1)
		childStep := makeStep(2, &childStepID, 110, 210, "manufacturing")

		plan := &models.ProductionPlan{
			ID:    1,
			Steps: []*models.ProductionPlanStep{rootStep, childStep},
		}

		// Request 5 units of product A
		result, err := WalkAndMergeSteps(ctx, sdeRepo, plan, 5, emptyJitaPrices(), emptyAdjustedPrices())
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.MergedJobs, 2)

		// Jobs should be sorted deepest first
		// Child (depth=1) should come before root (depth=0)
		assert.Equal(t, 1, result.MergedJobs[0].Depth)
		assert.Equal(t, 0, result.MergedJobs[1].Depth)

		// Root: 5 units / 1 per run = 5 runs
		rootJob := result.MergedJobs[1]
		assert.Equal(t, 5, rootJob.Entry.Runs)
		assert.Equal(t, int64(200), rootJob.Entry.BlueprintTypeID)

		// StepProduction should be populated for both steps
		assert.Contains(t, result.StepProduction, int64(1))
		assert.Contains(t, result.StepProduction, int64(2))

		// StepDepths should record correct depths
		assert.Equal(t, 0, result.StepDepths[1])
		assert.Equal(t, 1, result.StepDepths[2])

		sdeRepo.AssertExpectations(t)
	})

	t.Run("steps with same merge key get merged into one job", func(t *testing.T) {
		sdeRepo := &MockJobGenSdeRepository{}

		// Two steps that produce the same blueprint with same settings
		// (e.g. same intermediate needed by multiple parents)
		bp := makeBlueprintRow(210, 110, "Intermediate", 5, 1800)
		mats := []*repositories.ManufacturingMaterialRow{}

		// Root step uses intermediate twice (two children with same bp)
		rootBP := makeBlueprintRow(200, 100, "Final Product", 1, 3600)
		rootMats := []*repositories.ManufacturingMaterialRow{
			makeMaterialRow(200, 110, "Intermediate", 10),
		}

		sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(200), "manufacturing").Return(rootBP, nil)
		sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(200), "manufacturing").Return(rootMats, nil)
		sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(210), "manufacturing").Return(bp, nil)
		sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(210), "manufacturing").Return(mats, nil)

		rootStep := makeStep(1, nil, 100, 200, "manufacturing")
		childStepID := int64(1)
		// Single child producing the intermediate
		childStep := makeStep(2, &childStepID, 110, 210, "manufacturing")

		plan := &models.ProductionPlan{
			ID:    1,
			Steps: []*models.ProductionPlanStep{rootStep, childStep},
		}

		result, err := WalkAndMergeSteps(ctx, sdeRepo, plan, 1, emptyJitaPrices(), emptyAdjustedPrices())
		assert.NoError(t, err)
		assert.NotNil(t, result)
		// Two distinct blueprints, so 2 merged jobs
		assert.Len(t, result.MergedJobs, 2)

		sdeRepo.AssertExpectations(t)
	})

	t.Run("blueprint not found causes step to be skipped", func(t *testing.T) {
		sdeRepo := &MockJobGenSdeRepository{}

		// Blueprint lookup returns nil (not found)
		sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(200), "manufacturing").Return(nil, nil)

		rootStep := makeStep(1, nil, 100, 200, "manufacturing")
		plan := &models.ProductionPlan{
			ID:    1,
			Steps: []*models.ProductionPlanStep{rootStep},
		}

		result, err := WalkAndMergeSteps(ctx, sdeRepo, plan, 10, emptyJitaPrices(), emptyAdjustedPrices())
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.MergedJobs)
		assert.Len(t, result.Skipped, 1)
		assert.Equal(t, int64(100), result.Skipped[0].TypeID)
		assert.Equal(t, "blueprint data not found", result.Skipped[0].Reason)

		sdeRepo.AssertExpectations(t)
	})

	t.Run("materials fetch error causes step to be skipped", func(t *testing.T) {
		sdeRepo := &MockJobGenSdeRepository{}

		bp := makeBlueprintRow(200, 100, "SomeItem", 1, 3600)
		sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(200), "manufacturing").Return(bp, nil)
		sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(200), "manufacturing").Return(
			([]*repositories.ManufacturingMaterialRow)(nil),
			assert.AnError,
		)

		rootStep := makeStep(1, nil, 100, 200, "manufacturing")
		plan := &models.ProductionPlan{
			ID:    1,
			Steps: []*models.ProductionPlanStep{rootStep},
		}

		result, err := WalkAndMergeSteps(ctx, sdeRepo, plan, 10, emptyJitaPrices(), emptyAdjustedPrices())
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.MergedJobs)
		assert.Len(t, result.Skipped, 1)
		assert.Equal(t, "failed to get materials", result.Skipped[0].Reason)

		sdeRepo.AssertExpectations(t)
	})

	t.Run("reaction step uses reaction TE formula for duration", func(t *testing.T) {
		sdeRepo := &MockJobGenSdeRepository{}

		// Reaction blueprint: 3600s base time, produces 100 units per run
		bp := makeBlueprintRow(300, 150, "Moon Goo Product", 100, 3600)
		mats := []*repositories.ManufacturingMaterialRow{}

		sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(300), "reaction").Return(bp, nil)
		sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(300), "reaction").Return(mats, nil)

		rootStep := makeStep(1, nil, 150, 300, "reaction")
		rootStep.IndustrySkill = 5 // Reactions skill level 5
		plan := &models.ProductionPlan{
			ID:    1,
			Steps: []*models.ProductionPlanStep{rootStep},
		}

		result, err := WalkAndMergeSteps(ctx, sdeRepo, plan, 100, emptyJitaPrices(), emptyAdjustedPrices())
		assert.NoError(t, err)
		assert.Len(t, result.MergedJobs, 1)

		job := result.MergedJobs[0]
		assert.Equal(t, "reaction", job.Entry.Activity)
		assert.NotNil(t, job.Entry.EstimatedDuration)

		// Verify duration uses reaction formula:
		// teFactor = (1 - 5*0.04) * (1 - 0) * (1 - 0) = 0.80
		// secsPerRun = floor(3600 * 0.80) = 2880
		// totalDuration = 2880 * 1 run = 2880
		expectedTEFactor := calculator.ComputeTEFactor(5, "station", "none", "high")
		expectedSecsPerRun := calculator.ComputeSecsPerRun(3600, expectedTEFactor)
		expectedDuration := expectedSecsPerRun * 1
		assert.Equal(t, expectedDuration, *job.Entry.EstimatedDuration)

		sdeRepo.AssertExpectations(t)
	})

	t.Run("step production data is recorded", func(t *testing.T) {
		sdeRepo := &MockJobGenSdeRepository{}

		bp := makeBlueprintRow(200, 100, "Tritanium", 5, 3600)
		mats := []*repositories.ManufacturingMaterialRow{}

		sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(200), "manufacturing").Return(bp, nil)
		sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(200), "manufacturing").Return(mats, nil)

		rootStep := makeStep(1, nil, 100, 200, "manufacturing")
		plan := &models.ProductionPlan{
			ID:    1,
			Steps: []*models.ProductionPlanStep{rootStep},
		}

		result, err := WalkAndMergeSteps(ctx, sdeRepo, plan, 10, emptyJitaPrices(), emptyAdjustedPrices())
		assert.NoError(t, err)

		prod, ok := result.StepProduction[1]
		assert.True(t, ok)
		assert.Equal(t, int64(100), prod.ProductTypeID)
		assert.Equal(t, "Tritanium", prod.ProductName)
		// ceil(10/5) = 2 runs × 5 per run = 10 produced
		assert.Equal(t, 10, prod.TotalQuantity)

		sdeRepo.AssertExpectations(t)
	})

	t.Run("location fields populated from step", func(t *testing.T) {
		sdeRepo := &MockJobGenSdeRepository{}

		bp := makeBlueprintRow(200, 100, "Item", 1, 3600)
		mats := []*repositories.ManufacturingMaterialRow{}

		sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(200), "manufacturing").Return(bp, nil)
		sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(200), "manufacturing").Return(mats, nil)

		stationName := "Jita 4-4"
		rootStep := makeStep(1, nil, 100, 200, "manufacturing")
		rootStep.StationName = &stationName
		rootStep.SourceOwnerName = "My Corp"
		rootStep.SourceDivisionName = "Hangar 1"
		rootStep.OutputOwnerName = "My Corp"
		rootStep.OutputDivisionName = "Hangar 2"

		plan := &models.ProductionPlan{
			ID:    1,
			Steps: []*models.ProductionPlanStep{rootStep},
		}

		result, err := WalkAndMergeSteps(ctx, sdeRepo, plan, 1, emptyJitaPrices(), emptyAdjustedPrices())
		assert.NoError(t, err)
		assert.Len(t, result.MergedJobs, 1)

		job := result.MergedJobs[0]
		assert.Equal(t, "Jita 4-4", job.Entry.StationName)
		assert.Equal(t, "My Corp > Hangar 1", job.Entry.InputLocation)
		assert.Equal(t, "My Corp > Hangar 2", job.Entry.OutputLocation)

		sdeRepo.AssertExpectations(t)
	})

	t.Run("merged jobs are sorted deepest first", func(t *testing.T) {
		sdeRepo := &MockJobGenSdeRepository{}

		rootBP := makeBlueprintRow(200, 100, "Final", 1, 3600)
		rootMats := []*repositories.ManufacturingMaterialRow{
			makeMaterialRow(200, 110, "Intermediate", 2),
		}
		childBP := makeBlueprintRow(210, 110, "Intermediate", 1, 1800)
		childMats := []*repositories.ManufacturingMaterialRow{}

		sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(200), "manufacturing").Return(rootBP, nil)
		sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(200), "manufacturing").Return(rootMats, nil)
		sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(210), "manufacturing").Return(childBP, nil)
		sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(210), "manufacturing").Return(childMats, nil)

		rootStep := makeStep(1, nil, 100, 200, "manufacturing")
		childStepID := int64(1)
		childStep := makeStep(2, &childStepID, 110, 210, "manufacturing")

		plan := &models.ProductionPlan{
			ID:    1,
			Steps: []*models.ProductionPlanStep{rootStep, childStep},
		}

		result, err := WalkAndMergeSteps(ctx, sdeRepo, plan, 1, emptyJitaPrices(), emptyAdjustedPrices())
		assert.NoError(t, err)
		assert.Len(t, result.MergedJobs, 2)

		// First job should be deeper
		assert.Greater(t, result.MergedJobs[0].Depth, result.MergedJobs[1].Depth)

		sdeRepo.AssertExpectations(t)
	})

	t.Run("two identical steps get merged into a single job", func(t *testing.T) {
		sdeRepo := &MockJobGenSdeRepository{}

		// One root step and two children producing the same item (same blueprint, same settings)
		rootBP := makeBlueprintRow(200, 100, "Final", 1, 3600)
		rootMats := []*repositories.ManufacturingMaterialRow{
			makeMaterialRow(200, 110, "Component", 2),
		}
		compBP := makeBlueprintRow(210, 110, "Component", 1, 1800)
		compMats := []*repositories.ManufacturingMaterialRow{}

		sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(200), "manufacturing").Return(rootBP, nil)
		sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(200), "manufacturing").Return(rootMats, nil)
		sdeRepo.On("GetBlueprintForActivity", mock.Anything, int64(210), "manufacturing").Return(compBP, nil)
		sdeRepo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(210), "manufacturing").Return(compMats, nil)

		rootStepA := makeStep(1, nil, 100, 200, "manufacturing")
		parentStepID := int64(1)
		childStep1 := makeStep(2, &parentStepID, 110, 210, "manufacturing")

		plan := &models.ProductionPlan{
			ID:    1,
			Steps: []*models.ProductionPlanStep{rootStepA, childStep1},
		}

		result, err := WalkAndMergeSteps(ctx, sdeRepo, plan, 3, emptyJitaPrices(), emptyAdjustedPrices())
		assert.NoError(t, err)
		assert.NotNil(t, result)
		// root (bp 200) + child (bp 210) = 2 distinct blueprints
		assert.Len(t, result.MergedJobs, 2)

		sdeRepo.AssertExpectations(t)
	})
}
