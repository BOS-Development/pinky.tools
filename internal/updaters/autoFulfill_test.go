package updaters_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/updaters"
	"github.com/stretchr/testify/assert"
)

// --- Mocks ---

type mockAutoFulfillBuyOrdersRepo struct {
	allOrders    []*models.BuyOrder
	allOrdersErr error
	userOrders   []*models.BuyOrder
	userErr      error
}

func (m *mockAutoFulfillBuyOrdersRepo) GetAllActiveBuyOrders(ctx context.Context) ([]*models.BuyOrder, error) {
	return m.allOrders, m.allOrdersErr
}

func (m *mockAutoFulfillBuyOrdersRepo) GetActiveBuyOrdersForUser(ctx context.Context, userID int64) ([]*models.BuyOrder, error) {
	return m.userOrders, m.userErr
}

type mockAutoFulfillForSaleRepo struct {
	matchingItems []*models.ForSaleItem
	matchingErr   error
	updatedItems  map[int64]int64 // itemID -> newQuantity
	updateErr     error
	byID          map[int64]*models.ForSaleItem
}

func (m *mockAutoFulfillForSaleRepo) GetMatchingForSaleItems(ctx context.Context, typeID int64, minPrice, maxPrice float64, excludeUserID int64) ([]*models.ForSaleItem, error) {
	return m.matchingItems, m.matchingErr
}

func (m *mockAutoFulfillForSaleRepo) UpdateQuantity(ctx context.Context, tx *sql.Tx, itemID int64, newQuantity int64) error {
	if m.updatedItems == nil {
		m.updatedItems = map[int64]int64{}
	}
	m.updatedItems[itemID] = newQuantity
	return m.updateErr
}

func (m *mockAutoFulfillForSaleRepo) GetByID(ctx context.Context, itemID int64) (*models.ForSaleItem, error) {
	if item, ok := m.byID[itemID]; ok {
		return item, nil
	}
	return nil, fmt.Errorf("not found")
}

type mockAutoFulfillPurchaseRepo struct {
	createdPurchases  []*models.PurchaseTransaction
	createErr         error
	pendingByBuyOrder map[int64]int64 // buyOrderID -> pending qty
	pendingByOrderErr error
}

func (m *mockAutoFulfillPurchaseRepo) CreateAutoFulfill(ctx context.Context, tx *sql.Tx, purchase *models.PurchaseTransaction) error {
	if m.createErr != nil {
		return m.createErr
	}
	cp := *purchase
	cp.ID = int64(len(m.createdPurchases) + 1)
	m.createdPurchases = append(m.createdPurchases, &cp)
	return nil
}

func (m *mockAutoFulfillPurchaseRepo) GetPendingQuantityForBuyOrder(ctx context.Context, buyOrderID int64) (int64, error) {
	if m.pendingByOrderErr != nil {
		return 0, m.pendingByOrderErr
	}
	return m.pendingByBuyOrder[buyOrderID], nil
}

type mockAutoFulfillPermissionsRepo struct {
	allowed    bool
	allowedErr error
}

func (m *mockAutoFulfillPermissionsRepo) CheckPermission(ctx context.Context, grantingUserID, receivingUserID int64, serviceType string) (bool, error) {
	return m.allowed, m.allowedErr
}

type mockAutoFulfillUsersRepo struct {
	names map[int64]string
}

func (m *mockAutoFulfillUsersRepo) GetUserName(ctx context.Context, userID int64) (string, error) {
	if name, ok := m.names[userID]; ok {
		return name, nil
	}
	return "", fmt.Errorf("not found")
}

// --- Helpers ---

// newAutoFulfillUpdaterNoDB creates an updater with nil db — only for tests that don't reach createAutoFulfillPurchase.
func newAutoFulfillUpdaterNoDB(
	buyOrderRepo *mockAutoFulfillBuyOrdersRepo,
	forSaleRepo *mockAutoFulfillForSaleRepo,
	purchaseRepo *mockAutoFulfillPurchaseRepo,
	permRepo *mockAutoFulfillPermissionsRepo,
) *updaters.AutoFulfill {
	return updaters.NewAutoFulfill(
		nil,
		buyOrderRepo,
		forSaleRepo,
		purchaseRepo,
		permRepo,
		&mockAutoFulfillUsersRepo{names: map[int64]string{42: "Buyer", 99: "Seller"}},
		nil,
	)
}

// newAutoFulfillUpdaterWithDB creates an updater with a sqlmock db — for tests that exercise createAutoFulfillPurchase.
func newAutoFulfillUpdaterWithDB(
	t *testing.T,
	buyOrderRepo *mockAutoFulfillBuyOrdersRepo,
	forSaleRepo *mockAutoFulfillForSaleRepo,
	purchaseRepo *mockAutoFulfillPurchaseRepo,
	permRepo *mockAutoFulfillPermissionsRepo,
) (*updaters.AutoFulfill, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	u := updaters.NewAutoFulfill(
		db,
		buyOrderRepo,
		forSaleRepo,
		purchaseRepo,
		permRepo,
		&mockAutoFulfillUsersRepo{names: map[int64]string{42: "Buyer", 99: "Seller"}},
		nil,
	)
	return u, mock
}

// --- SyncForUser Tests ---

func Test_AutoFulfill_SyncForUser_NoOrders(t *testing.T) {
	buyOrderRepo := &mockAutoFulfillBuyOrdersRepo{
		userOrders: []*models.BuyOrder{},
	}
	forSaleRepo := &mockAutoFulfillForSaleRepo{}
	purchaseRepo := &mockAutoFulfillPurchaseRepo{pendingByBuyOrder: map[int64]int64{}}
	permRepo := &mockAutoFulfillPermissionsRepo{allowed: true}

	u := newAutoFulfillUpdaterNoDB(buyOrderRepo, forSaleRepo, purchaseRepo, permRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	assert.Len(t, purchaseRepo.createdPurchases, 0)
}

func Test_AutoFulfill_SyncForUser_GetOrdersError(t *testing.T) {
	buyOrderRepo := &mockAutoFulfillBuyOrdersRepo{
		userErr: fmt.Errorf("db error"),
	}
	forSaleRepo := &mockAutoFulfillForSaleRepo{}
	purchaseRepo := &mockAutoFulfillPurchaseRepo{pendingByBuyOrder: map[int64]int64{}}
	permRepo := &mockAutoFulfillPermissionsRepo{allowed: true}

	u := newAutoFulfillUpdaterNoDB(buyOrderRepo, forSaleRepo, purchaseRepo, permRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get active buy orders for user")
}

func Test_AutoFulfill_SyncForUser_NoMatchingItems(t *testing.T) {
	buyOrderRepo := &mockAutoFulfillBuyOrdersRepo{
		userOrders: []*models.BuyOrder{
			{ID: 1, BuyerUserID: 42, TypeID: 34, QuantityDesired: 100, MinPricePerUnit: 5.0, MaxPricePerUnit: 10.0, IsActive: true},
		},
	}
	forSaleRepo := &mockAutoFulfillForSaleRepo{matchingItems: []*models.ForSaleItem{}}
	purchaseRepo := &mockAutoFulfillPurchaseRepo{pendingByBuyOrder: map[int64]int64{}}
	permRepo := &mockAutoFulfillPermissionsRepo{allowed: true}

	u := newAutoFulfillUpdaterNoDB(buyOrderRepo, forSaleRepo, purchaseRepo, permRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	assert.Len(t, purchaseRepo.createdPurchases, 0)
}

func Test_AutoFulfill_SyncForUser_ZeroMaxPrice(t *testing.T) {
	buyOrderRepo := &mockAutoFulfillBuyOrdersRepo{
		userOrders: []*models.BuyOrder{
			{ID: 1, BuyerUserID: 42, TypeID: 34, QuantityDesired: 100, MaxPricePerUnit: 0, IsActive: true},
		},
	}
	forSaleRepo := &mockAutoFulfillForSaleRepo{}
	purchaseRepo := &mockAutoFulfillPurchaseRepo{pendingByBuyOrder: map[int64]int64{}}
	permRepo := &mockAutoFulfillPermissionsRepo{allowed: true}

	u := newAutoFulfillUpdaterNoDB(buyOrderRepo, forSaleRepo, purchaseRepo, permRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	assert.Len(t, purchaseRepo.createdPurchases, 0)
}

func Test_AutoFulfill_SyncForUser_NoPermission(t *testing.T) {
	buyOrderRepo := &mockAutoFulfillBuyOrdersRepo{
		userOrders: []*models.BuyOrder{
			{ID: 1, BuyerUserID: 42, TypeID: 34, QuantityDesired: 100, MinPricePerUnit: 5.0, MaxPricePerUnit: 10.0, IsActive: true},
		},
	}
	forSaleRepo := &mockAutoFulfillForSaleRepo{
		matchingItems: []*models.ForSaleItem{
			{ID: 1, UserID: 99, TypeID: 34, QuantityAvailable: 500, PricePerUnit: 8.0, IsActive: true},
		},
	}
	purchaseRepo := &mockAutoFulfillPurchaseRepo{pendingByBuyOrder: map[int64]int64{}}
	permRepo := &mockAutoFulfillPermissionsRepo{allowed: false}

	u := newAutoFulfillUpdaterNoDB(buyOrderRepo, forSaleRepo, purchaseRepo, permRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	assert.Len(t, purchaseRepo.createdPurchases, 0)
}

// --- Pending Quantity Awareness Tests ---

func Test_AutoFulfill_PendingPurchasesFullyCoverOrder_NoPurchase(t *testing.T) {
	orderID := int64(1)
	buyOrderRepo := &mockAutoFulfillBuyOrdersRepo{
		userOrders: []*models.BuyOrder{
			{ID: orderID, BuyerUserID: 42, TypeID: 34, QuantityDesired: 100, MinPricePerUnit: 5.0, MaxPricePerUnit: 10.0, IsActive: true},
		},
	}
	forSaleRepo := &mockAutoFulfillForSaleRepo{
		matchingItems: []*models.ForSaleItem{
			{ID: 1, UserID: 99, TypeID: 34, QuantityAvailable: 500, PricePerUnit: 8.0, IsActive: true},
		},
	}
	purchaseRepo := &mockAutoFulfillPurchaseRepo{
		pendingByBuyOrder: map[int64]int64{orderID: 100}, // Fully covered
	}
	permRepo := &mockAutoFulfillPermissionsRepo{allowed: true}

	// No db needed — pending check returns early before createAutoFulfillPurchase
	u := newAutoFulfillUpdaterNoDB(buyOrderRepo, forSaleRepo, purchaseRepo, permRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	assert.Len(t, purchaseRepo.createdPurchases, 0)
}

func Test_AutoFulfill_PendingPurchasesExceedOrder_NoPurchase(t *testing.T) {
	orderID := int64(1)
	buyOrderRepo := &mockAutoFulfillBuyOrdersRepo{
		userOrders: []*models.BuyOrder{
			{ID: orderID, BuyerUserID: 42, TypeID: 34, QuantityDesired: 100, MinPricePerUnit: 5.0, MaxPricePerUnit: 10.0, IsActive: true},
		},
	}
	forSaleRepo := &mockAutoFulfillForSaleRepo{
		matchingItems: []*models.ForSaleItem{
			{ID: 1, UserID: 99, TypeID: 34, QuantityAvailable: 500, PricePerUnit: 8.0, IsActive: true},
		},
	}
	purchaseRepo := &mockAutoFulfillPurchaseRepo{
		pendingByBuyOrder: map[int64]int64{orderID: 150}, // Exceeds desired
	}
	permRepo := &mockAutoFulfillPermissionsRepo{allowed: true}

	u := newAutoFulfillUpdaterNoDB(buyOrderRepo, forSaleRepo, purchaseRepo, permRepo)
	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	assert.Len(t, purchaseRepo.createdPurchases, 0)
}

func Test_AutoFulfill_PendingQueryError_LoggedNotReturned(t *testing.T) {
	orderID := int64(1)
	buyOrderRepo := &mockAutoFulfillBuyOrdersRepo{
		userOrders: []*models.BuyOrder{
			{ID: orderID, BuyerUserID: 42, TypeID: 34, QuantityDesired: 100, MinPricePerUnit: 5.0, MaxPricePerUnit: 10.0, IsActive: true},
		},
	}
	forSaleRepo := &mockAutoFulfillForSaleRepo{
		matchingItems: []*models.ForSaleItem{
			{ID: 1, UserID: 99, TypeID: 34, QuantityAvailable: 500, PricePerUnit: 8.0, IsActive: true},
		},
	}
	purchaseRepo := &mockAutoFulfillPurchaseRepo{
		pendingByOrderErr: fmt.Errorf("db error"),
	}
	permRepo := &mockAutoFulfillPermissionsRepo{allowed: true}

	u := newAutoFulfillUpdaterNoDB(buyOrderRepo, forSaleRepo, purchaseRepo, permRepo)
	err := u.SyncForUser(context.Background(), 42)

	// matchBuyOrder errors are logged, not returned
	assert.NoError(t, err)
	assert.Len(t, purchaseRepo.createdPurchases, 0)
}

func Test_AutoFulfill_PendingPurchasesReduceRemainingQuantity(t *testing.T) {
	orderID := int64(1)
	buyOrderRepo := &mockAutoFulfillBuyOrdersRepo{
		userOrders: []*models.BuyOrder{
			{ID: orderID, BuyerUserID: 42, TypeID: 34, QuantityDesired: 100, MinPricePerUnit: 5.0, MaxPricePerUnit: 10.0, IsActive: true},
		},
	}
	forSaleRepo := &mockAutoFulfillForSaleRepo{
		matchingItems: []*models.ForSaleItem{
			{ID: 1, UserID: 99, TypeID: 34, QuantityAvailable: 500, PricePerUnit: 8.0, IsActive: true},
		},
	}
	purchaseRepo := &mockAutoFulfillPurchaseRepo{
		pendingByBuyOrder: map[int64]int64{orderID: 60}, // 60 already pending
	}
	permRepo := &mockAutoFulfillPermissionsRepo{allowed: true}

	u, mock := newAutoFulfillUpdaterWithDB(t, buyOrderRepo, forSaleRepo, purchaseRepo, permRepo)
	mock.ExpectBegin()
	mock.ExpectCommit()

	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	assert.Len(t, purchaseRepo.createdPurchases, 1)
	// Should only purchase 40 (100 desired - 60 pending)
	assert.Equal(t, int64(40), purchaseRepo.createdPurchases[0].QuantityPurchased)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_AutoFulfill_NoPendingPurchases_FullQuantity(t *testing.T) {
	orderID := int64(1)
	buyOrderRepo := &mockAutoFulfillBuyOrdersRepo{
		userOrders: []*models.BuyOrder{
			{ID: orderID, BuyerUserID: 42, TypeID: 34, QuantityDesired: 100, MinPricePerUnit: 5.0, MaxPricePerUnit: 10.0, IsActive: true},
		},
	}
	forSaleRepo := &mockAutoFulfillForSaleRepo{
		matchingItems: []*models.ForSaleItem{
			{ID: 1, UserID: 99, TypeID: 34, QuantityAvailable: 500, PricePerUnit: 8.0, IsActive: true},
		},
	}
	purchaseRepo := &mockAutoFulfillPurchaseRepo{pendingByBuyOrder: map[int64]int64{}}
	permRepo := &mockAutoFulfillPermissionsRepo{allowed: true}

	u, mock := newAutoFulfillUpdaterWithDB(t, buyOrderRepo, forSaleRepo, purchaseRepo, permRepo)
	mock.ExpectBegin()
	mock.ExpectCommit()

	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	assert.Len(t, purchaseRepo.createdPurchases, 1)
	assert.Equal(t, int64(100), purchaseRepo.createdPurchases[0].QuantityPurchased)
	assert.Equal(t, int64(34), purchaseRepo.createdPurchases[0].TypeID)
	assert.Equal(t, int64(42), purchaseRepo.createdPurchases[0].BuyerUserID)
	assert.Equal(t, int64(99), purchaseRepo.createdPurchases[0].SellerUserID)
	assert.Equal(t, 8.0, purchaseRepo.createdPurchases[0].PricePerUnit)
	assert.Equal(t, 800.0, purchaseRepo.createdPurchases[0].TotalPrice)
	assert.Equal(t, "pending", purchaseRepo.createdPurchases[0].Status)
	assert.True(t, purchaseRepo.createdPurchases[0].IsAutoFulfilled)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_AutoFulfill_PurchaseCappedByAvailableQuantity(t *testing.T) {
	orderID := int64(1)
	buyOrderRepo := &mockAutoFulfillBuyOrdersRepo{
		userOrders: []*models.BuyOrder{
			{ID: orderID, BuyerUserID: 42, TypeID: 34, QuantityDesired: 1000, MinPricePerUnit: 5.0, MaxPricePerUnit: 10.0, IsActive: true},
		},
	}
	forSaleRepo := &mockAutoFulfillForSaleRepo{
		matchingItems: []*models.ForSaleItem{
			{ID: 1, UserID: 99, TypeID: 34, QuantityAvailable: 50, PricePerUnit: 8.0, IsActive: true},
		},
	}
	purchaseRepo := &mockAutoFulfillPurchaseRepo{pendingByBuyOrder: map[int64]int64{}}
	permRepo := &mockAutoFulfillPermissionsRepo{allowed: true}

	u, mock := newAutoFulfillUpdaterWithDB(t, buyOrderRepo, forSaleRepo, purchaseRepo, permRepo)
	mock.ExpectBegin()
	mock.ExpectCommit()

	err := u.SyncForUser(context.Background(), 42)

	assert.NoError(t, err)
	assert.Len(t, purchaseRepo.createdPurchases, 1)
	assert.Equal(t, int64(50), purchaseRepo.createdPurchases[0].QuantityPurchased)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- SyncForAllUsers Tests ---

func Test_AutoFulfill_SyncForAllUsers_NoOrders(t *testing.T) {
	buyOrderRepo := &mockAutoFulfillBuyOrdersRepo{allOrders: []*models.BuyOrder{}}
	forSaleRepo := &mockAutoFulfillForSaleRepo{}
	purchaseRepo := &mockAutoFulfillPurchaseRepo{pendingByBuyOrder: map[int64]int64{}}
	permRepo := &mockAutoFulfillPermissionsRepo{allowed: true}

	u := newAutoFulfillUpdaterNoDB(buyOrderRepo, forSaleRepo, purchaseRepo, permRepo)
	err := u.SyncForAllUsers(context.Background())

	assert.NoError(t, err)
}

func Test_AutoFulfill_SyncForAllUsers_Error(t *testing.T) {
	buyOrderRepo := &mockAutoFulfillBuyOrdersRepo{allOrdersErr: fmt.Errorf("db error")}
	forSaleRepo := &mockAutoFulfillForSaleRepo{}
	purchaseRepo := &mockAutoFulfillPurchaseRepo{pendingByBuyOrder: map[int64]int64{}}
	permRepo := &mockAutoFulfillPermissionsRepo{allowed: true}

	u := newAutoFulfillUpdaterNoDB(buyOrderRepo, forSaleRepo, purchaseRepo, permRepo)
	err := u.SyncForAllUsers(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get all active buy orders")
}

// --- Constructor Test ---

func Test_AutoFulfill_Constructor(t *testing.T) {
	u := updaters.NewAutoFulfill(
		nil,
		&mockAutoFulfillBuyOrdersRepo{},
		&mockAutoFulfillForSaleRepo{},
		&mockAutoFulfillPurchaseRepo{},
		&mockAutoFulfillPermissionsRepo{},
		&mockAutoFulfillUsersRepo{},
		nil,
	)
	assert.NotNil(t, u)
}
