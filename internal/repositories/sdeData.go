package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type SdeDataRepository struct {
	db *sql.DB
}

func NewSdeDataRepository(db *sql.DB) *SdeDataRepository {
	return &SdeDataRepository{db: db}
}

func (r *SdeDataRepository) GetMetadata(ctx context.Context, key string) (*models.SdeMetadata, error) {
	query := `SELECT key, value, updated_at FROM sde_metadata WHERE key = $1`

	var m models.SdeMetadata
	err := r.db.QueryRowContext(ctx, query, key).Scan(&m.Key, &m.Value, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get SDE metadata")
	}

	return &m, nil
}

func (r *SdeDataRepository) SetMetadata(ctx context.Context, key, value string) error {
	query := `
INSERT INTO sde_metadata (key, value, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
`
	_, err := r.db.ExecContext(ctx, query, key, value)
	if err != nil {
		return errors.Wrap(err, "failed to set SDE metadata")
	}
	return nil
}

func (r *SdeDataRepository) UpsertCategories(ctx context.Context, categories []models.SdeCategory) error {
	return batchUpsert(r.db, ctx,
		`INSERT INTO sde_categories (category_id, name, published, icon_id) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (category_id) DO UPDATE SET name = EXCLUDED.name, published = EXCLUDED.published, icon_id = EXCLUDED.icon_id`,
		categories,
		func(smt *sql.Stmt, c models.SdeCategory) error {
			_, err := smt.ExecContext(ctx, c.CategoryID, c.Name, c.Published, c.IconID)
			return err
		},
	)
}

func (r *SdeDataRepository) UpsertGroups(ctx context.Context, groups []models.SdeGroup) error {
	return batchUpsert(r.db, ctx,
		`INSERT INTO sde_groups (group_id, name, category_id, published, icon_id) VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (group_id) DO UPDATE SET name = EXCLUDED.name, category_id = EXCLUDED.category_id, published = EXCLUDED.published, icon_id = EXCLUDED.icon_id`,
		groups,
		func(smt *sql.Stmt, g models.SdeGroup) error {
			_, err := smt.ExecContext(ctx, g.GroupID, g.Name, g.CategoryID, g.Published, g.IconID)
			return err
		},
	)
}

func (r *SdeDataRepository) UpsertMetaGroups(ctx context.Context, metaGroups []models.SdeMetaGroup) error {
	return batchUpsert(r.db, ctx,
		`INSERT INTO sde_meta_groups (meta_group_id, name) VALUES ($1, $2)
		 ON CONFLICT (meta_group_id) DO UPDATE SET name = EXCLUDED.name`,
		metaGroups,
		func(smt *sql.Stmt, mg models.SdeMetaGroup) error {
			_, err := smt.ExecContext(ctx, mg.MetaGroupID, mg.Name)
			return err
		},
	)
}

func (r *SdeDataRepository) UpsertMarketGroups(ctx context.Context, marketGroups []models.SdeMarketGroup) error {
	return batchUpsert(r.db, ctx,
		`INSERT INTO sde_market_groups (market_group_id, parent_group_id, name, description, icon_id, has_types) VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (market_group_id) DO UPDATE SET parent_group_id = EXCLUDED.parent_group_id, name = EXCLUDED.name, description = EXCLUDED.description, icon_id = EXCLUDED.icon_id, has_types = EXCLUDED.has_types`,
		marketGroups,
		func(smt *sql.Stmt, mg models.SdeMarketGroup) error {
			_, err := smt.ExecContext(ctx, mg.MarketGroupID, mg.ParentGroupID, mg.Name, mg.Description, mg.IconID, mg.HasTypes)
			return err
		},
	)
}

func (r *SdeDataRepository) UpsertIcons(ctx context.Context, icons []models.SdeIcon) error {
	return batchUpsert(r.db, ctx,
		`INSERT INTO sde_icons (icon_id, description) VALUES ($1, $2)
		 ON CONFLICT (icon_id) DO UPDATE SET description = EXCLUDED.description`,
		icons,
		func(smt *sql.Stmt, i models.SdeIcon) error {
			_, err := smt.ExecContext(ctx, i.IconID, i.Description)
			return err
		},
	)
}

func (r *SdeDataRepository) UpsertGraphics(ctx context.Context, graphics []models.SdeGraphic) error {
	return batchUpsert(r.db, ctx,
		`INSERT INTO sde_graphics (graphic_id, description) VALUES ($1, $2)
		 ON CONFLICT (graphic_id) DO UPDATE SET description = EXCLUDED.description`,
		graphics,
		func(smt *sql.Stmt, g models.SdeGraphic) error {
			_, err := smt.ExecContext(ctx, g.GraphicID, g.Description)
			return err
		},
	)
}

func (r *SdeDataRepository) UpsertBlueprints(ctx context.Context, blueprints []models.SdeBlueprint, activities []models.SdeBlueprintActivity, materials []models.SdeBlueprintMaterial, products []models.SdeBlueprintProduct, skills []models.SdeBlueprintSkill) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin blueprint transaction")
	}
	defer tx.Rollback()

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_blueprints (blueprint_type_id, max_production_limit) VALUES ($1, $2)
		 ON CONFLICT (blueprint_type_id) DO UPDATE SET max_production_limit = EXCLUDED.max_production_limit`,
		blueprints,
		func(smt *sql.Stmt, b models.SdeBlueprint) error {
			_, err := smt.ExecContext(ctx, b.BlueprintTypeID, b.MaxProductionLimit)
			return err
		},
	); err != nil {
		return errors.Wrap(err, "failed to upsert blueprints")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_blueprint_activities (blueprint_type_id, activity, time) VALUES ($1, $2, $3)
		 ON CONFLICT (blueprint_type_id, activity) DO UPDATE SET time = EXCLUDED.time`,
		activities,
		func(smt *sql.Stmt, a models.SdeBlueprintActivity) error {
			_, err := smt.ExecContext(ctx, a.BlueprintTypeID, a.Activity, a.Time)
			return err
		},
	); err != nil {
		return errors.Wrap(err, "failed to upsert blueprint activities")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_blueprint_materials (blueprint_type_id, activity, type_id, quantity) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (blueprint_type_id, activity, type_id) DO UPDATE SET quantity = EXCLUDED.quantity`,
		materials,
		func(smt *sql.Stmt, m models.SdeBlueprintMaterial) error {
			_, err := smt.ExecContext(ctx, m.BlueprintTypeID, m.Activity, m.TypeID, m.Quantity)
			return err
		},
	); err != nil {
		return errors.Wrap(err, "failed to upsert blueprint materials")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_blueprint_products (blueprint_type_id, activity, type_id, quantity, probability) VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (blueprint_type_id, activity, type_id) DO UPDATE SET quantity = EXCLUDED.quantity, probability = EXCLUDED.probability`,
		products,
		func(smt *sql.Stmt, p models.SdeBlueprintProduct) error {
			_, err := smt.ExecContext(ctx, p.BlueprintTypeID, p.Activity, p.TypeID, p.Quantity, p.Probability)
			return err
		},
	); err != nil {
		return errors.Wrap(err, "failed to upsert blueprint products")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_blueprint_skills (blueprint_type_id, activity, type_id, level) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (blueprint_type_id, activity, type_id) DO UPDATE SET level = EXCLUDED.level`,
		skills,
		func(smt *sql.Stmt, s models.SdeBlueprintSkill) error {
			_, err := smt.ExecContext(ctx, s.BlueprintTypeID, s.Activity, s.TypeID, s.Level)
			return err
		},
	); err != nil {
		return errors.Wrap(err, "failed to upsert blueprint skills")
	}

	return tx.Commit()
}

func (r *SdeDataRepository) UpsertDogma(ctx context.Context, attrCats []models.SdeDogmaAttributeCategory, attrs []models.SdeDogmaAttribute, effects []models.SdeDogmaEffect, typeAttrs []models.SdeTypeDogmaAttribute, typeEffects []models.SdeTypeDogmaEffect) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin dogma transaction")
	}
	defer tx.Rollback()

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_dogma_attribute_categories (category_id, name, description) VALUES ($1, $2, $3)
		 ON CONFLICT (category_id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description`,
		attrCats,
		func(smt *sql.Stmt, c models.SdeDogmaAttributeCategory) error {
			_, err := smt.ExecContext(ctx, c.CategoryID, c.Name, c.Description)
			return err
		},
	); err != nil {
		return errors.Wrap(err, "failed to upsert dogma attribute categories")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_dogma_attributes (attribute_id, name, description, default_value, display_name, category_id, high_is_good, stackable, published, unit_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (attribute_id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description, default_value = EXCLUDED.default_value, display_name = EXCLUDED.display_name, category_id = EXCLUDED.category_id, high_is_good = EXCLUDED.high_is_good, stackable = EXCLUDED.stackable, published = EXCLUDED.published, unit_id = EXCLUDED.unit_id`,
		attrs,
		func(smt *sql.Stmt, a models.SdeDogmaAttribute) error {
			_, err := smt.ExecContext(ctx, a.AttributeID, a.Name, a.Description, a.DefaultValue, a.DisplayName, a.CategoryID, a.HighIsGood, a.Stackable, a.Published, a.UnitID)
			return err
		},
	); err != nil {
		return errors.Wrap(err, "failed to upsert dogma attributes")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_dogma_effects (effect_id, name, description, display_name, category_id) VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (effect_id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description, display_name = EXCLUDED.display_name, category_id = EXCLUDED.category_id`,
		effects,
		func(smt *sql.Stmt, e models.SdeDogmaEffect) error {
			_, err := smt.ExecContext(ctx, e.EffectID, e.Name, e.Description, e.DisplayName, e.CategoryID)
			return err
		},
	); err != nil {
		return errors.Wrap(err, "failed to upsert dogma effects")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_type_dogma_attributes (type_id, attribute_id, value) VALUES ($1, $2, $3)
		 ON CONFLICT (type_id, attribute_id) DO UPDATE SET value = EXCLUDED.value`,
		typeAttrs,
		func(smt *sql.Stmt, a models.SdeTypeDogmaAttribute) error {
			_, err := smt.ExecContext(ctx, a.TypeID, a.AttributeID, a.Value)
			return err
		},
	); err != nil {
		return errors.Wrap(err, "failed to upsert type dogma attributes")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_type_dogma_effects (type_id, effect_id, is_default) VALUES ($1, $2, $3)
		 ON CONFLICT (type_id, effect_id) DO UPDATE SET is_default = EXCLUDED.is_default`,
		typeEffects,
		func(smt *sql.Stmt, e models.SdeTypeDogmaEffect) error {
			_, err := smt.ExecContext(ctx, e.TypeID, e.EffectID, e.IsDefault)
			return err
		},
	); err != nil {
		return errors.Wrap(err, "failed to upsert type dogma effects")
	}

	return tx.Commit()
}

func (r *SdeDataRepository) UpsertNpcData(ctx context.Context, factions []models.SdeFaction, corps []models.SdeNpcCorporation, divs []models.SdeNpcCorporationDivision, agents []models.SdeAgent, agentsInSpace []models.SdeAgentInSpace, races []models.SdeRace, bloodlines []models.SdeBloodline, ancestries []models.SdeAncestry) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin NPC data transaction")
	}
	defer tx.Rollback()

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_factions (faction_id, name, description, corporation_id, icon_id) VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (faction_id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description, corporation_id = EXCLUDED.corporation_id, icon_id = EXCLUDED.icon_id`,
		factions,
		func(smt *sql.Stmt, f models.SdeFaction) error {
			_, err := smt.ExecContext(ctx, f.FactionID, f.Name, f.Description, f.CorporationID, f.IconID)
			return err
		},
	); err != nil {
		return errors.Wrap(err, "failed to upsert factions")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_npc_corporations (corporation_id, name, faction_id, icon_id) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (corporation_id) DO UPDATE SET name = EXCLUDED.name, faction_id = EXCLUDED.faction_id, icon_id = EXCLUDED.icon_id`,
		corps,
		func(smt *sql.Stmt, c models.SdeNpcCorporation) error {
			_, err := smt.ExecContext(ctx, c.CorporationID, c.Name, c.FactionID, c.IconID)
			return err
		},
	); err != nil {
		return errors.Wrap(err, "failed to upsert NPC corporations")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_npc_corporation_divisions (corporation_id, division_id, name) VALUES ($1, $2, $3)
		 ON CONFLICT (corporation_id, division_id) DO UPDATE SET name = EXCLUDED.name`,
		divs,
		func(smt *sql.Stmt, d models.SdeNpcCorporationDivision) error {
			_, err := smt.ExecContext(ctx, d.CorporationID, d.DivisionID, d.Name)
			return err
		},
	); err != nil {
		return errors.Wrap(err, "failed to upsert NPC corporation divisions")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_agents (agent_id, name, corporation_id, division_id, level) VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (agent_id) DO UPDATE SET name = EXCLUDED.name, corporation_id = EXCLUDED.corporation_id, division_id = EXCLUDED.division_id, level = EXCLUDED.level`,
		agents,
		func(smt *sql.Stmt, a models.SdeAgent) error {
			_, err := smt.ExecContext(ctx, a.AgentID, a.Name, a.CorporationID, a.DivisionID, a.Level)
			return err
		},
	); err != nil {
		return errors.Wrap(err, "failed to upsert agents")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_agents_in_space (agent_id, solar_system_id) VALUES ($1, $2)
		 ON CONFLICT (agent_id) DO UPDATE SET solar_system_id = EXCLUDED.solar_system_id`,
		agentsInSpace,
		func(smt *sql.Stmt, a models.SdeAgentInSpace) error {
			_, err := smt.ExecContext(ctx, a.AgentID, a.SolarSystemID)
			return err
		},
	); err != nil {
		return errors.Wrap(err, "failed to upsert agents in space")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_races (race_id, name, description, icon_id) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (race_id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description, icon_id = EXCLUDED.icon_id`,
		races,
		func(smt *sql.Stmt, r models.SdeRace) error {
			_, err := smt.ExecContext(ctx, r.RaceID, r.Name, r.Description, r.IconID)
			return err
		},
	); err != nil {
		return errors.Wrap(err, "failed to upsert races")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_bloodlines (bloodline_id, name, race_id, description, icon_id) VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (bloodline_id) DO UPDATE SET name = EXCLUDED.name, race_id = EXCLUDED.race_id, description = EXCLUDED.description, icon_id = EXCLUDED.icon_id`,
		bloodlines,
		func(smt *sql.Stmt, b models.SdeBloodline) error {
			_, err := smt.ExecContext(ctx, b.BloodlineID, b.Name, b.RaceID, b.Description, b.IconID)
			return err
		},
	); err != nil {
		return errors.Wrap(err, "failed to upsert bloodlines")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_ancestries (ancestry_id, name, bloodline_id, description, icon_id) VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (ancestry_id) DO UPDATE SET name = EXCLUDED.name, bloodline_id = EXCLUDED.bloodline_id, description = EXCLUDED.description, icon_id = EXCLUDED.icon_id`,
		ancestries,
		func(smt *sql.Stmt, a models.SdeAncestry) error {
			_, err := smt.ExecContext(ctx, a.AncestryID, a.Name, a.BloodlineID, a.Description, a.IconID)
			return err
		},
	); err != nil {
		return errors.Wrap(err, "failed to upsert ancestries")
	}

	return tx.Commit()
}

func (r *SdeDataRepository) UpsertIndustryData(ctx context.Context, schematics []models.SdePlanetSchematic, schematicTypes []models.SdePlanetSchematicType, towerResources []models.SdeControlTowerResource) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin industry data transaction")
	}
	defer tx.Rollback()

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_planet_schematics (schematic_id, name, cycle_time) VALUES ($1, $2, $3)
		 ON CONFLICT (schematic_id) DO UPDATE SET name = EXCLUDED.name, cycle_time = EXCLUDED.cycle_time`,
		schematics,
		func(smt *sql.Stmt, s models.SdePlanetSchematic) error {
			_, err := smt.ExecContext(ctx, s.SchematicID, s.Name, s.CycleTime)
			return err
		},
	); err != nil {
		return errors.Wrap(err, "failed to upsert planet schematics")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_planet_schematic_types (schematic_id, type_id, quantity, is_input) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (schematic_id, type_id) DO UPDATE SET quantity = EXCLUDED.quantity, is_input = EXCLUDED.is_input`,
		schematicTypes,
		func(smt *sql.Stmt, st models.SdePlanetSchematicType) error {
			_, err := smt.ExecContext(ctx, st.SchematicID, st.TypeID, st.Quantity, st.IsInput)
			return err
		},
	); err != nil {
		return errors.Wrap(err, "failed to upsert planet schematic types")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_control_tower_resources (control_tower_type_id, resource_type_id, purpose, quantity, min_security, faction_id) VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (control_tower_type_id, resource_type_id) DO UPDATE SET purpose = EXCLUDED.purpose, quantity = EXCLUDED.quantity, min_security = EXCLUDED.min_security, faction_id = EXCLUDED.faction_id`,
		towerResources,
		func(smt *sql.Stmt, r models.SdeControlTowerResource) error {
			_, err := smt.ExecContext(ctx, r.ControlTowerTypeID, r.ResourceTypeID, r.Purpose, r.Quantity, r.MinSecurity, r.FactionID)
			return err
		},
	); err != nil {
		return errors.Wrap(err, "failed to upsert control tower resources")
	}

	return tx.Commit()
}

func (r *SdeDataRepository) UpsertMiscData(ctx context.Context, skins []models.SdeSkin, skinLicenses []models.SdeSkinLicense, skinMaterials []models.SdeSkinMaterial, certificates []models.SdeCertificate, landmarks []models.SdeLandmark, stationOps []models.SdeStationOperation, stationSvcs []models.SdeStationService, contrabandTypes []models.SdeContrabandType, researchAgents []models.SdeResearchAgent, charAttrs []models.SdeCharacterAttribute, corpActivities []models.SdeCorporationActivity, tournamentRuleSets []models.SdeTournamentRuleSet) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin misc data transaction")
	}
	defer tx.Rollback()

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_skins (skin_id, type_id, material_id) VALUES ($1, $2, $3)
		 ON CONFLICT (skin_id) DO UPDATE SET type_id = EXCLUDED.type_id, material_id = EXCLUDED.material_id`,
		skins,
		func(smt *sql.Stmt, s models.SdeSkin) error {
			_, err := smt.ExecContext(ctx, s.SkinID, s.TypeID, s.MaterialID)
			return err
		}); err != nil {
		return errors.Wrap(err, "failed to upsert skins")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_skin_licenses (license_type_id, skin_id, duration) VALUES ($1, $2, $3)
		 ON CONFLICT (license_type_id) DO UPDATE SET skin_id = EXCLUDED.skin_id, duration = EXCLUDED.duration`,
		skinLicenses,
		func(smt *sql.Stmt, l models.SdeSkinLicense) error {
			_, err := smt.ExecContext(ctx, l.LicenseTypeID, l.SkinID, l.Duration)
			return err
		}); err != nil {
		return errors.Wrap(err, "failed to upsert skin licenses")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_skin_materials (skin_material_id, name) VALUES ($1, $2)
		 ON CONFLICT (skin_material_id) DO UPDATE SET name = EXCLUDED.name`,
		skinMaterials,
		func(smt *sql.Stmt, m models.SdeSkinMaterial) error {
			_, err := smt.ExecContext(ctx, m.SkinMaterialID, m.Name)
			return err
		}); err != nil {
		return errors.Wrap(err, "failed to upsert skin materials")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_certificates (certificate_id, name, description, group_id) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (certificate_id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description, group_id = EXCLUDED.group_id`,
		certificates,
		func(smt *sql.Stmt, c models.SdeCertificate) error {
			_, err := smt.ExecContext(ctx, c.CertificateID, c.Name, c.Description, c.GroupID)
			return err
		}); err != nil {
		return errors.Wrap(err, "failed to upsert certificates")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_landmarks (landmark_id, name, description) VALUES ($1, $2, $3)
		 ON CONFLICT (landmark_id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description`,
		landmarks,
		func(smt *sql.Stmt, l models.SdeLandmark) error {
			_, err := smt.ExecContext(ctx, l.LandmarkID, l.Name, l.Description)
			return err
		}); err != nil {
		return errors.Wrap(err, "failed to upsert landmarks")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_station_operations (operation_id, name, description) VALUES ($1, $2, $3)
		 ON CONFLICT (operation_id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description`,
		stationOps,
		func(smt *sql.Stmt, o models.SdeStationOperation) error {
			_, err := smt.ExecContext(ctx, o.OperationID, o.Name, o.Description)
			return err
		}); err != nil {
		return errors.Wrap(err, "failed to upsert station operations")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_station_services (service_id, name, description) VALUES ($1, $2, $3)
		 ON CONFLICT (service_id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description`,
		stationSvcs,
		func(smt *sql.Stmt, s models.SdeStationService) error {
			_, err := smt.ExecContext(ctx, s.ServiceID, s.Name, s.Description)
			return err
		}); err != nil {
		return errors.Wrap(err, "failed to upsert station services")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_contraband_types (faction_id, type_id, standing_loss, fine_by_value) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (faction_id, type_id) DO UPDATE SET standing_loss = EXCLUDED.standing_loss, fine_by_value = EXCLUDED.fine_by_value`,
		contrabandTypes,
		func(smt *sql.Stmt, ct models.SdeContrabandType) error {
			_, err := smt.ExecContext(ctx, ct.FactionID, ct.TypeID, ct.StandingLoss, ct.FineByValue)
			return err
		}); err != nil {
		return errors.Wrap(err, "failed to upsert contraband types")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_research_agents (agent_id, type_id) VALUES ($1, $2)
		 ON CONFLICT (agent_id, type_id) DO NOTHING`,
		researchAgents,
		func(smt *sql.Stmt, ra models.SdeResearchAgent) error {
			_, err := smt.ExecContext(ctx, ra.AgentID, ra.TypeID)
			return err
		}); err != nil {
		return errors.Wrap(err, "failed to upsert research agents")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_character_attributes (attribute_id, name, description, icon_id) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (attribute_id) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description, icon_id = EXCLUDED.icon_id`,
		charAttrs,
		func(smt *sql.Stmt, a models.SdeCharacterAttribute) error {
			_, err := smt.ExecContext(ctx, a.AttributeID, a.Name, a.Description, a.IconID)
			return err
		}); err != nil {
		return errors.Wrap(err, "failed to upsert character attributes")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_corporation_activities (activity_id, name) VALUES ($1, $2)
		 ON CONFLICT (activity_id) DO UPDATE SET name = EXCLUDED.name`,
		corpActivities,
		func(smt *sql.Stmt, a models.SdeCorporationActivity) error {
			_, err := smt.ExecContext(ctx, a.ActivityID, a.Name)
			return err
		}); err != nil {
		return errors.Wrap(err, "failed to upsert corporation activities")
	}

	if err := bulkUpsertTx(ctx, tx,
		`INSERT INTO sde_tournament_rule_sets (rule_set_id, data) VALUES ($1, $2)
		 ON CONFLICT (rule_set_id) DO UPDATE SET data = EXCLUDED.data`,
		tournamentRuleSets,
		func(smt *sql.Stmt, rs models.SdeTournamentRuleSet) error {
			_, err := smt.ExecContext(ctx, rs.RuleSetID, rs.Data)
			return err
		}); err != nil {
		return errors.Wrap(err, "failed to upsert tournament rule sets")
	}

	return tx.Commit()
}

// ReactionRow represents a reaction with its product info and group
type ReactionRow struct {
	BlueprintTypeID int64
	ProductTypeID   int64
	ProductName     string
	GroupName       string
	ProductQuantity int
	Time            int
	ProductVolume   float64
}

// ReactionMaterialRow represents an input material for a reaction
type ReactionMaterialRow struct {
	BlueprintTypeID int64
	TypeID          int64
	TypeName        string
	Quantity        int
	Volume          float64
}

// GetAllReactions returns all reactions with product info and group names
func (r *SdeDataRepository) GetAllReactions(ctx context.Context) ([]*ReactionRow, error) {
	query := `
SELECT
	ba.blueprint_type_id,
	bp.type_id AS product_type_id,
	ait.type_name AS product_name,
	g.name AS group_name,
	bp.quantity AS product_quantity,
	ba.time,
	COALESCE(ait.packaged_volume, ait.volume, 0) AS product_volume
FROM sde_blueprint_activities ba
JOIN sde_blueprint_products bp ON bp.blueprint_type_id = ba.blueprint_type_id AND bp.activity = ba.activity
JOIN asset_item_types ait ON ait.type_id = bp.type_id
JOIN sde_groups g ON g.group_id = ait.group_id
WHERE ba.activity = 'reaction'
  AND ba.time >= 3600
ORDER BY g.name, ait.type_name
`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query reactions")
	}
	defer rows.Close()

	results := []*ReactionRow{}
	for rows.Next() {
		var row ReactionRow
		err := rows.Scan(
			&row.BlueprintTypeID,
			&row.ProductTypeID,
			&row.ProductName,
			&row.GroupName,
			&row.ProductQuantity,
			&row.Time,
			&row.ProductVolume,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan reaction row")
		}
		results = append(results, &row)
	}

	return results, nil
}

// GetAllReactionMaterials returns input materials for all reactions
func (r *SdeDataRepository) GetAllReactionMaterials(ctx context.Context) ([]*ReactionMaterialRow, error) {
	query := `
SELECT
	bm.blueprint_type_id,
	bm.type_id,
	ait.type_name,
	bm.quantity,
	COALESCE(ait.packaged_volume, ait.volume, 0) AS volume
FROM sde_blueprint_materials bm
JOIN asset_item_types ait ON ait.type_id = bm.type_id
WHERE bm.activity = 'reaction'
ORDER BY bm.blueprint_type_id, ait.type_name
`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query reaction materials")
	}
	defer rows.Close()

	results := []*ReactionMaterialRow{}
	for rows.Next() {
		var row ReactionMaterialRow
		err := rows.Scan(
			&row.BlueprintTypeID,
			&row.TypeID,
			&row.TypeName,
			&row.Quantity,
			&row.Volume,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan reaction material row")
		}
		results = append(results, &row)
	}

	return results, nil
}

// GetReactionSystems returns systems with reaction cost indices
func (r *SdeDataRepository) GetReactionSystems(ctx context.Context) ([]*models.ReactionSystem, error) {
	query := `
SELECT
	i.system_id,
	s.name,
	s.security,
	i.cost_index
FROM industry_cost_indices i
JOIN solar_systems s ON s.solar_system_id = i.system_id
WHERE i.activity = 'reaction'
ORDER BY s.name
`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query reaction systems")
	}
	defer rows.Close()

	results := []*models.ReactionSystem{}
	for rows.Next() {
		var sys models.ReactionSystem
		err := rows.Scan(
			&sys.SystemID,
			&sys.Name,
			&sys.SecurityStatus,
			&sys.CostIndex,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan reaction system row")
		}
		results = append(results, &sys)
	}

	return results, nil
}

// ManufacturingBlueprintRow represents a manufacturing blueprint with its product info
type ManufacturingBlueprintRow struct {
	BlueprintTypeID int64
	ProductTypeID   int64
	ProductName     string
	GroupName       string
	ProductQuantity int
	Time            int
	ProductVolume   float64
	MaxProdLimit    int
}

// ManufacturingMaterialRow represents an input material for a manufacturing blueprint
type ManufacturingMaterialRow struct {
	BlueprintTypeID int64
	TypeID          int64
	TypeName        string
	Quantity        int
	Volume          float64
}

// BlueprintSearchRow represents a blueprint search result
type BlueprintSearchRow struct {
	BlueprintTypeID int64
	BlueprintName   string
	ProductTypeID   int64
	ProductName     string
	Activity        string
}

// GetManufacturingBlueprint returns a single manufacturing blueprint with product info
func (r *SdeDataRepository) GetManufacturingBlueprint(ctx context.Context, blueprintTypeID int64) (*ManufacturingBlueprintRow, error) {
	query := `
SELECT
	ba.blueprint_type_id,
	bp.type_id AS product_type_id,
	ait.type_name AS product_name,
	g.name AS group_name,
	bp.quantity AS product_quantity,
	ba.time,
	COALESCE(ait.packaged_volume, ait.volume, 0) AS product_volume,
	COALESCE(sb.max_production_limit, 0)
FROM sde_blueprint_activities ba
JOIN sde_blueprint_products bp ON bp.blueprint_type_id = ba.blueprint_type_id AND bp.activity = ba.activity
JOIN asset_item_types ait ON ait.type_id = bp.type_id
JOIN sde_groups g ON g.group_id = ait.group_id
LEFT JOIN sde_blueprints sb ON sb.blueprint_type_id = ba.blueprint_type_id
WHERE ba.activity = 'manufacturing'
  AND ba.blueprint_type_id = $1
`

	var row ManufacturingBlueprintRow
	err := r.db.QueryRowContext(ctx, query, blueprintTypeID).Scan(
		&row.BlueprintTypeID,
		&row.ProductTypeID,
		&row.ProductName,
		&row.GroupName,
		&row.ProductQuantity,
		&row.Time,
		&row.ProductVolume,
		&row.MaxProdLimit,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to query manufacturing blueprint")
	}

	return &row, nil
}

// GetManufacturingMaterials returns input materials for a manufacturing blueprint
func (r *SdeDataRepository) GetManufacturingMaterials(ctx context.Context, blueprintTypeID int64) ([]*ManufacturingMaterialRow, error) {
	query := `
SELECT
	bm.blueprint_type_id,
	bm.type_id,
	ait.type_name,
	bm.quantity,
	COALESCE(ait.packaged_volume, ait.volume, 0) AS volume
FROM sde_blueprint_materials bm
JOIN asset_item_types ait ON ait.type_id = bm.type_id
WHERE bm.activity = 'manufacturing'
  AND bm.blueprint_type_id = $1
ORDER BY ait.type_name
`

	rows, err := r.db.QueryContext(ctx, query, blueprintTypeID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query manufacturing materials")
	}
	defer rows.Close()

	results := []*ManufacturingMaterialRow{}
	for rows.Next() {
		var row ManufacturingMaterialRow
		err := rows.Scan(
			&row.BlueprintTypeID,
			&row.TypeID,
			&row.TypeName,
			&row.Quantity,
			&row.Volume,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan manufacturing material row")
		}
		results = append(results, &row)
	}

	return results, nil
}

// SearchBlueprints searches blueprints by product name (ILIKE) for a given activity
func (r *SdeDataRepository) SearchBlueprints(ctx context.Context, query string, activity string, limit int) ([]*BlueprintSearchRow, error) {
	sqlQuery := `
SELECT DISTINCT
	ba.blueprint_type_id,
	bpait.type_name AS blueprint_name,
	bp.type_id AS product_type_id,
	prodait.type_name AS product_name,
	ba.activity
FROM sde_blueprint_activities ba
JOIN sde_blueprint_products bp ON bp.blueprint_type_id = ba.blueprint_type_id AND bp.activity = ba.activity
JOIN asset_item_types prodait ON prodait.type_id = bp.type_id
JOIN asset_item_types bpait ON bpait.type_id = ba.blueprint_type_id
WHERE ba.activity = $1
  AND prodait.type_name ILIKE '%' || $2 || '%'
ORDER BY prodait.type_name
LIMIT $3
`

	rows, err := r.db.QueryContext(ctx, sqlQuery, activity, query, limit)
	if err != nil {
		return nil, errors.Wrap(err, "failed to search blueprints")
	}
	defer rows.Close()

	results := []*BlueprintSearchRow{}
	for rows.Next() {
		var row BlueprintSearchRow
		err := rows.Scan(
			&row.BlueprintTypeID,
			&row.BlueprintName,
			&row.ProductTypeID,
			&row.ProductName,
			&row.Activity,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan blueprint search row")
		}
		results = append(results, &row)
	}

	return results, nil
}

// GetManufacturingSystems returns systems with manufacturing cost indices
func (r *SdeDataRepository) GetManufacturingSystems(ctx context.Context) ([]*models.ReactionSystem, error) {
	query := `
SELECT
	i.system_id,
	s.name,
	s.security,
	i.cost_index
FROM industry_cost_indices i
JOIN solar_systems s ON s.solar_system_id = i.system_id
WHERE i.activity = 'manufacturing'
ORDER BY s.name
`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query manufacturing systems")
	}
	defer rows.Close()

	results := []*models.ReactionSystem{}
	for rows.Next() {
		var sys models.ReactionSystem
		err := rows.Scan(
			&sys.SystemID,
			&sys.Name,
			&sys.SecurityStatus,
			&sys.CostIndex,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan manufacturing system row")
		}
		results = append(results, &sys)
	}

	return results, nil
}

// batchUpsert executes upsert queries for a batch of items in a single transaction
func batchUpsert[T any](db *sql.DB, ctx context.Context, upsertQuery string, items []T, execFn func(*sql.Stmt, T) error) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if err := bulkUpsertTx(ctx, tx, upsertQuery, items, execFn); err != nil {
		return err
	}

	return tx.Commit()
}

// bulkUpsertTx prepares and executes bulk upserts within an existing transaction
func bulkUpsertTx[T any](ctx context.Context, tx *sql.Tx, query string, items []T, execFn func(*sql.Stmt, T) error) error {
	if len(items) == 0 {
		return nil
	}

	smt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return errors.Wrap(err, "failed to prepare statement")
	}

	for _, item := range items {
		if err := execFn(smt, item); err != nil {
			return err
		}
	}

	return nil
}

// GetMetadataLastUpdateTime returns the last update time for SDE metadata
func (r *SdeDataRepository) GetAllSchematics(ctx context.Context) ([]*models.SdePlanetSchematic, error) {
	rows, err := r.db.QueryContext(ctx, `select schematic_id, name, cycle_time from sde_planet_schematics`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query planet schematics")
	}
	defer rows.Close()

	schematics := []*models.SdePlanetSchematic{}
	for rows.Next() {
		var s models.SdePlanetSchematic
		if err := rows.Scan(&s.SchematicID, &s.Name, &s.CycleTime); err != nil {
			return nil, errors.Wrap(err, "failed to scan planet schematic")
		}
		schematics = append(schematics, &s)
	}
	return schematics, nil
}

func (r *SdeDataRepository) GetAllSchematicTypes(ctx context.Context) ([]*models.SdePlanetSchematicType, error) {
	rows, err := r.db.QueryContext(ctx, `select schematic_id, type_id, quantity, is_input from sde_planet_schematic_types`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query planet schematic types")
	}
	defer rows.Close()

	types := []*models.SdePlanetSchematicType{}
	for rows.Next() {
		var t models.SdePlanetSchematicType
		if err := rows.Scan(&t.SchematicID, &t.TypeID, &t.Quantity, &t.IsInput); err != nil {
			return nil, errors.Wrap(err, "failed to scan planet schematic type")
		}
		types = append(types, &t)
	}
	return types, nil
}

func (r *SdeDataRepository) GetMetadataLastUpdateTime(ctx context.Context) (*time.Time, error) {
	query := `SELECT MAX(updated_at) FROM sde_metadata`

	var lastUpdate *time.Time
	err := r.db.QueryRowContext(ctx, query).Scan(&lastUpdate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to query last SDE metadata update time")
	}

	return lastUpdate, nil
}
