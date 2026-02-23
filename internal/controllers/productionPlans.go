package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/annymsMthd/industry-tool/internal/calculator"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type ProductionPlansRepository interface {
	Create(ctx context.Context, plan *models.ProductionPlan) (*models.ProductionPlan, error)
	GetByUser(ctx context.Context, userID int64) ([]*models.ProductionPlan, error)
	GetByID(ctx context.Context, id, userID int64) (*models.ProductionPlan, error)
	Update(ctx context.Context, id, userID int64, name string, notes *string, defaultManufacturingStationID *int64, defaultReactionStationID *int64) error
	Delete(ctx context.Context, id, userID int64) error
	CreateStep(ctx context.Context, step *models.ProductionPlanStep) (*models.ProductionPlanStep, error)
	UpdateStep(ctx context.Context, stepID, planID, userID int64, step *models.ProductionPlanStep) error
	BatchUpdateSteps(ctx context.Context, stepIDs []int64, planID, userID int64, step *models.ProductionPlanStep) (int64, error)
	DeleteStep(ctx context.Context, stepID, planID, userID int64) error
	GetStepMaterials(ctx context.Context, stepID, planID, userID int64) ([]*models.PlanMaterial, error)
	GetContainersAtStation(ctx context.Context, userID, stationID int64) ([]*models.StationContainer, error)
}

type ProductionPlansCharacterRepository interface {
	GetNames(ctx context.Context, userID int64) (map[int64]string, error)
}

type ProductionPlansCorporationRepository interface {
	Get(ctx context.Context, user int64) ([]repositories.PlayerCorporation, error)
	GetDivisions(ctx context.Context, corp, user int64) (*models.CorporationDivisions, error)
}

type ProductionPlansUserStationRepository interface {
	GetByID(ctx context.Context, id, userID int64) (*models.UserStation, error)
}

type ProductionPlansSdeRepository interface {
	GetBlueprintByProduct(ctx context.Context, productTypeID int64) (*repositories.BlueprintProductRow, error)
	GetManufacturingBlueprint(ctx context.Context, blueprintTypeID int64) (*repositories.ManufacturingBlueprintRow, error)
	GetManufacturingMaterials(ctx context.Context, blueprintTypeID int64) ([]*repositories.ManufacturingMaterialRow, error)
}

type ProductionPlansMarketRepository interface {
	GetAllJitaPrices(ctx context.Context) (map[int64]*models.MarketPrice, error)
	GetAllAdjustedPrices(ctx context.Context) (map[int64]float64, error)
}

type ProductionPlansCostIndicesRepository interface {
	GetCostIndex(ctx context.Context, systemID int64, activity string) (*models.IndustryCostIndex, error)
}

type ProductionPlansJobQueueRepository interface {
	Create(ctx context.Context, entry *models.IndustryJobQueueEntry) (*models.IndustryJobQueueEntry, error)
}

type ProductionPlanRunsRepository interface {
	Create(ctx context.Context, run *models.ProductionPlanRun) (*models.ProductionPlanRun, error)
	GetByPlan(ctx context.Context, planID, userID int64) ([]*models.ProductionPlanRun, error)
	GetByID(ctx context.Context, runID, userID int64) (*models.ProductionPlanRun, error)
	Delete(ctx context.Context, runID, userID int64) error
}

type ProductionPlans struct {
	plansRepo       ProductionPlansRepository
	sdeRepo         ProductionPlansSdeRepository
	queueRepo       ProductionPlansJobQueueRepository
	marketRepo      ProductionPlansMarketRepository
	costIndicesRepo ProductionPlansCostIndicesRepository
	characterRepo   ProductionPlansCharacterRepository
	corpRepo        ProductionPlansCorporationRepository
	stationRepo     ProductionPlansUserStationRepository
	runsRepo        ProductionPlanRunsRepository
}

func NewProductionPlans(
	router Routerer,
	plansRepo ProductionPlansRepository,
	sdeRepo ProductionPlansSdeRepository,
	queueRepo ProductionPlansJobQueueRepository,
	marketRepo ProductionPlansMarketRepository,
	costIndicesRepo ProductionPlansCostIndicesRepository,
	characterRepo ProductionPlansCharacterRepository,
	corpRepo ProductionPlansCorporationRepository,
	stationRepo ProductionPlansUserStationRepository,
	runsRepo ProductionPlanRunsRepository,
) *ProductionPlans {
	c := &ProductionPlans{
		plansRepo:       plansRepo,
		sdeRepo:         sdeRepo,
		queueRepo:       queueRepo,
		marketRepo:      marketRepo,
		costIndicesRepo: costIndicesRepo,
		characterRepo:   characterRepo,
		corpRepo:        corpRepo,
		stationRepo:     stationRepo,
		runsRepo:        runsRepo,
	}

	router.RegisterRestAPIRoute("/v1/industry/plans", web.AuthAccessUser, c.GetPlans, "GET")
	router.RegisterRestAPIRoute("/v1/industry/plans", web.AuthAccessUser, c.CreatePlan, "POST")
	router.RegisterRestAPIRoute("/v1/industry/plans/hangars", web.AuthAccessUser, c.GetHangars, "GET")
	router.RegisterRestAPIRoute("/v1/industry/plans/{id}", web.AuthAccessUser, c.GetPlan, "GET")
	router.RegisterRestAPIRoute("/v1/industry/plans/{id}", web.AuthAccessUser, c.UpdatePlan, "PUT")
	router.RegisterRestAPIRoute("/v1/industry/plans/{id}", web.AuthAccessUser, c.DeletePlan, "DELETE")
	router.RegisterRestAPIRoute("/v1/industry/plans/{id}/steps", web.AuthAccessUser, c.CreateStep, "POST")
	router.RegisterRestAPIRoute("/v1/industry/plans/{id}/steps/batch", web.AuthAccessUser, c.BatchUpdateSteps, "PUT")
	router.RegisterRestAPIRoute("/v1/industry/plans/{id}/steps/{stepId}", web.AuthAccessUser, c.UpdateStep, "PUT")
	router.RegisterRestAPIRoute("/v1/industry/plans/{id}/steps/{stepId}", web.AuthAccessUser, c.DeleteStep, "DELETE")
	router.RegisterRestAPIRoute("/v1/industry/plans/{id}/steps/{stepId}/materials", web.AuthAccessUser, c.GetStepMaterials, "GET")
	router.RegisterRestAPIRoute("/v1/industry/plans/{id}/runs", web.AuthAccessUser, c.GetPlanRuns, "GET")
	router.RegisterRestAPIRoute("/v1/industry/plans/{id}/runs/{runId}", web.AuthAccessUser, c.GetPlanRun, "GET")
	router.RegisterRestAPIRoute("/v1/industry/plans/{id}/runs/{runId}", web.AuthAccessUser, c.DeletePlanRun, "DELETE")
	router.RegisterRestAPIRoute("/v1/industry/plans/{id}/generate", web.AuthAccessUser, c.GenerateJobs, "POST")

	return c
}

// GetPlans lists all production plans for the user.
func (c *ProductionPlans) GetPlans(args *web.HandlerArgs) (any, *web.HttpError) {
	plans, err := c.plansRepo.GetByUser(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get production plans")}
	}
	return plans, nil
}

type createPlanRequest struct {
	ProductTypeID                 int64   `json:"product_type_id"`
	Name                          string  `json:"name"`
	Notes                         *string `json:"notes"`
	DefaultManufacturingStationID *int64  `json:"default_manufacturing_station_id"`
	DefaultReactionStationID      *int64  `json:"default_reaction_station_id"`
}

// CreatePlan creates a new production plan and auto-creates the root step.
func (c *ProductionPlans) CreatePlan(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()

	var req createPlanRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	if req.ProductTypeID <= 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("product_type_id is required")}
	}

	// Look up the blueprint for this product
	bp, err := c.sdeRepo.GetBlueprintByProduct(ctx, req.ProductTypeID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to look up blueprint")}
	}
	if bp == nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("no manufacturing or reaction blueprint found for this product")}
	}

	// Get product name for default plan name
	planName := req.Name
	if planName == "" {
		bpInfo, err := c.sdeRepo.GetManufacturingBlueprint(ctx, bp.BlueprintTypeID)
		if err == nil && bpInfo != nil {
			planName = bpInfo.ProductName
		}
	}

	// Create the plan
	plan, err := c.plansRepo.Create(ctx, &models.ProductionPlan{
		UserID:                        *args.User,
		ProductTypeID:                 req.ProductTypeID,
		Name:                          planName,
		Notes:                         req.Notes,
		DefaultManufacturingStationID: req.DefaultManufacturingStationID,
		DefaultReactionStationID:      req.DefaultReactionStationID,
	})
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to create production plan")}
	}

	// Determine default station for root step based on activity
	var defaultStationID *int64
	if bp.Activity == "manufacturing" {
		defaultStationID = req.DefaultManufacturingStationID
	} else if bp.Activity == "reaction" {
		defaultStationID = req.DefaultReactionStationID
	}

	// Create the root step
	rootStep := &models.ProductionPlanStep{
		PlanID:           plan.ID,
		ProductTypeID:    req.ProductTypeID,
		BlueprintTypeID:  bp.BlueprintTypeID,
		Activity:         bp.Activity,
		MELevel:          10,
		TELevel:          20,
		IndustrySkill:    5,
		AdvIndustrySkill: 5,
		Structure:        "raitaru",
		Rig:              "t2",
		Security:         "high",
		FacilityTax:      1.0,
		UserStationID:    defaultStationID,
	}

	_, err = c.plansRepo.CreateStep(ctx, rootStep)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to create root step")}
	}

	// Return the full plan with steps
	fullPlan, err := c.plansRepo.GetByID(ctx, plan.ID, *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to fetch created plan")}
	}

	return fullPlan, nil
}

// GetPlan returns a single plan with its full step tree.
func (c *ProductionPlans) GetPlan(args *web.HandlerArgs) (any, *web.HttpError) {
	id, err := parseID(args.Params["id"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid plan ID")}
	}

	plan, err := c.plansRepo.GetByID(args.Request.Context(), id, *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get production plan")}
	}
	if plan == nil {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.New("production plan not found")}
	}

	return plan, nil
}

type updatePlanRequest struct {
	Name                          string  `json:"name"`
	Notes                         *string `json:"notes"`
	DefaultManufacturingStationID *int64  `json:"default_manufacturing_station_id"`
	DefaultReactionStationID      *int64  `json:"default_reaction_station_id"`
}

// UpdatePlan updates plan metadata (name, notes).
func (c *ProductionPlans) UpdatePlan(args *web.HandlerArgs) (any, *web.HttpError) {
	id, err := parseID(args.Params["id"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid plan ID")}
	}

	var req updatePlanRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	if err := c.plansRepo.Update(args.Request.Context(), id, *args.User, req.Name, req.Notes, req.DefaultManufacturingStationID, req.DefaultReactionStationID); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to update production plan")}
	}

	return map[string]string{"status": "updated"}, nil
}

// DeletePlan deletes a plan and all its steps.
func (c *ProductionPlans) DeletePlan(args *web.HandlerArgs) (any, *web.HttpError) {
	id, err := parseID(args.Params["id"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid plan ID")}
	}

	if err := c.plansRepo.Delete(args.Request.Context(), id, *args.User); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to delete production plan")}
	}

	return map[string]string{"status": "deleted"}, nil
}

type createStepRequest struct {
	ParentStepID  int64   `json:"parent_step_id"`
	ProductTypeID int64   `json:"product_type_id"`
	MELevel       *int    `json:"me_level"`
	TELevel       *int    `json:"te_level"`
	Structure     string  `json:"structure"`
	Rig           string  `json:"rig"`
	Security      string  `json:"security"`
	FacilityTax   *float64 `json:"facility_tax"`
}

// CreateStep adds a production step (toggles a material to "produce").
func (c *ProductionPlans) CreateStep(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()

	planID, err := parseID(args.Params["id"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid plan ID")}
	}

	var req createStepRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	if req.ParentStepID <= 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("parent_step_id is required")}
	}
	if req.ProductTypeID <= 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("product_type_id is required")}
	}

	// Verify the plan exists and belongs to user
	plan, err := c.plansRepo.GetByID(ctx, planID, *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to verify plan")}
	}
	if plan == nil {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.New("production plan not found")}
	}

	// Look up the blueprint for the material
	bp, err := c.sdeRepo.GetBlueprintByProduct(ctx, req.ProductTypeID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to look up blueprint")}
	}
	if bp == nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("no manufacturing or reaction blueprint found for this product")}
	}

	meLevel := 10
	if req.MELevel != nil {
		meLevel = *req.MELevel
	}
	teLevel := 20
	if req.TELevel != nil {
		teLevel = *req.TELevel
	}
	facilityTax := 1.0
	if req.FacilityTax != nil {
		facilityTax = *req.FacilityTax
	}

	// Auto-assign user station from plan defaults based on activity
	var stepStationID *int64
	if bp.Activity == "manufacturing" && plan.DefaultManufacturingStationID != nil {
		stepStationID = plan.DefaultManufacturingStationID
	} else if bp.Activity == "reaction" && plan.DefaultReactionStationID != nil {
		stepStationID = plan.DefaultReactionStationID
	}

	step := &models.ProductionPlanStep{
		PlanID:           planID,
		ParentStepID:     &req.ParentStepID,
		ProductTypeID:    req.ProductTypeID,
		BlueprintTypeID:  bp.BlueprintTypeID,
		Activity:         bp.Activity,
		MELevel:          meLevel,
		TELevel:          teLevel,
		IndustrySkill:    5,
		AdvIndustrySkill: 5,
		Structure:        withDefault(req.Structure, "raitaru"),
		Rig:              withDefault(req.Rig, "t2"),
		Security:         withDefault(req.Security, "high"),
		FacilityTax:      facilityTax,
		UserStationID:    stepStationID,
	}

	created, err := c.plansRepo.CreateStep(ctx, step)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to create step")}
	}

	return created, nil
}

type updateStepRequest struct {
	MELevel              int      `json:"me_level"`
	TELevel              int      `json:"te_level"`
	IndustrySkill        int      `json:"industry_skill"`
	AdvIndustrySkill     int      `json:"adv_industry_skill"`
	Structure            string   `json:"structure"`
	Rig                  string   `json:"rig"`
	Security             string   `json:"security"`
	FacilityTax          float64  `json:"facility_tax"`
	StationName          *string  `json:"station_name"`
	SourceLocationID     *int64   `json:"source_location_id"`
	SourceContainerID    *int64   `json:"source_container_id"`
	SourceDivisionNumber *int     `json:"source_division_number"`
	SourceOwnerType      *string  `json:"source_owner_type"`
	SourceOwnerID        *int64   `json:"source_owner_id"`
	OutputOwnerType      *string  `json:"output_owner_type"`
	OutputOwnerID        *int64   `json:"output_owner_id"`
	OutputDivisionNumber *int     `json:"output_division_number"`
	OutputContainerID    *int64   `json:"output_container_id"`
	UserStationID        *int64   `json:"user_station_id"`
}

// UpdateStep updates parameters of a production step.
func (c *ProductionPlans) UpdateStep(args *web.HandlerArgs) (any, *web.HttpError) {
	planID, err := parseID(args.Params["id"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid plan ID")}
	}
	stepID, err := parseID(args.Params["stepId"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid step ID")}
	}

	var req updateStepRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	step := &models.ProductionPlanStep{
		MELevel:              req.MELevel,
		TELevel:              req.TELevel,
		IndustrySkill:        req.IndustrySkill,
		AdvIndustrySkill:     req.AdvIndustrySkill,
		Structure:            req.Structure,
		Rig:                  req.Rig,
		Security:             req.Security,
		FacilityTax:          req.FacilityTax,
		StationName:          req.StationName,
		SourceLocationID:     req.SourceLocationID,
		SourceContainerID:    req.SourceContainerID,
		SourceDivisionNumber: req.SourceDivisionNumber,
		SourceOwnerType:      req.SourceOwnerType,
		SourceOwnerID:        req.SourceOwnerID,
		OutputOwnerType:      req.OutputOwnerType,
		OutputOwnerID:        req.OutputOwnerID,
		OutputDivisionNumber: req.OutputDivisionNumber,
		OutputContainerID:    req.OutputContainerID,
		UserStationID:        req.UserStationID,
	}

	if err := c.plansRepo.UpdateStep(args.Request.Context(), stepID, planID, *args.User, step); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to update step")}
	}

	return map[string]string{"status": "updated"}, nil
}

type batchUpdateStepsRequest struct {
	StepIDs              []int64  `json:"step_ids"`
	MELevel              int      `json:"me_level"`
	TELevel              int      `json:"te_level"`
	IndustrySkill        int      `json:"industry_skill"`
	AdvIndustrySkill     int      `json:"adv_industry_skill"`
	Structure            string   `json:"structure"`
	Rig                  string   `json:"rig"`
	Security             string   `json:"security"`
	FacilityTax          float64  `json:"facility_tax"`
	StationName          *string  `json:"station_name"`
	SourceLocationID     *int64   `json:"source_location_id"`
	SourceContainerID    *int64   `json:"source_container_id"`
	SourceDivisionNumber *int     `json:"source_division_number"`
	SourceOwnerType      *string  `json:"source_owner_type"`
	SourceOwnerID        *int64   `json:"source_owner_id"`
	OutputOwnerType      *string  `json:"output_owner_type"`
	OutputOwnerID        *int64   `json:"output_owner_id"`
	OutputDivisionNumber *int     `json:"output_division_number"`
	OutputContainerID    *int64   `json:"output_container_id"`
	UserStationID        *int64   `json:"user_station_id"`
}

// BatchUpdateSteps updates multiple steps at once with the same parameters.
func (c *ProductionPlans) BatchUpdateSteps(args *web.HandlerArgs) (any, *web.HttpError) {
	planID, err := parseID(args.Params["id"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid plan ID")}
	}

	var req batchUpdateStepsRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	if len(req.StepIDs) == 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("step_ids is required and must not be empty")}
	}

	step := &models.ProductionPlanStep{
		MELevel:              req.MELevel,
		TELevel:              req.TELevel,
		IndustrySkill:        req.IndustrySkill,
		AdvIndustrySkill:     req.AdvIndustrySkill,
		Structure:            req.Structure,
		Rig:                  req.Rig,
		Security:             req.Security,
		FacilityTax:          req.FacilityTax,
		StationName:          req.StationName,
		SourceLocationID:     req.SourceLocationID,
		SourceContainerID:    req.SourceContainerID,
		SourceDivisionNumber: req.SourceDivisionNumber,
		SourceOwnerType:      req.SourceOwnerType,
		SourceOwnerID:        req.SourceOwnerID,
		OutputOwnerType:      req.OutputOwnerType,
		OutputOwnerID:        req.OutputOwnerID,
		OutputDivisionNumber: req.OutputDivisionNumber,
		OutputContainerID:    req.OutputContainerID,
		UserStationID:        req.UserStationID,
	}

	rowsAffected, err := c.plansRepo.BatchUpdateSteps(args.Request.Context(), req.StepIDs, planID, *args.User, step)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to batch update steps")}
	}

	return map[string]any{"status": "updated", "rows_affected": rowsAffected}, nil
}

// DeleteStep removes a production step (toggles material back to "buy").
func (c *ProductionPlans) DeleteStep(args *web.HandlerArgs) (any, *web.HttpError) {
	planID, err := parseID(args.Params["id"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid plan ID")}
	}
	stepID, err := parseID(args.Params["stepId"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid step ID")}
	}

	if err := c.plansRepo.DeleteStep(args.Request.Context(), stepID, planID, *args.User); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to delete step")}
	}

	return map[string]string{"status": "deleted"}, nil
}

// GetStepMaterials returns the materials for a step with producibility info.
func (c *ProductionPlans) GetStepMaterials(args *web.HandlerArgs) (any, *web.HttpError) {
	planID, err := parseID(args.Params["id"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid plan ID")}
	}
	stepID, err := parseID(args.Params["stepId"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid step ID")}
	}

	materials, err := c.plansRepo.GetStepMaterials(args.Request.Context(), stepID, planID, *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get step materials")}
	}

	return materials, nil
}

type hangarsResponse struct {
	Characters   []hangarsCharacter   `json:"characters"`
	Corporations []hangarsCorporation `json:"corporations"`
	Containers   []*models.StationContainer `json:"containers"`
}

type hangarsCharacter struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type hangarsCorporation struct {
	ID        int64             `json:"id"`
	Name      string            `json:"name"`
	Divisions map[string]string `json:"divisions"`
}

// GetHangars returns characters, corporations (with divisions), and containers at a station.
func (c *ProductionPlans) GetHangars(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User

	userStationIDStr := args.Request.URL.Query().Get("user_station_id")
	if userStationIDStr == "" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("user_station_id is required")}
	}

	userStationID, err := parseID(userStationIDStr)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid user_station_id")}
	}

	// Resolve user_station_id to real station_id
	station, err := c.stationRepo.GetByID(ctx, userStationID, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get station")}
	}
	if station == nil {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.New("station not found")}
	}

	// Get character names
	charNames, err := c.characterRepo.GetNames(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get character names")}
	}

	characters := []hangarsCharacter{}
	for id, name := range charNames {
		characters = append(characters, hangarsCharacter{ID: id, Name: name})
	}

	// Get corporations with divisions
	corps, err := c.corpRepo.Get(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get corporations")}
	}

	corporations := []hangarsCorporation{}
	for _, corp := range corps {
		divisions, err := c.corpRepo.GetDivisions(ctx, corp.ID, userID)
		if err != nil {
			return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get corporation divisions")}
		}

		divMap := map[string]string{}
		if divisions != nil {
			for num, name := range divisions.Hanger {
				divMap[fmt.Sprintf("%d", num)] = name
			}
		}

		corporations = append(corporations, hangarsCorporation{
			ID:        corp.ID,
			Name:      corp.Name,
			Divisions: divMap,
		})
	}

	// Get containers at station
	containers, err := c.plansRepo.GetContainersAtStation(ctx, userID, station.StationID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get containers")}
	}

	return &hangarsResponse{
		Characters:   characters,
		Corporations: corporations,
		Containers:   containers,
	}, nil
}

type generateJobsRequest struct {
	Quantity int `json:"quantity"`
}

// GenerateJobs creates job queue entries from a production plan for a given quantity.
func (c *ProductionPlans) GenerateJobs(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()

	planID, err := parseID(args.Params["id"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid plan ID")}
	}

	var req generateJobsRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}
	if req.Quantity <= 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("quantity must be positive")}
	}

	// Get the full plan
	plan, err := c.plansRepo.GetByID(ctx, planID, *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get plan")}
	}
	if plan == nil {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.New("production plan not found")}
	}
	if len(plan.Steps) == 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("plan has no steps")}
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
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("plan has no root step")}
	}

	// Fetch market data for cost calculations
	jitaPrices, err := c.marketRepo.GetAllJitaPrices(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get Jita prices")}
	}
	adjustedPrices, err := c.marketRepo.GetAllAdjustedPrices(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get adjusted prices")}
	}

	// Create the plan run
	run, err := c.runsRepo.Create(ctx, &models.ProductionPlanRun{
		PlanID:   planID,
		UserID:   *args.User,
		Quantity: req.Quantity,
	})
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to create plan run")}
	}

	result := &models.GenerateJobsResult{
		Run:     run,
		Created: []*models.IndustryJobQueueEntry{},
		Skipped: []*models.GenerateJobSkipped{},
	}

	// Walk the tree depth-first, collect jobs bottom-up
	var walkStep func(step *models.ProductionPlanStep, quantity int)
	walkStep = func(step *models.ProductionPlanStep, quantity int) {
		// Look up blueprint to get product quantity per run
		bp, err := c.sdeRepo.GetManufacturingBlueprint(ctx, step.BlueprintTypeID)
		if err != nil || bp == nil {
			result.Skipped = append(result.Skipped, &models.GenerateJobSkipped{
				TypeID:   step.ProductTypeID,
				TypeName: step.ProductName,
				Reason:   "blueprint data not found",
			})
			return
		}

		// Calculate runs needed
		runs := int(math.Ceil(float64(quantity) / float64(bp.ProductQuantity)))
		if runs <= 0 {
			runs = 1
		}

		// Get materials for this step to calculate child needs
		materials, err := c.sdeRepo.GetManufacturingMaterials(ctx, step.BlueprintTypeID)
		if err != nil {
			result.Skipped = append(result.Skipped, &models.GenerateJobSkipped{
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
				walkStep(childStep, int(batchQty))
			}
		}

		// Calculate cost for this step (manufacturing only)
		var estimatedCost *float64
		var estimatedDuration *int

		if step.Activity == "manufacturing" {
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
		}

		// Create queue entry
		productTypeID := step.ProductTypeID
		stepID := step.ID
		note := "Generated from plan: " + plan.Name
		entry := &models.IndustryJobQueueEntry{
			UserID:            *args.User,
			BlueprintTypeID:   step.BlueprintTypeID,
			Activity:          step.Activity,
			Runs:              runs,
			MELevel:           step.MELevel,
			TELevel:           step.TELevel,
			FacilityTax:       step.FacilityTax,
			ProductTypeID:     &productTypeID,
			EstimatedCost:     estimatedCost,
			EstimatedDuration: estimatedDuration,
			Notes:             &note,
			PlanRunID:         &run.ID,
			PlanStepID:        &stepID,
		}

		created, err := c.queueRepo.Create(ctx, entry)
		if err != nil {
			result.Skipped = append(result.Skipped, &models.GenerateJobSkipped{
				TypeID:   step.ProductTypeID,
				TypeName: step.ProductName,
				Reason:   "failed to create queue entry: " + err.Error(),
			})
			return
		}

		result.Created = append(result.Created, created)
	}

	walkStep(rootStep, req.Quantity)

	return result, nil
}

// GetPlanRuns lists all runs for a production plan.
func (c *ProductionPlans) GetPlanRuns(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()

	planID, err := parseID(args.Params["id"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid plan ID")}
	}

	// Verify plan exists and belongs to user
	plan, err := c.plansRepo.GetByID(ctx, planID, *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get plan")}
	}
	if plan == nil {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.New("production plan not found")}
	}

	runs, err := c.runsRepo.GetByPlan(ctx, planID, *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get plan runs")}
	}

	return runs, nil
}

// GetPlanRun returns a single plan run with its jobs.
func (c *ProductionPlans) GetPlanRun(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()

	planID, err := parseID(args.Params["id"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid plan ID")}
	}

	runID, err := parseID(args.Params["runId"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid run ID")}
	}

	run, err := c.runsRepo.GetByID(ctx, runID, *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get plan run")}
	}
	if run == nil || run.PlanID != planID {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.New("plan run not found")}
	}

	return run, nil
}

// DeletePlanRun deletes a plan run. Jobs survive but lose their plan_run_id link.
func (c *ProductionPlans) DeletePlanRun(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()

	_, err := parseID(args.Params["id"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid plan ID")}
	}

	runID, err := parseID(args.Params["runId"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid run ID")}
	}

	if err := c.runsRepo.Delete(ctx, runID, *args.User); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to delete plan run")}
	}

	return map[string]string{"status": "deleted"}, nil
}
