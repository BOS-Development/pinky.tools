package updaters_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/updaters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHaulingNotificationsRepo mocks the NotificationsDiscordRepo interface.
type MockHaulingNotificationsRepo struct {
	mock.Mock
}

func (m *MockHaulingNotificationsRepo) GetActiveTargetsForEvent(ctx context.Context, userID int64, eventType string) ([]*models.DiscordNotificationTarget, error) {
	args := m.Called(ctx, userID, eventType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DiscordNotificationTarget), args.Error(1)
}

func (m *MockHaulingNotificationsRepo) GetLinkByUser(ctx context.Context, userID int64) (*models.DiscordLink, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DiscordLink), args.Error(1)
}

// MockHaulingDiscordClient mocks the DiscordClientInterface.
type MockHaulingDiscordClient struct {
	mock.Mock
}

func (m *MockHaulingDiscordClient) SendDM(ctx context.Context, discordUserID string, embed *client.DiscordEmbed) error {
	args := m.Called(ctx, discordUserID, embed)
	return args.Error(0)
}

func (m *MockHaulingDiscordClient) SendChannelMessage(ctx context.Context, channelID string, embed *client.DiscordEmbed) error {
	args := m.Called(ctx, channelID, embed)
	return args.Error(0)
}

func (m *MockHaulingDiscordClient) GetUserGuilds(ctx context.Context, userAccessToken string) ([]*models.DiscordGuild, error) {
	args := m.Called(ctx, userAccessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DiscordGuild), args.Error(1)
}

func (m *MockHaulingDiscordClient) GetBotGuilds(ctx context.Context) ([]*models.DiscordGuild, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DiscordGuild), args.Error(1)
}

func (m *MockHaulingDiscordClient) GetGuildChannels(ctx context.Context, guildID string) ([]*models.DiscordChannel, error) {
	args := m.Called(ctx, guildID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DiscordChannel), args.Error(1)
}

func setupHaulingNotifier() (*updaters.HaulingNotificationsUpdater, *MockHaulingNotificationsRepo, *MockHaulingDiscordClient) {
	repo := new(MockHaulingNotificationsRepo)
	discord := new(MockHaulingDiscordClient)
	notifier := updaters.NewHaulingNotifications(repo, discord, "https://example.com/")
	return notifier, repo, discord
}

func makeHaulingRun() *models.HaulingRun {
	threshold := 5000000.0
	return &models.HaulingRun{
		ID:               int64(1),
		UserID:           int64(100),
		Name:             "Test Run",
		Status:           "ACCUMULATING",
		FromRegionID:     int64(10000002),
		ToRegionID:       int64(10000043),
		NotifyTier2:      true,
		HaulThresholdISK: &threshold,
	}
}

// --- NotifyHaulingTier2 tests ---

func Test_HaulingNotifications_NotifyHaulingTier2_NoTargets(t *testing.T) {
	notifier, repo, discord := setupHaulingNotifier()
	userID := int64(100)
	run := makeHaulingRun()

	repo.On("GetActiveTargetsForEvent", mock.Anything, userID, "hauling_tier2").
		Return([]*models.DiscordNotificationTarget{}, nil)

	notifier.NotifyHaulingTier2(context.Background(), userID, run, 85.0)

	repo.AssertExpectations(t)
	discord.AssertNotCalled(t, "SendDM", mock.Anything, mock.Anything, mock.Anything)
	discord.AssertNotCalled(t, "SendChannelMessage", mock.Anything, mock.Anything, mock.Anything)
}

func Test_HaulingNotifications_NotifyHaulingTier2_SendsDM(t *testing.T) {
	notifier, repo, discord := setupHaulingNotifier()
	userID := int64(100)
	run := makeHaulingRun()

	discordLink := &models.DiscordLink{
		UserID:        userID,
		DiscordUserID: "discord-123",
	}
	target := &models.DiscordNotificationTarget{
		ID:         int64(1),
		UserID:     userID,
		TargetType: "dm",
	}

	repo.On("GetActiveTargetsForEvent", mock.Anything, userID, "hauling_tier2").
		Return([]*models.DiscordNotificationTarget{target}, nil)
	repo.On("GetLinkByUser", mock.Anything, userID).Return(discordLink, nil)
	discord.On("SendDM", mock.Anything, "discord-123", mock.AnythingOfType("*client.DiscordEmbed")).Return(nil)

	notifier.NotifyHaulingTier2(context.Background(), userID, run, 85.0)

	repo.AssertExpectations(t)
	discord.AssertExpectations(t)
}

func Test_HaulingNotifications_NotifyHaulingTier2_SendsChannelMessage(t *testing.T) {
	notifier, repo, discord := setupHaulingNotifier()
	userID := int64(100)
	run := makeHaulingRun()
	channelID := "channel-456"

	target := &models.DiscordNotificationTarget{
		ID:         int64(2),
		UserID:     userID,
		TargetType: "channel",
		ChannelID:  &channelID,
	}

	repo.On("GetActiveTargetsForEvent", mock.Anything, userID, "hauling_tier2").
		Return([]*models.DiscordNotificationTarget{target}, nil)
	discord.On("SendChannelMessage", mock.Anything, channelID, mock.AnythingOfType("*client.DiscordEmbed")).Return(nil)

	notifier.NotifyHaulingTier2(context.Background(), userID, run, 85.0)

	repo.AssertExpectations(t)
	discord.AssertExpectations(t)
}

func Test_HaulingNotifications_NotifyHaulingTier2_RepoError(t *testing.T) {
	notifier, repo, discord := setupHaulingNotifier()
	userID := int64(100)
	run := makeHaulingRun()

	repo.On("GetActiveTargetsForEvent", mock.Anything, userID, "hauling_tier2").
		Return(nil, assert.AnError)

	// Should not panic; error is logged
	notifier.NotifyHaulingTier2(context.Background(), userID, run, 85.0)

	repo.AssertExpectations(t)
	discord.AssertNotCalled(t, "SendDM", mock.Anything, mock.Anything, mock.Anything)
}

// --- NotifyHaulingComplete tests ---

func Test_HaulingNotifications_NotifyHaulingComplete_NoTargets(t *testing.T) {
	notifier, repo, discord := setupHaulingNotifier()
	userID := int64(100)
	run := makeHaulingRun()
	summary := &models.HaulingRunPnlSummary{
		TotalRevenueISK: 150000.0,
		TotalCostISK:    100000.0,
		NetProfitISK:    50000.0,
		MarginPct:       33.33,
	}

	repo.On("GetActiveTargetsForEvent", mock.Anything, userID, "hauling_complete").
		Return([]*models.DiscordNotificationTarget{}, nil)

	notifier.NotifyHaulingComplete(context.Background(), userID, run, summary)

	repo.AssertExpectations(t)
	discord.AssertNotCalled(t, "SendDM", mock.Anything, mock.Anything, mock.Anything)
}

func Test_HaulingNotifications_NotifyHaulingComplete_WithSummary(t *testing.T) {
	notifier, repo, discord := setupHaulingNotifier()
	userID := int64(100)
	run := makeHaulingRun()
	summary := &models.HaulingRunPnlSummary{
		TotalRevenueISK: 150000.0,
		TotalCostISK:    100000.0,
		NetProfitISK:    50000.0,
		MarginPct:       33.33,
	}
	channelID := "channel-789"

	target := &models.DiscordNotificationTarget{
		ID:         int64(3),
		UserID:     userID,
		TargetType: "channel",
		ChannelID:  &channelID,
	}

	repo.On("GetActiveTargetsForEvent", mock.Anything, userID, "hauling_complete").
		Return([]*models.DiscordNotificationTarget{target}, nil)

	var capturedEmbed *client.DiscordEmbed
	discord.On("SendChannelMessage", mock.Anything, channelID, mock.AnythingOfType("*client.DiscordEmbed")).
		Run(func(args mock.Arguments) {
			capturedEmbed = args.Get(2).(*client.DiscordEmbed)
		}).
		Return(nil)

	notifier.NotifyHaulingComplete(context.Background(), userID, run, summary)

	repo.AssertExpectations(t)
	discord.AssertExpectations(t)
	assert.NotNil(t, capturedEmbed)
	assert.Equal(t, "Hauling Run Complete", capturedEmbed.Title)
	assert.Contains(t, capturedEmbed.Description, "50000.00 ISK")
}

func Test_HaulingNotifications_NotifyHaulingComplete_NilSummary(t *testing.T) {
	notifier, repo, discord := setupHaulingNotifier()
	userID := int64(100)
	run := makeHaulingRun()
	channelID := "channel-999"

	target := &models.DiscordNotificationTarget{
		ID:         int64(4),
		UserID:     userID,
		TargetType: "channel",
		ChannelID:  &channelID,
	}

	repo.On("GetActiveTargetsForEvent", mock.Anything, userID, "hauling_complete").
		Return([]*models.DiscordNotificationTarget{target}, nil)
	discord.On("SendChannelMessage", mock.Anything, channelID, mock.AnythingOfType("*client.DiscordEmbed")).Return(nil)

	// nil summary should not panic
	notifier.NotifyHaulingComplete(context.Background(), userID, run, nil)

	repo.AssertExpectations(t)
	discord.AssertExpectations(t)
}

// --- SendHaulingDailyDigest tests ---

func Test_HaulingNotifications_SendHaulingDailyDigest_NoRuns(t *testing.T) {
	notifier, repo, discord := setupHaulingNotifier()
	userID := int64(100)

	// Should return early without querying targets
	notifier.SendHaulingDailyDigest(context.Background(), userID, []*models.HaulingRun{})

	repo.AssertNotCalled(t, "GetActiveTargetsForEvent", mock.Anything, mock.Anything, mock.Anything)
	discord.AssertNotCalled(t, "SendDM", mock.Anything, mock.Anything, mock.Anything)
}

func Test_HaulingNotifications_SendHaulingDailyDigest_WithRuns(t *testing.T) {
	notifier, repo, discord := setupHaulingNotifier()
	userID := int64(100)
	runs := []*models.HaulingRun{
		makeHaulingRun(),
		{ID: int64(2), Name: "Run 2", Status: "ACCUMULATING", FromRegionID: int64(10000030), ToRegionID: int64(10000042)},
	}
	channelID := "channel-digest"

	target := &models.DiscordNotificationTarget{
		ID:         int64(5),
		UserID:     userID,
		TargetType: "channel",
		ChannelID:  &channelID,
	}

	repo.On("GetActiveTargetsForEvent", mock.Anything, userID, "hauling_daily_digest").
		Return([]*models.DiscordNotificationTarget{target}, nil)

	var capturedEmbed *client.DiscordEmbed
	discord.On("SendChannelMessage", mock.Anything, channelID, mock.AnythingOfType("*client.DiscordEmbed")).
		Run(func(args mock.Arguments) {
			capturedEmbed = args.Get(2).(*client.DiscordEmbed)
		}).
		Return(nil)

	notifier.SendHaulingDailyDigest(context.Background(), userID, runs)

	repo.AssertExpectations(t)
	discord.AssertExpectations(t)
	assert.NotNil(t, capturedEmbed)
	assert.Equal(t, "Hauling Daily Digest", capturedEmbed.Title)
	assert.Len(t, capturedEmbed.Fields, 2)
}

// --- Region name mapping tests ---

func Test_HaulingNotifications_EmbedContainsRegionNames(t *testing.T) {
	notifier, repo, discord := setupHaulingNotifier()
	userID := int64(100)
	run := &models.HaulingRun{
		ID:           int64(1),
		Name:         "Jita to Amarr",
		Status:       "ACCUMULATING",
		FromRegionID: int64(10000002), // The Forge
		ToRegionID:   int64(10000043), // Domain
		NotifyTier2:  true,
	}
	channelID := "channel-route"

	target := &models.DiscordNotificationTarget{
		ID:         int64(6),
		UserID:     userID,
		TargetType: "channel",
		ChannelID:  &channelID,
	}

	repo.On("GetActiveTargetsForEvent", mock.Anything, userID, "hauling_tier2").
		Return([]*models.DiscordNotificationTarget{target}, nil)

	var capturedEmbed *client.DiscordEmbed
	discord.On("SendChannelMessage", mock.Anything, channelID, mock.AnythingOfType("*client.DiscordEmbed")).
		Run(func(args mock.Arguments) {
			capturedEmbed = args.Get(2).(*client.DiscordEmbed)
		}).
		Return(nil)

	notifier.NotifyHaulingTier2(context.Background(), userID, run, 90.0)

	discord.AssertExpectations(t)
	assert.NotNil(t, capturedEmbed)
	// Should contain region names
	found := false
	for _, f := range capturedEmbed.Fields {
		if f.Name == "Route" {
			assert.Contains(t, f.Value, "The Forge")
			assert.Contains(t, f.Value, "Domain")
			found = true
		}
	}
	assert.True(t, found, "Route field not found in embed")
}
