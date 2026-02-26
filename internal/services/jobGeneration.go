package services

import (
	"context"
	"math"
	"sort"
	"strings"

	"github.com/annymsMthd/industry-tool/internal/calculator"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/pkg/errors"
)

// JobGenSdeRepository is the SDE interface required by WalkAndMergeSteps.
type JobGenSdeRepository interface {
	GetBlueprintForActivity(ctx context.Context, blueprintTypeID int64, activity string) (*repositories.ManufacturingBlueprintRow, error)
	GetBlueprintMaterialsForActivity(ctx context.Context, blueprintTypeID int64, activity string) ([]*repositories.ManufacturingMaterialRow, error)
}

// StepProductionData holds production output data for a single plan step,
// used when generating transport jobs.
type StepProductionData struct {
	ProductTypeID int64
	ProductName   string
	TotalQuantity int
	ProductVolume float64
}

// PendingJob is a job ready to be persisted or simulated, with tree metadata.
type PendingJob struct {
	Entry         *models.IndustryJobQueueEntry
	BlueprintName string
	ProductName   string
	Depth         int
	// Fields for per-character TE recalculation in preview
	BaseBlueprintTime int    // base seconds from SDE blueprint
	Activity          string // "manufacturing" or "reaction"
	Structure         string
	Rig               string
	Security          string
	BlueprintTE       int
}

// MergeKey identifies jobs that can be merged (same blueprint, settings).
type MergeKey struct {
	BlueprintTypeID int64
	Activity        string
	MELevel         int
	TELevel         int
	FacilityTax     float64
}

// WalkResult is the output of WalkAndMergeSteps.
type WalkResult struct {
	MergedJobs     []*PendingJob
	StepProduction map[int64]*StepProductionData
	StepDepths     map[int64]int
	Skipped        []*models.GenerateJobSkipped
}

// AssignedJob is a single job fragment assigned to a character during simulation.
type AssignedJob struct {
	Original    *PendingJob // reference to the originating merged job
	Activity    string
	Runs        int
	DurationSec int
	Depth       int
	CharacterID int64
}

// GetCharacterID is a helper accessor for AssignedJob to work with map grouping.
// CharacterID is also a field name, so the accessor is named GetCharacterID.
func (aj *AssignedJob) GetCharacterID() int64 {
	return aj.CharacterID
}

// FormatLocation builds a human-readable location string from owner/division/container names.
func FormatLocation(ownerName, divisionName, containerName string) string {
	parts := []string{}
	if ownerName != "" {
		parts = append(parts, ownerName)
	}
	if divisionName != "" {
		parts = append(parts, divisionName)
	}
	if containerName != "" {
		parts = append(parts, containerName)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " > ")
}

// WalkAndMergeSteps walks the production plan tree and returns merged pending jobs.
// It is shared by GenerateJobs and PreviewPlan.
func WalkAndMergeSteps(
	ctx context.Context,
	sdeRepo JobGenSdeRepository,
	plan *models.ProductionPlan,
	quantity int,
	jitaPrices map[int64]*models.MarketPrice,
	adjustedPrices map[int64]float64,
) (*WalkResult, error) {
	if len(plan.Steps) == 0 {
		return nil, errors.New("plan has no steps")
	}

	// Build step index and find root
	stepsByID := make(map[int64]*models.ProductionPlanStep)
	childStepsByParent := make(map[int64][]*models.ProductionPlanStep)
	var rootStep *models.ProductionPlanStep

	for _, step := range plan.Steps {
		stepsByID[step.ID] = step
		if step.ParentStepID == nil {
			rootStep = step
		} else {
			childStepsByParent[*step.ParentStepID] = append(childStepsByParent[*step.ParentStepID], step)
		}
	}

	if rootStep == nil {
		return nil, errors.New("plan has no root step")
	}

	wr := &WalkResult{
		StepProduction: make(map[int64]*StepProductionData),
		StepDepths:     make(map[int64]int),
		Skipped:        []*models.GenerateJobSkipped{},
	}

	pendingJobs := []*PendingJob{}

	// Walk the tree depth-first, collect jobs by depth level
	var walkStep func(step *models.ProductionPlanStep, qty int, depth int)
	walkStep = func(step *models.ProductionPlanStep, qty int, depth int) {
		// Record depth for this step (used by transport ordering)
		wr.StepDepths[step.ID] = depth

		// Look up blueprint to get product quantity per run (activity-aware)
		bp, err := sdeRepo.GetBlueprintForActivity(ctx, step.BlueprintTypeID, step.Activity)
		if err != nil || bp == nil {
			wr.Skipped = append(wr.Skipped, &models.GenerateJobSkipped{
				TypeID:   step.ProductTypeID,
				TypeName: step.ProductName,
				Reason:   "blueprint data not found",
			})
			return
		}

		// Calculate runs needed
		runs := int(math.Ceil(float64(qty) / float64(bp.ProductQuantity)))
		if runs <= 0 {
			runs = 1
		}

		// Record production data for transport generation
		totalProduced := runs * bp.ProductQuantity
		wr.StepProduction[step.ID] = &StepProductionData{
			ProductTypeID: step.ProductTypeID,
			ProductName:   bp.ProductName,
			TotalQuantity: totalProduced,
			ProductVolume: bp.ProductVolume,
		}

		// Get materials for this step to calculate child needs
		materials, err := sdeRepo.GetBlueprintMaterialsForActivity(ctx, step.BlueprintTypeID, step.Activity)
		if err != nil {
			wr.Skipped = append(wr.Skipped, &models.GenerateJobSkipped{
				TypeID:   step.ProductTypeID,
				TypeName: step.ProductName,
				Reason:   "failed to get materials",
			})
			return
		}

		// Calculate ME factor for batch quantity calculation
		meFactor := calculator.ComputeManufacturingME(step.MELevel, step.Structure, step.Rig, step.Security)

		// Process child steps (materials that are produced)
		children := childStepsByParent[step.ID]
		childProductTypeIDs := make(map[int64]*models.ProductionPlanStep)
		for _, child := range children {
			childProductTypeIDs[child.ProductTypeID] = child
		}

		for _, mat := range materials {
			if childStep, ok := childProductTypeIDs[mat.TypeID]; ok {
				// This material is produced — calculate needed quantity
				batchQty := calculator.ComputeBatchQty(runs, mat.Quantity, meFactor)
				walkStep(childStep, int(batchQty), depth+1)
			}
		}

		// Calculate cost and duration for manufacturing and reaction steps
		var estimatedCost *float64
		var estimatedDuration *int

		if step.Activity == "manufacturing" || step.Activity == "reaction" {
			params := &calculator.ManufacturingParams{
				BlueprintME:      step.MELevel,
				BlueprintTE:      step.TELevel,
				Runs:             runs,
				Structure:        step.Structure,
				Rig:              step.Rig,
				Security:         step.Security,
				IndustrySkill:    step.IndustrySkill,
				AdvIndustrySkill: step.AdvIndustrySkill,
				FacilityTax:      step.FacilityTax,
			}

			data := &calculator.ManufacturingData{
				Blueprint:      bp,
				Materials:      materials,
				CostIndex:      0,
				AdjustedPrices: adjustedPrices,
				JitaPrices:     jitaPrices,
			}

			calcResult := calculator.CalculateManufacturingJob(params, data)
			estimatedCost = &calcResult.TotalCost
			estimatedDuration = &calcResult.TotalDuration

			// For reaction steps, override the duration using the reactions TE formula.
			// Reactions use only the Reactions skill (4% per level) with no blueprint TE
			// and no Advanced Industry skill reduction. step.IndustrySkill holds the
			// Reactions skill level when activity == "reaction".
			if step.Activity == "reaction" {
				reactionTEFactor := calculator.ComputeTEFactor(step.IndustrySkill, step.Structure, step.Rig, step.Security)
				secsPerRun := calculator.ComputeSecsPerRun(bp.Time, reactionTEFactor)
				totalDuration := secsPerRun * runs
				estimatedDuration = &totalDuration
			}
		}

		// Build location context from plan step
		stationName := ""
		if step.StationName != nil {
			stationName = *step.StationName
		}
		inputLoc := FormatLocation(step.SourceOwnerName, step.SourceDivisionName, step.SourceContainerName)
		outputLoc := FormatLocation(step.OutputOwnerName, step.OutputDivisionName, step.OutputContainerName)

		productTypeID := step.ProductTypeID
		pendingJobs = append(pendingJobs, &PendingJob{
			Entry: &models.IndustryJobQueueEntry{
				BlueprintTypeID:   step.BlueprintTypeID,
				Activity:          step.Activity,
				Runs:              runs,
				MELevel:           step.MELevel,
				TELevel:           step.TELevel,
				FacilityTax:       step.FacilityTax,
				ProductTypeID:     &productTypeID,
				EstimatedCost:     estimatedCost,
				EstimatedDuration: estimatedDuration,
				SortOrder:         depth * 2,
				StationName:       stationName,
				InputLocation:     inputLoc,
				OutputLocation:    outputLoc,
			},
			BlueprintName:     step.BlueprintName,
			ProductName:       bp.ProductName,
			Depth:             depth,
			BaseBlueprintTime: bp.Time,
			Activity:          step.Activity,
			Structure:         step.Structure,
			Rig:               step.Rig,
			Security:          step.Security,
			BlueprintTE:       step.TELevel,
		})
	}

	walkStep(rootStep, quantity, 0)

	// Merge pending jobs with identical blueprint + settings into combined entries
	merged := make(map[MergeKey]*PendingJob)
	mergeOrder := []MergeKey{}
	for _, pj := range pendingJobs {
		key := MergeKey{
			BlueprintTypeID: pj.Entry.BlueprintTypeID,
			Activity:        pj.Entry.Activity,
			MELevel:         pj.Entry.MELevel,
			TELevel:         pj.Entry.TELevel,
			FacilityTax:     pj.Entry.FacilityTax,
		}
		if existing, ok := merged[key]; ok {
			existing.Entry.Runs += pj.Entry.Runs
			if pj.Entry.EstimatedCost != nil {
				if existing.Entry.EstimatedCost == nil {
					cost := *pj.Entry.EstimatedCost
					existing.Entry.EstimatedCost = &cost
				} else {
					*existing.Entry.EstimatedCost += *pj.Entry.EstimatedCost
				}
			}
			if pj.Entry.EstimatedDuration != nil {
				if existing.Entry.EstimatedDuration == nil {
					dur := *pj.Entry.EstimatedDuration
					existing.Entry.EstimatedDuration = &dur
				} else {
					*existing.Entry.EstimatedDuration += *pj.Entry.EstimatedDuration
				}
			}
			// Keep the deeper depth for ordering
			if pj.Depth > existing.Depth {
				existing.Depth = pj.Depth
				existing.Entry.SortOrder = pj.Entry.SortOrder
			}
		} else {
			merged[key] = pj
			mergeOrder = append(mergeOrder, key)
		}
	}

	// Collect merged jobs preserving insertion order
	mergedJobs := make([]*PendingJob, 0, len(mergeOrder))
	for _, key := range mergeOrder {
		mergedJobs = append(mergedJobs, merged[key])
	}

	// Sort merged jobs by depth descending (deepest first = leaves before parents)
	sort.Slice(mergedJobs, func(i, j int) bool {
		return mergedJobs[i].Depth > mergedJobs[j].Depth
	})

	wr.MergedJobs = mergedJobs
	return wr, nil
}

// SimulateAssignment distributes merged jobs across up to parallelism characters.
// It clones capacity state so the originals are not mutated.
// Returns the list of assigned job fragments and count of unassigned runs.
//
// Jobs with no eligible character are still appended to the returned slice with
// CharacterID=0 so they are not silently dropped from queue creation.
//
// Slot availability resets at each depth level: children must finish before
// parents start, so a slot used at depth N is free again at depth N-1.
func SimulateAssignment(
	mergedJobs []*PendingJob,
	capacities []*calculator.CharacterCapacity,
	parallelism int,
) ([]*AssignedJob, int) {
	if parallelism > len(capacities) {
		parallelism = len(capacities)
	}
	pool := capacities[:parallelism]

	// Record initial available slots (capacity minus already-running ESI jobs).
	// These are reset whenever the depth level changes.
	initialMfg := make([]int, parallelism)
	initialReact := make([]int, parallelism)
	for i, cap := range pool {
		initialMfg[i] = calculator.MfgSlotsAvailable(cap)
		initialReact[i] = calculator.ReactSlotsAvailable(cap)
	}

	// Working copies reset at each new depth level.
	mfgAvail := make([]int, parallelism)
	reactAvail := make([]int, parallelism)
	copy(mfgAvail, initialMfg)
	copy(reactAvail, initialReact)

	assigned := []*AssignedJob{}
	unassigned := 0
	currentDepth := -1

	for _, pj := range mergedJobs {
		// Jobs are sorted deepest-first. When we move to a shallower depth, the
		// previous depth's jobs are done and all slots are free again.
		jobDepth := pj.Depth
		if currentDepth == -1 {
			currentDepth = jobDepth
		} else if jobDepth != currentDepth {
			currentDepth = jobDepth
			copy(mfgAvail, initialMfg)
			copy(reactAvail, initialReact)
		}

		activity := pj.Activity
		if activity == "" {
			activity = pj.Entry.Activity
		}
		if activity != "manufacturing" && activity != "reaction" {
			continue
		}

		// Collect characters with available slots for this activity
		type eligibleChar struct {
			idx int
			cap *calculator.CharacterCapacity
		}
		eligible := []eligibleChar{}
		for i, cap := range pool {
			if activity == "manufacturing" && mfgAvail[i] > 0 {
				eligible = append(eligible, eligibleChar{idx: i, cap: cap})
			} else if activity == "reaction" && reactAvail[i] > 0 {
				eligible = append(eligible, eligibleChar{idx: i, cap: cap})
			}
		}

		if len(eligible) == 0 {
			// No slot available — still persist the job so it appears in the queue
			// without a character assignment (CharacterID=0 signals unassigned).
			var dur int
			if pj.Entry.EstimatedDuration != nil {
				dur = *pj.Entry.EstimatedDuration
			}
			assigned = append(assigned, &AssignedJob{
				Original:    pj,
				Activity:    activity,
				Runs:        pj.Entry.Runs,
				DurationSec: dur,
				Depth:       jobDepth,
				CharacterID: 0,
			})
			unassigned += pj.Entry.Runs
			continue
		}

		totalRuns := pj.Entry.Runs
		numChars := len(eligible)
		runsPerChar := int(math.Ceil(float64(totalRuns) / float64(numChars)))

		for j, ec := range eligible {
			runsForThis := runsPerChar
			// Last character gets the remainder
			if j == numChars-1 {
				runsForThis = totalRuns - runsPerChar*(numChars-1)
				if runsForThis <= 0 {
					break
				}
			}

			// Recalculate duration with this character's actual skills
			var durationSec int
			bpTime := pj.BaseBlueprintTime
			if bpTime <= 0 {
				// Fall back to stored estimated duration if base time is missing
				if pj.Entry.EstimatedDuration != nil {
					durationSec = *pj.Entry.EstimatedDuration * runsForThis
					if pj.Entry.Runs > 0 {
						durationSec = (*pj.Entry.EstimatedDuration * runsForThis) / pj.Entry.Runs
					}
				}
			} else if activity == "manufacturing" {
				teFactor := calculator.ComputeManufacturingTE(
					pj.BlueprintTE,
					ec.cap.IndustrySkill,
					ec.cap.AdvIndustrySkill,
					pj.Structure, pj.Rig, pj.Security,
				)
				secsPerRun := calculator.ComputeSecsPerRun(bpTime, teFactor)
				durationSec = secsPerRun * runsForThis
			} else { // reaction
				teFactor := calculator.ComputeTEFactor(
					ec.cap.ReactionsSkill,
					pj.Structure, pj.Rig, pj.Security,
				)
				secsPerRun := calculator.ComputeSecsPerRun(bpTime, teFactor)
				durationSec = secsPerRun * runsForThis
			}

			assigned = append(assigned, &AssignedJob{
				Original:    pj,
				Activity:    activity,
				Runs:        runsForThis,
				DurationSec: durationSec,
				Depth:       jobDepth,
				CharacterID: ec.cap.CharacterID,
			})

			// Consume one slot
			if activity == "manufacturing" {
				mfgAvail[ec.idx]--
			} else {
				reactAvail[ec.idx]--
			}
		}
	}

	return assigned, unassigned
}

// EstimateWallClock returns the total wall-clock seconds for all assigned jobs,
// using a depth-aware LPT (Longest Processing Time) scheduling model.
// Depths are processed sequentially (deepest first); within each depth, characters
// run in parallel across their slots using LPT lane assignment.
func EstimateWallClock(jobs []*AssignedJob, capacities []*calculator.CharacterCapacity) int {
	if len(jobs) == 0 {
		return 0
	}

	// Gather unique depth levels (descending order = deepest/leaves first)
	depthSet := make(map[int]bool)
	for _, aj := range jobs {
		depthSet[aj.Depth] = true
	}
	depths := make([]int, 0, len(depthSet))
	for d := range depthSet {
		depths = append(depths, d)
	}
	sort.Slice(depths, func(i, j int) bool { return depths[i] > depths[j] })

	// Build capacity lookup: CharacterID → CharacterCapacity
	capByChar := make(map[int64]*calculator.CharacterCapacity, len(capacities))
	for _, cap := range capacities {
		capByChar[cap.CharacterID] = cap
	}

	total := 0

	for _, depth := range depths {
		// Collect jobs at this depth, grouped by character
		charJobs := make(map[int64][]*AssignedJob)
		for _, aj := range jobs {
			if aj.Depth == depth {
				charJobs[aj.GetCharacterID()] = append(charJobs[aj.GetCharacterID()], aj)
			}
		}

		depthTime := 0

		for charID, cjobs := range charJobs {
			cap := capByChar[charID]

			// Determine slot count for this activity type
			// (all jobs for this character at this depth share the same activity)
			activity := ""
			if len(cjobs) > 0 {
				activity = cjobs[0].Activity
			}

			var numSlots int
			if cap == nil {
				numSlots = 1
			} else if activity == "manufacturing" {
				numSlots = cap.MfgSlotsMax
				if numSlots <= 0 {
					numSlots = 1
				}
			} else {
				numSlots = cap.ReactSlotsMax
				if numSlots <= 0 {
					numSlots = 1
				}
			}

			// LPT: sort jobs by duration descending
			sort.Slice(cjobs, func(i, j int) bool {
				return cjobs[i].DurationSec > cjobs[j].DurationSec
			})

			// Assign jobs to lanes (one lane = one slot)
			lanes := make([]int, numSlots)
			for _, aj := range cjobs {
				// Find lane with least total time
				minIdx := 0
				for k := 1; k < numSlots; k++ {
					if lanes[k] < lanes[minIdx] {
						minIdx = k
					}
				}
				lanes[minIdx] += aj.DurationSec
			}

			// Character completion time = max lane time
			charTime := 0
			for _, lt := range lanes {
				if lt > charTime {
					charTime = lt
				}
			}

			if charTime > depthTime {
				depthTime = charTime
			}
		}

		total += depthTime
	}

	return total
}
