package controllers

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/services"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

// ArbiterSettingsRepository handles arbiter settings reads and writes.
type ArbiterSettingsRepository interface {
	GetArbiterSettings(ctx context.Context, userID int64) (*models.ArbiterSettings, error)
	UpsertArbiterSettings(ctx context.Context, settings *models.ArbiterSettings) error
	GetArbiterEnabled(ctx context.Context, userID int64) (bool, error)
}

// ArbiterScannerRepository is the interface used by the scan endpoint.
// It embeds services.ArbiterScanRepository so the same concrete repo can satisfy both.
type ArbiterScannerRepository interface {
	services.ArbiterScanRepository
}

// ArbiterScopesRepository handles arbiter scopes CRUD.
type ArbiterScopesRepository interface {
	GetScopes(ctx context.Context, userID int64) ([]*models.ArbiterScope, error)
	GetScope(ctx context.Context, scopeID, userID int64) (*models.ArbiterScope, error)
	CreateScope(ctx context.Context, scope *models.ArbiterScope) (int64, error)
	UpdateScope(ctx context.Context, scope *models.ArbiterScope) error
	DeleteScope(ctx context.Context, scopeID, userID int64) error
	GetScopeMembers(ctx context.Context, scopeID int64) ([]*models.ArbiterScopeMember, error)
	AddScopeMember(ctx context.Context, member *models.ArbiterScopeMember) error
	RemoveScopeMember(ctx context.Context, memberID, scopeID int64) error
}

// ArbiterTaxProfileRepository handles tax profile reads and writes.
type ArbiterTaxProfileRepository interface {
	GetTaxProfile(ctx context.Context, userID int64) (*models.ArbiterTaxProfile, error)
	UpsertTaxProfile(ctx context.Context, profile *models.ArbiterTaxProfile) error
}

// ArbiterListsRepository handles blacklist and whitelist operations.
type ArbiterListsRepository interface {
	GetBlacklist(ctx context.Context, userID int64) ([]*models.ArbiterListItem, error)
	AddToBlacklist(ctx context.Context, userID, typeID int64) error
	RemoveFromBlacklist(ctx context.Context, userID, typeID int64) error
	GetWhitelist(ctx context.Context, userID int64) ([]*models.ArbiterListItem, error)
	AddToWhitelist(ctx context.Context, userID, typeID int64) error
	RemoveFromWhitelist(ctx context.Context, userID, typeID int64) error
}

// ArbiterBOMRepositoryInterface wraps the BOM repo for the controller.
// It extends the service-level interface with scope asset loading.
type ArbiterBOMRepositoryInterface interface {
	services.ArbiterBOMRepository
	GetScopeAssets(ctx context.Context, scopeID, userID int64) (map[int64]int64, error)
}

// ArbiterSolarSystemRepository handles solar system search.
type ArbiterSolarSystemRepository interface {
	SearchSolarSystems(ctx context.Context, query string, limit int) ([]*models.SolarSystemSearchResult, error)
}

// ArbiterDecryptorRepository handles decryptor lookups.
type ArbiterDecryptorRepository interface {
	GetDecryptors(ctx context.Context) ([]*models.Decryptor, error)
}

// Arbiter is the HTTP controller for the Arbiter feature.
type Arbiter struct {
	settingsRepo   ArbiterSettingsRepository
	scanRepo       ArbiterScannerRepository
	scopesRepo     ArbiterScopesRepository
	taxRepo        ArbiterTaxProfileRepository
	listsRepo      ArbiterListsRepository
	bomRepo        ArbiterBOMRepositoryInterface
	solarSysRepo   ArbiterSolarSystemRepository
	decryptorsRepo ArbiterDecryptorRepository
}

// NewArbiter creates and registers the Arbiter controller routes.
func NewArbiter(
	router Routerer,
	settingsRepo ArbiterSettingsRepository,
	scanRepo ArbiterScannerRepository,
) *Arbiter {
	c := &Arbiter{
		settingsRepo: settingsRepo,
		scanRepo:     scanRepo,
	}
	router.RegisterRestAPIRoute("/v1/arbiter/settings", web.AuthAccessUser, c.GetArbiterSettings, "GET")
	router.RegisterRestAPIRoute("/v1/arbiter/settings", web.AuthAccessUser, c.UpdateArbiterSettings, "PUT")
	router.RegisterRestAPIRoute("/v1/arbiter/opportunities", web.AuthAccessUser, c.GetArbiterOpportunities, "GET")
	return c
}

// NewArbiterFull creates the Arbiter controller with all optional repositories.
func NewArbiterFull(
	router Routerer,
	settingsRepo ArbiterSettingsRepository,
	scanRepo ArbiterScannerRepository,
	scopesRepo ArbiterScopesRepository,
	taxRepo ArbiterTaxProfileRepository,
	listsRepo ArbiterListsRepository,
	bomRepo ArbiterBOMRepositoryInterface,
	solarSysRepo ArbiterSolarSystemRepository,
	decryptorsRepo ArbiterDecryptorRepository,
) *Arbiter {
	c := &Arbiter{
		settingsRepo:   settingsRepo,
		scanRepo:       scanRepo,
		scopesRepo:     scopesRepo,
		taxRepo:        taxRepo,
		listsRepo:      listsRepo,
		bomRepo:        bomRepo,
		solarSysRepo:   solarSysRepo,
		decryptorsRepo: decryptorsRepo,
	}
	router.RegisterRestAPIRoute("/v1/arbiter/settings", web.AuthAccessUser, c.GetArbiterSettings, "GET")
	router.RegisterRestAPIRoute("/v1/arbiter/settings", web.AuthAccessUser, c.UpdateArbiterSettings, "PUT")
	router.RegisterRestAPIRoute("/v1/arbiter/opportunities", web.AuthAccessUser, c.GetArbiterOpportunities, "GET")

	// Scopes
	router.RegisterRestAPIRoute("/v1/arbiter/scopes", web.AuthAccessUser, c.GetScopes, "GET")
	router.RegisterRestAPIRoute("/v1/arbiter/scopes", web.AuthAccessUser, c.CreateScope, "POST")
	router.RegisterRestAPIRoute("/v1/arbiter/scopes/{id}", web.AuthAccessUser, c.UpdateScope, "PUT")
	router.RegisterRestAPIRoute("/v1/arbiter/scopes/{id}", web.AuthAccessUser, c.DeleteScope, "DELETE")
	router.RegisterRestAPIRoute("/v1/arbiter/scopes/{id}/members", web.AuthAccessUser, c.GetScopeMembers, "GET")
	router.RegisterRestAPIRoute("/v1/arbiter/scopes/{id}/members", web.AuthAccessUser, c.AddScopeMember, "POST")
	router.RegisterRestAPIRoute("/v1/arbiter/scopes/{id}/members/{memberID}", web.AuthAccessUser, c.RemoveScopeMember, "DELETE")

	// Tax profile
	router.RegisterRestAPIRoute("/v1/arbiter/tax-profile", web.AuthAccessUser, c.GetTaxProfile, "GET")
	router.RegisterRestAPIRoute("/v1/arbiter/tax-profile", web.AuthAccessUser, c.UpdateTaxProfile, "PUT")

	// Blacklist
	router.RegisterRestAPIRoute("/v1/arbiter/blacklist", web.AuthAccessUser, c.GetBlacklist, "GET")
	router.RegisterRestAPIRoute("/v1/arbiter/blacklist", web.AuthAccessUser, c.AddToBlacklist, "POST")
	router.RegisterRestAPIRoute("/v1/arbiter/blacklist/{typeID}", web.AuthAccessUser, c.RemoveFromBlacklist, "DELETE")

	// Whitelist
	router.RegisterRestAPIRoute("/v1/arbiter/whitelist", web.AuthAccessUser, c.GetWhitelist, "GET")
	router.RegisterRestAPIRoute("/v1/arbiter/whitelist", web.AuthAccessUser, c.AddToWhitelist, "POST")
	router.RegisterRestAPIRoute("/v1/arbiter/whitelist/{typeID}", web.AuthAccessUser, c.RemoveFromWhitelist, "DELETE")

	// BOM tree
	router.RegisterRestAPIRoute("/v1/arbiter/opportunities/{typeID}/bom", web.AuthAccessUser, c.GetBOMTree, "GET")

	// Solar systems search (shared route)
	router.RegisterRestAPIRoute("/v1/solar-systems/search", web.AuthAccessUser, c.SearchSolarSystems, "GET")

	// Decryptors
	router.RegisterRestAPIRoute("/v1/arbiter/decryptors", web.AuthAccessUser, c.GetDecryptors, "GET")

	return c
}

// checkFeatureGate returns an HttpError if the user does not have the Arbiter feature enabled.
func (c *Arbiter) checkFeatureGate(ctx context.Context, userID int64) *web.HttpError {
	enabled, err := c.settingsRepo.GetArbiterEnabled(ctx, userID)
	if err != nil {
		return &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to check arbiter feature flag")}
	}
	if !enabled {
		return &web.HttpError{StatusCode: 403, Error: errors.New("Arbiter feature not enabled for this user")}
	}
	return nil
}

// GetArbiterSettings returns the current Arbiter settings for the user.
func (c *Arbiter) GetArbiterSettings(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User

	if httpErr := c.checkFeatureGate(ctx, userID); httpErr != nil {
		return nil, httpErr
	}

	settings, err := c.settingsRepo.GetArbiterSettings(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get arbiter settings")}
	}
	return settings, nil
}

// validStructures and validRigs for validation
var validStructures = map[string]bool{
	"raitaru": true, "azbel": true, "sotiyo": true,
	"tatara": true, "athanor": true,
	"station": true,
}
var validRigs = map[string]bool{"none": true, "t1": true, "t2": true}

type updateArbiterSettingsRequest struct {
	ReactionStructure string `json:"reaction_structure"`
	ReactionRig       string `json:"reaction_rig"`
	ReactionSystemID  *int64 `json:"reaction_system_id"`

	InventionStructure string `json:"invention_structure"`
	InventionRig       string `json:"invention_rig"`
	InventionSystemID  *int64 `json:"invention_system_id"`

	ComponentStructure string `json:"component_structure"`
	ComponentRig       string `json:"component_rig"`
	ComponentSystemID  *int64 `json:"component_system_id"`

	FinalStructure string `json:"final_structure"`
	FinalRig       string `json:"final_rig"`
	FinalSystemID  *int64 `json:"final_system_id"`

	FinalFacilityTax     float64 `json:"final_facility_tax"`
	ComponentFacilityTax float64 `json:"component_facility_tax"`
	ReactionFacilityTax  float64 `json:"reaction_facility_tax"`
	InventionFacilityTax float64 `json:"invention_facility_tax"`

	UseWhitelist    bool   `json:"use_whitelist"`
	UseBlacklist    bool   `json:"use_blacklist"`
	DecryptorTypeID *int64 `json:"decryptor_type_id"`
	DefaultScopeID  *int64 `json:"default_scope_id"`
}

// UpdateArbiterSettings saves new Arbiter settings for the user.
func (c *Arbiter) UpdateArbiterSettings(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User

	if httpErr := c.checkFeatureGate(ctx, userID); httpErr != nil {
		return nil, httpErr
	}

	var req updateArbiterSettingsRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	// Validate
	if !validStructures[req.ReactionStructure] {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid reaction_structure")}
	}
	if !validRigs[req.ReactionRig] {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid reaction_rig")}
	}
	if !validStructures[req.InventionStructure] {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid invention_structure")}
	}
	if !validRigs[req.InventionRig] {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid invention_rig")}
	}
	if !validStructures[req.ComponentStructure] {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid component_structure")}
	}
	if !validRigs[req.ComponentRig] {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid component_rig")}
	}
	if !validStructures[req.FinalStructure] {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid final_structure")}
	}
	if !validRigs[req.FinalRig] {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid final_rig")}
	}

	settings := &models.ArbiterSettings{
		UserID:             userID,
		ReactionStructure:  req.ReactionStructure,
		ReactionRig:        req.ReactionRig,
		ReactionSystemID:   req.ReactionSystemID,
		InventionStructure: req.InventionStructure,
		InventionRig:       req.InventionRig,
		InventionSystemID:  req.InventionSystemID,
		ComponentStructure: req.ComponentStructure,
		ComponentRig:       req.ComponentRig,
		ComponentSystemID:  req.ComponentSystemID,
		FinalStructure:      req.FinalStructure,
		FinalRig:            req.FinalRig,
		FinalSystemID:       req.FinalSystemID,
		FinalFacilityTax:    req.FinalFacilityTax,
		ComponentFacilityTax: req.ComponentFacilityTax,
		ReactionFacilityTax:  req.ReactionFacilityTax,
		InventionFacilityTax: req.InventionFacilityTax,
		UseWhitelist:        req.UseWhitelist,
		UseBlacklist:       req.UseBlacklist,
		DecryptorTypeID:    req.DecryptorTypeID,
		DefaultScopeID:     req.DefaultScopeID,
	}

	if err := c.settingsRepo.UpsertArbiterSettings(ctx, settings); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to save arbiter settings")}
	}
	return settings, nil
}

// GetArbiterOpportunities runs the full T2 opportunity scan and returns ranked results.
func (c *Arbiter) GetArbiterOpportunities(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User

	if httpErr := c.checkFeatureGate(ctx, userID); httpErr != nil {
		return nil, httpErr
	}

	settings, err := c.settingsRepo.GetArbiterSettings(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get arbiter settings")}
	}

	var taxProfile *models.ArbiterTaxProfile
	if c.taxRepo != nil {
		taxProfile, err = c.taxRepo.GetTaxProfile(ctx, userID)
		if err != nil {
			return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get tax profile")}
		}
	}

	buildAll := args.Request.URL.Query().Get("build_all") == "true"
	result, err := services.ScanOpportunities(ctx, userID, settings, taxProfile, buildAll, c.scanRepo)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to scan opportunities")}
	}

	return result, nil
}

// --- Scopes ---

// GetScopes returns all arbiter scopes for the user.
func (c *Arbiter) GetScopes(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User

	if httpErr := c.checkFeatureGate(ctx, userID); httpErr != nil {
		return nil, httpErr
	}
	if c.scopesRepo == nil {
		return nil, &web.HttpError{StatusCode: 501, Error: errors.New("scopes not configured")}
	}

	scopes, err := c.scopesRepo.GetScopes(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get scopes")}
	}
	return scopes, nil
}

type createScopeRequest struct {
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

// CreateScope creates a new arbiter scope for the user.
func (c *Arbiter) CreateScope(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User

	if httpErr := c.checkFeatureGate(ctx, userID); httpErr != nil {
		return nil, httpErr
	}
	if c.scopesRepo == nil {
		return nil, &web.HttpError{StatusCode: 501, Error: errors.New("scopes not configured")}
	}

	var req createScopeRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}
	if req.Name == "" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("name is required")}
	}

	scope := &models.ArbiterScope{
		UserID:    userID,
		Name:      req.Name,
		IsDefault: req.IsDefault,
	}
	id, err := c.scopesRepo.CreateScope(ctx, scope)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to create scope")}
	}
	scope.ID = id
	return scope, nil
}

type updateScopeRequest struct {
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

// UpdateScope updates an existing arbiter scope.
func (c *Arbiter) UpdateScope(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User

	if httpErr := c.checkFeatureGate(ctx, userID); httpErr != nil {
		return nil, httpErr
	}
	if c.scopesRepo == nil {
		return nil, &web.HttpError{StatusCode: 501, Error: errors.New("scopes not configured")}
	}

	scopeID, err := strconv.ParseInt(args.Params["id"], 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid scope id")}
	}

	var req updateScopeRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}
	if req.Name == "" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("name is required")}
	}

	scope := &models.ArbiterScope{
		ID:        scopeID,
		UserID:    userID,
		Name:      req.Name,
		IsDefault: req.IsDefault,
	}
	if err := c.scopesRepo.UpdateScope(ctx, scope); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to update scope")}
	}
	return scope, nil
}

// DeleteScope deletes an arbiter scope.
func (c *Arbiter) DeleteScope(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User

	if httpErr := c.checkFeatureGate(ctx, userID); httpErr != nil {
		return nil, httpErr
	}
	if c.scopesRepo == nil {
		return nil, &web.HttpError{StatusCode: 501, Error: errors.New("scopes not configured")}
	}

	scopeID, err := strconv.ParseInt(args.Params["id"], 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid scope id")}
	}

	if err := c.scopesRepo.DeleteScope(ctx, scopeID, userID); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to delete scope")}
	}
	return map[string]bool{"ok": true}, nil
}

// GetScopeMembers returns all members of a scope.
func (c *Arbiter) GetScopeMembers(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User

	if httpErr := c.checkFeatureGate(ctx, userID); httpErr != nil {
		return nil, httpErr
	}
	if c.scopesRepo == nil {
		return nil, &web.HttpError{StatusCode: 501, Error: errors.New("scopes not configured")}
	}

	scopeID, err := strconv.ParseInt(args.Params["id"], 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid scope id")}
	}

	// Verify user owns the scope
	scope, err := c.scopesRepo.GetScope(ctx, scopeID, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to verify scope ownership")}
	}
	if scope == nil {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.New("scope not found")}
	}

	members, err := c.scopesRepo.GetScopeMembers(ctx, scopeID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get scope members")}
	}
	return members, nil
}

type addScopeMemberRequest struct {
	MemberType string `json:"member_type"` // "character" or "corporation"
	MemberID   int64  `json:"member_id"`
}

// AddScopeMember adds a character or corporation to a scope.
func (c *Arbiter) AddScopeMember(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User

	if httpErr := c.checkFeatureGate(ctx, userID); httpErr != nil {
		return nil, httpErr
	}
	if c.scopesRepo == nil {
		return nil, &web.HttpError{StatusCode: 501, Error: errors.New("scopes not configured")}
	}

	scopeID, err := strconv.ParseInt(args.Params["id"], 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid scope id")}
	}

	// Verify user owns the scope
	scope, err := c.scopesRepo.GetScope(ctx, scopeID, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to verify scope ownership")}
	}
	if scope == nil {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.New("scope not found")}
	}

	var req addScopeMemberRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}
	if req.MemberType != "character" && req.MemberType != "corporation" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("member_type must be 'character' or 'corporation'")}
	}
	if req.MemberID == 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("member_id is required")}
	}

	member := &models.ArbiterScopeMember{
		ScopeID:    scopeID,
		MemberType: req.MemberType,
		MemberID:   req.MemberID,
	}
	if err := c.scopesRepo.AddScopeMember(ctx, member); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to add scope member")}
	}
	return member, nil
}

// RemoveScopeMember removes a member from a scope.
func (c *Arbiter) RemoveScopeMember(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User

	if httpErr := c.checkFeatureGate(ctx, userID); httpErr != nil {
		return nil, httpErr
	}
	if c.scopesRepo == nil {
		return nil, &web.HttpError{StatusCode: 501, Error: errors.New("scopes not configured")}
	}

	scopeID, err := strconv.ParseInt(args.Params["id"], 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid scope id")}
	}
	memberID, err := strconv.ParseInt(args.Params["memberID"], 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid member id")}
	}

	// Verify user owns the scope
	scope, err := c.scopesRepo.GetScope(ctx, scopeID, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to verify scope ownership")}
	}
	if scope == nil {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.New("scope not found")}
	}

	if err := c.scopesRepo.RemoveScopeMember(ctx, memberID, scopeID); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to remove scope member")}
	}
	return map[string]bool{"ok": true}, nil
}

// --- Tax Profile ---

// GetTaxProfile returns the user's Arbiter tax profile.
func (c *Arbiter) GetTaxProfile(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User

	if httpErr := c.checkFeatureGate(ctx, userID); httpErr != nil {
		return nil, httpErr
	}
	if c.taxRepo == nil {
		return nil, &web.HttpError{StatusCode: 501, Error: errors.New("tax profile not configured")}
	}

	profile, err := c.taxRepo.GetTaxProfile(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get tax profile")}
	}
	return profile, nil
}

type updateTaxProfileRequest struct {
	TraderCharacterID  *int64  `json:"trader_character_id"`
	SalesTaxRate       float64 `json:"sales_tax_rate"`
	BrokerFeeRate      float64 `json:"broker_fee_rate"`
	StructureBrokerFee float64 `json:"structure_broker_fee"`
	InputPriceType     string  `json:"input_price_type"`
	OutputPriceType    string  `json:"output_price_type"`
}

// UpdateTaxProfile updates the user's Arbiter tax profile.
func (c *Arbiter) UpdateTaxProfile(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User

	if httpErr := c.checkFeatureGate(ctx, userID); httpErr != nil {
		return nil, httpErr
	}
	if c.taxRepo == nil {
		return nil, &web.HttpError{StatusCode: 501, Error: errors.New("tax profile not configured")}
	}

	var req updateTaxProfileRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}
	if req.InputPriceType != "buy" && req.InputPriceType != "sell" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("input_price_type must be 'buy' or 'sell'")}
	}
	if req.OutputPriceType != "buy" && req.OutputPriceType != "sell" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("output_price_type must be 'buy' or 'sell'")}
	}

	profile := &models.ArbiterTaxProfile{
		UserID:             userID,
		TraderCharacterID:  req.TraderCharacterID,
		SalesTaxRate:       req.SalesTaxRate,
		BrokerFeeRate:      req.BrokerFeeRate,
		StructureBrokerFee: req.StructureBrokerFee,
		InputPriceType:     req.InputPriceType,
		OutputPriceType:    req.OutputPriceType,
	}
	if err := c.taxRepo.UpsertTaxProfile(ctx, profile); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to update tax profile")}
	}
	return profile, nil
}

// --- Blacklist ---

// GetBlacklist returns all blacklisted items for the user.
func (c *Arbiter) GetBlacklist(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User

	if httpErr := c.checkFeatureGate(ctx, userID); httpErr != nil {
		return nil, httpErr
	}
	if c.listsRepo == nil {
		return nil, &web.HttpError{StatusCode: 501, Error: errors.New("lists not configured")}
	}

	items, err := c.listsRepo.GetBlacklist(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get blacklist")}
	}
	return items, nil
}

type addListItemRequest struct {
	TypeID int64 `json:"type_id"`
}

// AddToBlacklist adds a type to the user's blacklist.
func (c *Arbiter) AddToBlacklist(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User

	if httpErr := c.checkFeatureGate(ctx, userID); httpErr != nil {
		return nil, httpErr
	}
	if c.listsRepo == nil {
		return nil, &web.HttpError{StatusCode: 501, Error: errors.New("lists not configured")}
	}

	var req addListItemRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}
	if req.TypeID == 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("type_id is required")}
	}

	if err := c.listsRepo.AddToBlacklist(ctx, userID, req.TypeID); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to add to blacklist")}
	}
	return map[string]bool{"ok": true}, nil
}

// RemoveFromBlacklist removes a type from the user's blacklist.
func (c *Arbiter) RemoveFromBlacklist(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User

	if httpErr := c.checkFeatureGate(ctx, userID); httpErr != nil {
		return nil, httpErr
	}
	if c.listsRepo == nil {
		return nil, &web.HttpError{StatusCode: 501, Error: errors.New("lists not configured")}
	}

	typeIDStr := args.Params["typeID"]
	typeID, err := strconv.ParseInt(typeIDStr, 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid type id")}
	}

	if err := c.listsRepo.RemoveFromBlacklist(ctx, userID, typeID); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to remove from blacklist")}
	}
	return map[string]bool{"ok": true}, nil
}

// --- Whitelist ---

// GetWhitelist returns all whitelisted items for the user.
func (c *Arbiter) GetWhitelist(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User

	if httpErr := c.checkFeatureGate(ctx, userID); httpErr != nil {
		return nil, httpErr
	}
	if c.listsRepo == nil {
		return nil, &web.HttpError{StatusCode: 501, Error: errors.New("lists not configured")}
	}

	items, err := c.listsRepo.GetWhitelist(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get whitelist")}
	}
	return items, nil
}

// AddToWhitelist adds a type to the user's whitelist.
func (c *Arbiter) AddToWhitelist(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User

	if httpErr := c.checkFeatureGate(ctx, userID); httpErr != nil {
		return nil, httpErr
	}
	if c.listsRepo == nil {
		return nil, &web.HttpError{StatusCode: 501, Error: errors.New("lists not configured")}
	}

	var req addListItemRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}
	if req.TypeID == 0 {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("type_id is required")}
	}

	if err := c.listsRepo.AddToWhitelist(ctx, userID, req.TypeID); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to add to whitelist")}
	}
	return map[string]bool{"ok": true}, nil
}

// RemoveFromWhitelist removes a type from the user's whitelist.
func (c *Arbiter) RemoveFromWhitelist(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User

	if httpErr := c.checkFeatureGate(ctx, userID); httpErr != nil {
		return nil, httpErr
	}
	if c.listsRepo == nil {
		return nil, &web.HttpError{StatusCode: 501, Error: errors.New("lists not configured")}
	}

	typeIDStr := args.Params["typeID"]
	typeID, err := strconv.ParseInt(typeIDStr, 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid type id")}
	}

	if err := c.listsRepo.RemoveFromWhitelist(ctx, userID, typeID); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to remove from whitelist")}
	}
	return map[string]bool{"ok": true}, nil
}

// --- BOM Tree ---

// GetBOMTree returns a full recursive BOM tree for a specific item.
func (c *Arbiter) GetBOMTree(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	userID := *args.User

	if httpErr := c.checkFeatureGate(ctx, userID); httpErr != nil {
		return nil, httpErr
	}
	if c.bomRepo == nil || c.settingsRepo == nil {
		return nil, &web.HttpError{StatusCode: 501, Error: errors.New("BOM tree not configured")}
	}

	typeIDStr := args.Params["typeID"]
	typeID, err := strconv.ParseInt(typeIDStr, 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid type id")}
	}

	qtyStr := args.Request.URL.Query().Get("quantity")
	qty := int64(1)
	if qtyStr != "" {
		if v, err := strconv.ParseInt(qtyStr, 10, 64); err == nil && v > 0 {
			qty = v
		}
	}

	meStr := args.Request.URL.Query().Get("me")
	me := 0
	if meStr != "" {
		if v, err := strconv.Atoi(meStr); err == nil {
			me = v
		}
	}

	// Resolve settings for BOM calculation
	settings, err := c.settingsRepo.GetArbiterSettings(ctx, userID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get settings")}
	}

	// Build empty blacklist/whitelist/assets
	blacklist := map[int64]bool{}
	whitelist := map[int64]bool{}
	assets := map[int64]int64{}

	// Use lists if available
	if c.listsRepo != nil {
		bl, err := c.listsRepo.GetBlacklist(ctx, userID)
		if err == nil {
			for _, item := range bl {
				blacklist[item.TypeID] = true
			}
		}
		wl, err := c.listsRepo.GetWhitelist(ctx, userID)
		if err == nil {
			for _, item := range wl {
				whitelist[item.TypeID] = true
			}
		}
	}

	// Load scope assets if scope_id is provided
	scopeIDStr := args.Request.URL.Query().Get("scope_id")
	if scopeIDStr != "" {
		if scopeID, err := strconv.ParseInt(scopeIDStr, 10, 64); err == nil {
			if sa, err := c.bomRepo.GetScopeAssets(ctx, scopeID, userID); err == nil {
				assets = sa
			}
		}
	}

	// Parse build_all flag
	buildAll := args.Request.URL.Query().Get("build_all") == "true"

	// Get the blueprint for this product
	bpID, err := c.bomRepo.GetBlueprintForProduct(ctx, typeID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get blueprint")}
	}
	if bpID == 0 {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.New("no blueprint found for this type")}
	}

	inputPriceType := args.Request.URL.Query().Get("input_price_type")
	if inputPriceType == "" {
		inputPriceType = "sell"
	}

	tree, err := services.BuildBOMTree(ctx, bpID, typeID, "", qty, me, c.bomRepo, settings, blacklist, whitelist, assets, inputPriceType, buildAll)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to build BOM tree")}
	}
	return tree, nil
}

// --- Solar System Search ---

// SearchSolarSystems returns solar systems matching the query string.
func (c *Arbiter) SearchSolarSystems(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	// No feature gate — solar system search is general utility

	if c.solarSysRepo == nil {
		return nil, &web.HttpError{StatusCode: 501, Error: errors.New("solar system search not configured")}
	}

	q := args.Request.URL.Query().Get("q")
	if q == "" {
		return []*models.SolarSystemSearchResult{}, nil
	}

	limitStr := args.Request.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			limit = v
		}
	}

	results, err := c.solarSysRepo.SearchSolarSystems(ctx, q, limit)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to search solar systems")}
	}
	return results, nil
}

// --- Decryptors ---

// GetDecryptors returns all available decryptors.
func (c *Arbiter) GetDecryptors(args *web.HandlerArgs) (any, *web.HttpError) {
	ctx := args.Request.Context()
	decryptors, err := c.decryptorsRepo.GetDecryptors(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get decryptors")}
	}
	return decryptors, nil
}
