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

// --- Mock ---

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

func defaultArbiterSettings() *models.ArbiterSettings {
	return &models.ArbiterSettings{
		UserID:             1,
		ReactionStructure:  "athanor",
		ReactionRig:        "t1",
		ReactionSecurity:   "null",
		InventionStructure: "raitaru",
		InventionRig:       "t1",
		InventionSecurity:  "high",
		ComponentStructure: "raitaru",
		ComponentRig:       "t2",
		ComponentSecurity:  "null",
		FinalStructure:     "raitaru",
		FinalRig:           "t2",
		FinalSecurity:      "null",
	}
}
