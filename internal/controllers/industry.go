package controllers

import (
	"context"
	"encoding/json"

	"github.com/annymsMthd/industry-tool/internal/calculator"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type IndustryJobsRepository interface {
	GetActiveJobs(ctx context.Context, userID int64) ([]*models.IndustryJob, error)
	GetAllJobs(ctx context.Context, userID int64) ([]*models.IndustryJob, error)
}

type IndustryJobQueueRepository interface {
	Create(ctx context.Context, entry *models.IndustryJobQueueEntry) (*models.IndustryJobQueueEntry, error)
	GetByUser(ctx context.Context, userID int64) ([]*models.IndustryJobQueueEntry, error)
	Update(ctx context.Context, id, userID int64, entry *models.IndustryJobQueueEntry) (*models.IndustryJobQueueEntry, error)
	Cancel(ctx context.Context, id, userID int64) error
}

type IndustrySDERepository interface {
	GetManufacturingBlueprint(ctx context.Context, blueprintTypeID int64) (*repositories.ManufacturingBlueprintRow, error)
	GetManufacturingMaterials(ctx context.Context, blueprintTypeID int64) ([]*repositories.ManufacturingMaterialRow, error)
	SearchBlueprints(ctx context.Context, query string, activity string, limit int) ([]*repositories.BlueprintSearchRow, error)
	GetManufacturingSystems(ctx context.Context) ([]*models.ReactionSystem, error)
}

type IndustryMarketRepository interface {
	GetAllJitaPrices(ctx context.Context) (map[int64]*models.MarketPrice, error)
	GetAllAdjustedPrices(ctx context.Context) (map[int64]float64, error)
}

type IndustryCostIndicesRepository interface {
	GetCostIndex(ctx context.Context, systemID int64, activity string) (*models.IndustryCostIndex, error)
}

type Industry struct {
	jobsRepo        IndustryJobsRepository
	queueRepo       IndustryJobQueueRepository
	sdeRepo         IndustrySDERepository
	marketRepo      IndustryMarketRepository
	costIndicesRepo IndustryCostIndicesRepository
}

func NewIndustry(
	router Routerer,
	jobsRepo IndustryJobsRepository,
	queueRepo IndustryJobQueueRepository,
	sdeRepo IndustrySDERepository,
	marketRepo IndustryMarketRepository,
	costIndicesRepo IndustryCostIndicesRepository,
) *Industry {
	c := &Industry{
		jobsRepo:        jobsRepo,
		queueRepo:       queueRepo,
		sdeRepo:         sdeRepo,
		marketRepo:      marketRepo,
		costIndicesRepo: costIndicesRepo,
	}

	// User-scoped endpoints
	router.RegisterRestAPIRoute("/v1/industry/jobs", web.AuthAccessUser, c.GetActiveJobs, "GET")
	router.RegisterRestAPIRoute("/v1/industry/jobs/all", web.AuthAccessUser, c.GetAllJobs, "GET")
	router.RegisterRestAPIRoute("/v1/industry/queue", web.AuthAccessUser, c.GetQueue, "GET")
	router.RegisterRestAPIRoute("/v1/industry/queue", web.AuthAccessUser, c.CreateQueueEntry, "POST")
	router.RegisterRestAPIRoute("/v1/industry/queue/{id}", web.AuthAccessUser, c.UpdateQueueEntry, "PUT")
	router.RegisterRestAPIRoute("/v1/industry/queue/{id}", web.AuthAccessUser, c.CancelQueueEntry, "DELETE")

	// Backend-scoped endpoints (no user required)
	router.RegisterRestAPIRoute("/v1/industry/calculate", web.AuthAccessBackend, c.Calculate, "POST")
	router.RegisterRestAPIRoute("/v1/industry/blueprints", web.AuthAccessBackend, c.SearchBlueprints, "GET")
	router.RegisterRestAPIRoute("/v1/industry/systems", web.AuthAccessBackend, c.GetSystems, "GET")

	return c
}

func (c *Industry) GetActiveJobs(args *web.HandlerArgs) (any, *web.HttpError) {
	jobs, err := c.jobsRepo.GetActiveJobs(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get active jobs")}
	}
	return jobs, nil
}

func (c *Industry) GetAllJobs(args *web.HandlerArgs) (any, *web.HttpError) {
	jobs, err := c.jobsRepo.GetAllJobs(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get all jobs")}
	}
	return jobs, nil
}

func (c *Industry) GetQueue(args *web.HandlerArgs) (any, *web.HttpError) {
	entries, err := c.queueRepo.GetByUser(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get job queue")}
	}
	return entries, nil
}

type createQueueRequest struct {
	BlueprintTypeID   int64    `json:"blueprint_type_id"`
	Activity          string   `json:"activity"`
	Runs              int      `json:"runs"`
	MELevel           int      `json:"me_level"`
	TELevel           int      `json:"te_level"`
	CharacterID       *int64   `json:"character_id"`
	SystemID          *int64   `json:"system_id"`
	FacilityTax       float64  `json:"facility_tax"`
	Structure         string   `json:"structure"`
	Rig               string   `json:"rig"`
	Security          string   `json:"security"`
	IndustrySkill     int      `json:"industry_skill"`
	AdvIndustrySkill  int      `json:"adv_industry_skill"`
	ProductTypeID     *int64   `json:"product_type_id"`
	Notes             *string  `json:"notes"`
}

func (c *Industry) CreateQueueEntry(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()

	var req createQueueRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	if req.BlueprintTypeID <= 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("blueprint_type_id is required")}
	}
	if req.Activity == "" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("activity is required")}
	}
	if req.Runs <= 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("runs must be positive")}
	}

	// Calculate estimated cost and duration if this is manufacturing
	var estimatedCost *float64
	var estimatedDuration *int
	var productTypeID *int64

	if req.Activity == "manufacturing" {
		calcResult, httpErr := c.calculateForBlueprint(ctx, req.BlueprintTypeID, req.Runs, req.MELevel, req.TELevel,
			req.IndustrySkill, req.AdvIndustrySkill, req.SystemID, req.FacilityTax,
			withDefault(req.Structure, "station"), withDefault(req.Rig, "none"), withDefault(req.Security, "high"))
		if httpErr != nil {
			// Non-fatal: still create the entry without estimates
			estimatedCost = nil
			estimatedDuration = nil
		} else {
			estimatedCost = &calcResult.TotalCost
			estimatedDuration = &calcResult.TotalDuration
			productTypeID = &calcResult.ProductTypeID
		}
	}

	// Use explicitly provided product_type_id if given
	if req.ProductTypeID != nil {
		productTypeID = req.ProductTypeID
	}

	entry := &models.IndustryJobQueueEntry{
		UserID:            *args.User,
		CharacterID:       req.CharacterID,
		BlueprintTypeID:   req.BlueprintTypeID,
		Activity:          req.Activity,
		Runs:              req.Runs,
		MELevel:           req.MELevel,
		TELevel:           req.TELevel,
		SystemID:          req.SystemID,
		FacilityTax:       req.FacilityTax,
		ProductTypeID:     productTypeID,
		EstimatedCost:     estimatedCost,
		EstimatedDuration: estimatedDuration,
		Notes:             req.Notes,
	}

	created, err := c.queueRepo.Create(ctx, entry)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to create job queue entry")}
	}

	return created, nil
}

func (c *Industry) UpdateQueueEntry(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()

	id, err := parseID(args.Params["id"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid queue entry ID")}
	}

	var req createQueueRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	// Recalculate estimates for manufacturing
	var estimatedCost *float64
	var estimatedDuration *int
	var productTypeID *int64

	if req.Activity == "manufacturing" {
		calcResult, httpErr := c.calculateForBlueprint(ctx, req.BlueprintTypeID, req.Runs, req.MELevel, req.TELevel,
			req.IndustrySkill, req.AdvIndustrySkill, req.SystemID, req.FacilityTax,
			withDefault(req.Structure, "station"), withDefault(req.Rig, "none"), withDefault(req.Security, "high"))
		if httpErr == nil {
			estimatedCost = &calcResult.TotalCost
			estimatedDuration = &calcResult.TotalDuration
			productTypeID = &calcResult.ProductTypeID
		}
	}

	if req.ProductTypeID != nil {
		productTypeID = req.ProductTypeID
	}

	entry := &models.IndustryJobQueueEntry{
		CharacterID:       req.CharacterID,
		BlueprintTypeID:   req.BlueprintTypeID,
		Activity:          req.Activity,
		Runs:              req.Runs,
		MELevel:           req.MELevel,
		TELevel:           req.TELevel,
		SystemID:          req.SystemID,
		FacilityTax:       req.FacilityTax,
		ProductTypeID:     productTypeID,
		EstimatedCost:     estimatedCost,
		EstimatedDuration: estimatedDuration,
		Notes:             req.Notes,
	}

	updated, err := c.queueRepo.Update(ctx, id, *args.User, entry)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to update job queue entry")}
	}
	if updated == nil {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.New("queue entry not found or not editable")}
	}

	return updated, nil
}

func (c *Industry) CancelQueueEntry(args *web.HandlerArgs) (any, *web.HttpError) {
	id, err := parseID(args.Params["id"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid queue entry ID")}
	}

	err = c.queueRepo.Cancel(args.Request.Context(), id, *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.Wrap(err, "failed to cancel queue entry")}
	}

	return map[string]string{"status": "cancelled"}, nil
}

type calculateRequest struct {
	BlueprintTypeID  int64   `json:"blueprint_type_id"`
	Runs             int     `json:"runs"`
	MELevel          int     `json:"me_level"`
	TELevel          int     `json:"te_level"`
	IndustrySkill    int     `json:"industry_skill"`
	AdvIndustrySkill int     `json:"adv_industry_skill"`
	SystemID         *int64  `json:"system_id"`
	FacilityTax      float64 `json:"facility_tax"`
	Structure        string  `json:"structure"`
	Rig              string  `json:"rig"`
	Security         string  `json:"security"`
}

func (c *Industry) Calculate(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()

	var req calculateRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	if req.BlueprintTypeID <= 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("blueprint_type_id is required")}
	}
	if req.Runs <= 0 {
		req.Runs = 1
	}

	result, httpErr := c.calculateForBlueprint(ctx, req.BlueprintTypeID, req.Runs, req.MELevel, req.TELevel,
		req.IndustrySkill, req.AdvIndustrySkill, req.SystemID, req.FacilityTax,
		withDefault(req.Structure, "station"), withDefault(req.Rig, "none"), withDefault(req.Security, "high"))
	if httpErr != nil {
		return nil, httpErr
	}

	return result, nil
}

func (c *Industry) SearchBlueprints(args *web.HandlerArgs) (any, *web.HttpError) {
	q := args.Request.URL.Query()
	query := q.Get("q")
	if query == "" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("query parameter 'q' is required")}
	}

	activity := q.Get("activity")
	limit := int(parseInt64(q.Get("limit"), 20))
	if limit > 100 {
		limit = 100
	}

	results, err := c.sdeRepo.SearchBlueprints(args.Request.Context(), query, activity, limit)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to search blueprints")}
	}

	if results == nil {
		results = []*repositories.BlueprintSearchRow{}
	}

	return results, nil
}

func (c *Industry) GetSystems(args *web.HandlerArgs) (any, *web.HttpError) {
	systems, err := c.sdeRepo.GetManufacturingSystems(args.Request.Context())
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get manufacturing systems")}
	}

	if systems == nil {
		systems = []*models.ReactionSystem{}
	}

	return systems, nil
}

// calculateForBlueprint performs the full manufacturing calculation for a given blueprint.
func (c *Industry) calculateForBlueprint(
	ctx context.Context,
	blueprintTypeID int64,
	runs, meLevel, teLevel, industrySkill, advIndustrySkill int,
	systemID *int64,
	facilityTax float64,
	structure, rig, security string,
) (*models.ManufacturingCalcResult, *web.HttpError) {
	blueprint, err := c.sdeRepo.GetManufacturingBlueprint(ctx, blueprintTypeID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get blueprint")}
	}
	if blueprint == nil {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.New("blueprint not found")}
	}

	materials, err := c.sdeRepo.GetManufacturingMaterials(ctx, blueprintTypeID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get materials")}
	}

	jitaPrices, err := c.marketRepo.GetAllJitaPrices(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get Jita prices")}
	}

	adjustedPrices, err := c.marketRepo.GetAllAdjustedPrices(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get adjusted prices")}
	}

	var costIndex float64
	if systemID != nil && *systemID > 0 {
		idx, err := c.costIndicesRepo.GetCostIndex(ctx, *systemID, "manufacturing")
		if err != nil {
			return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get cost index")}
		}
		if idx != nil {
			costIndex = idx.CostIndex
		}
	}

	params := &calculator.ManufacturingParams{
		BlueprintME:      meLevel,
		BlueprintTE:      teLevel,
		Runs:             runs,
		Structure:        structure,
		Rig:              rig,
		Security:         security,
		IndustrySkill:    industrySkill,
		AdvIndustrySkill: advIndustrySkill,
		FacilityTax:      facilityTax,
	}
	if systemID != nil {
		params.SystemID = *systemID
	}

	data := &calculator.ManufacturingData{
		Blueprint:      blueprint,
		Materials:      materials,
		CostIndex:      costIndex,
		AdjustedPrices: adjustedPrices,
		JitaPrices:     jitaPrices,
	}

	result := calculator.CalculateManufacturingJob(params, data)
	return result, nil
}
