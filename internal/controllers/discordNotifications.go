package controllers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/pkg/errors"
)

type DiscordNotificationsRepository interface {
	CreateLink(ctx context.Context, link *models.DiscordLink) error
	GetLinkByUser(ctx context.Context, userID int64) (*models.DiscordLink, error)
	UpdateLinkTokens(ctx context.Context, userID int64, accessToken, refreshToken string, expiresAt time.Time) error
	DeleteLink(ctx context.Context, userID int64) error
	CreateTarget(ctx context.Context, target *models.DiscordNotificationTarget) error
	GetTargetsByUser(ctx context.Context, userID int64) ([]*models.DiscordNotificationTarget, error)
	GetTargetByID(ctx context.Context, id int64) (*models.DiscordNotificationTarget, error)
	UpdateTarget(ctx context.Context, target *models.DiscordNotificationTarget) error
	DeleteTarget(ctx context.Context, id int64, userID int64) error
	UpsertPreference(ctx context.Context, pref *models.NotificationPreference) error
	GetPreferencesByTarget(ctx context.Context, targetID int64) ([]*models.NotificationPreference, error)
}

type DiscordClientProvider interface {
	GetUserGuilds(ctx context.Context, userAccessToken string) ([]*models.DiscordGuild, error)
	GetBotGuilds(ctx context.Context) ([]*models.DiscordGuild, error)
	GetGuildChannels(ctx context.Context, guildID string) ([]*models.DiscordChannel, error)
	SendDM(ctx context.Context, discordUserID string, embed *client.DiscordEmbed) error
	SendChannelMessage(ctx context.Context, channelID string, embed *client.DiscordEmbed) error
}

type DiscordTestNotifier interface {
	SendTestNotification(ctx context.Context, target *models.DiscordNotificationTarget, discordLink *models.DiscordLink) error
}

type DiscordNotifications struct {
	router        Routerer
	repository    DiscordNotificationsRepository
	discordClient DiscordClientProvider
	notifier      DiscordTestNotifier
}

func NewDiscordNotifications(router Routerer, repository DiscordNotificationsRepository, discordClient DiscordClientProvider, notifier DiscordTestNotifier) *DiscordNotifications {
	controller := &DiscordNotifications{
		router:        router,
		repository:    repository,
		discordClient: discordClient,
		notifier:      notifier,
	}

	// Discord link
	router.RegisterRestAPIRoute("/v1/discord/link", web.AuthAccessUser, controller.GetMyLink, "GET")
	router.RegisterRestAPIRoute("/v1/discord/link", web.AuthAccessUser, controller.SaveDiscordLink, "POST")
	router.RegisterRestAPIRoute("/v1/discord/link", web.AuthAccessUser, controller.UnlinkDiscord, "DELETE")

	// Guilds and channels
	router.RegisterRestAPIRoute("/v1/discord/guilds", web.AuthAccessUser, controller.GetAvailableGuilds, "GET")
	router.RegisterRestAPIRoute("/v1/discord/guilds/{id}/channels", web.AuthAccessUser, controller.GetGuildChannels, "GET")

	// Targets
	router.RegisterRestAPIRoute("/v1/discord/targets", web.AuthAccessUser, controller.GetMyTargets, "GET")
	router.RegisterRestAPIRoute("/v1/discord/targets", web.AuthAccessUser, controller.CreateTarget, "POST")
	router.RegisterRestAPIRoute("/v1/discord/targets/{id}", web.AuthAccessUser, controller.UpdateTarget, "PUT")
	router.RegisterRestAPIRoute("/v1/discord/targets/{id}", web.AuthAccessUser, controller.DeleteTarget, "DELETE")

	// Preferences
	router.RegisterRestAPIRoute("/v1/discord/targets/{id}/prefs", web.AuthAccessUser, controller.GetPreferences, "GET")
	router.RegisterRestAPIRoute("/v1/discord/targets/{id}/prefs", web.AuthAccessUser, controller.UpsertPreference, "POST")

	// Test
	router.RegisterRestAPIRoute("/v1/discord/targets/{id}/test", web.AuthAccessUser, controller.TestTarget, "POST")

	return controller
}

// GetMyLink returns the user's Discord link
func (c *DiscordNotifications) GetMyLink(args *web.HandlerArgs) (any, *web.HttpError) {
	link, err := c.repository.GetLinkByUser(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get discord link")}
	}

	if link == nil {
		return map[string]any{"linked": false}, nil
	}

	return map[string]any{
		"linked":          true,
		"discordUserId":   link.DiscordUserID,
		"discordUsername": link.DiscordUsername,
		"linkedAt":        link.CreatedAt,
	}, nil
}

type saveDiscordLinkRequest struct {
	DiscordUserID  string `json:"discordUserId"`
	DiscordUsername string `json:"discordUsername"`
	AccessToken    string `json:"accessToken"`
	RefreshToken   string `json:"refreshToken"`
	ExpiresIn      int    `json:"expiresIn"`
}

// SaveDiscordLink saves or updates the user's Discord link (called from frontend OAuth callback)
func (c *DiscordNotifications) SaveDiscordLink(args *web.HandlerArgs) (any, *web.HttpError) {
	var req saveDiscordLinkRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	if req.DiscordUserID == "" || req.AccessToken == "" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("discordUserId and accessToken are required")}
	}

	link := &models.DiscordLink{
		UserID:         *args.User,
		DiscordUserID:  req.DiscordUserID,
		DiscordUsername: req.DiscordUsername,
		AccessToken:    req.AccessToken,
		RefreshToken:   req.RefreshToken,
		TokenExpiresAt: time.Now().Add(time.Duration(req.ExpiresIn) * time.Second),
	}

	if err := c.repository.CreateLink(args.Request.Context(), link); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to save discord link")}
	}

	return map[string]any{"success": true}, nil
}

// UnlinkDiscord removes the user's Discord link and all associated targets/preferences
func (c *DiscordNotifications) UnlinkDiscord(args *web.HandlerArgs) (any, *web.HttpError) {
	if err := c.repository.DeleteLink(args.Request.Context(), *args.User); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to unlink discord")}
	}

	return map[string]any{"success": true}, nil
}

// GetAvailableGuilds returns guilds that both the user and bot are members of
func (c *DiscordNotifications) GetAvailableGuilds(args *web.HandlerArgs) (any, *web.HttpError) {
	link, err := c.repository.GetLinkByUser(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get discord link")}
	}
	if link == nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("discord account not linked")}
	}

	ctx := args.Request.Context()

	userGuilds, err := c.discordClient.GetUserGuilds(ctx, link.AccessToken)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get user guilds")}
	}

	botGuilds, err := c.discordClient.GetBotGuilds(ctx)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get bot guilds")}
	}

	// Intersection: only guilds where both user and bot are present
	botGuildMap := map[string]bool{}
	for _, g := range botGuilds {
		botGuildMap[g.ID] = true
	}

	available := []*models.DiscordGuild{}
	for _, g := range userGuilds {
		if botGuildMap[g.ID] {
			available = append(available, g)
		}
	}

	return available, nil
}

// GetGuildChannels returns text channels for a guild
func (c *DiscordNotifications) GetGuildChannels(args *web.HandlerArgs) (any, *web.HttpError) {
	guildID := args.Params["id"]
	if guildID == "" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("guild id is required")}
	}

	channels, err := c.discordClient.GetGuildChannels(args.Request.Context(), guildID)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get guild channels")}
	}

	return channels, nil
}

// GetMyTargets returns the user's notification targets
func (c *DiscordNotifications) GetMyTargets(args *web.HandlerArgs) (any, *web.HttpError) {
	targets, err := c.repository.GetTargetsByUser(args.Request.Context(), *args.User)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get targets")}
	}

	return targets, nil
}

type createTargetRequest struct {
	TargetType  string  `json:"targetType"`
	ChannelID   *string `json:"channelId"`
	GuildName   string  `json:"guildName"`
	ChannelName string  `json:"channelName"`
}

// CreateTarget creates a new notification target
func (c *DiscordNotifications) CreateTarget(args *web.HandlerArgs) (any, *web.HttpError) {
	var req createTargetRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	if req.TargetType != "dm" && req.TargetType != "channel" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("targetType must be 'dm' or 'channel'")}
	}

	if req.TargetType == "channel" && (req.ChannelID == nil || *req.ChannelID == "") {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("channelId is required for channel targets")}
	}

	target := &models.DiscordNotificationTarget{
		UserID:      *args.User,
		TargetType:  req.TargetType,
		ChannelID:   req.ChannelID,
		GuildName:   req.GuildName,
		ChannelName: req.ChannelName,
	}

	if err := c.repository.CreateTarget(args.Request.Context(), target); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to create target")}
	}

	return target, nil
}

type updateTargetRequest struct {
	IsActive bool `json:"isActive"`
}

// UpdateTarget updates a notification target
func (c *DiscordNotifications) UpdateTarget(args *web.HandlerArgs) (any, *web.HttpError) {
	id, err := parseID(args.Params["id"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid target id")}
	}

	var req updateTargetRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	target := &models.DiscordNotificationTarget{
		ID:       id,
		UserID:   *args.User,
		IsActive: req.IsActive,
	}

	if err := c.repository.UpdateTarget(args.Request.Context(), target); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to update target")}
	}

	return target, nil
}

// DeleteTarget deletes a notification target
func (c *DiscordNotifications) DeleteTarget(args *web.HandlerArgs) (any, *web.HttpError) {
	id, err := parseID(args.Params["id"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid target id")}
	}

	if err := c.repository.DeleteTarget(args.Request.Context(), id, *args.User); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to delete target")}
	}

	return map[string]any{"success": true}, nil
}

// GetPreferences returns preferences for a target
func (c *DiscordNotifications) GetPreferences(args *web.HandlerArgs) (any, *web.HttpError) {
	id, err := parseID(args.Params["id"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid target id")}
	}

	// Verify the target belongs to this user
	target, err := c.repository.GetTargetByID(args.Request.Context(), id)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.New("target not found")}
	}
	if target.UserID != *args.User {
		return nil, &web.HttpError{StatusCode: 403, Error: errors.New("unauthorized")}
	}

	prefs, err := c.repository.GetPreferencesByTarget(args.Request.Context(), id)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to get preferences")}
	}

	return prefs, nil
}

type upsertPreferenceRequest struct {
	EventType string `json:"eventType"`
	IsEnabled bool   `json:"isEnabled"`
}

// UpsertPreference creates or updates a preference for a target
func (c *DiscordNotifications) UpsertPreference(args *web.HandlerArgs) (any, *web.HttpError) {
	id, err := parseID(args.Params["id"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid target id")}
	}

	// Verify the target belongs to this user
	target, err := c.repository.GetTargetByID(args.Request.Context(), id)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.New("target not found")}
	}
	if target.UserID != *args.User {
		return nil, &web.HttpError{StatusCode: 403, Error: errors.New("unauthorized")}
	}

	var req upsertPreferenceRequest
	if err := json.NewDecoder(args.Request.Body).Decode(&req); err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.Wrap(err, "invalid request body")}
	}

	if req.EventType == "" {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("eventType is required")}
	}

	pref := &models.NotificationPreference{
		TargetID:  id,
		EventType: req.EventType,
		IsEnabled: req.IsEnabled,
	}

	if err := c.repository.UpsertPreference(args.Request.Context(), pref); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to upsert preference")}
	}

	return pref, nil
}

// TestTarget sends a test notification to verify a target works
func (c *DiscordNotifications) TestTarget(args *web.HandlerArgs) (any, *web.HttpError) {
	id, err := parseID(args.Params["id"])
	if err != nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("invalid target id")}
	}

	target, err := c.repository.GetTargetByID(args.Request.Context(), id)
	if err != nil {
		return nil, &web.HttpError{StatusCode: 404, Error: errors.New("target not found")}
	}
	if target.UserID != *args.User {
		return nil, &web.HttpError{StatusCode: 403, Error: errors.New("unauthorized")}
	}

	link, err := c.repository.GetLinkByUser(args.Request.Context(), *args.User)
	if err != nil || link == nil {
		return nil, &web.HttpError{StatusCode: 400, Error: errors.New("discord account not linked")}
	}

	if err := c.notifier.SendTestNotification(args.Request.Context(), target, link); err != nil {
		return nil, &web.HttpError{StatusCode: 500, Error: errors.Wrap(err, "failed to send test notification")}
	}

	return map[string]any{"success": true}, nil
}
