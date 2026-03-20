package controllers

import (
	"context"
	"encoding/json"

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

// Arbiter is the HTTP controller for the Arbiter feature.
type Arbiter struct {
	settingsRepo ArbiterSettingsRepository
	scanRepo     ArbiterScannerRepository
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

// validStructures and validRigs/Securities for validation
var validStructures = map[string]bool{
	"raitaru": true, "azbel": true, "sotiyo": true,
	"tatara": true, "athanor": true,
	"station": true,
}
var validRigs = map[string]bool{"none": true, "t1": true, "t2": true}
var validSecurities = map[string]bool{"null": true, "low": true, "high": true}

type updateArbiterSettingsRequest struct {
	ReactionStructure string  `json:"reaction_structure"`
	ReactionRig       string  `json:"reaction_rig"`
	ReactionSecurity  string  `json:"reaction_security"`
	ReactionSystemID  *int64  `json:"reaction_system_id"`

	InventionStructure string  `json:"invention_structure"`
	InventionRig       string  `json:"invention_rig"`
	InventionSecurity  string  `json:"invention_security"`
	InventionSystemID  *int64  `json:"invention_system_id"`

	ComponentStructure string  `json:"component_structure"`
	ComponentRig       string  `json:"component_rig"`
	ComponentSecurity  string  `json:"component_security"`
	ComponentSystemID  *int64  `json:"component_system_id"`

	FinalStructure string  `json:"final_structure"`
	FinalRig       string  `json:"final_rig"`
	FinalSecurity  string  `json:"final_security"`
	FinalSystemID  *int64  `json:"final_system_id"`
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
	if !validSecurities[req.ReactionSecurity] {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid reaction_security")}
	}
	if !validStructures[req.InventionStructure] {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid invention_structure")}
	}
	if !validRigs[req.InventionRig] {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid invention_rig")}
	}
	if !validSecurities[req.InventionSecurity] {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid invention_security")}
	}
	if !validStructures[req.ComponentStructure] {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid component_structure")}
	}
	if !validRigs[req.ComponentRig] {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid component_rig")}
	}
	if !validSecurities[req.ComponentSecurity] {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid component_security")}
	}
	if !validStructures[req.FinalStructure] {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid final_structure")}
	}
	if !validRigs[req.FinalRig] {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid final_rig")}
	}
	if !validSecurities[req.FinalSecurity] {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid final_security")}
	}

	settings := &models.ArbiterSettings{
		UserID:             userID,
		ReactionStructure:  req.ReactionStructure,
		ReactionRig:        req.ReactionRig,
		ReactionSecurity:   req.ReactionSecurity,
		ReactionSystemID:   req.ReactionSystemID,
		InventionStructure: req.InventionStructure,
		InventionRig:       req.InventionRig,
		InventionSecurity:  req.InventionSecurity,
		InventionSystemID:  req.InventionSystemID,
		ComponentStructure: req.ComponentStructure,
		ComponentRig:       req.ComponentRig,
		ComponentSecurity:  req.ComponentSecurity,
		ComponentSystemID:  req.ComponentSystemID,
		FinalStructure:     req.FinalStructure,
		FinalRig:           req.FinalRig,
		FinalSecurity:      req.FinalSecurity,
		FinalSystemID:      req.FinalSystemID,
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

	result, err := services.ScanOpportunities(ctx, userID, settings, c.scanRepo)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to scan opportunities")}
	}

	return result, nil
}
