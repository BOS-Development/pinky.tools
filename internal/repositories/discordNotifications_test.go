package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func Test_DiscordNotificationsShouldCreateAndGetLink(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	discordRepo := repositories.NewDiscordNotifications(db)

	user := &repositories.User{ID: 500, Name: "Discord User 1"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 50000, Name: "Discord Char 1", UserID: user.ID}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	expiresAt := time.Now().Add(time.Hour).Truncate(time.Second)
	link := &models.DiscordLink{
		UserID:         user.ID,
		DiscordUserID:  "123456789",
		DiscordUsername: "testuser#1234",
		AccessToken:    "access-token-1",
		RefreshToken:   "refresh-token-1",
		TokenExpiresAt: expiresAt,
	}

	err = discordRepo.CreateLink(context.Background(), link)
	assert.NoError(t, err)
	assert.NotZero(t, link.ID)
	assert.NotZero(t, link.CreatedAt)
	assert.NotZero(t, link.UpdatedAt)

	// Get the link back
	retrieved, err := discordRepo.GetLinkByUser(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, link.ID, retrieved.ID)
	assert.Equal(t, user.ID, retrieved.UserID)
	assert.Equal(t, "123456789", retrieved.DiscordUserID)
	assert.Equal(t, "testuser#1234", retrieved.DiscordUsername)
	assert.Equal(t, "access-token-1", retrieved.AccessToken)
	assert.Equal(t, "refresh-token-1", retrieved.RefreshToken)
	assert.Equal(t, expiresAt.UTC(), retrieved.TokenExpiresAt.UTC())
}

func Test_DiscordNotificationsShouldUpsertLink(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	discordRepo := repositories.NewDiscordNotifications(db)

	user := &repositories.User{ID: 501, Name: "Discord User 2"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 50100, Name: "Discord Char 2", UserID: user.ID}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	expiresAt := time.Now().Add(time.Hour).Truncate(time.Second)
	link := &models.DiscordLink{
		UserID:         user.ID,
		DiscordUserID:  "111111111",
		DiscordUsername: "originaluser",
		AccessToken:    "original-access",
		RefreshToken:   "original-refresh",
		TokenExpiresAt: expiresAt,
	}

	err = discordRepo.CreateLink(context.Background(), link)
	assert.NoError(t, err)
	originalID := link.ID

	// Upsert with new Discord account info
	newExpiresAt := time.Now().Add(2 * time.Hour).Truncate(time.Second)
	updatedLink := &models.DiscordLink{
		UserID:         user.ID,
		DiscordUserID:  "222222222",
		DiscordUsername: "updateduser",
		AccessToken:    "updated-access",
		RefreshToken:   "updated-refresh",
		TokenExpiresAt: newExpiresAt,
	}

	err = discordRepo.CreateLink(context.Background(), updatedLink)
	assert.NoError(t, err)
	assert.Equal(t, originalID, updatedLink.ID)

	// Verify updated values
	retrieved, err := discordRepo.GetLinkByUser(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "222222222", retrieved.DiscordUserID)
	assert.Equal(t, "updateduser", retrieved.DiscordUsername)
	assert.Equal(t, "updated-access", retrieved.AccessToken)
	assert.Equal(t, "updated-refresh", retrieved.RefreshToken)
	assert.Equal(t, newExpiresAt.UTC(), retrieved.TokenExpiresAt.UTC())
}

func Test_DiscordNotificationsShouldReturnNilForNonExistentUser(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	discordRepo := repositories.NewDiscordNotifications(db)

	link, err := discordRepo.GetLinkByUser(context.Background(), 999999)
	assert.NoError(t, err)
	assert.Nil(t, link)
}

func Test_DiscordNotificationsShouldUpdateLinkTokens(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	discordRepo := repositories.NewDiscordNotifications(db)

	user := &repositories.User{ID: 503, Name: "Discord User 4"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 50300, Name: "Discord Char 4", UserID: user.ID}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	expiresAt := time.Now().Add(time.Hour).Truncate(time.Second)
	link := &models.DiscordLink{
		UserID:         user.ID,
		DiscordUserID:  "444444444",
		DiscordUsername: "tokenuser",
		AccessToken:    "old-access",
		RefreshToken:   "old-refresh",
		TokenExpiresAt: expiresAt,
	}

	err = discordRepo.CreateLink(context.Background(), link)
	assert.NoError(t, err)

	// Update tokens
	newExpiresAt := time.Now().Add(3 * time.Hour).Truncate(time.Second)
	err = discordRepo.UpdateLinkTokens(context.Background(), user.ID, "new-access", "new-refresh", newExpiresAt)
	assert.NoError(t, err)

	// Verify tokens were updated
	retrieved, err := discordRepo.GetLinkByUser(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "new-access", retrieved.AccessToken)
	assert.Equal(t, "new-refresh", retrieved.RefreshToken)
	assert.Equal(t, newExpiresAt.UTC(), retrieved.TokenExpiresAt.UTC())
	// Discord user info should remain the same
	assert.Equal(t, "444444444", retrieved.DiscordUserID)
	assert.Equal(t, "tokenuser", retrieved.DiscordUsername)
}

func Test_DiscordNotificationsShouldUpdateLinkTokensFailForNonExistentUser(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	discordRepo := repositories.NewDiscordNotifications(db)

	err = discordRepo.UpdateLinkTokens(context.Background(), 999998, "token", "refresh", time.Now())
	assert.Error(t, err)
}

func Test_DiscordNotificationsShouldCreateTargetsAndGetByUser(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	discordRepo := repositories.NewDiscordNotifications(db)

	user := &repositories.User{ID: 504, Name: "Discord User 5"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 50400, Name: "Discord Char 5", UserID: user.ID}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	// Create a DM target
	dmTarget := &models.DiscordNotificationTarget{
		UserID:      user.ID,
		TargetType:  "dm",
		ChannelID:   nil,
		GuildName:   "",
		ChannelName: "",
	}
	err = discordRepo.CreateTarget(context.Background(), dmTarget)
	assert.NoError(t, err)
	assert.NotZero(t, dmTarget.ID)
	assert.True(t, dmTarget.IsActive)

	// Create a channel target
	channelID := "987654321"
	channelTarget := &models.DiscordNotificationTarget{
		UserID:      user.ID,
		TargetType:  "channel",
		ChannelID:   &channelID,
		GuildName:   "Test Guild",
		ChannelName: "notifications",
	}
	err = discordRepo.CreateTarget(context.Background(), channelTarget)
	assert.NoError(t, err)
	assert.NotZero(t, channelTarget.ID)
	assert.True(t, channelTarget.IsActive)

	// Get targets by user
	targets, err := discordRepo.GetTargetsByUser(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, targets, 2)

	// Verify DM target
	assert.Equal(t, "dm", targets[0].TargetType)
	assert.Nil(t, targets[0].ChannelID)
	assert.Equal(t, user.ID, targets[0].UserID)

	// Verify channel target
	assert.Equal(t, "channel", targets[1].TargetType)
	assert.NotNil(t, targets[1].ChannelID)
	assert.Equal(t, "987654321", *targets[1].ChannelID)
	assert.Equal(t, "Test Guild", targets[1].GuildName)
	assert.Equal(t, "notifications", targets[1].ChannelName)
}

func Test_DiscordNotificationsShouldGetTargetByID(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	discordRepo := repositories.NewDiscordNotifications(db)

	user := &repositories.User{ID: 505, Name: "Discord User 6"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 50500, Name: "Discord Char 6", UserID: user.ID}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	channelID := "111222333"
	target := &models.DiscordNotificationTarget{
		UserID:      user.ID,
		TargetType:  "channel",
		ChannelID:   &channelID,
		GuildName:   "My Guild",
		ChannelName: "alerts",
	}
	err = discordRepo.CreateTarget(context.Background(), target)
	assert.NoError(t, err)

	// Get by ID
	retrieved, err := discordRepo.GetTargetByID(context.Background(), target.ID)
	assert.NoError(t, err)
	assert.Equal(t, target.ID, retrieved.ID)
	assert.Equal(t, user.ID, retrieved.UserID)
	assert.Equal(t, "channel", retrieved.TargetType)
	assert.Equal(t, "111222333", *retrieved.ChannelID)
	assert.Equal(t, "My Guild", retrieved.GuildName)
	assert.Equal(t, "alerts", retrieved.ChannelName)
	assert.True(t, retrieved.IsActive)
}

func Test_DiscordNotificationsShouldGetTargetByIDFailForNonExistent(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	discordRepo := repositories.NewDiscordNotifications(db)

	_, err = discordRepo.GetTargetByID(context.Background(), 999999)
	assert.Error(t, err)
}

func Test_DiscordNotificationsShouldUpdateTarget(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	discordRepo := repositories.NewDiscordNotifications(db)

	user := &repositories.User{ID: 506, Name: "Discord User 7"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 50600, Name: "Discord Char 7", UserID: user.ID}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	target := &models.DiscordNotificationTarget{
		UserID:      user.ID,
		TargetType:  "dm",
		GuildName:   "",
		ChannelName: "",
	}
	err = discordRepo.CreateTarget(context.Background(), target)
	assert.NoError(t, err)
	assert.True(t, target.IsActive)

	// Toggle active to false
	target.IsActive = false
	err = discordRepo.UpdateTarget(context.Background(), target)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := discordRepo.GetTargetByID(context.Background(), target.ID)
	assert.NoError(t, err)
	assert.False(t, retrieved.IsActive)

	// Toggle back to true
	target.IsActive = true
	err = discordRepo.UpdateTarget(context.Background(), target)
	assert.NoError(t, err)

	retrieved, err = discordRepo.GetTargetByID(context.Background(), target.ID)
	assert.NoError(t, err)
	assert.True(t, retrieved.IsActive)
}

func Test_DiscordNotificationsShouldDeleteTarget(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	discordRepo := repositories.NewDiscordNotifications(db)

	user := &repositories.User{ID: 507, Name: "Discord User 8"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 50700, Name: "Discord Char 8", UserID: user.ID}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	target := &models.DiscordNotificationTarget{
		UserID:      user.ID,
		TargetType:  "dm",
		GuildName:   "",
		ChannelName: "",
	}
	err = discordRepo.CreateTarget(context.Background(), target)
	assert.NoError(t, err)

	// Delete the target
	err = discordRepo.DeleteTarget(context.Background(), target.ID, user.ID)
	assert.NoError(t, err)

	// Verify target is gone
	targets, err := discordRepo.GetTargetsByUser(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, targets, 0)
}

func Test_DiscordNotificationsShouldDeleteTargetFailForWrongUser(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	discordRepo := repositories.NewDiscordNotifications(db)

	user := &repositories.User{ID: 508, Name: "Discord User 9"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 50800, Name: "Discord Char 9", UserID: user.ID}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	target := &models.DiscordNotificationTarget{
		UserID:      user.ID,
		TargetType:  "dm",
		GuildName:   "",
		ChannelName: "",
	}
	err = discordRepo.CreateTarget(context.Background(), target)
	assert.NoError(t, err)

	// Try to delete with wrong user ID
	err = discordRepo.DeleteTarget(context.Background(), target.ID, 999997)
	assert.Error(t, err)
}

func Test_DiscordNotificationsShouldUpsertAndGetPreferences(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	discordRepo := repositories.NewDiscordNotifications(db)

	user := &repositories.User{ID: 509, Name: "Discord User 10"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 50900, Name: "Discord Char 10", UserID: user.ID}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	target := &models.DiscordNotificationTarget{
		UserID:      user.ID,
		TargetType:  "dm",
		GuildName:   "",
		ChannelName: "",
	}
	err = discordRepo.CreateTarget(context.Background(), target)
	assert.NoError(t, err)

	// Create preferences
	pref1 := &models.NotificationPreference{
		TargetID:  target.ID,
		EventType: "asset_change",
		IsEnabled: true,
	}
	err = discordRepo.UpsertPreference(context.Background(), pref1)
	assert.NoError(t, err)
	assert.NotZero(t, pref1.ID)

	pref2 := &models.NotificationPreference{
		TargetID:  target.ID,
		EventType: "price_alert",
		IsEnabled: false,
	}
	err = discordRepo.UpsertPreference(context.Background(), pref2)
	assert.NoError(t, err)
	assert.NotZero(t, pref2.ID)

	// Get preferences by target
	prefs, err := discordRepo.GetPreferencesByTarget(context.Background(), target.ID)
	assert.NoError(t, err)
	assert.Len(t, prefs, 2)

	// Verify preferences exist
	foundAssetChange := false
	foundPriceAlert := false
	for _, p := range prefs {
		if p.EventType == "asset_change" {
			foundAssetChange = true
			assert.True(t, p.IsEnabled)
		}
		if p.EventType == "price_alert" {
			foundPriceAlert = true
			assert.False(t, p.IsEnabled)
		}
	}
	assert.True(t, foundAssetChange)
	assert.True(t, foundPriceAlert)

	// Upsert to update existing preference
	pref2.IsEnabled = true
	err = discordRepo.UpsertPreference(context.Background(), pref2)
	assert.NoError(t, err)

	// Verify updated preference
	prefs, err = discordRepo.GetPreferencesByTarget(context.Background(), target.ID)
	assert.NoError(t, err)
	assert.Len(t, prefs, 2)

	for _, p := range prefs {
		if p.EventType == "price_alert" {
			assert.True(t, p.IsEnabled)
		}
	}
}

func Test_DiscordNotificationsShouldGetActiveTargetsForEvent(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	discordRepo := repositories.NewDiscordNotifications(db)

	user := &repositories.User{ID: 510, Name: "Discord User 11"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 51000, Name: "Discord Char 11", UserID: user.ID}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	// Create an active target with the event enabled
	activeTarget := &models.DiscordNotificationTarget{
		UserID:      user.ID,
		TargetType:  "dm",
		GuildName:   "",
		ChannelName: "",
	}
	err = discordRepo.CreateTarget(context.Background(), activeTarget)
	assert.NoError(t, err)

	pref1 := &models.NotificationPreference{
		TargetID:  activeTarget.ID,
		EventType: "asset_change",
		IsEnabled: true,
	}
	err = discordRepo.UpsertPreference(context.Background(), pref1)
	assert.NoError(t, err)

	// Create an inactive target with the event enabled
	inactiveTarget := &models.DiscordNotificationTarget{
		UserID:      user.ID,
		TargetType:  "channel",
		GuildName:   "Some Guild",
		ChannelName: "some-channel",
	}
	err = discordRepo.CreateTarget(context.Background(), inactiveTarget)
	assert.NoError(t, err)

	// Deactivate the target
	inactiveTarget.IsActive = false
	err = discordRepo.UpdateTarget(context.Background(), inactiveTarget)
	assert.NoError(t, err)

	pref2 := &models.NotificationPreference{
		TargetID:  inactiveTarget.ID,
		EventType: "asset_change",
		IsEnabled: true,
	}
	err = discordRepo.UpsertPreference(context.Background(), pref2)
	assert.NoError(t, err)

	// Create an active target with the event disabled
	disabledPrefTarget := &models.DiscordNotificationTarget{
		UserID:      user.ID,
		TargetType:  "dm",
		GuildName:   "",
		ChannelName: "",
	}
	err = discordRepo.CreateTarget(context.Background(), disabledPrefTarget)
	assert.NoError(t, err)

	pref3 := &models.NotificationPreference{
		TargetID:  disabledPrefTarget.ID,
		EventType: "asset_change",
		IsEnabled: false,
	}
	err = discordRepo.UpsertPreference(context.Background(), pref3)
	assert.NoError(t, err)

	// Get active targets for asset_change event
	targets, err := discordRepo.GetActiveTargetsForEvent(context.Background(), user.ID, "asset_change")
	assert.NoError(t, err)
	assert.Len(t, targets, 1)
	assert.Equal(t, activeTarget.ID, targets[0].ID)
	assert.True(t, targets[0].IsActive)
}

func Test_DiscordNotificationsShouldGetActiveTargetsForEventEmpty(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	discordRepo := repositories.NewDiscordNotifications(db)

	user := &repositories.User{ID: 511, Name: "Discord User 12"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 51100, Name: "Discord Char 12", UserID: user.ID}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	// No targets at all
	targets, err := discordRepo.GetActiveTargetsForEvent(context.Background(), user.ID, "asset_change")
	assert.NoError(t, err)
	assert.Len(t, targets, 0)
}

func Test_DiscordNotificationsShouldDeleteLinkCascadesToTargets(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	discordRepo := repositories.NewDiscordNotifications(db)

	user := &repositories.User{ID: 512, Name: "Discord User 13"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 51200, Name: "Discord Char 13", UserID: user.ID}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	// Create link
	expiresAt := time.Now().Add(time.Hour).Truncate(time.Second)
	link := &models.DiscordLink{
		UserID:         user.ID,
		DiscordUserID:  "999888777",
		DiscordUsername: "cascadeuser",
		AccessToken:    "cascade-access",
		RefreshToken:   "cascade-refresh",
		TokenExpiresAt: expiresAt,
	}
	err = discordRepo.CreateLink(context.Background(), link)
	assert.NoError(t, err)

	// Create targets
	target1 := &models.DiscordNotificationTarget{
		UserID:      user.ID,
		TargetType:  "dm",
		GuildName:   "",
		ChannelName: "",
	}
	err = discordRepo.CreateTarget(context.Background(), target1)
	assert.NoError(t, err)

	channelID := "cascade-channel"
	target2 := &models.DiscordNotificationTarget{
		UserID:      user.ID,
		TargetType:  "channel",
		ChannelID:   &channelID,
		GuildName:   "Cascade Guild",
		ChannelName: "cascade-channel",
	}
	err = discordRepo.CreateTarget(context.Background(), target2)
	assert.NoError(t, err)

	// Create preferences on a target
	pref := &models.NotificationPreference{
		TargetID:  target1.ID,
		EventType: "asset_change",
		IsEnabled: true,
	}
	err = discordRepo.UpsertPreference(context.Background(), pref)
	assert.NoError(t, err)

	// Delete the link (should cascade to targets and preferences)
	err = discordRepo.DeleteLink(context.Background(), user.ID)
	assert.NoError(t, err)

	// Verify link is gone
	retrieved, err := discordRepo.GetLinkByUser(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Nil(t, retrieved)

	// Verify targets are gone
	targets, err := discordRepo.GetTargetsByUser(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, targets, 0)

	// Verify preferences are gone (target was deleted, so preferences cascade deleted)
	prefs, err := discordRepo.GetPreferencesByTarget(context.Background(), target1.ID)
	assert.NoError(t, err)
	assert.Len(t, prefs, 0)
}

func Test_DiscordNotificationsShouldDeleteTargetCascadePreferences(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	discordRepo := repositories.NewDiscordNotifications(db)

	user := &repositories.User{ID: 513, Name: "Discord User 14"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 51300, Name: "Discord Char 14", UserID: user.ID}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	// Create target with preferences
	target := &models.DiscordNotificationTarget{
		UserID:      user.ID,
		TargetType:  "dm",
		GuildName:   "",
		ChannelName: "",
	}
	err = discordRepo.CreateTarget(context.Background(), target)
	assert.NoError(t, err)

	pref1 := &models.NotificationPreference{
		TargetID:  target.ID,
		EventType: "asset_change",
		IsEnabled: true,
	}
	err = discordRepo.UpsertPreference(context.Background(), pref1)
	assert.NoError(t, err)

	pref2 := &models.NotificationPreference{
		TargetID:  target.ID,
		EventType: "price_alert",
		IsEnabled: true,
	}
	err = discordRepo.UpsertPreference(context.Background(), pref2)
	assert.NoError(t, err)

	// Delete target
	err = discordRepo.DeleteTarget(context.Background(), target.ID, user.ID)
	assert.NoError(t, err)

	// Verify preferences are cascade deleted
	prefs, err := discordRepo.GetPreferencesByTarget(context.Background(), target.ID)
	assert.NoError(t, err)
	assert.Len(t, prefs, 0)
}

func Test_DiscordNotificationsShouldGetTargetsByUserEmpty(t *testing.T) {
	db, err := setupDatabase(t)
	assert.NoError(t, err)

	userRepo := repositories.NewUserRepository(db)
	characterRepo := repositories.NewCharacterRepository(db)
	discordRepo := repositories.NewDiscordNotifications(db)

	user := &repositories.User{ID: 514, Name: "Discord User 15"}
	err = userRepo.Add(context.Background(), user)
	assert.NoError(t, err)

	char := &repositories.Character{ID: 51400, Name: "Discord Char 15", UserID: user.ID}
	err = characterRepo.Add(context.Background(), char)
	assert.NoError(t, err)

	// No targets created
	targets, err := discordRepo.GetTargetsByUser(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, targets, 0)
}
