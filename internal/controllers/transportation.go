package controllers

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/annymsMthd/industry-tool/internal/calculator"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

// Repository interfaces

type TransportProfilesRepository interface {
	GetByUser(ctx context.Context, userID int64) ([]*models.TransportProfile, error)
	GetByID(ctx context.Context, id, userID int64) (*models.TransportProfile, error)
	GetDefaultByMethod(ctx context.Context, userID int64, method string) (*models.TransportProfile, error)
	Create(ctx context.Context, p *models.TransportProfile) (*models.TransportProfile, error)
	Update(ctx context.Context, p *models.TransportProfile) (*models.TransportProfile, error)
	Delete(ctx context.Context, id, userID int64) error
}

type JFRoutesRepository interface {
	GetByUser(ctx context.Context, userID int64) ([]*models.JFRoute, error)
	GetByID(ctx context.Context, id, userID int64) (*models.JFRoute, error)
	Create(ctx context.Context, route *models.JFRoute, systemCoords map[int64]*models.SolarSystem) (*models.JFRoute, error)
	Update(ctx context.Context, route *models.JFRoute, systemCoords map[int64]*models.SolarSystem) (*models.JFRoute, error)
	Delete(ctx context.Context, id, userID int64) error
}

type TransportJobsRepository interface {
	GetByUser(ctx context.Context, userID int64) ([]*models.TransportJob, error)
	GetByID(ctx context.Context, id, userID int64) (*models.TransportJob, error)
	Create(ctx context.Context, job *models.TransportJob) (*models.TransportJob, error)
	UpdateStatus(ctx context.Context, id, userID int64, status string) error
	SetQueueEntryID(ctx context.Context, id int64, queueEntryID int64) error
	Cancel(ctx context.Context, id, userID int64) error
}

type TransportTriggerConfigRepository interface {
	GetByUser(ctx context.Context, userID int64) ([]*models.TransportTriggerConfig, error)
	Upsert(ctx context.Context, c *models.TransportTriggerConfig) (*models.TransportTriggerConfig, error)
}

type TransportJobQueueRepository interface {
	Create(ctx context.Context, entry *models.IndustryJobQueueEntry) (*models.IndustryJobQueueEntry, error)
}

type TransportMarketPricesRepository interface {
	GetAllJitaPrices(ctx context.Context) (map[int64]*models.MarketPrice, error)
}

type TransportSolarSystemsRepository interface {
	GetByIDs(ctx context.Context, ids []int64) ([]*models.SolarSystem, error)
	Search(ctx context.Context, query string, limit int) ([]*models.SolarSystem, error)
}

type TransportEsiClient interface {
	GetRoute(ctx context.Context, origin, destination int64, flag string) ([]int32, error)
}

// Controller

type Transportation struct {
	profilesRepo   TransportProfilesRepository
	jfRoutesRepo   JFRoutesRepository
	jobsRepo       TransportJobsRepository
	triggerRepo    TransportTriggerConfigRepository
	queueRepo      TransportJobQueueRepository
	marketRepo     TransportMarketPricesRepository
	solarSysRepo   TransportSolarSystemsRepository
	esiClient      TransportEsiClient
}

func NewTransportation(
	router Routerer,
	profilesRepo TransportProfilesRepository,
	jfRoutesRepo JFRoutesRepository,
	jobsRepo TransportJobsRepository,
	triggerRepo TransportTriggerConfigRepository,
	queueRepo TransportJobQueueRepository,
	marketRepo TransportMarketPricesRepository,
	solarSysRepo TransportSolarSystemsRepository,
	esiClient TransportEsiClient,
) *Transportation {
	c := &Transportation{
		profilesRepo: profilesRepo,
		jfRoutesRepo: jfRoutesRepo,
		jobsRepo:     jobsRepo,
		triggerRepo:  triggerRepo,
		queueRepo:    queueRepo,
		marketRepo:   marketRepo,
		solarSysRepo: solarSysRepo,
		esiClient:    esiClient,
	}

	// Transport Profiles
	router.RegisterRestAPIRoute("/v1/transport/profiles", web.AuthAccessUser, c.GetProfiles, "GET")
	router.RegisterRestAPIRoute("/v1/transport/profiles", web.AuthAccessUser, c.CreateProfile, "POST")
	router.RegisterRestAPIRoute("/v1/transport/profiles/{id:[0-9]+}", web.AuthAccessUser, c.UpdateProfile, "PUT")
	router.RegisterRestAPIRoute("/v1/transport/profiles/{id:[0-9]+}", web.AuthAccessUser, c.DeleteProfile, "DELETE")

	// JF Routes
	router.RegisterRestAPIRoute("/v1/transport/jf-routes", web.AuthAccessUser, c.GetJFRoutes, "GET")
	router.RegisterRestAPIRoute("/v1/transport/jf-routes", web.AuthAccessUser, c.CreateJFRoute, "POST")
	router.RegisterRestAPIRoute("/v1/transport/jf-routes/{id:[0-9]+}", web.AuthAccessUser, c.UpdateJFRoute, "PUT")
	router.RegisterRestAPIRoute("/v1/transport/jf-routes/{id:[0-9]+}", web.AuthAccessUser, c.DeleteJFRoute, "DELETE")

	// Transport Jobs
	router.RegisterRestAPIRoute("/v1/transport/jobs", web.AuthAccessUser, c.GetJobs, "GET")
	router.RegisterRestAPIRoute("/v1/transport/jobs", web.AuthAccessUser, c.CreateJob, "POST")
	router.RegisterRestAPIRoute("/v1/transport/jobs/{id:[0-9]+}/status", web.AuthAccessUser, c.UpdateJobStatus, "POST")

	// Route Calculation
	router.RegisterRestAPIRoute("/v1/transport/route", web.AuthAccessUser, c.GetRoute, "GET")

	// Solar System Search
	router.RegisterRestAPIRoute("/v1/transport/systems/search", web.AuthAccessUser, c.SearchSystems, "GET")

	// Trigger Config
	router.RegisterRestAPIRoute("/v1/transport/trigger-config", web.AuthAccessUser, c.GetTriggerConfig, "GET")
	router.RegisterRestAPIRoute("/v1/transport/trigger-config", web.AuthAccessUser, c.UpsertTriggerConfig, "PUT")

	return c
}

// ---- Transport Profiles ----

func (c *Transportation) GetProfiles(args *web.HandlerArgs) (any, *web.HttpError) {
	profiles, err := c.profilesRepo.GetByUser(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get transport profiles")}
	}
	return profiles, nil
}

type createProfileRequest struct {
	Name                  string   `json:"name"`
	TransportMethod       string   `json:"transportMethod"`
	CharacterID           *int64   `json:"characterId"`
	CargoM3               float64  `json:"cargoM3"`
	RatePerM3PerJump      float64  `json:"ratePerM3PerJump"`
	CollateralRate        float64  `json:"collateralRate"`
	CollateralPriceBasis  string   `json:"collateralPriceBasis"`
	FuelTypeID            *int64   `json:"fuelTypeId"`
	FuelPerLY             *float64 `json:"fuelPerLy"`
	FuelConservationLevel int      `json:"fuelConservationLevel"`
	RoutePreference       string   `json:"routePreference"`
	IsDefault             bool     `json:"isDefault"`
}

func (c *Transportation) CreateProfile(args *web.HandlerArgs) (any, *web.HttpError) {
	var req createProfileRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	if req.Name == "" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("name is required")}
	}
	if req.TransportMethod == "" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("transportMethod is required")}
	}
	if req.CargoM3 <= 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("cargoM3 must be positive")}
	}

	profile := &models.TransportProfile{
		UserID:                *args.User,
		Name:                  req.Name,
		TransportMethod:       req.TransportMethod,
		CharacterID:           req.CharacterID,
		CargoM3:               req.CargoM3,
		RatePerM3PerJump:      req.RatePerM3PerJump,
		CollateralRate:        req.CollateralRate,
		CollateralPriceBasis:  req.CollateralPriceBasis,
		FuelTypeID:            req.FuelTypeID,
		FuelPerLY:             req.FuelPerLY,
		FuelConservationLevel: req.FuelConservationLevel,
		RoutePreference:       req.RoutePreference,
		IsDefault:             req.IsDefault,
	}

	if profile.CollateralPriceBasis == "" {
		profile.CollateralPriceBasis = "sell"
	}
	if profile.RoutePreference == "" {
		profile.RoutePreference = "shortest"
	}

	created, err := c.profilesRepo.Create(args.Request.Context(), profile)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to create transport profile")}
	}

	return created, nil
}

func (c *Transportation) UpdateProfile(args *web.HandlerArgs) (any, *web.HttpError) {
	id, err := strconv.ParseInt(args.Params["id"], 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid profile ID")}
	}

	var req createProfileRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	profile := &models.TransportProfile{
		ID:                    id,
		UserID:                *args.User,
		Name:                  req.Name,
		TransportMethod:       req.TransportMethod,
		CharacterID:           req.CharacterID,
		CargoM3:               req.CargoM3,
		RatePerM3PerJump:      req.RatePerM3PerJump,
		CollateralRate:        req.CollateralRate,
		CollateralPriceBasis:  req.CollateralPriceBasis,
		FuelTypeID:            req.FuelTypeID,
		FuelPerLY:             req.FuelPerLY,
		FuelConservationLevel: req.FuelConservationLevel,
		RoutePreference:       req.RoutePreference,
		IsDefault:             req.IsDefault,
	}

	if profile.CollateralPriceBasis == "" {
		profile.CollateralPriceBasis = "sell"
	}
	if profile.RoutePreference == "" {
		profile.RoutePreference = "shortest"
	}

	updated, err := c.profilesRepo.Update(args.Request.Context(), profile)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to update transport profile")}
	}
	if updated == nil {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.New("transport profile not found")}
	}

	return updated, nil
}

func (c *Transportation) DeleteProfile(args *web.HandlerArgs) (any, *web.HttpError) {
	id, err := strconv.ParseInt(args.Params["id"], 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid profile ID")}
	}

	if err := c.profilesRepo.Delete(args.Request.Context(), id, *args.User); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to delete transport profile")}
	}

	return map[string]bool{"success": true}, nil
}

// ---- JF Routes ----

func (c *Transportation) GetJFRoutes(args *web.HandlerArgs) (any, *web.HttpError) {
	routes, err := c.jfRoutesRepo.GetByUser(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get JF routes")}
	}
	return routes, nil
}

type jfRouteWaypointRequest struct {
	Sequence int   `json:"sequence"`
	SystemID int64 `json:"systemId"`
}

type createJFRouteRequest struct {
	Name                string                   `json:"name"`
	OriginSystemID      int64                    `json:"originSystemId"`
	DestinationSystemID int64                    `json:"destinationSystemId"`
	Waypoints           []jfRouteWaypointRequest `json:"waypoints"`
}

func (c *Transportation) CreateJFRoute(args *web.HandlerArgs) (any, *web.HttpError) {
	var req createJFRouteRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	if req.Name == "" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("name is required")}
	}
	if len(req.Waypoints) < 2 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("at least 2 waypoints required")}
	}

	// Collect all system IDs for coordinate lookup
	systemIDs := []int64{}
	waypoints := []*models.JFRouteWaypoint{}
	for _, wp := range req.Waypoints {
		systemIDs = append(systemIDs, wp.SystemID)
		waypoints = append(waypoints, &models.JFRouteWaypoint{
			Sequence: wp.Sequence,
			SystemID: wp.SystemID,
		})
	}

	// Fetch coordinates
	systems, err := c.solarSysRepo.GetByIDs(args.Request.Context(), systemIDs)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get solar system coordinates")}
	}
	systemCoords := make(map[int64]*models.SolarSystem)
	for _, s := range systems {
		systemCoords[s.ID] = s
	}

	route := &models.JFRoute{
		UserID:              *args.User,
		Name:                req.Name,
		OriginSystemID:      req.OriginSystemID,
		DestinationSystemID: req.DestinationSystemID,
		Waypoints:           waypoints,
	}

	created, err := c.jfRoutesRepo.Create(args.Request.Context(), route, systemCoords)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to create JF route")}
	}

	return created, nil
}

func (c *Transportation) UpdateJFRoute(args *web.HandlerArgs) (any, *web.HttpError) {
	id, err := strconv.ParseInt(args.Params["id"], 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid route ID")}
	}

	var req createJFRouteRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	systemIDs := []int64{}
	waypoints := []*models.JFRouteWaypoint{}
	for _, wp := range req.Waypoints {
		systemIDs = append(systemIDs, wp.SystemID)
		waypoints = append(waypoints, &models.JFRouteWaypoint{
			Sequence: wp.Sequence,
			SystemID: wp.SystemID,
		})
	}

	systems, err := c.solarSysRepo.GetByIDs(args.Request.Context(), systemIDs)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get solar system coordinates")}
	}
	systemCoords := make(map[int64]*models.SolarSystem)
	for _, s := range systems {
		systemCoords[s.ID] = s
	}

	route := &models.JFRoute{
		ID:                  id,
		UserID:              *args.User,
		Name:                req.Name,
		OriginSystemID:      req.OriginSystemID,
		DestinationSystemID: req.DestinationSystemID,
		Waypoints:           waypoints,
	}

	updated, err := c.jfRoutesRepo.Update(args.Request.Context(), route, systemCoords)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to update JF route")}
	}
	if updated == nil {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.New("JF route not found")}
	}

	return updated, nil
}

func (c *Transportation) DeleteJFRoute(args *web.HandlerArgs) (any, *web.HttpError) {
	id, err := strconv.ParseInt(args.Params["id"], 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid route ID")}
	}

	if err := c.jfRoutesRepo.Delete(args.Request.Context(), id, *args.User); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to delete JF route")}
	}

	return map[string]bool{"success": true}, nil
}

// ---- Transport Jobs ----

func (c *Transportation) GetJobs(args *web.HandlerArgs) (any, *web.HttpError) {
	jobs, err := c.jobsRepo.GetByUser(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get transport jobs")}
	}
	return jobs, nil
}

type transportJobItemRequest struct {
	TypeID         int64   `json:"typeId"`
	Quantity       int     `json:"quantity"`
	VolumeM3       float64 `json:"volumeM3"`
	EstimatedValue float64 `json:"estimatedValue"`
}

type createJobRequest struct {
	OriginStationID      int64                     `json:"originStationId"`
	DestinationStationID int64                     `json:"destinationStationId"`
	OriginSystemID       int64                     `json:"originSystemId"`
	DestinationSystemID  int64                     `json:"destinationSystemId"`
	TransportMethod      string                    `json:"transportMethod"`
	RoutePreference      string                    `json:"routePreference"`
	FulfillmentType      string                    `json:"fulfillmentType"`
	TransportProfileID   *int64                    `json:"transportProfileId"`
	JFRouteID            *int64                    `json:"jfRouteId"`
	Notes                *string                   `json:"notes"`
	Items                []transportJobItemRequest `json:"items"`
}

func (c *Transportation) CreateJob(args *web.HandlerArgs) (any, *web.HttpError) {
	var req createJobRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	if req.OriginStationID <= 0 || req.DestinationStationID <= 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("origin and destination station IDs are required")}
	}
	if req.TransportMethod == "" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("transportMethod is required")}
	}

	ctx := args.Request.Context()

	// Build items
	items := []*models.TransportJobItem{}
	totalVolume := 0.0
	totalValue := 0.0
	for _, item := range req.Items {
		items = append(items, &models.TransportJobItem{
			TypeID:         item.TypeID,
			Quantity:       item.Quantity,
			VolumeM3:       item.VolumeM3,
			EstimatedValue: item.EstimatedValue,
		})
		totalVolume += item.VolumeM3
		totalValue += item.EstimatedValue
	}

	// Calculate cost based on transport method and fulfillment
	var estimatedCost float64
	var jumps int
	var distanceLY *float64

	if req.RoutePreference == "" {
		req.RoutePreference = "shortest"
	}
	if req.FulfillmentType == "" {
		req.FulfillmentType = "self_haul"
	}

	if req.FulfillmentType == "self_haul" && req.TransportProfileID != nil {
		profile, err := c.profilesRepo.GetByID(ctx, *req.TransportProfileID, *args.User)
		if err != nil {
			return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get transport profile")}
		}
		if profile == nil {
			return nil, &web.HttpError{StatusCode: 400, Error: errors.New("transport profile not found")}
		}

		switch req.TransportMethod {
		case "freighter", "dst", "blockade_runner":
			// Gate-based: use ESI route for jump count
			route, err := c.esiClient.GetRoute(ctx, req.OriginSystemID, req.DestinationSystemID, req.RoutePreference)
			if err != nil {
				return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to calculate route")}
			}
			jumps = len(route) - 1
			if jumps < 0 {
				jumps = 0
			}

			result := calculator.CalculateGateTransportCost(&calculator.GateTransportCostParams{
				TotalVolumeM3:    totalVolume,
				TotalCollateral:  totalValue,
				Jumps:            jumps,
				CargoM3:          profile.CargoM3,
				RatePerM3PerJump: profile.RatePerM3PerJump,
				CollateralRate:   profile.CollateralRate,
			})
			estimatedCost = result.Cost

		case "jump_freighter":
			if req.JFRouteID != nil {
				jfRoute, err := c.jfRoutesRepo.GetByID(ctx, *req.JFRouteID, *args.User)
				if err != nil {
					return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get JF route")}
				}
				if jfRoute != nil {
					d := jfRoute.TotalDistanceLY
					distanceLY = &d
					jumps = len(jfRoute.Waypoints) - 1

					// Get isotope price
					isotopePrice := 0.0
					if profile.FuelTypeID != nil {
						prices, err := c.marketRepo.GetAllJitaPrices(ctx)
						if err == nil {
							if p, ok := prices[*profile.FuelTypeID]; ok {
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
					}

					fuelPerLY := 0.0
					if profile.FuelPerLY != nil {
						fuelPerLY = *profile.FuelPerLY
					}

					result := calculator.CalculateJFTransportCost(&calculator.JFTransportCostParams{
						TotalVolumeM3:         totalVolume,
						TotalCollateral:       totalValue,
						CargoM3:               profile.CargoM3,
						CollateralRate:        profile.CollateralRate,
						FuelPerLY:             fuelPerLY,
						FuelConservationLevel: profile.FuelConservationLevel,
						IsotopePrice:          isotopePrice,
						Waypoints:             jfRoute.Waypoints,
					})
					estimatedCost = result.Cost
				}
			}
		}
	}

	job := &models.TransportJob{
		UserID:               *args.User,
		OriginStationID:      req.OriginStationID,
		DestinationStationID: req.DestinationStationID,
		OriginSystemID:       req.OriginSystemID,
		DestinationSystemID:  req.DestinationSystemID,
		TransportMethod:      req.TransportMethod,
		RoutePreference:      req.RoutePreference,
		TotalVolumeM3:        totalVolume,
		TotalCollateral:      totalValue,
		EstimatedCost:        estimatedCost,
		Jumps:                jumps,
		DistanceLY:           distanceLY,
		JFRouteID:            req.JFRouteID,
		FulfillmentType:      req.FulfillmentType,
		TransportProfileID:   req.TransportProfileID,
		Notes:                req.Notes,
		Items:                items,
	}

	created, err := c.jobsRepo.Create(ctx, job)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to create transport job")}
	}

	// Create corresponding queue entry
	queueEntry, err := c.queueRepo.Create(ctx, &models.IndustryJobQueueEntry{
		UserID:         *args.User,
		Activity:       "transport",
		EstimatedCost:  &estimatedCost,
		TransportJobID: &created.ID,
	})
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to create queue entry for transport job")}
	}

	// Link queue entry back to transport job
	if err := c.jobsRepo.SetQueueEntryID(ctx, created.ID, queueEntry.ID); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to link queue entry to transport job")}
	}
	created.QueueEntryID = &queueEntry.ID

	return created, nil
}

type updateJobStatusRequest struct {
	Status string `json:"status"`
}

func (c *Transportation) UpdateJobStatus(args *web.HandlerArgs) (any, *web.HttpError) {
	id, err := strconv.ParseInt(args.Params["id"], 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid job ID")}
	}

	var req updateJobStatusRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	validStatuses := map[string]bool{
		"in_transit": true,
		"delivered":  true,
		"cancelled":  true,
	}
	if !validStatuses[req.Status] {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid status; must be in_transit, delivered, or cancelled")}
	}

	if err := c.jobsRepo.UpdateStatus(args.Request.Context(), id, *args.User, req.Status); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to update transport job status")}
	}

	return map[string]bool{"success": true}, nil
}

// ---- Route Calculation ----

func (c *Transportation) GetRoute(args *web.HandlerArgs) (any, *web.HttpError) {
	originStr := args.Request.URL.Query().Get("origin")
	destStr := args.Request.URL.Query().Get("destination")
	flag := args.Request.URL.Query().Get("flag")

	if originStr == "" || destStr == "" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("origin and destination query params required")}
	}

	origin, err := strconv.ParseInt(originStr, 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid origin")}
	}
	dest, err := strconv.ParseInt(destStr, 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid destination")}
	}

	if flag == "" {
		flag = "shortest"
	}

	route, err := c.esiClient.GetRoute(args.Request.Context(), origin, dest, flag)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get route")}
	}

	return map[string]any{
		"route": route,
		"jumps": len(route) - 1,
	}, nil
}

// ---- Solar System Search ----

func (c *Transportation) SearchSystems(args *web.HandlerArgs) (any, *web.HttpError) {
	query := args.Request.URL.Query().Get("q")
	if query == "" {
		return []*models.SolarSystem{}, nil
	}

	systems, err := c.solarSysRepo.Search(args.Request.Context(), query, 20)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to search solar systems")}
	}

	return systems, nil
}

// ---- Trigger Config ----

func (c *Transportation) GetTriggerConfig(args *web.HandlerArgs) (any, *web.HttpError) {
	configs, err := c.triggerRepo.GetByUser(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get trigger configs")}
	}
	return configs, nil
}

type upsertTriggerConfigRequest struct {
	TriggerType           string   `json:"triggerType"`
	DefaultFulfillment    string   `json:"defaultFulfillment"`
	AllowedFulfillments   []string `json:"allowedFulfillments"`
	DefaultProfileID      *int64   `json:"defaultProfileId"`
	DefaultMethod         *string  `json:"defaultMethod"`
	CourierRatePerM3      float64  `json:"courierRatePerM3"`
	CourierCollateralRate float64  `json:"courierCollateralRate"`
}

func (c *Transportation) UpsertTriggerConfig(args *web.HandlerArgs) (any, *web.HttpError) {
	var req upsertTriggerConfigRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	if req.TriggerType == "" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("triggerType is required")}
	}

	config := &models.TransportTriggerConfig{
		UserID:                *args.User,
		TriggerType:           req.TriggerType,
		DefaultFulfillment:    req.DefaultFulfillment,
		AllowedFulfillments:   req.AllowedFulfillments,
		DefaultProfileID:      req.DefaultProfileID,
		DefaultMethod:         req.DefaultMethod,
		CourierRatePerM3:      req.CourierRatePerM3,
		CourierCollateralRate: req.CourierCollateralRate,
	}

	if config.DefaultFulfillment == "" {
		config.DefaultFulfillment = "self_haul"
	}
	if config.AllowedFulfillments == nil {
		config.AllowedFulfillments = []string{"self_haul"}
	}

	result, err := c.triggerRepo.Upsert(args.Request.Context(), config)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to upsert trigger config")}
	}

	return result, nil
}
