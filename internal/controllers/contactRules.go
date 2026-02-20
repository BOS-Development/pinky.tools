package controllers

import (
	"context"
	"encoding/json"
	"strconv"

	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type ContactRulesRepository interface {
	GetByUser(ctx context.Context, userID int64) ([]*models.ContactRule, error)
	Create(ctx context.Context, rule *models.ContactRule) error
	Delete(ctx context.Context, ruleID int64, userID int64) error
	DeleteAutoContactsForRule(ctx context.Context, ruleID int64) error
	SearchCorporations(ctx context.Context, query string) ([]*repositories.SearchResult, error)
	SearchAlliances(ctx context.Context, query string) ([]*repositories.SearchResult, error)
}

type ContactRuleApplier interface {
	ApplyRule(ctx context.Context, rule *models.ContactRule) error
	ApplyRulesForNewCorporation(ctx context.Context, userID int64, corpID int64, allianceID int64) error
}

type ContactRules struct {
	repository ContactRulesRepository
	applier    ContactRuleApplier
}

func NewContactRules(
	router Routerer,
	repository ContactRulesRepository,
	applier ContactRuleApplier,
) *ContactRules {
	controller := &ContactRules{
		repository: repository,
		applier:    applier,
	}

	router.RegisterRestAPIRoute("/v1/contact-rules", web.AuthAccessUser, controller.GetMyRules, "GET")
	router.RegisterRestAPIRoute("/v1/contact-rules", web.AuthAccessUser, controller.CreateRule, "POST")
	router.RegisterRestAPIRoute("/v1/contact-rules/{id}", web.AuthAccessUser, controller.DeleteRule, "DELETE")
	router.RegisterRestAPIRoute("/v1/contact-rules/corporations", web.AuthAccessUser, controller.SearchCorporations, "GET")
	router.RegisterRestAPIRoute("/v1/contact-rules/alliances", web.AuthAccessUser, controller.SearchAlliances, "GET")

	return controller
}

func (c *ContactRules) GetMyRules(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	rules, err := c.repository.GetByUser(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get contact rules")}
	}

	return rules, nil
}

func (c *ContactRules) CreateRule(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	var req struct {
		RuleType    string   `json:"ruleType"`
		EntityID    *int64   `json:"entityId"`
		EntityName  string   `json:"entityName"`
		Permissions []string `json:"permissions"`
	}

	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	if req.RuleType != "corporation" && req.RuleType != "alliance" && req.RuleType != "everyone" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("ruleType must be 'corporation', 'alliance', or 'everyone'")}
	}

	if req.RuleType != "everyone" && req.EntityID == nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("entityId is required for corporation and alliance rules")}
	}

	// Default to for_sale_browse if no permissions specified
	if len(req.Permissions) == 0 {
		req.Permissions = []string{"for_sale_browse"}
	}

	// Validate permissions against allowed service types
	allowedPermissions := map[string]bool{"for_sale_browse": true}
	for _, p := range req.Permissions {
		if !allowedPermissions[p] {
			return nil, &web.HttpError{StatusCode: 400, Error: errors.Errorf("invalid permission: %s", p)}
		}
	}

	var entityName *string
	if req.EntityName != "" {
		entityName = &req.EntityName
	}

	rule := &models.ContactRule{
		UserID:      *args.User,
		RuleType:    req.RuleType,
		EntityID:    req.EntityID,
		EntityName:  entityName,
		Permissions: req.Permissions,
	}

	if err := c.repository.Create(args.Request.Context(), rule); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to create contact rule")}
	}

	// Apply rule asynchronously
	go func() {
		ctx := context.Background()
		if err := c.applier.ApplyRule(ctx, rule); err != nil {
			log.Error("failed to apply contact rule", "ruleID", rule.ID, "ruleType", rule.RuleType, "error", err)
		}
	}()

	return rule, nil
}

func (c *ContactRules) DeleteRule(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	idStr, ok := args.Params["id"]
	if !ok {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("id is required")}
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid id")}
	}

	// Delete auto-created contacts first (CASCADE would handle this, but being explicit)
	if err := c.repository.DeleteAutoContactsForRule(args.Request.Context(), id); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to delete auto-contacts")}
	}

	if err := c.repository.Delete(args.Request.Context(), id, *args.User); err != nil {
		if err.Error() == "contact rule not found or user is not the owner" {
			return nil, &web.HttpError{StatusCode: 404, Error: err}
		}
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to delete contact rule")}
	}

	return nil, nil
}

func (c *ContactRules) SearchCorporations(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	q := args.Request.URL.Query().Get("q")
	if q == "" {
		return []*repositories.SearchResult{}, nil
	}

	results, err := c.repository.SearchCorporations(args.Request.Context(), q)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to search corporations")}
	}

	return results, nil
}

func (c *ContactRules) SearchAlliances(args *web.HandlerArgs) (any, *web.HttpError) {
	if args.User == nil {
		return nil, &web.HttpError{StatusCode: 401, Error: errors.New("unauthorized")}
	}

	q := args.Request.URL.Query().Get("q")
	if q == "" {
		return []*repositories.SearchResult{}, nil
	}

	results, err := c.repository.SearchAlliances(args.Request.Context(), q)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to search alliances")}
	}

	return results, nil
}
