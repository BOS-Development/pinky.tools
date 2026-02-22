package calculator

import (
	"math"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
)

// CalcParams holds all user-configurable settings for the reactions calculator
type CalcParams struct {
	SystemID           int64
	Structure          string  // "tatara", "athanor"
	Rig                string  // "none", "t1", "t2"
	Security           string  // "null", "low", "high"
	ReactionsSkill     int     // 0-5
	FacilityTax        float64 // percentage
	CycleDays          int     // 1-30
	BrokerFee          float64 // percentage
	SalesTax           float64 // percentage
	ShippingM3         float64
	ShippingCollateral float64
	InputPrice         string // "sell", "buy", "split"
	OutputPrice        string // "sell", "buy"
	ShipInputs         bool
	ShipOutputs        bool
}

// CalcData holds all data fetched from the database needed for calculations
type CalcData struct {
	Reactions      []*repositories.ReactionRow
	Materials      []*repositories.ReactionMaterialRow
	CostIndex      float64
	JitaPrices     map[int64]*models.MarketPrice
	AdjustedPrices map[int64]float64
}

// Simple reaction group names — these are intermediate reactions, not selectable
var SimpleGroups = map[string]bool{
	"Intermediate Materials": true,
	"Unrefined Mineral":     true,
}

// ComputeMEFactor calculates the material efficiency factor
func ComputeMEFactor(rig, security string) float64 {
	rigME := RigMEValue(rig)
	secMult := SecurityMultiplier(security)
	return 1.0 - rigME*secMult
}

// ComputeTEFactor calculates the time efficiency factor
func ComputeTEFactor(skill int, structure, rig, security string) float64 {
	rigTE := RigTEValue(rig)
	structTE := StructureTEValue(structure)
	secMult := SecurityMultiplier(security)
	return (1.0 - float64(skill)*0.04) * (1.0 - structTE) * (1.0 - rigTE*secMult)
}

// ComputeSecsPerRun calculates seconds per run after TE
func ComputeSecsPerRun(baseTime int, teFactor float64) int {
	return int(math.Floor(float64(baseTime) * teFactor))
}

// ComputeRunsPerCycle calculates how many runs fit in a cycle
func ComputeRunsPerCycle(secsPerRun, cycleDays int) int {
	cycleSeconds := cycleDays * 86400
	if secsPerRun <= 0 {
		return 0
	}
	return int(math.Floor(float64(cycleSeconds) / float64(secsPerRun)))
}

// ComputeAdjQty calculates per-run adjusted quantity (for display)
func ComputeAdjQty(baseQty int, meFactor float64) int {
	return int(math.Ceil(float64(baseQty) * meFactor))
}

// ComputeBatchQty calculates actual batch quantity consumed across all runs (EVE batch ME)
func ComputeBatchQty(runs, baseQty int, meFactor float64) int64 {
	raw := math.Ceil(float64(runs) * float64(baseQty) * meFactor)
	result := int64(raw)
	if int64(runs) > result {
		return int64(runs)
	}
	return result
}

// Calculate processes all reaction data and returns the full response
func Calculate(params *CalcParams, data *CalcData) *models.ReactionsResponse {
	meFactor := ComputeMEFactor(params.Rig, params.Security)
	teFactor := ComputeTEFactor(params.ReactionsSkill, params.Structure, params.Rig, params.Security)

	// Build material lookup: blueprintTypeID -> []*ReactionMaterialRow
	materialsByReaction := make(map[int64][]*repositories.ReactionMaterialRow)
	for _, mat := range data.Materials {
		materialsByReaction[mat.BlueprintTypeID] = append(materialsByReaction[mat.BlueprintTypeID], mat)
	}

	// Build reaction product set (to identify intermediates)
	reactionProductIDs := make(map[int64]bool)
	for _, r := range data.Reactions {
		reactionProductIDs[r.ProductTypeID] = true
	}

	// Build simple reaction lookup: productTypeID -> ReactionRow
	simpleReactions := make(map[int64]*repositories.ReactionRow)
	for _, r := range data.Reactions {
		if SimpleGroups[r.GroupName] {
			simpleReactions[r.ProductTypeID] = r
		}
	}

	// Compute per-unit costs for each intermediate reaction using batch ME.
	// This traces intermediates back to their raw moon goo inputs so that
	// complex reaction margins reflect actual production cost, not market price.
	// We use the same runs-per-cycle as the complex reactions to get accurate batch ME.
	intermediateRawCostPerUnit := make(map[int64]float64)
	intermediateJobCostPerUnit := make(map[int64]float64)
	for productTypeID, simpleReaction := range simpleReactions {
		simpleMats := materialsByReaction[simpleReaction.BlueprintTypeID]
		secsPerRun := ComputeSecsPerRun(simpleReaction.Time, teFactor)
		runsPerCycle := ComputeRunsPerCycle(secsPerRun, params.CycleDays)
		if runsPerCycle <= 0 {
			continue
		}
		var batchCost float64
		for _, mat := range simpleMats {
			batchQty := ComputeBatchQty(runsPerCycle, mat.Quantity, meFactor)
			price := GetPrice(mat.TypeID, params.InputPrice, data.JitaPrices)
			batchCost += price * float64(batchQty)
		}
		jobCostPerRun := ComputeReactionJobCost(simpleMats, data.AdjustedPrices, data.CostIndex, params.FacilityTax)
		totalProduced := float64(simpleReaction.ProductQuantity) * float64(runsPerCycle)
		if totalProduced > 0 {
			intermediateRawCostPerUnit[productTypeID] = batchCost / totalProduced
			intermediateJobCostPerUnit[productTypeID] = (jobCostPerRun * float64(runsPerCycle)) / totalProduced
		}
	}

	reactions := []*models.Reaction{}

	for _, r := range data.Reactions {
		secsPerRun := ComputeSecsPerRun(r.Time, teFactor)
		runsPerCycle := ComputeRunsPerCycle(secsPerRun, params.CycleDays)

		mats := materialsByReaction[r.BlueprintTypeID]
		isComplex := !SimpleGroups[r.GroupName]

		// Calculate materials using batch ME for accurate cost/profit
		reactionMaterials := []*models.ReactionMaterial{}
		var inputCostPerRun float64
		var inputVolumePerRun float64
		numIntermediates := 0

		for _, mat := range mats {
			adjQty := ComputeAdjQty(mat.Quantity, meFactor)
			isIntermediate := reactionProductIDs[mat.TypeID]
			if isIntermediate {
				numIntermediates++
			}

			// For intermediates in complex reactions, use raw material production cost
			// instead of market price so margins reflect actual production cost
			var price float64
			if isIntermediate && isComplex {
				price = intermediateRawCostPerUnit[mat.TypeID]
			} else {
				price = GetPrice(mat.TypeID, params.InputPrice, data.JitaPrices)
			}

			// Display values use per-run adjQty
			cost := price * float64(adjQty)
			volume := mat.Volume * float64(adjQty)

			reactionMaterials = append(reactionMaterials, &models.ReactionMaterial{
				TypeID:         mat.TypeID,
				Name:           mat.TypeName,
				BaseQty:        mat.Quantity,
				AdjQty:         adjQty,
				Price:          price,
				Cost:           cost,
				Volume:         volume,
				IsIntermediate: isIntermediate,
			})

			// Aggregate costs use batch ME (actual consumption across all runs)
			if runsPerCycle > 0 {
				batchQty := ComputeBatchQty(runsPerCycle, mat.Quantity, meFactor)
				inputCostPerRun += price * float64(batchQty) / float64(runsPerCycle)
				inputVolumePerRun += mat.Volume * float64(batchQty) / float64(runsPerCycle)
			}
		}

		// Job cost: complexJobCostPerRun is this reaction's own job cost (used by plan),
		// jobCostPerRun includes intermediate production job costs (used for per-row margin)
		complexJobCostPerRun := ComputeReactionJobCost(mats, data.AdjustedPrices, data.CostIndex, params.FacilityTax)
		jobCostPerRun := complexJobCostPerRun
		if isComplex && runsPerCycle > 0 {
			for _, mat := range mats {
				if reactionProductIDs[mat.TypeID] {
					batchQty := ComputeBatchQty(runsPerCycle, mat.Quantity, meFactor)
					jobCostPerRun += intermediateJobCostPerUnit[mat.TypeID] * float64(batchQty) / float64(runsPerCycle)
				}
			}
		}

		// Output value
		outputPrice := GetPrice(r.ProductTypeID, params.OutputPrice, data.JitaPrices)
		outputValuePerRun := outputPrice * float64(r.ProductQuantity)
		outputFeesPerRun := outputValuePerRun * (params.BrokerFee + params.SalesTax) / 100.0
		outputVolumePerRun := r.ProductVolume * float64(r.ProductQuantity)

		// Shipping
		var shippingInPerRun, shippingOutPerRun float64
		if params.ShipInputs {
			shippingInPerRun = inputVolumePerRun*params.ShippingM3 + inputCostPerRun*params.ShippingCollateral
		}
		if params.ShipOutputs {
			shippingOutPerRun = outputVolumePerRun*params.ShippingM3 + outputValuePerRun*params.ShippingCollateral
		}

		// Profit
		totalCostPerRun := inputCostPerRun + jobCostPerRun + shippingInPerRun + shippingOutPerRun
		profitPerRun := outputValuePerRun - outputFeesPerRun - totalCostPerRun
		profitPerCycle := profitPerRun * float64(runsPerCycle)

		var margin float64
		if outputValuePerRun > 0 {
			margin = (profitPerRun / outputValuePerRun) * 100.0
		}

		// Complex instances
		complexInstances := 0
		if isComplex {
			complexInstances = computeComplexInstances(mats, simpleReactions, meFactor)
		}

		reactions = append(reactions, &models.Reaction{
			ReactionTypeID:    r.BlueprintTypeID,
			ProductTypeID:     r.ProductTypeID,
			ProductName:       r.ProductName,
			GroupName:         r.GroupName,
			ProductQtyPerRun:  r.ProductQuantity,
			RunsPerCycle:      runsPerCycle,
			SecsPerRun:        secsPerRun,
			ComplexInstances:  complexInstances,
			NumIntermediates:  numIntermediates,
			InputCostPerRun:      math.Round(inputCostPerRun*100) / 100,
			JobCostPerRun:        math.Round(jobCostPerRun*100) / 100,
			ComplexJobCostPerRun: math.Round(complexJobCostPerRun*100) / 100,
			OutputValuePerRun: math.Round(outputValuePerRun*100) / 100,
			OutputFeesPerRun:  math.Round(outputFeesPerRun*100) / 100,
			ShippingInPerRun:  math.Round(shippingInPerRun*100) / 100,
			ShippingOutPerRun: math.Round(shippingOutPerRun*100) / 100,
			ProfitPerRun:      math.Round(profitPerRun*100) / 100,
			ProfitPerCycle:    math.Round(profitPerCycle*100) / 100,
			Margin:            math.Round(margin*100) / 100,
			Materials:         reactionMaterials,
		})
	}

	// Compute a base runs_per_cycle using the standard reaction time (10800s)
	baseSecsPerRun := ComputeSecsPerRun(10800, teFactor)
	baseRunsPerCycle := ComputeRunsPerCycle(baseSecsPerRun, params.CycleDays)

	return &models.ReactionsResponse{
		Reactions:    reactions,
		Count:        len(reactions),
		CostIndex:    data.CostIndex,
		MEFactor:     math.Round(meFactor*10000) / 10000,
		TEFactor:     math.Round(teFactor*10000) / 10000,
		RunsPerCycle: baseRunsPerCycle,
	}
}

// computeComplexInstances calculates how many parallel complex lines one intermediate cycle feeds
func computeComplexInstances(mats []*repositories.ReactionMaterialRow, simpleReactions map[int64]*repositories.ReactionRow, meFactor float64) int {
	minRatio := math.MaxFloat64
	hasIntermediate := false

	for _, mat := range mats {
		simpleReaction, ok := simpleReactions[mat.TypeID]
		if !ok {
			continue
		}
		hasIntermediate = true

		adjQty := ComputeAdjQty(mat.Quantity, meFactor)
		if adjQty <= 0 {
			continue
		}

		ratio := float64(simpleReaction.ProductQuantity) / float64(adjQty)
		if ratio < minRatio {
			minRatio = ratio
		}
	}

	if !hasIntermediate {
		return 1
	}

	result := int(math.Floor(minRatio))
	if result < 1 {
		return 1
	}
	return result
}

// computeJobCost calculates the industry job cost for one run of a reaction.
// EVE uses base quantities (ME 0), not ME-adjusted quantities, for EIV calculation.
// The total cost is: EIV × system_cost_index + EIV × scc_surcharge + EIV × facility_tax
// All three components are additive and applied independently to the EIV.
// SCC surcharge for reactions is 4% (increased from 1.5% in the Viridian expansion).
const SccSurchargeRate = 0.04

func ComputeReactionJobCost(mats []*repositories.ReactionMaterialRow, adjustedPrices map[int64]float64, costIndex, facilityTax float64) float64 {
	var eiv float64
	for _, mat := range mats {
		adjPrice, ok := adjustedPrices[mat.TypeID]
		if !ok {
			continue
		}
		eiv += float64(mat.Quantity) * adjPrice
	}
	return eiv * (costIndex + SccSurchargeRate + facilityTax/100.0)
}

// getPrice resolves price for a type using the specified method
func GetPrice(typeID int64, method string, jitaPrices map[int64]*models.MarketPrice) float64 {
	mp, ok := jitaPrices[typeID]
	if !ok {
		return 0
	}

	var sell, buy float64
	if mp.SellPrice != nil {
		sell = *mp.SellPrice
	}
	if mp.BuyPrice != nil {
		buy = *mp.BuyPrice
	}

	switch method {
	case "buy":
		return buy
	case "split":
		return (sell + buy) / 2.0
	default: // "sell"
		return sell
	}
}

func RigMEValue(rig string) float64 {
	switch rig {
	case "t1":
		return 0.02
	case "t2":
		return 0.024
	default:
		return 0
	}
}

func RigTEValue(rig string) float64 {
	switch rig {
	case "t1":
		return 0.20
	case "t2":
		return 0.24
	default:
		return 0
	}
}

func StructureTEValue(structure string) float64 {
	switch structure {
	case "tatara":
		return 0.25
	default: // "athanor"
		return 0
	}
}

func SecurityMultiplier(security string) float64 {
	switch security {
	case "null":
		return 1.1
	default: // "low", "high"
		return 1.0
	}
}
