package calculator

import (
	"math"
	"sort"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
)

// ComputePlan aggregates intermediate demand across selected complex reactions,
// calculates intermediate slots/runs, generates a shopping list, and computes financials.
func ComputePlan(selections []models.PlanSelection, params *CalcParams, data *CalcData, response *models.ReactionsResponse) *models.PlanResponse {
	meFactor := ComputeMEFactor(params.Rig, params.Security)

	// Build reaction lookup by type ID
	reactionByTypeID := make(map[int64]*models.Reaction)
	for _, r := range response.Reactions {
		reactionByTypeID[r.ReactionTypeID] = r
	}

	// Build material lookup: blueprintTypeID -> []*ReactionMaterialRow
	materialsByReaction := make(map[int64][]*repositories.ReactionMaterialRow)
	for _, mat := range data.Materials {
		materialsByReaction[mat.BlueprintTypeID] = append(materialsByReaction[mat.BlueprintTypeID], mat)
	}

	// Build simple reaction lookup: productTypeID -> ReactionRow
	simpleReactions := make(map[int64]*repositories.ReactionRow)
	simpleReactionByProduct := make(map[int64]*repositories.ReactionRow)
	for _, r := range data.Reactions {
		if SimpleGroups[r.GroupName] {
			simpleReactions[r.ProductTypeID] = r
			simpleReactionByProduct[r.ProductTypeID] = r
		}
	}

	// Build reaction product set
	reactionProductIDs := make(map[int64]bool)
	for _, r := range data.Reactions {
		reactionProductIDs[r.ProductTypeID] = true
	}

	// Aggregate intermediate demand across all selected complex reactions
	intermediateDemand := make(map[int64]int64) // productTypeID -> total quantity needed

	// Track complex slots and financials
	var totalComplexSlots int
	var totalRevenue, totalComplexJobCost, totalOutputFees, totalShipping float64

	for _, sel := range selections {
		reaction, ok := reactionByTypeID[sel.ReactionTypeID]
		if !ok || sel.Instances <= 0 {
			continue
		}

		complexLines := sel.Instances * reaction.ComplexInstances
		totalComplexSlots += complexLines

		runs := float64(reaction.RunsPerCycle) * float64(complexLines)
		totalRevenue += reaction.OutputValuePerRun * runs
		totalComplexJobCost += reaction.ComplexJobCostPerRun * runs
		totalOutputFees += reaction.OutputFeesPerRun * runs
		totalShipping += (reaction.ShippingInPerRun + reaction.ShippingOutPerRun) * runs

		// Aggregate intermediate demand
		mats := materialsByReaction[reaction.ReactionTypeID]
		for _, mat := range mats {
			if reactionProductIDs[mat.TypeID] {
				batchQty := ComputeBatchQty(reaction.RunsPerCycle, mat.Quantity, meFactor)
				intermediateDemand[mat.TypeID] += batchQty * int64(complexLines)
			}
		}
	}

	// Calculate intermediate slots and runs
	teFactor := ComputeTEFactor(params.ReactionsSkill, params.Structure, params.Rig, params.Security)
	intermediates := []*models.IntermediatePlan{}
	var totalIntermediateSlots int

	// Sort intermediate type IDs for deterministic output
	intermediateTypeIDs := make([]int64, 0, len(intermediateDemand))
	for typeID := range intermediateDemand {
		intermediateTypeIDs = append(intermediateTypeIDs, typeID)
	}
	sort.Slice(intermediateTypeIDs, func(i, j int) bool {
		return intermediateTypeIDs[i] < intermediateTypeIDs[j]
	})

	// intermediateRuns tracks how many runs each intermediate needs (for shopping list)
	intermediateRunsMap := make(map[int64]int)    // productTypeID -> runs per slot
	intermediateSlotsMap := make(map[int64]int)   // productTypeID -> slots

	for _, typeID := range intermediateTypeIDs {
		demand := intermediateDemand[typeID]
		simpleReaction, ok := simpleReactionByProduct[typeID]
		if !ok {
			continue
		}

		secsPerRun := ComputeSecsPerRun(simpleReaction.Time, teFactor)
		runsPerCycle := ComputeRunsPerCycle(secsPerRun, params.CycleDays)

		supplyPerSlot := int64(simpleReaction.ProductQuantity) * int64(runsPerCycle)
		if supplyPerSlot <= 0 {
			continue
		}

		slotsNeeded := int(math.Ceil(float64(demand) / float64(supplyPerSlot)))
		if slotsNeeded < 1 {
			slotsNeeded = 1
		}

		runsNeeded := int(math.Ceil(float64(demand) / (float64(slotsNeeded) * float64(simpleReaction.ProductQuantity))))
		produced := int64(slotsNeeded) * int64(runsNeeded) * int64(simpleReaction.ProductQuantity)

		intermediateRunsMap[typeID] = runsNeeded
		intermediateSlotsMap[typeID] = slotsNeeded
		totalIntermediateSlots += slotsNeeded

		intermediates = append(intermediates, &models.IntermediatePlan{
			TypeID:   typeID,
			Name:     simpleReaction.ProductName,
			Slots:    slotsNeeded,
			Runs:     runsNeeded,
			Produced: produced,
		})
	}

	// Build shopping list
	shoppingMap := make(map[int64]*models.ShoppingItem)

	// 1. Raw materials from intermediate reactions
	for _, typeID := range intermediateTypeIDs {
		simpleReaction, ok := simpleReactionByProduct[typeID]
		if !ok {
			continue
		}
		runs := intermediateRunsMap[typeID]
		slots := intermediateSlotsMap[typeID]
		mats := materialsByReaction[simpleReaction.BlueprintTypeID]

		for _, mat := range mats {
			qtyPerSlot := ComputeBatchQty(runs, mat.Quantity, meFactor)
			totalQty := qtyPerSlot * int64(slots)

			if existing, ok := shoppingMap[mat.TypeID]; ok {
				existing.Quantity += totalQty
			} else {
				price := GetPrice(mat.TypeID, params.InputPrice, data.JitaPrices)
				shoppingMap[mat.TypeID] = &models.ShoppingItem{
					TypeID:   mat.TypeID,
					Name:     mat.TypeName,
					Quantity: totalQty,
					Price:    price,
					Volume:   mat.Volume,
				}
			}
		}
	}

	// 2. Non-intermediate materials from complex reactions (e.g., fuel blocks)
	for _, sel := range selections {
		reaction, ok := reactionByTypeID[sel.ReactionTypeID]
		if !ok || sel.Instances <= 0 {
			continue
		}

		complexLines := sel.Instances * reaction.ComplexInstances
		mats := materialsByReaction[reaction.ReactionTypeID]

		for _, mat := range mats {
			if reactionProductIDs[mat.TypeID] {
				continue // skip intermediates, already handled
			}

			qtyPerLine := ComputeBatchQty(reaction.RunsPerCycle, mat.Quantity, meFactor)
			totalQty := qtyPerLine * int64(complexLines)

			if existing, ok := shoppingMap[mat.TypeID]; ok {
				existing.Quantity += totalQty
			} else {
				price := GetPrice(mat.TypeID, params.InputPrice, data.JitaPrices)
				shoppingMap[mat.TypeID] = &models.ShoppingItem{
					TypeID:   mat.TypeID,
					Name:     mat.TypeName,
					Quantity: totalQty,
					Price:    price,
					Volume:   mat.Volume,
				}
			}
		}
	}

	// Convert shopping map to sorted list and compute costs
	shoppingList := []*models.ShoppingItem{}
	for _, item := range shoppingMap {
		item.Cost = item.Price * float64(item.Quantity)
		item.Volume = item.Volume * float64(item.Quantity)
		shoppingList = append(shoppingList, item)
	}
	sort.Slice(shoppingList, func(i, j int) bool {
		return shoppingList[i].Name < shoppingList[j].Name
	})

	// Plan summary: investment = shopping list (raw mats) + all job costs + output fees + shipping
	var shoppingCost float64
	for _, item := range shoppingList {
		shoppingCost += item.Cost
	}

	// Compute intermediate job costs
	var totalIntermediateJobCost float64
	for _, typeID := range intermediateTypeIDs {
		simpleReaction, ok := simpleReactionByProduct[typeID]
		if !ok {
			continue
		}
		runs := intermediateRunsMap[typeID]
		slots := intermediateSlotsMap[typeID]
		mats := materialsByReaction[simpleReaction.BlueprintTypeID]
		jobCostPerRun := ComputeReactionJobCost(mats, data.AdjustedPrices, data.CostIndex, params.FacilityTax)
		totalIntermediateJobCost += jobCostPerRun * float64(runs) * float64(slots)
	}

	totalInvestment := shoppingCost + totalComplexJobCost + totalIntermediateJobCost + totalOutputFees + totalShipping
	totalSlots := totalIntermediateSlots + totalComplexSlots
	profit := totalRevenue - totalInvestment
	var margin float64
	if totalRevenue > 0 {
		margin = (profit / totalRevenue) * 100.0
	}

	summary := &models.PlanSummary{
		TotalSlots:        totalSlots,
		IntermediateSlots: totalIntermediateSlots,
		ComplexSlots:      totalComplexSlots,
		Investment:        math.Round(totalInvestment*100) / 100,
		Revenue:           math.Round(totalRevenue*100) / 100,
		Profit:            math.Round(profit*100) / 100,
		Margin:            math.Round(margin*100) / 100,
	}

	return &models.PlanResponse{
		Intermediates: intermediates,
		ShoppingList:  shoppingList,
		Summary:       summary,
	}
}
