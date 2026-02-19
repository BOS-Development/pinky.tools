package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type ContactPermissions struct {
	db *sql.DB
}

func NewContactPermissions(db *sql.DB) *ContactPermissions {
	return &ContactPermissions{db: db}
}

// GetByContact returns all permissions for a specific contact
func (r *ContactPermissions) GetByContact(ctx context.Context, contactID int64, userID int64) ([]*models.ContactPermission, error) {
	query := `
		SELECT id, contact_id, granting_user_id, receiving_user_id, service_type, can_access
		FROM contact_permissions
		WHERE contact_id = $1 AND (granting_user_id = $2 OR receiving_user_id = $2)
		ORDER BY service_type
	`

	rows, err := r.db.QueryContext(ctx, query, contactID, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query contact permissions")
	}
	defer rows.Close()

	var permissions []*models.ContactPermission
	for rows.Next() {
		var perm models.ContactPermission
		err = rows.Scan(
			&perm.ID,
			&perm.ContactID,
			&perm.GrantingUserID,
			&perm.ReceivingUserID,
			&perm.ServiceType,
			&perm.CanAccess,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan contact permission")
		}
		permissions = append(permissions, &perm)
	}

	return permissions, nil
}

// Upsert creates or updates a permission
func (r *ContactPermissions) Upsert(ctx context.Context, perm *models.ContactPermission) error {
	query := `
		INSERT INTO contact_permissions
		(contact_id, granting_user_id, receiving_user_id, service_type, can_access, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (contact_id, granting_user_id, receiving_user_id, service_type)
		DO UPDATE SET
			can_access = EXCLUDED.can_access,
			updated_at = NOW()
	`

	_, err := r.db.ExecContext(ctx, query,
		perm.ContactID,
		perm.GrantingUserID,
		perm.ReceivingUserID,
		perm.ServiceType,
		perm.CanAccess,
	)
	if err != nil {
		return errors.Wrap(err, "failed to upsert contact permission")
	}

	return nil
}

// UpsertInTx creates or updates a permission within an existing transaction
func (r *ContactPermissions) UpsertInTx(ctx context.Context, tx *sql.Tx, perm *models.ContactPermission) error {
	query := `
		INSERT INTO contact_permissions
		(contact_id, granting_user_id, receiving_user_id, service_type, can_access, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (contact_id, granting_user_id, receiving_user_id, service_type)
		DO UPDATE SET
			can_access = EXCLUDED.can_access,
			updated_at = NOW()
	`

	_, err := tx.ExecContext(ctx, query,
		perm.ContactID,
		perm.GrantingUserID,
		perm.ReceivingUserID,
		perm.ServiceType,
		perm.CanAccess,
	)
	if err != nil {
		return errors.Wrap(err, "failed to upsert contact permission in transaction")
	}

	return nil
}

// CheckPermission verifies if grantingUser allows receivingUser to access serviceType
func (r *ContactPermissions) CheckPermission(ctx context.Context, grantingUserID, receivingUserID int64, serviceType string) (bool, error) {
	query := `
		SELECT can_access
		FROM contact_permissions
		WHERE granting_user_id = $1 AND receiving_user_id = $2 AND service_type = $3
	`

	var canAccess bool
	err := r.db.QueryRowContext(ctx, query, grantingUserID, receivingUserID, serviceType).Scan(&canAccess)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrap(err, "failed to check permission")
	}

	return canAccess, nil
}

// GetUserPermissionsForService returns all users who granted permission to viewerUserID for a specific service
func (r *ContactPermissions) GetUserPermissionsForService(ctx context.Context, viewerUserID int64, serviceType string) ([]int64, error) {
	query := `
		SELECT granting_user_id
		FROM contact_permissions
		WHERE receiving_user_id = $1 AND service_type = $2 AND can_access = true
	`

	rows, err := r.db.QueryContext(ctx, query, viewerUserID, serviceType)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query user permissions")
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

// InitializePermissionsForContact creates default (all denied) permissions when contact accepted
func (r *ContactPermissions) InitializePermissionsForContact(ctx context.Context, tx *sql.Tx, contactID, userID1, userID2 int64) error {
	// Default service types
	serviceTypes := []string{"for_sale_browse"}

	query := `
		INSERT INTO contact_permissions
		(contact_id, granting_user_id, receiving_user_id, service_type, can_access)
		VALUES ($1, $2, $3, $4, false)
		ON CONFLICT (contact_id, granting_user_id, receiving_user_id, service_type) DO NOTHING
	`

	for _, serviceType := range serviceTypes {
		// User1 grants to User2
		_, err := tx.ExecContext(ctx, query, contactID, userID1, userID2, serviceType)
		if err != nil {
			return errors.Wrap(err, "failed to initialize permission for user1->user2")
		}

		// User2 grants to User1
		_, err = tx.ExecContext(ctx, query, contactID, userID2, userID1, serviceType)
		if err != nil {
			return errors.Wrap(err, "failed to initialize permission for user2->user1")
		}
	}

	return nil
}
