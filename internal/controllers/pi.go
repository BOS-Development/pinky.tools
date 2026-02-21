package controllers

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type PiPlanetsRepository interface {
	GetPlanetsForUser(ctx context.Context, userID int64) ([]*models.PiPlanet, error)
	GetPinsForPlanets(ctx context.Context, userID int64) ([]*models.PiPin, error)
	GetPinContentsForUser(ctx context.Context, userID int64) ([]*models.PiPinContent, error)
	GetRoutesForUser(ctx context.Context, userID int64) ([]*models.PiRoute, error)
}

type PiTaxConfigRepository interface {
	GetForUser(ctx context.Context, userID int64) ([]*models.PiTaxConfig, error)
	Upsert(ctx context.Context, config *models.PiTaxConfig) error
	Delete(ctx context.Context, userID int64, planetID *int64) error
}

type PiSchematicRepository interface {
	GetAllSchematics(ctx context.Context) ([]*models.SdePlanetSchematic, error)
	GetAllSchematicTypes(ctx context.Context) ([]*models.SdePlanetSchematicType, error)
}

type PiCharacterRepository interface {
	GetNames(ctx context.Context, userID int64) (map[int64]string, error)
}

type PiSolarSystemRepository interface {
	GetNames(ctx context.Context, ids []int64) (map[int64]string, error)
}

type PiItemTypeRepository interface {
	GetNames(ctx context.Context, ids []int64) (map[int64]string, error)
}

type PiMarketPriceRepository interface {
	GetAllJitaPrices(ctx context.Context) (map[int64]*models.MarketPrice, error)
}

type PiLaunchpadLabelRepository interface {
	GetForUser(ctx context.Context, userID int64) ([]*models.PiLaunchpadLabel, error)
	Upsert(ctx context.Context, label *models.PiLaunchpadLabel) error
	Delete(ctx context.Context, userID, characterID, planetID, pinID int64) error
}

type PiStockpileRepository interface {
	GetByUser(ctx context.Context, userID int64) ([]*models.StockpileMarker, error)
}

type Pi struct {
	piRepo        PiPlanetsRepository
	taxRepo       PiTaxConfigRepository
	schematicRepo PiSchematicRepository
	charRepo      PiCharacterRepository
	systemRepo    PiSolarSystemRepository
	itemTypeRepo  PiItemTypeRepository
	marketRepo    PiMarketPriceRepository
	labelRepo     PiLaunchpadLabelRepository
	stockpileRepo PiStockpileRepository
}

func NewPi(
	router Routerer,
	piRepo PiPlanetsRepository,
	taxRepo PiTaxConfigRepository,
	schematicRepo PiSchematicRepository,
	charRepo PiCharacterRepository,
	systemRepo PiSolarSystemRepository,
	itemTypeRepo PiItemTypeRepository,
	marketRepo PiMarketPriceRepository,
	labelRepo PiLaunchpadLabelRepository,
	stockpileRepo PiStockpileRepository,
) *Pi {
	c := &Pi{
		piRepo:        piRepo,
		taxRepo:       taxRepo,
		schematicRepo: schematicRepo,
		charRepo:      charRepo,
		systemRepo:    systemRepo,
		itemTypeRepo:  itemTypeRepo,
		marketRepo:    marketRepo,
		labelRepo:     labelRepo,
		stockpileRepo: stockpileRepo,
	}

	router.RegisterRestAPIRoute("/v1/pi/planets", web.AuthAccessUser, c.GetPlanets, "GET")
	router.RegisterRestAPIRoute("/v1/pi/profit", web.AuthAccessUser, c.GetProfit, "GET")
	router.RegisterRestAPIRoute("/v1/pi/tax", web.AuthAccessUser, c.GetTaxConfig, "GET")
	router.RegisterRestAPIRoute("/v1/pi/tax", web.AuthAccessUser, c.UpsertTaxConfig, "POST")
	router.RegisterRestAPIRoute("/v1/pi/tax", web.AuthAccessUser, c.DeleteTaxConfig, "DELETE")
	router.RegisterRestAPIRoute("/v1/pi/launchpad-labels", web.AuthAccessUser, c.UpsertLaunchpadLabel, "POST")
	router.RegisterRestAPIRoute("/v1/pi/launchpad-labels", web.AuthAccessUser, c.DeleteLaunchpadLabel, "DELETE")
	router.RegisterRestAPIRoute("/v1/pi/launchpad-detail", web.AuthAccessUser, c.GetLaunchpadDetail, "GET")
	router.RegisterRestAPIRoute("/v1/pi/supply-chain", web.AuthAccessUser, c.GetSupplyChain, "GET")

	return c
}

// Stall detection thresholds
const (
	staleDataThreshold  = 48 * time.Hour
	factoryIdleMultiple = 2 // factory is stalled if last_cycle_start + cycle_time*2 < now
)

// Response types

type PiExtractorResponse struct {
	PinID            int64      `json:"pinId"`
	TypeID           int64      `json:"typeId"`
	ProductTypeID    int64      `json:"productTypeId"`
	ProductName      string     `json:"productName"`
	QtyPerCycle      int        `json:"qtyPerCycle"`
	CycleTimeSec     int        `json:"cycleTimeSec"`
	RatePerHour      float64    `json:"ratePerHour"`
	ExpiryTime       *time.Time `json:"expiryTime"`
	Status           string     `json:"status"`
	NumHeads         int        `json:"numHeads"`
}

type PiFactoryResponse struct {
	PinID          int64      `json:"pinId"`
	TypeID         int64      `json:"typeId"`
	SchematicID    int        `json:"schematicId"`
	SchematicName  string     `json:"schematicName"`
	OutputTypeID   int64      `json:"outputTypeId"`
	OutputName     string     `json:"outputName"`
	OutputQty      int        `json:"outputQty"`
	CycleTimeSec   int        `json:"cycleTimeSec"`
	RatePerHour    float64    `json:"ratePerHour"`
	LastCycleStart *time.Time `json:"lastCycleStart"`
	Status         string     `json:"status"`
	PinCategory    string     `json:"pinCategory"`
}

type PiLaunchpadResponse struct {
	PinID    int64                   `json:"pinId"`
	TypeID   int64                   `json:"typeId"`
	Label    string                  `json:"label,omitempty"`
	Contents []*PiPinContentResponse `json:"contents"`
}

type PiPinContentResponse struct {
	TypeID int64  `json:"typeId"`
	Name   string `json:"name"`
	Amount int64  `json:"amount"`
}

type PiPlanetResponse struct {
	PlanetID        int64                  `json:"planetId"`
	PlanetType      string                 `json:"planetType"`
	SolarSystemID   int64                  `json:"solarSystemId"`
	SolarSystemName string                 `json:"solarSystemName"`
	CharacterID     int64                  `json:"characterId"`
	CharacterName   string                 `json:"characterName"`
	UpgradeLevel    int                    `json:"upgradeLevel"`
	NumPins         int                    `json:"numPins"`
	LastUpdate      time.Time              `json:"lastUpdate"`
	Status          string                 `json:"status"`
	Extractors      []*PiExtractorResponse `json:"extractors"`
	Factories       []*PiFactoryResponse   `json:"factories"`
	Launchpads      []*PiLaunchpadResponse `json:"launchpads"`
}

type PiPlanetsListResponse struct {
	Planets []*PiPlanetResponse `json:"planets"`
}

func (c *Pi) GetPlanets(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User

	planets, err := c.piRepo.GetPlanetsForUser(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get PI planets")}
	}

	pins, err := c.piRepo.GetPinsForPlanets(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get PI pins")}
	}

	contents, err := c.piRepo.GetPinContentsForUser(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get PI pin contents")}
	}

	routes, err := c.piRepo.GetRoutesForUser(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get PI routes")}
	}

	labels, err := c.labelRepo.GetForUser(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get launchpad labels")}
	}

	schematics, err := c.schematicRepo.GetAllSchematics(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get schematics")}
	}

	schematicTypes, err := c.schematicRepo.GetAllSchematicTypes(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get schematic types")}
	}

	charNames, err := c.charRepo.GetNames(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get character names")}
	}

	// Collect solar system IDs for name resolution
	systemIDs := []int64{}
	systemIDSet := map[int64]bool{}
	for _, p := range planets {
		if !systemIDSet[p.SolarSystemID] {
			systemIDs = append(systemIDs, p.SolarSystemID)
			systemIDSet[p.SolarSystemID] = true
		}
	}

	systemNames, err := c.systemRepo.GetNames(ctx, systemIDs)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get solar system names")}
	}

	// Build lookup maps
	schematicMap := buildSchematicMap(schematics)
	schematicOutputMap := buildSchematicOutputMap(schematicTypes)

	// Collect all item type IDs for name resolution
	typeIDSet := map[int64]bool{}
	for _, pin := range pins {
		if pin.ExtractorProductTypeID != nil {
			typeIDSet[*pin.ExtractorProductTypeID] = true
		}
	}
	for _, c := range contents {
		typeIDSet[c.TypeID] = true
	}
	for _, r := range routes {
		typeIDSet[r.ContentTypeID] = true
	}
	// Include factory output type IDs from schematics
	for _, output := range schematicOutputMap {
		typeIDSet[output.TypeID] = true
	}
	typeIDs := []int64{}
	for id := range typeIDSet {
		typeIDs = append(typeIDs, id)
	}

	typeNames, err := c.itemTypeRepo.GetNames(ctx, typeIDs)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get item type names")}
	}

	// Group pins by (characterID, planetID)
	type planetKey struct {
		CharacterID int64
		PlanetID    int64
	}
	pinsByPlanet := map[planetKey][]*models.PiPin{}
	for _, pin := range pins {
		key := planetKey{pin.CharacterID, pin.PlanetID}
		pinsByPlanet[key] = append(pinsByPlanet[key], pin)
	}

	// Group contents by (characterID, planetID, pinID)
	type pinKey struct {
		CharacterID int64
		PlanetID    int64
		PinID       int64
	}
	contentsByPin := map[pinKey][]*models.PiPinContent{}
	for _, c := range contents {
		key := pinKey{c.CharacterID, c.PlanetID, c.PinID}
		contentsByPin[key] = append(contentsByPin[key], c)
	}

	// Build label lookup by pinKey
	labelsByPin := map[pinKey]string{}
	for _, l := range labels {
		key := pinKey{l.CharacterID, l.PlanetID, l.PinID}
		labelsByPin[key] = l.Label
	}

	// Group routes by source pin to determine required input types
	routesBySrcPin := map[pinKey][]int64{}
	for _, r := range routes {
		key := pinKey{r.CharacterID, r.PlanetID, r.SourcePinID}
		routesBySrcPin[key] = append(routesBySrcPin[key], r.ContentTypeID)
	}

	now := time.Now()
	result := &PiPlanetsListResponse{
		Planets: []*PiPlanetResponse{},
	}

	for _, planet := range planets {
		key := planetKey{planet.CharacterID, planet.PlanetID}
		planetPins := pinsByPlanet[key]

		extractors := []*PiExtractorResponse{}
		factories := []*PiFactoryResponse{}
		launchpads := []*PiLaunchpadResponse{}
		planetStatus := "running"

		// Check for stale data
		if now.Sub(planet.LastUpdate) > staleDataThreshold {
			planetStatus = "stale_data"
		}

		for _, pin := range planetPins {
			switch {
			case pin.PinCategory == "extractor" && pin.ExtractorProductTypeID != nil:
				ext := buildExtractorResponse(pin, now, typeNames)
				if ext.Status == "expired" && planetStatus == "running" {
					planetStatus = "stalled"
				}
				extractors = append(extractors, ext)

			case pin.PinCategory == "factory" && pin.SchematicID != nil:
				fac := buildFactoryResponse(pin, schematicMap, schematicOutputMap, now, typeNames)
				if fac.Status == "stalled" && planetStatus == "running" {
					planetStatus = "stalled"
				}
				factories = append(factories, fac)

			case pin.PinCategory == "launchpad" || pin.PinCategory == "storage":
				pk := pinKey{pin.CharacterID, pin.PlanetID, pin.PinID}
				pinContents := contentsByPin[pk]
				expectedTypeIDs := routesBySrcPin[pk]
				lp := buildLaunchpadResponse(pin, pinContents, expectedTypeIDs, typeNames)
				lp.Label = labelsByPin[pk]
				launchpads = append(launchpads, lp)
			}
		}

		result.Planets = append(result.Planets, &PiPlanetResponse{
			PlanetID:        planet.PlanetID,
			PlanetType:      planet.PlanetType,
			SolarSystemID:   planet.SolarSystemID,
			SolarSystemName: systemNames[planet.SolarSystemID],
			CharacterID:     planet.CharacterID,
			CharacterName:   charNames[planet.CharacterID],
			UpgradeLevel:    planet.UpgradeLevel,
			NumPins:         planet.NumPins,
			LastUpdate:      planet.LastUpdate,
			Status:          planetStatus,
			Extractors:      extractors,
			Factories:       factories,
			Launchpads:      launchpads,
		})
	}

	return result, nil
}

func buildExtractorResponse(pin *models.PiPin, now time.Time, typeNames map[int64]string) *PiExtractorResponse {
	cycleTime := 0
	if pin.ExtractorCycleTime != nil {
		cycleTime = *pin.ExtractorCycleTime
	}
	qtyPerCycle := 0
	if pin.ExtractorQtyPerCycle != nil {
		qtyPerCycle = *pin.ExtractorQtyPerCycle
	}
	numHeads := 0
	if pin.ExtractorNumHeads != nil {
		numHeads = *pin.ExtractorNumHeads
	}

	ratePerHour := 0.0
	if cycleTime > 0 {
		ratePerHour = float64(qtyPerCycle) / float64(cycleTime) * 3600.0
	}

	status := "running"
	if pin.ExpiryTime != nil && pin.ExpiryTime.Before(now) {
		status = "expired"
	}

	return &PiExtractorResponse{
		PinID:         pin.PinID,
		TypeID:        pin.TypeID,
		ProductTypeID: *pin.ExtractorProductTypeID,
		ProductName:   typeNames[*pin.ExtractorProductTypeID],
		QtyPerCycle:   qtyPerCycle,
		CycleTimeSec:  cycleTime,
		RatePerHour:   ratePerHour,
		ExpiryTime:    pin.ExpiryTime,
		Status:        status,
		NumHeads:      numHeads,
	}
}

type schematicOutput struct {
	TypeID   int64
	Quantity int
}

func buildSchematicMap(schematics []*models.SdePlanetSchematic) map[int]*models.SdePlanetSchematic {
	m := map[int]*models.SdePlanetSchematic{}
	for _, s := range schematics {
		sid := int(s.SchematicID)
		m[sid] = s
	}
	return m
}

func buildSchematicOutputMap(types []*models.SdePlanetSchematicType) map[int]*schematicOutput {
	m := map[int]*schematicOutput{}
	for _, t := range types {
		if !t.IsInput {
			sid := int(t.SchematicID)
			m[sid] = &schematicOutput{
				TypeID:   t.TypeID,
				Quantity: t.Quantity,
			}
		}
	}
	return m
}

func buildFactoryResponse(pin *models.PiPin, schematicMap map[int]*models.SdePlanetSchematic, outputMap map[int]*schematicOutput, now time.Time, typeNames map[int64]string) *PiFactoryResponse {
	schematicID := *pin.SchematicID
	schematic := schematicMap[schematicID]
	output := outputMap[schematicID]

	resp := &PiFactoryResponse{
		PinID:          pin.PinID,
		TypeID:         pin.TypeID,
		SchematicID:    schematicID,
		LastCycleStart: pin.LastCycleStart,
		Status:         "running",
		PinCategory:    pin.PinCategory,
	}

	if schematic != nil {
		resp.SchematicName = schematic.Name
		resp.CycleTimeSec = schematic.CycleTime
		if schematic.CycleTime > 0 && output != nil {
			resp.RatePerHour = float64(output.Quantity) / float64(schematic.CycleTime) * 3600.0
		}
	}

	if output != nil {
		resp.OutputTypeID = output.TypeID
		resp.OutputName = typeNames[output.TypeID]
		resp.OutputQty = output.Quantity
	}

	// Stall detection: factory hasn't cycled in 2x the expected cycle time
	if pin.LastCycleStart != nil && schematic != nil && schematic.CycleTime > 0 {
		expectedNextCycle := pin.LastCycleStart.Add(time.Duration(schematic.CycleTime*factoryIdleMultiple) * time.Second)
		if now.After(expectedNextCycle) {
			resp.Status = "stalled"
		}
	}

	return resp
}

func buildLaunchpadResponse(pin *models.PiPin, contents []*models.PiPinContent, expectedTypeIDs []int64, typeNames map[int64]string) *PiLaunchpadResponse {
	resp := &PiLaunchpadResponse{
		PinID:    pin.PinID,
		TypeID:   pin.TypeID,
		Contents: []*PiPinContentResponse{},
	}

	presentTypes := map[int64]bool{}
	for _, c := range contents {
		presentTypes[c.TypeID] = true
		resp.Contents = append(resp.Contents, &PiPinContentResponse{
			TypeID: c.TypeID,
			Name:   typeNames[c.TypeID],
			Amount: c.Amount,
		})
	}

	// Add expected types from routes that have no actual contents
	seen := map[int64]bool{}
	for _, typeID := range expectedTypeIDs {
		if !presentTypes[typeID] && !seen[typeID] {
			seen[typeID] = true
			resp.Contents = append(resp.Contents, &PiPinContentResponse{
				TypeID: typeID,
				Name:   typeNames[typeID],
				Amount: 0,
			})
		}
	}

	return resp
}

// Tax config handlers

func (c *Pi) GetTaxConfig(args *web.HandlerArgs) (any, *web.HttpError) {
	configs, err := c.taxRepo.GetForUser(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get PI tax config")}
	}
	return configs, nil
}

func (c *Pi) UpsertTaxConfig(args *web.HandlerArgs) (any, *web.HttpError) {
	d := json.NewDecoder(args.Request.Body)
	var config models.PiTaxConfig
	if err := d.Decode(&config); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "failed to decode json")}
	}

	config.UserID = *args.User

	if err := c.taxRepo.Upsert(args.Request.Context(), &config); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to upsert PI tax config")}
	}

	return nil, nil
}

func (c *Pi) DeleteTaxConfig(args *web.HandlerArgs) (any, *web.HttpError) {
	d := json.NewDecoder(args.Request.Body)
	var config models.PiTaxConfig
	if err := d.Decode(&config); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "failed to decode json")}
	}

	if err := c.taxRepo.Delete(args.Request.Context(), *args.User, config.PlanetID); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to delete PI tax config")}
	}

	return nil, nil
}

// --- Phase 3: Launchpad Labels + Detail ---

func (c *Pi) UpsertLaunchpadLabel(args *web.HandlerArgs) (any, *web.HttpError) {
	d := json.NewDecoder(args.Request.Body)
	var label models.PiLaunchpadLabel
	if err := d.Decode(&label); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "failed to decode json")}
	}

	label.UserID = *args.User

	if err := c.labelRepo.Upsert(args.Request.Context(), &label); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to upsert launchpad label")}
	}

	return nil, nil
}

func (c *Pi) DeleteLaunchpadLabel(args *web.HandlerArgs) (any, *web.HttpError) {
	d := json.NewDecoder(args.Request.Body)
	var label models.PiLaunchpadLabel
	if err := d.Decode(&label); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "failed to decode json")}
	}

	if err := c.labelRepo.Delete(args.Request.Context(), *args.User, label.CharacterID, label.PlanetID, label.PinID); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to delete launchpad label")}
	}

	return nil, nil
}

// Launchpad detail response types

type LaunchpadInputRequirement struct {
	TypeID          int64   `json:"typeId"`
	Name            string  `json:"name"`
	QtyPerCycle     int     `json:"qtyPerCycle"`
	CyclesPerHour   float64 `json:"cyclesPerHour"`
	ConsumedPerHour float64 `json:"consumedPerHour"`
	CurrentStock    int64   `json:"currentStock"`
	DepletionHours  float64 `json:"depletionHours"`
}

type LaunchpadConnectedFactory struct {
	PinID         int64                        `json:"pinId"`
	SchematicName string                       `json:"schematicName"`
	OutputName    string                       `json:"outputName"`
	OutputTypeID  int64                        `json:"outputTypeId"`
	CycleTimeSec  int                          `json:"cycleTimeSec"`
	Inputs        []*LaunchpadInputRequirement `json:"inputs"`
}

type LaunchpadDetailResponse struct {
	PinID       int64                        `json:"pinId"`
	CharacterID int64                        `json:"characterId"`
	PlanetID    int64                        `json:"planetId"`
	Label       string                       `json:"label,omitempty"`
	Contents    []*PiPinContentResponse      `json:"contents"`
	Factories   []*LaunchpadConnectedFactory `json:"factories"`
}

func (c *Pi) GetLaunchpadDetail(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User
	q := args.Request.URL.Query()

	characterID, err := strconv.ParseInt(q.Get("characterId"), 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid characterId")}
	}
	planetID, err := strconv.ParseInt(q.Get("planetId"), 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid planetId")}
	}
	pinID, err := strconv.ParseInt(q.Get("pinId"), 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid pinId")}
	}

	// Fetch all data for this user's planets
	pins, err := c.piRepo.GetPinsForPlanets(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get PI pins")}
	}

	contents, err := c.piRepo.GetPinContentsForUser(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get PI pin contents")}
	}

	routes, err := c.piRepo.GetRoutesForUser(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get PI routes")}
	}

	labels, err := c.labelRepo.GetForUser(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get launchpad labels")}
	}

	schematics, err := c.schematicRepo.GetAllSchematics(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get schematics")}
	}

	schematicTypes, err := c.schematicRepo.GetAllSchematicTypes(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get schematic types")}
	}

	schematicMap := buildSchematicMap(schematics)
	schematicOutputMap := buildSchematicOutputMap(schematicTypes)
	schematicInputMap := buildSchematicInputMap(schematicTypes)

	// Collect type IDs for name resolution
	typeIDSet := map[int64]bool{}
	for _, c := range contents {
		if c.CharacterID == characterID && c.PlanetID == planetID && c.PinID == pinID {
			typeIDSet[c.TypeID] = true
		}
	}
	for _, output := range schematicOutputMap {
		typeIDSet[output.TypeID] = true
	}
	for _, inputs := range schematicInputMap {
		for _, inp := range inputs {
			typeIDSet[inp.TypeID] = true
		}
	}
	typeIDs := []int64{}
	for id := range typeIDSet {
		typeIDs = append(typeIDs, id)
	}

	typeNames, err := c.itemTypeRepo.GetNames(ctx, typeIDs)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get item type names")}
	}

	// Build pin lookup for this planet
	pinsByID := map[int64]*models.PiPin{}
	for _, pin := range pins {
		if pin.CharacterID == characterID && pin.PlanetID == planetID {
			pinsByID[pin.PinID] = pin
		}
	}

	// Build contents lookup for the launchpad
	launchpadContents := []*PiPinContentResponse{}
	stockByType := map[int64]int64{}
	for _, ct := range contents {
		if ct.CharacterID == characterID && ct.PlanetID == planetID && ct.PinID == pinID {
			stockByType[ct.TypeID] = ct.Amount
			launchpadContents = append(launchpadContents, &PiPinContentResponse{
				TypeID: ct.TypeID,
				Name:   typeNames[ct.TypeID],
				Amount: ct.Amount,
			})
		}
	}

	// Find routes FROM this launchpad TO factories (source = this launchpad).
	// Track which type is routed to which factory so we only show types
	// the launchpad actually feeds out.
	type outboundRoute struct {
		factoryPinID  int64
		contentTypeID int64
	}
	outboundRoutes := []outboundRoute{}
	for _, r := range routes {
		if r.CharacterID == characterID && r.PlanetID == planetID && r.SourcePinID == pinID {
			outboundRoutes = append(outboundRoutes, outboundRoute{
				factoryPinID:  r.DestinationPinID,
				contentTypeID: r.ContentTypeID,
			})
		}
	}

	// Group outbound routes by factory, collecting which types are routed to each
	factoryRoutedTypes := map[int64]map[int64]bool{}
	for _, ob := range outboundRoutes {
		if factoryRoutedTypes[ob.factoryPinID] == nil {
			factoryRoutedTypes[ob.factoryPinID] = map[int64]bool{}
		}
		factoryRoutedTypes[ob.factoryPinID][ob.contentTypeID] = true
	}

	// Build factory detail — only include inputs that are routed from this launchpad
	factoryResps := []*LaunchpadConnectedFactory{}
	for factoryPinID, routedTypes := range factoryRoutedTypes {
		pin := pinsByID[factoryPinID]
		if pin == nil || pin.PinCategory != "factory" || pin.SchematicID == nil {
			continue
		}

		schematicID := *pin.SchematicID
		schematic := schematicMap[schematicID]
		output := schematicOutputMap[schematicID]
		inputs := schematicInputMap[schematicID]

		if schematic == nil || output == nil || schematic.CycleTime <= 0 {
			continue
		}

		cyclesPerHour := 3600.0 / float64(schematic.CycleTime)

		inputResps := []*LaunchpadInputRequirement{}
		for _, inp := range inputs {
			// Only include types that are actually routed from this launchpad
			if !routedTypes[inp.TypeID] {
				continue
			}
			consumedPerHour := float64(inp.Quantity) * cyclesPerHour
			stock := stockByType[inp.TypeID]
			depletionHours := 0.0
			if consumedPerHour > 0 && stock > 0 {
				depletionHours = float64(stock) / consumedPerHour
			}

			inputResps = append(inputResps, &LaunchpadInputRequirement{
				TypeID:          inp.TypeID,
				Name:            typeNames[inp.TypeID],
				QtyPerCycle:     inp.Quantity,
				CyclesPerHour:   cyclesPerHour,
				ConsumedPerHour: consumedPerHour,
				CurrentStock:    stock,
				DepletionHours:  depletionHours,
			})
		}

		factoryResps = append(factoryResps, &LaunchpadConnectedFactory{
			PinID:         pin.PinID,
			SchematicName: schematic.Name,
			OutputName:    typeNames[output.TypeID],
			OutputTypeID:  output.TypeID,
			CycleTimeSec:  schematic.CycleTime,
			Inputs:        inputResps,
		})
	}

	// Find label for this launchpad
	labelStr := ""
	for _, l := range labels {
		if l.CharacterID == characterID && l.PlanetID == planetID && l.PinID == pinID {
			labelStr = l.Label
			break
		}
	}

	return &LaunchpadDetailResponse{
		PinID:       pinID,
		CharacterID: characterID,
		PlanetID:    planetID,
		Label:       labelStr,
		Contents:    launchpadContents,
		Factories:   factoryResps,
	}, nil
}

// --- Phase 2: Profit Calculation ---

// PI tier base costs (ISK per unit) used for POCO tax calculation
var piTierBaseCost = map[int]float64{
	0: 5,         // R0
	1: 400,       // P1
	2: 7200,      // P2
	3: 60000,     // P3
	4: 1200000,   // P4
}

var tierName = map[int]string{
	0: "R0",
	1: "P1",
	2: "P2",
	3: "P3",
	4: "P4",
}

// Profit response types

type PiFactoryInputResponse struct {
	TypeID           int64   `json:"typeId"`
	Name             string  `json:"name"`
	Tier             string  `json:"tier"`
	Quantity         int     `json:"quantity"`
	PricePerUnit     float64 `json:"pricePerUnit"`
	CostPerHour      float64 `json:"costPerHour"`
	ImportTaxPerHour float64 `json:"importTaxPerHour"`
	IsLocal          bool    `json:"isLocal"`
}

type PiFactoryProfitResponse struct {
	PinID              int64                      `json:"pinId"`
	SchematicID        int                        `json:"schematicId"`
	SchematicName      string                     `json:"schematicName"`
	OutputTypeID       int64                      `json:"outputTypeId"`
	OutputName         string                     `json:"outputName"`
	OutputTier         string                     `json:"outputTier"`
	OutputQty          int                        `json:"outputQty"`
	CycleTimeSec       int                        `json:"cycleTimeSec"`
	RatePerHour        float64                    `json:"ratePerHour"`
	OutputValuePerHour float64                    `json:"outputValuePerHour"`
	InputCostPerHour   float64                    `json:"inputCostPerHour"`
	ExportTaxPerHour   float64                    `json:"exportTaxPerHour"`
	ImportTaxPerHour   float64                    `json:"importTaxPerHour"`
	ProfitPerHour      float64                    `json:"profitPerHour"`
	Inputs             []*PiFactoryInputResponse  `json:"inputs"`
}

type PiPlanetProfitResponse struct {
	PlanetID           int64                       `json:"planetId"`
	PlanetType         string                      `json:"planetType"`
	SolarSystemID      int64                       `json:"solarSystemId"`
	SolarSystemName    string                      `json:"solarSystemName"`
	CharacterID        int64                       `json:"characterId"`
	CharacterName      string                      `json:"characterName"`
	TaxRate            float64                     `json:"taxRate"`
	TotalOutputValue   float64                     `json:"totalOutputValue"`
	TotalInputCost     float64                     `json:"totalInputCost"`
	TotalExportTax     float64                     `json:"totalExportTax"`
	TotalImportTax     float64                     `json:"totalImportTax"`
	NetProfitPerHour   float64                     `json:"netProfitPerHour"`
	Factories          []*PiFactoryProfitResponse  `json:"factories"`
}

type PiProfitListResponse struct {
	Planets          []*PiPlanetProfitResponse `json:"planets"`
	PriceSource      string                    `json:"priceSource"`
	TotalOutputValue float64                   `json:"totalOutputValue"`
	TotalInputCost   float64                   `json:"totalInputCost"`
	TotalExportTax   float64                   `json:"totalExportTax"`
	TotalImportTax   float64                   `json:"totalImportTax"`
	TotalProfit      float64                   `json:"totalProfit"`
}

// buildTierMap classifies every PI material type into a tier (0=R0, 1=P1, ..., 4=P4)
// by walking the schematic type graph. A type that is never the output of any schematic
// is R0 (raw resource). Outputs inherit tier = max(input tiers) + 1.
func buildTierMap(schematicTypes []*models.SdePlanetSchematicType) map[int64]int {
	outputToSchematic := map[int64]int64{} // typeID -> schematicID that produces it
	schematicInputTypeIDs := map[int64][]int64{} // schematicID -> input typeIDs

	for _, st := range schematicTypes {
		if st.IsInput {
			schematicInputTypeIDs[st.SchematicID] = append(schematicInputTypeIDs[st.SchematicID], st.TypeID)
		} else {
			outputToSchematic[st.TypeID] = st.SchematicID
		}
	}

	// Collect all type IDs that appear as inputs
	allInputIDs := map[int64]bool{}
	for _, inputs := range schematicInputTypeIDs {
		for _, id := range inputs {
			allInputIDs[id] = true
		}
	}

	tierMap := map[int64]int{}

	// R0: any type that is used as input but is never the output of a schematic
	for typeID := range allInputIDs {
		if _, isOutput := outputToSchematic[typeID]; !isOutput {
			tierMap[typeID] = 0
		}
	}

	// Iteratively classify outputs based on their inputs' tiers
	changed := true
	for changed {
		changed = false
		for outputTypeID, schematicID := range outputToSchematic {
			if _, done := tierMap[outputTypeID]; done {
				continue
			}

			inputs := schematicInputTypeIDs[schematicID]
			allClassified := true
			maxInputTier := 0
			for _, inputID := range inputs {
				t, ok := tierMap[inputID]
				if !ok {
					allClassified = false
					break
				}
				if t > maxInputTier {
					maxInputTier = t
				}
			}

			if allClassified && len(inputs) > 0 {
				tierMap[outputTypeID] = maxInputTier + 1
				changed = true
			}
		}
	}

	return tierMap
}

// schematicInput describes one input material for a schematic
type schematicInput struct {
	TypeID   int64
	Quantity int
}

// buildSchematicInputMap returns schematicID -> list of inputs
func buildSchematicInputMap(types []*models.SdePlanetSchematicType) map[int][]*schematicInput {
	m := map[int][]*schematicInput{}
	for _, t := range types {
		if t.IsInput {
			sid := int(t.SchematicID)
			m[sid] = append(m[sid], &schematicInput{
				TypeID:   t.TypeID,
				Quantity: t.Quantity,
			})
		}
	}
	return m
}

func getPrice(mp *models.MarketPrice, source string) float64 {
	if mp == nil {
		return 0
	}
	switch source {
	case "buy":
		if mp.BuyPrice != nil {
			return *mp.BuyPrice
		}
	case "sell":
		if mp.SellPrice != nil {
			return *mp.SellPrice
		}
	case "split":
		buy, sell := 0.0, 0.0
		if mp.BuyPrice != nil {
			buy = *mp.BuyPrice
		}
		if mp.SellPrice != nil {
			sell = *mp.SellPrice
		}
		return (buy + sell) / 2
	}
	return 0
}

func (c *Pi) GetProfit(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User

	priceSource := args.Request.URL.Query().Get("priceSource")
	if priceSource == "" {
		priceSource = "sell"
	}

	planets, err := c.piRepo.GetPlanetsForUser(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get PI planets")}
	}

	pins, err := c.piRepo.GetPinsForPlanets(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get PI pins")}
	}

	schematics, err := c.schematicRepo.GetAllSchematics(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get schematics")}
	}

	schematicTypes, err := c.schematicRepo.GetAllSchematicTypes(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get schematic types")}
	}

	jitaPrices, err := c.marketRepo.GetAllJitaPrices(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get Jita prices")}
	}

	taxConfigs, err := c.taxRepo.GetForUser(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get tax configs")}
	}

	charNames, err := c.charRepo.GetNames(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get character names")}
	}

	systemIDs := []int64{}
	systemIDSet := map[int64]bool{}
	for _, p := range planets {
		if !systemIDSet[p.SolarSystemID] {
			systemIDs = append(systemIDs, p.SolarSystemID)
			systemIDSet[p.SolarSystemID] = true
		}
	}
	systemNames, err := c.systemRepo.GetNames(ctx, systemIDs)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get system names")}
	}

	// Collect type IDs for name resolution
	typeIDSet := map[int64]bool{}
	for _, st := range schematicTypes {
		typeIDSet[st.TypeID] = true
	}
	typeIDs := []int64{}
	for id := range typeIDSet {
		typeIDs = append(typeIDs, id)
	}
	typeNames, err := c.itemTypeRepo.GetNames(ctx, typeIDs)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get item type names")}
	}

	// Build lookup maps
	schematicMap := buildSchematicMap(schematics)
	schematicOutputMap := buildSchematicOutputMap(schematicTypes)
	schematicInputMap := buildSchematicInputMap(schematicTypes)
	tierMap := buildTierMap(schematicTypes)

	// Build tax rate lookup: planetID -> rate, with global default fallback
	globalTaxRate := 2.5
	planetTaxRates := map[int64]float64{}
	for _, tc := range taxConfigs {
		if tc.PlanetID == nil {
			globalTaxRate = tc.TaxRate
		} else {
			planetTaxRates[*tc.PlanetID] = tc.TaxRate
		}
	}

	// Group pins by planet
	type planetKey struct {
		CharacterID int64
		PlanetID    int64
	}
	pinsByPlanet := map[planetKey][]*models.PiPin{}
	for _, pin := range pins {
		key := planetKey{pin.CharacterID, pin.PlanetID}
		pinsByPlanet[key] = append(pinsByPlanet[key], pin)
	}

	// Build global set of all types produced across ALL user planets.
	// This lets us recognize cross-planet inputs (cost=0 but taxes still apply).
	userProducedTypes := map[int64]bool{}
	for _, pin := range pins {
		if pin.PinCategory == "extractor" && pin.ExtractorProductTypeID != nil {
			userProducedTypes[*pin.ExtractorProductTypeID] = true
		}
		if pin.PinCategory == "factory" && pin.SchematicID != nil {
			if output := schematicOutputMap[*pin.SchematicID]; output != nil {
				userProducedTypes[output.TypeID] = true
			}
		}
	}

	result := &PiProfitListResponse{
		Planets:     []*PiPlanetProfitResponse{},
		PriceSource: priceSource,
	}

	for _, planet := range planets {
		taxRate := globalTaxRate
		if rate, ok := planetTaxRates[planet.PlanetID]; ok {
			taxRate = rate
		}

		key := planetKey{planet.CharacterID, planet.PlanetID}
		planetPins := pinsByPlanet[key]

		// First pass: collect all types produced on this planet (extractors + factories)
		// so we know which factory inputs are locally sourced vs imported.
		locallyProduced := map[int64]bool{}
		// Track per-type production and consumption rates for net planet totals.
		productionRates := map[int64]float64{} // typeID -> units/hr produced
		consumptionRates := map[int64]float64{} // typeID -> units/hr consumed

		for _, pin := range planetPins {
			if pin.PinCategory == "extractor" && pin.ExtractorProductTypeID != nil {
				locallyProduced[*pin.ExtractorProductTypeID] = true
				cycleTime := 0
				if pin.ExtractorCycleTime != nil {
					cycleTime = *pin.ExtractorCycleTime
				}
				qtyPerCycle := 0
				if pin.ExtractorQtyPerCycle != nil {
					qtyPerCycle = *pin.ExtractorQtyPerCycle
				}
				if cycleTime > 0 {
					productionRates[*pin.ExtractorProductTypeID] += float64(qtyPerCycle) / float64(cycleTime) * 3600.0
				}
			}
			if pin.PinCategory == "factory" && pin.SchematicID != nil {
				output := schematicOutputMap[*pin.SchematicID]
				schematic := schematicMap[*pin.SchematicID]
				if output != nil && schematic != nil && schematic.CycleTime > 0 {
					locallyProduced[output.TypeID] = true
					cyclesPerHour := 3600.0 / float64(schematic.CycleTime)
					productionRates[output.TypeID] += float64(output.Quantity) * cyclesPerHour

					inputs := schematicInputMap[*pin.SchematicID]
					for _, inp := range inputs {
						consumptionRates[inp.TypeID] += float64(inp.Quantity) * cyclesPerHour
					}
				}
			}
		}

		planetResp := &PiPlanetProfitResponse{
			PlanetID:        planet.PlanetID,
			PlanetType:      planet.PlanetType,
			SolarSystemID:   planet.SolarSystemID,
			SolarSystemName: systemNames[planet.SolarSystemID],
			CharacterID:     planet.CharacterID,
			CharacterName:   charNames[planet.CharacterID],
			TaxRate:         taxRate,
			Factories:       []*PiFactoryProfitResponse{},
		}

		// Second pass: build per-factory profit (inputs sourced locally have zero cost)
		for _, pin := range planetPins {
			if pin.PinCategory != "factory" || pin.SchematicID == nil {
				continue
			}

			schematicID := *pin.SchematicID
			schematic := schematicMap[schematicID]
			output := schematicOutputMap[schematicID]
			inputs := schematicInputMap[schematicID]

			if schematic == nil || output == nil || schematic.CycleTime <= 0 {
				continue
			}

			cyclesPerHour := 3600.0 / float64(schematic.CycleTime)

			outputPrice := getPrice(jitaPrices[output.TypeID], priceSource)
			outputValuePerHour := float64(output.Quantity) * outputPrice * cyclesPerHour

			outputTier := tierMap[output.TypeID]
			exportTaxPerHour := piTierBaseCost[outputTier] * float64(output.Quantity) * (taxRate / 100) * cyclesPerHour

			inputCostPerHour := 0.0
			importTaxPerHour := 0.0
			inputResps := []*PiFactoryInputResponse{}

			for _, inp := range inputs {
				inputTier := tierMap[inp.TypeID]
				samePlanet := locallyProduced[inp.TypeID]
				crossPlanet := !samePlanet && userProducedTypes[inp.TypeID]
				isLocal := samePlanet || crossPlanet

				costPerHour := 0.0
				impTaxPerHour := 0.0
				inputPrice := getPrice(jitaPrices[inp.TypeID], priceSource)

				if samePlanet {
					// Produced on same planet: no material cost, no import tax (no transfer)
				} else if crossPlanet {
					// Produced on another user planet: no material cost, but import tax applies (transfer)
					impTaxPerHour = piTierBaseCost[inputTier] * float64(inp.Quantity) * (taxRate / 100) * 0.5 * cyclesPerHour
				} else {
					// Truly external: Jita price + import tax
					costPerHour = float64(inp.Quantity) * inputPrice * cyclesPerHour
					impTaxPerHour = piTierBaseCost[inputTier] * float64(inp.Quantity) * (taxRate / 100) * 0.5 * cyclesPerHour
				}

				inputCostPerHour += costPerHour
				importTaxPerHour += impTaxPerHour

				inputResps = append(inputResps, &PiFactoryInputResponse{
					TypeID:           inp.TypeID,
					Name:             typeNames[inp.TypeID],
					Tier:             tierName[inputTier],
					Quantity:         inp.Quantity,
					PricePerUnit:     inputPrice,
					CostPerHour:      costPerHour,
					ImportTaxPerHour: impTaxPerHour,
					IsLocal:          isLocal,
				})
			}

			profitPerHour := outputValuePerHour - inputCostPerHour - exportTaxPerHour - importTaxPerHour

			planetResp.Factories = append(planetResp.Factories, &PiFactoryProfitResponse{
				PinID:              pin.PinID,
				SchematicID:        schematicID,
				SchematicName:      schematic.Name,
				OutputTypeID:       output.TypeID,
				OutputName:         typeNames[output.TypeID],
				OutputTier:         tierName[outputTier],
				OutputQty:          output.Quantity,
				CycleTimeSec:       schematic.CycleTime,
				RatePerHour:        float64(output.Quantity) * cyclesPerHour,
				OutputValuePerHour: outputValuePerHour,
				InputCostPerHour:   inputCostPerHour,
				ExportTaxPerHour:   exportTaxPerHour,
				ImportTaxPerHour:   importTaxPerHour,
				ProfitPerHour:      profitPerHour,
				Inputs:             inputResps,
			})
		}

		// Planet totals: use net production/consumption to avoid double-counting intermediates.
		// Net exported types = production > consumption → revenue
		// Net imported types = consumption > production → cost
		allTypeIDs := map[int64]bool{}
		for id := range productionRates {
			allTypeIDs[id] = true
		}
		for id := range consumptionRates {
			allTypeIDs[id] = true
		}

		for typeID := range allTypeIDs {
			produced := productionRates[typeID]
			consumed := consumptionRates[typeID]
			net := produced - consumed
			t := tierMap[typeID]

			if net > 0 {
				// Net export
				price := getPrice(jitaPrices[typeID], priceSource)
				planetResp.TotalOutputValue += net * price
				planetResp.TotalExportTax += net * piTierBaseCost[t] * (taxRate / 100)
			} else if net < 0 {
				// Net import
				imported := -net
				planetResp.TotalImportTax += imported * piTierBaseCost[t] * (taxRate / 100) * 0.5
				if !userProducedTypes[typeID] {
					// Only charge Jita price for materials not produced on any user planet
					price := getPrice(jitaPrices[typeID], priceSource)
					planetResp.TotalInputCost += imported * price
				}
			}
		}
		planetResp.NetProfitPerHour = planetResp.TotalOutputValue - planetResp.TotalInputCost - planetResp.TotalExportTax - planetResp.TotalImportTax

		result.Planets = append(result.Planets, planetResp)
	}

	// Compute global totals across ALL planets.
	// Only types that are net-exported across all planets count as revenue (sold at Jita).
	// Only types that are net-imported across all planets count as costs (bought from Jita).
	// Taxes are summed from per-factory calculations (actual POCO payments).
	globalProduction := map[int64]float64{}
	globalConsumption := map[int64]float64{}
	for _, pin := range pins {
		if pin.PinCategory == "extractor" && pin.ExtractorProductTypeID != nil {
			cycleTime := 0
			if pin.ExtractorCycleTime != nil {
				cycleTime = *pin.ExtractorCycleTime
			}
			qtyPerCycle := 0
			if pin.ExtractorQtyPerCycle != nil {
				qtyPerCycle = *pin.ExtractorQtyPerCycle
			}
			if cycleTime > 0 {
				globalProduction[*pin.ExtractorProductTypeID] += float64(qtyPerCycle) / float64(cycleTime) * 3600.0
			}
		}
		if pin.PinCategory == "factory" && pin.SchematicID != nil {
			output := schematicOutputMap[*pin.SchematicID]
			schematic := schematicMap[*pin.SchematicID]
			if output != nil && schematic != nil && schematic.CycleTime > 0 {
				cyclesPerHour := 3600.0 / float64(schematic.CycleTime)
				globalProduction[output.TypeID] += float64(output.Quantity) * cyclesPerHour
				for _, inp := range schematicInputMap[*pin.SchematicID] {
					globalConsumption[inp.TypeID] += float64(inp.Quantity) * cyclesPerHour
				}
			}
		}
	}

	globalTypeIDs := map[int64]bool{}
	for id := range globalProduction {
		globalTypeIDs[id] = true
	}
	for id := range globalConsumption {
		globalTypeIDs[id] = true
	}

	for typeID := range globalTypeIDs {
		produced := globalProduction[typeID]
		consumed := globalConsumption[typeID]
		net := produced - consumed
		t := tierMap[typeID]

		if net > 0 {
			// Truly exported to market
			price := getPrice(jitaPrices[typeID], priceSource)
			result.TotalOutputValue += net * price
			result.TotalExportTax += net * piTierBaseCost[t] * (globalTaxRate / 100)
		} else if net < 0 {
			// Truly imported from market
			imported := -net
			price := getPrice(jitaPrices[typeID], priceSource)
			result.TotalInputCost += imported * price
			result.TotalImportTax += imported * piTierBaseCost[t] * (globalTaxRate / 100) * 0.5
		}
	}

	// For total taxes, sum actual per-factory taxes instead of the global net estimate,
	// since each planet has its own tax rate and inter-planet transfers incur taxes.
	result.TotalExportTax = 0
	result.TotalImportTax = 0
	for _, planet := range result.Planets {
		for _, factory := range planet.Factories {
			result.TotalExportTax += factory.ExportTaxPerHour
			result.TotalImportTax += factory.ImportTaxPerHour
		}
	}
	result.TotalProfit = result.TotalOutputValue - result.TotalInputCost - result.TotalExportTax - result.TotalImportTax

	return result, nil
}

// --- Phase 5: Supply Chain Analysis ---

type SupplyChainItem struct {
	TypeID            int64   `json:"typeId"`
	Name              string  `json:"name"`
	Tier              int     `json:"tier"`
	TierName          string  `json:"tierName"`
	ProducedPerHour   float64 `json:"producedPerHour"`
	ConsumedPerHour   float64 `json:"consumedPerHour"`
	NetPerHour        float64 `json:"netPerHour"`
	StockpileQty      int64   `json:"stockpileQty"`
	DepletionHours    float64 `json:"depletionHours"`  // 0 = no consumption or no stock
	Source            string  `json:"source"`           // "extracted", "produced", "bought", "mixed"
	Producers         []*SupplyChainPlanetEntry `json:"producers"`
	Consumers         []*SupplyChainPlanetEntry `json:"consumers"`
	StockpileMarkers  []*models.StockpileMarker `json:"stockpileMarkers,omitempty"`
}

type SupplyChainPlanetEntry struct {
	CharacterID     int64   `json:"characterId"`
	CharacterName   string  `json:"characterName"`
	PlanetID        int64   `json:"planetId"`
	SolarSystemName string  `json:"solarSystemName"`
	PlanetType      string  `json:"planetType"`
	RatePerHour     float64 `json:"ratePerHour"`
}

type SupplyChainResponse struct {
	Items []*SupplyChainItem `json:"items"`
}

// rollUpByPlanet aggregates multiple pin-level entries into one entry per planet,
// summing their rates.
func rollUpByPlanet(entries []*SupplyChainPlanetEntry) []*SupplyChainPlanetEntry {
	if len(entries) <= 1 {
		return entries
	}

	type key struct {
		characterID int64
		planetID    int64
	}
	order := []key{}
	grouped := map[key]*SupplyChainPlanetEntry{}
	for _, e := range entries {
		k := key{e.CharacterID, e.PlanetID}
		if existing, ok := grouped[k]; ok {
			existing.RatePerHour += e.RatePerHour
		} else {
			grouped[k] = &SupplyChainPlanetEntry{
				CharacterID:     e.CharacterID,
				CharacterName:   e.CharacterName,
				PlanetID:        e.PlanetID,
				SolarSystemName: e.SolarSystemName,
				PlanetType:      e.PlanetType,
				RatePerHour:     e.RatePerHour,
			}
			order = append(order, k)
		}
	}

	result := make([]*SupplyChainPlanetEntry, 0, len(order))
	for _, k := range order {
		result = append(result, grouped[k])
	}
	return result
}

func (c *Pi) GetSupplyChain(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User

	planets, err := c.piRepo.GetPlanetsForUser(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get PI planets")}
	}

	pins, err := c.piRepo.GetPinsForPlanets(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get PI pins")}
	}

	schematics, err := c.schematicRepo.GetAllSchematics(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get schematics")}
	}

	schematicTypes, err := c.schematicRepo.GetAllSchematicTypes(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get schematic types")}
	}

	stockpileMarkers, err := c.stockpileRepo.GetByUser(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get stockpile markers")}
	}

	charNames, err := c.charRepo.GetNames(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get character names")}
	}

	systemIDs := []int64{}
	systemIDSet := map[int64]bool{}
	for _, p := range planets {
		if !systemIDSet[p.SolarSystemID] {
			systemIDs = append(systemIDs, p.SolarSystemID)
			systemIDSet[p.SolarSystemID] = true
		}
	}
	systemNames, err := c.systemRepo.GetNames(ctx, systemIDs)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get system names")}
	}

	schematicMap := buildSchematicMap(schematics)
	schematicOutputMap := buildSchematicOutputMap(schematicTypes)
	schematicInputMap := buildSchematicInputMap(schematicTypes)
	tierMap := buildTierMap(schematicTypes)

	// Build planet lookup for producer/consumer entries
	planetByKey := map[int64]*models.PiPlanet{} // characterID<<32 | planetID -> planet
	for _, p := range planets {
		planetByKey[p.CharacterID<<32|p.PlanetID] = p
	}

	// Aggregate production and consumption across all planets
	type typeEntry struct {
		producedPerHour float64
		consumedPerHour float64
		producers       []*SupplyChainPlanetEntry
		consumers       []*SupplyChainPlanetEntry
		isExtracted     bool
		isProduced      bool
	}
	entries := map[int64]*typeEntry{}

	getEntry := func(typeID int64) *typeEntry {
		e := entries[typeID]
		if e == nil {
			e = &typeEntry{}
			entries[typeID] = e
		}
		return e
	}

	makePlanetEntry := func(pin *models.PiPin, rate float64) *SupplyChainPlanetEntry {
		planet := planetByKey[pin.CharacterID<<32|pin.PlanetID]
		if planet == nil {
			return &SupplyChainPlanetEntry{
				CharacterID: pin.CharacterID,
				PlanetID:    pin.PlanetID,
				RatePerHour: rate,
			}
		}
		return &SupplyChainPlanetEntry{
			CharacterID:     pin.CharacterID,
			CharacterName:   charNames[pin.CharacterID],
			PlanetID:        pin.PlanetID,
			SolarSystemName: systemNames[planet.SolarSystemID],
			PlanetType:      planet.PlanetType,
			RatePerHour:     rate,
		}
	}

	for _, pin := range pins {
		if pin.PinCategory == "extractor" && pin.ExtractorProductTypeID != nil {
			cycleTime := 0
			if pin.ExtractorCycleTime != nil {
				cycleTime = *pin.ExtractorCycleTime
			}
			qtyPerCycle := 0
			if pin.ExtractorQtyPerCycle != nil {
				qtyPerCycle = *pin.ExtractorQtyPerCycle
			}
			if cycleTime > 0 {
				rate := float64(qtyPerCycle) / float64(cycleTime) * 3600.0
				e := getEntry(*pin.ExtractorProductTypeID)
				e.producedPerHour += rate
				e.isExtracted = true
				e.producers = append(e.producers, makePlanetEntry(pin, rate))
			}
		}

		if pin.PinCategory == "factory" && pin.SchematicID != nil {
			output := schematicOutputMap[*pin.SchematicID]
			schematic := schematicMap[*pin.SchematicID]
			if output != nil && schematic != nil && schematic.CycleTime > 0 {
				cyclesPerHour := 3600.0 / float64(schematic.CycleTime)
				outputRate := float64(output.Quantity) * cyclesPerHour
				e := getEntry(output.TypeID)
				e.producedPerHour += outputRate
				e.isProduced = true
				e.producers = append(e.producers, makePlanetEntry(pin, outputRate))

				inputs := schematicInputMap[*pin.SchematicID]
				for _, inp := range inputs {
					inputRate := float64(inp.Quantity) * cyclesPerHour
					ce := getEntry(inp.TypeID)
					ce.consumedPerHour += inputRate
					ce.consumers = append(ce.consumers, makePlanetEntry(pin, inputRate))
				}
			}
		}
	}

	// Build set of PI material type IDs from schematic data
	piTypeIDs := map[int64]bool{}
	for _, st := range schematicTypes {
		piTypeIDs[st.TypeID] = true
	}

	// Aggregate stockpile quantities by typeID for PI materials
	stockpileByType := map[int64]int64{}
	markersByType := map[int64][]*models.StockpileMarker{}
	for _, m := range stockpileMarkers {
		if _, isPiType := entries[m.TypeID]; isPiType {
			stockpileByType[m.TypeID] += m.DesiredQuantity
			markersByType[m.TypeID] = append(markersByType[m.TypeID], m)
		} else if piTypeIDs[m.TypeID] {
			// Stockpile-only type not produced/consumed — add as "bought" entry
			stockpileByType[m.TypeID] += m.DesiredQuantity
			markersByType[m.TypeID] = append(markersByType[m.TypeID], m)
			entries[m.TypeID] = &typeEntry{}
		}
	}

	// Collect type IDs for name resolution
	typeIDSet := map[int64]bool{}
	for typeID := range entries {
		typeIDSet[typeID] = true
	}
	typeIDs := []int64{}
	for id := range typeIDSet {
		typeIDs = append(typeIDs, id)
	}
	typeNames, err := c.itemTypeRepo.GetNames(ctx, typeIDs)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get item type names")}
	}

	// Build response items
	items := []*SupplyChainItem{}
	for typeID, e := range entries {
		net := e.producedPerHour - e.consumedPerHour
		stockQty := stockpileByType[typeID]

		depletionHours := 0.0
		if e.consumedPerHour > 0 && net < 0 {
			// Net deficit — how long until stockpile runs out
			deficit := -net
			if stockQty > 0 {
				depletionHours = float64(stockQty) / deficit
			}
		}

		source := "produced"
		if e.isExtracted && !e.isProduced {
			source = "extracted"
		} else if e.isExtracted && e.isProduced {
			source = "mixed"
		} else if !e.isExtracted && !e.isProduced && stockQty > 0 {
			source = "bought"
		}

		tier := tierMap[typeID]
		tn := "R0"
		if t, ok := tierName[tier]; ok {
			tn = t
		}

		items = append(items, &SupplyChainItem{
			TypeID:          typeID,
			Name:            typeNames[typeID],
			Tier:            tier,
			TierName:        tn,
			ProducedPerHour: e.producedPerHour,
			ConsumedPerHour: e.consumedPerHour,
			NetPerHour:      net,
			StockpileQty:    stockQty,
			DepletionHours:  depletionHours,
			Source:          source,
			Producers:        rollUpByPlanet(e.producers),
			Consumers:        rollUpByPlanet(e.consumers),
			StockpileMarkers: markersByType[typeID],
		})
	}

	return &SupplyChainResponse{Items: items}, nil
}
