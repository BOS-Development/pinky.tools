package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type ContactRules struct {
	db *sql.DB
}

func NewContactRules(db *sql.DB) *ContactRules {
	return &ContactRules{db: db}
}

// GetByUser returns all active rules for a user
func (r *ContactRules) GetByUser(ctx context.Context, userID int64) ([]*models.ContactRule, error) {
	query := `
		SELECT id, user_id, rule_type, entity_id, entity_name, is_active, created_at, updated_at
		FROM contact_rules
		WHERE user_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query contact rules")
	}
	defer rows.Close()

	items := []*models.ContactRule{}
	for rows.Next() {
		var item models.ContactRule
		err = rows.Scan(
			&item.ID, &item.UserID, &item.RuleType, &item.EntityID,
			&item.EntityName, &item.IsActive, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan contact rule")
		}
		items = append(items, &item)
	}

	return items, nil
}

// Create inserts a new contact rule
func (r *ContactRules) Create(ctx context.Context, rule *models.ContactRule) error {
	query := `
		INSERT INTO contact_rules (user_id, rule_type, entity_id, entity_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, is_active, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		rule.UserID, rule.RuleType, rule.EntityID, rule.EntityName,
	).Scan(&rule.ID, &rule.IsActive, &rule.CreatedAt, &rule.UpdatedAt)

	if err != nil {
		return errors.Wrap(err, "failed to create contact rule")
	}

	return nil
}

// Delete soft-deletes a contact rule
func (r *ContactRules) Delete(ctx context.Context, ruleID int64, userID int64) error {
	query := `
		UPDATE contact_rules
		SET is_active = false, updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND is_active = true
	`

	result, err := r.db.ExecContext(ctx, query, ruleID, userID)
	if err != nil {
		return errors.Wrap(err, "failed to delete contact rule")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.New("contact rule not found or user is not the owner")
	}

	return nil
}

// GetMatchingRulesForCorporation returns all active rules matching a corporation ID
func (r *ContactRules) GetMatchingRulesForCorporation(ctx context.Context, corpID int64) ([]*models.ContactRule, error) {
	query := `
		SELECT id, user_id, rule_type, entity_id, entity_name, is_active, created_at, updated_at
		FROM contact_rules
		WHERE rule_type = 'corporation' AND entity_id = $1 AND is_active = true
	`

	rows, err := r.db.QueryContext(ctx, query, corpID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query matching corporation rules")
	}
	defer rows.Close()

	items := []*models.ContactRule{}
	for rows.Next() {
		var item models.ContactRule
		err = rows.Scan(
			&item.ID, &item.UserID, &item.RuleType, &item.EntityID,
			&item.EntityName, &item.IsActive, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan contact rule")
		}
		items = append(items, &item)
	}

	return items, nil
}

// GetMatchingRulesForAlliance returns all active rules matching an alliance ID
func (r *ContactRules) GetMatchingRulesForAlliance(ctx context.Context, allianceID int64) ([]*models.ContactRule, error) {
	query := `
		SELECT id, user_id, rule_type, entity_id, entity_name, is_active, created_at, updated_at
		FROM contact_rules
		WHERE rule_type = 'alliance' AND entity_id = $1 AND is_active = true
	`

	rows, err := r.db.QueryContext(ctx, query, allianceID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query matching alliance rules")
	}
	defer rows.Close()

	items := []*models.ContactRule{}
	for rows.Next() {
		var item models.ContactRule
		err = rows.Scan(
			&item.ID, &item.UserID, &item.RuleType, &item.EntityID,
			&item.EntityName, &item.IsActive, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan contact rule")
		}
		items = append(items, &item)
	}

	return items, nil
}

// GetEveryoneRules returns all active "everyone" rules
func (r *ContactRules) GetEveryoneRules(ctx context.Context) ([]*models.ContactRule, error) {
	query := `
		SELECT id, user_id, rule_type, entity_id, entity_name, is_active, created_at, updated_at
		FROM contact_rules
		WHERE rule_type = 'everyone' AND is_active = true
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query everyone rules")
	}
	defer rows.Close()

	items := []*models.ContactRule{}
	for rows.Next() {
		var item models.ContactRule
		err = rows.Scan(
			&item.ID, &item.UserID, &item.RuleType, &item.EntityID,
			&item.EntityName, &item.IsActive, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan contact rule")
		}
		items = append(items, &item)
	}

	return items, nil
}

// GetUsersForCorporation returns all user IDs that have a given corporation, excluding a specific user
func (r *ContactRules) GetUsersForCorporation(ctx context.Context, corpID int64, excludeUserID int64) ([]int64, error) {
	query := `
		SELECT DISTINCT user_id
		FROM player_corporations
		WHERE id = $1 AND user_id != $2
	`

	rows, err := r.db.QueryContext(ctx, query, corpID, excludeUserID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query users for corporation")
	}
	defer rows.Close()

	var userIDs []int64
	for rows.Next() {
		var userID int64
		err = rows.Scan(&userID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan user ID")
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

// GetUsersForAlliance returns all user IDs that have a corporation in the given alliance
func (r *ContactRules) GetUsersForAlliance(ctx context.Context, allianceID int64, excludeUserID int64) ([]int64, error) {
	query := `
		SELECT DISTINCT user_id
		FROM player_corporations
		WHERE alliance_id = $1 AND user_id != $2
	`

	rows, err := r.db.QueryContext(ctx, query, allianceID, excludeUserID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query users for alliance")
	}
	defer rows.Close()

	var userIDs []int64
	for rows.Next() {
		var userID int64
		err = rows.Scan(&userID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan user ID")
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

// GetAllUsers returns all user IDs except the specified one
func (r *ContactRules) GetAllUsers(ctx context.Context, excludeUserID int64) ([]int64, error) {
	query := `
		SELECT id FROM users WHERE id != $1
	`

	rows, err := r.db.QueryContext(ctx, query, excludeUserID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query all users")
	}
	defer rows.Close()

	var userIDs []int64
	for rows.Next() {
		var userID int64
		err = rows.Scan(&userID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan user ID")
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

type SearchResult struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// SearchCorporations searches for corporations by name
func (r *ContactRules) SearchCorporations(ctx context.Context, query string) ([]*SearchResult, error) {
	sqlQuery := `
		SELECT DISTINCT id, name
		FROM player_corporations
		WHERE name ILIKE '%' || $1 || '%'
		ORDER BY name
		LIMIT 20
	`

	rows, err := r.db.QueryContext(ctx, sqlQuery, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search corporations")
	}
	defer rows.Close()

	items := []*SearchResult{}
	for rows.Next() {
		var item SearchResult
		err = rows.Scan(&item.ID, &item.Name)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan search result")
		}
		items = append(items, &item)
	}

	return items, nil
}

// SearchAlliances searches for alliances by name
func (r *ContactRules) SearchAlliances(ctx context.Context, query string) ([]*SearchResult, error) {
	sqlQuery := `
		SELECT DISTINCT alliance_id, alliance_name
		FROM player_corporations
		WHERE alliance_id IS NOT NULL
			AND alliance_name ILIKE '%' || $1 || '%'
		ORDER BY alliance_name
		LIMIT 20
	`

	rows, err := r.db.QueryContext(ctx, sqlQuery, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search alliances")
	}
	defer rows.Close()

	items := []*SearchResult{}
	for rows.Next() {
		var item SearchResult
		err = rows.Scan(&item.ID, &item.Name)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan search result")
		}
		items = append(items, &item)
	}

	return items, nil
}

// DeleteAutoContactsForRule removes all contacts created by a specific rule
func (r *ContactRules) DeleteAutoContactsForRule(ctx context.Context, ruleID int64) error {
	query := `DELETE FROM contacts WHERE contact_rule_id = $1`

	_, err := r.db.ExecContext(ctx, query, ruleID)
	if err != nil {
		return errors.Wrap(err, "failed to delete auto-contacts for rule")
	}

	return nil
}
