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
		TechLevel:     "T2",
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
		ME:                    resultME,
		TE:                    decryptor.TEModifier,
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
	systemID := ac.settings.FinalSystemID

	// Derive security class from system if system is set
	security := "null"
	if systemID != nil {
		security = deriveSecurityClassFromSettings(ac)
	}

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
