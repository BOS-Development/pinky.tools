package updaters

import (
	"context"
	"fmt"

	"github.com/annymsMthd/industry-tool/internal/calculator"
	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/services"
	"github.com/pkg/errors"
)

// autoProductionGroup holds data for a single (userID, planID) production group.
type autoProductionGroup struct {
	userID      int64
	planID      int64
	typeIDs     map[int64]bool
	parallelism int
}

type AutoProductionStockpileMarkersRepository interface {
	GetAutoProductionMarkers(ctx context.Context) ([]*models.StockpileMarker, error)
}

type AutoProductionAssetsRepository interface {
	GetStockpileDeficits(ctx context.Context, user int64) (*repositories.StockpilesResponse, error)
}

type AutoProductionPlansRepository interface {
	GetByID(ctx context.Context, id, userID int64) (*models.ProductionPlan, error)
}

type AutoProductionPlanRunsRepository interface {
	Create(ctx context.Context, run *models.ProductionPlanRun) (*models.ProductionPlanRun, error)
	GetPendingOutputForPlan(ctx context.Context, planID, userID int64) (int64, error)
}

type AutoProductionMarketRepository interface {
	GetAllJitaPrices(ctx context.Context) (map[int64]*models.MarketPrice, error)
	GetAllAdjustedPrices(ctx context.Context) (map[int64]float64, error)
}

type AutoProductionJobQueueRepository interface {
	Create(ctx context.Context, entry *models.IndustryJobQueueEntry) (*models.IndustryJobQueueEntry, error)
	GetSlotUsage(ctx context.Context, userID int64) (map[int64]map[string]int, error)
}

type AutoProductionCharacterRepository interface {
	GetNames(ctx context.Context, userID int64) (map[int64]string, error)
}

type AutoProductionSkillsRepository interface {
	GetSkillsForUser(ctx context.Context, userID int64) ([]*models.CharacterSkill, error)
}

type AutoProductionSdeRepository interface {
	services.JobGenSdeRepository
}

type AutoProductionUpdater struct {
	markersRepo AutoProductionStockpileMarkersRepository
	assetsRepo  AutoProductionAssetsRepository
	plansRepo   AutoProductionPlansRepository
	runsRepo    AutoProductionPlanRunsRepository
	marketRepo  AutoProductionMarketRepository
	queueRepo   AutoProductionJobQueueRepository
	charRepo    AutoProductionCharacterRepository
	skillsRepo  AutoProductionSkillsRepository
	sdeRepo     AutoProductionSdeRepository
}

func NewAutoProductionUpdater(
	markersRepo AutoProductionStockpileMarkersRepository,
	assetsRepo AutoProductionAssetsRepository,
	plansRepo AutoProductionPlansRepository,
	runsRepo AutoProductionPlanRunsRepository,
	marketRepo AutoProductionMarketRepository,
	queueRepo AutoProductionJobQueueRepository,
	charRepo AutoProductionCharacterRepository,
	skillsRepo AutoProductionSkillsRepository,
	sdeRepo AutoProductionSdeRepository,
) *AutoProductionUpdater {
	return &AutoProductionUpdater{
		markersRepo: markersRepo,
		assetsRepo:  assetsRepo,
		plansRepo:   plansRepo,
		runsRepo:    runsRepo,
		marketRepo:  marketRepo,
		queueRepo:   queueRepo,
		charRepo:    charRepo,
		skillsRepo:  skillsRepo,
		sdeRepo:     sdeRepo,
	}
}

// RunAll fetches all auto-production markers, groups them by (userID, planID),
// calculates net deficits, and generates production plan runs for any outstanding demand.
func (u *AutoProductionUpdater) RunAll(ctx context.Context) error {
	// 1. Fetch all auto-production markers
	markers, err := u.markersRepo.GetAutoProductionMarkers(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get auto-production markers")
	}
	if len(markers) == 0 {
		return nil
	}

	// 2. Group markers by (userID, planID)
	groups := map[string]*autoProductionGroup{}

	for _, m := range markers {
		if m.PlanID == nil {
			continue
		}
		key := fmt.Sprintf("%d:%d", m.UserID, *m.PlanID)
		g, ok := groups[key]
		if !ok {
			g = &autoProductionGroup{
				userID:  m.UserID,
				planID:  *m.PlanID,
				typeIDs: map[int64]bool{},
			}
			groups[key] = g
		}
		g.typeIDs[m.TypeID] = true
		if m.AutoProductionParallelism > g.parallelism {
			g.parallelism = m.AutoProductionParallelism
		}
	}

	// 3. Process each group
	for _, group := range groups {
		if err := u.processGroup(ctx, group); err != nil {
			log.Error("auto-production: failed to process group",
				"userID", group.userID, "planID", group.planID, "error", err)
		}
	}

	return nil
}

func (u *AutoProductionUpdater) processGroup(ctx context.Context, group *autoProductionGroup) error {
	// 1. Get the plan with steps
	plan, err := u.plansRepo.GetByID(ctx, group.planID, group.userID)
	if err != nil {
		return errors.Wrap(err, "failed to get plan")
	}
	if plan == nil || len(plan.Steps) == 0 {
		return nil // plan missing or has no steps
	}

	// 2. Calculate gross deficit from stockpile data
	deficitsResp, err := u.assetsRepo.GetStockpileDeficits(ctx, group.userID)
	if err != nil {
		return errors.Wrap(err, "failed to get stockpile deficits")
	}

	// Sum deficits for type_ids in this group
	var grossDeficit int64
	if deficitsResp != nil {
		for _, item := range deficitsResp.Items {
			if group.typeIDs[item.TypeID] && item.StockpileDelta < 0 {
				grossDeficit += int64(-item.StockpileDelta) // delta is negative
			}
		}
	}

	if grossDeficit <= 0 {
		return nil // no deficit
	}

	// 3. Calculate pending output from existing runs
	pendingOutput, err := u.runsRepo.GetPendingOutputForPlan(ctx, group.planID, group.userID)
	if err != nil {
		return errors.Wrap(err, "failed to get pending output")
	}

	netDeficit := grossDeficit - pendingOutput
	if netDeficit <= 0 {
		log.Info("auto-production: deficit covered by pending output",
			"userID", group.userID, "planID", group.planID,
			"grossDeficit", grossDeficit, "pendingOutput", pendingOutput)
		return nil
	}

	// 4. Fetch market data
	jitaPrices, err := u.marketRepo.GetAllJitaPrices(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get Jita prices")
	}
	adjustedPrices, err := u.marketRepo.GetAllAdjustedPrices(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get adjusted prices")
	}

	// 5. Walk and merge steps
	wr, err := services.WalkAndMergeSteps(ctx, u.sdeRepo, plan, int(netDeficit), jitaPrices, adjustedPrices)
	if err != nil {
		return errors.Wrap(err, "failed to walk and merge steps")
	}

	// 6. Optional character assignment
	var assignedJobs []*services.AssignedJob
	if group.parallelism >= 1 {
		characterNames, err := u.charRepo.GetNames(ctx, group.userID)
		if err != nil {
			return errors.Wrap(err, "failed to get character names")
		}

		allSkills, err := u.skillsRepo.GetSkillsForUser(ctx, group.userID)
		if err != nil {
			return errors.Wrap(err, "failed to get character skills")
		}

		industrySkillSet := make(map[int64]bool, len(calculator.IndustrySkillIDs))
		for _, id := range calculator.IndustrySkillIDs {
			industrySkillSet[id] = true
		}
		skillsByCharacter := make(map[int64]map[int64]int)
		for _, sk := range allSkills {
			if !industrySkillSet[sk.SkillID] {
				continue
			}
			if skillsByCharacter[sk.CharacterID] == nil {
				skillsByCharacter[sk.CharacterID] = make(map[int64]int)
			}
			skillsByCharacter[sk.CharacterID][sk.SkillID] = sk.ActiveLevel
		}

		slotUsage, err := u.queueRepo.GetSlotUsage(ctx, group.userID)
		if err != nil {
			return errors.Wrap(err, "failed to get slot usage")
		}

		capacities := calculator.BuildCharacterCapacities(characterNames, skillsByCharacter, slotUsage)
		assignedJobs, _ = services.SimulateAssignment(wr.MergedJobs, capacities, group.parallelism)
	}

	// 7. Create plan run
	run, err := u.runsRepo.Create(ctx, &models.ProductionPlanRun{
		PlanID:   group.planID,
		UserID:   group.userID,
		Quantity: int(netDeficit),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create plan run")
	}

	// 8. Create job queue entries
	note := fmt.Sprintf("Auto-production from plan: %s", plan.Name)

	if group.parallelism >= 1 && len(assignedJobs) > 0 {
		for _, aj := range assignedJobs {
			orig := aj.Original
			origEntry := orig.Entry

			var charIDPtr *int64
			if aj.CharacterID != 0 {
				cid := aj.CharacterID
				charIDPtr = &cid
			}

			var estimatedCost *float64
			if origEntry.EstimatedCost != nil && origEntry.Runs > 0 {
				cost := *origEntry.EstimatedCost * float64(aj.Runs) / float64(origEntry.Runs)
				estimatedCost = &cost
			}

			entryNote := fmt.Sprintf("%s x%d", orig.ProductName, aj.Runs)
			newEntry := &models.IndustryJobQueueEntry{
				UserID:            group.userID,
				CharacterID:       charIDPtr,
				BlueprintTypeID:   origEntry.BlueprintTypeID,
				Activity:          aj.Activity,
				Runs:              aj.Runs,
				MELevel:           origEntry.MELevel,
				TELevel:           origEntry.TELevel,
				FacilityTax:       origEntry.FacilityTax,
				ProductTypeID:     origEntry.ProductTypeID,
				PlanRunID:         &run.ID,
				PlanStepID:        origEntry.PlanStepID,
				SortOrder:         origEntry.SortOrder,
				StationName:       origEntry.StationName,
				InputLocation:     origEntry.InputLocation,
				OutputLocation:    origEntry.OutputLocation,
				EstimatedCost:     estimatedCost,
				EstimatedDuration: &aj.DurationSec,
				Notes:             &entryNote,
			}

			_, err := u.queueRepo.Create(ctx, newEntry)
			if err != nil {
				log.Error("auto-production: failed to create queue entry",
					"userID", group.userID, "planID", group.planID, "error", err)
			}
		}
	} else {
		for _, pj := range wr.MergedJobs {
			pj.Entry.UserID = group.userID
			pj.Entry.Notes = &note
			pj.Entry.PlanRunID = &run.ID

			_, err := u.queueRepo.Create(ctx, pj.Entry)
			if err != nil {
				log.Error("auto-production: failed to create queue entry",
					"userID", group.userID, "planID", group.planID, "error", err)
			}
		}
	}

	log.Info("auto-production: generated plan run",
		"userID", group.userID, "planID", group.planID,
		"quantity", netDeficit, "jobs", len(wr.MergedJobs))

	return nil
}
