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
	GetMarketPricesLastUpdated(ctx context.Context) (*time.Time, error)
}

// ArbiterBOMRepository is the interface needed for building a BOM tree.
type ArbiterBOMRepository interface {
	GetBlueprintMaterialsForActivity(ctx context.Context, blueprintTypeID int64, activity string) ([]*models.BlueprintMaterial, error)
	GetBlueprintProductForActivity(ctx context.Context, blueprintTypeID int64, activity string) (*models.BlueprintProduct, error)
	GetBlueprintForProduct(ctx context.Context, productTypeID int64) (int64, error)
	GetReactionBlueprintForProduct(ctx context.Context, productTypeID int64) (int64, error)
	GetMarketPricesForTypes(ctx context.Context, typeIDs []int64) (map[int64]*models.MarketPrice, error)
	GetAdjustedPricesForTypes(ctx context.Context, typeIDs []int64) (map[int64]float64, error)
	GetBlueprintActivityTime(ctx context.Context, blueprintTypeID int64, activity string) (int64, error)
	GetCostIndexForSystem(ctx context.Context, systemID int64, activity string) (float64, error)
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
	buildAll          bool                                   // when false, buy sub-components from market instead of building full chain
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
// Falls back to the other price type if the preferred one is unavailable.
func (ac *arbiterContext) getInputPrice(typeID int64) float64 {
	mp := ac.loadPrice(typeID)
	if mp == nil {
		return 0
	}
	if ac.taxProfile.InputPriceType == "buy" {
		if mp.BuyPrice != nil {
			return *mp.BuyPrice
		}
		// fallback: use sell price if no buy orders exist
		if mp.SellPrice != nil {
			return *mp.SellPrice
		}
		return 0
	}
	if mp.SellPrice != nil {
		return *mp.SellPrice
	}
	// fallback: use buy price if no sell orders exist
	if mp.BuyPrice != nil {
		return *mp.BuyPrice
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
func ScanOpportunities(ctx context.Context, userID int64, settings *models.ArbiterSettings, taxProfile *models.ArbiterTaxProfile, buildAll bool, repo ArbiterScanRepository) (*models.ArbiterScanResult, error) {
	if taxProfile == nil {
		taxProfile = defaultTaxProfile()
	}
	items, err := repo.GetT2BlueprintsForScan(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get T2 blueprints")
	}

	if len(items) == 0 {
		pricesUpdatedAt, _ := repo.GetMarketPricesLastUpdated(ctx)
		return &models.ArbiterScanResult{
			Opportunities:   []*models.ArbiterOpportunity{},
			GeneratedAt:     time.Now(),
			TotalScanned:    0,
			PricesUpdatedAt: pricesUpdatedAt,
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
		buildAll:          buildAll,
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

	// Create one shared BOM cache for the entire scan so that sub-components shared across
	// multiple T2 items (e.g., Nanoelectrical Microprocessors) are fetched from the DB at most once.
	shared := NewBOMSharedCache()

	for _, item := range items {
		opp, err := calculateOpportunity(ac, item, decryptors, bestChar, shared)
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

	pricesUpdatedAt, _ := repo.GetMarketPricesLastUpdated(ctx)

	result := &models.ArbiterScanResult{
		Opportunities:   opportunities,
		GeneratedAt:     time.Now(),
		TotalScanned:    len(items),
		PricesUpdatedAt: pricesUpdatedAt,
	}
	if bestChar != nil {
		result.BestCharacterID = bestChar.CharacterID
		result.BestCharacterName = bestChar.Name
	}
	return result, nil
}

func calculateOpportunity(ac *arbiterContext, item *models.T2BlueprintScanItem, decryptors []*models.Decryptor, bestChar *models.InventionCharacter, shared *BOMSharedCache) (*models.ArbiterOpportunity, error) {
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

	// Build all decryptor options: no-decryptor + each decryptor.
	// The shared BOM cache is passed in from ScanOpportunities so that sub-components used by
	// multiple T2 items are fetched from the DB at most once across the entire scan.
	allOptions := make([]*models.DecryptorOption, 0, len(decryptors)+1)

	noDecOpt, err := calculateDecryptorOption(ac, item, &models.Decryptor{
		ProbabilityMultiplier: 1.0,
	}, nil, effectiveChance, copyAndDatacoreCost, buildTime, shared)
	if err == nil && noDecOpt != nil {
		allOptions = append(allOptions, noDecOpt)
	}

	for _, dec := range decryptors {
		d := dec
		opt, err := calculateDecryptorOption(ac, item, d, &d.TypeID, effectiveChance, copyAndDatacoreCost, buildTime, shared)
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
		Profit:             best.Profit,
		ROI:                best.ROI,
		BestDecryptor:      best,
		AllDecryptors:      allOptions,
		InventionMaterials: best.InventionMaterials,
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
	shared *BOMSharedCache,
) (*models.DecryptorOption, error) {
	chanceMod := effectiveChance * decryptor.ProbabilityMultiplier
	chanceMod = math.Round(chanceMod*100) / 100

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

	// decryptorAdjPrice contributes to the invention EIV job cost per attempt
	var decryptorJobCost float64
	if decryptorTypeID != nil && ac.settings.InventionSystemID != nil {
		invIdx := ac.getCostIndex(*ac.settings.InventionSystemID, "invention")
		if invIdx > 0 {
			facilityTaxRate := ac.settings.InventionFacilityTax / 100.0
			adjPrice := ac.getAdjustedPrice(*decryptorTypeID)
			decryptorJobCost = adjPrice*invIdx*(1.0+facilityTaxRate) + adjPrice*calculator.SccSurchargeRate
		}
	}

	var inventionCost float64
	if chanceMod > 0 {
		inventionCost = (copyAndDatacoreCost + decryptorCost + decryptorJobCost) / chanceMod
	}

	// Resolve decryptor name before building the materials list so it can be used there.
	name := "No Decryptor"
	if decryptor.Name != "" {
		name = decryptor.Name
	}

	// Build invention materials list: datacores scaled by 1/success_rate, plus decryptor if present.
	inventionMats, _ := ac.getBlueprintMaterials(item.T1BlueprintTypeID, "invention")
	inventionMaterials := make([]*models.InventionMaterial, 0, len(inventionMats)+1)
	for _, m := range inventionMats {
		var qty int64
		if chanceMod > 0 {
			qty = int64(math.Ceil(float64(m.Quantity) / chanceMod))
		}
		inventionMaterials = append(inventionMaterials, &models.InventionMaterial{
			TypeID:    m.TypeID,
			Name:      m.TypeName,
			Quantity:  qty,
			UnitPrice: ac.getPrice(m.TypeID),
		})
	}
	if decryptorTypeID != nil {
		var qty int64
		if chanceMod > 0 {
			qty = int64(math.Ceil(1.0 / chanceMod))
		}
		inventionMaterials = append(inventionMaterials, &models.InventionMaterial{
			TypeID:    *decryptorTypeID,
			Name:      name,
			Quantity:  qty,
			UnitPrice: decryptorCost,
		})
	}

	// BOM for the final T2 product using the T2 blueprint and result ME.
	// Pass the scan repo directly — ArbiterScanRepository is a superset of ArbiterBOMRepository.
	// shared cache is populated once per T2 item and reused across all decryptor options.
	tree, treeErr := BuildBOMTree(
		ac.ctx,
		item.BlueprintTypeID,
		item.ProductTypeID,
		item.ProductName,
		int64(resultRuns),
		resultME,
		ac.repo,
		ac.settings,
		map[int64]bool{},   // no blacklist during scan
		map[int64]bool{},   // no whitelist during scan
		map[int64]int64{},  // no assets during scan
		ac.taxProfile.InputPriceType,
		ac.buildAll,
		shared,
	)
	if treeErr != nil || tree == nil {
		return nil, treeErr
	}
	materialCost := tree.MaterialCost
	jobCost := tree.TotalJobCost

	totalCost := materialCost + jobCost + inventionCost
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

	var iskPerDay float64
	if buildTimeSec > 0 {
		iskPerDay = profit / (float64(buildTimeSec) / 86400.0)
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
		MaterialCost:          math.Round((materialCost+inventionCost)*100) / 100,
		JobCost:               math.Round(jobCost*100) / 100,
		TotalCost:             math.Round(totalCost*100) / 100,
		Profit:                math.Round(profit*100) / 100,
		ROI:                   math.Round(roi*100) / 100,
		ISKPerDay:             math.Round(iskPerDay*100) / 100,
		BuildTimeSec:          buildTimeSec,
		InventionMaterials:    inventionMaterials,
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

	// facilityTaxRate is shared by both the copy job and the invention job (same structure).
	facilityTaxRate := ac.settings.InventionFacilityTax / 100.0

	// Copy cost: EIV of T1 product × (cost_index + facility_tax + scc_surcharge)
	var copyCost float64
	if ac.settings.InventionSystemID != nil {
		copyIdx := ac.getCostIndex(*ac.settings.InventionSystemID, "copying")
		if copyIdx > 0 {
			t1Product, err := ac.getBlueprintProduct(item.T1BlueprintTypeID, "manufacturing")
			if err == nil && t1Product != nil {
				adjPrice := ac.getAdjustedPrice(t1Product.TypeID)
				copyCost = adjPrice*copyIdx*(1.0+facilityTaxRate) + adjPrice*calculator.SccSurchargeRate
			}
		}
	}

	// Invention job cost: EIV of datacores × (cost_index + facility_tax + scc_surcharge)
	if ac.settings.InventionSystemID != nil {
		invIdx := ac.getCostIndex(*ac.settings.InventionSystemID, "invention")
		if invIdx > 0 {
			var invEIV float64
			for _, m := range datecoreMats {
				invEIV += float64(m.Quantity) * ac.getAdjustedPrice(m.TypeID)
			}
			inventionJobCost := invEIV*invIdx*(1.0+facilityTaxRate) + invEIV*calculator.SccSurchargeRate
			dataCoreCost += inventionJobCost
		}
	}

	return dataCoreCost + copyCost, nil
}

// --- BOM tree ---

// BOMSharedCache holds DB query results that can be reused across multiple BuildBOMTree calls.
// Pass a single instance through calculateDecryptorOption calls for a given T2 item so that
// blueprint materials, product lookups, prices, and cost indices are fetched at most once per scan.
type BOMSharedCache struct {
	bpMatsCache       map[string][]*models.BlueprintMaterial // key: "blueprintID:activity"
	bpProductCache    map[string]*models.BlueprintProduct    // key: "blueprintID:activity"
	marketPrices      map[int64]*models.MarketPrice          // key: typeID
	adjPrices         map[int64]float64                      // key: typeID
	costIndices       map[string]float64                     // key: "systemID:activity"
	bpForProductCache map[int64]int64                        // mat typeID → manufacturing blueprint typeID (0 if none)
	rxForProductCache map[int64]int64                        // mat typeID → reaction blueprint typeID (0 if none)
}

// NewBOMSharedCache creates an empty shared cache for use across multiple BuildBOMTree calls.
// Callers that build the same blueprints repeatedly (e.g., scanning decryptor options)
// should create one cache and pass it to every BuildBOMTree call.
func NewBOMSharedCache() *BOMSharedCache {
	return &BOMSharedCache{
		bpMatsCache:       map[string][]*models.BlueprintMaterial{},
		bpProductCache:    map[string]*models.BlueprintProduct{},
		marketPrices:      map[int64]*models.MarketPrice{},
		adjPrices:         map[int64]float64{},
		costIndices:       map[string]float64{},
		bpForProductCache: map[int64]int64{},
		rxForProductCache: map[int64]int64{},
	}
}

// bomTreeContext holds shared state for building a BOM tree.
type bomTreeContext struct {
	ctx               context.Context
	repo              ArbiterBOMRepository
	settings          *models.ArbiterSettings
	prices            map[int64]*models.MarketPrice
	adjustedPrices    map[int64]float64
	blacklist         map[int64]bool
	whitelist         map[int64]bool
	assets            map[int64]int64
	inputPriceType    string
	buildAll          bool
	bpProductCache    map[string]*models.BlueprintProduct
	bpMatsCache       map[string][]*models.BlueprintMaterial // key: "blueprintID:activity"
	costIndexCache    map[string]float64                     // key: "systemID:activity"
	bpForProductCache map[int64]int64                        // mat typeID → manufacturing blueprint typeID (0 if none)
	rxForProductCache map[int64]int64                        // mat typeID → reaction blueprint typeID (0 if none)
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
		// fallback: use sell price if no buy orders exist
		if mp.SellPrice != nil {
			return *mp.SellPrice
		}
		return 0
	}
	// default: sell
	if mp.SellPrice != nil {
		return *mp.SellPrice
	}
	// fallback: use buy price if no sell orders exist
	if mp.BuyPrice != nil {
		return *mp.BuyPrice
	}
	return 0
}

// getBuyPrice returns the price to purchase this item from the market,
// respecting inputPriceType (buy or sell order pricing).
func (btc *bomTreeContext) getBuyPrice(typeID int64) float64 {
	return btc.getInputPrice(typeID)
}

// getAdjustedPrice returns the adjusted price for a type, lazily loading from the repo if missing.
func (btc *bomTreeContext) getAdjustedPrice(typeID int64) float64 {
	if p, ok := btc.adjustedPrices[typeID]; ok {
		return p
	}
	newAdj, err := btc.repo.GetAdjustedPricesForTypes(btc.ctx, []int64{typeID})
	if err == nil {
		for k, v := range newAdj {
			btc.adjustedPrices[k] = v
		}
		return btc.adjustedPrices[typeID]
	}
	return 0
}

// getBlueprintMaterials returns materials for a blueprint activity, using cache to avoid repeated DB calls.
func (btc *bomTreeContext) getBlueprintMaterials(blueprintTypeID int64, activity string) ([]*models.BlueprintMaterial, error) {
	key := fmt.Sprintf("%d:%s", blueprintTypeID, activity)
	if mats, ok := btc.bpMatsCache[key]; ok {
		return mats, nil
	}
	mats, err := btc.repo.GetBlueprintMaterialsForActivity(btc.ctx, blueprintTypeID, activity)
	if err != nil {
		return nil, err
	}
	btc.bpMatsCache[key] = mats
	return mats, nil
}

// getCostIndex returns the cost index for a system and activity, using cache to avoid repeated DB calls.
func (btc *bomTreeContext) getCostIndex(systemID int64, activity string) float64 {
	key := fmt.Sprintf("%d:%s", systemID, activity)
	if c, ok := btc.costIndexCache[key]; ok {
		return c
	}
	c, _ := btc.repo.GetCostIndexForSystem(btc.ctx, systemID, activity)
	btc.costIndexCache[key] = c
	return c
}

// BuildBOMTree builds a full recursive BOM tree for a product.
// blueprintTypeID is the blueprint used to manufacture the product.
// qty is the number of units needed.
// depth prevents infinite recursion (max 10).
// shared is an optional cross-call cache; pass nil for single calls (e.g., the BOM display endpoint)
// or a *BOMSharedCache created via newBOMSharedCache() when calling in a loop (e.g., scan path).
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
	shared *BOMSharedCache,
) (*models.BOMNode, error) {
	var prices map[int64]*models.MarketPrice
	var adjPrices map[int64]float64
	var bpProductCache map[string]*models.BlueprintProduct
	var bpMatsCache map[string][]*models.BlueprintMaterial
	var costIndexCache map[string]float64
	var bpForProductCache map[int64]int64
	var rxForProductCache map[int64]int64

	if shared != nil {
		// Use shared caches — add the product price if not already cached
		if _, ok := shared.marketPrices[productTypeID]; !ok {
			newPrices, err := repo.GetMarketPricesForTypes(ctx, []int64{productTypeID})
			if err == nil {
				for k, v := range newPrices {
					shared.marketPrices[k] = v
				}
			}
		}
		if _, ok := shared.adjPrices[productTypeID]; !ok {
			newAdj, err := repo.GetAdjustedPricesForTypes(ctx, []int64{productTypeID})
			if err == nil {
				for k, v := range newAdj {
					shared.adjPrices[k] = v
				}
			}
		}
		prices = shared.marketPrices
		adjPrices = shared.adjPrices
		bpProductCache = shared.bpProductCache
		bpMatsCache = shared.bpMatsCache
		costIndexCache = shared.costIndices
		bpForProductCache = shared.bpForProductCache
		rxForProductCache = shared.rxForProductCache
	} else {
		// No shared cache — create fresh local maps (single-call path, e.g., BOM display endpoint)
		var err error
		prices, err = repo.GetMarketPricesForTypes(ctx, []int64{productTypeID})
		if err != nil {
			prices = map[int64]*models.MarketPrice{}
		}
		adjPrices, err = repo.GetAdjustedPricesForTypes(ctx, []int64{productTypeID})
		if err != nil {
			adjPrices = map[int64]float64{}
		}
		bpProductCache = map[string]*models.BlueprintProduct{}
		bpMatsCache = map[string][]*models.BlueprintMaterial{}
		costIndexCache = map[string]float64{}
		bpForProductCache = map[int64]int64{}
		rxForProductCache = map[int64]int64{}
	}

	btc := &bomTreeContext{
		ctx:               ctx,
		repo:              repo,
		settings:          settings,
		prices:            prices,
		adjustedPrices:    adjPrices,
		blacklist:         blacklist,
		whitelist:         whitelist,
		assets:            assets,
		inputPriceType:    inputPriceType,
		buildAll:          buildAll,
		bpProductCache:    bpProductCache,
		bpMatsCache:       bpMatsCache,
		costIndexCache:    costIndexCache,
		bpForProductCache: bpForProductCache,
		rxForProductCache: rxForProductCache,
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
		node.MaterialCost = float64(qty) * buyPrice
		return node, nil
	}

	// Get materials for this blueprint — fall back to reaction if no manufacturing materials
	mats, err := btc.getBlueprintMaterials(blueprintTypeID, "manufacturing")
	activity := "manufacturing"
	if err != nil || len(mats) == 0 {
		mats, err = btc.getBlueprintMaterials(blueprintTypeID, "reaction")
		activity = "reaction"
		if err != nil || len(mats) == 0 {
			// Blueprint exists but has no materials — treat as zero-cost build leaf.
			node.Decision = "buy"
			return node, nil
		}
	}

	// Use depth-appropriate structure/rig for ME factor
	var structure, rig string
	switch {
	case depth == 0:
		structure = btc.settings.FinalStructure
		rig = btc.settings.FinalRig
	case depth == 1:
		structure = btc.settings.ComponentStructure
		rig = btc.settings.ComponentRig
	default:
		structure = btc.settings.ReactionStructure
		rig = btc.settings.ReactionRig
	}

	var meFactor float64
	if activity == "manufacturing" {
		meFactor = calculator.ComputeManufacturingME(me, structure, rig, "null")
	} else {
		meFactor = calculator.ComputeMEFactor(rig, "null")
	}

	productQtyPerRun := int64(1)
	cacheKey := fmt.Sprintf("%d:%s", blueprintTypeID, activity)
	if prod, ok := btc.bpProductCache[cacheKey]; ok {
		if prod != nil && prod.Quantity > 1 {
			productQtyPerRun = int64(prod.Quantity)
		}
	} else {
		prod, _ := btc.repo.GetBlueprintProductForActivity(btc.ctx, blueprintTypeID, activity)
		btc.bpProductCache[cacheKey] = prod
		if prod != nil && prod.Quantity > 1 {
			productQtyPerRun = int64(prod.Quantity)
		}
	}
	runs := (qty + productQtyPerRun - 1) / productQtyPerRun // ceil(qty / productQtyPerRun)

	var buildCost float64
	var eiv float64
	children := []*models.BOMNode{}

	for _, mat := range mats {
		batchQty := calculator.ComputeBatchQty(int(runs), mat.Quantity, meFactor)
		matBuyPrice := btc.getBuyPrice(mat.TypeID)
		matBuildCost := matBuyPrice // default: buy

		// Check if this material has a blueprint for sub-building — try manufacturing first, then reaction.
		// Both lookups go through the shared cache to avoid repeated DB calls across recursion levels.
		var subBpID int64
		if cached, ok := btc.bpForProductCache[mat.TypeID]; ok {
			subBpID = cached
		} else {
			subBpID, err = btc.repo.GetBlueprintForProduct(btc.ctx, mat.TypeID)
			if err != nil {
				subBpID = 0
			}
			btc.bpForProductCache[mat.TypeID] = subBpID
		}
		if subBpID == 0 {
			if cached, ok := btc.rxForProductCache[mat.TypeID]; ok {
				subBpID = cached
			} else {
				subBpID, _ = btc.repo.GetReactionBlueprintForProduct(btc.ctx, mat.TypeID)
				btc.rxForProductCache[mat.TypeID] = subBpID
			}
		}

		var childNode *models.BOMNode
		if subBpID != 0 && depth+1 < maxDepth && !btc.blacklist[mat.TypeID] {
			// Recurse to build sub-tree; sub-component blueprints are assumed researched to ME 10/20
			// reactions ignore the me param entirely via ComputeMEFactor
			childNode, err = buildBOMNode(btc, subBpID, mat.TypeID, mat.TypeName, int64(batchQty), 10, depth+1)
			if err != nil || childNode == nil {
				childAvail := btc.assets[mat.TypeID]
				childDelta := int64(batchQty) - childAvail
				if childDelta < 0 {
					childDelta = 0
				}
				childNode = &models.BOMNode{
					TypeID:        mat.TypeID,
					Name:          mat.TypeName,
					Quantity:      int64(batchQty),
					Available:     childAvail,
					Needed:        int64(batchQty),
					Delta:         childDelta,
					UnitBuyPrice:  matBuyPrice,
					Decision:      "buy",
					Children:      []*models.BOMNode{},
					MaterialCost:  float64(batchQty) * matBuyPrice,
				}
			}
			if childNode.Decision == "build" || childNode.Decision == "build_override" {
				matBuildCost = childNode.UnitBuildCost
			}
		} else {
			childAvail := btc.assets[mat.TypeID]
			childDelta := int64(batchQty) - childAvail
			if childDelta < 0 {
				childDelta = 0
			}
			decision := "buy"
			if btc.blacklist[mat.TypeID] {
				decision = "buy_override"
			}
			childNode = &models.BOMNode{
				TypeID:       mat.TypeID,
				Name:         mat.TypeName,
				Quantity:     int64(batchQty),
				Available:    childAvail,
				Needed:       int64(batchQty),
				Delta:        childDelta,
				UnitBuyPrice: matBuyPrice,
				Decision:     decision,
				Children:     []*models.BOMNode{},
				MaterialCost: float64(batchQty) * matBuyPrice,
			}
		}

		buildCost += matBuildCost * float64(batchQty)
		// EIV uses per-blueprint-run quantities (not ME-adjusted) for job cost calculation
		eiv += float64(mat.Quantity) * btc.getAdjustedPrice(mat.TypeID)
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

	// Compute this node's job cost using depth-appropriate structure settings
	var systemID *int64
	var facilityTax float64
	var structureName string
	switch {
	case depth == 0:
		systemID = btc.settings.FinalSystemID
		facilityTax = btc.settings.FinalFacilityTax
		structureName = btc.settings.FinalStructure
	case depth == 1:
		systemID = btc.settings.ComponentSystemID
		facilityTax = btc.settings.ComponentFacilityTax
		structureName = btc.settings.ComponentStructure
	default:
		systemID = btc.settings.ReactionSystemID
		facilityTax = btc.settings.ReactionFacilityTax
		structureName = btc.settings.ReactionStructure
	}
	if systemID != nil {
		costIdx := btc.getCostIndex(*systemID, activity)
		if costIdx > 0 {
			facilityTaxRate := facilityTax / 100.0
			if activity == "manufacturing" {
				structBonus := calculator.ManufacturingStructureCostBonus(structureName)
				node.JobCost = eiv * (costIdx*(1.0-structBonus) + facilityTaxRate + calculator.ManufacturingSccSurchargeRate) * float64(runs)
			} else {
				node.JobCost = eiv * (costIdx + facilityTaxRate + calculator.SccSurchargeRate) * float64(runs)
			}
		}
	}

	// Accumulate bottom-up cost fields.
	// MaterialCost: for buy/buy_override leaves this is market value; for build nodes it's the sum of children.
	if node.Decision == "buy" || node.Decision == "buy_override" {
		node.MaterialCost = float64(node.Quantity) * node.UnitBuyPrice
	} else {
		for _, child := range node.Children {
			node.MaterialCost += child.MaterialCost
		}
	}

	// TotalJobCost: this node's job fee plus all descendant job fees.
	node.TotalJobCost = node.JobCost
	for _, child := range node.Children {
		node.TotalJobCost += child.TotalJobCost
	}

	return node, nil
}
