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
	reaction_structure, reaction_rig, reaction_security, reaction_system_id,
	invention_structure, invention_rig, invention_security, invention_system_id,
	component_structure, component_rig, component_security, component_system_id,
	final_structure, final_rig, final_security, final_system_id
FROM arbiter_settings
WHERE user_id = $1
`
	var s models.ArbiterSettings
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&s.UserID,
		&s.ReactionStructure, &s.ReactionRig, &s.ReactionSecurity, &s.ReactionSystemID,
		&s.InventionStructure, &s.InventionRig, &s.InventionSecurity, &s.InventionSystemID,
		&s.ComponentStructure, &s.ComponentRig, &s.ComponentSecurity, &s.ComponentSystemID,
		&s.FinalStructure, &s.FinalRig, &s.FinalSecurity, &s.FinalSystemID,
	)
	if err == sql.ErrNoRows {
		// Return defaults matching DB column defaults
		return &models.ArbiterSettings{
			UserID:             userID,
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
	reaction_structure, reaction_rig, reaction_security, reaction_system_id,
	invention_structure, invention_rig, invention_security, invention_system_id,
	component_structure, component_rig, component_security, component_system_id,
	final_structure, final_rig, final_security, final_system_id,
	updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, NOW())
ON CONFLICT (user_id) DO UPDATE SET
	reaction_structure  = EXCLUDED.reaction_structure,
	reaction_rig        = EXCLUDED.reaction_rig,
	reaction_security   = EXCLUDED.reaction_security,
	reaction_system_id  = EXCLUDED.reaction_system_id,
	invention_structure = EXCLUDED.invention_structure,
	invention_rig       = EXCLUDED.invention_rig,
	invention_security  = EXCLUDED.invention_security,
	invention_system_id = EXCLUDED.invention_system_id,
	component_structure = EXCLUDED.component_structure,
	component_rig       = EXCLUDED.component_rig,
	component_security  = EXCLUDED.component_security,
	component_system_id = EXCLUDED.component_system_id,
	final_structure     = EXCLUDED.final_structure,
	final_rig           = EXCLUDED.final_rig,
	final_security      = EXCLUDED.final_security,
	final_system_id     = EXCLUDED.final_system_id,
	updated_at          = NOW()
`
	_, err := r.db.ExecContext(ctx, query,
		settings.UserID,
		settings.ReactionStructure, settings.ReactionRig, settings.ReactionSecurity, settings.ReactionSystemID,
		settings.InventionStructure, settings.InventionRig, settings.InventionSecurity, settings.InventionSystemID,
		settings.ComponentStructure, settings.ComponentRig, settings.ComponentSecurity, settings.ComponentSystemID,
		settings.FinalStructure, settings.FinalRig, settings.FinalSecurity, settings.FinalSystemID,
	)
	if err != nil {
		return errors.Wrap(err, "failed to upsert arbiter settings")
	}
	return nil
}

// GetDecryptors returns all decryptors from sde_decryptors.
func (r *ArbiterRepository) GetDecryptors(ctx context.Context) ([]*models.Decryptor, error) {
	query := `
SELECT type_id, name, probability_multiplier, me_modifier, te_modifier, run_modifier
FROM sde_decryptors
ORDER BY name
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
	c.character_name,
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

	// Identify which skills are encryption vs science by convention:
	// encryption skills are in group 1116 (Encryption Methods), science skills are others.
	// Simple heuristic: the skill with the highest type_id in the required set is usually
	// one of the two science skills; the one matching known encryption type IDs is the encryption skill.
	// For robustness, we look up group membership for the required skills.
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

	// Score each character: encryption_skill * 0.01 + (sum of science skills) * 0.005
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
		score := float64(encLevel)*0.01 + float64(sciSum)*0.005
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
SELECT type_id, region_id, buy_price, sell_price, daily_volume, updated_at
FROM market_prices
WHERE type_id = $1 AND region_id = 10000002
`
	var p models.MarketPrice
	var updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query, typeID).Scan(
		&p.TypeID, &p.RegionID, &p.BuyPrice, &p.SellPrice, &p.DailyVolume, &updatedAt,
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
SELECT type_id, region_id, buy_price, sell_price, daily_volume, updated_at
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
		if err := rows.Scan(&p.TypeID, &p.RegionID, &p.BuyPrice, &p.SellPrice, &p.DailyVolume, &updatedAt); err != nil {
			return nil, errors.Wrap(err, "failed to scan market price")
		}
		p.UpdatedAt = updatedAt.Format(time.RFC3339)
		prices[p.TypeID] = &p
	}
	return prices, nil
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
FROM market_prices
WHERE type_id = ANY($1)
  AND adjusted_price IS NOT NULL
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
