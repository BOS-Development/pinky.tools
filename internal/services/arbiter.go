package services

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/annymsMthd/industry-tool/internal/calculator"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

// ArbiterScanRepository groups all the repository interfaces needed by ScanOpportunities.
type ArbiterScanRepository interface {
	GetT2BlueprintsForScan(ctx context.Context) ([]*models.T2BlueprintScanItem, error)
	GetDecryptors(ctx context.Context) ([]*models.Decryptor, error)
	GetMarketPricesForTypes(ctx context.Context, typeIDs []int64) (map[int64]*models.MarketPrice, error)
	GetBlueprintMaterialsForActivity(ctx context.Context, blueprintTypeID int64, activity string) ([]*models.BlueprintMaterial, error)
	GetBlueprintProductForActivity(ctx context.Context, blueprintTypeID int64, activity string) (*models.BlueprintProduct, error)
	GetBlueprintActivityTime(ctx context.Context, blueprintTypeID int64, activity string) (int64, error)
	GetBestInventionCharacter(ctx context.Context, userID int64, blueprintTypeID int64) (*models.InventionCharacter, error)
	GetCostIndexForSystem(ctx context.Context, systemID int64, activity string) (float64, error)
	GetAdjustedPricesForTypes(ctx context.Context, typeIDs []int64) (map[int64]float64, error)
}

// bomResult holds the accumulated material and job cost for a BOM tree calculation.
type bomResult struct {
	MaterialCost float64
	JobCost      float64
	BuildTimeSec int64
}

// arbiterContext holds all pre-loaded (and lazily cached) data for a single scan run.
type arbiterContext struct {
	settings       *models.ArbiterSettings
	prices         map[int64]*models.MarketPrice
	adjustedPrices map[int64]float64
	repo           ArbiterScanRepository
	ctx            context.Context
}

// getPrice returns the Jita sell price for a type, fetching lazily if needed.
func (ac *arbiterContext) getPrice(typeID int64) float64 {
	if mp, ok := ac.prices[typeID]; ok {
		if mp.SellPrice != nil {
			return *mp.SellPrice
		}
		return 0
	}
	// Lazily fetch
	newPrices, err := ac.repo.GetMarketPricesForTypes(ac.ctx, []int64{typeID})
	if err == nil {
		for k, v := range newPrices {
			ac.prices[k] = v
		}
		if mp, ok := ac.prices[typeID]; ok && mp.SellPrice != nil {
			return *mp.SellPrice
		}
	}
	return 0
}

// getAdjustedPrice returns the adjusted price for a type.
func (ac *arbiterContext) getAdjustedPrice(typeID int64) float64 {
	if p, ok := ac.adjustedPrices[typeID]; ok {
		return p
	}
	newAdj, err := ac.repo.GetAdjustedPricesForTypes(ac.ctx, []int64{typeID})
	if err == nil {
		for k, v := range newAdj {
			ac.adjustedPrices[k] = v
		}
		return ac.adjustedPrices[typeID]
	}
	return 0
}

// ScanOpportunities runs the full Arbiter scan and returns ranked T2 opportunities.
func ScanOpportunities(ctx context.Context, userID int64, settings *models.ArbiterSettings, repo ArbiterScanRepository) (*models.ArbiterScanResult, error) {
	items, err := repo.GetT2BlueprintsForScan(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get T2 blueprints")
	}

	if len(items) == 0 {
		return &models.ArbiterScanResult{
			Opportunities: []*models.ArbiterOpportunity{},
			GeneratedAt:   time.Now(),
			TotalScanned:  0,
		}, nil
	}

	decryptors, err := repo.GetDecryptors(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get decryptors")
	}

	// Bulk-load Jita prices for all product type IDs up front
	typeIDs := make([]int64, 0, len(items))
	for _, item := range items {
		typeIDs = append(typeIDs, item.ProductTypeID)
	}
	prices, err := repo.GetMarketPricesForTypes(ctx, typeIDs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load market prices")
	}

	adjustedPrices, err := repo.GetAdjustedPricesForTypes(ctx, typeIDs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load adjusted prices")
	}

	ac := &arbiterContext{
		settings:       settings,
		prices:         prices,
		adjustedPrices: adjustedPrices,
		repo:           repo,
		ctx:            ctx,
	}

	// Find best invention character using the first T1 blueprint as representative
	var bestChar *models.InventionCharacter
	if len(items) > 0 {
		bestChar, err = repo.GetBestInventionCharacter(ctx, userID, items[0].T1BlueprintTypeID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get best invention character")
		}
	}

	opportunities := []*models.ArbiterOpportunity{}

	for _, item := range items {
		opp, err := calculateOpportunity(ac, item, decryptors, bestChar)
		if err != nil || opp == nil {
			continue
		}
		opportunities = append(opportunities, opp)
	}

	sort.Slice(opportunities, func(i, j int) bool {
		pi, pj := float64(0), float64(0)
		if opportunities[i].BestDecryptor != nil {
			pi = opportunities[i].BestDecryptor.Profit
		}
		if opportunities[j].BestDecryptor != nil {
			pj = opportunities[j].BestDecryptor.Profit
		}
		return pi > pj
	})

	result := &models.ArbiterScanResult{
		Opportunities: opportunities,
		GeneratedAt:   time.Now(),
		TotalScanned:  len(items),
	}
	if bestChar != nil {
		result.BestCharacterID = bestChar.CharacterID
		result.BestCharacterName = bestChar.Name
	}
	return result, nil
}

func calculateOpportunity(ac *arbiterContext, item *models.T2BlueprintScanItem, decryptors []*models.Decryptor, bestChar *models.InventionCharacter) (*models.ArbiterOpportunity, error) {
	mp := ac.prices[item.ProductTypeID]
	var jitaSell, jitaBuy float64
	if mp != nil {
		if mp.SellPrice != nil {
			jitaSell = *mp.SellPrice
		}
		if mp.BuyPrice != nil {
			jitaBuy = *mp.BuyPrice
		}
	}

	var encLevel, sci1, sci2 int
	if bestChar != nil {
		encLevel = bestChar.EncryptionSkillLevel
		sci1 = bestChar.Science1SkillLevel
		sci2 = bestChar.Science2SkillLevel
	}

	// base_chance * (1 + enc*0.01 + (sci1+sci2)*0.005)
	skillBonus := 1.0 + float64(encLevel)*0.01 + float64(sci1+sci2)*0.005
	effectiveChance := item.BaseInventionChance * skillBonus

	// Copy + datacore cost for one invention attempt
	copyAndDatacoreCost, err := calculateInventionBaseCost(ac, item)
	if err != nil {
		return nil, err
	}

	// Manufacturing build time for the T2 product
	buildTime, err := ac.repo.GetBlueprintActivityTime(ac.ctx, item.BlueprintTypeID, "manufacturing")
	if err != nil {
		buildTime = 0
	}

	// Build all decryptor options: no-decryptor + each decryptor
	allOptions := make([]*models.DecryptorOption, 0, len(decryptors)+1)

	noDecOpt, err := calculateDecryptorOption(ac, item, &models.Decryptor{
		ProbabilityMultiplier: 1.0,
	}, nil, effectiveChance, copyAndDatacoreCost, jitaSell, buildTime)
	if err == nil && noDecOpt != nil {
		allOptions = append(allOptions, noDecOpt)
	}

	for _, dec := range decryptors {
		d := dec
		opt, err := calculateDecryptorOption(ac, item, d, &d.TypeID, effectiveChance, copyAndDatacoreCost, jitaSell, buildTime)
		if err == nil && opt != nil {
			allOptions = append(allOptions, opt)
		}
	}

	if len(allOptions) == 0 {
		return nil, nil
	}

	var best *models.DecryptorOption
	for _, opt := range allOptions {
		if best == nil || opt.Profit > best.Profit {
			best = opt
		}
	}

	return &models.ArbiterOpportunity{
		ProductTypeID: item.ProductTypeID,
		ProductName:   item.ProductName,
		Category:      item.Category,
		JitaSellPrice: jitaSell,
		JitaBuyPrice:  jitaBuy,
		BestDecryptor: best,
		AllDecryptors: allOptions,
	}, nil
}

func calculateDecryptorOption(
	ac *arbiterContext,
	item *models.T2BlueprintScanItem,
	decryptor *models.Decryptor,
	decryptorTypeID *int64,
	effectiveChance float64,
	copyAndDatacoreCost float64,
	jitaSellPrice float64,
	buildTime int64,
) (*models.DecryptorOption, error) {
	chanceMod := effectiveChance * decryptor.ProbabilityMultiplier

	resultME := item.BaseResultME + decryptor.MEModifier
	if resultME < 0 {
		resultME = 0
	}
	if resultME > 10 {
		resultME = 10
	}

	resultRuns := item.BaseResultRuns + decryptor.RunModifier
	if resultRuns < 1 {
		resultRuns = 1
	}

	var decryptorCost float64
	if decryptorTypeID != nil {
		decryptorCost = ac.getPrice(*decryptorTypeID)
	}

	var inventionCost float64
	if chanceMod > 0 {
		inventionCost = (copyAndDatacoreCost + decryptorCost) / chanceMod
	}

	// BOM for the final T2 product using the T2 blueprint and result ME
	bom, err := calculateFinalBOM(ac, item, resultME, resultRuns)
	if err != nil {
		return nil, err
	}

	totalCost := bom.MaterialCost + bom.JobCost + inventionCost
	profit := jitaSellPrice - totalCost
	var roi float64
	if totalCost > 0 {
		roi = profit / totalCost * 100.0
	}

	buildTimeSec := buildTime
	if buildTimeSec <= 0 {
		buildTimeSec = bom.BuildTimeSec
	}

	var iskPerDay float64
	if buildTimeSec > 0 {
		iskPerDay = profit / (float64(buildTimeSec) / 86400.0)
	}

	name := "No Decryptor"
	if decryptor.Name != "" {
		name = decryptor.Name
	}

	return &models.DecryptorOption{
		TypeID:                decryptorTypeID,
		Name:                  name,
		ProbabilityMultiplier: decryptor.ProbabilityMultiplier,
		MEModifier:            decryptor.MEModifier,
		TEModifier:            decryptor.TEModifier,
		RunModifier:           decryptor.RunModifier,
		ResultingME:           resultME,
		ResultingRuns:         resultRuns,
		InventionCost:         math.Round(inventionCost*100) / 100,
		MaterialCost:          math.Round(bom.MaterialCost*100) / 100,
		JobCost:               math.Round(bom.JobCost*100) / 100,
		TotalCost:             math.Round(totalCost*100) / 100,
		Profit:                math.Round(profit*100) / 100,
		ROI:                   math.Round(roi*100) / 100,
		ISKPerDay:             math.Round(iskPerDay*100) / 100,
		BuildTimeSec:          buildTimeSec,
	}, nil
}

// calculateInventionBaseCost returns copy cost + datacore cost for one invention attempt.
func calculateInventionBaseCost(ac *arbiterContext, item *models.T2BlueprintScanItem) (float64, error) {
	datecoreMats, err := ac.repo.GetBlueprintMaterialsForActivity(ac.ctx, item.T1BlueprintTypeID, "invention")
	if err != nil {
		return 0, errors.Wrap(err, "failed to get datacore materials")
	}

	// Ensure datacore prices are loaded
	datacoreTypeIDs := make([]int64, 0, len(datecoreMats))
	for _, m := range datecoreMats {
		datacoreTypeIDs = append(datacoreTypeIDs, m.TypeID)
	}
	if len(datacoreTypeIDs) > 0 {
		newPrices, err := ac.repo.GetMarketPricesForTypes(ac.ctx, datacoreTypeIDs)
		if err == nil {
			for k, v := range newPrices {
				ac.prices[k] = v
			}
		}
	}

	var dataCoreCost float64
	for _, m := range datecoreMats {
		dataCoreCost += ac.getPrice(m.TypeID) * float64(m.Quantity)
	}

	// Copy cost: approximate using invention system cost index if configured
	var copyCost float64
	if ac.settings.InventionSystemID != nil {
		copyIdx, err := ac.repo.GetCostIndexForSystem(ac.ctx, *ac.settings.InventionSystemID, "copying")
		if err == nil && copyIdx > 0 {
			t1Product, err := ac.repo.GetBlueprintProductForActivity(ac.ctx, item.T1BlueprintTypeID, "manufacturing")
			if err == nil && t1Product != nil {
				adjPrice := ac.getAdjustedPrice(t1Product.TypeID)
				copyCost = adjPrice * copyIdx
			}
		}
	}

	return dataCoreCost + copyCost, nil
}

// calculateFinalBOM computes the material + job cost for building the T2 product.
// Uses only the T2 blueprint's direct materials (one level deep), buying sub-materials at market.
func calculateFinalBOM(ac *arbiterContext, item *models.T2BlueprintScanItem, me int, runs int) (*bomResult, error) {
	structure := ac.settings.FinalStructure
	rig := ac.settings.FinalRig
	security := ac.settings.FinalSecurity
	systemID := ac.settings.FinalSystemID

	meFactor := calculator.ComputeManufacturingME(me, structure, rig, security)

	mats, err := ac.repo.GetBlueprintMaterialsForActivity(ac.ctx, item.BlueprintTypeID, "manufacturing")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get T2 manufacturing materials")
	}

	// Ensure material prices are loaded
	matTypeIDs := make([]int64, 0, len(mats))
	for _, m := range mats {
		matTypeIDs = append(matTypeIDs, m.TypeID)
	}
	if len(matTypeIDs) > 0 {
		newPrices, err := ac.repo.GetMarketPricesForTypes(ac.ctx, matTypeIDs)
		if err == nil {
			for k, v := range newPrices {
				ac.prices[k] = v
			}
		}
		// Also ensure adjusted prices
		newAdj, err := ac.repo.GetAdjustedPricesForTypes(ac.ctx, matTypeIDs)
		if err == nil {
			for k, v := range newAdj {
				ac.adjustedPrices[k] = v
			}
		}
	}

	var totalMatCost float64
	var eiv float64

	for _, m := range mats {
		batchQty := calculator.ComputeBatchQty(runs, m.Quantity, meFactor)
		price := ac.getPrice(m.TypeID)
		totalMatCost += price * float64(batchQty)

		adjPrice := ac.getAdjustedPrice(m.TypeID)
		eiv += float64(m.Quantity) * adjPrice
	}

	var jobCost float64
	if systemID != nil {
		costIndex, err := ac.repo.GetCostIndexForSystem(ac.ctx, *systemID, "manufacturing")
		if err == nil && costIndex > 0 {
			structBonus := calculator.ManufacturingStructureCostBonus(structure)
			jobCost = (eiv*costIndex*(1.0-structBonus) + eiv*calculator.SccSurchargeRate) * float64(runs)
		}
	}

	buildTime, err := ac.repo.GetBlueprintActivityTime(ac.ctx, item.BlueprintTypeID, "manufacturing")
	if err != nil {
		buildTime = 0
	}

	return &bomResult{
		MaterialCost: totalMatCost,
		JobCost:      jobCost,
		BuildTimeSec: buildTime * int64(runs),
	}, nil
}
