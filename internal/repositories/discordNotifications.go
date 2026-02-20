package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type DiscordNotifications struct {
	db *sql.DB
}

func NewDiscordNotifications(db *sql.DB) *DiscordNotifications {
	return &DiscordNotifications{db: db}
}

// CreateLink creates a Discord link for a user
func (r *DiscordNotifications) CreateLink(ctx context.Context, link *models.DiscordLink) error {
	query := `
		INSERT INTO discord_links (user_id, discord_user_id, discord_username, access_token, refresh_token, token_expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE SET
			discord_user_id = EXCLUDED.discord_user_id,
			discord_username = EXCLUDED.discord_username,
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			token_expires_at = EXCLUDED.token_expires_at,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		link.UserID,
		link.DiscordUserID,
		link.DiscordUsername,
		link.AccessToken,
		link.RefreshToken,
		link.TokenExpiresAt,
	).Scan(&link.ID, &link.CreatedAt, &link.UpdatedAt)

	if err != nil {
		return errors.Wrap(err, "failed to create discord link")
	}

	return nil
}

// GetLinkByUser gets the Discord link for a user
func (r *DiscordNotifications) GetLinkByUser(ctx context.Context, userID int64) (*models.DiscordLink, error) {
	query := `
		SELECT id, user_id, discord_user_id, discord_username, access_token, refresh_token, token_expires_at, created_at, updated_at
		FROM discord_links
		WHERE user_id = $1
	`

	link := &models.DiscordLink{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&link.ID,
		&link.UserID,
		&link.DiscordUserID,
		&link.DiscordUsername,
		&link.AccessToken,
		&link.RefreshToken,
		&link.TokenExpiresAt,
		&link.CreatedAt,
		&link.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get discord link")
	}

	return link, nil
}

// UpdateLinkTokens updates the OAuth tokens for a Discord link
func (r *DiscordNotifications) UpdateLinkTokens(ctx context.Context, userID int64, accessToken, refreshToken string, expiresAt time.Time) error {
	query := `
		UPDATE discord_links
		SET access_token = $2, refresh_token = $3, token_expires_at = $4, updated_at = NOW()
		WHERE user_id = $1
	`

	result, err := r.db.ExecContext(ctx, query, userID, accessToken, refreshToken, expiresAt)
	if err != nil {
		return errors.Wrap(err, "failed to update discord link tokens")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("discord link not found")
	}

	return nil
}

// DeleteLink removes a Discord link and all associated targets/preferences (via cascade)
func (r *DiscordNotifications) DeleteLink(ctx context.Context, userID int64) error {
	// First delete targets (which cascades to preferences)
	_, err := r.db.ExecContext(ctx, `DELETE FROM discord_notification_targets WHERE user_id = $1`, userID)
	if err != nil {
		return errors.Wrap(err, "failed to delete discord targets")
	}

	// Then delete the link
	_, err = r.db.ExecContext(ctx, `DELETE FROM discord_links WHERE user_id = $1`, userID)
	if err != nil {
		return errors.Wrap(err, "failed to delete discord link")
	}

	return nil
}

// CreateTarget creates a notification target
func (r *DiscordNotifications) CreateTarget(ctx context.Context, target *models.DiscordNotificationTarget) error {
	query := `
		INSERT INTO discord_notification_targets (user_id, target_type, channel_id, guild_name, channel_name)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, is_active, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		target.UserID,
		target.TargetType,
		target.ChannelID,
		target.GuildName,
		target.ChannelName,
	).Scan(&target.ID, &target.IsActive, &target.CreatedAt, &target.UpdatedAt)

	if err != nil {
		return errors.Wrap(err, "failed to create notification target")
	}

	return nil
}

// GetTargetsByUser returns all notification targets for a user
func (r *DiscordNotifications) GetTargetsByUser(ctx context.Context, userID int64) ([]*models.DiscordNotificationTarget, error) {
	query := `
		SELECT id, user_id, target_type, channel_id, guild_name, channel_name, is_active, created_at, updated_at
		FROM discord_notification_targets
		WHERE user_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query notification targets")
	}
	defer rows.Close()

	targets := []*models.DiscordNotificationTarget{}
	for rows.Next() {
		t := &models.DiscordNotificationTarget{}
		err := rows.Scan(&t.ID, &t.UserID, &t.TargetType, &t.ChannelID, &t.GuildName, &t.ChannelName, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan notification target")
		}
		targets = append(targets, t)
	}

	return targets, nil
}

// GetTargetByID returns a notification target by ID
func (r *DiscordNotifications) GetTargetByID(ctx context.Context, id int64) (*models.DiscordNotificationTarget, error) {
	query := `
		SELECT id, user_id, target_type, channel_id, guild_name, channel_name, is_active, created_at, updated_at
		FROM discord_notification_targets
		WHERE id = $1
	`

	t := &models.DiscordNotificationTarget{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.UserID, &t.TargetType, &t.ChannelID, &t.GuildName, &t.ChannelName, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("notification target not found")
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get notification target")
	}

	return t, nil
}

// UpdateTarget updates a notification target
func (r *DiscordNotifications) UpdateTarget(ctx context.Context, target *models.DiscordNotificationTarget) error {
	query := `
		UPDATE discord_notification_targets
		SET is_active = $2, updated_at = NOW()
		WHERE id = $1 AND user_id = $3
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query, target.ID, target.IsActive, target.UserID).Scan(&target.UpdatedAt)

	if err == sql.ErrNoRows {
		return errors.New("notification target not found")
	}
	if err != nil {
		return errors.Wrap(err, "failed to update notification target")
	}

	return nil
}

// DeleteTarget deletes a notification target (cascade deletes preferences)
func (r *DiscordNotifications) DeleteTarget(ctx context.Context, id int64, userID int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM discord_notification_targets WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return errors.Wrap(err, "failed to delete notification target")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("notification target not found")
	}

	return nil
}

// UpsertPreference creates or updates a notification preference
func (r *DiscordNotifications) UpsertPreference(ctx context.Context, pref *models.NotificationPreference) error {
	query := `
		INSERT INTO notification_preferences (target_id, event_type, is_enabled)
		VALUES ($1, $2, $3)
		ON CONFLICT (target_id, event_type) DO UPDATE SET
			is_enabled = EXCLUDED.is_enabled
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query, pref.TargetID, pref.EventType, pref.IsEnabled).Scan(&pref.ID)
	if err != nil {
		return errors.Wrap(err, "failed to upsert notification preference")
	}

	return nil
}

// GetPreferencesByTarget returns all preferences for a target
func (r *DiscordNotifications) GetPreferencesByTarget(ctx context.Context, targetID int64) ([]*models.NotificationPreference, error) {
	query := `
		SELECT id, target_id, event_type, is_enabled
		FROM notification_preferences
		WHERE target_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, targetID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query notification preferences")
	}
	defer rows.Close()

	prefs := []*models.NotificationPreference{}
	for rows.Next() {
		p := &models.NotificationPreference{}
		err := rows.Scan(&p.ID, &p.TargetID, &p.EventType, &p.IsEnabled)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan notification preference")
		}
		prefs = append(prefs, p)
	}

	return prefs, nil
}

// GetActiveTargetsForEvent returns active targets with the given event type enabled
func (r *DiscordNotifications) GetActiveTargetsForEvent(ctx context.Context, userID int64, eventType string) ([]*models.DiscordNotificationTarget, error) {
	query := `
		SELECT dt.id, dt.user_id, dt.target_type, dt.channel_id, dt.guild_name, dt.channel_name, dt.is_active, dt.created_at, dt.updated_at
		FROM discord_notification_targets dt
		INNER JOIN notification_preferences np ON np.target_id = dt.id
		WHERE dt.user_id = $1 AND dt.is_active = true
			AND np.event_type = $2 AND np.is_enabled = true
	`

	rows, err := r.db.QueryContext(ctx, query, userID, eventType)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query active targets for event")
	}
	defer rows.Close()

	targets := []*models.DiscordNotificationTarget{}
	for rows.Next() {
		t := &models.DiscordNotificationTarget{}
		err := rows.Scan(&t.ID, &t.UserID, &t.TargetType, &t.ChannelID, &t.GuildName, &t.ChannelName, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan active target")
		}
		targets = append(targets, t)
	}

	return targets, nil
}
