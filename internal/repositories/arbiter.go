package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type ArbiterRepository struct {
	db *sql.DB
}

func NewArbiterRepository(db *sql.DB) *ArbiterRepository {
	return &ArbiterRepository{db: db}
}

// GetArbiterSettings returns the arbiter settings for a user, or default settings if none exist.
func (r *ArbiterRepository) GetArbiterSettings(ctx context.Context, userID int64) (*models.ArbiterSettings, error) {
	query := `
SELECT
	user_id,
	reaction_structure, reaction_rig, reaction_system_id,
	invention_structure, invention_rig, invention_system_id,
	component_structure, component_rig, component_system_id,
	final_structure, final_rig, final_system_id,
	final_facility_tax, component_facility_tax, reaction_facility_tax, invention_facility_tax,
	use_whitelist, use_blacklist, decryptor_type_id, default_scope_id
FROM arbiter_settings
WHERE user_id = $1
`
	var s models.ArbiterSettings
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&s.UserID,
		&s.ReactionStructure, &s.ReactionRig, &s.ReactionSystemID,
		&s.InventionStructure, &s.InventionRig, &s.InventionSystemID,
		&s.ComponentStructure, &s.ComponentRig, &s.ComponentSystemID,
		&s.FinalStructure, &s.FinalRig, &s.FinalSystemID,
		&s.FinalFacilityTax, &s.ComponentFacilityTax, &s.ReactionFacilityTax, &s.InventionFacilityTax,
		&s.UseWhitelist, &s.UseBlacklist, &s.DecryptorTypeID, &s.DefaultScopeID,
	)
	if err == sql.ErrNoRows {
		// Return defaults matching DB column defaults
		return &models.ArbiterSettings{
			UserID:             userID,
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
		}, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to query arbiter settings")
	}
	return &s, nil
}

// UpsertArbiterSettings saves arbiter settings for a user.
func (r *ArbiterRepository) UpsertArbiterSettings(ctx context.Context, settings *models.ArbiterSettings) error {
	query := `
INSERT INTO arbiter_settings (
	user_id,
	reaction_structure, reaction_rig, reaction_system_id,
	invention_structure, invention_rig, invention_system_id,
	component_structure, component_rig, component_system_id,
	final_structure, final_rig, final_system_id,
	final_facility_tax, component_facility_tax, reaction_facility_tax, invention_facility_tax,
	use_whitelist, use_blacklist, decryptor_type_id, default_scope_id,
	updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, NOW())
ON CONFLICT (user_id) DO UPDATE SET
	reaction_structure     = EXCLUDED.reaction_structure,
	reaction_rig           = EXCLUDED.reaction_rig,
	reaction_system_id     = EXCLUDED.reaction_system_id,
	invention_structure    = EXCLUDED.invention_structure,
	invention_rig          = EXCLUDED.invention_rig,
	invention_system_id    = EXCLUDED.invention_system_id,
	component_structure    = EXCLUDED.component_structure,
	component_rig          = EXCLUDED.component_rig,
	component_system_id    = EXCLUDED.component_system_id,
	final_structure        = EXCLUDED.final_structure,
	final_rig              = EXCLUDED.final_rig,
	final_system_id        = EXCLUDED.final_system_id,
	final_facility_tax     = EXCLUDED.final_facility_tax,
	component_facility_tax = EXCLUDED.component_facility_tax,
	reaction_facility_tax  = EXCLUDED.reaction_facility_tax,
	invention_facility_tax = EXCLUDED.invention_facility_tax,
	use_whitelist          = EXCLUDED.use_whitelist,
	use_blacklist          = EXCLUDED.use_blacklist,
	decryptor_type_id      = EXCLUDED.decryptor_type_id,
	default_scope_id       = EXCLUDED.default_scope_id,
	updated_at             = NOW()
`
	_, err := r.db.ExecContext(ctx, query,
		settings.UserID,
		settings.ReactionStructure, settings.ReactionRig, settings.ReactionSystemID,
		settings.InventionStructure, settings.InventionRig, settings.InventionSystemID,
		settings.ComponentStructure, settings.ComponentRig, settings.ComponentSystemID,
		settings.FinalStructure, settings.FinalRig, settings.FinalSystemID,
		settings.FinalFacilityTax, settings.ComponentFacilityTax, settings.ReactionFacilityTax, settings.InventionFacilityTax,
		settings.UseWhitelist, settings.UseBlacklist, settings.DecryptorTypeID, settings.DefaultScopeID,
	)
	if err != nil {
		return errors.Wrap(err, "failed to upsert arbiter settings")
	}
	return nil
}

// GetDecryptors returns all decryptors from sde_decryptors.
func (r *ArbiterRepository) GetDecryptors(ctx context.Context) ([]*models.Decryptor, error) {
	query := `
SELECT
  type_id,
  name,
  probability_multiplier,
  me_modifier,
  te_modifier,
  run_modifier
FROM sde_decryptors
WHERE name IN (
  'Accelerant Decryptor',
  'Attainment Decryptor',
  'Augmentation Decryptor',
  'Optimized Attainment Decryptor',
  'Optimized Augmentation Decryptor',
  'Parity Decryptor',
  'Process Decryptor',
  'Symmetry Decryptor'
)
ORDER BY run_modifier DESC
`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query decryptors")
	}
	defer rows.Close()

	decryptors := []*models.Decryptor{}
	for rows.Next() {
		var d models.Decryptor
		if err := rows.Scan(&d.TypeID, &d.Name, &d.ProbabilityMultiplier, &d.MEModifier, &d.TEModifier, &d.RunModifier); err != nil {
			return nil, errors.Wrap(err, "failed to scan decryptor")
		}
		decryptors = append(decryptors, &d)
	}
	return decryptors, nil
}

// GetT2BlueprintsForScan returns all T2 ship and module blueprint type IDs with their product info.
// T2 = meta_group_id = 2, categories = Ship (6) and Module (7).
// Only returns items that have invention blueprint activities (confirming they are T2).
func (r *ArbiterRepository) GetT2BlueprintsForScan(ctx context.Context) ([]*models.T2BlueprintScanItem, error) {
	query := `
SELECT
	product_ait.type_id           AS product_type_id,
	product_ait.type_name         AS product_name,
	t2_bp.blueprint_type_id       AS blueprint_type_id,
	invent_bp.blueprint_type_id   AS t1_blueprint_type_id,
	COALESCE(invent_prod.probability, 0.0) AS base_invention_chance,
	COALESCE(t2_mfg_prod.quantity, 0) AS base_result_runs,
	CASE
		WHEN sc.name = 'Ship' THEN 'ship'
		ELSE 'module'
	END                           AS category
FROM sde_blueprint_products t2_mfg_prod
JOIN sde_blueprint_activities t2_mfg_act
	ON t2_mfg_act.blueprint_type_id = t2_mfg_prod.blueprint_type_id
	AND t2_mfg_act.activity = 'manufacturing'
JOIN asset_item_types product_ait
	ON product_ait.type_id = t2_mfg_prod.type_id
	AND product_ait.meta_group_id = 2
JOIN sde_groups sg
	ON sg.group_id = product_ait.group_id
JOIN sde_categories sc
	ON sc.category_id = sg.category_id
	AND sc.category_id IN (6, 7)
-- The T2 blueprint itself
JOIN sde_blueprints t2_bp
	ON t2_bp.blueprint_type_id = t2_mfg_prod.blueprint_type_id
-- The T1 blueprint used to invent the T2 blueprint (invention product = T2 blueprint)
JOIN sde_blueprint_products invent_prod
	ON invent_prod.type_id = t2_mfg_prod.blueprint_type_id
	AND invent_prod.activity = 'invention'
JOIN sde_blueprints invent_bp
	ON invent_bp.blueprint_type_id = invent_prod.blueprint_type_id
WHERE t2_mfg_prod.activity = 'manufacturing'
ORDER BY sc.name, product_ait.type_name
`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query T2 blueprints for scan")
	}
	defer rows.Close()

	items := []*models.T2BlueprintScanItem{}
	for rows.Next() {
		var item models.T2BlueprintScanItem
		if err := rows.Scan(
			&item.ProductTypeID,
			&item.ProductName,
			&item.BlueprintTypeID,
			&item.T1BlueprintTypeID,
			&item.BaseInventionChance,
			&item.BaseResultRuns,
			&item.Category,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan T2 blueprint scan item")
		}
		// Base ME for T2 BPCs produced by invention is always 2
		item.BaseResultME = 2
		items = append(items, &item)
	}
	return items, nil
}

// GetAllBlueprintsForScan returns all manufacturable items across all tech levels and categories.
func (r *ArbiterRepository) GetAllBlueprintsForScan(ctx context.Context) ([]*models.BlueprintScanItem, error) {
	query := `
SELECT
	product_ait.type_id,
	product_ait.type_name,
	mfg_bp.blueprint_type_id,
	invent_prod.blueprint_type_id AS t1_blueprint_type_id,
	COALESCE(invent_prod.probability, 0) AS base_invention_chance,
	COALESCE(mfg_prod.quantity, 1) AS base_result_runs,
	COALESCE(mg.name, 'T1') AS tech_level,
	sc.name AS category_name,
	sg.name AS group_name,
	CASE WHEN sc.name = 'Ship' THEN 'ship' ELSE lower(sc.name) END AS category
FROM sde_blueprint_products mfg_prod
JOIN sde_blueprint_activities mfg_act
	ON mfg_act.blueprint_type_id = mfg_prod.blueprint_type_id
	AND mfg_act.activity = 'manufacturing'
JOIN asset_item_types product_ait ON product_ait.type_id = mfg_prod.type_id
JOIN sde_groups sg ON sg.group_id = product_ait.group_id
JOIN sde_categories sc ON sc.category_id = sg.category_id
	AND sc.name IN ('Ship', 'Module', 'Charge', 'Drone', 'Implant', 'Booster',
	                'Deployable', 'Structure', 'Fighter')
JOIN sde_blueprints mfg_bp ON mfg_bp.blueprint_type_id = mfg_prod.blueprint_type_id
LEFT JOIN sde_meta_groups mg ON mg.meta_group_id = product_ait.meta_group_id
-- T2 invention chain (optional)
LEFT JOIN sde_blueprint_products invent_prod
	ON invent_prod.type_id = mfg_prod.blueprint_type_id
	AND invent_prod.activity = 'invention'
WHERE mfg_prod.activity = 'manufacturing'
ORDER BY sc.name, product_ait.type_name
`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query all blueprints for scan")
	}
	defer rows.Close()

	items := []*models.BlueprintScanItem{}
	for rows.Next() {
		var item models.BlueprintScanItem
		if err := rows.Scan(
			&item.ProductTypeID,
			&item.ProductName,
			&item.BlueprintTypeID,
			&item.T1BlueprintTypeID,
			&item.BaseInventionChance,
			&item.BaseResultRuns,
			&item.TechLevel,
			&item.CategoryName,
			&item.GroupName,
			&item.Category,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan blueprint scan item")
		}
		items = append(items, &item)
	}
	return items, nil
}

// GetBlueprintMaterialsForActivity returns materials for a blueprint activity.
func (r *ArbiterRepository) GetBlueprintMaterialsForActivity(ctx context.Context, blueprintTypeID int64, activity string) ([]*models.BlueprintMaterial, error) {
	query := `
SELECT
	bm.type_id,
	ait.type_name,
	bm.quantity
FROM sde_blueprint_materials bm
JOIN asset_item_types ait ON ait.type_id = bm.type_id
WHERE bm.blueprint_type_id = $1
  AND bm.activity = $2
ORDER BY ait.type_name
`
	rows, err := r.db.QueryContext(ctx, query, blueprintTypeID, activity)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query blueprint materials")
	}
	defer rows.Close()

	materials := []*models.BlueprintMaterial{}
	for rows.Next() {
		var m models.BlueprintMaterial
		if err := rows.Scan(&m.TypeID, &m.TypeName, &m.Quantity); err != nil {
			return nil, errors.Wrap(err, "failed to scan blueprint material")
		}
		materials = append(materials, &m)
	}
	return materials, nil
}

// GetBlueprintProductForActivity returns product info for a blueprint activity.
func (r *ArbiterRepository) GetBlueprintProductForActivity(ctx context.Context, blueprintTypeID int64, activity string) (*models.BlueprintProduct, error) {
	query := `
SELECT
	bp.type_id,
	ait.type_name,
	bp.quantity,
	COALESCE(bp.probability, 1.0)
FROM sde_blueprint_products bp
JOIN asset_item_types ait ON ait.type_id = bp.type_id
WHERE bp.blueprint_type_id = $1
  AND bp.activity = $2
LIMIT 1
`
	var p models.BlueprintProduct
	err := r.db.QueryRowContext(ctx, query, blueprintTypeID, activity).Scan(
		&p.TypeID, &p.TypeName, &p.Quantity, &p.Probability,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to query blueprint product")
	}
	return &p, nil
}

// GetBlueprintActivityTime returns the activity time (seconds) for a blueprint.
func (r *ArbiterRepository) GetBlueprintActivityTime(ctx context.Context, blueprintTypeID int64, activity string) (int64, error) {
	var t int64
	err := r.db.QueryRowContext(ctx,
		`SELECT time FROM sde_blueprint_activities WHERE blueprint_type_id = $1 AND activity = $2`,
		blueprintTypeID, activity,
	).Scan(&t)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, errors.Wrap(err, "failed to query blueprint activity time")
	}
	return t, nil
}

// GetInventionBlueprintForProduct returns the T1 blueprint type ID used to invent a T2 blueprint.
// Given T2 blueprint type_id, finds the T1 blueprint that produces it via invention.
func (r *ArbiterRepository) GetInventionBlueprintForProduct(ctx context.Context, t2BlueprintTypeID int64) (int64, error) {
	var t1BpID int64
	err := r.db.QueryRowContext(ctx,
		`SELECT blueprint_type_id FROM sde_blueprint_products WHERE type_id = $1 AND activity = 'invention' LIMIT 1`,
		t2BlueprintTypeID,
	).Scan(&t1BpID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, errors.Wrap(err, "failed to query invention blueprint for product")
	}
	return t1BpID, nil
}

// GetBlueprintForProduct returns the blueprint type ID that manufactures the given product type.
// Returns 0 if no blueprint found.
func (r *ArbiterRepository) GetBlueprintForProduct(ctx context.Context, productTypeID int64) (int64, error) {
	var bpID int64
	err := r.db.QueryRowContext(ctx,
		`SELECT blueprint_type_id FROM sde_blueprint_products WHERE type_id = $1 AND activity = 'manufacturing' LIMIT 1`,
		productTypeID,
	).Scan(&bpID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, errors.Wrap(err, "failed to query blueprint for product")
	}
	return bpID, nil
}

// GetReactionBlueprintForProduct returns the blueprint type ID that reacts/produces the given product type.
// Returns 0 if no reaction blueprint found.
func (r *ArbiterRepository) GetReactionBlueprintForProduct(ctx context.Context, productTypeID int64) (int64, error) {
	var bpID int64
	err := r.db.QueryRowContext(ctx,
		`SELECT blueprint_type_id FROM sde_blueprint_products WHERE type_id = $1 AND activity = 'reaction' LIMIT 1`,
		productTypeID,
	).Scan(&bpID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, errors.Wrap(err, "failed to query reaction blueprint for product")
	}
	return bpID, nil
}

// GetBestInventionCharacter returns the character with the highest invention chance for a blueprint.
// Considers Racial Encryption Methods skill and two Science skills required by the blueprint.
func (r *ArbiterRepository) GetBestInventionCharacter(ctx context.Context, userID int64, blueprintTypeID int64) (*models.InventionCharacter, error) {
	// Get required skills for this blueprint's invention activity
	skillQuery := `
SELECT type_id, level
FROM sde_blueprint_skills
WHERE blueprint_type_id = $1
  AND activity = 'invention'
ORDER BY type_id
LIMIT 3
`
	skillRows, err := r.db.QueryContext(ctx, skillQuery, blueprintTypeID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query invention skills")
	}
	defer skillRows.Close()

	type skillReq struct {
		typeID int64
		level  int
	}
	reqs := []skillReq{}
	for skillRows.Next() {
		var s skillReq
		if err := skillRows.Scan(&s.typeID, &s.level); err != nil {
			return nil, errors.Wrap(err, "failed to scan skill req")
		}
		reqs = append(reqs, s)
	}
	skillRows.Close()

	if len(reqs) == 0 {
		return nil, nil
	}

	// Collect skill type IDs
	skillTypeIDs := make([]int64, len(reqs))
	for i, s := range reqs {
		skillTypeIDs[i] = s.typeID
	}

	// Get characters and their skill levels for the user
	charQuery := `
SELECT
	c.id,
	c.name,
	cs.skill_id,
	COALESCE(cs.active_level, 0)
FROM characters c
LEFT JOIN character_skills cs
	ON cs.character_id = c.id
	AND cs.skill_id = ANY($2)
WHERE c.user_id = $1
ORDER BY c.id
`
	rows, err := r.db.QueryContext(ctx, charQuery, userID, pq.Array(skillTypeIDs))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query character invention skills")
	}
	defer rows.Close()

	type charSkills struct {
		id     int64
		name   string
		skills map[int64]int
	}
	chars := map[int64]*charSkills{}
	charOrder := []int64{}

	for rows.Next() {
		var charID int64
		var charName string
		var skillID sql.NullInt64
		var level int
		if err := rows.Scan(&charID, &charName, &skillID, &level); err != nil {
			return nil, errors.Wrap(err, "failed to scan character skill row")
		}
		if _, ok := chars[charID]; !ok {
			chars[charID] = &charSkills{id: charID, name: charName, skills: map[int64]int{}}
			charOrder = append(charOrder, charID)
		}
		if skillID.Valid {
			chars[charID].skills[skillID.Int64] = level
		}
	}

	if len(chars) == 0 {
		return nil, nil
	}

	encQuery := `
SELECT tda.type_id
FROM sde_type_dogma_attributes tda
JOIN asset_item_types ait ON ait.type_id = tda.type_id
JOIN sde_groups sg ON sg.group_id = ait.group_id
WHERE tda.type_id = ANY($1)
  AND sg.name = 'Encryption Methods'
`
	encRows, err := r.db.QueryContext(ctx, encQuery, pq.Array(skillTypeIDs))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query encryption skill group")
	}
	defer encRows.Close()

	encSkillIDs := map[int64]bool{}
	for encRows.Next() {
		var typeID int64
		if err := encRows.Scan(&typeID); err != nil {
			return nil, errors.Wrap(err, "failed to scan encryption skill type ID")
		}
		encSkillIDs[typeID] = true
	}
	encRows.Close()

	var bestChar *models.InventionCharacter
	var bestScore float64

	for _, charID := range charOrder {
		c := chars[charID]
		var encLevel int
		sciLevels := []int{}
		for _, sID := range skillTypeIDs {
			lvl := c.skills[sID]
			if encSkillIDs[sID] {
				encLevel = lvl
			} else {
				sciLevels = append(sciLevels, lvl)
			}
		}
		var sciSum int
		for _, l := range sciLevels {
			sciSum += l
		}
		score := float64(encLevel)*0.01 + float64(sciSum)*0.1
		if bestChar == nil || score > bestScore {
			bestScore = score
			sci1, sci2 := 0, 0
			if len(sciLevels) >= 1 {
				sci1 = sciLevels[0]
			}
			if len(sciLevels) >= 2 {
				sci2 = sciLevels[1]
			}
			bestChar = &models.InventionCharacter{
				CharacterID:          charID,
				Name:                 c.name,
				EncryptionSkillLevel: encLevel,
				Science1SkillLevel:   sci1,
				Science2SkillLevel:   sci2,
			}
		}
	}

	return bestChar, nil
}

// GetMarketPriceForType returns current Jita buy/sell price for a type.
func (r *ArbiterRepository) GetMarketPriceForType(ctx context.Context, typeID int64) (*models.MarketPrice, error) {
	query := `
SELECT type_id, region_id, buy_price, sell_price, daily_volume, order_book_volume, updated_at
FROM market_prices
WHERE type_id = $1 AND region_id = 10000002
`
	var p models.MarketPrice
	var updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query, typeID).Scan(
		&p.TypeID, &p.RegionID, &p.BuyPrice, &p.SellPrice, &p.DailyVolume, &p.OrderBookVolume, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to query market price")
	}
	p.UpdatedAt = updatedAt.Format(time.RFC3339)
	return &p, nil
}

// GetMarketPricesForTypes returns current Jita prices for multiple type IDs (bulk).
func (r *ArbiterRepository) GetMarketPricesForTypes(ctx context.Context, typeIDs []int64) (map[int64]*models.MarketPrice, error) {
	if len(typeIDs) == 0 {
		return map[int64]*models.MarketPrice{}, nil
	}
	query := `
SELECT type_id, region_id, buy_price, sell_price, daily_volume, order_book_volume, updated_at
FROM market_prices
WHERE type_id = ANY($1) AND region_id = 10000002
`
	rows, err := r.db.QueryContext(ctx, query, pq.Array(typeIDs))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query market prices for types")
	}
	defer rows.Close()

	prices := map[int64]*models.MarketPrice{}
	for rows.Next() {
		var p models.MarketPrice
		var updatedAt time.Time
		if err := rows.Scan(&p.TypeID, &p.RegionID, &p.BuyPrice, &p.SellPrice, &p.DailyVolume, &p.OrderBookVolume, &updatedAt); err != nil {
			return nil, errors.Wrap(err, "failed to scan market price")
		}
		p.UpdatedAt = updatedAt.Format(time.RFC3339)
		prices[p.TypeID] = &p
	}
	return prices, nil
}

// GetMarketPricesLastUpdated returns the most recent updated_at timestamp across all Jita market prices.
func (r *ArbiterRepository) GetMarketPricesLastUpdated(ctx context.Context) (*time.Time, error) {
	var t sql.NullTime
	err := r.db.QueryRowContext(ctx,
		`SELECT MAX(updated_at) FROM market_prices WHERE region_id = 10000002`,
	).Scan(&t)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to query market prices last updated")
	}
	if !t.Valid {
		return nil, nil
	}
	return &t.Time, nil
}

// GetCostIndexForSystem returns the manufacturing/invention cost index for a system.
func (r *ArbiterRepository) GetCostIndexForSystem(ctx context.Context, systemID int64, activity string) (float64, error) {
	var costIndex float64
	err := r.db.QueryRowContext(ctx,
		`SELECT cost_index FROM industry_cost_indices WHERE system_id = $1 AND activity = $2`,
		systemID, activity,
	).Scan(&costIndex)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, errors.Wrap(err, "failed to query cost index")
	}
	return costIndex, nil
}

// GetAdjustedPricesForTypes returns adjusted_price values for EIV calculation.
func (r *ArbiterRepository) GetAdjustedPricesForTypes(ctx context.Context, typeIDs []int64) (map[int64]float64, error) {
	if len(typeIDs) == 0 {
		return map[int64]float64{}, nil
	}
	query := `
SELECT type_id, adjusted_price
FROM adjusted_prices
WHERE type_id = ANY($1)
`
	rows, err := r.db.QueryContext(ctx, query, pq.Array(typeIDs))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query adjusted prices")
	}
	defer rows.Close()

	prices := map[int64]float64{}
	for rows.Next() {
		var typeID int64
		var price float64
		if err := rows.Scan(&typeID, &price); err != nil {
			return nil, errors.Wrap(err, "failed to scan adjusted price")
		}
		prices[typeID] = price
	}
	return prices, nil
}

// UpsertDecryptors populates sde_decryptors from sde_type_dogma_attributes.
func (r *ArbiterRepository) UpsertDecryptors(ctx context.Context) error {
	query := `
INSERT INTO sde_decryptors (type_id, name, probability_multiplier, me_modifier, te_modifier, run_modifier)
SELECT
	tda.type_id,
	ait.type_name,
	MAX(CASE WHEN tda.attribute_id = 1112 THEN tda.value END)           AS probability_multiplier,
	MAX(CASE WHEN tda.attribute_id = 1113 THEN tda.value END)::int      AS me_modifier,
	MAX(CASE WHEN tda.attribute_id = 1114 THEN tda.value END)::int      AS te_modifier,
	MAX(CASE WHEN tda.attribute_id = 1124 THEN tda.value END)::int      AS run_modifier
FROM sde_type_dogma_attributes tda
JOIN asset_item_types ait ON ait.type_id = tda.type_id
WHERE tda.attribute_id IN (1112, 1113, 1114, 1124)
GROUP BY tda.type_id, ait.type_name
HAVING COUNT(DISTINCT tda.attribute_id) = 4
ON CONFLICT (type_id) DO UPDATE SET
	name                   = EXCLUDED.name,
	probability_multiplier = EXCLUDED.probability_multiplier,
	me_modifier            = EXCLUDED.me_modifier,
	te_modifier            = EXCLUDED.te_modifier,
	run_modifier           = EXCLUDED.run_modifier
`
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return errors.Wrap(err, "failed to upsert decryptors")
	}
	return nil
}

// GetArbiterEnabled returns whether the Arbiter feature is enabled for a user.
func (r *ArbiterRepository) GetArbiterEnabled(ctx context.Context, userID int64) (bool, error) {
	var enabled bool
	err := r.db.QueryRowContext(ctx,
		`SELECT arbiter_enabled FROM users WHERE id = $1`,
		userID,
	).Scan(&enabled)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrap(err, "failed to query arbiter_enabled")
	}
	return enabled, nil
}

// --- Scopes ---

// GetScopes returns all arbiter scopes for a user.
func (r *ArbiterRepository) GetScopes(ctx context.Context, userID int64) ([]*models.ArbiterScope, error) {
	query := `
SELECT id, user_id, name, is_default
FROM arbiter_scopes
WHERE user_id = $1
ORDER BY name
`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query arbiter scopes")
	}
	defer rows.Close()

	scopes := []*models.ArbiterScope{}
	for rows.Next() {
		var s models.ArbiterScope
		if err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.IsDefault); err != nil {
			return nil, errors.Wrap(err, "failed to scan arbiter scope")
		}
		scopes = append(scopes, &s)
	}
	return scopes, nil
}

// GetScope returns a single arbiter scope by ID for a user.
func (r *ArbiterRepository) GetScope(ctx context.Context, scopeID, userID int64) (*models.ArbiterScope, error) {
	var s models.ArbiterScope
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, is_default FROM arbiter_scopes WHERE id = $1 AND user_id = $2`,
		scopeID, userID,
	).Scan(&s.ID, &s.UserID, &s.Name, &s.IsDefault)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to query arbiter scope")
	}
	return &s, nil
}

// CreateScope inserts a new arbiter scope and returns the new ID.
func (r *ArbiterRepository) CreateScope(ctx context.Context, scope *models.ArbiterScope) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO arbiter_scopes (user_id, name, is_default) VALUES ($1, $2, $3) RETURNING id`,
		scope.UserID, scope.Name, scope.IsDefault,
	).Scan(&id)
	if err != nil {
		return 0, errors.Wrap(err, "failed to create arbiter scope")
	}
	return id, nil
}

// UpdateScope updates name and is_default for an existing scope owned by the user.
func (r *ArbiterRepository) UpdateScope(ctx context.Context, scope *models.ArbiterScope) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE arbiter_scopes SET name = $1, is_default = $2 WHERE id = $3 AND user_id = $4`,
		scope.Name, scope.IsDefault, scope.ID, scope.UserID,
	)
	if err != nil {
		return errors.Wrap(err, "failed to update arbiter scope")
	}
	return nil
}

// DeleteScope deletes an arbiter scope owned by the user.
func (r *ArbiterRepository) DeleteScope(ctx context.Context, scopeID, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM arbiter_scopes WHERE id = $1 AND user_id = $2`,
		scopeID, userID,
	)
	if err != nil {
		return errors.Wrap(err, "failed to delete arbiter scope")
	}
	return nil
}

// GetScopeMembers returns all members of a scope, resolving names from characters/corporations.
func (r *ArbiterRepository) GetScopeMembers(ctx context.Context, scopeID int64) ([]*models.ArbiterScopeMember, error) {
	query := `
SELECT
	asm.id,
	asm.scope_id,
	asm.member_type,
	asm.member_id,
	CASE
		WHEN asm.member_type = 'character'    THEN COALESCE(c.name, '')
		WHEN asm.member_type = 'corporation'  THEN COALESCE(pc.name, '')
		ELSE ''
	END AS name
FROM arbiter_scope_members asm
LEFT JOIN characters c
	ON c.id = asm.member_id AND asm.member_type = 'character'
LEFT JOIN player_corporations pc
	ON pc.id = asm.member_id AND asm.member_type = 'corporation'
WHERE asm.scope_id = $1
ORDER BY asm.member_type, name
`
	rows, err := r.db.QueryContext(ctx, query, scopeID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query scope members")
	}
	defer rows.Close()

	members := []*models.ArbiterScopeMember{}
	for rows.Next() {
		var m models.ArbiterScopeMember
		if err := rows.Scan(&m.ID, &m.ScopeID, &m.MemberType, &m.MemberID, &m.Name); err != nil {
			return nil, errors.Wrap(err, "failed to scan scope member")
		}
		members = append(members, &m)
	}
	return members, nil
}

// AddScopeMember inserts a member into an arbiter scope.
func (r *ArbiterRepository) AddScopeMember(ctx context.Context, member *models.ArbiterScopeMember) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO arbiter_scope_members (scope_id, member_type, member_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
		member.ScopeID, member.MemberType, member.MemberID,
	)
	if err != nil {
		return errors.Wrap(err, "failed to add scope member")
	}
	return nil
}

// RemoveScopeMember removes a member from a scope.
func (r *ArbiterRepository) RemoveScopeMember(ctx context.Context, memberID, scopeID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM arbiter_scope_members WHERE id = $1 AND scope_id = $2`,
		memberID, scopeID,
	)
	if err != nil {
		return errors.Wrap(err, "failed to remove scope member")
	}
	return nil
}

// --- Tax Profile ---

// GetTaxProfile returns the tax profile for a user, or sensible defaults.
func (r *ArbiterRepository) GetTaxProfile(ctx context.Context, userID int64) (*models.ArbiterTaxProfile, error) {
	var p models.ArbiterTaxProfile
	err := r.db.QueryRowContext(ctx,
		`SELECT user_id, trader_character_id, sales_tax_rate, broker_fee_rate, structure_broker_fee, input_price_type, output_price_type
		 FROM arbiter_tax_profile WHERE user_id = $1`,
		userID,
	).Scan(&p.UserID, &p.TraderCharacterID, &p.SalesTaxRate, &p.BrokerFeeRate, &p.StructureBrokerFee, &p.InputPriceType, &p.OutputPriceType)
	if err == sql.ErrNoRows {
		return &models.ArbiterTaxProfile{
			UserID:             userID,
			SalesTaxRate:       0.036,
			BrokerFeeRate:      0.03,
			StructureBrokerFee: 0.02,
			InputPriceType:     "sell",
			OutputPriceType:    "buy",
		}, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to query tax profile")
	}
	return &p, nil
}

// UpsertTaxProfile saves or updates the tax profile for a user.
func (r *ArbiterRepository) UpsertTaxProfile(ctx context.Context, profile *models.ArbiterTaxProfile) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO arbiter_tax_profile (user_id, trader_character_id, sales_tax_rate, broker_fee_rate, structure_broker_fee, input_price_type, output_price_type, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
ON CONFLICT (user_id) DO UPDATE SET
	trader_character_id = EXCLUDED.trader_character_id,
	sales_tax_rate      = EXCLUDED.sales_tax_rate,
	broker_fee_rate     = EXCLUDED.broker_fee_rate,
	structure_broker_fee = EXCLUDED.structure_broker_fee,
	input_price_type    = EXCLUDED.input_price_type,
	output_price_type   = EXCLUDED.output_price_type,
	updated_at          = NOW()
`,
		profile.UserID, profile.TraderCharacterID, profile.SalesTaxRate, profile.BrokerFeeRate,
		profile.StructureBrokerFee, profile.InputPriceType, profile.OutputPriceType,
	)
	if err != nil {
		return errors.Wrap(err, "failed to upsert tax profile")
	}
	return nil
}

// --- Blacklist ---

// GetBlacklist returns all blacklisted items for a user.
func (r *ArbiterRepository) GetBlacklist(ctx context.Context, userID int64) ([]*models.ArbiterListItem, error) {
	query := `
SELECT bl.user_id, bl.type_id, COALESCE(ait.type_name, ''), bl.added_at
FROM arbiter_blacklist bl
LEFT JOIN asset_item_types ait ON ait.type_id = bl.type_id
WHERE bl.user_id = $1
ORDER BY ait.type_name
`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query blacklist")
	}
	defer rows.Close()

	items := []*models.ArbiterListItem{}
	for rows.Next() {
		var item models.ArbiterListItem
		if err := rows.Scan(&item.UserID, &item.TypeID, &item.Name, &item.AddedAt); err != nil {
			return nil, errors.Wrap(err, "failed to scan blacklist item")
		}
		items = append(items, &item)
	}
	return items, nil
}

// AddToBlacklist adds a type to the user's blacklist.
func (r *ArbiterRepository) AddToBlacklist(ctx context.Context, userID, typeID int64) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO arbiter_blacklist (user_id, type_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, typeID,
	)
	if err != nil {
		return errors.Wrap(err, "failed to add to blacklist")
	}
	return nil
}

// RemoveFromBlacklist removes a type from the user's blacklist.
func (r *ArbiterRepository) RemoveFromBlacklist(ctx context.Context, userID, typeID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM arbiter_blacklist WHERE user_id = $1 AND type_id = $2`,
		userID, typeID,
	)
	if err != nil {
		return errors.Wrap(err, "failed to remove from blacklist")
	}
	return nil
}

// --- Whitelist ---

// GetWhitelist returns all whitelisted items for a user.
func (r *ArbiterRepository) GetWhitelist(ctx context.Context, userID int64) ([]*models.ArbiterListItem, error) {
	query := `
SELECT wl.user_id, wl.type_id, COALESCE(ait.type_name, ''), wl.added_at
FROM arbiter_whitelist wl
LEFT JOIN asset_item_types ait ON ait.type_id = wl.type_id
WHERE wl.user_id = $1
ORDER BY ait.type_name
`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query whitelist")
	}
	defer rows.Close()

	items := []*models.ArbiterListItem{}
	for rows.Next() {
		var item models.ArbiterListItem
		if err := rows.Scan(&item.UserID, &item.TypeID, &item.Name, &item.AddedAt); err != nil {
			return nil, errors.Wrap(err, "failed to scan whitelist item")
		}
		items = append(items, &item)
	}
	return items, nil
}

// AddToWhitelist adds a type to the user's whitelist.
func (r *ArbiterRepository) AddToWhitelist(ctx context.Context, userID, typeID int64) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO arbiter_whitelist (user_id, type_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, typeID,
	)
	if err != nil {
		return errors.Wrap(err, "failed to add to whitelist")
	}
	return nil
}

// RemoveFromWhitelist removes a type from the user's whitelist.
func (r *ArbiterRepository) RemoveFromWhitelist(ctx context.Context, userID, typeID int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM arbiter_whitelist WHERE user_id = $1 AND type_id = $2`,
		userID, typeID,
	)
	if err != nil {
		return errors.Wrap(err, "failed to remove from whitelist")
	}
	return nil
}

// --- Assets for scope ---

// GetScopeAssets returns total quantity per type_id across all scope members.
func (r *ArbiterRepository) GetScopeAssets(ctx context.Context, scopeID, userID int64) (map[int64]int64, error) {
	// Verify scope belongs to user
	scope, err := r.GetScope(ctx, scopeID, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to verify scope ownership")
	}
	if scope == nil {
		return map[int64]int64{}, nil
	}

	query := `
SELECT type_id, SUM(quantity) AS total
FROM (
	SELECT ca.type_id, ca.quantity
	FROM character_assets ca
	JOIN arbiter_scope_members asm ON asm.member_id = ca.character_id
		AND asm.member_type = 'character'
	WHERE asm.scope_id = $1
	UNION ALL
	SELECT ca.type_id, ca.quantity
	FROM corporation_assets ca
	JOIN arbiter_scope_members asm ON asm.member_id = ca.corporation_id
		AND asm.member_type = 'corporation'
	WHERE asm.scope_id = $1
) combined
GROUP BY type_id
`
	rows, err := r.db.QueryContext(ctx, query, scopeID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query scope assets")
	}
	defer rows.Close()

	assets := map[int64]int64{}
	for rows.Next() {
		var typeID, total int64
		if err := rows.Scan(&typeID, &total); err != nil {
			return nil, errors.Wrap(err, "failed to scan scope asset row")
		}
		assets[typeID] = total
	}
	return assets, nil
}

// --- Demand / DOS ---

// GetDemandStats returns 30d avg daily volume and current order_book_volume for each type.
func (r *ArbiterRepository) GetDemandStats(ctx context.Context, typeIDs []int64) (map[int64]*models.DemandStats, error) {
	if len(typeIDs) == 0 {
		return map[int64]*models.DemandStats{}, nil
	}

	// 30-day avg daily volume from history
	histQuery := `
SELECT
	type_id,
	AVG(daily_volume) AS avg_daily_volume
FROM market_price_history
WHERE type_id = ANY($1)
  AND snapshot_date >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY type_id
`
	histRows, err := r.db.QueryContext(ctx, histQuery, pq.Array(typeIDs))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query price history for demand stats")
	}
	defer histRows.Close()

	result := map[int64]*models.DemandStats{}
	for histRows.Next() {
		var typeID int64
		var avg float64
		if err := histRows.Scan(&typeID, &avg); err != nil {
			return nil, errors.Wrap(err, "failed to scan demand stats row")
		}
		result[typeID] = &models.DemandStats{
			TypeID:       typeID,
			DemandPerDay: avg,
		}
	}
	histRows.Close()

	// Current order book volume
	obQuery := `
SELECT type_id, COALESCE(order_book_volume, 0)
FROM market_prices
WHERE type_id = ANY($1) AND region_id = 10000002
`
	obRows, err := r.db.QueryContext(ctx, obQuery, pq.Array(typeIDs))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query order book volume")
	}
	defer obRows.Close()

	for obRows.Next() {
		var typeID int64
		var obv int64
		if err := obRows.Scan(&typeID, &obv); err != nil {
			return nil, errors.Wrap(err, "failed to scan order book volume")
		}
		if s, ok := result[typeID]; ok {
			s.OrderBookVolume = obv
			if s.DemandPerDay > 0 {
				s.DaysOfSupply = float64(obv) / s.DemandPerDay
			}
		} else {
			result[typeID] = &models.DemandStats{
				TypeID:          typeID,
				OrderBookVolume: obv,
			}
		}
	}

	return result, nil
}

// --- Solar system search ---

// SearchSolarSystems searches for solar systems by name prefix.
func (r *ArbiterRepository) SearchSolarSystems(ctx context.Context, query string, limit int) ([]*models.SolarSystemSearchResult, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	q := `
SELECT
	solar_system_id,
	name,
	CASE
		WHEN security >= 0.45 THEN 'high'
		WHEN security > 0.0  THEN 'low'
		WHEN security <= 0.0 THEN 'null'
		ELSE 'null'
	END AS security_class,
	security
FROM solar_systems
WHERE name ILIKE $1
ORDER BY name
LIMIT $2
`
	rows, err := r.db.QueryContext(ctx, q, query+"%", limit)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search solar systems")
	}
	defer rows.Close()

	results := []*models.SolarSystemSearchResult{}
	for rows.Next() {
		var s models.SolarSystemSearchResult
		if err := rows.Scan(&s.SolarSystemID, &s.Name, &s.SecurityClass, &s.Security); err != nil {
			return nil, errors.Wrap(err, "failed to scan solar system search result")
		}
		results = append(results, &s)
	}
	return results, nil
}
