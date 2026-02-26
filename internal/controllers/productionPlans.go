package controllers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/annymsMthd/industry-tool/internal/calculator"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/services"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type ProductionPlansRepository interface {
	Create(ctx context.Context, plan *models.ProductionPlan) (*models.ProductionPlan, error)
	GetByUser(ctx context.Context, userID int64) ([]*models.ProductionPlan, error)
	GetByID(ctx context.Context, id, userID int64) (*models.ProductionPlan, error)
	Update(ctx context.Context, id, userID int64, name string, notes *string, defaultManufacturingStationID *int64, defaultReactionStationID *int64, transportFulfillment *string, transportMethod *string, transportProfileID *int64, courierRatePerM3 float64, courierCollateralRate float64) error
	Delete(ctx context.Context, id, userID int64) error
	CreateStep(ctx context.Context, step *models.ProductionPlanStep) (*models.ProductionPlanStep, error)
	UpdateStep(ctx context.Context, stepID, planID, userID int64, step *models.ProductionPlanStep) error
	BatchUpdateSteps(ctx context.Context, stepIDs []int64, planID, userID int64, step *models.ProductionPlanStep) (int64, error)
	DeleteStep(ctx context.Context, stepID, planID, userID int64) error
	GetStepMaterials(ctx context.Context, stepID, planID, userID int64) ([]*models.PlanMaterial, error)
	GetContainersAtStation(ctx context.Context, userID, stationID int64) ([]*models.StationContainer, error)
	GetByProductTypeAndUser(ctx context.Context, productTypeID, userID int64) ([]*models.ProductionPlan, error)
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
	GetBlueprintForActivity(ctx context.Context, blueprintTypeID int64, activity string) (*repositories.ManufacturingBlueprintRow, error)
	GetBlueprintMaterialsForActivity(ctx context.Context, blueprintTypeID int64, activity string) ([]*repositories.ManufacturingMaterialRow, error)
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
	GetSlotUsage(ctx context.Context, userID int64) (map[int64]map[string]int, error)
}

type ProductionPlanRunsRepository interface {
	Create(ctx context.Context, run *models.ProductionPlanRun) (*models.ProductionPlanRun, error)
	GetByPlan(ctx context.Context, planID, userID int64) ([]*models.ProductionPlanRun, error)
	GetByUser(ctx context.Context, userID int64) ([]*models.ProductionPlanRun, error)
	GetByID(ctx context.Context, runID, userID int64) (*models.ProductionPlanRun, error)
	Delete(ctx context.Context, runID, userID int64) error
	CancelPlannedJobs(ctx context.Context, runID, userID int64) (int64, error)
}

type ProductionPlansTransportJobsRepository interface {
	Create(ctx context.Context, job *models.TransportJob) (*models.TransportJob, error)
	SetQueueEntryID(ctx context.Context, id int64, queueEntryID int64) error
}

type ProductionPlansTransportProfilesRepository interface {
	GetByID(ctx context.Context, id, userID int64) (*models.TransportProfile, error)
	GetDefaultByMethod(ctx context.Context, userID int64, method string) (*models.TransportProfile, error)
}

type ProductionPlansJFRoutesRepository interface {
	FindBySystemPair(ctx context.Context, userID, originSystemID, destSystemID int64) (*models.JFRoute, error)
}

type ProductionPlansEsiClient interface {
	GetRoute(ctx context.Context, origin, destination int64, flag string) ([]int32, error)
}

type ProductionPlansCharacterSkillsRepository interface {
	GetSkillsForUser(ctx context.Context, userID int64) ([]*models.CharacterSkill, error)
}

type ProductionPlans struct {
	plansRepo        ProductionPlansRepository
	sdeRepo          ProductionPlansSdeRepository
	queueRepo        ProductionPlansJobQueueRepository
	marketRepo       ProductionPlansMarketRepository
	costIndicesRepo  ProductionPlansCostIndicesRepository
	characterRepo    ProductionPlansCharacterRepository
	corpRepo         ProductionPlansCorporationRepository
	stationRepo      ProductionPlansUserStationRepository
	runsRepo         ProductionPlanRunsRepository
	transportJobRepo ProductionPlansTransportJobsRepository
	profilesRepo     ProductionPlansTransportProfilesRepository
	jfRoutesRepo     ProductionPlansJFRoutesRepository
	esiClient        ProductionPlansEsiClient
	skillsRepo       ProductionPlansCharacterSkillsRepository
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
	transportJobRepo ProductionPlansTransportJobsRepository,
	profilesRepo ProductionPlansTransportProfilesRepository,
	jfRoutesRepo ProductionPlansJFRoutesRepository,
	esiClient ProductionPlansEsiClient,
	skillsRepo ProductionPlansCharacterSkillsRepository,
) *ProductionPlans {
	c := &ProductionPlans{
		plansRepo:        plansRepo,
		sdeRepo:          sdeRepo,
		queueRepo:        queueRepo,
		marketRepo:       marketRepo,
		costIndicesRepo:  costIndicesRepo,
		characterRepo:    characterRepo,
		corpRepo:         corpRepo,
		stationRepo:      stationRepo,
		runsRepo:         runsRepo,
		transportJobRepo: transportJobRepo,
		profilesRepo:     profilesRepo,
		jfRoutesRepo:     jfRoutesRepo,
		esiClient:        esiClient,
		skillsRepo:       skillsRepo,
	}

	router.RegisterRestAPIRoute("/v1/industry/plans", web.AuthAccessUser, c.GetPlans, "GET")
	router.RegisterRestAPIRoute("/v1/industry/plans", web.AuthAccessUser, c.CreatePlan, "POST")
	router.RegisterRestAPIRoute("/v1/industry/plans/hangars", web.AuthAccessUser, c.GetHangars, "GET")
	router.RegisterRestAPIRoute("/v1/industry/plans/runs", web.AuthAccessUser, c.GetAllRuns, "GET")
	router.RegisterRestAPIRoute("/v1/industry/plans/runs/{runId}/cancel", web.AuthAccessUser, c.CancelPlanRun, "POST")
	router.RegisterRestAPIRoute("/v1/industry/plans/by-product/{typeId}", web.AuthAccessUser, c.GetPlansByProduct, "GET")
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
	router.RegisterRestAPIRoute("/v1/industry/plans/{id}/preview", web.AuthAccessUser, c.PreviewPlan, "POST")
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

// GetPlansByProduct returns production plans that produce a specific item type.
func (c *ProductionPlans) GetPlansByProduct(args *web.HandlerArgs) (any, *web.HttpError) {
	typeID, err := parseID(args.Params["typeId"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid type ID")}
	}

	plans, err := c.plansRepo.GetByProductTypeAndUser(args.Request.Context(), typeID, *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get plans by product type")}
	}

	return plans, nil
}

type createPlanRequest struct {
	ProductTypeID                 int64    `json:"product_type_id"`
	Name                          string   `json:"name"`
	Notes                         *string  `json:"notes"`
	DefaultManufacturingStationID *int64   `json:"default_manufacturing_station_id"`
	DefaultReactionStationID      *int64   `json:"default_reaction_station_id"`
	TransportFulfillment          *string  `json:"transport_fulfillment"`
	TransportMethod               *string  `json:"transport_method"`
	TransportProfileID            *int64   `json:"transport_profile_id"`
	CourierRatePerM3              float64  `json:"courier_rate_per_m3"`
	CourierCollateralRate         float64  `json:"courier_collateral_rate"`
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
		bpInfo, err := c.sdeRepo.GetBlueprintForActivity(ctx, bp.BlueprintTypeID, bp.Activity)
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
		TransportFulfillment:          req.TransportFulfillment,
		TransportMethod:               req.TransportMethod,
		TransportProfileID:            req.TransportProfileID,
		CourierRatePerM3:              req.CourierRatePerM3,
		CourierCollateralRate:         req.CourierCollateralRate,
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
	Name                          string   `json:"name"`
	Notes                         *string  `json:"notes"`
	DefaultManufacturingStationID *int64   `json:"default_manufacturing_station_id"`
	DefaultReactionStationID      *int64   `json:"default_reaction_station_id"`
	TransportFulfillment          *string  `json:"transport_fulfillment"`
	TransportMethod               *string  `json:"transport_method"`
	TransportProfileID            *int64   `json:"transport_profile_id"`
	CourierRatePerM3              float64  `json:"courier_rate_per_m3"`
	CourierCollateralRate         float64  `json:"courier_collateral_rate"`
}

// UpdatePlan updates plan metadata (name, notes, transport settings).
func (c *ProductionPlans) UpdatePlan(args *web.HandlerArgs) (any, *web.HttpError) {
	id, err := parseID(args.Params["id"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid plan ID")}
	}

	var req updatePlanRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	if err := c.plansRepo.Update(args.Request.Context(), id, *args.User, req.Name, req.Notes, req.DefaultManufacturingStationID, req.DefaultReactionStationID, req.TransportFulfillment, req.TransportMethod, req.TransportProfileID, req.CourierRatePerM3, req.CourierCollateralRate); err != nil {
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
	Quantity    int `json:"quantity"`
	Parallelism int `json:"parallelism"`
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

	// Fetch market data for cost calculations
	jitaPrices, err := c.marketRepo.GetAllJitaPrices(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get Jita prices")}
	}
	adjustedPrices, err := c.marketRepo.GetAllAdjustedPrices(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get adjusted prices")}
	}

	// Walk the tree and merge jobs
	wr, err := services.WalkAndMergeSteps(ctx, c.sdeRepo, plan, req.Quantity, jitaPrices, adjustedPrices)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: err}
	}

	// Character assignment (when parallelism >= 1)
	var characterAssignments map[int64]string
	var unassignedCount int
	var assignedJobs []*services.AssignedJob

	if req.Parallelism >= 1 {
		characterNames, err := c.characterRepo.GetNames(ctx, *args.User)
		if err != nil {
			return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get character names")}
		}

		allSkills, err := c.skillsRepo.GetSkillsForUser(ctx, *args.User)
		if err != nil {
			return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get character skills")}
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

		slotUsage, err := c.queueRepo.GetSlotUsage(ctx, *args.User)
		if err != nil {
			return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get slot usage")}
		}

		capacities := calculator.BuildCharacterCapacities(characterNames, skillsByCharacter, slotUsage)

		assignedJobs, unassignedCount = services.SimulateAssignment(wr.MergedJobs, capacities, req.Parallelism)

		characterAssignments = make(map[int64]string)
		for _, aj := range assignedJobs {
			if aj.CharacterID != 0 {
				characterAssignments[aj.CharacterID] = characterNames[aj.CharacterID]
			}
		}
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
		Run:                  run,
		Created:              []*models.IndustryJobQueueEntry{},
		Skipped:              wr.Skipped,
		TransportJobs:        []*models.TransportJob{},
		CharacterAssignments: characterAssignments,
		UnassignedCount:      unassignedCount,
	}

	// Build the step index needed for transport generation
	stepsByID := make(map[int64]*models.ProductionPlanStep)
	childStepsByParent := make(map[int64][]*models.ProductionPlanStep)
	for _, step := range plan.Steps {
		stepsByID[step.ID] = step
		if step.ParentStepID != nil {
			childStepsByParent[*step.ParentStepID] = append(childStepsByParent[*step.ParentStepID], step)
		}
	}

	// Determine the list of jobs to persist.
	// When parallelism >= 1, we use the assigned job fragments (each carrying a
	// character assignment and a recalculated duration).  Otherwise we fall back
	// to the merged jobs produced by WalkAndMergeSteps.
	note := "Generated from plan: " + plan.Name

	if req.Parallelism >= 1 && len(assignedJobs) > 0 {
		for _, aj := range assignedJobs {
			orig := aj.Original
			origEntry := orig.Entry

			// Build the queue entry from the assigned fragment
			productTypeID := origEntry.ProductTypeID
			dur := aj.DurationSec

			// CharacterID=0 means no eligible character was found; persist the
			// entry without a character assignment so the job is not silently dropped.
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
				UserID:            *args.User,
				CharacterID:       charIDPtr,
				BlueprintTypeID:   origEntry.BlueprintTypeID,
				Activity:          aj.Activity,
				Runs:              aj.Runs,
				MELevel:           origEntry.MELevel,
				TELevel:           origEntry.TELevel,
				FacilityTax:       origEntry.FacilityTax,
				ProductTypeID:     productTypeID,
				PlanRunID:         &run.ID,
				PlanStepID:        origEntry.PlanStepID,
				SortOrder:         origEntry.SortOrder,
				StationName:       origEntry.StationName,
				InputLocation:     origEntry.InputLocation,
				OutputLocation:    origEntry.OutputLocation,
				EstimatedCost:     estimatedCost,
				EstimatedDuration: &dur,
				Notes:             &entryNote,
			}

			created, err := c.queueRepo.Create(ctx, newEntry)
			if err != nil {
				result.Skipped = append(result.Skipped, &models.GenerateJobSkipped{
					TypeID:   *origEntry.ProductTypeID,
					TypeName: orig.ProductName,
					Reason:   "failed to create queue entry: " + err.Error(),
				})
				continue
			}

			created.BlueprintName = orig.BlueprintName
			created.ProductName = orig.ProductName
			result.Created = append(result.Created, created)
		}
	} else {
		// No assignment — create one entry per merged job (original behaviour)
		for _, pj := range wr.MergedJobs {
			pj.Entry.UserID = *args.User
			pj.Entry.Notes = &note
			pj.Entry.PlanRunID = &run.ID

			created, err := c.queueRepo.Create(ctx, pj.Entry)
			if err != nil {
				result.Skipped = append(result.Skipped, &models.GenerateJobSkipped{
					TypeID:   *pj.Entry.ProductTypeID,
					TypeName: pj.ProductName,
					Reason:   "failed to create queue entry: " + err.Error(),
				})
				continue
			}

			created.BlueprintName = pj.BlueprintName
			created.ProductName = pj.ProductName
			result.Created = append(result.Created, created)
		}
	}

	// Phase 2: Generate transport jobs if plan has transport settings
	if plan.TransportFulfillment != nil {
		c.generateTransportJobs(ctx, plan, stepsByID, childStepsByParent, wr.StepProduction, wr.StepDepths, jitaPrices, run, *args.User, result)
	}

	return result, nil
}

// generateTransportJobs detects cross-station dependencies and creates transport jobs.
func (c *ProductionPlans) generateTransportJobs(
	ctx context.Context,
	plan *models.ProductionPlan,
	stepsByID map[int64]*models.ProductionPlanStep,
	childStepsByParent map[int64][]*models.ProductionPlanStep,
	stepProduction map[int64]*services.StepProductionData,
	stepDepths map[int64]int,
	jitaPrices map[int64]*models.MarketPrice,
	run *models.ProductionPlanRun,
	userID int64,
	result *models.GenerateJobsResult,
) {
	// Resolve user stations for each step (cached)
	stationCache := make(map[int64]*models.UserStation)
	for _, step := range stepsByID {
		if step.UserStationID == nil {
			continue
		}
		if _, ok := stationCache[*step.UserStationID]; !ok {
			station, err := c.stationRepo.GetByID(ctx, *step.UserStationID, userID)
			if err == nil && station != nil {
				stationCache[*step.UserStationID] = station
			}
		}
	}

	// Collect transport needs grouped by origin→destination route
	type transportNeed struct {
		originStation *models.UserStation
		destStation   *models.UserStation
		items         []*models.TransportJobItem
		maxChildDepth int
	}
	needsByRoute := make(map[string]*transportNeed)

	for parentID, children := range childStepsByParent {
		parent := stepsByID[parentID]
		if parent == nil || parent.UserStationID == nil {
			continue
		}
		parentStation := stationCache[*parent.UserStationID]
		if parentStation == nil {
			continue
		}

		for _, child := range children {
			if child.UserStationID == nil {
				continue
			}
			childStation := stationCache[*child.UserStationID]
			if childStation == nil {
				continue
			}
			if childStation.StationID == parentStation.StationID {
				continue // same station, no transport needed
			}

			prod := stepProduction[child.ID]
			if prod == nil {
				continue
			}

			// Calculate item value from Jita prices
			estimatedValue := 0.0
			if p, ok := jitaPrices[prod.ProductTypeID]; ok && p.SellPrice != nil {
				estimatedValue = *p.SellPrice * float64(prod.TotalQuantity)
			}

			key := fmt.Sprintf("%d-%d", childStation.StationID, parentStation.StationID)
			if needsByRoute[key] == nil {
				needsByRoute[key] = &transportNeed{
					originStation: childStation,
					destStation:   parentStation,
					items:         []*models.TransportJobItem{},
				}
			}
			// Track the max child depth for sort ordering
			if d, ok := stepDepths[child.ID]; ok && d > needsByRoute[key].maxChildDepth {
				needsByRoute[key].maxChildDepth = d
			}
			needsByRoute[key].items = append(needsByRoute[key].items, &models.TransportJobItem{
				TypeID:         prod.ProductTypeID,
				Quantity:       prod.TotalQuantity,
				VolumeM3:       prod.ProductVolume * float64(prod.TotalQuantity),
				EstimatedValue: estimatedValue,
			})
		}
	}

	if len(needsByRoute) == 0 {
		return
	}

	// Resolve transport profile
	var profile *models.TransportProfile
	if plan.TransportProfileID != nil {
		p, err := c.profilesRepo.GetByID(ctx, *plan.TransportProfileID, userID)
		if err == nil && p != nil {
			profile = p
		}
	}
	if profile == nil && plan.TransportMethod != nil {
		p, err := c.profilesRepo.GetDefaultByMethod(ctx, userID, *plan.TransportMethod)
		if err == nil && p != nil {
			profile = p
		}
	}

	fulfillment := *plan.TransportFulfillment
	method := ""
	if plan.TransportMethod != nil {
		method = *plan.TransportMethod
	} else if profile != nil {
		method = profile.TransportMethod
	}

	routePreference := "shortest"
	if profile != nil && profile.RoutePreference != "" {
		routePreference = profile.RoutePreference
	}

	// Create a transport job for each route
	for _, need := range needsByRoute {
		totalVolume := 0.0
		totalCollateral := 0.0
		for _, item := range need.items {
			totalVolume += item.VolumeM3
			totalCollateral += item.EstimatedValue
		}

		var estimatedCost float64
		var jumps int
		var distanceLY *float64
		var jfRouteID *int64

		if fulfillment == "self_haul" && profile != nil {
			switch method {
			case "freighter", "dst", "blockade_runner":
				route, err := c.esiClient.GetRoute(ctx, need.originStation.SolarSystemID, need.destStation.SolarSystemID, routePreference)
				if err == nil && len(route) > 1 {
					jumps = len(route) - 1
					costResult := calculator.CalculateGateTransportCost(&calculator.GateTransportCostParams{
						TotalVolumeM3:    totalVolume,
						TotalCollateral:  totalCollateral,
						Jumps:            jumps,
						CargoM3:          profile.CargoM3,
						RatePerM3PerJump: profile.RatePerM3PerJump,
						CollateralRate:   profile.CollateralRate,
					})
					estimatedCost = costResult.Cost
				}

			case "jump_freighter":
				jfRoute, err := c.jfRoutesRepo.FindBySystemPair(ctx, userID, need.originStation.SolarSystemID, need.destStation.SolarSystemID)
				if err == nil && jfRoute != nil {
					d := jfRoute.TotalDistanceLY
					distanceLY = &d
					jumps = len(jfRoute.Waypoints) - 1
					if jumps < 0 {
						jumps = 0
					}
					jfRouteID = &jfRoute.ID

					// Get isotope price
					isotopePrice := 0.0
					if profile.FuelTypeID != nil {
						if p, ok := jitaPrices[*profile.FuelTypeID]; ok {
							switch profile.CollateralPriceBasis {
							case "buy":
								if p.BuyPrice != nil {
									isotopePrice = *p.BuyPrice
								}
							case "split":
								buy, sell := 0.0, 0.0
								if p.BuyPrice != nil {
									buy = *p.BuyPrice
								}
								if p.SellPrice != nil {
									sell = *p.SellPrice
								}
								isotopePrice = (buy + sell) / 2
							default:
								if p.SellPrice != nil {
									isotopePrice = *p.SellPrice
								}
							}
						}
					}

					fuelPerLY := 0.0
					if profile.FuelPerLY != nil {
						fuelPerLY = *profile.FuelPerLY
					}

					costResult := calculator.CalculateJFTransportCost(&calculator.JFTransportCostParams{
						TotalVolumeM3:         totalVolume,
						TotalCollateral:       totalCollateral,
						CargoM3:               profile.CargoM3,
						CollateralRate:        profile.CollateralRate,
						FuelPerLY:             fuelPerLY,
						FuelConservationLevel: profile.FuelConservationLevel,
						IsotopePrice:          isotopePrice,
						Waypoints:             jfRoute.Waypoints,
					})
					estimatedCost = costResult.Cost
				}
			}
		} else if fulfillment == "courier_contract" || fulfillment == "contact_haul" {
			estimatedCost = calculator.CalculateCourierCost(&calculator.CourierCostParams{
				TotalVolumeM3:        totalVolume,
				TotalCollateral:      totalCollateral,
				CourierRatePerM3:     plan.CourierRatePerM3,
				CourierCollateralRate: plan.CourierCollateralRate,
			})
		}

		var profileID *int64
		if profile != nil {
			profileID = &profile.ID
		}

		note := fmt.Sprintf("Auto-generated from plan: %s", plan.Name)
		job := &models.TransportJob{
			UserID:               userID,
			OriginStationID:      need.originStation.StationID,
			DestinationStationID: need.destStation.StationID,
			OriginSystemID:       need.originStation.SolarSystemID,
			DestinationSystemID:  need.destStation.SolarSystemID,
			TransportMethod:      method,
			RoutePreference:      routePreference,
			TotalVolumeM3:        totalVolume,
			TotalCollateral:      totalCollateral,
			EstimatedCost:        estimatedCost,
			Jumps:                jumps,
			DistanceLY:           distanceLY,
			JFRouteID:            jfRouteID,
			FulfillmentType:      fulfillment,
			TransportProfileID:   profileID,
			PlanRunID:            &run.ID,
			Notes:                &note,
			Items:                need.items,
		}

		created, err := c.transportJobRepo.Create(ctx, job)
		if err != nil {
			result.Skipped = append(result.Skipped, &models.GenerateJobSkipped{
				Reason: fmt.Sprintf("failed to create transport job (%s → %s): %s", need.originStation.StationName, need.destStation.StationName, err.Error()),
			})
			continue
		}

		// Create corresponding queue entry
		// Sort order: child depth * 2 - 1 places transport between child build and parent build
		transportSortOrder := need.maxChildDepth*2 - 1
		if transportSortOrder < 0 {
			transportSortOrder = 0
		}
		queueEntry, err := c.queueRepo.Create(ctx, &models.IndustryJobQueueEntry{
			UserID:         userID,
			Activity:       "transport",
			EstimatedCost:  &created.EstimatedCost,
			TransportJobID: &created.ID,
			PlanRunID:      &run.ID,
			SortOrder:      transportSortOrder,
			StationName:    need.originStation.StationName + " → " + need.destStation.StationName,
		})
		if err != nil {
			result.Skipped = append(result.Skipped, &models.GenerateJobSkipped{
				Reason: "failed to create transport queue entry: " + err.Error(),
			})
			continue
		}

		// Link queue entry back to transport job
		if err := c.transportJobRepo.SetQueueEntryID(ctx, created.ID, queueEntry.ID); err != nil {
			result.Skipped = append(result.Skipped, &models.GenerateJobSkipped{
				Reason: "failed to link transport queue entry: " + err.Error(),
			})
			continue
		}
		created.QueueEntryID = &queueEntry.ID

		result.TransportJobs = append(result.TransportJobs, created)
		result.Created = append(result.Created, queueEntry)
	}
}

// GetAllRuns lists all runs across all plans for the user.
func (c *ProductionPlans) GetAllRuns(args *web.HandlerArgs) (any, *web.HttpError) {
	runs, err := c.runsRepo.GetByUser(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get all plan runs")}
	}
	return runs, nil
}

// CancelPlanRun cancels all planned jobs in a run. Active/completed jobs are untouched.
func (c *ProductionPlans) CancelPlanRun(args *web.HandlerArgs) (any, *web.HttpError) {
	runID, err := parseID(args.Params["runId"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid run ID")}
	}

	cancelled, err := c.runsRepo.CancelPlannedJobs(args.Request.Context(), runID, *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to cancel plan run")}
	}

	return map[string]any{"status": "cancelled", "jobsCancelled": cancelled}, nil
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

// --- Plan Preview ---

type previewPlanRequest struct {
	Quantity int `json:"quantity"`
}

// PreviewPlan simulates job assignment at every parallelism level and returns
// estimated wall-clock durations without creating any database records.
func (c *ProductionPlans) PreviewPlan(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()

	planID, err := parseID(args.Params["id"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid plan ID")}
	}

	var req previewPlanRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}
	if req.Quantity <= 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("quantity must be positive")}
	}

	// Fetch the plan
	plan, err := c.plansRepo.GetByID(ctx, planID, *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get plan")}
	}
	if plan == nil {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.New("production plan not found")}
	}

	// Fetch market data
	jitaPrices, err := c.marketRepo.GetAllJitaPrices(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get Jita prices")}
	}
	adjustedPrices, err := c.marketRepo.GetAllAdjustedPrices(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get adjusted prices")}
	}

	// Walk and merge steps
	wr, err := services.WalkAndMergeSteps(ctx, c.sdeRepo, plan, req.Quantity, jitaPrices, adjustedPrices)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: err}
	}

	// Discover eligible characters
	characterNames, err := c.characterRepo.GetNames(ctx, *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get character names")}
	}

	allSkills, err := c.skillsRepo.GetSkillsForUser(ctx, *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get character skills")}
	}

	// Build skillsByCharacter filtered to industry skill IDs
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

	// No slot usage data for preview — we're simulating fresh slots
	capacities := calculator.BuildCharacterCapacities(characterNames, skillsByCharacter, nil)

	result := &models.PlanPreviewResult{
		Options:            []*models.PlanPreviewOption{},
		EligibleCharacters: len(capacities),
		TotalJobs:          len(wr.MergedJobs),
	}

	if len(capacities) == 0 || len(wr.MergedJobs) == 0 {
		return result, nil
	}

	// Generate one option per parallelism level
	for p := 1; p <= len(capacities); p++ {
		assigned, _ := services.SimulateAssignment(wr.MergedJobs, capacities, p)
		wallClock := services.EstimateWallClock(assigned, capacities[:p])

		// Build per-character info
		charJobCount := make(map[int64]int)
		charDuration := make(map[int64]int)
		for _, aj := range assigned {
			charJobCount[aj.CharacterID]++
			if aj.DurationSec > charDuration[aj.CharacterID] {
				charDuration[aj.CharacterID] = aj.DurationSec
			}
		}

		chars := []*models.PreviewCharacterInfo{}
		for _, cap := range capacities[:p] {
			chars = append(chars, &models.PreviewCharacterInfo{
				CharacterID:    cap.CharacterID,
				Name:           cap.CharacterName,
				JobCount:       charJobCount[cap.CharacterID],
				DurationSec:    charDuration[cap.CharacterID],
				MfgSlotsUsed:   cap.MfgSlotsUsed,
				MfgSlotsMax:    cap.MfgSlotsMax,
				ReactSlotsUsed: cap.ReactSlotsUsed,
				ReactSlotsMax:  cap.ReactSlotsMax,
			})
		}

		result.Options = append(result.Options, &models.PlanPreviewOption{
			Parallelism:            p,
			EstimatedDurationSec:   wallClock,
			EstimatedDurationLabel: models.FormatDurationLabel(wallClock),
			Characters:             chars,
		})
	}

	return result, nil
}

