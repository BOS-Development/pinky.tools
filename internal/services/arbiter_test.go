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

	result, err := services.ScanOpportunities(context.Background(), 1, settings, repo)
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

	result, err := services.ScanOpportunities(context.Background(), 1, settings, repo)
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
	repo.On("GetBestInventionCharacter", mock.Anything, mock.Anything, int64(3001)).Return((*models.InventionCharacter)(nil), nil)
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(3001), "invention").Return([]*models.BlueprintMaterial{}, nil)
	repo.On("GetBlueprintMaterialsForActivity", mock.Anything, int64(2001), "manufacturing").Return([]*models.BlueprintMaterial{}, nil)
	repo.On("GetBlueprintProductForActivity", mock.Anything, mock.Anything, mock.Anything).Return((*models.BlueprintProduct)(nil), nil).Maybe()
	repo.On("GetBlueprintActivityTime", mock.Anything, int64(2001), "manufacturing").Return(int64(86400), nil)

	result, err := services.ScanOpportunities(context.Background(), 1, settings, repo)
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
	repo.On("GetAdjustedPricesForTypes", mock.Anything, mock.Anything).Return(map[int64]float64{}, nil).Maybe()

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
