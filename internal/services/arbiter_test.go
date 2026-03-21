package services_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- Mock for ArbiterScanRepository ---

type MockArbiterScanRepository struct {
	mock.Mock
}

func (m *MockArbiterScanRepository) GetT2BlueprintsForScan(ctx context.Context) ([]*models.T2BlueprintScanItem, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.T2BlueprintScanItem), args.Error(1)
}

func (m *MockArbiterScanRepository) GetDecryptors(ctx context.Context) ([]*models.Decryptor, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Decryptor), args.Error(1)
}

func (m *MockArbiterScanRepository) GetMarketPricesForTypes(ctx context.Context, typeIDs []int64) (map[int64]*models.MarketPrice, error) {
	args := m.Called(ctx, typeIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int64]*models.MarketPrice), args.Error(1)
}

func (m *MockArbiterScanRepository) GetBlueprintMaterialsForActivity(ctx context.Context, blueprintTypeID int64, activity string) ([]*models.BlueprintMaterial, error) {
	args := m.Called(ctx, blueprintTypeID, activity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.BlueprintMaterial), args.Error(1)
}

func (m *MockArbiterScanRepository) GetBlueprintProductForActivity(ctx context.Context, blueprintTypeID int64, activity string) (*models.BlueprintProduct, error) {
	args := m.Called(ctx, blueprintTypeID, activity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.BlueprintProduct), args.Error(1)
}

func (m *MockArbiterScanRepository) GetBlueprintActivityTime(ctx context.Context, blueprintTypeID int64, activity string) (int64, error) {
	args := m.Called(ctx, blueprintTypeID, activity)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockArbiterScanRepository) GetBestInventionCharacter(ctx context.Context, userID int64, blueprintTypeID int64) (*models.InventionCharacter, error) {
	args := m.Called(ctx, userID, blueprintTypeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.InventionCharacter), args.Error(1)
}

func (m *MockArbiterScanRepository) GetCostIndexForSystem(ctx context.Context, systemID int64, activity string) (float64, error) {
	args := m.Called(ctx, systemID, activity)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockArbiterScanRepository) GetAdjustedPricesForTypes(ctx context.Context, typeIDs []int64) (map[int64]float64, error) {
	args := m.Called(ctx, typeIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int64]float64), args.Error(1)
}

func (m *MockArbiterScanRepository) GetDemandStats(ctx context.Context, typeIDs []int64) (map[int64]*models.DemandStats, error) {
	args := m.Called(ctx, typeIDs)
	if args.Get(0) == nil {
		return map[int64]*models.DemandStats{}, args.Error(1)
	}
	return args.Get(0).(map[int64]*models.DemandStats), args.Error(1)
}

func (m *MockArbiterScanRepository) GetBlueprintForProduct(ctx context.Context, productTypeID int64) (int64, error) {
	args := m.Called(ctx, productTypeID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockArbiterScanRepository) GetReactionBlueprintForProduct(ctx context.Context, productTypeID int64) (int64, error) {
	args := m.Called(ctx, productTypeID)
	return args.Get(0).(int64), args.Error(1)
}

// Ensure interface is satisfied
var _ services.ArbiterScanRepository = &MockArbiterScanRepository{}

// --- Mock for ArbiterBOMRepository ---

type MockArbiterBOMRepository struct {
	mock.Mock
}

func (m *MockArbiterBOMRepository) GetBlueprintMaterialsForActivity(ctx context.Context, blueprintTypeID int64, activity string) ([]*models.BlueprintMaterial, error) {
	args := m.Called(ctx, blueprintTypeID, activity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.BlueprintMaterial), args.Error(1)
}

func (m *MockArbiterBOMRepository) GetBlueprintForProduct(ctx context.Context, productTypeID int64) (int64, error) {
	args := m.Called(ctx, productTypeID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockArbiterBOMRepository) GetMarketPricesForTypes(ctx context.Context, typeIDs []int64) (map[int64]*models.MarketPrice, error) {
	args := m.Called(ctx, typeIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int64]*models.MarketPrice), args.Error(1)
}

func (m *MockArbiterBOMRepository) GetAdjustedPricesForTypes(ctx context.Context, typeIDs []int64) (map[int64]float64, error) {
	args := m.Called(ctx, typeIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int64]float64), args.Error(1)
}

func (m *MockArbiterBOMRepository) GetBlueprintActivityTime(ctx context.Context, blueprintTypeID int64, activity string) (int64, error) {
	args := m.Called(ctx, blueprintTypeID, activity)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockArbiterBOMRepository) GetCostIndexForSystem(ctx context.Context, systemID int64, activity string) (float64, error) {
	args := m.Called(ctx, systemID, activity)
	return args.Get(0).(float64), args.Error(1)
}

// Ensure interface is satisfied
var _ services.ArbiterBOMRepository = &MockArbiterBOMRepository{}

// --- ScanOpportunities tests ---

func Test_ScanOpportunities_ReturnsEmptyResult_WhenNoBlueprints(t *testing.T) {
	repo := &MockArbiterScanRepository{}
	settings := defaultArbiterSettings()

	repo.On("GetT2BlueprintsForScan", mock.Anything).Return([]*models.T2BlueprintScanItem{}, nil)

	result, err := services.ScanOpportunities(context.Background(), 1, settings, nil, false, repo)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.TotalScanned)
	assert.NotNil(t, result.Opportunities)
	assert.Empty(t, result.Opportunities)

	repo.AssertExpectations(t)
}

func Test_ScanOpportunities_ReturnsSortedByProfit(t *testing.T) {
	repo := &MockArbiterScanRepository{}
	settings := defaultArbiterSettings()

	sellPrice1 := float64(100_000_000) // 100M
	sellPrice2 := float64(50_000_000)  // 50M

	blueprints := []*models.T2BlueprintScanItem{
		{
			ProductTypeID:       1001,
			ProductName:         "Product A",
			BlueprintTypeID:     2001,
			T1BlueprintTypeID:   3001,
			BaseInventionChance: 0.34,
			BaseResultME:        2,
			BaseResultRuns:      1,
			Category:            "module",
		},
		{
			ProductTypeID:       1002,
			ProductName:         "Product B",
			BlueprintTypeID:     2002,
			T1BlueprintTypeID:   3002,
			BaseInventionChance: 0.34,
			BaseResultME:        2,
			BaseResultRuns:      1,
			Category:            "module",
		},
	}

	repo.On("GetT2BlueprintsForScan", mock.Anything).Return(blueprints, nil)
	repo.On("GetDecryptors", mock.Anything).Return([]*models.Decryptor{}, nil)
	repo.On("GetMarketPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]*models.MarketPrice{
		1001: {TypeID: 1001, SellPrice: &sellPrice1},
		1002: {TypeID: 1002, SellPrice: &sellPrice2},
	}, nil)
	repo.On("GetAdjustedPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]float64{}, nil)
	repo.On("GetDemandStats", mock.Anything, mock.Anything).Return(map[int64]*models.DemandStats{}, nil)
	repo.On("GetBestInventionCharacter", mock.Anything, mock.Anything, int64(3001)).Return((*models.InventionCharacter)(nil), nil)

	// Invention materials for both blueprints
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(3001), "invention").Return([]*models.BlueprintMaterial{}, nil)
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(3002), "invention").Return([]*models.BlueprintMaterial{}, nil)

	// Manufacturing materials for final products
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(2001), "manufacturing").Return([]*models.BlueprintMaterial{}, nil)
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(2002), "manufacturing").Return([]*models.BlueprintMaterial{}, nil)

	repo.On("GetBlueprintProductForActivity", mock.Anything, mock.Anything, mock.Anything).Return((*models.BlueprintProduct)(nil), nil).Maybe()

	repo.On("GetBlueprintActivityTime", mock.Anything, int64(2001), "manufacturing").Return(int64(86400), nil)
	repo.On("GetBlueprintActivityTime", mock.Anything, int64(2002), "manufacturing").Return(int64(86400), nil)

	result, err := services.ScanOpportunities(context.Background(), 1, settings, nil, false, repo)
	require.NoError(t, err)
	assert.Equal(t, 2, result.TotalScanned)
	assert.Len(t, result.Opportunities, 2)

	// Product A (100M sell price) should have higher profit than Product B (50M)
	if len(result.Opportunities) >= 2 {
		assert.Equal(t, int64(1001), result.Opportunities[0].ProductTypeID,
			"Product A should rank first due to higher sell price with same costs")
	}

	repo.AssertExpectations(t)
}

func Test_ScanOpportunities_NoDecryptorOption_IncludedByDefault(t *testing.T) {
	repo := &MockArbiterScanRepository{}
	settings := defaultArbiterSettings()

	sellPrice := float64(50_000_000)

	blueprint := &models.T2BlueprintScanItem{
		ProductTypeID:       1001,
		ProductName:         "Test Module",
		BlueprintTypeID:     2001,
		T1BlueprintTypeID:   3001,
		BaseInventionChance: 0.34,
		BaseResultME:        2,
		BaseResultRuns:      1,
		Category:            "module",
	}

	repo.On("GetT2BlueprintsForScan", mock.Anything).Return([]*models.T2BlueprintScanItem{blueprint}, nil)
	repo.On("GetDecryptors", mock.Anything).Return([]*models.Decryptor{}, nil) // no decryptors
	repo.On("GetMarketPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]*models.MarketPrice{
		1001: {TypeID: 1001, SellPrice: &sellPrice},
	}, nil)
	repo.On("GetAdjustedPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]float64{}, nil)
	repo.On("GetDemandStats", mock.Anything, mock.Anything).Return(map[int64]*models.DemandStats{}, nil)
	repo.On("GetBestInventionCharacter", mock.Anything, mock.Anything, int64(3001)).Return((*models.InventionCharacter)(nil), nil)
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(3001), "invention").Return([]*models.BlueprintMaterial{}, nil)
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(2001), "manufacturing").Return([]*models.BlueprintMaterial{}, nil)
	repo.On("GetBlueprintProductForActivity", mock.Anything, mock.Anything, mock.Anything).Return((*models.BlueprintProduct)(nil), nil).Maybe()
	repo.On("GetBlueprintActivityTime", mock.Anything, int64(2001), "manufacturing").Return(int64(86400), nil)

	result, err := services.ScanOpportunities(context.Background(), 1, settings, nil, false, repo)
	require.NoError(t, err)
	require.Len(t, result.Opportunities, 1)

	opp := result.Opportunities[0]
	assert.Equal(t, int64(1001), opp.ProductTypeID)
	// Should have exactly 1 option: the no-decryptor option
	assert.Len(t, opp.AllDecryptors, 1)
	assert.Nil(t, opp.AllDecryptors[0].TypeID, "no-decryptor option should have nil TypeID")
	assert.Equal(t, "No Decryptor", opp.AllDecryptors[0].Name)

	repo.AssertExpectations(t)
}

func Test_ScanOpportunities_TaxProfile_AffectsProfitCalculation(t *testing.T) {
	repo := &MockArbiterScanRepository{}
	settings := defaultArbiterSettings()

	sellPrice := float64(10_000_000) // 10M sell price
	taxProfile := &models.ArbiterTaxProfile{
		InputPriceType:  "sell",
		OutputPriceType: "sell",
		SalesTaxRate:    10.0, // 10% sales tax for easy math
		BrokerFeeRate:   0.0,
	}

	blueprint := &models.T2BlueprintScanItem{
		ProductTypeID:       1001,
		ProductName:         "Test Module",
		BlueprintTypeID:     2001,
		T1BlueprintTypeID:   3001,
		BaseInventionChance: 1.0, // 100% chance for predictable cost
		BaseResultME:        2,
		BaseResultRuns:      1,
		Category:            "module",
	}

	repo.On("GetT2BlueprintsForScan", mock.Anything).Return([]*models.T2BlueprintScanItem{blueprint}, nil)
	repo.On("GetDecryptors", mock.Anything).Return([]*models.Decryptor{}, nil)
	repo.On("GetMarketPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]*models.MarketPrice{
		1001: {TypeID: 1001, SellPrice: &sellPrice},
	}, nil)
	repo.On("GetAdjustedPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]float64{}, nil)
	repo.On("GetDemandStats", mock.Anything, mock.Anything).Return(map[int64]*models.DemandStats{}, nil)
	repo.On("GetBestInventionCharacter", mock.Anything, mock.Anything, int64(3001)).Return((*models.InventionCharacter)(nil), nil)
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(3001), "invention").Return([]*models.BlueprintMaterial{}, nil)
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(2001), "manufacturing").Return([]*models.BlueprintMaterial{}, nil)
	repo.On("GetBlueprintProductForActivity", mock.Anything, mock.Anything, mock.Anything).Return((*models.BlueprintProduct)(nil), nil).Maybe()
	repo.On("GetBlueprintActivityTime", mock.Anything, int64(2001), "manufacturing").Return(int64(86400), nil)

	result, err := services.ScanOpportunities(context.Background(), 1, settings, taxProfile, false, repo)
	require.NoError(t, err)
	require.Len(t, result.Opportunities, 1)

	opp := result.Opportunities[0]
	// Revenue = 10M * 1 run = 10M
	assert.Equal(t, float64(10_000_000), opp.Revenue)
	// SalesTax = 10M * 10% = 1M
	assert.Equal(t, float64(1_000_000), opp.SalesTax)
	// BrokerFee = 0
	assert.Equal(t, float64(0), opp.BrokerFee)
	// Profit = 10M - 0 (total cost, no materials) - 1M - 0 = 9M
	assert.Equal(t, float64(9_000_000), opp.Profit)

	repo.AssertExpectations(t)
}

func Test_ScanOpportunities_MultiRunBPC_ProfitUsesFullRevenue(t *testing.T) {
	repo := &MockArbiterScanRepository{}
	settings := defaultArbiterSettings()

	sellPrice := float64(5_000_000) // 5M per unit

	blueprint := &models.T2BlueprintScanItem{
		ProductTypeID:       1002,
		ProductName:         "Module Batch",
		BlueprintTypeID:     2002,
		T1BlueprintTypeID:   3002,
		BaseInventionChance: 1.0,
		BaseResultME:        2,
		BaseResultRuns:      10, // 10-run BPC
		Category:            "module",
	}

	repo.On("GetT2BlueprintsForScan", mock.Anything).Return([]*models.T2BlueprintScanItem{blueprint}, nil)
	repo.On("GetDecryptors", mock.Anything).Return([]*models.Decryptor{}, nil)
	repo.On("GetMarketPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]*models.MarketPrice{
		1002: {TypeID: 1002, SellPrice: &sellPrice},
	}, nil)
	repo.On("GetAdjustedPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]float64{}, nil)
	repo.On("GetDemandStats", mock.Anything, mock.Anything).Return(map[int64]*models.DemandStats{}, nil)
	repo.On("GetBestInventionCharacter", mock.Anything, mock.Anything, int64(3002)).Return((*models.InventionCharacter)(nil), nil)
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(3002), "invention").Return([]*models.BlueprintMaterial{}, nil)
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(2002), "manufacturing").Return([]*models.BlueprintMaterial{}, nil)
	repo.On("GetBlueprintProductForActivity", mock.Anything, mock.Anything, mock.Anything).Return((*models.BlueprintProduct)(nil), nil).Maybe()
	repo.On("GetBlueprintActivityTime", mock.Anything, int64(2002), "manufacturing").Return(int64(86400), nil)

	// nil taxProfile → defaults: 3.6% sales tax, 0% broker fee
	result, err := services.ScanOpportunities(context.Background(), 1, settings, nil, false, repo)
	require.NoError(t, err)
	require.Len(t, result.Opportunities, 1)

	opp := result.Opportunities[0]
	// Revenue = 5M * 10 runs = 50M
	assert.Equal(t, float64(50_000_000), opp.Revenue)
	// Runs should be 10
	assert.Equal(t, 10, opp.Runs)

	repo.AssertExpectations(t)
}

// --- BuildBOMTree tests ---

func Test_BuildBOMTree_ReturnsLeafNode_WhenNoMaterials(t *testing.T) {
	repo := &MockArbiterBOMRepository{}
	settings := defaultArbiterSettings()

	sellPrice := float64(1_000_000)
	repo.On("GetMarketPricesForTypes", mock.Anything, mock.Anything).Return(
		map[int64]*models.MarketPrice{
			5001: {TypeID: 5001, SellPrice: &sellPrice},
		}, nil)
	repo.On("GetAdjustedPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]float64{}, nil)
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(6001), "manufacturing").Return([]*models.BlueprintMaterial{}, nil)

	tree, err := services.BuildBOMTree(
		context.Background(),
		6001,   // blueprint type ID
		5001,   // product type ID
		"Widget",
		10,     // qty
		2,      // me
		repo,
		settings,
		map[int64]bool{},
		map[int64]bool{},
		map[int64]int64{},
		"sell",
		false,
	)
	require.NoError(t, err)
	require.NotNil(t, tree)
	assert.Equal(t, int64(5001), tree.TypeID)
	assert.Equal(t, int64(10), tree.Quantity)
	assert.Equal(t, "buy", tree.Decision)
	assert.Empty(t, tree.Children)

	repo.AssertExpectations(t)
}

func Test_BuildBOMTree_BlacklistForcesBuy(t *testing.T) {
	repo := &MockArbiterBOMRepository{}
	settings := defaultArbiterSettings()

	sellPrice := float64(5_000_000)
	repo.On("GetMarketPricesForTypes", mock.Anything, mock.Anything).Return(
		map[int64]*models.MarketPrice{
			5002: {TypeID: 5002, SellPrice: &sellPrice},
		}, nil)
	repo.On("GetAdjustedPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]float64{}, nil)
	// Even though blueprint exists, blacklisted items should not recurse
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, mock.Anything, mock.Anything).Return([]*models.BlueprintMaterial{}, nil).Maybe()

	blacklist := map[int64]bool{5002: true}

	tree, err := services.BuildBOMTree(
		context.Background(),
		6002,
		5002,
		"Blacklisted Item",
		1,
		0,
		repo,
		settings,
		blacklist,
		map[int64]bool{},
		map[int64]int64{},
		"sell",
		false,
	)
	require.NoError(t, err)
	require.NotNil(t, tree)
	assert.Equal(t, "buy_override", tree.Decision)
	assert.True(t, tree.IsBlacklisted)
}

func Test_BuildBOMTree_WhitelistForcesBuild(t *testing.T) {
	repo := &MockArbiterBOMRepository{}
	settings := defaultArbiterSettings()

	sellPrice := float64(500_000)
	matSellPrice := float64(100_000)

	repo.On("GetMarketPricesForTypes", mock.Anything, mock.Anything).Return(
		map[int64]*models.MarketPrice{
			5003: {TypeID: 5003, SellPrice: &sellPrice},
			5004: {TypeID: 5004, SellPrice: &matSellPrice},
		}, nil)
	repo.On("GetAdjustedPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]float64{}, nil)
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(6003), "manufacturing").Return(
		[]*models.BlueprintMaterial{
			{TypeID: 5004, TypeName: "Component", Quantity: 2},
		}, nil)
	// No sub-blueprint for the component
	repo.On("GetBlueprintForProduct", mock.Anything, int64(5004)).Return(int64(0), nil)

	whitelist := map[int64]bool{5003: true}

	tree, err := services.BuildBOMTree(
		context.Background(),
		6003,
		5003,
		"Whitelisted Item",
		1,
		0,
		repo,
		settings,
		map[int64]bool{},
		whitelist,
		map[int64]int64{},
		"sell",
		false,
	)
	require.NoError(t, err)
	require.NotNil(t, tree)
	assert.Equal(t, "build_override", tree.Decision)
	assert.True(t, tree.IsWhitelisted)
	assert.Len(t, tree.Children, 1)

	repo.AssertExpectations(t)
}

func Test_BuildBOMTree_ChoosesBuild_WhenCheaperThanBuy(t *testing.T) {
	repo := &MockArbiterBOMRepository{}
	settings := defaultArbiterSettings()

	// Product costs 10M to buy, but materials cost 5M (cheaper to build)
	productSellPrice := float64(10_000_000)
	matSellPrice := float64(2_500_000) // 2 units * 2.5M = 5M < 10M

	repo.On("GetMarketPricesForTypes", mock.Anything, mock.Anything).Return(
		map[int64]*models.MarketPrice{
			5005: {TypeID: 5005, SellPrice: &productSellPrice},
			5006: {TypeID: 5006, SellPrice: &matSellPrice},
		}, nil)
	repo.On("GetAdjustedPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]float64{}, nil)
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(6005), "manufacturing").Return(
		[]*models.BlueprintMaterial{
			{TypeID: 5006, TypeName: "Cheap Part", Quantity: 2},
		}, nil)
	repo.On("GetBlueprintForProduct", mock.Anything, int64(5006)).Return(int64(0), nil)

	tree, err := services.BuildBOMTree(
		context.Background(),
		6005,
		5005,
		"Expensive Item",
		1,
		0,
		repo,
		settings,
		map[int64]bool{},
		map[int64]bool{},
		map[int64]int64{},
		"sell",
		false,
	)
	require.NoError(t, err)
	require.NotNil(t, tree)
	assert.Equal(t, "build", tree.Decision)

	repo.AssertExpectations(t)
}

func Test_BuildBOMTree_ChoosesBuy_WhenCheaperThanBuild(t *testing.T) {
	repo := &MockArbiterBOMRepository{}
	settings := defaultArbiterSettings()

	// Product costs 1M to buy, but materials cost 8M (cheaper to buy)
	productSellPrice := float64(1_000_000)
	matSellPrice := float64(4_000_000) // 2 units * 4M = 8M > 1M

	repo.On("GetMarketPricesForTypes", mock.Anything, mock.Anything).Return(
		map[int64]*models.MarketPrice{
			5007: {TypeID: 5007, SellPrice: &productSellPrice},
			5008: {TypeID: 5008, SellPrice: &matSellPrice},
		}, nil)
	repo.On("GetAdjustedPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]float64{}, nil)
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(6007), "manufacturing").Return(
		[]*models.BlueprintMaterial{
			{TypeID: 5008, TypeName: "Expensive Part", Quantity: 2},
		}, nil)
	repo.On("GetBlueprintForProduct", mock.Anything, int64(5008)).Return(int64(0), nil)

	tree, err := services.BuildBOMTree(
		context.Background(),
		6007,
		5007,
		"Cheap Item",
		1,
		0,
		repo,
		settings,
		map[int64]bool{},
		map[int64]bool{},
		map[int64]int64{},
		"sell",
		false,
	)
	require.NoError(t, err)
	require.NotNil(t, tree)
	assert.Equal(t, "buy", tree.Decision)

	repo.AssertExpectations(t)
}

func Test_BuildBOMTree_AvailableAssets_ReducesDelta(t *testing.T) {
	repo := &MockArbiterBOMRepository{}
	settings := defaultArbiterSettings()

	sellPrice := float64(1_000_000)
	repo.On("GetMarketPricesForTypes", mock.Anything, mock.Anything).Return(
		map[int64]*models.MarketPrice{
			5009: {TypeID: 5009, SellPrice: &sellPrice},
		}, nil)
	repo.On("GetAdjustedPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]float64{}, nil)
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(6009), "manufacturing").Return([]*models.BlueprintMaterial{}, nil)

	// 5 units available, need 10 → delta = 5
	assets := map[int64]int64{5009: 5}

	tree, err := services.BuildBOMTree(
		context.Background(),
		6009,
		5009,
		"Partially Available",
		10,
		0,
		repo,
		settings,
		map[int64]bool{},
		map[int64]bool{},
		assets,
		"sell",
		false,
	)
	require.NoError(t, err)
	require.NotNil(t, tree)
	assert.Equal(t, int64(5), tree.Available)
	assert.Equal(t, int64(10), tree.Needed)
	assert.Equal(t, int64(5), tree.Delta)

	repo.AssertExpectations(t)
}

func Test_BuildBOMTree_DepthLimit_StopsRecursion(t *testing.T) {
	repo := &MockArbiterBOMRepository{}
	settings := defaultArbiterSettings()

	// Every type has a blueprint → would recurse forever without depth limit
	sellPrice := float64(1_000)
	for i := int64(5100); i <= 5110; i++ {
		p := sellPrice
		repo.On("GetMarketPricesForTypes", mock.Anything, mock.Anything).Return(
			map[int64]*models.MarketPrice{i: {TypeID: i, SellPrice: &p}}, nil).Maybe()
	}
	repo.On("GetAdjustedPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]float64{}, nil)
	repo.On("GetDemandStats", mock.Anything, mock.Anything).Return(map[int64]*models.DemandStats{}, nil).Maybe()

	// Blueprint 6100 produces 5100, which needs 5101, which needs 5102, ... chain
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(6100), "manufacturing").Return(
		[]*models.BlueprintMaterial{{TypeID: 5101, TypeName: "Sub1", Quantity: 1}}, nil)
	// Sub-items also have blueprints but return empty materials (simulating any nesting)
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, mock.Anything, "manufacturing").Return(
		[]*models.BlueprintMaterial{}, nil).Maybe()
	// Return a blueprint for sub-items (triggers recursion)
	repo.On("GetBlueprintForProduct", mock.Anything, mock.Anything).Return(int64(7000), nil).Maybe()

	// Should complete without stack overflow due to depth limit
	tree, err := services.BuildBOMTree(
		context.Background(),
		6100,
		5100,
		"Deep Item",
		1,
		0,
		repo,
		settings,
		map[int64]bool{},
		map[int64]bool{},
		map[int64]int64{},
		"sell",
		false,
	)
	require.NoError(t, err)
	require.NotNil(t, tree)
}

func Test_BuildBOMTree_BuildAll_ForcesAllNodesToBuild(t *testing.T) {
	repo := &MockArbiterBOMRepository{}
	settings := defaultArbiterSettings()

	// Product would normally be cheaper to buy (1M buy vs 8M build)
	productSellPrice := float64(1_000_000)
	matSellPrice := float64(4_000_000) // 2 units * 4M = 8M > 1M → normally "buy"

	repo.On("GetMarketPricesForTypes", mock.Anything, mock.Anything).Return(
		map[int64]*models.MarketPrice{
			5020: {TypeID: 5020, SellPrice: &productSellPrice},
			5021: {TypeID: 5021, SellPrice: &matSellPrice},
		}, nil)
	repo.On("GetAdjustedPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]float64{}, nil)
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(6020), "manufacturing").Return(
		[]*models.BlueprintMaterial{
			{TypeID: 5021, TypeName: "Expensive Part", Quantity: 2},
		}, nil)
	repo.On("GetBlueprintForProduct", mock.Anything, int64(5021)).Return(int64(0), nil)

	// With buildAll=true, the root node should be "build_override" even though it's more expensive to build
	tree, err := services.BuildBOMTree(
		context.Background(),
		6020,
		5020,
		"Force Build Item",
		1,
		0,
		repo,
		settings,
		map[int64]bool{},
		map[int64]bool{},
		map[int64]int64{},
		"sell",
		true, // buildAll
	)
	require.NoError(t, err)
	require.NotNil(t, tree)
	assert.Equal(t, "build_override", tree.Decision)

	repo.AssertExpectations(t)
}

func Test_BuildBOMTree_BuildAll_RespectsBlacklist(t *testing.T) {
	repo := &MockArbiterBOMRepository{}
	settings := defaultArbiterSettings()

	sellPrice := float64(1_000_000)
	repo.On("GetMarketPricesForTypes", mock.Anything, mock.Anything).Return(
		map[int64]*models.MarketPrice{
			5022: {TypeID: 5022, SellPrice: &sellPrice},
		}, nil)
	repo.On("GetAdjustedPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]float64{}, nil)

	// Blacklisted item — buildAll should NOT override the blacklist
	blacklist := map[int64]bool{5022: true}

	tree, err := services.BuildBOMTree(
		context.Background(),
		6022,
		5022,
		"Blacklisted Even With BuildAll",
		1,
		0,
		repo,
		settings,
		blacklist,
		map[int64]bool{},
		map[int64]int64{},
		"sell",
		true, // buildAll
	)
	require.NoError(t, err)
	require.NotNil(t, tree)
	// Blacklist takes priority over buildAll
	assert.Equal(t, "buy_override", tree.Decision)
	assert.True(t, tree.IsBlacklisted)

	repo.AssertExpectations(t)
}

// --- calculateFinalBOM recursive chain tests ---

func Test_ScanOpportunities_RecursiveChain_BuildsSubComponents(t *testing.T) {
	// Verify that when a T2 blueprint material has a manufacturing sub-blueprint,
	// the cost reflects building the sub-component rather than buying it at market.
	repo := &MockArbiterScanRepository{}
	settings := defaultArbiterSettings()

	// T2 module: typeID 1001, blueprint 2001, T1 blueprint 3001
	// It requires 1x Component (typeID 4001)
	// Component (4001) has manufacturing blueprint (4002)
	// Component blueprint requires 10x Raw (typeID 5001) — raw has no blueprint
	//
	// At market, Component 4001 costs 5_000_000 each
	// Built from 10x Raw at 100_000 each = 1_000_000 (much cheaper to build)

	productSellPrice := float64(10_000_000)
	compMarketPrice := float64(5_000_000)
	rawPrice := float64(100_000)

	blueprint := &models.T2BlueprintScanItem{
		ProductTypeID:       1001,
		ProductName:         "T2 Module",
		BlueprintTypeID:     2001,
		T1BlueprintTypeID:   3001,
		BaseInventionChance: 1.0,
		BaseResultME:        2,
		BaseResultRuns:      1,
		Category:            "module",
	}

	repo.On("GetT2BlueprintsForScan", mock.Anything).Return([]*models.T2BlueprintScanItem{blueprint}, nil)
	repo.On("GetDecryptors", mock.Anything).Return([]*models.Decryptor{}, nil)
	repo.On("GetMarketPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]*models.MarketPrice{
		1001: {TypeID: 1001, SellPrice: &productSellPrice},
		4001: {TypeID: 4001, SellPrice: &compMarketPrice},
		5001: {TypeID: 5001, SellPrice: &rawPrice},
	}, nil)
	repo.On("GetAdjustedPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]float64{}, nil)
	repo.On("GetDemandStats", mock.Anything, mock.Anything).Return(map[int64]*models.DemandStats{}, nil)
	repo.On("GetBestInventionCharacter", mock.Anything, mock.Anything, int64(3001)).Return((*models.InventionCharacter)(nil), nil)
	repo.On("GetBlueprintProductForActivity", mock.Anything, mock.Anything, mock.Anything).Return((*models.BlueprintProduct)(nil), nil).Maybe()
	repo.On("GetBlueprintActivityTime", mock.Anything, int64(2001), "manufacturing").Return(int64(86400), nil)
	repo.On("GetBlueprintActivityTime", mock.Anything, int64(4002), "manufacturing").Return(int64(3600), nil)

	// Invention materials for T1 blueprint
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(3001), "invention").Return([]*models.BlueprintMaterial{}, nil)

	// T2 blueprint materials: 1x Component 4001
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(2001), "manufacturing").Return(
		[]*models.BlueprintMaterial{
			{TypeID: 4001, TypeName: "T2 Component", Quantity: 1},
		}, nil)

	// Component blueprint (4002) materials: 10x Raw 5001
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(4002), "manufacturing").Return(
		[]*models.BlueprintMaterial{
			{TypeID: 5001, TypeName: "Raw Material", Quantity: 10},
		}, nil)
	// Reaction fallback for sub-blueprints returning no mfg mats
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, mock.Anything, "reaction").Return([]*models.BlueprintMaterial{}, nil).Maybe()

	// Component 4001 has manufacturing blueprint 4002
	repo.On("GetBlueprintForProduct", mock.Anything, int64(4001)).Return(int64(4002), nil)
	// Raw material 5001 has no blueprint
	repo.On("GetBlueprintForProduct", mock.Anything, int64(5001)).Return(int64(0), nil)
	repo.On("GetReactionBlueprintForProduct", mock.Anything, int64(5001)).Return(int64(0), nil)
	repo.On("GetReactionBlueprintForProduct", mock.Anything, int64(4001)).Return(int64(0), nil).Maybe()

	result, err := services.ScanOpportunities(context.Background(), 1, settings, nil, true, repo)
	require.NoError(t, err)
	require.Len(t, result.Opportunities, 1)

	opp := result.Opportunities[0]
	// Material cost should reflect building the component from raw (1_000_000),
	// NOT buying it at market (5_000_000).
	// With ME=2 on T2 BPC and raitaru t2 rig: meFactor ≈ 0.95*0.99*... ≈ some reduction
	// We just verify it's less than the market price of the component (5M).
	assert.Less(t, opp.MaterialCost, float64(5_000_000),
		"Material cost should reflect building sub-component cheaper than buying at market")

	repo.AssertExpectations(t)
}

func Test_ScanOpportunities_RecursiveChain_BuysAtMarket_WhenNoBlueprintFound(t *testing.T) {
	// When a material has no blueprint at all (manufacturing or reaction), it should buy at market.
	repo := &MockArbiterScanRepository{}
	settings := defaultArbiterSettings()

	sellPrice := float64(10_000_000)
	matPrice := float64(3_000_000)

	blueprint := &models.T2BlueprintScanItem{
		ProductTypeID:       1010,
		ProductName:         "T2 Widget",
		BlueprintTypeID:     2010,
		T1BlueprintTypeID:   3010,
		BaseInventionChance: 1.0,
		BaseResultME:        2,
		BaseResultRuns:      1,
		Category:            "module",
	}

	repo.On("GetT2BlueprintsForScan", mock.Anything).Return([]*models.T2BlueprintScanItem{blueprint}, nil)
	repo.On("GetDecryptors", mock.Anything).Return([]*models.Decryptor{}, nil)
	repo.On("GetMarketPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]*models.MarketPrice{
		1010: {TypeID: 1010, SellPrice: &sellPrice},
		6010: {TypeID: 6010, SellPrice: &matPrice},
	}, nil)
	repo.On("GetAdjustedPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]float64{}, nil)
	repo.On("GetDemandStats", mock.Anything, mock.Anything).Return(map[int64]*models.DemandStats{}, nil)
	repo.On("GetBestInventionCharacter", mock.Anything, mock.Anything, int64(3010)).Return((*models.InventionCharacter)(nil), nil)
	repo.On("GetBlueprintProductForActivity", mock.Anything, mock.Anything, mock.Anything).Return((*models.BlueprintProduct)(nil), nil).Maybe()
	repo.On("GetBlueprintActivityTime", mock.Anything, int64(2010), "manufacturing").Return(int64(86400), nil)

	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(3010), "invention").Return([]*models.BlueprintMaterial{}, nil)
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(2010), "manufacturing").Return(
		[]*models.BlueprintMaterial{
			{TypeID: 6010, TypeName: "No-Blueprint Mat", Quantity: 1},
		}, nil)

	// No manufacturing or reaction blueprint for this material
	repo.On("GetBlueprintForProduct", mock.Anything, int64(6010)).Return(int64(0), nil)
	repo.On("GetReactionBlueprintForProduct", mock.Anything, int64(6010)).Return(int64(0), nil)

	result, err := services.ScanOpportunities(context.Background(), 1, settings, nil, true, repo)
	require.NoError(t, err)
	require.Len(t, result.Opportunities, 1)

	// Material cost should equal the market price of the material
	assert.Greater(t, result.Opportunities[0].MaterialCost, float64(0))
	repo.AssertExpectations(t)
}

// Test_ScanOpportunities_ReactionProRating verifies that when a T2 component requires fewer
// units than a reaction run produces, costs are scaled proportionally and not inflated by the
// full run output. E.g. needing 1,000 units from a 10,000-unit reaction run should cost 10%
// of the full run, not 100%.
func Test_ScanOpportunities_ReactionProRating(t *testing.T) {
	// Setup:
	//   T2 product (typeID 7001, blueprint 7002, T1 blueprint 7003)
	//   T2 blueprint requires 1,000x composite mat (typeID 7010)
	//   Composite mat (7010) is produced by reaction blueprint (7011) at 10,000 units/run
	//   Reaction blueprint requires 5x raw input (typeID 7020), market price 1,000 ISK each
	//
	// Full reaction run cost (inputs only, no adjusted prices / job cost):
	//   5 raw × 1,000 ISK = 5,000 ISK per run → produces 10,000 units
	//
	// Pro-rated cost for 1,000 units (10% of run):
	//   5,000 × (1,000/10,000) = 500 ISK
	//
	// Without the fix, we'd pay 5,000 ISK (10× inflation).

	repo := &MockArbiterScanRepository{}
	settings := defaultArbiterSettings()

	productSellPrice := float64(50_000_000) // generous sell price so opportunity appears
	compMarketPrice := float64(10_000_000)  // high market price so build path is chosen
	rawPrice := float64(1_000)

	blueprint := &models.T2BlueprintScanItem{
		ProductTypeID:       7001,
		ProductName:         "T2 Armor Plate",
		BlueprintTypeID:     7002,
		T1BlueprintTypeID:   7003,
		BaseInventionChance: 1.0,
		BaseResultME:        0,
		BaseResultRuns:      1,
		Category:            "module",
	}

	repo.On("GetT2BlueprintsForScan", mock.Anything).Return([]*models.T2BlueprintScanItem{blueprint}, nil)
	repo.On("GetDecryptors", mock.Anything).Return([]*models.Decryptor{}, nil)
	repo.On("GetMarketPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]*models.MarketPrice{
		7001: {TypeID: 7001, SellPrice: &productSellPrice},
		7010: {TypeID: 7010, SellPrice: &compMarketPrice},
		7020: {TypeID: 7020, SellPrice: &rawPrice},
	}, nil)
	repo.On("GetAdjustedPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]float64{}, nil)
	repo.On("GetDemandStats", mock.Anything, mock.Anything).Return(map[int64]*models.DemandStats{}, nil)
	repo.On("GetBestInventionCharacter", mock.Anything, mock.Anything, int64(7003)).Return((*models.InventionCharacter)(nil), nil)
	repo.On("GetBlueprintActivityTime", mock.Anything, int64(7002), "manufacturing").Return(int64(86400), nil)
	repo.On("GetBlueprintActivityTime", mock.Anything, int64(7011), "reaction").Return(int64(3600), nil)

	// Invention materials for T1 blueprint
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(7003), "invention").Return([]*models.BlueprintMaterial{}, nil)

	// T2 blueprint materials: 1,000x composite (7010)
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(7002), "manufacturing").Return(
		[]*models.BlueprintMaterial{
			{TypeID: 7010, TypeName: "Crystalline Carbonide", Quantity: 1000},
		}, nil)

	// Reaction blueprint (7011) requires 5x raw input, produces 10,000 composite per run
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(7011), "reaction").Return(
		[]*models.BlueprintMaterial{
			{TypeID: 7020, TypeName: "Raw Ore", Quantity: 5},
		}, nil)
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, mock.Anything, "manufacturing").Return([]*models.BlueprintMaterial{}, nil).Maybe()

	// The reaction blueprint product: 10,000 units per run
	reactionProduct := &models.BlueprintProduct{TypeID: 7010, Quantity: 10000}
	repo.On("GetBlueprintProductForActivity", mock.Anything, int64(7011), "reaction").Return(reactionProduct, nil)
	repo.On("GetBlueprintProductForActivity", mock.Anything, mock.Anything, mock.Anything).Return((*models.BlueprintProduct)(nil), nil).Maybe()

	// Composite mat (7010) has a reaction blueprint (7011), no manufacturing blueprint
	repo.On("GetBlueprintForProduct", mock.Anything, int64(7010)).Return(int64(0), nil)
	repo.On("GetReactionBlueprintForProduct", mock.Anything, int64(7010)).Return(int64(7011), nil)
	// Raw input (7020) has no blueprint
	repo.On("GetBlueprintForProduct", mock.Anything, int64(7020)).Return(int64(0), nil)
	repo.On("GetReactionBlueprintForProduct", mock.Anything, int64(7020)).Return(int64(0), nil)

	result, err := services.ScanOpportunities(context.Background(), 1, settings, nil, true, repo)
	require.NoError(t, err)
	require.Len(t, result.Opportunities, 1)

	opp := result.Opportunities[0]

	// The reaction produces 10,000 units/run but we only need 1,000.
	// With pro-rating: mat cost ≈ 5 raw × 1,000 ISK × (1,000/10,000) = 500 ISK.
	// Without pro-rating: mat cost ≈ 5 raw × 1,000 ISK = 5,000 ISK (10× too high).
	//
	// The actual cost may include ME reductions and job costs, so we assert it is well
	// below the inflated un-pro-rated ceiling (5,000 ISK), confirming the fix is active.
	const inflatedCeiling = float64(5_000)
	assert.Less(t, opp.MaterialCost, inflatedCeiling,
		"Material cost should be pro-rated to 1,000/10,000 of the reaction run, not the full run cost")
	assert.Greater(t, opp.MaterialCost, float64(0),
		"Material cost should be positive")

	repo.AssertExpectations(t)
}

func Test_ScanOpportunities_BuildIfProfitable_PicksCheaperOption(t *testing.T) {
	// Verify the "build if profitable" logic for buildAll=false:
	// for each sub-component that has a blueprint, compute both the recursive build cost AND
	// the market buy price, and use whichever is cheaper.
	// Also verifies that with buildAll=true, the build cost is always used regardless.
	//
	// Setup:
	//   T2 product (typeID 8001, blueprint 8002, T1 blueprint 8003)
	//   T2 blueprint requires 1x Component (typeID 8010)
	//   Component (8010) has manufacturing blueprint (8011)
	//   Component blueprint requires 5x Raw (typeID 8020) — raw has no blueprint

	settings := defaultArbiterSettings()

	productSellPrice := float64(20_000_000)
	blueprint := &models.T2BlueprintScanItem{
		ProductTypeID:       8001,
		ProductName:         "T2 Test Module",
		BlueprintTypeID:     8002,
		T1BlueprintTypeID:   8003,
		BaseInventionChance: 1.0,
		BaseResultME:        0, // no ME reduction for simple math
		BaseResultRuns:      1,
		Category:            "module",
	}

	// commonMocks wires up everything except market prices for 8010 and 8020,
	// and the sub-blueprint (8011) materials — those differ per scenario.
	commonMocks := func(r *MockArbiterScanRepository, compMarketPrice, rawPrice float64) {
		r.On("GetT2BlueprintsForScan", mock.Anything).Return([]*models.T2BlueprintScanItem{blueprint}, nil)
		r.On("GetDecryptors", mock.Anything).Return([]*models.Decryptor{}, nil)
		r.On("GetMarketPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]*models.MarketPrice{
			8001: {TypeID: 8001, SellPrice: &productSellPrice},
			8010: {TypeID: 8010, SellPrice: &compMarketPrice},
			8020: {TypeID: 8020, SellPrice: &rawPrice},
		}, nil)
		r.On("GetAdjustedPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]float64{}, nil)
		r.On("GetDemandStats", mock.Anything, mock.Anything).Return(map[int64]*models.DemandStats{}, nil)
		r.On("GetBestInventionCharacter", mock.Anything, mock.Anything, int64(8003)).Return((*models.InventionCharacter)(nil), nil)
		r.On("GetBlueprintProductForActivity", mock.Anything, mock.Anything, mock.Anything).Return((*models.BlueprintProduct)(nil), nil).Maybe()
		r.On("GetBlueprintActivityTime", mock.Anything, int64(8002), "manufacturing").Return(int64(86400), nil)
		r.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(8003), "invention").Return([]*models.BlueprintMaterial{}, nil)
		r.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(8002), "manufacturing").Return(
			[]*models.BlueprintMaterial{
				{TypeID: 8010, TypeName: "T2 Component", Quantity: 1},
			}, nil)
		// Component 8010 has manufacturing blueprint 8011
		r.On("GetBlueprintForProduct", mock.Anything, int64(8010)).Return(int64(8011), nil)
		r.On("GetReactionBlueprintForProduct", mock.Anything, mock.Anything).Return(int64(0), nil).Maybe()
		// Sub-blueprint (8011) materials: 5x Raw 8020
		r.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(8011), "manufacturing").Return(
			[]*models.BlueprintMaterial{
				{TypeID: 8020, TypeName: "Raw Material", Quantity: 5},
			}, nil)
		r.On("GetBlueprintActivityTime", mock.Anything, int64(8011), "manufacturing").Return(int64(3600), nil)
		r.On("GetBlueprintForProduct", mock.Anything, int64(8020)).Return(int64(0), nil)
		r.On("GetReactionBlueprintForProduct", mock.Anything, int64(8020)).Return(int64(0), nil)
	}

	// --- Scenario 1: build cost < market → buildAll=false should use build cost ---
	//   Component market price: 2_000_000 ISK
	//   Raw price:                100_000 ISK → build cost = 5 × 100_000 = 500_000 ISK (cheaper)
	t.Run("build cheaper than market — uses build cost", func(t *testing.T) {
		compMarketPrice := float64(2_000_000)
		rawPrice := float64(100_000)

		repo := &MockArbiterScanRepository{}
		commonMocks(repo, compMarketPrice, rawPrice)

		result, err := services.ScanOpportunities(context.Background(), 1, settings, nil, false, repo)
		require.NoError(t, err)
		require.Len(t, result.Opportunities, 1)

		opp := result.Opportunities[0]
		// Build cost is ~500_000 (raw materials only, no system set so no job cost).
		// Material cost should be well below the 2_000_000 market price.
		assert.Less(t, opp.MaterialCost, float64(1_000_000),
			"buildAll=false: build is cheaper, so material cost should use build cost (~500K)")

		repo.AssertExpectations(t)
	})

	// --- Scenario 2: market cheaper than build → buildAll=false should use market price ---
	//   Component market price: 2_000_000 ISK
	//   Raw price:                500_000 ISK → build cost = 5 × 500_000 = 2_500_000 ISK (more expensive)
	t.Run("market cheaper than build — uses market price", func(t *testing.T) {
		compMarketPrice := float64(2_000_000)
		rawPrice := float64(500_000)

		repo := &MockArbiterScanRepository{}
		commonMocks(repo, compMarketPrice, rawPrice)

		result, err := services.ScanOpportunities(context.Background(), 1, settings, nil, false, repo)
		require.NoError(t, err)
		require.Len(t, result.Opportunities, 1)

		opp := result.Opportunities[0]
		// Build cost is ~2_500_000 (more expensive than market at 2_000_000).
		// Material cost should be ~2_000_000 (market price was chosen).
		assert.GreaterOrEqual(t, opp.MaterialCost, float64(1_800_000),
			"buildAll=false: market is cheaper, so material cost should use market price (~2M)")
		assert.Less(t, opp.MaterialCost, float64(2_200_000),
			"buildAll=false: market is cheaper, so material cost should be close to market price (~2M)")

		repo.AssertExpectations(t)
	})

	// --- Scenario 3: buildAll=true always uses build cost regardless of market price ---
	//   Same prices as Scenario 2: market is cheaper, but buildAll forces build cost
	t.Run("buildAll=true always uses build cost", func(t *testing.T) {
		compMarketPrice := float64(2_000_000)
		rawPrice := float64(500_000)

		repo := &MockArbiterScanRepository{}
		commonMocks(repo, compMarketPrice, rawPrice)

		result, err := services.ScanOpportunities(context.Background(), 1, settings, nil, true, repo)
		require.NoError(t, err)
		require.Len(t, result.Opportunities, 1)

		opp := result.Opportunities[0]
		// Build cost is ~2_500_000, market is 2_000_000.
		// With buildAll=true the build cost is always used.
		assert.GreaterOrEqual(t, opp.MaterialCost, float64(2_200_000),
			"buildAll=true: should always use build cost even when market is cheaper")

		repo.AssertExpectations(t)
	})
}

func defaultArbiterSettings() *models.ArbiterSettings {
	return &models.ArbiterSettings{
		UserID:             1,
		ReactionStructure:  "athanor",
		ReactionRig:        "t1",
		InventionStructure: "raitaru",
		InventionRig:       "t1",
		ComponentStructure: "raitaru",
		ComponentRig:       "t2",
		FinalStructure:     "raitaru",
		FinalRig:           "t2",
		UseWhitelist:       true,
		UseBlacklist:       true,
	}
}
