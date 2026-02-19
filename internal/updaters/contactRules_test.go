package updaters_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/updaters"
	"github.com/stretchr/testify/assert"
)

// Mock contacts repository
type mockContactsRepo struct {
	createAutoContactFn func(ctx context.Context, tx *sql.Tx, requesterID, recipientID, ruleID int64) (int64, bool, error)
	calls               []createAutoContactCall
}

type createAutoContactCall struct {
	RequesterID int64
	RecipientID int64
	RuleID      int64
}

func (m *mockContactsRepo) CreateAutoContact(ctx context.Context, tx *sql.Tx, requesterID, recipientID, ruleID int64) (int64, bool, error) {
	m.calls = append(m.calls, createAutoContactCall{RequesterID: requesterID, RecipientID: recipientID, RuleID: ruleID})
	if m.createAutoContactFn != nil {
		return m.createAutoContactFn(ctx, tx, requesterID, recipientID, ruleID)
	}
	return 100, true, nil
}

// Mock contact rules repository
type mockRulesRepo struct {
	corpRules    []*models.ContactRule
	corpRulesErr error
	alliRules    []*models.ContactRule
	alliRulesErr error
	everyoneRules []*models.ContactRule
	everyoneErr  error
	corpUsers    []int64
	corpUsersErr error
	alliUsers    []int64
	alliUsersErr error
	allUsers     []int64
	allUsersErr  error
}

func (m *mockRulesRepo) GetMatchingRulesForCorporation(ctx context.Context, corpID int64) ([]*models.ContactRule, error) {
	return m.corpRules, m.corpRulesErr
}
func (m *mockRulesRepo) GetMatchingRulesForAlliance(ctx context.Context, allianceID int64) ([]*models.ContactRule, error) {
	return m.alliRules, m.alliRulesErr
}
func (m *mockRulesRepo) GetEveryoneRules(ctx context.Context) ([]*models.ContactRule, error) {
	return m.everyoneRules, m.everyoneErr
}
func (m *mockRulesRepo) GetUsersForCorporation(ctx context.Context, corpID int64, excludeUserID int64) ([]int64, error) {
	return m.corpUsers, m.corpUsersErr
}
func (m *mockRulesRepo) GetUsersForAlliance(ctx context.Context, allianceID int64, excludeUserID int64) ([]int64, error) {
	return m.alliUsers, m.alliUsersErr
}
func (m *mockRulesRepo) GetAllUsers(ctx context.Context, excludeUserID int64) ([]int64, error) {
	return m.allUsers, m.allUsersErr
}
func (m *mockRulesRepo) DeleteAutoContactsForRule(ctx context.Context, ruleID int64) error {
	return nil
}

// Mock permissions repository
type mockPermsRepo struct {
	initErr     error
	upsertErr   error
	initCalls   int
	upsertCalls int
}

func (m *mockPermsRepo) InitializePermissionsForContact(ctx context.Context, tx *sql.Tx, contactID, userID1, userID2 int64) error {
	m.initCalls++
	return m.initErr
}
func (m *mockPermsRepo) UpsertInTx(ctx context.Context, tx *sql.Tx, perm *models.ContactPermission) error {
	m.upsertCalls++
	return m.upsertErr
}

// Helper to create a sqlmock DB that expects N transactions (begin + commit)
func newMockDB(t *testing.T, txCount int) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	for i := 0; i < txCount; i++ {
		mock.ExpectBegin()
		mock.ExpectCommit()
	}
	return db, mock
}

// --- ApplyRule tests ---

func Test_ContactRulesUpdater_ApplyRule_Corporation(t *testing.T) {
	db, dbMock := newMockDB(t, 2)
	defer db.Close()

	entityID := int64(2001)
	rule := &models.ContactRule{
		ID: 1, UserID: 100, RuleType: "corporation", EntityID: &entityID,
	}

	contactsRepo := &mockContactsRepo{}
	rulesRepo := &mockRulesRepo{corpUsers: []int64{200, 300}}
	permsRepo := &mockPermsRepo{}

	updater := updaters.NewContactRules(contactsRepo, rulesRepo, permsRepo, db)
	err := updater.ApplyRule(context.Background(), rule)

	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
	assert.Len(t, contactsRepo.calls, 2)
	assert.Equal(t, int64(100), contactsRepo.calls[0].RequesterID)
	assert.Equal(t, int64(200), contactsRepo.calls[0].RecipientID)
	assert.Equal(t, int64(1), contactsRepo.calls[0].RuleID)
	assert.Equal(t, int64(300), contactsRepo.calls[1].RecipientID)
	assert.Equal(t, 2, permsRepo.initCalls)
	assert.Equal(t, 2, permsRepo.upsertCalls)
}

func Test_ContactRulesUpdater_ApplyRule_Alliance(t *testing.T) {
	db, dbMock := newMockDB(t, 1)
	defer db.Close()

	entityID := int64(5001)
	rule := &models.ContactRule{
		ID: 2, UserID: 100, RuleType: "alliance", EntityID: &entityID,
	}

	contactsRepo := &mockContactsRepo{}
	rulesRepo := &mockRulesRepo{alliUsers: []int64{400}}
	permsRepo := &mockPermsRepo{}

	updater := updaters.NewContactRules(contactsRepo, rulesRepo, permsRepo, db)
	err := updater.ApplyRule(context.Background(), rule)

	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
	assert.Len(t, contactsRepo.calls, 1)
	assert.Equal(t, int64(400), contactsRepo.calls[0].RecipientID)
}

func Test_ContactRulesUpdater_ApplyRule_Everyone(t *testing.T) {
	db, dbMock := newMockDB(t, 3)
	defer db.Close()

	rule := &models.ContactRule{ID: 3, UserID: 100, RuleType: "everyone"}

	contactsRepo := &mockContactsRepo{}
	rulesRepo := &mockRulesRepo{allUsers: []int64{200, 300, 400}}
	permsRepo := &mockPermsRepo{}

	updater := updaters.NewContactRules(contactsRepo, rulesRepo, permsRepo, db)
	err := updater.ApplyRule(context.Background(), rule)

	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
	assert.Len(t, contactsRepo.calls, 3)
}

func Test_ContactRulesUpdater_ApplyRule_CorporationMissingEntityID(t *testing.T) {
	rule := &models.ContactRule{ID: 1, UserID: 100, RuleType: "corporation", EntityID: nil}

	updater := updaters.NewContactRules(&mockContactsRepo{}, &mockRulesRepo{}, &mockPermsRepo{}, nil)
	err := updater.ApplyRule(context.Background(), rule)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "entity_id required")
}

func Test_ContactRulesUpdater_ApplyRule_AllianceMissingEntityID(t *testing.T) {
	rule := &models.ContactRule{ID: 1, UserID: 100, RuleType: "alliance", EntityID: nil}

	updater := updaters.NewContactRules(&mockContactsRepo{}, &mockRulesRepo{}, &mockPermsRepo{}, nil)
	err := updater.ApplyRule(context.Background(), rule)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "entity_id required")
}

func Test_ContactRulesUpdater_ApplyRule_UnknownRuleType(t *testing.T) {
	rule := &models.ContactRule{ID: 1, UserID: 100, RuleType: "invalid"}

	updater := updaters.NewContactRules(&mockContactsRepo{}, &mockRulesRepo{}, &mockPermsRepo{}, nil)
	err := updater.ApplyRule(context.Background(), rule)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown rule type")
}

func Test_ContactRulesUpdater_ApplyRule_GetUsersError(t *testing.T) {
	entityID := int64(2001)
	rule := &models.ContactRule{ID: 1, UserID: 100, RuleType: "corporation", EntityID: &entityID}

	rulesRepo := &mockRulesRepo{corpUsersErr: errors.New("database error")}

	updater := updaters.NewContactRules(&mockContactsRepo{}, rulesRepo, &mockPermsRepo{}, nil)
	err := updater.ApplyRule(context.Background(), rule)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get matching users")
}

func Test_ContactRulesUpdater_ApplyRule_NoMatchingUsers(t *testing.T) {
	entityID := int64(2001)
	rule := &models.ContactRule{ID: 1, UserID: 100, RuleType: "corporation", EntityID: &entityID}

	contactsRepo := &mockContactsRepo{}
	rulesRepo := &mockRulesRepo{corpUsers: []int64{}}

	updater := updaters.NewContactRules(contactsRepo, rulesRepo, &mockPermsRepo{}, nil)
	err := updater.ApplyRule(context.Background(), rule)

	assert.NoError(t, err)
	assert.Len(t, contactsRepo.calls, 0)
}

func Test_ContactRulesUpdater_ApplyRule_SkippedContact(t *testing.T) {
	db, _ := newMockDB(t, 1)
	defer db.Close()

	entityID := int64(2001)
	rule := &models.ContactRule{ID: 1, UserID: 100, RuleType: "corporation", EntityID: &entityID}

	contactsRepo := &mockContactsRepo{
		createAutoContactFn: func(ctx context.Context, tx *sql.Tx, requesterID, recipientID, ruleID int64) (int64, bool, error) {
			return 0, false, nil // Skipped (pending/rejected)
		},
	}
	rulesRepo := &mockRulesRepo{corpUsers: []int64{200}}
	permsRepo := &mockPermsRepo{}

	updater := updaters.NewContactRules(contactsRepo, rulesRepo, permsRepo, db)
	err := updater.ApplyRule(context.Background(), rule)

	assert.NoError(t, err)
	assert.Equal(t, 0, permsRepo.initCalls)
	assert.Equal(t, 0, permsRepo.upsertCalls)
}

func Test_ContactRulesUpdater_ApplyRule_ExistingAcceptedContact(t *testing.T) {
	db, dbMock := newMockDB(t, 1)
	defer db.Close()

	entityID := int64(2001)
	rule := &models.ContactRule{ID: 1, UserID: 100, RuleType: "corporation", EntityID: &entityID}

	contactsRepo := &mockContactsRepo{
		createAutoContactFn: func(ctx context.Context, tx *sql.Tx, requesterID, recipientID, ruleID int64) (int64, bool, error) {
			return 50, false, nil // Existing accepted contact
		},
	}
	rulesRepo := &mockRulesRepo{corpUsers: []int64{200}}
	permsRepo := &mockPermsRepo{}

	updater := updaters.NewContactRules(contactsRepo, rulesRepo, permsRepo, db)
	err := updater.ApplyRule(context.Background(), rule)

	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
	assert.Equal(t, 0, permsRepo.initCalls)   // Not new, skip init
	assert.Equal(t, 1, permsRepo.upsertCalls)  // Still grant permission
}

// --- ApplyRulesForNewCorporation tests ---

func Test_ContactRulesUpdater_ApplyRulesForNewCorporation_CorpRule(t *testing.T) {
	db, dbMock := newMockDB(t, 1)
	defer db.Close()

	contactsRepo := &mockContactsRepo{}
	rulesRepo := &mockRulesRepo{
		corpRules:     []*models.ContactRule{{ID: 1, UserID: 100, RuleType: "corporation"}},
		everyoneRules: []*models.ContactRule{},
	}
	permsRepo := &mockPermsRepo{}

	updater := updaters.NewContactRules(contactsRepo, rulesRepo, permsRepo, db)
	err := updater.ApplyRulesForNewCorporation(context.Background(), 200, 2001, 0)

	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
	assert.Len(t, contactsRepo.calls, 1)
	assert.Equal(t, int64(100), contactsRepo.calls[0].RequesterID)
	assert.Equal(t, int64(200), contactsRepo.calls[0].RecipientID)
}

func Test_ContactRulesUpdater_ApplyRulesForNewCorporation_SkipsSelfRules(t *testing.T) {
	contactsRepo := &mockContactsRepo{}
	rulesRepo := &mockRulesRepo{
		corpRules:     []*models.ContactRule{{ID: 1, UserID: 200, RuleType: "corporation"}},
		everyoneRules: []*models.ContactRule{},
	}

	updater := updaters.NewContactRules(contactsRepo, rulesRepo, &mockPermsRepo{}, nil)
	err := updater.ApplyRulesForNewCorporation(context.Background(), 200, 2001, 0)

	assert.NoError(t, err)
	assert.Len(t, contactsRepo.calls, 0)
}

func Test_ContactRulesUpdater_ApplyRulesForNewCorporation_AllianceRule(t *testing.T) {
	db, dbMock := newMockDB(t, 1)
	defer db.Close()

	contactsRepo := &mockContactsRepo{}
	rulesRepo := &mockRulesRepo{
		corpRules:     []*models.ContactRule{},
		alliRules:     []*models.ContactRule{{ID: 2, UserID: 100, RuleType: "alliance"}},
		everyoneRules: []*models.ContactRule{},
	}

	updater := updaters.NewContactRules(contactsRepo, rulesRepo, &mockPermsRepo{}, db)
	err := updater.ApplyRulesForNewCorporation(context.Background(), 200, 2001, 5001)

	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
	assert.Len(t, contactsRepo.calls, 1)
	assert.Equal(t, int64(100), contactsRepo.calls[0].RequesterID)
}

func Test_ContactRulesUpdater_ApplyRulesForNewCorporation_SkipsAllianceWhenZero(t *testing.T) {
	contactsRepo := &mockContactsRepo{}
	rulesRepo := &mockRulesRepo{
		corpRules:     []*models.ContactRule{},
		everyoneRules: []*models.ContactRule{},
	}

	updater := updaters.NewContactRules(contactsRepo, rulesRepo, &mockPermsRepo{}, nil)
	err := updater.ApplyRulesForNewCorporation(context.Background(), 200, 2001, 0)

	assert.NoError(t, err)
	assert.Len(t, contactsRepo.calls, 0)
}

func Test_ContactRulesUpdater_ApplyRulesForNewCorporation_EveryoneRule(t *testing.T) {
	db, dbMock := newMockDB(t, 1)
	defer db.Close()

	contactsRepo := &mockContactsRepo{}
	rulesRepo := &mockRulesRepo{
		corpRules:     []*models.ContactRule{},
		everyoneRules: []*models.ContactRule{{ID: 3, UserID: 100, RuleType: "everyone"}},
	}

	updater := updaters.NewContactRules(contactsRepo, rulesRepo, &mockPermsRepo{}, db)
	err := updater.ApplyRulesForNewCorporation(context.Background(), 200, 2001, 0)

	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
	assert.Len(t, contactsRepo.calls, 1)
}

func Test_ContactRulesUpdater_ApplyRulesForNewCorporation_CorpRulesError(t *testing.T) {
	rulesRepo := &mockRulesRepo{corpRulesErr: errors.New("database error")}

	updater := updaters.NewContactRules(&mockContactsRepo{}, rulesRepo, &mockPermsRepo{}, nil)
	err := updater.ApplyRulesForNewCorporation(context.Background(), 200, 2001, 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get matching corporation rules")
}

func Test_ContactRulesUpdater_ApplyRulesForNewCorporation_AllianceRulesError(t *testing.T) {
	rulesRepo := &mockRulesRepo{
		corpRules:    []*models.ContactRule{},
		alliRulesErr: errors.New("database error"),
	}

	updater := updaters.NewContactRules(&mockContactsRepo{}, rulesRepo, &mockPermsRepo{}, nil)
	err := updater.ApplyRulesForNewCorporation(context.Background(), 200, 2001, 5001)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get matching alliance rules")
}

func Test_ContactRulesUpdater_ApplyRulesForNewCorporation_EveryoneRulesError(t *testing.T) {
	rulesRepo := &mockRulesRepo{
		corpRules:   []*models.ContactRule{},
		everyoneErr: errors.New("database error"),
	}

	updater := updaters.NewContactRules(&mockContactsRepo{}, rulesRepo, &mockPermsRepo{}, nil)
	err := updater.ApplyRulesForNewCorporation(context.Background(), 200, 2001, 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get everyone rules")
}

func Test_ContactRulesUpdater_ApplyRulesForNewCorporation_MultipleRuleTypes(t *testing.T) {
	db, dbMock := newMockDB(t, 3)
	defer db.Close()

	contactsRepo := &mockContactsRepo{}
	rulesRepo := &mockRulesRepo{
		corpRules:     []*models.ContactRule{{ID: 1, UserID: 100}},
		alliRules:     []*models.ContactRule{{ID: 2, UserID: 300}},
		everyoneRules: []*models.ContactRule{{ID: 3, UserID: 400}},
	}

	updater := updaters.NewContactRules(contactsRepo, rulesRepo, &mockPermsRepo{}, db)
	err := updater.ApplyRulesForNewCorporation(context.Background(), 200, 2001, 5001)

	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
	assert.Len(t, contactsRepo.calls, 3)
}
