package updaters_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/updaters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Discord repo
type MockNotificationsDiscordRepo struct {
	mock.Mock
}

func (m *MockNotificationsDiscordRepo) GetActiveTargetsForEvent(ctx context.Context, userID int64, eventType string) ([]*models.DiscordNotificationTarget, error) {
	args := m.Called(ctx, userID, eventType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DiscordNotificationTarget), args.Error(1)
}

func (m *MockNotificationsDiscordRepo) GetLinkByUser(ctx context.Context, userID int64) (*models.DiscordLink, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DiscordLink), args.Error(1)
}

// Mock Discord client
type MockDiscordClient struct {
	mock.Mock
}

func (m *MockDiscordClient) SendDM(ctx context.Context, discordUserID string, embed *client.DiscordEmbed) error {
	args := m.Called(ctx, discordUserID, embed)
	return args.Error(0)
}

func (m *MockDiscordClient) SendChannelMessage(ctx context.Context, channelID string, embed *client.DiscordEmbed) error {
	args := m.Called(ctx, channelID, embed)
	return args.Error(0)
}

func (m *MockDiscordClient) GetUserGuilds(ctx context.Context, userAccessToken string) ([]*models.DiscordGuild, error) {
	args := m.Called(ctx, userAccessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DiscordGuild), args.Error(1)
}

func (m *MockDiscordClient) GetBotGuilds(ctx context.Context) ([]*models.DiscordGuild, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DiscordGuild), args.Error(1)
}

func (m *MockDiscordClient) GetGuildChannels(ctx context.Context, guildID string) ([]*models.DiscordChannel, error) {
	args := m.Called(ctx, guildID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DiscordChannel), args.Error(1)
}

func Test_NotifyPurchase_SendsDM(t *testing.T) {
	mockRepo := new(MockNotificationsDiscordRepo)
	mockClient := new(MockDiscordClient)

	notifier := updaters.NewNotifications(mockRepo, mockClient, "")

	channelID := "dm-target"
	purchase := &models.PurchaseTransaction{
		ID:                1,
		BuyerUserID:       100,
		BuyerName:         "Alice",
		SellerUserID:      200,
		TypeID:            34,
		TypeName:          "Tritanium",
		LocationName:      "Jita IV - Moon 4",
		QuantityPurchased: 1000,
		PricePerUnit:      5.0,
		TotalPrice:        5000.0,
		PurchasedAt:       time.Now(),
	}

	targets := []*models.DiscordNotificationTarget{
		{
			ID:         1,
			UserID:     200,
			TargetType: "dm",
			ChannelID:  &channelID,
			IsActive:   true,
		},
	}

	link := &models.DiscordLink{
		ID:            1,
		UserID:        200,
		DiscordUserID: "discord-user-200",
	}

	mockRepo.On("GetActiveTargetsForEvent", mock.Anything, int64(200), "purchase_created").Return(targets, nil)
	mockRepo.On("GetLinkByUser", mock.Anything, int64(200)).Return(link, nil)
	mockClient.On("SendDM", mock.Anything, "discord-user-200", mock.AnythingOfType("*client.DiscordEmbed")).Return(nil)

	notifier.NotifyPurchase(context.Background(), purchase)

	mockRepo.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

func Test_NotifyPurchase_SendsChannelMessage(t *testing.T) {
	mockRepo := new(MockNotificationsDiscordRepo)
	mockClient := new(MockDiscordClient)

	notifier := updaters.NewNotifications(mockRepo, mockClient, "")

	channelID := "channel-456"
	purchase := &models.PurchaseTransaction{
		ID:                1,
		BuyerUserID:       100,
		BuyerName:         "Alice",
		SellerUserID:      200,
		TypeName:          "Tritanium",
		LocationName:      "Jita IV - Moon 4",
		QuantityPurchased: 500,
		PricePerUnit:      10.0,
		TotalPrice:        5000.0,
		PurchasedAt:       time.Now(),
	}

	targets := []*models.DiscordNotificationTarget{
		{
			ID:         2,
			UserID:     200,
			TargetType: "channel",
			ChannelID:  &channelID,
			IsActive:   true,
		},
	}

	mockRepo.On("GetActiveTargetsForEvent", mock.Anything, int64(200), "purchase_created").Return(targets, nil)
	mockClient.On("SendChannelMessage", mock.Anything, "channel-456", mock.AnythingOfType("*client.DiscordEmbed")).Return(nil)

	notifier.NotifyPurchase(context.Background(), purchase)

	mockRepo.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

func Test_NotifyPurchase_NoTargets(t *testing.T) {
	mockRepo := new(MockNotificationsDiscordRepo)
	mockClient := new(MockDiscordClient)

	notifier := updaters.NewNotifications(mockRepo, mockClient, "")

	purchase := &models.PurchaseTransaction{
		SellerUserID: 200,
		PurchasedAt:  time.Now(),
	}

	mockRepo.On("GetActiveTargetsForEvent", mock.Anything, int64(200), "purchase_created").Return([]*models.DiscordNotificationTarget{}, nil)

	notifier.NotifyPurchase(context.Background(), purchase)

	mockRepo.AssertExpectations(t)
	// SendDM/SendChannelMessage should not be called
	mockClient.AssertNotCalled(t, "SendDM")
	mockClient.AssertNotCalled(t, "SendChannelMessage")
}

func Test_NotifyPurchase_RepoError(t *testing.T) {
	mockRepo := new(MockNotificationsDiscordRepo)
	mockClient := new(MockDiscordClient)

	notifier := updaters.NewNotifications(mockRepo, mockClient, "")

	purchase := &models.PurchaseTransaction{
		SellerUserID: 200,
		PurchasedAt:  time.Now(),
	}

	mockRepo.On("GetActiveTargetsForEvent", mock.Anything, int64(200), "purchase_created").Return(nil, errors.New("db error"))

	// Should not panic, just log
	notifier.NotifyPurchase(context.Background(), purchase)

	mockRepo.AssertExpectations(t)
}

func Test_NotifyPurchase_SendError(t *testing.T) {
	mockRepo := new(MockNotificationsDiscordRepo)
	mockClient := new(MockDiscordClient)

	notifier := updaters.NewNotifications(mockRepo, mockClient, "")

	channelID := "channel-456"
	purchase := &models.PurchaseTransaction{
		BuyerName:         "Alice",
		SellerUserID:      200,
		TypeName:          "Tritanium",
		LocationName:      "Jita",
		QuantityPurchased: 100,
		PricePerUnit:      5.0,
		TotalPrice:        500.0,
		PurchasedAt:       time.Now(),
	}

	targets := []*models.DiscordNotificationTarget{
		{ID: 1, UserID: 200, TargetType: "channel", ChannelID: &channelID, IsActive: true},
	}

	mockRepo.On("GetActiveTargetsForEvent", mock.Anything, int64(200), "purchase_created").Return(targets, nil)
	mockClient.On("SendChannelMessage", mock.Anything, "channel-456", mock.AnythingOfType("*client.DiscordEmbed")).Return(errors.New("discord error"))

	// Should not panic, just log
	notifier.NotifyPurchase(context.Background(), purchase)

	mockRepo.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

func Test_SendTestNotification_DM(t *testing.T) {
	mockRepo := new(MockNotificationsDiscordRepo)
	mockClient := new(MockDiscordClient)

	notifier := updaters.NewNotifications(mockRepo, mockClient, "")

	target := &models.DiscordNotificationTarget{
		ID:         1,
		UserID:     200,
		TargetType: "dm",
	}

	link := &models.DiscordLink{
		DiscordUserID: "discord-user-200",
	}

	mockClient.On("SendDM", mock.Anything, "discord-user-200", mock.AnythingOfType("*client.DiscordEmbed")).Return(nil)

	err := notifier.SendTestNotification(context.Background(), target, link)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func Test_SendTestNotification_Channel(t *testing.T) {
	mockRepo := new(MockNotificationsDiscordRepo)
	mockClient := new(MockDiscordClient)

	notifier := updaters.NewNotifications(mockRepo, mockClient, "")

	channelID := "channel-789"
	target := &models.DiscordNotificationTarget{
		ID:         2,
		UserID:     200,
		TargetType: "channel",
		ChannelID:  &channelID,
	}

	link := &models.DiscordLink{
		DiscordUserID: "discord-user-200",
	}

	mockClient.On("SendChannelMessage", mock.Anything, "channel-789", mock.AnythingOfType("*client.DiscordEmbed")).Return(nil)

	err := notifier.SendTestNotification(context.Background(), target, link)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func Test_SendTestNotification_ChannelWithoutID(t *testing.T) {
	mockRepo := new(MockNotificationsDiscordRepo)
	mockClient := new(MockDiscordClient)

	notifier := updaters.NewNotifications(mockRepo, mockClient, "")

	target := &models.DiscordNotificationTarget{
		ID:         2,
		UserID:     200,
		TargetType: "channel",
		ChannelID:  nil,
	}

	link := &models.DiscordLink{
		DiscordUserID: "discord-user-200",
	}

	err := notifier.SendTestNotification(context.Background(), target, link)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no channel_id")
}

func Test_NotifyContractCreated_SendsDM(t *testing.T) {
	mockRepo := new(MockNotificationsDiscordRepo)
	mockClient := new(MockDiscordClient)

	notifier := updaters.NewNotifications(mockRepo, mockClient, "")

	contractKey := "PT-42"
	purchase := &models.PurchaseTransaction{
		ID:                42,
		BuyerUserID:       100,
		BuyerName:         "Alice",
		SellerUserID:      200,
		SellerName:        "Bob",
		TypeID:            34,
		TypeName:          "Tritanium",
		LocationName:      "Jita IV - Moon 4",
		QuantityPurchased: 1000,
		PricePerUnit:      5.0,
		TotalPrice:        5000.0,
		ContractKey:       &contractKey,
		PurchasedAt:       time.Now(),
	}

	targets := []*models.DiscordNotificationTarget{
		{
			ID:         1,
			UserID:     100,
			TargetType: "dm",
			IsActive:   true,
		},
	}

	link := &models.DiscordLink{
		ID:            1,
		UserID:        100,
		DiscordUserID: "discord-user-100",
	}

	mockRepo.On("GetActiveTargetsForEvent", mock.Anything, int64(100), "contract_created").Return(targets, nil)
	mockRepo.On("GetLinkByUser", mock.Anything, int64(100)).Return(link, nil)
	mockClient.On("SendDM", mock.Anything, "discord-user-100", mock.AnythingOfType("*client.DiscordEmbed")).Return(nil)

	notifier.NotifyContractCreated(context.Background(), purchase)

	mockRepo.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

func Test_NotifyContractCreated_NoTargets(t *testing.T) {
	mockRepo := new(MockNotificationsDiscordRepo)
	mockClient := new(MockDiscordClient)

	notifier := updaters.NewNotifications(mockRepo, mockClient, "")

	purchase := &models.PurchaseTransaction{
		BuyerUserID: 100,
		PurchasedAt: time.Now(),
	}

	mockRepo.On("GetActiveTargetsForEvent", mock.Anything, int64(100), "contract_created").Return([]*models.DiscordNotificationTarget{}, nil)

	notifier.NotifyContractCreated(context.Background(), purchase)

	mockRepo.AssertExpectations(t)
	mockClient.AssertNotCalled(t, "SendDM")
	mockClient.AssertNotCalled(t, "SendChannelMessage")
}

func Test_NotifyContractCreated_SendError(t *testing.T) {
	mockRepo := new(MockNotificationsDiscordRepo)
	mockClient := new(MockDiscordClient)

	notifier := updaters.NewNotifications(mockRepo, mockClient, "")

	channelID := "channel-456"
	contractKey := "PT-99"
	purchase := &models.PurchaseTransaction{
		BuyerUserID:       100,
		SellerName:        "Bob",
		TypeName:          "Tritanium",
		LocationName:      "Jita",
		QuantityPurchased: 100,
		PricePerUnit:      5.0,
		TotalPrice:        500.0,
		ContractKey:       &contractKey,
		PurchasedAt:       time.Now(),
	}

	targets := []*models.DiscordNotificationTarget{
		{ID: 1, UserID: 100, TargetType: "channel", ChannelID: &channelID, IsActive: true},
	}

	mockRepo.On("GetActiveTargetsForEvent", mock.Anything, int64(100), "contract_created").Return(targets, nil)
	mockClient.On("SendChannelMessage", mock.Anything, "channel-456", mock.AnythingOfType("*client.DiscordEmbed")).Return(errors.New("discord error"))

	// Should not panic, just log
	notifier.NotifyContractCreated(context.Background(), purchase)

	mockRepo.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

func Test_NotifyPiStalls_WithDepletionAndURL(t *testing.T) {
	mockRepo := new(MockNotificationsDiscordRepo)
	mockClient := new(MockDiscordClient)

	frontendURL := "https://example.com/"
	notifier := updaters.NewNotifications(mockRepo, mockClient, frontendURL)

	channelID := "pi-channel-123"
	depletionTime := time.Date(2026, 2, 25, 3, 0, 0, 0, time.UTC)
	alerts := []*updaters.PiStallAlert{
		{
			CharacterName:     "TestChar",
			PlanetType:        "barren",
			SolarSystemName:   "Jita",
			StalledPins:       []updaters.PiStalledPin{{PinCategory: "extractor", Reason: "expired"}},
			DepletionTime:     &depletionTime,
			DepletedInputName: "Water",
		},
	}

	targets := []*models.DiscordNotificationTarget{
		{ID: 1, UserID: 42, TargetType: "channel", ChannelID: &channelID, IsActive: true},
	}

	var capturedEmbed *client.DiscordEmbed
	mockRepo.On("GetActiveTargetsForEvent", mock.Anything, int64(42), "pi_stall").Return(targets, nil)
	mockClient.On("SendChannelMessage", mock.Anything, "pi-channel-123", mock.AnythingOfType("*client.DiscordEmbed")).
		Run(func(args mock.Arguments) {
			capturedEmbed = args.Get(2).(*client.DiscordEmbed)
		}).
		Return(nil)

	notifier.NotifyPiStalls(context.Background(), 42, alerts)

	mockRepo.AssertExpectations(t)
	mockClient.AssertExpectations(t)

	assert.NotNil(t, capturedEmbed)
	assert.Equal(t, "PI Stall Detected", capturedEmbed.Title)
	assert.Equal(t, "https://example.com/pi", capturedEmbed.URL)
	assert.Len(t, capturedEmbed.Fields, 1)
	assert.Contains(t, capturedEmbed.Fields[0].Value, "1 extractor expired")
	assert.Contains(t, capturedEmbed.Fields[0].Value, "Inputs depleted since")
	assert.Contains(t, capturedEmbed.Fields[0].Value, "Water")
}

func Test_NotifyPiStalls_NoURL_WhenFrontendURLEmpty(t *testing.T) {
	mockRepo := new(MockNotificationsDiscordRepo)
	mockClient := new(MockDiscordClient)

	notifier := updaters.NewNotifications(mockRepo, mockClient, "")

	channelID := "pi-channel-456"
	alerts := []*updaters.PiStallAlert{
		{
			CharacterName:   "TestChar",
			PlanetType:      "barren",
			SolarSystemName: "Amarr",
			StalledPins:     []updaters.PiStalledPin{{PinCategory: "extractor", Reason: "expired"}},
		},
	}

	targets := []*models.DiscordNotificationTarget{
		{ID: 2, UserID: 99, TargetType: "channel", ChannelID: &channelID, IsActive: true},
	}

	var capturedEmbed *client.DiscordEmbed
	mockRepo.On("GetActiveTargetsForEvent", mock.Anything, int64(99), "pi_stall").Return(targets, nil)
	mockClient.On("SendChannelMessage", mock.Anything, "pi-channel-456", mock.AnythingOfType("*client.DiscordEmbed")).
		Run(func(args mock.Arguments) {
			capturedEmbed = args.Get(2).(*client.DiscordEmbed)
		}).
		Return(nil)

	notifier.NotifyPiStalls(context.Background(), 99, alerts)

	mockRepo.AssertExpectations(t)
	mockClient.AssertExpectations(t)

	assert.NotNil(t, capturedEmbed)
	assert.Equal(t, "", capturedEmbed.URL)
}
