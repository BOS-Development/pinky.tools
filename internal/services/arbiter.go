package services

import (
	"context"
	"fmt"
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
	GetDemandStats(ctx context.Context, typeIDs []int64) (map[int64]*models.DemandStats, error)
	GetBlueprintForProduct(ctx context.Context, productTypeID int64) (int64, error)
	GetReactionBlueprintForProduct(ctx context.Context, productTypeID int64) (int64, error)
}

// ArbiterBOMRepository is the interface needed for building a BOM tree.
type ArbiterBOMRepository interface {
	GetBlueprintMaterialsForActivity(ctx context.Context, blueprintTypeID int64, activity string) ([]*models.BlueprintMaterial, error)
	GetBlueprintForProduct(ctx context.Context, productTypeID int64) (int64, error)
	GetMarketPricesForTypes(ctx context.Context, typeIDs []int64) (map[int64]*models.MarketPrice, error)
	GetAdjustedPricesForTypes(ctx context.Context, typeIDs []int64) (map[int64]float64, error)
	GetBlueprintActivityTime(ctx context.Context, blueprintTypeID int64, activity string) (int64, error)
	GetCostIndexForSystem(ctx context.Context, systemID int64, activity string) (float64, error)
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
	taxProfile     *models.ArbiterTaxProfile
	prices         map[int64]*models.MarketPrice
	adjustedPrices map[int64]float64
	repo           ArbiterScanRepository
	ctx            context.Context

	// Blueprint data caches — avoid repeated DB roundtrips
	bpMatsCache       map[string][]*models.BlueprintMaterial // key: "blueprintID:activity"
	bpProductCache    map[string]*models.BlueprintProduct    // key: "blueprintID:activity"
	bpTimeCache       map[string]int64                       // key: "blueprintID:activity"
	bpForProductCache map[int64]int64                        // typeID → blueprintTypeID (0 = not found)
	rxForProductCache map[int64]int64                        // typeID → reactionBlueprintTypeID (0 = not found)
	costIndexCache    map[string]float64                     // key: "systemID:activity"
}

// ensurePrices loads market prices for any typeIDs not already cached in ac.prices.
func (ac *arbiterContext) ensurePrices(typeIDs []int64) {
	missing := make([]int64, 0, len(typeIDs))
	for _, id := range typeIDs {
		if _, ok := ac.prices[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) == 0 {
		return
	}
	newPrices, err := ac.repo.GetMarketPricesForTypes(ac.ctx, missing)
	if err == nil {
		for k, v := range newPrices {
			ac.prices[k] = v
		}
	}
}

// ensureAdjustedPrices loads adjusted prices for any typeIDs not already cached.
func (ac *arbiterContext) ensureAdjustedPrices(typeIDs []int64) {
	missing := make([]int64, 0, len(typeIDs))
	for _, id := range typeIDs {
		if _, ok := ac.adjustedPrices[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) == 0 {
		return
	}
	newAdj, err := ac.repo.GetAdjustedPricesForTypes(ac.ctx, missing)
	if err == nil {
		for k, v := range newAdj {
			ac.adjustedPrices[k] = v
		}
	}
}

// loadPrice fetches the price record for a typeID, lazily fetching if missing.
func (ac *arbiterContext) loadPrice(typeID int64) *models.MarketPrice {
	if mp, ok := ac.prices[typeID]; ok {
		return mp
	}
	newPrices, err := ac.repo.GetMarketPricesForTypes(ac.ctx, []int64{typeID})
	if err == nil {
		for k, v := range newPrices {
			ac.prices[k] = v
		}
		if mp, ok := ac.prices[typeID]; ok {
			return mp
		}
	}
	return nil
}

// getPrice returns the Jita sell price for a type (used for decryptor cost lookups).
func (ac *arbiterContext) getPrice(typeID int64) float64 {
	mp := ac.loadPrice(typeID)
	if mp == nil {
		return 0
	}
	if mp.SellPrice != nil {
		return *mp.SellPrice
	}
	return 0
}

// getInputPrice returns price for input materials, respecting InputPriceType.
func (ac *arbiterContext) getInputPrice(typeID int64) float64 {
	mp := ac.loadPrice(typeID)
	if mp == nil {
		return 0
	}
	if ac.taxProfile.InputPriceType == "buy" {
		if mp.BuyPrice != nil {
			return *mp.BuyPrice
		}
		return 0
	}
	if mp.SellPrice != nil {
		return *mp.SellPrice
	}
	return 0
}

// getOutputPrice returns price for the T2 product being sold, respecting OutputPriceType.
func (ac *arbiterContext) getOutputPrice(typeID int64) float64 {
	mp := ac.loadPrice(typeID)
	if mp == nil {
		return 0
	}
	if ac.taxProfile.OutputPriceType == "buy" {
		if mp.BuyPrice != nil {
			return *mp.BuyPrice
		}
		return 0
	}
	if mp.SellPrice != nil {
		return *mp.SellPrice
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

// getBlueprintMaterials returns materials for a blueprint activity, using cache to avoid repeated DB calls.
func (ac *arbiterContext) getBlueprintMaterials(blueprintTypeID int64, activity string) ([]*models.BlueprintMaterial, error) {
	key := fmt.Sprintf("%d:%s", blueprintTypeID, activity)
	if mats, ok := ac.bpMatsCache[key]; ok {
		return mats, nil
	}
	mats, err := ac.repo.GetBlueprintMaterialsForActivity(ac.ctx, blueprintTypeID, activity)
	if err != nil {
		return nil, err
	}
	ac.bpMatsCache[key] = mats
	return mats, nil
}

// getBlueprintProduct returns the product for a blueprint activity, using cache to avoid repeated DB calls.
func (ac *arbiterContext) getBlueprintProduct(blueprintTypeID int64, activity string) (*models.BlueprintProduct, error) {
	key := fmt.Sprintf("%d:%s", blueprintTypeID, activity)
	if prod, ok := ac.bpProductCache[key]; ok {
		return prod, nil
	}
	prod, err := ac.repo.GetBlueprintProductForActivity(ac.ctx, blueprintTypeID, activity)
	if err != nil {
		return nil, err
	}
	ac.bpProductCache[key] = prod
	return prod, nil
}

// getBlueprintTime returns the activity time for a blueprint, using cache to avoid repeated DB calls.
func (ac *arbiterContext) getBlueprintTime(blueprintTypeID int64, activity string) int64 {
	key := fmt.Sprintf("%d:%s", blueprintTypeID, activity)
	if t, ok := ac.bpTimeCache[key]; ok {
		return t
	}
	t, _ := ac.repo.GetBlueprintActivityTime(ac.ctx, blueprintTypeID, activity)
	ac.bpTimeCache[key] = t
	return t
}

// getBlueprintForProduct returns the manufacturing blueprint ID for a product type, using cache.
// Returns 0 if not found.
func (ac *arbiterContext) getBlueprintForProduct(typeID int64) int64 {
	if id, ok := ac.bpForProductCache[typeID]; ok {
		return id
	}
	id, err := ac.repo.GetBlueprintForProduct(ac.ctx, typeID)
	if err != nil {
		id = 0
	}
	ac.bpForProductCache[typeID] = id
	return id
}

// getReactionForProduct returns the reaction blueprint ID for a product type, using cache.
// Returns 0 if not found.
func (ac *arbiterContext) getReactionForProduct(typeID int64) int64 {
	if id, ok := ac.rxForProductCache[typeID]; ok {
		return id
	}
	id, err := ac.repo.GetReactionBlueprintForProduct(ac.ctx, typeID)
	if err != nil {
		id = 0
	}
	ac.rxForProductCache[typeID] = id
	return id
}

// getCostIndex returns the cost index for a system and activity, using cache to avoid repeated DB calls.
func (ac *arbiterContext) getCostIndex(systemID int64, activity string) float64 {
	key := fmt.Sprintf("%d:%s", systemID, activity)
	if c, ok := ac.costIndexCache[key]; ok {
		return c
	}
	c, _ := ac.repo.GetCostIndexForSystem(ac.ctx, systemID, activity)
	ac.costIndexCache[key] = c
	return c
}

// defaultTaxProfile returns a default tax profile with sensible defaults.
func defaultTaxProfile() *models.ArbiterTaxProfile {
	return &models.ArbiterTaxProfile{
		InputPriceType:  "sell",
		OutputPriceType: "sell",
		SalesTaxRate:    3.6,
		BrokerFeeRate:   0,
	}
}

// ScanOpportunities runs the full Arbiter scan and returns ranked T2 opportunities.
func ScanOpportunities(ctx context.Context, userID int64, settings *models.ArbiterSettings, taxProfile *models.ArbiterTaxProfile, repo ArbiterScanRepository) (*models.ArbiterScanResult, error) {
	if taxProfile == nil {
		taxProfile = defaultTaxProfile()
	}
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

	// Bulk-load demand stats (30d history + order book volume).
	// Fallback: use daily_volume from market_prices when history is empty (new installs).
	demandStats, err := repo.GetDemandStats(ctx, typeIDs)
	if err != nil {
		demandStats = map[int64]*models.DemandStats{}
	}
	// Fill in missing demand from market_prices.daily_volume as fallback
	for _, typeID := range typeIDs {
		if ds, ok := demandStats[typeID]; ok && ds.DemandPerDay > 0 {
			continue
		}
		if mp, ok := prices[typeID]; ok && mp.DailyVolume != nil && *mp.DailyVolume > 0 {
			vol := float64(*mp.DailyVolume)
			if existing, ok := demandStats[typeID]; ok {
				existing.DemandPerDay = vol
				if existing.OrderBookVolume > 0 {
					existing.DaysOfSupply = float64(existing.OrderBookVolume) / vol
				}
			} else {
				demandStats[typeID] = &models.DemandStats{
					TypeID:       typeID,
					DemandPerDay: vol,
				}
			}
		}
	}

	ac := &arbiterContext{
		settings:          settings,
		taxProfile:        taxProfile,
		prices:            prices,
		adjustedPrices:    adjustedPrices,
		repo:              repo,
		ctx:               ctx,
		bpMatsCache:       map[string][]*models.BlueprintMaterial{},
		bpProductCache:    map[string]*models.BlueprintProduct{},
		bpTimeCache:       map[string]int64{},
		bpForProductCache: map[int64]int64{},
		rxForProductCache: map[int64]int64{},
		costIndexCache:    map[string]float64{},
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
		// Apply demand stats
		if ds, ok := demandStats[item.ProductTypeID]; ok {
			opp.DemandPerDay = math.Round(ds.DemandPerDay*100) / 100
			opp.DaysOfSupply = math.Round(ds.DaysOfSupply*100) / 100
		}
		opportunities = append(opportunities, opp)
	}

	sort.Slice(opportunities, func(i, j int) bool {
		return opportunities[i].Profit > opportunities[j].Profit
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
	buildTime := ac.getBlueprintTime(item.BlueprintTypeID, "manufacturing")

	// Build all decryptor options: no-decryptor + each decryptor
	allOptions := make([]*models.DecryptorOption, 0, len(decryptors)+1)

	noDecOpt, err := calculateDecryptorOption(ac, item, &models.Decryptor{
		ProbabilityMultiplier: 1.0,
	}, nil, effectiveChance, copyAndDatacoreCost, buildTime)
	if err == nil && noDecOpt != nil {
		allOptions = append(allOptions, noDecOpt)
	}

	for _, dec := range decryptors {
		d := dec
		opt, err := calculateDecryptorOption(ac, item, d, &d.TypeID, effectiveChance, copyAndDatacoreCost, buildTime)
		if err == nil && opt != nil {
			allOptions = append(allOptions, opt)
		}
	}

	if len(allOptions) == 0 {
		return nil, nil
	}

	var best *models.DecryptorOption
	for _, opt := range allOptions {
		if best == nil || opt.ROI > best.ROI {
			best = opt
		}
	}

	outputPrice := ac.getOutputPrice(item.ProductTypeID)
	revenue := outputPrice * float64(best.ResultingRuns)
	salesTaxRate := ac.taxProfile.SalesTaxRate / 100.0
	brokerFeeRate := ac.taxProfile.BrokerFeeRate / 100.0
	salesTax := revenue * salesTaxRate
	brokerFee := revenue * brokerFeeRate

	return &models.ArbiterOpportunity{
		ProductTypeID: item.ProductTypeID,
		ProductName:   item.ProductName,
		Category:      item.Category,
		TechLevel:     "Tech II",
		JitaSellPrice: jitaSell,
		JitaBuyPrice:  jitaBuy,
		Duration:      best.BuildTimeSec,
		Runs:          best.ResultingRuns,
		BaseRuns:      item.BaseResultRuns,
		ME:            best.ME,
		TE:            best.TE,
		MaterialCost:  best.MaterialCost,
		JobCost:       best.JobCost,
		InventionCost: best.InventionCost,
		TotalCost:     best.TotalCost,
		Revenue:       math.Round(revenue*100) / 100,
		SalesTax:      math.Round(salesTax*100) / 100,
		BrokerFee:     math.Round(brokerFee*100) / 100,
		Profit:        best.Profit,
		ROI:           best.ROI,
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
	outputPrice := ac.getOutputPrice(item.ProductTypeID)
	revenue := outputPrice * float64(resultRuns)
	salesTaxRate := ac.taxProfile.SalesTaxRate / 100.0
	brokerFeeRate := ac.taxProfile.BrokerFeeRate / 100.0
	salesTax := revenue * salesTaxRate
	brokerFee := revenue * brokerFeeRate
	profit := revenue - totalCost - salesTax - brokerFee
	roi := 0.0
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

	resultTE := 4 + decryptor.TEModifier
	if resultTE < 0 {
		resultTE = 0
	}
	if resultTE > 20 {
		resultTE = 20
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
		ME:                    resultME,
		TE:                    resultTE,
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
	datecoreMats, err := ac.getBlueprintMaterials(item.T1BlueprintTypeID, "invention")
	if err != nil {
		return 0, errors.Wrap(err, "failed to get datacore materials")
	}

	// Ensure datacore prices are loaded
	datacoreTypeIDs := make([]int64, 0, len(datecoreMats))
	for _, m := range datecoreMats {
		datacoreTypeIDs = append(datacoreTypeIDs, m.TypeID)
	}
	ac.ensurePrices(datacoreTypeIDs)

	var dataCoreCost float64
	for _, m := range datecoreMats {
		dataCoreCost += ac.getInputPrice(m.TypeID) * float64(m.Quantity)
	}

	// Copy cost: approximate using invention system cost index if configured
	var copyCost float64
	if ac.settings.InventionSystemID != nil {
		copyIdx := ac.getCostIndex(*ac.settings.InventionSystemID, "copying")
		if copyIdx > 0 {
			t1Product, err := ac.getBlueprintProduct(item.T1BlueprintTypeID, "manufacturing")
			if err == nil && t1Product != nil {
				adjPrice := ac.getAdjustedPrice(t1Product.TypeID)
				copyCost = adjPrice * copyIdx
			}
		}
	}

	return dataCoreCost + copyCost, nil
}

// levelSettings holds the structure/rig/system configuration for a given manufacturing depth.
type levelSettings struct {
	structure string
	rig       string
	systemID  *int64
}

// settingsForDepth returns the structure settings to apply at a given recursion depth.
// depth 0 = final T2 build, depth 1 = T2 component build, depth >= 2 = reaction.
func (ac *arbiterContext) settingsForDepth(depth int) levelSettings {
	s := ac.settings
	switch depth {
	case 0:
		return levelSettings{s.FinalStructure, s.FinalRig, s.FinalSystemID}
	case 1:
		return levelSettings{s.ComponentStructure, s.ComponentRig, s.ComponentSystemID}
	default: // depth >= 2 = reactions
		return levelSettings{s.ReactionStructure, s.ReactionRig, s.ReactionSystemID}
	}
}

// calcChainCost returns the total material cost and job cost to produce qty units of a product
// using blueprintTypeID at the given depth. Recursively builds sub-components up to maxDepth.
func calcChainCost(ac *arbiterContext, blueprintTypeID int64, qty int, depth int) (materialCost float64, jobCost float64, buildTimeSec int64) {
	const maxDepth = 4

	if depth >= maxDepth || blueprintTypeID == 0 {
		return 0, 0, 0
	}

	lvl := ac.settingsForDepth(depth)

	// Try manufacturing activity first, then reaction
	mats, err := ac.getBlueprintMaterials(blueprintTypeID, "manufacturing")
	activity := "manufacturing"
	if err != nil || len(mats) == 0 {
		mats, err = ac.getBlueprintMaterials(blueprintTypeID, "reaction")
		activity = "reaction"
		if err != nil || len(mats) == 0 {
			return 0, 0, 0
		}
	}

	// Ensure prices are loaded for all materials
	matTypeIDs := make([]int64, 0, len(mats))
	for _, m := range mats {
		matTypeIDs = append(matTypeIDs, m.TypeID)
	}
	ac.ensurePrices(matTypeIDs)
	ac.ensureAdjustedPrices(matTypeIDs)

	// Determine how many units this blueprint produces per run (reactions: 50–150; manufacturing: usually 1).
	// We need this to convert "units needed" into "runs needed".
	productQtyPerRun := 1
	if prod, perr := ac.getBlueprintProduct(blueprintTypeID, activity); perr == nil && prod != nil && prod.Quantity > 1 {
		productQtyPerRun = prod.Quantity
	}
	runs := (qty + productQtyPerRun - 1) / productQtyPerRun // ceil(qty / productQtyPerRun)

	var meFactor float64
	if activity == "manufacturing" {
		meFactor = calculator.ComputeManufacturingME(0, lvl.structure, lvl.rig, "null")
	} else {
		meFactor = calculator.ComputeMEFactor(lvl.rig, "null")
	}

	var totalMatCost, totalJobCost float64
	var eiv float64

	for _, m := range mats {
		batchQty := int(calculator.ComputeBatchQty(runs, m.Quantity, meFactor))

		// Try to find a manufacturing sub-blueprint first, then a reaction blueprint
		subBpID := ac.getBlueprintForProduct(m.TypeID)
		if subBpID == 0 {
			subBpID = ac.getReactionForProduct(m.TypeID)
		}

		if subBpID != 0 {
			subMat, subJob, _ := calcChainCost(ac, subBpID, batchQty, depth+1)
			totalMatCost += subMat
			totalJobCost += subJob
		} else {
			totalMatCost += ac.getInputPrice(m.TypeID) * float64(batchQty)
		}

		adjPrice := ac.getAdjustedPrice(m.TypeID)
		eiv += float64(m.Quantity) * adjPrice
	}

	// Job cost for this level
	if lvl.systemID != nil {
		costIdx := ac.getCostIndex(*lvl.systemID, activity)
		if costIdx > 0 {
			if activity == "manufacturing" {
				structBonus := calculator.ManufacturingStructureCostBonus(lvl.structure)
				totalJobCost += (eiv*costIdx*(1.0-structBonus) + eiv*calculator.SccSurchargeRate) * float64(runs)
			} else {
				// Reactions: EIV × (cost_index + scc_surcharge) per run
				totalJobCost += eiv * (costIdx + calculator.SccSurchargeRate) * float64(runs)
			}
		}
	}

	buildTime := ac.getBlueprintTime(blueprintTypeID, activity)
	totalBuildTime := buildTime * int64(runs)

	// Pro-rate costs: we computed for `runs × productQtyPerRun` units,
	// but only `qty` are consumed. Scale proportionally.
	// For manufacturing (productQtyPerRun=1), scale=1.0 (no change).
	if productQtyPerRun > 1 && runs > 0 {
		scale := float64(qty) / float64(runs*productQtyPerRun)
		totalMatCost *= scale
		totalJobCost *= scale
		totalBuildTime = int64(math.Round(float64(totalBuildTime) * scale))
	}

	return totalMatCost, totalJobCost, totalBuildTime
}

// calculateFinalBOM computes the full production chain material + job cost for building the T2 product.
// It applies the correct ME from the T2 BPC at the final level, then recursively builds sub-components
// (T2 components at depth=1, reactions at depth=2) using the appropriate structure settings.
func calculateFinalBOM(ac *arbiterContext, item *models.T2BlueprintScanItem, me int, runs int) (*bomResult, error) {
	structure := ac.settings.FinalStructure
	rig := ac.settings.FinalRig
	systemID := ac.settings.FinalSystemID

	meFactor := calculator.ComputeManufacturingME(me, structure, rig, "null")

	mats, err := ac.getBlueprintMaterials(item.BlueprintTypeID, "manufacturing")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get T2 manufacturing materials")
	}

	// Ensure prices are loaded for direct materials
	matTypeIDs := make([]int64, 0, len(mats))
	for _, m := range mats {
		matTypeIDs = append(matTypeIDs, m.TypeID)
	}
	ac.ensurePrices(matTypeIDs)
	ac.ensureAdjustedPrices(matTypeIDs)

	var totalMatCost, totalJobCost float64
	var eiv float64

	for _, m := range mats {
		batchQty := int(calculator.ComputeBatchQty(runs, m.Quantity, meFactor))

		// Recurse into sub-components starting at depth=1 (component level)
		subBpID := ac.getBlueprintForProduct(m.TypeID)
		if subBpID == 0 {
			subBpID = ac.getReactionForProduct(m.TypeID)
		}

		if subBpID != 0 {
			subMat, subJob, _ := calcChainCost(ac, subBpID, batchQty, 1)
			totalMatCost += subMat
			totalJobCost += subJob
		} else {
			totalMatCost += ac.getInputPrice(m.TypeID) * float64(batchQty)
		}

		adjPrice := ac.getAdjustedPrice(m.TypeID)
		eiv += float64(m.Quantity) * adjPrice
	}

	// Final manufacturing job cost
	if systemID != nil {
		costIndex := ac.getCostIndex(*systemID, "manufacturing")
		if costIndex > 0 {
			structBonus := calculator.ManufacturingStructureCostBonus(structure)
			totalJobCost += (eiv*costIndex*(1.0-structBonus) + eiv*calculator.SccSurchargeRate) * float64(runs)
		}
	}

	buildTime := ac.getBlueprintTime(item.BlueprintTypeID, "manufacturing")

	return &bomResult{
		MaterialCost: totalMatCost,
		JobCost:      totalJobCost,
		BuildTimeSec: buildTime * int64(runs),
	}, nil
}

// deriveSecurityClassFromSettings returns a security class string from settings.
// Since security fields were removed from the model, we use a sensible default.
func deriveSecurityClassFromSettings(ac *arbiterContext) string {
	// Without the security columns, we can't determine it from settings.
	// The security class must be derived from the system at runtime.
	// Default to "null" (most conservative for ME bonus calculations).
	return "null"
}

// --- BOM tree ---

// bomTreeContext holds shared state for building a BOM tree.
type bomTreeContext struct {
	ctx            context.Context
	repo           ArbiterBOMRepository
	settings       *models.ArbiterSettings
	prices         map[int64]*models.MarketPrice
	adjustedPrices map[int64]float64
	blacklist      map[int64]bool
	whitelist      map[int64]bool
	assets         map[int64]int64
	inputPriceType string
	buildAll       bool
}

// getInputPrice returns the price to use for material cost based on inputPriceType.
func (btc *bomTreeContext) getInputPrice(typeID int64) float64 {
	mp, ok := btc.prices[typeID]
	if !ok {
		// Lazily fetch
		newPrices, err := btc.repo.GetMarketPricesForTypes(btc.ctx, []int64{typeID})
		if err == nil {
			for k, v := range newPrices {
				btc.prices[k] = v
			}
		}
		mp = btc.prices[typeID]
	}
	if mp == nil {
		return 0
	}
	if btc.inputPriceType == "buy" {
		if mp.BuyPrice != nil {
			return *mp.BuyPrice
		}
		return 0
	}
	// default: sell
	if mp.SellPrice != nil {
		return *mp.SellPrice
	}
	return 0
}

// getBuyPrice returns the sell price (what it costs to buy from market).
func (btc *bomTreeContext) getBuyPrice(typeID int64) float64 {
	mp, ok := btc.prices[typeID]
	if !ok {
		newPrices, err := btc.repo.GetMarketPricesForTypes(btc.ctx, []int64{typeID})
		if err == nil {
			for k, v := range newPrices {
				btc.prices[k] = v
			}
		}
		mp = btc.prices[typeID]
	}
	if mp == nil {
		return 0
	}
	if mp.SellPrice != nil {
		return *mp.SellPrice
	}
	return 0
}

// BuildBOMTree builds a full recursive BOM tree for a product.
// blueprintTypeID is the blueprint used to manufacture the product.
// qty is the number of units needed.
// depth prevents infinite recursion (max 10).
func BuildBOMTree(
	ctx context.Context,
	blueprintTypeID int64,
	productTypeID int64,
	productName string,
	qty int64,
	me int,
	repo ArbiterBOMRepository,
	settings *models.ArbiterSettings,
	blacklist map[int64]bool,
	whitelist map[int64]bool,
	assets map[int64]int64,
	inputPriceType string,
	buildAll bool,
) (*models.BOMNode, error) {
	prices, err := repo.GetMarketPricesForTypes(ctx, []int64{productTypeID})
	if err != nil {
		prices = map[int64]*models.MarketPrice{}
	}
	adjPrices, err := repo.GetAdjustedPricesForTypes(ctx, []int64{productTypeID})
	if err != nil {
		adjPrices = map[int64]float64{}
	}

	btc := &bomTreeContext{
		ctx:            ctx,
		repo:           repo,
		settings:       settings,
		prices:         prices,
		adjustedPrices: adjPrices,
		blacklist:      blacklist,
		whitelist:      whitelist,
		assets:         assets,
		inputPriceType: inputPriceType,
		buildAll:       buildAll,
	}

	return buildBOMNode(btc, blueprintTypeID, productTypeID, productName, qty, me, 0)
}

// buildBOMNode recursively builds a BOM node.
func buildBOMNode(
	btc *bomTreeContext,
	blueprintTypeID int64,
	productTypeID int64,
	productName string,
	qty int64,
	me int,
	depth int,
) (*models.BOMNode, error) {
	const maxDepth = 10

	buyPrice := btc.getBuyPrice(productTypeID)
	available := btc.assets[productTypeID]
	needed := qty
	delta := needed - available
	if delta < 0 {
		delta = 0
	}

	isBlacklisted := btc.blacklist[productTypeID]
	isWhitelisted := btc.whitelist[productTypeID]

	// build_all forces every node to be treated as whitelisted unless explicitly blacklisted
	if btc.buildAll && !isBlacklisted {
		isWhitelisted = true
	}

	node := &models.BOMNode{
		TypeID:        productTypeID,
		Name:          productName,
		Quantity:      qty,
		Available:     available,
		Needed:        needed,
		Delta:         delta,
		UnitBuyPrice:  buyPrice,
		IsBlacklisted: isBlacklisted,
		IsWhitelisted: isWhitelisted,
		Children:      []*models.BOMNode{},
	}

	// If depth limit reached, or no blueprint, or blacklisted → buy
	if depth >= maxDepth || blueprintTypeID == 0 || isBlacklisted {
		if isBlacklisted {
			node.Decision = "buy_override"
		} else {
			node.Decision = "buy"
		}
		return node, nil
	}

	// Get materials for this blueprint
	mats, err := btc.repo.GetBlueprintMaterialsForActivity(btc.ctx, blueprintTypeID, "manufacturing")
	if err != nil {
		node.Decision = "buy"
		return node, nil
	}
	if len(mats) == 0 {
		node.Decision = "buy"
		return node, nil
	}

	// Calculate build cost for the materials
	structure := btc.settings.FinalStructure
	rig := btc.settings.FinalRig
	meFactor := calculator.ComputeManufacturingME(me, structure, rig, "null")

	var buildCost float64
	children := []*models.BOMNode{}

	for _, mat := range mats {
		batchQty := calculator.ComputeBatchQty(int(qty), mat.Quantity, meFactor)
		matBuyPrice := btc.getBuyPrice(mat.TypeID)
		matBuildCost := matBuyPrice // default: buy

		// Check if this material has a blueprint for sub-building
		subBpID, err := btc.repo.GetBlueprintForProduct(btc.ctx, mat.TypeID)
		if err != nil {
			subBpID = 0
		}

		var childNode *models.BOMNode
		if subBpID != 0 && depth+1 < maxDepth && !btc.blacklist[mat.TypeID] {
			// Recurse to build sub-tree
			childNode, err = buildBOMNode(btc, subBpID, mat.TypeID, mat.TypeName, int64(batchQty), 0, depth+1)
			if err != nil || childNode == nil {
				childNode = &models.BOMNode{
					TypeID:       mat.TypeID,
					Name:         mat.TypeName,
					Quantity:     int64(batchQty),
					Available:    btc.assets[mat.TypeID],
					Needed:       int64(batchQty),
					UnitBuyPrice: matBuyPrice,
					Decision:     "buy",
					Children:     []*models.BOMNode{},
				}
			}
			if childNode.Decision == "build" || childNode.Decision == "build_override" {
				matBuildCost = childNode.UnitBuildCost
			}
		} else {
			childNode = &models.BOMNode{
				TypeID:       mat.TypeID,
				Name:         mat.TypeName,
				Quantity:     int64(batchQty),
				Available:    btc.assets[mat.TypeID],
				Needed:       int64(batchQty),
				UnitBuyPrice: matBuyPrice,
				Decision:     "buy",
				Children:     []*models.BOMNode{},
			}
			if btc.blacklist[mat.TypeID] {
				childNode.Decision = "buy_override"
			}
		}

		buildCost += matBuildCost * float64(batchQty)
		children = append(children, childNode)
	}

	// Compute per-unit build cost
	var unitBuildCost float64
	if qty > 0 {
		unitBuildCost = buildCost / float64(qty)
	}
	node.UnitBuildCost = math.Round(unitBuildCost*100) / 100
	node.Children = children

	// Decide: build or buy
	if isWhitelisted {
		node.Decision = "build_override"
	} else if isBlacklisted {
		node.Decision = "buy_override"
	} else if unitBuildCost > 0 && buyPrice > 0 {
		if unitBuildCost < buyPrice {
			node.Decision = "build"
		} else {
			node.Decision = "buy"
		}
	} else if unitBuildCost > 0 {
		node.Decision = "build"
	} else {
		node.Decision = "buy"
	}

	return node, nil
}
