package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/annymsMthd/industry-tool/internal/controllers"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDiscordNotificationsRepository is a mock implementation of DiscordNotificationsRepository
type MockDiscordNotificationsRepository struct {
	mock.Mock
}

func (m *MockDiscordNotificationsRepository) CreateLink(ctx context.Context, link *models.DiscordLink) error {
	args := m.Called(ctx, link)
	return args.Error(0)
}

func (m *MockDiscordNotificationsRepository) GetLinkByUser(ctx context.Context, userID int64) (*models.DiscordLink, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DiscordLink), args.Error(1)
}

func (m *MockDiscordNotificationsRepository) UpdateLinkTokens(ctx context.Context, userID int64, accessToken, refreshToken string, expiresAt time.Time) error {
	args := m.Called(ctx, userID, accessToken, refreshToken, expiresAt)
	return args.Error(0)
}

func (m *MockDiscordNotificationsRepository) DeleteLink(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockDiscordNotificationsRepository) CreateTarget(ctx context.Context, target *models.DiscordNotificationTarget) error {
	args := m.Called(ctx, target)
	return args.Error(0)
}

func (m *MockDiscordNotificationsRepository) GetTargetsByUser(ctx context.Context, userID int64) ([]*models.DiscordNotificationTarget, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DiscordNotificationTarget), args.Error(1)
}

func (m *MockDiscordNotificationsRepository) GetTargetByID(ctx context.Context, id int64) (*models.DiscordNotificationTarget, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DiscordNotificationTarget), args.Error(1)
}

func (m *MockDiscordNotificationsRepository) UpdateTarget(ctx context.Context, target *models.DiscordNotificationTarget) error {
	args := m.Called(ctx, target)
	return args.Error(0)
}

func (m *MockDiscordNotificationsRepository) DeleteTarget(ctx context.Context, id int64, userID int64) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockDiscordNotificationsRepository) UpsertPreference(ctx context.Context, pref *models.NotificationPreference) error {
	args := m.Called(ctx, pref)
	return args.Error(0)
}

func (m *MockDiscordNotificationsRepository) GetPreferencesByTarget(ctx context.Context, targetID int64) ([]*models.NotificationPreference, error) {
	args := m.Called(ctx, targetID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.NotificationPreference), args.Error(1)
}

// MockDiscordClientProvider is a mock implementation of DiscordClientProvider
type MockDiscordClientProvider struct {
	mock.Mock
}

func (m *MockDiscordClientProvider) GetUserGuilds(ctx context.Context, userAccessToken string) ([]*models.DiscordGuild, error) {
	args := m.Called(ctx, userAccessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DiscordGuild), args.Error(1)
}

func (m *MockDiscordClientProvider) GetBotGuilds(ctx context.Context) ([]*models.DiscordGuild, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DiscordGuild), args.Error(1)
}

func (m *MockDiscordClientProvider) GetGuildChannels(ctx context.Context, guildID string) ([]*models.DiscordChannel, error) {
	args := m.Called(ctx, guildID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DiscordChannel), args.Error(1)
}

func (m *MockDiscordClientProvider) SendDM(ctx context.Context, discordUserID string, embed *client.DiscordEmbed) error {
	args := m.Called(ctx, discordUserID, embed)
	return args.Error(0)
}

func (m *MockDiscordClientProvider) SendChannelMessage(ctx context.Context, channelID string, embed *client.DiscordEmbed) error {
	args := m.Called(ctx, channelID, embed)
	return args.Error(0)
}

// MockDiscordTestNotifier is a mock implementation of DiscordTestNotifier
type MockDiscordTestNotifier struct {
	mock.Mock
}

func (m *MockDiscordTestNotifier) SendTestNotification(ctx context.Context, target *models.DiscordNotificationTarget, discordLink *models.DiscordLink) error {
	args := m.Called(ctx, target, discordLink)
	return args.Error(0)
}

// --- GetMyLink tests ---

func Test_DiscordNotifications_GetMyLink_Success_Linked(t *testing.T) {
	mockRepo := new(MockDiscordNotificationsRepository)
	mockDiscord := new(MockDiscordClientProvider)
	mockNotifier := new(MockDiscordTestNotifier)
	mockRouter := &MockRouter{}

	userID := int64(123)
	expectedLink := &models.DiscordLink{
		ID:             1,
		UserID:         userID,
		DiscordUserID:  "discord-user-123",
		DiscordUsername: "TestUser#1234",
		CreatedAt:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	mockRepo.On("GetLinkByUser", mock.Anything, userID).Return(expectedLink, nil)

	req := httptest.NewRequest("GET", "/v1/discord/link", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewDiscordNotifications(mockRouter, mockRepo, mockDiscord, mockNotifier)
	result, httpErr := controller.GetMyLink(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	resultMap := result.(map[string]any)
	assert.Equal(t, true, resultMap["linked"])
	assert.Equal(t, "discord-user-123", resultMap["discordUserId"])
	assert.Equal(t, "TestUser#1234", resultMap["discordUsername"])
	assert.Equal(t, expectedLink.CreatedAt, resultMap["linkedAt"])

	mockRepo.AssertExpectations(t)
}

func Test_DiscordNotifications_GetMyLink_Success_NotLinked(t *testing.T) {
	mockRepo := new(MockDiscordNotificationsRepository)
	mockDiscord := new(MockDiscordClientProvider)
	mockNotifier := new(MockDiscordTestNotifier)
	mockRouter := &MockRouter{}

	userID := int64(123)

	mockRepo.On("GetLinkByUser", mock.Anything, userID).Return(nil, nil)

	req := httptest.NewRequest("GET", "/v1/discord/link", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewDiscordNotifications(mockRouter, mockRepo, mockDiscord, mockNotifier)
	result, httpErr := controller.GetMyLink(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	resultMap := result.(map[string]any)
	assert.Equal(t, false, resultMap["linked"])

	mockRepo.AssertExpectations(t)
}

// --- SaveDiscordLink tests ---

func Test_DiscordNotifications_SaveDiscordLink_Success(t *testing.T) {
	mockRepo := new(MockDiscordNotificationsRepository)
	mockDiscord := new(MockDiscordClientProvider)
	mockNotifier := new(MockDiscordTestNotifier)
	mockRouter := &MockRouter{}

	userID := int64(123)

	mockRepo.On("CreateLink", mock.Anything, mock.MatchedBy(func(link *models.DiscordLink) bool {
		return link.UserID == userID &&
			link.DiscordUserID == "discord-user-456" &&
			link.DiscordUsername == "TestUser#5678" &&
			link.AccessToken == "access-token-abc" &&
			link.RefreshToken == "refresh-token-xyz"
	})).Return(nil)

	body := map[string]any{
		"discordUserId":  "discord-user-456",
		"discordUsername": "TestUser#5678",
		"accessToken":    "access-token-abc",
		"refreshToken":   "refresh-token-xyz",
		"expiresIn":      3600,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/discord/link", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewDiscordNotifications(mockRouter, mockRepo, mockDiscord, mockNotifier)
	result, httpErr := controller.SaveDiscordLink(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	resultMap := result.(map[string]any)
	assert.Equal(t, true, resultMap["success"])

	mockRepo.AssertExpectations(t)
}

func Test_DiscordNotifications_SaveDiscordLink_MissingFields_Returns400(t *testing.T) {
	mockRepo := new(MockDiscordNotificationsRepository)
	mockDiscord := new(MockDiscordClientProvider)
	mockNotifier := new(MockDiscordTestNotifier)
	mockRouter := &MockRouter{}

	userID := int64(123)

	// Missing discordUserId and accessToken
	body := map[string]any{
		"discordUsername": "TestUser#5678",
		"refreshToken":   "refresh-token-xyz",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/discord/link", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewDiscordNotifications(mockRouter, mockRepo, mockDiscord, mockNotifier)
	result, httpErr := controller.SaveDiscordLink(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

// --- UnlinkDiscord tests ---

func Test_DiscordNotifications_UnlinkDiscord_Success(t *testing.T) {
	mockRepo := new(MockDiscordNotificationsRepository)
	mockDiscord := new(MockDiscordClientProvider)
	mockNotifier := new(MockDiscordTestNotifier)
	mockRouter := &MockRouter{}

	userID := int64(123)

	mockRepo.On("DeleteLink", mock.Anything, userID).Return(nil)

	req := httptest.NewRequest("DELETE", "/v1/discord/link", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewDiscordNotifications(mockRouter, mockRepo, mockDiscord, mockNotifier)
	result, httpErr := controller.UnlinkDiscord(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	resultMap := result.(map[string]any)
	assert.Equal(t, true, resultMap["success"])

	mockRepo.AssertExpectations(t)
}

// --- GetAvailableGuilds tests ---

func Test_DiscordNotifications_GetAvailableGuilds_ReturnsIntersection(t *testing.T) {
	mockRepo := new(MockDiscordNotificationsRepository)
	mockDiscord := new(MockDiscordClientProvider)
	mockNotifier := new(MockDiscordTestNotifier)
	mockRouter := &MockRouter{}

	userID := int64(123)
	link := &models.DiscordLink{
		ID:          1,
		UserID:      userID,
		AccessToken: "user-access-token",
	}

	userGuilds := []*models.DiscordGuild{
		{ID: "guild-1", Name: "Shared Guild", Icon: "icon1"},
		{ID: "guild-2", Name: "User Only Guild", Icon: "icon2"},
		{ID: "guild-3", Name: "Another Shared Guild", Icon: "icon3"},
	}

	botGuilds := []*models.DiscordGuild{
		{ID: "guild-1", Name: "Shared Guild", Icon: "icon1"},
		{ID: "guild-3", Name: "Another Shared Guild", Icon: "icon3"},
		{ID: "guild-4", Name: "Bot Only Guild", Icon: "icon4"},
	}

	mockRepo.On("GetLinkByUser", mock.Anything, userID).Return(link, nil)
	mockDiscord.On("GetUserGuilds", mock.Anything, "user-access-token").Return(userGuilds, nil)
	mockDiscord.On("GetBotGuilds", mock.Anything).Return(botGuilds, nil)

	req := httptest.NewRequest("GET", "/v1/discord/guilds", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewDiscordNotifications(mockRouter, mockRepo, mockDiscord, mockNotifier)
	result, httpErr := controller.GetAvailableGuilds(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	guilds := result.([]*models.DiscordGuild)
	assert.Len(t, guilds, 2)
	assert.Equal(t, "guild-1", guilds[0].ID)
	assert.Equal(t, "Shared Guild", guilds[0].Name)
	assert.Equal(t, "guild-3", guilds[1].ID)
	assert.Equal(t, "Another Shared Guild", guilds[1].Name)

	mockRepo.AssertExpectations(t)
	mockDiscord.AssertExpectations(t)
}

// --- GetMyTargets tests ---

func Test_DiscordNotifications_GetMyTargets_Success(t *testing.T) {
	mockRepo := new(MockDiscordNotificationsRepository)
	mockDiscord := new(MockDiscordClientProvider)
	mockNotifier := new(MockDiscordTestNotifier)
	mockRouter := &MockRouter{}

	userID := int64(123)
	channelID := "channel-abc"
	expectedTargets := []*models.DiscordNotificationTarget{
		{
			ID:          1,
			UserID:      userID,
			TargetType:  "dm",
			IsActive:    true,
			GuildName:   "",
			ChannelName: "",
		},
		{
			ID:          2,
			UserID:      userID,
			TargetType:  "channel",
			ChannelID:   &channelID,
			GuildName:   "Test Guild",
			ChannelName: "general",
			IsActive:    true,
		},
	}

	mockRepo.On("GetTargetsByUser", mock.Anything, userID).Return(expectedTargets, nil)

	req := httptest.NewRequest("GET", "/v1/discord/targets", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewDiscordNotifications(mockRouter, mockRepo, mockDiscord, mockNotifier)
	result, httpErr := controller.GetMyTargets(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	targets := result.([]*models.DiscordNotificationTarget)
	assert.Len(t, targets, 2)
	assert.Equal(t, "dm", targets[0].TargetType)
	assert.Equal(t, "channel", targets[1].TargetType)
	assert.Equal(t, &channelID, targets[1].ChannelID)

	mockRepo.AssertExpectations(t)
}

// --- CreateTarget tests ---

func Test_DiscordNotifications_CreateTarget_DM_Success(t *testing.T) {
	mockRepo := new(MockDiscordNotificationsRepository)
	mockDiscord := new(MockDiscordClientProvider)
	mockNotifier := new(MockDiscordTestNotifier)
	mockRouter := &MockRouter{}

	userID := int64(123)

	mockRepo.On("CreateTarget", mock.Anything, mock.MatchedBy(func(target *models.DiscordNotificationTarget) bool {
		return target.UserID == userID && target.TargetType == "dm"
	})).Return(nil)

	body := map[string]any{
		"targetType": "dm",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/discord/targets", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewDiscordNotifications(mockRouter, mockRepo, mockDiscord, mockNotifier)
	result, httpErr := controller.CreateTarget(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	target := result.(*models.DiscordNotificationTarget)
	assert.Equal(t, userID, target.UserID)
	assert.Equal(t, "dm", target.TargetType)

	mockRepo.AssertExpectations(t)
}

func Test_DiscordNotifications_CreateTarget_Channel_Success(t *testing.T) {
	mockRepo := new(MockDiscordNotificationsRepository)
	mockDiscord := new(MockDiscordClientProvider)
	mockNotifier := new(MockDiscordTestNotifier)
	mockRouter := &MockRouter{}

	userID := int64(123)
	channelID := "channel-abc"

	mockRepo.On("CreateTarget", mock.Anything, mock.MatchedBy(func(target *models.DiscordNotificationTarget) bool {
		return target.UserID == userID &&
			target.TargetType == "channel" &&
			target.ChannelID != nil &&
			*target.ChannelID == channelID &&
			target.GuildName == "Test Guild" &&
			target.ChannelName == "general"
	})).Return(nil)

	body := map[string]any{
		"targetType":  "channel",
		"channelId":   channelID,
		"guildName":   "Test Guild",
		"channelName": "general",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/discord/targets", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewDiscordNotifications(mockRouter, mockRepo, mockDiscord, mockNotifier)
	result, httpErr := controller.CreateTarget(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	target := result.(*models.DiscordNotificationTarget)
	assert.Equal(t, userID, target.UserID)
	assert.Equal(t, "channel", target.TargetType)
	assert.Equal(t, &channelID, target.ChannelID)
	assert.Equal(t, "Test Guild", target.GuildName)
	assert.Equal(t, "general", target.ChannelName)

	mockRepo.AssertExpectations(t)
}

func Test_DiscordNotifications_CreateTarget_InvalidType_Returns400(t *testing.T) {
	mockRepo := new(MockDiscordNotificationsRepository)
	mockDiscord := new(MockDiscordClientProvider)
	mockNotifier := new(MockDiscordTestNotifier)
	mockRouter := &MockRouter{}

	userID := int64(123)

	body := map[string]any{
		"targetType": "invalid",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/discord/targets", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{},
	}

	controller := controllers.NewDiscordNotifications(mockRouter, mockRepo, mockDiscord, mockNotifier)
	result, httpErr := controller.CreateTarget(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 400, httpErr.StatusCode)
}

// --- DeleteTarget tests ---

func Test_DiscordNotifications_DeleteTarget_Success(t *testing.T) {
	mockRepo := new(MockDiscordNotificationsRepository)
	mockDiscord := new(MockDiscordClientProvider)
	mockNotifier := new(MockDiscordTestNotifier)
	mockRouter := &MockRouter{}

	userID := int64(123)
	targetID := int64(42)

	mockRepo.On("DeleteTarget", mock.Anything, targetID, userID).Return(nil)

	req := httptest.NewRequest("DELETE", "/v1/discord/targets/42", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "42"},
	}

	controller := controllers.NewDiscordNotifications(mockRouter, mockRepo, mockDiscord, mockNotifier)
	result, httpErr := controller.DeleteTarget(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	resultMap := result.(map[string]any)
	assert.Equal(t, true, resultMap["success"])

	mockRepo.AssertExpectations(t)
}

// --- GetPreferences tests ---

func Test_DiscordNotifications_GetPreferences_Success(t *testing.T) {
	mockRepo := new(MockDiscordNotificationsRepository)
	mockDiscord := new(MockDiscordClientProvider)
	mockNotifier := new(MockDiscordTestNotifier)
	mockRouter := &MockRouter{}

	userID := int64(123)
	targetID := int64(10)

	target := &models.DiscordNotificationTarget{
		ID:     targetID,
		UserID: userID,
	}

	expectedPrefs := []*models.NotificationPreference{
		{ID: 1, TargetID: targetID, EventType: "asset_change", IsEnabled: true},
		{ID: 2, TargetID: targetID, EventType: "price_alert", IsEnabled: false},
	}

	mockRepo.On("GetTargetByID", mock.Anything, targetID).Return(target, nil)
	mockRepo.On("GetPreferencesByTarget", mock.Anything, targetID).Return(expectedPrefs, nil)

	req := httptest.NewRequest("GET", "/v1/discord/targets/10/prefs", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "10"},
	}

	controller := controllers.NewDiscordNotifications(mockRouter, mockRepo, mockDiscord, mockNotifier)
	result, httpErr := controller.GetPreferences(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	prefs := result.([]*models.NotificationPreference)
	assert.Len(t, prefs, 2)
	assert.Equal(t, "asset_change", prefs[0].EventType)
	assert.Equal(t, true, prefs[0].IsEnabled)
	assert.Equal(t, "price_alert", prefs[1].EventType)
	assert.Equal(t, false, prefs[1].IsEnabled)

	mockRepo.AssertExpectations(t)
}

func Test_DiscordNotifications_GetPreferences_Unauthorized_Returns403(t *testing.T) {
	mockRepo := new(MockDiscordNotificationsRepository)
	mockDiscord := new(MockDiscordClientProvider)
	mockNotifier := new(MockDiscordTestNotifier)
	mockRouter := &MockRouter{}

	userID := int64(123)
	otherUserID := int64(999)
	targetID := int64(10)

	// Target belongs to a different user
	target := &models.DiscordNotificationTarget{
		ID:     targetID,
		UserID: otherUserID,
	}

	mockRepo.On("GetTargetByID", mock.Anything, targetID).Return(target, nil)

	req := httptest.NewRequest("GET", "/v1/discord/targets/10/prefs", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "10"},
	}

	controller := controllers.NewDiscordNotifications(mockRouter, mockRepo, mockDiscord, mockNotifier)
	result, httpErr := controller.GetPreferences(args)

	assert.Nil(t, result)
	assert.NotNil(t, httpErr)
	assert.Equal(t, 403, httpErr.StatusCode)

	mockRepo.AssertExpectations(t)
}

// --- UpsertPreference tests ---

func Test_DiscordNotifications_UpsertPreference_Success(t *testing.T) {
	mockRepo := new(MockDiscordNotificationsRepository)
	mockDiscord := new(MockDiscordClientProvider)
	mockNotifier := new(MockDiscordTestNotifier)
	mockRouter := &MockRouter{}

	userID := int64(123)
	targetID := int64(10)

	target := &models.DiscordNotificationTarget{
		ID:     targetID,
		UserID: userID,
	}

	mockRepo.On("GetTargetByID", mock.Anything, targetID).Return(target, nil)
	mockRepo.On("UpsertPreference", mock.Anything, mock.MatchedBy(func(pref *models.NotificationPreference) bool {
		return pref.TargetID == targetID &&
			pref.EventType == "asset_change" &&
			pref.IsEnabled == true
	})).Return(nil)

	body := map[string]any{
		"eventType": "asset_change",
		"isEnabled": true,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/v1/discord/targets/10/prefs", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "10"},
	}

	controller := controllers.NewDiscordNotifications(mockRouter, mockRepo, mockDiscord, mockNotifier)
	result, httpErr := controller.UpsertPreference(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	pref := result.(*models.NotificationPreference)
	assert.Equal(t, targetID, pref.TargetID)
	assert.Equal(t, "asset_change", pref.EventType)
	assert.Equal(t, true, pref.IsEnabled)

	mockRepo.AssertExpectations(t)
}

// --- TestTarget tests ---

func Test_DiscordNotifications_TestTarget_Success(t *testing.T) {
	mockRepo := new(MockDiscordNotificationsRepository)
	mockDiscord := new(MockDiscordClientProvider)
	mockNotifier := new(MockDiscordTestNotifier)
	mockRouter := &MockRouter{}

	userID := int64(123)
	targetID := int64(10)

	target := &models.DiscordNotificationTarget{
		ID:         targetID,
		UserID:     userID,
		TargetType: "dm",
		IsActive:   true,
	}

	link := &models.DiscordLink{
		ID:            1,
		UserID:        userID,
		DiscordUserID: "discord-user-123",
		AccessToken:   "access-token",
	}

	mockRepo.On("GetTargetByID", mock.Anything, targetID).Return(target, nil)
	mockRepo.On("GetLinkByUser", mock.Anything, userID).Return(link, nil)
	mockNotifier.On("SendTestNotification", mock.Anything, target, link).Return(nil)

	req := httptest.NewRequest("POST", "/v1/discord/targets/10/test", nil)
	args := &web.HandlerArgs{
		Request: req,
		User:    &userID,
		Params:  map[string]string{"id": "10"},
	}

	controller := controllers.NewDiscordNotifications(mockRouter, mockRepo, mockDiscord, mockNotifier)
	result, httpErr := controller.TestTarget(args)

	assert.Nil(t, httpErr)
	assert.NotNil(t, result)

	resultMap := result.(map[string]any)
	assert.Equal(t, true, resultMap["success"])

	mockRepo.AssertExpectations(t)
	mockNotifier.AssertExpectations(t)
}
