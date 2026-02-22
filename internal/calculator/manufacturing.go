package calculator

import (
	"math"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
)

// ManufacturingParams holds user-configurable settings for a manufacturing job calculation
type ManufacturingParams struct {
	BlueprintME    int     // 0-10
	BlueprintTE    int     // 0-20
	Runs           int
	Structure      string  // "raitaru", "azbel", "sotiyo", "station"
	Rig            string  // "none", "t1", "t2"
	Security       string  // "null", "low", "high"
	IndustrySkill  int     // 0-5
	AdvIndustrySkill int   // 0-5
	FacilityTax    float64 // percentage
	SystemID       int64
}

// ManufacturingData holds data fetched from the database for calculations
type ManufacturingData struct {
	Blueprint      *repositories.ManufacturingBlueprintRow
	Materials      []*repositories.ManufacturingMaterialRow
	CostIndex      float64
	AdjustedPrices map[int64]float64
	JitaPrices     map[int64]*models.MarketPrice
}

// EngineeringSecurityMultiplier returns the rig bonus multiplier for engineering complexes.
// Engineering complexes use different security multipliers than refineries:
// Null/WH: 2.1x, Low: 1.9x, High: 1.0x
func EngineeringSecurityMultiplier(security string) float64 {
	switch security {
	case "null":
		return 2.1
	case "low":
		return 1.9
	default: // "high"
		return 1.0
	}
}

// ComputeManufacturingME calculates the combined ME factor for manufacturing.
// Manufacturing ME combines blueprint ME (0-10) with structure ME bonus and rig ME.
// Engineering complexes (Raitaru, Azbel, Sotiyo) provide a 1% material reduction role bonus.
// combined_me = (1 - blueprint_me/100) * (1 - structure_me) * (1 - rig_me * security_mult)
func ComputeManufacturingME(blueprintME int, structure, rig, security string) float64 {
	structME := ManufacturingStructureMEValue(structure)
	rigME := RigMEValue(rig)
	secMult := EngineeringSecurityMultiplier(security)
	return (1.0 - float64(blueprintME)/100.0) * (1.0 - structME) * (1.0 - rigME*secMult)
}

// ManufacturingStructureMEValue returns the material efficiency bonus for manufacturing structures.
// Engineering complexes (Raitaru, Azbel, Sotiyo) provide a 1% material reduction role bonus.
// NPC stations provide no bonus.
func ManufacturingStructureMEValue(structure string) float64 {
	switch structure {
	case "raitaru", "azbel", "sotiyo":
		return 0.01
	default: // "station" or unknown
		return 0
	}
}

// ComputeManufacturingTE calculates the combined TE factor for manufacturing.
// Manufacturing TE combines blueprint TE (0-20) with skills, structure, and rig bonuses.
// combined_te = (1 - blueprint_te/100) * (1 - industry*0.04) * (1 - adv_industry*0.03) * (1 - structure_te) * (1 - rig_te * sec_mult)
func ComputeManufacturingTE(blueprintTE, industrySkill, advIndustrySkill int, structure, rig, security string) float64 {
	rigTE := RigTEValue(rig)
	structTE := ManufacturingStructureTEValue(structure)
	secMult := EngineeringSecurityMultiplier(security)
	return (1.0 - float64(blueprintTE)/100.0) *
		(1.0 - float64(industrySkill)*0.04) *
		(1.0 - float64(advIndustrySkill)*0.03) *
		(1.0 - structTE) *
		(1.0 - rigTE*secMult)
}

// ManufacturingStructureTEValue returns the time efficiency bonus for manufacturing structures.
// Sotiyo: 30%, Azbel: 20%, Raitaru: 15%, Station: 0%
func ManufacturingStructureTEValue(structure string) float64 {
	switch structure {
	case "sotiyo":
		return 0.30
	case "azbel":
		return 0.20
	case "raitaru":
		return 0.15
	default: // "station" or unknown
		return 0
	}
}

// ManufacturingStructureCostBonus returns the job cost reduction for engineering complexes.
// This bonus reduces only the system cost index portion of the job cost.
// Sotiyo: 5%, Azbel: 3%, Raitaru: 1%, Station: 0%
func ManufacturingStructureCostBonus(structure string) float64 {
	switch structure {
	case "sotiyo":
		return 0.05
	case "azbel":
		return 0.03
	case "raitaru":
		return 0.01
	default: // "station" or unknown
		return 0
	}
}

// ComputeManufacturingJobCost calculates the industry job cost for one run of a manufacturing job.
// Uses base quantities (ME 0) for EIV calculation, same as reactions.
// The structure cost bonus only reduces the system cost index portion:
// job_cost = EIV × cost_index × (1 - structure_bonus) + EIV × scc_surcharge + EIV × facility_tax
func ComputeManufacturingJobCost(materials []*repositories.ManufacturingMaterialRow, adjustedPrices map[int64]float64, costIndex, facilityTax float64, structure string) float64 {
	var eiv float64
	for _, mat := range materials {
		adjPrice, ok := adjustedPrices[mat.TypeID]
		if !ok {
			continue
		}
		eiv += float64(mat.Quantity) * adjPrice
	}
	structBonus := ManufacturingStructureCostBonus(structure)
	return eiv*costIndex*(1.0-structBonus) + eiv*SccSurchargeRate + eiv*(facilityTax/100.0)
}

// CalculateManufacturingJob calculates the full cost breakdown for a manufacturing job.
func CalculateManufacturingJob(params *ManufacturingParams, data *ManufacturingData) *models.ManufacturingCalcResult {
	meFactor := ComputeManufacturingME(params.BlueprintME, params.Structure, params.Rig, params.Security)
	teFactor := ComputeManufacturingTE(params.BlueprintTE, params.IndustrySkill, params.AdvIndustrySkill, params.Structure, params.Rig, params.Security)

	// Calculate time per run
	secsPerRun := ComputeSecsPerRun(data.Blueprint.Time, teFactor)
	totalDuration := secsPerRun * params.Runs

	// Calculate materials
	materials := []*models.ManufacturingMaterial{}
	var totalInputCost float64

	for _, mat := range data.Materials {
		batchQty := ComputeBatchQty(params.Runs, mat.Quantity, meFactor)
		price := GetPrice(mat.TypeID, "sell", data.JitaPrices)
		cost := price * float64(batchQty)

		materials = append(materials, &models.ManufacturingMaterial{
			TypeID:   mat.TypeID,
			Name:     mat.TypeName,
			BaseQty:  mat.Quantity,
			BatchQty: batchQty,
			Price:    price,
			Cost:     math.Round(cost*100) / 100,
		})

		totalInputCost += cost
	}

	// Job installation cost
	jobCostPerRun := ComputeManufacturingJobCost(data.Materials, data.AdjustedPrices, data.CostIndex, params.FacilityTax, params.Structure)
	totalJobCost := jobCostPerRun * float64(params.Runs)

	// Total output
	totalProducts := data.Blueprint.ProductQuantity * params.Runs

	// Output value
	outputPrice := GetPrice(data.Blueprint.ProductTypeID, "sell", data.JitaPrices)
	totalOutputValue := outputPrice * float64(totalProducts)

	// Total cost and profit
	totalCost := totalInputCost + totalJobCost
	profit := totalOutputValue - totalCost

	var margin float64
	if totalOutputValue > 0 {
		margin = (profit / totalOutputValue) * 100.0
	}

	return &models.ManufacturingCalcResult{
		BlueprintTypeID: data.Blueprint.BlueprintTypeID,
		ProductTypeID:   data.Blueprint.ProductTypeID,
		ProductName:     data.Blueprint.ProductName,
		Runs:            params.Runs,
		MEFactor:        math.Round(meFactor*10000) / 10000,
		TEFactor:        math.Round(teFactor*10000) / 10000,
		SecsPerRun:      secsPerRun,
		TotalDuration:   totalDuration,
		TotalProducts:   totalProducts,
		InputCost:       math.Round(totalInputCost*100) / 100,
		JobCost:         math.Round(totalJobCost*100) / 100,
		TotalCost:       math.Round(totalCost*100) / 100,
		OutputValue:     math.Round(totalOutputValue*100) / 100,
		Profit:          math.Round(profit*100) / 100,
		Margin:          math.Round(margin*100) / 100,
		Materials:       materials,
	}
}
