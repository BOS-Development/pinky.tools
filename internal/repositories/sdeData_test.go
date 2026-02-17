package repositories_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SdeDataShouldSetAndGetMetadata(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewSdeDataRepository(db)
	ctx := context.Background()

	// Should return nil when key doesn't exist
	result, err := repo.GetMetadata(ctx, "checksum")
	assert.NoError(t, err)
	assert.Nil(t, result)

	// Set metadata
	err = repo.SetMetadata(ctx, "checksum", "abc123")
	assert.NoError(t, err)

	// Get metadata
	result, err = repo.GetMetadata(ctx, "checksum")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "checksum", result.Key)
	assert.Equal(t, "abc123", result.Value)

	// Update metadata (ON CONFLICT upsert)
	err = repo.SetMetadata(ctx, "checksum", "def456")
	assert.NoError(t, err)

	result, err = repo.GetMetadata(ctx, "checksum")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "def456", result.Value)
}

func Test_SdeDataShouldGetMetadataLastUpdateTime(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewSdeDataRepository(db)
	ctx := context.Background()

	// Empty table returns nil
	lastUpdate, err := repo.GetMetadataLastUpdateTime(ctx)
	assert.NoError(t, err)
	assert.Nil(t, lastUpdate)

	// After inserting metadata, returns non-nil
	err = repo.SetMetadata(ctx, "checksum", "abc123")
	assert.NoError(t, err)

	lastUpdate, err = repo.GetMetadataLastUpdateTime(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, lastUpdate)
}

func Test_SdeDataShouldUpsertCategories(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewSdeDataRepository(db)
	ctx := context.Background()

	iconID := int64(100)
	categories := []models.SdeCategory{
		{CategoryID: 1, Name: "Ship", Published: true, IconID: &iconID},
		{CategoryID: 2, Name: "Module", Published: true, IconID: nil},
		{CategoryID: 3, Name: "Charge", Published: false, IconID: nil},
	}

	err = repo.UpsertCategories(ctx, categories)
	assert.NoError(t, err)

	// Re-upsert should update existing rows without error
	categories = []models.SdeCategory{
		{CategoryID: 1, Name: "Ship Updated", Published: true, IconID: &iconID},
		{CategoryID: 4, Name: "Drone", Published: true, IconID: nil},
	}

	err = repo.UpsertCategories(ctx, categories)
	assert.NoError(t, err)
}

func Test_SdeDataShouldUpsertGroups(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewSdeDataRepository(db)
	ctx := context.Background()

	iconID := int64(200)
	groups := []models.SdeGroup{
		{GroupID: 10, Name: "Frigate", CategoryID: 1, Published: true, IconID: &iconID},
		{GroupID: 20, Name: "Cruiser", CategoryID: 1, Published: true, IconID: nil},
	}

	err = repo.UpsertGroups(ctx, groups)
	assert.NoError(t, err)
}

func Test_SdeDataShouldUpsertMetaGroups(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewSdeDataRepository(db)
	ctx := context.Background()

	metaGroups := []models.SdeMetaGroup{
		{MetaGroupID: 1, Name: "Tech I"},
		{MetaGroupID: 2, Name: "Tech II"},
		{MetaGroupID: 14, Name: "Tech III"},
	}

	err = repo.UpsertMetaGroups(ctx, metaGroups)
	assert.NoError(t, err)
}

func Test_SdeDataShouldUpsertMarketGroups(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewSdeDataRepository(db)
	ctx := context.Background()

	parentID := int64(1)
	desc := "Ships of all sizes"
	iconID := int64(300)
	marketGroups := []models.SdeMarketGroup{
		{MarketGroupID: 1, ParentGroupID: nil, Name: "Ships", Description: &desc, IconID: &iconID, HasTypes: false},
		{MarketGroupID: 2, ParentGroupID: &parentID, Name: "Frigates", Description: nil, IconID: nil, HasTypes: true},
	}

	err = repo.UpsertMarketGroups(ctx, marketGroups)
	assert.NoError(t, err)
}

func Test_SdeDataShouldUpsertIcons(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewSdeDataRepository(db)
	ctx := context.Background()

	desc := "Tritanium icon"
	icons := []models.SdeIcon{
		{IconID: 1, Description: &desc},
		{IconID: 2, Description: nil},
	}

	err = repo.UpsertIcons(ctx, icons)
	assert.NoError(t, err)
}

func Test_SdeDataShouldUpsertGraphics(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewSdeDataRepository(db)
	ctx := context.Background()

	desc := "Model graphic"
	graphics := []models.SdeGraphic{
		{GraphicID: 1, Description: &desc},
		{GraphicID: 2, Description: nil},
	}

	err = repo.UpsertGraphics(ctx, graphics)
	assert.NoError(t, err)
}

func Test_SdeDataShouldUpsertBlueprints(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewSdeDataRepository(db)
	ctx := context.Background()

	maxProd := 10
	prob := 0.5

	blueprints := []models.SdeBlueprint{
		{BlueprintTypeID: 1000, MaxProductionLimit: &maxProd},
		{BlueprintTypeID: 1001, MaxProductionLimit: nil},
	}

	activities := []models.SdeBlueprintActivity{
		{BlueprintTypeID: 1000, Activity: "manufacturing", Time: 3600},
		{BlueprintTypeID: 1000, Activity: "research_time", Time: 1800},
		{BlueprintTypeID: 1001, Activity: "manufacturing", Time: 7200},
	}

	materials := []models.SdeBlueprintMaterial{
		{BlueprintTypeID: 1000, Activity: "manufacturing", TypeID: 34, Quantity: 100},
		{BlueprintTypeID: 1000, Activity: "manufacturing", TypeID: 35, Quantity: 50},
	}

	products := []models.SdeBlueprintProduct{
		{BlueprintTypeID: 1000, Activity: "manufacturing", TypeID: 200, Quantity: 1, Probability: nil},
		{BlueprintTypeID: 1001, Activity: "manufacturing", TypeID: 201, Quantity: 1, Probability: &prob},
	}

	skills := []models.SdeBlueprintSkill{
		{BlueprintTypeID: 1000, Activity: "manufacturing", TypeID: 3380, Level: 1},
		{BlueprintTypeID: 1000, Activity: "manufacturing", TypeID: 3385, Level: 3},
	}

	err = repo.UpsertBlueprints(ctx, blueprints, activities, materials, products, skills)
	assert.NoError(t, err)

	// Re-upsert should update existing rows
	err = repo.UpsertBlueprints(ctx, blueprints, activities, materials, products, skills)
	assert.NoError(t, err)
}

func Test_SdeDataShouldUpsertDogma(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewSdeDataRepository(db)
	ctx := context.Background()

	catName := "Fitting"
	catDesc := "Fitting attributes"
	attrName := "power"
	attrDesc := "CPU Power"
	attrDisplay := "CPU"
	defaultVal := 0.0
	catID := int64(1)
	highIsGood := true
	stackable := false
	published := true
	unitID := int64(5)
	effectName := "online"
	effectDesc := "Module online effect"
	effectDisplay := "Online"

	attrCats := []models.SdeDogmaAttributeCategory{
		{CategoryID: 1, Name: &catName, Description: &catDesc},
		{CategoryID: 2, Name: nil, Description: nil},
	}

	attrs := []models.SdeDogmaAttribute{
		{AttributeID: 50, Name: &attrName, Description: &attrDesc, DefaultValue: &defaultVal, DisplayName: &attrDisplay, CategoryID: &catID, HighIsGood: &highIsGood, Stackable: &stackable, Published: &published, UnitID: &unitID},
		{AttributeID: 51, Name: nil, Description: nil, DefaultValue: nil, DisplayName: nil, CategoryID: nil, HighIsGood: nil, Stackable: nil, Published: nil, UnitID: nil},
	}

	effects := []models.SdeDogmaEffect{
		{EffectID: 10, Name: &effectName, Description: &effectDesc, DisplayName: &effectDisplay, CategoryID: &catID},
		{EffectID: 11, Name: nil, Description: nil, DisplayName: nil, CategoryID: nil},
	}

	typeAttrs := []models.SdeTypeDogmaAttribute{
		{TypeID: 34, AttributeID: 50, Value: 100.0},
		{TypeID: 35, AttributeID: 50, Value: 200.0},
	}

	typeEffects := []models.SdeTypeDogmaEffect{
		{TypeID: 34, EffectID: 10, IsDefault: true},
		{TypeID: 35, EffectID: 11, IsDefault: false},
	}

	err = repo.UpsertDogma(ctx, attrCats, attrs, effects, typeAttrs, typeEffects)
	assert.NoError(t, err)

	// Re-upsert should update existing rows
	err = repo.UpsertDogma(ctx, attrCats, attrs, effects, typeAttrs, typeEffects)
	assert.NoError(t, err)
}

func Test_SdeDataShouldUpsertNpcData(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewSdeDataRepository(db)
	ctx := context.Background()

	corpID := int64(1000125)
	factionIconID := int64(500)
	factionDesc := "The Caldari State"
	agentName := "Agent Smith"
	agentCorpID := int64(1000125)
	agentDivID := int64(1)
	agentLevel := 3
	systemID := int64(30000142)
	raceDesc := "Caldari"
	raceIconID := int64(600)
	blDesc := "Civire"
	blIconID := int64(700)
	ancestryDesc := "Tube child"
	ancestryIconID := int64(800)

	factions := []models.SdeFaction{
		{FactionID: 500001, Name: "Caldari State", Description: &factionDesc, CorporationID: &corpID, IconID: &factionIconID},
	}

	corps := []models.SdeNpcCorporation{
		{CorporationID: 1000125, Name: "Caldari Navy", FactionID: nil, IconID: nil},
	}

	divs := []models.SdeNpcCorporationDivision{
		{CorporationID: 1000125, DivisionID: 1, Name: "Accounting"},
	}

	agents := []models.SdeAgent{
		{AgentID: 3000001, Name: &agentName, CorporationID: &agentCorpID, DivisionID: &agentDivID, Level: &agentLevel},
	}

	agentsInSpace := []models.SdeAgentInSpace{
		{AgentID: 3000001, SolarSystemID: &systemID},
	}

	races := []models.SdeRace{
		{RaceID: 1, Name: "Caldari", Description: &raceDesc, IconID: &raceIconID},
	}

	bloodlines := []models.SdeBloodline{
		{BloodlineID: 1, Name: "Civire", RaceID: nil, Description: &blDesc, IconID: &blIconID},
	}

	ancestries := []models.SdeAncestry{
		{AncestryID: 1, Name: "Tube Child", BloodlineID: nil, Description: &ancestryDesc, IconID: &ancestryIconID},
	}

	err = repo.UpsertNpcData(ctx, factions, corps, divs, agents, agentsInSpace, races, bloodlines, ancestries)
	assert.NoError(t, err)

	// Re-upsert should update existing rows
	err = repo.UpsertNpcData(ctx, factions, corps, divs, agents, agentsInSpace, races, bloodlines, ancestries)
	assert.NoError(t, err)
}

func Test_SdeDataShouldUpsertIndustryData(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewSdeDataRepository(db)
	ctx := context.Background()

	purpose := 1
	minSec := 0.5
	factionID := int64(500001)

	schematics := []models.SdePlanetSchematic{
		{SchematicID: 1, Name: "Water", CycleTime: 1800},
		{SchematicID: 2, Name: "Electrolytes", CycleTime: 3600},
	}

	schematicTypes := []models.SdePlanetSchematicType{
		{SchematicID: 1, TypeID: 2268, Quantity: 3000, IsInput: true},
		{SchematicID: 1, TypeID: 2390, Quantity: 20, IsInput: false},
	}

	towerResources := []models.SdeControlTowerResource{
		{ControlTowerTypeID: 12235, ResourceTypeID: 4246, Purpose: &purpose, Quantity: 40, MinSecurity: &minSec, FactionID: &factionID},
		{ControlTowerTypeID: 12235, ResourceTypeID: 4247, Purpose: nil, Quantity: 20, MinSecurity: nil, FactionID: nil},
	}

	err = repo.UpsertIndustryData(ctx, schematics, schematicTypes, towerResources)
	assert.NoError(t, err)

	// Re-upsert should update existing rows
	err = repo.UpsertIndustryData(ctx, schematics, schematicTypes, towerResources)
	assert.NoError(t, err)
}

func Test_SdeDataShouldUpsertMiscData(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewSdeDataRepository(db)
	ctx := context.Background()

	skinTypeID := int64(34)
	skinMatID := int64(1)
	skinID := int64(100)
	duration := 30
	matName := "Red Paint"
	certName := "Core Competency"
	certDesc := "Basic certificate"
	certGroupID := int64(10)
	landmarkName := "Eve Gate"
	landmarkDesc := "The collapsed wormhole"
	opName := "Manufacturing"
	opDesc := "Station operation"
	svcName := "Repair"
	svcDesc := "Repair service"
	standingLoss := 1.5
	fineByValue := 0.25
	charAttrName := "Charisma"
	charAttrDesc := "Social attribute"
	charAttrIconID := int64(900)
	corpActName := "Manufacturing"
	ruleData := "{\"rules\":[]}"

	skins := []models.SdeSkin{
		{SkinID: 100, TypeID: &skinTypeID, MaterialID: &skinMatID},
		{SkinID: 101, TypeID: nil, MaterialID: nil},
	}

	skinLicenses := []models.SdeSkinLicense{
		{LicenseTypeID: 200, SkinID: &skinID, Duration: &duration},
		{LicenseTypeID: 201, SkinID: nil, Duration: nil},
	}

	skinMaterials := []models.SdeSkinMaterial{
		{SkinMaterialID: 1, Name: &matName},
		{SkinMaterialID: 2, Name: nil},
	}

	certificates := []models.SdeCertificate{
		{CertificateID: 1, Name: &certName, Description: &certDesc, GroupID: &certGroupID},
	}

	landmarks := []models.SdeLandmark{
		{LandmarkID: 1, Name: &landmarkName, Description: &landmarkDesc},
	}

	stationOps := []models.SdeStationOperation{
		{OperationID: 1, Name: &opName, Description: &opDesc},
	}

	stationSvcs := []models.SdeStationService{
		{ServiceID: 1, Name: &svcName, Description: &svcDesc},
	}

	contrabandTypes := []models.SdeContrabandType{
		{FactionID: 500001, TypeID: 34, StandingLoss: &standingLoss, FineByValue: &fineByValue},
	}

	researchAgents := []models.SdeResearchAgent{
		{AgentID: 3000001, TypeID: 34},
	}

	charAttrs := []models.SdeCharacterAttribute{
		{AttributeID: 1, Name: &charAttrName, Description: &charAttrDesc, IconID: &charAttrIconID},
	}

	corpActivities := []models.SdeCorporationActivity{
		{ActivityID: 1, Name: &corpActName},
	}

	tournamentRuleSets := []models.SdeTournamentRuleSet{
		{RuleSetID: 1, Data: &ruleData},
	}

	err = repo.UpsertMiscData(ctx, skins, skinLicenses, skinMaterials, certificates, landmarks, stationOps, stationSvcs, contrabandTypes, researchAgents, charAttrs, corpActivities, tournamentRuleSets)
	assert.NoError(t, err)

	// Re-upsert should update existing rows
	err = repo.UpsertMiscData(ctx, skins, skinLicenses, skinMaterials, certificates, landmarks, stationOps, stationSvcs, contrabandTypes, researchAgents, charAttrs, corpActivities, tournamentRuleSets)
	assert.NoError(t, err)
}

func Test_SdeDataShouldHandleEmptySlices(t *testing.T) {
	db, err := setupDatabase(t)
	require.NoError(t, err)

	repo := repositories.NewSdeDataRepository(db)
	ctx := context.Background()

	// All empty slice operations should succeed
	err = repo.UpsertCategories(ctx, []models.SdeCategory{})
	assert.NoError(t, err)

	err = repo.UpsertGroups(ctx, []models.SdeGroup{})
	assert.NoError(t, err)

	err = repo.UpsertMetaGroups(ctx, []models.SdeMetaGroup{})
	assert.NoError(t, err)

	err = repo.UpsertMarketGroups(ctx, []models.SdeMarketGroup{})
	assert.NoError(t, err)

	err = repo.UpsertIcons(ctx, []models.SdeIcon{})
	assert.NoError(t, err)

	err = repo.UpsertGraphics(ctx, []models.SdeGraphic{})
	assert.NoError(t, err)

	err = repo.UpsertBlueprints(ctx, []models.SdeBlueprint{}, []models.SdeBlueprintActivity{}, []models.SdeBlueprintMaterial{}, []models.SdeBlueprintProduct{}, []models.SdeBlueprintSkill{})
	assert.NoError(t, err)

	err = repo.UpsertDogma(ctx, []models.SdeDogmaAttributeCategory{}, []models.SdeDogmaAttribute{}, []models.SdeDogmaEffect{}, []models.SdeTypeDogmaAttribute{}, []models.SdeTypeDogmaEffect{})
	assert.NoError(t, err)

	err = repo.UpsertNpcData(ctx, []models.SdeFaction{}, []models.SdeNpcCorporation{}, []models.SdeNpcCorporationDivision{}, []models.SdeAgent{}, []models.SdeAgentInSpace{}, []models.SdeRace{}, []models.SdeBloodline{}, []models.SdeAncestry{})
	assert.NoError(t, err)

	err = repo.UpsertIndustryData(ctx, []models.SdePlanetSchematic{}, []models.SdePlanetSchematicType{}, []models.SdeControlTowerResource{})
	assert.NoError(t, err)

	err = repo.UpsertMiscData(ctx, []models.SdeSkin{}, []models.SdeSkinLicense{}, []models.SdeSkinMaterial{}, []models.SdeCertificate{}, []models.SdeLandmark{}, []models.SdeStationOperation{}, []models.SdeStationService{}, []models.SdeContrabandType{}, []models.SdeResearchAgent{}, []models.SdeCharacterAttribute{}, []models.SdeCorporationActivity{}, []models.SdeTournamentRuleSet{})
	assert.NoError(t, err)
}
