package updaters

import (
	"context"
	"os"

	"github.com/annymsMthd/industry-tool/internal/client"
	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type SdeClientInterface interface {
	GetChecksum(ctx context.Context) (string, error)
	DownloadSDE(ctx context.Context) (string, error)
	ParseSDE(zipPath string) (*client.SdeData, error)
}

type SdeDataRepository interface {
	GetMetadata(ctx context.Context, key string) (*models.SdeMetadata, error)
	SetMetadata(ctx context.Context, key, value string) error
	UpsertCategories(ctx context.Context, categories []models.SdeCategory) error
	UpsertGroups(ctx context.Context, groups []models.SdeGroup) error
	UpsertMetaGroups(ctx context.Context, metaGroups []models.SdeMetaGroup) error
	UpsertMarketGroups(ctx context.Context, marketGroups []models.SdeMarketGroup) error
	UpsertIcons(ctx context.Context, icons []models.SdeIcon) error
	UpsertGraphics(ctx context.Context, graphics []models.SdeGraphic) error
	UpsertBlueprints(ctx context.Context, blueprints []models.SdeBlueprint, activities []models.SdeBlueprintActivity, materials []models.SdeBlueprintMaterial, products []models.SdeBlueprintProduct, skills []models.SdeBlueprintSkill) error
	UpsertDogma(ctx context.Context, attrCats []models.SdeDogmaAttributeCategory, attrs []models.SdeDogmaAttribute, effects []models.SdeDogmaEffect, typeAttrs []models.SdeTypeDogmaAttribute, typeEffects []models.SdeTypeDogmaEffect) error
	UpsertNpcData(ctx context.Context, factions []models.SdeFaction, corps []models.SdeNpcCorporation, divs []models.SdeNpcCorporationDivision, agents []models.SdeAgent, agentsInSpace []models.SdeAgentInSpace, races []models.SdeRace, bloodlines []models.SdeBloodline, ancestries []models.SdeAncestry) error
	UpsertIndustryData(ctx context.Context, schematics []models.SdePlanetSchematic, schematicTypes []models.SdePlanetSchematicType, towerResources []models.SdeControlTowerResource) error
	UpsertMiscData(ctx context.Context, skins []models.SdeSkin, skinLicenses []models.SdeSkinLicense, skinMaterials []models.SdeSkinMaterial, certificates []models.SdeCertificate, landmarks []models.SdeLandmark, stationOps []models.SdeStationOperation, stationSvcs []models.SdeStationService, contrabandTypes []models.SdeContrabandType, researchAgents []models.SdeResearchAgent, charAttrs []models.SdeCharacterAttribute, corpActivities []models.SdeCorporationActivity, tournamentRuleSets []models.SdeTournamentRuleSet) error
}

type SdeItemTypeRepository interface {
	UpsertItemTypes(ctx context.Context, itemTypes []models.EveInventoryType) error
}

type SdeRegionRepository interface {
	Upsert(ctx context.Context, regions []models.Region) error
}

type SdeConstellationRepository interface {
	Upsert(ctx context.Context, constellations []models.Constellation) error
}

type SdeSolarSystemRepository interface {
	Upsert(ctx context.Context, systems []models.SolarSystem) error
}

type SdeStationRepository interface {
	Upsert(ctx context.Context, stations []models.Station) error
}

type Sde struct {
	client                  SdeClientInterface
	sdeDataRepo             SdeDataRepository
	itemTypeRepository      SdeItemTypeRepository
	regionRepository        SdeRegionRepository
	constellationRepository SdeConstellationRepository
	solarSystemRepository   SdeSolarSystemRepository
	stationRepository       SdeStationRepository
}

func NewSde(
	client SdeClientInterface,
	sdeDataRepo SdeDataRepository,
	itemTypeRepository SdeItemTypeRepository,
	regionRepository SdeRegionRepository,
	constellationRepository SdeConstellationRepository,
	solarSystemRepository SdeSolarSystemRepository,
	stationRepository SdeStationRepository,
) *Sde {
	return &Sde{
		client:                  client,
		sdeDataRepo:             sdeDataRepo,
		itemTypeRepository:      itemTypeRepository,
		regionRepository:        regionRepository,
		constellationRepository: constellationRepository,
		solarSystemRepository:   solarSystemRepository,
		stationRepository:       stationRepository,
	}
}

func (u *Sde) Update(ctx context.Context) error {
	// Step 1: Get checksum from CCP
	checksum, err := u.client.GetChecksum(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get SDE checksum")
	}

	// Step 2: Compare with stored checksum
	existing, err := u.sdeDataRepo.GetMetadata(ctx, "checksum")
	if err != nil {
		return errors.Wrap(err, "failed to get stored SDE checksum")
	}

	if existing != nil && existing.Value == checksum {
		log.Info("SDE already up to date", "checksum", checksum)
		return nil
	}

	log.Info("SDE update available", "new_checksum", checksum)

	// Step 3: Download SDE ZIP
	zipPath, err := u.client.DownloadSDE(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to download SDE")
	}
	defer os.Remove(zipPath)

	log.Info("SDE downloaded", "path", zipPath)

	// Step 4: Parse all YAML files
	data, err := u.client.ParseSDE(zipPath)
	if err != nil {
		return errors.Wrap(err, "failed to parse SDE")
	}

	log.Info("SDE parsed",
		"types", len(data.Types),
		"blueprints", len(data.Blueprints),
		"regions", len(data.Regions),
	)

	// Step 5: Populate existing tables
	if err := u.regionRepository.Upsert(ctx, data.Regions); err != nil {
		return errors.Wrap(err, "failed to upsert regions")
	}

	if err := u.constellationRepository.Upsert(ctx, data.Constellations); err != nil {
		return errors.Wrap(err, "failed to upsert constellations")
	}

	if err := u.solarSystemRepository.Upsert(ctx, data.SolarSystems); err != nil {
		return errors.Wrap(err, "failed to upsert solar systems")
	}

	if err := u.stationRepository.Upsert(ctx, data.Stations); err != nil {
		return errors.Wrap(err, "failed to upsert stations")
	}

	if err := u.itemTypeRepository.UpsertItemTypes(ctx, data.Types); err != nil {
		return errors.Wrap(err, "failed to upsert item types")
	}

	// Step 6: Populate SDE-specific tables
	if err := u.sdeDataRepo.UpsertCategories(ctx, data.Categories); err != nil {
		return errors.Wrap(err, "failed to upsert categories")
	}

	if err := u.sdeDataRepo.UpsertGroups(ctx, data.Groups); err != nil {
		return errors.Wrap(err, "failed to upsert groups")
	}

	if err := u.sdeDataRepo.UpsertMetaGroups(ctx, data.MetaGroups); err != nil {
		return errors.Wrap(err, "failed to upsert meta groups")
	}

	if err := u.sdeDataRepo.UpsertMarketGroups(ctx, data.MarketGroups); err != nil {
		return errors.Wrap(err, "failed to upsert market groups")
	}

	if err := u.sdeDataRepo.UpsertIcons(ctx, data.Icons); err != nil {
		return errors.Wrap(err, "failed to upsert icons")
	}

	if err := u.sdeDataRepo.UpsertGraphics(ctx, data.Graphics); err != nil {
		return errors.Wrap(err, "failed to upsert graphics")
	}

	if err := u.sdeDataRepo.UpsertBlueprints(ctx, data.Blueprints, data.BlueprintActivities, data.BlueprintMaterials, data.BlueprintProducts, data.BlueprintSkills); err != nil {
		return errors.Wrap(err, "failed to upsert blueprints")
	}

	if err := u.sdeDataRepo.UpsertDogma(ctx, data.DogmaAttributeCategories, data.DogmaAttributes, data.DogmaEffects, data.TypeDogmaAttributes, data.TypeDogmaEffects); err != nil {
		return errors.Wrap(err, "failed to upsert dogma")
	}

	if err := u.sdeDataRepo.UpsertNpcData(ctx, data.Factions, data.NpcCorporations, data.NpcCorporationDivisions, data.Agents, data.AgentsInSpace, data.Races, data.Bloodlines, data.Ancestries); err != nil {
		return errors.Wrap(err, "failed to upsert NPC data")
	}

	if err := u.sdeDataRepo.UpsertIndustryData(ctx, data.PlanetSchematics, data.PlanetSchematicTypes, data.ControlTowerResources); err != nil {
		return errors.Wrap(err, "failed to upsert industry data")
	}

	if err := u.sdeDataRepo.UpsertMiscData(ctx, data.Skins, data.SkinLicenses, data.SkinMaterials, data.Certificates, data.Landmarks, data.StationOperations, data.StationServices, data.ContrabandTypes, data.ResearchAgents, data.CharacterAttributes, data.CorporationActivities, data.TournamentRuleSets); err != nil {
		return errors.Wrap(err, "failed to upsert misc data")
	}

	// Step 7: Update checksum
	if err := u.sdeDataRepo.SetMetadata(ctx, "checksum", checksum); err != nil {
		return errors.Wrap(err, "failed to update SDE checksum")
	}

	log.Info("SDE update complete", "checksum", checksum)
	return nil
}
