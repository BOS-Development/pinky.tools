package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

// TradingStationsRepository provides access to NPC station presets.
type TradingStationsRepository interface {
	ListStations(ctx context.Context) ([]*models.TradingStation, error)
}

// UserTradingStructuresRepository provides CRUD for user-configured structures.
type UserTradingStructuresRepository interface {
	List(ctx context.Context, userID int64) ([]*models.UserTradingStructure, error)
	Upsert(ctx context.Context, s *models.UserTradingStructure) (*models.UserTradingStructure, error)
	Delete(ctx context.Context, id int64, userID int64) error
	UpdateAccessStatus(ctx context.Context, userID int64, structureID int64, accessOK bool) error
}

// TradingStructureMarketUpdater scans a player structure's market orders.
type TradingStructureMarketUpdater interface {
	ScanStructure(ctx context.Context, structureID int64, token string) (bool, error)
}

// TradingStructureCharacterRepository provides character list for a user.
type TradingStructureCharacterRepository interface {
	GetAll(ctx context.Context, userID int64) ([]*repositories.Character, error)
}

// TradingStructureEsiClient provides ESI calls for structure info and token refresh.
type TradingStructureEsiClient interface {
	GetStructureInfo(ctx context.Context, structureID int64, token string) (*client.StructureInfo, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*client.RefreshedToken, error)
}

// TradingStructureSolarSystemRepository provides region lookup from system ID.
type TradingStructureSolarSystemRepository interface {
	GetRegionIDBySystemID(ctx context.Context, systemID int64) (int64, error)
}

// TradingStructureAssetRepository returns structure IDs from character assets.
type TradingStructureAssetRepository interface {
	GetPlayerOwnedStationIDs(ctx context.Context, characterID, userID int64) ([]int64, error)
}

// TradingStructuresController handles station and structure configuration endpoints.
type TradingStructuresController struct {
	stations   TradingStationsRepository
	structures UserTradingStructuresRepository
	scanner    TradingStructureMarketUpdater
	characters TradingStructureCharacterRepository
	esi        TradingStructureEsiClient
	systems    TradingStructureSolarSystemRepository
	assets     TradingStructureAssetRepository
}

func NewTradingStructures(
	router Routerer,
	stations TradingStationsRepository,
	structures UserTradingStructuresRepository,
	scanner TradingStructureMarketUpdater,
	characters TradingStructureCharacterRepository,
	esi TradingStructureEsiClient,
	systems TradingStructureSolarSystemRepository,
	assets TradingStructureAssetRepository,
) *TradingStructuresController {
	c := &TradingStructuresController{
		stations:   stations,
		structures: structures,
		scanner:    scanner,
		characters: characters,
		esi:        esi,
		systems:    systems,
		assets:     assets,
	}
	router.RegisterRestAPIRoute("/v1/hauling/stations", web.AuthAccessUser, c.ListStations, "GET")
	router.RegisterRestAPIRoute("/v1/hauling/structures", web.AuthAccessUser, c.ListStructures, "GET")
	router.RegisterRestAPIRoute("/v1/hauling/structures", web.AuthAccessUser, c.AddStructure, "POST")
	router.RegisterRestAPIRoute("/v1/hauling/structures/{id}", web.AuthAccessUser, c.DeleteStructure, "DELETE")
	router.RegisterRestAPIRoute("/v1/hauling/structures/{id}/scan", web.AuthAccessUser, c.ScanStructure, "POST")
	router.RegisterRestAPIRoute("/v1/hauling/characters/{id}/asset-structures", web.AuthAccessUser, c.ListCharacterAssetStructures, "GET")
	return c
}

// ListStations returns all trading stations (NPC presets first).
func (c *TradingStructuresController) ListStations(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	stations, err := c.stations.ListStations(args.Request.Context())
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to list stations")}
	}
	return stations, nil
}

// ListStructures returns all user-configured structures.
func (c *TradingStructuresController) ListStructures(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	structures, err := c.structures.List(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to list structures")}
	}
	return structures, nil
}

// AddStructure resolves a structure via ESI and stores it.
func (c *TradingStructuresController) AddStructure(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	var body struct {
		StructureID int64 `json:"structureId"`
		CharacterID int64 `json:"characterId"`
	}
	if err := json.NewDecoder(args.Request.Body).Decode(&body); err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.Wrap(err, "invalid request body")}
	}
	if body.StructureID == 0 {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.New("structureId is required")}
	}
	if body.CharacterID == 0 {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.New("characterId is required")}
	}

	ctx := args.Request.Context()

	// Find the character to get their ESI token
	chars, err := c.characters.GetAll(ctx, *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to get characters")}
	}
	var char *repositories.Character
	for _, ch := range chars {
		if ch.ID == body.CharacterID {
			char = ch
			break
		}
	}
	if char == nil {
		return nil, &web.HttpError{StatusCode: http.StatusNotFound, Error: errors.New("character not found")}
	}

	// Refresh token
	refreshed, err := c.esi.RefreshAccessToken(ctx, char.EsiRefreshToken)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to refresh token")}
	}

	// Fetch structure info from ESI
	info, err := c.esi.GetStructureInfo(ctx, body.StructureID, refreshed.AccessToken)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to get structure info")}
	}
	if info == nil {
		// 403 — no access
		return map[string]interface{}{"accessOk": false, "error": "Access denied to structure"}, nil
	}

	// Look up region from system
	regionID, err := c.systems.GetRegionIDBySystemID(ctx, info.SolarSystemID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to get region for system")}
	}

	// Upsert the structure
	s := &models.UserTradingStructure{
		UserID:      *args.User,
		StructureID: body.StructureID,
		Name:        info.Name,
		SystemID:    info.SolarSystemID,
		RegionID:    regionID,
		CharacterID: body.CharacterID,
		AccessOK:    true,
	}
	created, err := c.structures.Upsert(ctx, s)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to upsert structure")}
	}

	// Trigger market scan async
	structureID := body.StructureID
	token := refreshed.AccessToken
	go func() {
		bgCtx := context.Background()
		if _, err := c.scanner.ScanStructure(bgCtx, structureID, token); err != nil {
			// Log but don't fail the request
			_ = err
		}
	}()

	return created, nil
}

// DeleteStructure removes a user-configured structure.
func (c *TradingStructuresController) DeleteStructure(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	id, err := strconv.ParseInt(args.Params["id"], 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.New("invalid id")}
	}
	if err := c.structures.Delete(args.Request.Context(), id, *args.User); err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to delete structure")}
	}
	return nil, nil
}

// ListCharacterAssetStructures returns structure IDs from the character's assets, with names where known.
func (c *TradingStructuresController) ListCharacterAssetStructures(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	charID, err := strconv.ParseInt(args.Params["id"], 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.New("invalid character id")}
	}

	ctx := args.Request.Context()

	// Validate character belongs to user
	chars, err := c.characters.GetAll(ctx, *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to get characters")}
	}
	var targetChar *repositories.Character
	for _, ch := range chars {
		if ch.ID == charID {
			targetChar = ch
			break
		}
	}
	if targetChar == nil {
		return nil, &web.HttpError{StatusCode: http.StatusNotFound, Error: errors.New("character not found")}
	}

	// Get structure IDs from character assets
	structureIDs, err := c.assets.GetPlayerOwnedStationIDs(ctx, charID, *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to get asset structures")}
	}

	// Build name map from already-saved structures
	savedStructures, err := c.structures.List(ctx, *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to list structures")}
	}
	nameByID := make(map[int64]string)
	for _, s := range savedStructures {
		nameByID[s.StructureID] = s.Name
	}

	// Resolve names for structures not in saved structures via ESI
	unknownIDs := []int64{}
	for _, id := range structureIDs {
		if nameByID[id] == "" {
			unknownIDs = append(unknownIDs, id)
		}
	}
	if len(unknownIDs) > 0 {
		refreshed, err := c.esi.RefreshAccessToken(ctx, targetChar.EsiRefreshToken)
		if err == nil {
			for _, id := range unknownIDs {
				info, err := c.esi.GetStructureInfo(ctx, id, refreshed.AccessToken)
				if err == nil && info != nil {
					nameByID[id] = info.Name
				}
			}
		}
	}

	type assetStructure struct {
		StructureID int64  `json:"structureId"`
		Name        string `json:"name"`
	}
	results := []assetStructure{}
	for _, id := range structureIDs {
		results = append(results, assetStructure{
			StructureID: id,
			Name:        nameByID[id],
		})
	}
	return results, nil
}

// ScanStructure triggers a fresh market scan for a structure.
func (c *TradingStructuresController) ScanStructure(args *web.HandlerArgs) (interface{}, *web.HttpError) {
	id, err := strconv.ParseInt(args.Params["id"], 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusBadRequest, Error: errors.New("invalid id")}
	}

	ctx := args.Request.Context()

	// Find the structure record
	structures, err := c.structures.List(ctx, *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to list structures")}
	}
	var s *models.UserTradingStructure
	for _, st := range structures {
		if st.ID == id {
			s = st
			break
		}
	}
	if s == nil {
		return nil, &web.HttpError{StatusCode: http.StatusNotFound, Error: errors.New("structure not found")}
	}

	// Find the character
	chars, err := c.characters.GetAll(ctx, *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to get characters")}
	}
	var char *repositories.Character
	for _, ch := range chars {
		if ch.ID == s.CharacterID {
			char = ch
			break
		}
	}
	if char == nil {
		return nil, &web.HttpError{StatusCode: http.StatusNotFound, Error: errors.New("character not found")}
	}

	// Refresh token
	refreshed, err := c.esi.RefreshAccessToken(ctx, char.EsiRefreshToken)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to refresh token")}
	}

	// Scan structure
	accessOK, err := c.scanner.ScanStructure(ctx, s.StructureID, refreshed.AccessToken)
	if err != nil {
		return nil, &web.HttpError{StatusCode: http.StatusInternalServerError, Error: errors.Wrap(err, "failed to scan structure")}
	}

	// Update access status if denied
	if !accessOK {
		_ = c.structures.UpdateAccessStatus(ctx, *args.User, s.StructureID, false)
	}

	return map[string]interface{}{"accessOk": accessOK}, nil
}
