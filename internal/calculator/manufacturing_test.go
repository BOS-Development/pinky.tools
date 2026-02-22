package calculator

import (
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
)

func TestComputeManufacturingME(t *testing.T) {
	tests := []struct {
		name        string
		blueprintME int
		structure   string
		rig         string
		security    string
		expected    float64
	}{
		{"ME10 Sotiyo T2 null", 10, "sotiyo", "t2", "null", (1.0 - 0.10) * (1.0 - 0.01) * (1.0 - 0.024*2.1)},
		{"ME10 Raitaru T2 low", 10, "raitaru", "t2", "low", (1.0 - 0.10) * (1.0 - 0.01) * (1.0 - 0.024*1.9)},
		{"ME10 Azbel T1 null", 10, "azbel", "t1", "null", (1.0 - 0.10) * (1.0 - 0.01) * (1.0 - 0.02*2.1)},
		{"ME0 station none low", 0, "station", "none", "low", 1.0},
		{"ME5 station none low", 5, "station", "none", "low", 0.95},
		{"ME10 station none high", 10, "station", "none", "high", 0.90},
		{"ME10 Sotiyo none null", 10, "sotiyo", "none", "null", (1.0 - 0.10) * (1.0 - 0.01)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ComputeManufacturingME(tc.blueprintME, tc.structure, tc.rig, tc.security)
			assert.InDelta(t, tc.expected, result, 0.0001)
		})
	}
}

func TestEngineeringSecurityMultiplier(t *testing.T) {
	assert.Equal(t, 2.1, EngineeringSecurityMultiplier("null"))
	assert.Equal(t, 1.9, EngineeringSecurityMultiplier("low"))
	assert.Equal(t, 1.0, EngineeringSecurityMultiplier("high"))
}

func TestManufacturingStructureMEValue(t *testing.T) {
	assert.Equal(t, 0.01, ManufacturingStructureMEValue("raitaru"))
	assert.Equal(t, 0.01, ManufacturingStructureMEValue("azbel"))
	assert.Equal(t, 0.01, ManufacturingStructureMEValue("sotiyo"))
	assert.Equal(t, 0.0, ManufacturingStructureMEValue("station"))
	assert.Equal(t, 0.0, ManufacturingStructureMEValue("unknown"))
}

func TestComputeManufacturingTE(t *testing.T) {
	tests := []struct {
		name             string
		blueprintTE      int
		industrySkill    int
		advIndustrySkill int
		structure        string
		rig              string
		security         string
		expected         float64
	}{
		{
			"TE20 Industry5 AdvIndustry5 Raitaru T2 null",
			20, 5, 5, "raitaru", "t2", "null",
			(1.0 - 0.20) * (1.0 - 0.20) * (1.0 - 0.15) * (1.0 - 0.15) * (1.0 - 0.24*2.1),
		},
		{
			"TE0 Industry0 AdvIndustry0 station none high",
			0, 0, 0, "station", "none", "high",
			1.0,
		},
		{
			"TE20 Industry5 AdvIndustry0 Sotiyo T1 low",
			20, 5, 0, "sotiyo", "t1", "low",
			(1.0 - 0.20) * (1.0 - 0.20) * (1.0) * (1.0 - 0.30) * (1.0 - 0.20*1.9),
		},
		{
			"TE10 Industry3 AdvIndustry2 Azbel none high",
			10, 3, 2, "azbel", "none", "high",
			(1.0 - 0.10) * (1.0 - 0.12) * (1.0 - 0.06) * (1.0 - 0.20) * (1.0),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ComputeManufacturingTE(tc.blueprintTE, tc.industrySkill, tc.advIndustrySkill, tc.structure, tc.rig, tc.security)
			assert.InDelta(t, tc.expected, result, 0.0001)
		})
	}
}

func TestManufacturingStructureTEValue(t *testing.T) {
	assert.Equal(t, 0.15, ManufacturingStructureTEValue("raitaru"))
	assert.Equal(t, 0.20, ManufacturingStructureTEValue("azbel"))
	assert.Equal(t, 0.30, ManufacturingStructureTEValue("sotiyo"))
	assert.Equal(t, 0.0, ManufacturingStructureTEValue("station"))
	assert.Equal(t, 0.0, ManufacturingStructureTEValue("unknown"))
}

func TestManufacturingStructureCostBonus(t *testing.T) {
	assert.Equal(t, 0.05, ManufacturingStructureCostBonus("sotiyo"))
	assert.Equal(t, 0.03, ManufacturingStructureCostBonus("azbel"))
	assert.Equal(t, 0.01, ManufacturingStructureCostBonus("raitaru"))
	assert.Equal(t, 0.0, ManufacturingStructureCostBonus("station"))
	assert.Equal(t, 0.0, ManufacturingStructureCostBonus("unknown"))
}

func TestComputeManufacturingJobCost(t *testing.T) {
	materials := []*repositories.ManufacturingMaterialRow{
		{TypeID: 34, Quantity: 1000}, // Tritanium
		{TypeID: 35, Quantity: 500},  // Pyerite
	}

	adjustedPrices := map[int64]float64{
		34: 5.0,  // Tritanium adjusted price
		35: 10.0, // Pyerite adjusted price
	}

	// EIV = (1000 * 5.0) + (500 * 10.0) = 5000 + 5000 = 10000
	// Station (no bonus): EIV*cost_index + EIV*scc + EIV*tax
	// = 10000*0.01 + 10000*0.04 + 10000*0.05 = 100 + 400 + 500 = 1000
	result := ComputeManufacturingJobCost(materials, adjustedPrices, 0.01, 5.0, "station")
	assert.InDelta(t, 1000.0, result, 0.01)

	// Sotiyo (5% bonus on cost index only):
	// = 10000*0.01*0.95 + 10000*0.04 + 10000*0.05 = 95 + 400 + 500 = 995
	result = ComputeManufacturingJobCost(materials, adjustedPrices, 0.01, 5.0, "sotiyo")
	assert.InDelta(t, 995.0, result, 0.01)
}

func TestComputeManufacturingJobCost_MissingPrices(t *testing.T) {
	materials := []*repositories.ManufacturingMaterialRow{
		{TypeID: 34, Quantity: 1000},
		{TypeID: 99, Quantity: 500}, // no adjusted price
	}

	adjustedPrices := map[int64]float64{
		34: 5.0,
	}

	// EIV = (1000 * 5.0) = 5000 (type 99 missing, skipped)
	// Station: 5000*0.02 + 5000*0.04 + 0 = 100 + 200 = 300
	result := ComputeManufacturingJobCost(materials, adjustedPrices, 0.02, 0, "station")
	assert.InDelta(t, 300.0, result, 0.01)
}

func TestCalculateManufacturingJob(t *testing.T) {
	sellPrice34 := 6.0
	sellPrice35 := 12.0
	sellProductPrice := 50000.0

	params := &ManufacturingParams{
		BlueprintME:      10,
		BlueprintTE:      20,
		Runs:             10,
		Structure:        "raitaru",
		Rig:              "t1",
		Security:         "low",
		IndustrySkill:    5,
		AdvIndustrySkill: 4,
		FacilityTax:      5.0,
		SystemID:         30000142,
	}

	data := &ManufacturingData{
		Blueprint: &repositories.ManufacturingBlueprintRow{
			BlueprintTypeID: 787,
			ProductTypeID:   786,
			ProductName:     "Test Ship",
			GroupName:       "Frigate",
			ProductQuantity: 1,
			Time:            3600,
			ProductVolume:   2500,
			MaxProdLimit:    300,
		},
		Materials: []*repositories.ManufacturingMaterialRow{
			{BlueprintTypeID: 787, TypeID: 34, TypeName: "Tritanium", Quantity: 1000, Volume: 0.01},
			{BlueprintTypeID: 787, TypeID: 35, TypeName: "Pyerite", Quantity: 500, Volume: 0.01},
		},
		CostIndex: 0.01,
		AdjustedPrices: map[int64]float64{
			34: 5.0,
			35: 10.0,
		},
		JitaPrices: map[int64]*models.MarketPrice{
			34:  {TypeID: 34, SellPrice: &sellPrice34},
			35:  {TypeID: 35, SellPrice: &sellPrice35},
			786: {TypeID: 786, SellPrice: &sellProductPrice},
		},
	}

	result := CalculateManufacturingJob(params, data)

	assert.Equal(t, int64(787), result.BlueprintTypeID)
	assert.Equal(t, int64(786), result.ProductTypeID)
	assert.Equal(t, "Test Ship", result.ProductName)
	assert.Equal(t, 10, result.Runs)
	assert.Equal(t, 10, result.TotalProducts)

	// ME factor: (1 - 10/100) * (1 - 0.01) * (1 - 0.02 * 1.9) = 0.90 * 0.99 * 0.962
	expectedME := 0.90 * 0.99 * (1.0 - 0.02*1.9)
	assert.InDelta(t, expectedME, result.MEFactor, 0.001)

	// TE factor: (1 - 20/100) * (1 - 5*0.04) * (1 - 4*0.03) * (1 - 0.15) * (1 - 0.20*1.9)
	expectedTE := 0.80 * 0.80 * 0.88 * 0.85 * (1.0 - 0.20*1.9)
	assert.InDelta(t, expectedTE, result.TEFactor, 0.001)

	// Verify materials are populated
	assert.Len(t, result.Materials, 2)
	assert.Equal(t, int64(34), result.Materials[0].TypeID)
	assert.Equal(t, "Tritanium", result.Materials[0].Name)
	assert.Equal(t, 1000, result.Materials[0].BaseQty)

	// Job cost should be positive
	assert.Greater(t, result.JobCost, 0.0)
	assert.Greater(t, result.InputCost, 0.0)
	assert.Greater(t, result.TotalCost, 0.0)

	// Output value = 50000 * 10 = 500000
	assert.InDelta(t, 500000.0, result.OutputValue, 0.01)

	// Profit = output - total cost
	assert.InDelta(t, result.OutputValue-result.TotalCost, result.Profit, 0.01)

	// Duration should be computed
	assert.Greater(t, result.SecsPerRun, 0)
	assert.Equal(t, result.SecsPerRun*10, result.TotalDuration)
}

func TestCalculateManufacturingJob_NoJitaPrices(t *testing.T) {
	params := &ManufacturingParams{
		BlueprintME:      0,
		BlueprintTE:      0,
		Runs:             1,
		Structure:        "station",
		Rig:              "none",
		Security:         "high",
		IndustrySkill:    0,
		AdvIndustrySkill: 0,
		FacilityTax:      0,
		SystemID:         30000142,
	}

	data := &ManufacturingData{
		Blueprint: &repositories.ManufacturingBlueprintRow{
			BlueprintTypeID: 787,
			ProductTypeID:   786,
			ProductName:     "Test Ship",
			ProductQuantity: 1,
			Time:            3600,
		},
		Materials: []*repositories.ManufacturingMaterialRow{
			{BlueprintTypeID: 787, TypeID: 34, TypeName: "Tritanium", Quantity: 100},
		},
		CostIndex:      0.01,
		AdjustedPrices: map[int64]float64{},
		JitaPrices:     map[int64]*models.MarketPrice{},
	}

	result := CalculateManufacturingJob(params, data)

	// With no prices, costs and output should be 0
	assert.Equal(t, 0.0, result.InputCost)
	assert.Equal(t, 0.0, result.OutputValue)
	assert.Equal(t, 1, result.Runs)
	assert.Equal(t, 1, result.TotalProducts)
	// ME/TE at 0 blueprint, no rig, no structure = 1.0
	assert.InDelta(t, 1.0, result.MEFactor, 0.0001)
	assert.InDelta(t, 1.0, result.TEFactor, 0.0001)
}
