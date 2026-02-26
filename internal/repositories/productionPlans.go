package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type ProductionPlans struct {
	db *sql.DB
}

func NewProductionPlans(db *sql.DB) *ProductionPlans {
	return &ProductionPlans{db: db}
}

func (r *ProductionPlans) Create(ctx context.Context, plan *models.ProductionPlan) (*models.ProductionPlan, error) {
	query := `
		INSERT INTO production_plans (user_id, product_type_id, name, notes,
			default_manufacturing_station_id, default_reaction_station_id,
			transport_fulfillment, transport_method, transport_profile_id,
			courier_rate_per_m3, courier_collateral_rate)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, user_id, product_type_id, name, notes,
			default_manufacturing_station_id, default_reaction_station_id,
			transport_fulfillment, transport_method, transport_profile_id,
			courier_rate_per_m3, courier_collateral_rate,
			created_at, updated_at
	`

	var result models.ProductionPlan
	err := r.db.QueryRowContext(ctx, query,
		plan.UserID,
		plan.ProductTypeID,
		plan.Name,
		plan.Notes,
		plan.DefaultManufacturingStationID,
		plan.DefaultReactionStationID,
		plan.TransportFulfillment,
		plan.TransportMethod,
		plan.TransportProfileID,
		plan.CourierRatePerM3,
		plan.CourierCollateralRate,
	).Scan(
		&result.ID,
		&result.UserID,
		&result.ProductTypeID,
		&result.Name,
		&result.Notes,
		&result.DefaultManufacturingStationID,
		&result.DefaultReactionStationID,
		&result.TransportFulfillment,
		&result.TransportMethod,
		&result.TransportProfileID,
		&result.CourierRatePerM3,
		&result.CourierCollateralRate,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create production plan")
	}

	return &result, nil
}

func (r *ProductionPlans) GetByUser(ctx context.Context, userID int64) ([]*models.ProductionPlan, error) {
	query := `
		SELECT p.id, p.user_id, p.product_type_id, p.name, p.notes,
		       p.default_manufacturing_station_id, p.default_reaction_station_id,
		       p.transport_fulfillment, p.transport_method, p.transport_profile_id,
		       p.courier_rate_per_m3, p.courier_collateral_rate,
		       p.created_at, p.updated_at,
		       COALESCE(ait.type_name, '') as product_name,
		       (SELECT COUNT(*) FROM production_plan_steps WHERE plan_id = p.id) as step_count
		FROM production_plans p
		LEFT JOIN asset_item_types ait ON ait.type_id = p.product_type_id
		WHERE p.user_id = $1
		ORDER BY p.updated_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query production plans")
	}
	defer rows.Close()

	plans := []*models.ProductionPlan{}
	for rows.Next() {
		var plan models.ProductionPlan
		var stepCount int
		err := rows.Scan(
			&plan.ID,
			&plan.UserID,
			&plan.ProductTypeID,
			&plan.Name,
			&plan.Notes,
			&plan.DefaultManufacturingStationID,
			&plan.DefaultReactionStationID,
			&plan.TransportFulfillment,
			&plan.TransportMethod,
			&plan.TransportProfileID,
			&plan.CourierRatePerM3,
			&plan.CourierCollateralRate,
			&plan.CreatedAt,
			&plan.UpdatedAt,
			&plan.ProductName,
			&stepCount,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan production plan")
		}
		plans = append(plans, &plan)
	}

	return plans, nil
}

func (r *ProductionPlans) GetByProductTypeAndUser(ctx context.Context, productTypeID, userID int64) ([]*models.ProductionPlan, error) {
	query := `
		SELECT p.id, p.user_id, p.product_type_id, p.name, p.notes,
		       p.default_manufacturing_station_id, p.default_reaction_station_id,
		       p.transport_fulfillment, p.transport_method, p.transport_profile_id,
		       p.courier_rate_per_m3, p.courier_collateral_rate,
		       p.created_at, p.updated_at,
		       COALESCE(ait.type_name, '') as product_name,
		       (SELECT COUNT(*) FROM production_plan_steps WHERE plan_id = p.id) as step_count
		FROM production_plans p
		LEFT JOIN asset_item_types ait ON ait.type_id = p.product_type_id
		WHERE p.user_id = $1 AND p.product_type_id = $2
		ORDER BY p.updated_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, productTypeID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query production plans by product type")
	}
	defer rows.Close()

	plans := []*models.ProductionPlan{}
	for rows.Next() {
		var plan models.ProductionPlan
		var stepCount int
		err := rows.Scan(
			&plan.ID,
			&plan.UserID,
			&plan.ProductTypeID,
			&plan.Name,
			&plan.Notes,
			&plan.DefaultManufacturingStationID,
			&plan.DefaultReactionStationID,
			&plan.TransportFulfillment,
			&plan.TransportMethod,
			&plan.TransportProfileID,
			&plan.CourierRatePerM3,
			&plan.CourierCollateralRate,
			&plan.CreatedAt,
			&plan.UpdatedAt,
			&plan.ProductName,
			&stepCount,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan production plan")
		}
		plans = append(plans, &plan)
	}

	return plans, nil
}

func (r *ProductionPlans) GetByID(ctx context.Context, id, userID int64) (*models.ProductionPlan, error) {
	// Get plan
	planQuery := `
		SELECT p.id, p.user_id, p.product_type_id, p.name, p.notes,
		       p.default_manufacturing_station_id, p.default_reaction_station_id,
		       p.transport_fulfillment, p.transport_method, p.transport_profile_id,
		       p.courier_rate_per_m3, p.courier_collateral_rate,
		       p.created_at, p.updated_at,
		       COALESCE(ait.type_name, '') as product_name
		FROM production_plans p
		LEFT JOIN asset_item_types ait ON ait.type_id = p.product_type_id
		WHERE p.id = $1 AND p.user_id = $2
	`

	var plan models.ProductionPlan
	err := r.db.QueryRowContext(ctx, planQuery, id, userID).Scan(
		&plan.ID,
		&plan.UserID,
		&plan.ProductTypeID,
		&plan.Name,
		&plan.Notes,
		&plan.DefaultManufacturingStationID,
		&plan.DefaultReactionStationID,
		&plan.TransportFulfillment,
		&plan.TransportMethod,
		&plan.TransportProfileID,
		&plan.CourierRatePerM3,
		&plan.CourierCollateralRate,
		&plan.CreatedAt,
		&plan.UpdatedAt,
		&plan.ProductName,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to query production plan")
	}

	// Get all steps with rigCategory enrichment and source name resolution
	stepsQuery := `
		SELECT s.id, s.plan_id, s.parent_step_id, s.product_type_id, s.blueprint_type_id,
		       s.activity, s.me_level, s.te_level, s.industry_skill, s.adv_industry_skill,
		       s.structure, s.rig, s.security, s.facility_tax, s.station_name,
		       s.source_location_id, s.source_container_id, s.source_division_number,
		       s.source_owner_type, s.source_owner_id,
		       s.user_station_id,
		       COALESCE(product.type_name, '') as product_name,
		       COALESCE(bp.type_name, '') as blueprint_name,
		       CASE
		           WHEN s.activity = 'reaction' THEN 'reaction'
		           WHEN sg.category_id = 6 THEN 'ship'
		           WHEN sg.category_id = 7 THEN 'equipment'
		           WHEN sg.category_id = 8 THEN 'ammo'
		           WHEN sg.category_id IN (18, 87) THEN 'drone'
		           ELSE 'component'
		       END AS rig_category,
		       COALESCE(src_char.name, src_corp.name, '') AS source_owner_name,
		       COALESCE(src_div.name, '') AS source_division_name,
		       COALESCE(src_cname.name, src_ccname.name, '') AS source_container_name,
		       s.output_owner_type, s.output_owner_id, s.output_division_number, s.output_container_id,
		       COALESCE(out_char.name, out_corp.name, '') AS output_owner_name,
		       COALESCE(out_div.name, '') AS output_division_name,
		       COALESCE(out_cname.name, out_ccname.name, '') AS output_container_name
		FROM production_plan_steps s
		LEFT JOIN asset_item_types product ON product.type_id = s.product_type_id
		LEFT JOIN asset_item_types bp ON bp.type_id = s.blueprint_type_id
		LEFT JOIN sde_groups sg ON sg.group_id = product.group_id
		LEFT JOIN characters src_char ON s.source_owner_type = 'character' AND src_char.id = s.source_owner_id
		LEFT JOIN player_corporations src_corp ON s.source_owner_type = 'corporation' AND src_corp.id = s.source_owner_id
		LEFT JOIN corporation_divisions src_div ON s.source_owner_type = 'corporation'
		    AND src_div.corporation_id = s.source_owner_id
		    AND src_div.division_number = s.source_division_number
		    AND src_div.division_type = 'hangar'
		    AND src_div.user_id = $2
		LEFT JOIN character_asset_location_names src_cname ON src_cname.item_id = s.source_container_id
		LEFT JOIN corporation_asset_location_names src_ccname ON src_ccname.item_id = s.source_container_id
		LEFT JOIN characters out_char ON s.output_owner_type = 'character' AND out_char.id = s.output_owner_id
		LEFT JOIN player_corporations out_corp ON s.output_owner_type = 'corporation' AND out_corp.id = s.output_owner_id
		LEFT JOIN corporation_divisions out_div ON s.output_owner_type = 'corporation'
		    AND out_div.corporation_id = s.output_owner_id
		    AND out_div.division_number = s.output_division_number
		    AND out_div.division_type = 'hangar'
		    AND out_div.user_id = $2
		LEFT JOIN character_asset_location_names out_cname ON out_cname.item_id = s.output_container_id
		LEFT JOIN corporation_asset_location_names out_ccname ON out_ccname.item_id = s.output_container_id
		WHERE s.plan_id = $1
		ORDER BY s.parent_step_id NULLS FIRST, s.id
	`

	rows, err := r.db.QueryContext(ctx, stepsQuery, id, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query production plan steps")
	}
	defer rows.Close()

	plan.Steps = []*models.ProductionPlanStep{}
	stationIDs := []int64{}
	stationIDSet := map[int64]bool{}
	for rows.Next() {
		var step models.ProductionPlanStep
		err := rows.Scan(
			&step.ID,
			&step.PlanID,
			&step.ParentStepID,
			&step.ProductTypeID,
			&step.BlueprintTypeID,
			&step.Activity,
			&step.MELevel,
			&step.TELevel,
			&step.IndustrySkill,
			&step.AdvIndustrySkill,
			&step.Structure,
			&step.Rig,
			&step.Security,
			&step.FacilityTax,
			&step.StationName,
			&step.SourceLocationID,
			&step.SourceContainerID,
			&step.SourceDivisionNumber,
			&step.SourceOwnerType,
			&step.SourceOwnerID,
			&step.UserStationID,
			&step.ProductName,
			&step.BlueprintName,
			&step.RigCategory,
			&step.SourceOwnerName,
			&step.SourceDivisionName,
			&step.SourceContainerName,
			&step.OutputOwnerType,
			&step.OutputOwnerID,
			&step.OutputDivisionNumber,
			&step.OutputContainerID,
			&step.OutputOwnerName,
			&step.OutputDivisionName,
			&step.OutputContainerName,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan production plan step")
		}
		if step.UserStationID != nil && !stationIDSet[*step.UserStationID] {
			stationIDs = append(stationIDs, *step.UserStationID)
			stationIDSet[*step.UserStationID] = true
		}
		plan.Steps = append(plan.Steps, &step)
	}

	// Enrich steps that reference a user station
	if len(stationIDs) > 0 {
		if err := r.enrichStepsWithStationData(ctx, plan.Steps, stationIDs); err != nil {
			return nil, err
		}
	}

	return &plan, nil
}

func (r *ProductionPlans) Update(ctx context.Context, id, userID int64, name string, notes *string, defaultManufacturingStationID *int64, defaultReactionStationID *int64, transportFulfillment *string, transportMethod *string, transportProfileID *int64, courierRatePerM3 float64, courierCollateralRate float64) error {
	query := `
		UPDATE production_plans
		SET name = $3, notes = $4,
		    default_manufacturing_station_id = $5,
		    default_reaction_station_id = $6,
		    transport_fulfillment = $7,
		    transport_method = $8,
		    transport_profile_id = $9,
		    courier_rate_per_m3 = $10,
		    courier_collateral_rate = $11,
		    updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, id, userID, name, notes, defaultManufacturingStationID, defaultReactionStationID, transportFulfillment, transportMethod, transportProfileID, courierRatePerM3, courierCollateralRate)
	if err != nil {
		return errors.Wrap(err, "failed to update production plan")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}
	if rowsAffected == 0 {
		return errors.New("production plan not found")
	}

	return nil
}

func (r *ProductionPlans) Delete(ctx context.Context, id, userID int64) error {
	query := `DELETE FROM production_plans WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return errors.Wrap(err, "failed to delete production plan")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}
	if rowsAffected == 0 {
		return errors.New("production plan not found")
	}

	return nil
}

func (r *ProductionPlans) CreateStep(ctx context.Context, step *models.ProductionPlanStep) (*models.ProductionPlanStep, error) {
	query := `
		INSERT INTO production_plan_steps
		(plan_id, parent_step_id, product_type_id, blueprint_type_id, activity,
		 me_level, te_level, industry_skill, adv_industry_skill,
		 structure, rig, security, facility_tax, station_name,
		 source_location_id, source_container_id, source_division_number,
		 source_owner_type, source_owner_id, user_station_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query,
		step.PlanID,
		step.ParentStepID,
		step.ProductTypeID,
		step.BlueprintTypeID,
		step.Activity,
		step.MELevel,
		step.TELevel,
		step.IndustrySkill,
		step.AdvIndustrySkill,
		step.Structure,
		step.Rig,
		step.Security,
		step.FacilityTax,
		step.StationName,
		step.SourceLocationID,
		step.SourceContainerID,
		step.SourceDivisionNumber,
		step.SourceOwnerType,
		step.SourceOwnerID,
		step.UserStationID,
	).Scan(&step.ID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create production plan step")
	}

	return step, nil
}

func (r *ProductionPlans) UpdateStep(ctx context.Context, stepID, planID, userID int64, step *models.ProductionPlanStep) error {
	query := `
		UPDATE production_plan_steps s
		SET me_level = $4, te_level = $5,
		    industry_skill = $6, adv_industry_skill = $7,
		    structure = $8, rig = $9, security = $10,
		    facility_tax = $11, station_name = $12,
		    source_location_id = $13, source_container_id = $14,
		    source_division_number = $15, source_owner_type = $16, source_owner_id = $17,
		    user_station_id = $18,
		    output_owner_type = $19, output_owner_id = $20,
		    output_division_number = $21, output_container_id = $22
		FROM production_plans p
		WHERE s.id = $1 AND s.plan_id = $2 AND p.id = s.plan_id AND p.user_id = $3
	`

	result, err := r.db.ExecContext(ctx, query,
		stepID, planID, userID,
		step.MELevel,
		step.TELevel,
		step.IndustrySkill,
		step.AdvIndustrySkill,
		step.Structure,
		step.Rig,
		step.Security,
		step.FacilityTax,
		step.StationName,
		step.SourceLocationID,
		step.SourceContainerID,
		step.SourceDivisionNumber,
		step.SourceOwnerType,
		step.SourceOwnerID,
		step.UserStationID,
		step.OutputOwnerType,
		step.OutputOwnerID,
		step.OutputDivisionNumber,
		step.OutputContainerID,
	)
	if err != nil {
		return errors.Wrap(err, "failed to update production plan step")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}
	if rowsAffected == 0 {
		return errors.New("production plan step not found")
	}

	return nil
}

func (r *ProductionPlans) DeleteStep(ctx context.Context, stepID, planID, userID int64) error {
	// Verify ownership via join, then delete (CASCADE will remove children)
	query := `
		DELETE FROM production_plan_steps s
		USING production_plans p
		WHERE s.id = $1 AND s.plan_id = $2 AND p.id = s.plan_id AND p.user_id = $3
	`

	result, err := r.db.ExecContext(ctx, query, stepID, planID, userID)
	if err != nil {
		return errors.Wrap(err, "failed to delete production plan step")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}
	if rowsAffected == 0 {
		return errors.New("production plan step not found")
	}

	return nil
}

func (r *ProductionPlans) BatchUpdateSteps(ctx context.Context, stepIDs []int64, planID, userID int64, step *models.ProductionPlanStep) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	query := `
		UPDATE production_plan_steps s
		SET me_level = $4, te_level = $5,
		    industry_skill = $6, adv_industry_skill = $7,
		    structure = $8, rig = $9, security = $10,
		    facility_tax = $11, station_name = $12,
		    source_location_id = $13, source_container_id = $14,
		    source_division_number = $15, source_owner_type = $16, source_owner_id = $17,
		    user_station_id = $18,
		    output_owner_type = $19, output_owner_id = $20,
		    output_division_number = $21, output_container_id = $22
		FROM production_plans p
		WHERE s.id = ANY($1) AND s.plan_id = $2 AND p.id = s.plan_id AND p.user_id = $3
	`

	result, err := tx.ExecContext(ctx, query,
		pq.Array(stepIDs), planID, userID,
		step.MELevel,
		step.TELevel,
		step.IndustrySkill,
		step.AdvIndustrySkill,
		step.Structure,
		step.Rig,
		step.Security,
		step.FacilityTax,
		step.StationName,
		step.SourceLocationID,
		step.SourceContainerID,
		step.SourceDivisionNumber,
		step.SourceOwnerType,
		step.SourceOwnerID,
		step.UserStationID,
		step.OutputOwnerType,
		step.OutputOwnerID,
		step.OutputDivisionNumber,
		step.OutputContainerID,
	)
	if err != nil {
		return 0, errors.Wrap(err, "failed to batch update production plan steps")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "failed to get rows affected")
	}

	if err := tx.Commit(); err != nil {
		return 0, errors.Wrap(err, "failed to commit transaction")
	}

	return rowsAffected, nil
}

// GetStepMaterials returns the blueprint materials for a step, with producibility info
// and whether each material already has a production step in this plan.
func (r *ProductionPlans) GetStepMaterials(ctx context.Context, stepID, planID, userID int64) ([]*models.PlanMaterial, error) {
	query := `
		SELECT
			bm.type_id,
			COALESCE(ait.type_name, '') as type_name,
			bm.quantity,
			COALESCE(ait.volume, 0) as volume,
			CASE WHEN prod.blueprint_type_id IS NOT NULL THEN true ELSE false END as has_blueprint,
			prod.blueprint_type_id,
			prod.activity as blueprint_activity,
			CASE WHEN existing.id IS NOT NULL THEN true ELSE false END as is_produced
		FROM production_plan_steps s
		INNER JOIN production_plans p ON p.id = s.plan_id
		INNER JOIN sde_blueprint_materials bm ON bm.blueprint_type_id = s.blueprint_type_id AND bm.activity = s.activity
		LEFT JOIN asset_item_types ait ON ait.type_id = bm.type_id
		LEFT JOIN LATERAL (
			SELECT bp.blueprint_type_id, bp.activity
			FROM sde_blueprint_products bp
			WHERE bp.type_id = bm.type_id
			  AND bp.activity IN ('manufacturing', 'reaction')
			ORDER BY CASE bp.activity WHEN 'manufacturing' THEN 1 WHEN 'reaction' THEN 2 END
			LIMIT 1
		) prod ON true
		LEFT JOIN production_plan_steps existing ON existing.plan_id = s.plan_id
			AND existing.parent_step_id = s.id
			AND existing.product_type_id = bm.type_id
		WHERE s.id = $1 AND s.plan_id = $2 AND p.user_id = $3
		ORDER BY ait.type_name
	`

	rows, err := r.db.QueryContext(ctx, query, stepID, planID, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query step materials")
	}
	defer rows.Close()

	materials := []*models.PlanMaterial{}
	for rows.Next() {
		var mat models.PlanMaterial
		err := rows.Scan(
			&mat.TypeID,
			&mat.TypeName,
			&mat.Quantity,
			&mat.Volume,
			&mat.HasBlueprint,
			&mat.BlueprintTypeID,
			&mat.Activity,
			&mat.IsProduced,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan step material")
		}
		materials = append(materials, &mat)
	}

	return materials, nil
}

// GetContainersAtStation returns containers (character and corporation) at a given station.
func (r *ProductionPlans) GetContainersAtStation(ctx context.Context, userID, stationID int64) ([]*models.StationContainer, error) {
	query := `
		SELECT ca.item_id, COALESCE(loc.name, ait.type_name), 'character', ca.character_id, NULL::int
		FROM character_assets ca
		JOIN asset_item_types ait ON ait.type_id = ca.type_id
		LEFT JOIN character_asset_location_names loc ON loc.item_id = ca.item_id
		WHERE ca.user_id = $1 AND ca.location_id = $2
		  AND ca.is_singleton = true AND ait.type_name LIKE '%Container'

		UNION ALL

		SELECT ca.item_id, COALESCE(loc.name, ait.type_name), 'corporation', ca.corporation_id,
		       SUBSTRING(ca.location_flag, 8, 1)::int
		FROM corporation_assets ca
		JOIN asset_item_types ait ON ait.type_id = ca.type_id
		LEFT JOIN corporation_asset_location_names loc ON loc.item_id = ca.item_id
		JOIN corporation_assets office
		  ON office.item_id = ca.location_id
		  AND office.location_flag = 'OfficeFolder'
		  AND office.corporation_id = ca.corporation_id
		  AND office.user_id = ca.user_id
		WHERE ca.user_id = $1 AND office.location_id = $2
		  AND ca.is_singleton = true AND ait.type_name LIKE '%Container'
		  AND ca.location_flag LIKE 'CorpSAG%'
	`

	rows, err := r.db.QueryContext(ctx, query, userID, stationID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query containers at station")
	}
	defer rows.Close()

	containers := []*models.StationContainer{}
	for rows.Next() {
		var c models.StationContainer
		err := rows.Scan(&c.ID, &c.Name, &c.OwnerType, &c.OwnerID, &c.DivisionNumber)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan station container")
		}
		containers = append(containers, &c)
	}

	return containers, nil
}

// enrichStepsWithStationData overrides step fields with data from referenced user stations.
// When a step has user_station_id set, its structure, facilityTax, stationName, security,
// and rig are derived from the station rather than the step's own fields.
func (r *ProductionPlans) enrichStepsWithStationData(ctx context.Context, steps []*models.ProductionPlanStep, stationIDs []int64) error {
	// Fetch station data
	stationQuery := `
		SELECT us.id, us.structure, us.facility_tax,
		       COALESCE(st.name, '') as station_name,
		       CASE
		           WHEN ss.security >= 0.45 THEN 'high'
		           WHEN ss.security > 0.0 THEN 'low'
		           ELSE 'null'
		       END AS security
		FROM user_stations us
		JOIN stations st ON st.station_id = us.station_id
		JOIN solar_systems ss ON ss.solar_system_id = st.solar_system_id
		WHERE us.id = ANY($1)
	`

	type stationData struct {
		ID          int64
		Structure   string
		FacilityTax float64
		StationName string
		Security    string
	}

	rows, err := r.db.QueryContext(ctx, stationQuery, pq.Array(stationIDs))
	if err != nil {
		return errors.Wrap(err, "failed to query user stations for enrichment")
	}
	defer rows.Close()

	stationMap := map[int64]*stationData{}
	for rows.Next() {
		var sd stationData
		err := rows.Scan(&sd.ID, &sd.Structure, &sd.FacilityTax, &sd.StationName, &sd.Security)
		if err != nil {
			return errors.Wrap(err, "failed to scan station data")
		}
		stationMap[sd.ID] = &sd
	}

	// Fetch rigs for these stations
	rigQuery := `
		SELECT user_station_id, category, tier
		FROM user_station_rigs
		WHERE user_station_id = ANY($1)
	`

	type rigData struct {
		StationID int64
		Category  string
		Tier      string
	}

	rigRows, err := r.db.QueryContext(ctx, rigQuery, pq.Array(stationIDs))
	if err != nil {
		return errors.Wrap(err, "failed to query user station rigs for enrichment")
	}
	defer rigRows.Close()

	// Map station ID → category → tier
	rigMap := map[int64]map[string]string{}
	for rigRows.Next() {
		var rd rigData
		err := rigRows.Scan(&rd.StationID, &rd.Category, &rd.Tier)
		if err != nil {
			return errors.Wrap(err, "failed to scan rig data")
		}
		if rigMap[rd.StationID] == nil {
			rigMap[rd.StationID] = map[string]string{}
		}
		rigMap[rd.StationID][rd.Category] = rd.Tier
	}

	// Override step fields from station data
	for _, step := range steps {
		if step.UserStationID == nil {
			continue
		}
		sd, ok := stationMap[*step.UserStationID]
		if !ok {
			continue
		}

		step.Structure = sd.Structure
		step.FacilityTax = sd.FacilityTax
		step.StationName = &sd.StationName
		step.Security = sd.Security

		// Derive rig from station's rigs + step's rigCategory
		step.Rig = "none"
		if rigs, ok := rigMap[*step.UserStationID]; ok {
			if tier, ok := rigs[step.RigCategory]; ok {
				step.Rig = tier
			}
		}
	}

	return nil
}
