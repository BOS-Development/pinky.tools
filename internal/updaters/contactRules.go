package updaters

import (
	"context"
	"database/sql"

	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type ContactRulesContactsRepository interface {
	CreateAutoContact(ctx context.Context, tx *sql.Tx, requesterID, recipientID, ruleID int64) (int64, bool, error)
}

type ContactRulesRepository interface {
	GetMatchingRulesForCorporation(ctx context.Context, corpID int64) ([]*models.ContactRule, error)
	GetMatchingRulesForAlliance(ctx context.Context, allianceID int64) ([]*models.ContactRule, error)
	GetEveryoneRules(ctx context.Context) ([]*models.ContactRule, error)
	GetUsersForCorporation(ctx context.Context, corpID int64, excludeUserID int64) ([]int64, error)
	GetUsersForAlliance(ctx context.Context, allianceID int64, excludeUserID int64) ([]int64, error)
	GetAllUsers(ctx context.Context, excludeUserID int64) ([]int64, error)
	DeleteAutoContactsForRule(ctx context.Context, ruleID int64) error
}

type ContactRulesPermissionsRepository interface {
	InitializePermissionsForContact(ctx context.Context, tx *sql.Tx, contactID, userID1, userID2 int64) error
	UpsertInTx(ctx context.Context, tx *sql.Tx, perm *models.ContactPermission) error
}

type ContactRulesUpdater struct {
	contactsRepo     ContactRulesContactsRepository
	contactRulesRepo ContactRulesRepository
	permissionsRepo  ContactRulesPermissionsRepository
	db               *sql.DB
}

func NewContactRules(
	contactsRepo ContactRulesContactsRepository,
	contactRulesRepo ContactRulesRepository,
	permissionsRepo ContactRulesPermissionsRepository,
	db *sql.DB,
) *ContactRulesUpdater {
	return &ContactRulesUpdater{
		contactsRepo:     contactsRepo,
		contactRulesRepo: contactRulesRepo,
		permissionsRepo:  permissionsRepo,
		db:               db,
	}
}

// ApplyRule creates auto-contacts for all users matching the given rule
func (u *ContactRulesUpdater) ApplyRule(ctx context.Context, rule *models.ContactRule) error {
	var userIDs []int64
	var err error

	switch rule.RuleType {
	case "corporation":
		if rule.EntityID == nil {
			return errors.New("entity_id required for corporation rule")
		}
		userIDs, err = u.contactRulesRepo.GetUsersForCorporation(ctx, *rule.EntityID, rule.UserID)
	case "alliance":
		if rule.EntityID == nil {
			return errors.New("entity_id required for alliance rule")
		}
		userIDs, err = u.contactRulesRepo.GetUsersForAlliance(ctx, *rule.EntityID, rule.UserID)
	case "everyone":
		userIDs, err = u.contactRulesRepo.GetAllUsers(ctx, rule.UserID)
	default:
		return errors.Errorf("unknown rule type: %s", rule.RuleType)
	}

	if err != nil {
		return errors.Wrap(err, "failed to get matching users for rule")
	}

	for _, userID := range userIDs {
		if err := u.createContactAndPermissions(ctx, rule.UserID, userID, rule.ID); err != nil {
			log.Error("failed to create auto-contact", "ruleID", rule.ID, "targetUserID", userID, "error", err)
			continue
		}
	}

	return nil
}

// ApplyRulesForNewCorporation checks all rules and creates contacts for a user who just added a corporation
func (u *ContactRulesUpdater) ApplyRulesForNewCorporation(ctx context.Context, userID int64, corpID int64, allianceID int64) error {
	// Check corporation rules
	corpRules, err := u.contactRulesRepo.GetMatchingRulesForCorporation(ctx, corpID)
	if err != nil {
		return errors.Wrap(err, "failed to get matching corporation rules")
	}

	for _, rule := range corpRules {
		if rule.UserID == userID {
			continue
		}
		if err := u.createContactAndPermissions(ctx, rule.UserID, userID, rule.ID); err != nil {
			log.Error("failed to create auto-contact from corp rule", "ruleID", rule.ID, "userID", userID, "error", err)
		}
	}

	// Check alliance rules
	if allianceID > 0 {
		allianceRules, err := u.contactRulesRepo.GetMatchingRulesForAlliance(ctx, allianceID)
		if err != nil {
			return errors.Wrap(err, "failed to get matching alliance rules")
		}

		for _, rule := range allianceRules {
			if rule.UserID == userID {
				continue
			}
			if err := u.createContactAndPermissions(ctx, rule.UserID, userID, rule.ID); err != nil {
				log.Error("failed to create auto-contact from alliance rule", "ruleID", rule.ID, "userID", userID, "error", err)
			}
		}
	}

	// Check everyone rules
	everyoneRules, err := u.contactRulesRepo.GetEveryoneRules(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get everyone rules")
	}

	for _, rule := range everyoneRules {
		if rule.UserID == userID {
			continue
		}
		if err := u.createContactAndPermissions(ctx, rule.UserID, userID, rule.ID); err != nil {
			log.Error("failed to create auto-contact from everyone rule", "ruleID", rule.ID, "userID", userID, "error", err)
		}
	}

	return nil
}

// createContactAndPermissions creates an auto-contact and grants for_sale_browse permission
func (u *ContactRulesUpdater) createContactAndPermissions(ctx context.Context, ruleOwnerID, targetUserID, ruleID int64) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	contactID, isNew, err := u.contactsRepo.CreateAutoContact(ctx, tx, ruleOwnerID, targetUserID, ruleID)
	if err != nil {
		return errors.Wrap(err, "failed to create auto-contact")
	}

	// contactID == 0 means the contact was skipped (pending/rejected)
	if contactID == 0 {
		return tx.Commit()
	}

	if isNew {
		// Initialize bilateral permissions (all false by default)
		err = u.permissionsRepo.InitializePermissionsForContact(ctx, tx, contactID, ruleOwnerID, targetUserID)
		if err != nil {
			return errors.Wrap(err, "failed to initialize permissions")
		}
	}

	// Grant for_sale_browse from rule owner to target
	err = u.permissionsRepo.UpsertInTx(ctx, tx, &models.ContactPermission{
		ContactID:       contactID,
		GrantingUserID:  ruleOwnerID,
		ReceivingUserID: targetUserID,
		ServiceType:     "for_sale_browse",
		CanAccess:       true,
	})
	if err != nil {
		return errors.Wrap(err, "failed to grant for_sale_browse permission")
	}

	return tx.Commit()
}
