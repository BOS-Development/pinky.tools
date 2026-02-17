package updaters_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/updaters"
	"github.com/stretchr/testify/assert"
)

// Mock SDE client

type mockSdeClient struct {
	checksumResult string
	checksumErr    error
	downloadResult string
	downloadErr    error
	parseResult    *client.SdeData
	parseErr       error
}

func (m *mockSdeClient) GetChecksum(ctx context.Context) (string, error) {
	return m.checksumResult, m.checksumErr
}
func (m *mockSdeClient) DownloadSDE(ctx context.Context) (string, error) {
	return m.downloadResult, m.downloadErr
}
func (m *mockSdeClient) ParseSDE(zipPath string) (*client.SdeData, error) {
	return m.parseResult, m.parseErr
}

// Mock SDE data repository

type mockSdeDataRepo struct {
	metadata    map[string]*models.SdeMetadata
	getMetaErr  error
	setMetaErr  error
	upsertErr   error
	upsertCalls int
}

func newMockSdeDataRepo() *mockSdeDataRepo {
	return &mockSdeDataRepo{metadata: make(map[string]*models.SdeMetadata)}
}

func (m *mockSdeDataRepo) GetMetadata(ctx context.Context, key string) (*models.SdeMetadata, error) {
	if m.getMetaErr != nil {
		return nil, m.getMetaErr
	}
	return m.metadata[key], nil
}
func (m *mockSdeDataRepo) SetMetadata(ctx context.Context, key, value string) error {
	if m.setMetaErr != nil {
		return m.setMetaErr
	}
	m.metadata[key] = &models.SdeMetadata{Key: key, Value: value}
	return nil
}
func (m *mockSdeDataRepo) UpsertCategories(ctx context.Context, categories []models.SdeCategory) error {
	m.upsertCalls++
	return m.upsertErr
}
func (m *mockSdeDataRepo) UpsertGroups(ctx context.Context, groups []models.SdeGroup) error {
	m.upsertCalls++
	return m.upsertErr
}
func (m *mockSdeDataRepo) UpsertMetaGroups(ctx context.Context, metaGroups []models.SdeMetaGroup) error {
	m.upsertCalls++
	return m.upsertErr
}
func (m *mockSdeDataRepo) UpsertMarketGroups(ctx context.Context, marketGroups []models.SdeMarketGroup) error {
	m.upsertCalls++
	return m.upsertErr
}
func (m *mockSdeDataRepo) UpsertIcons(ctx context.Context, icons []models.SdeIcon) error {
	m.upsertCalls++
	return m.upsertErr
}
func (m *mockSdeDataRepo) UpsertGraphics(ctx context.Context, graphics []models.SdeGraphic) error {
	m.upsertCalls++
	return m.upsertErr
}
func (m *mockSdeDataRepo) UpsertBlueprints(ctx context.Context, blueprints []models.SdeBlueprint, activities []models.SdeBlueprintActivity, materials []models.SdeBlueprintMaterial, products []models.SdeBlueprintProduct, skills []models.SdeBlueprintSkill) error {
	m.upsertCalls++
	return m.upsertErr
}
func (m *mockSdeDataRepo) UpsertDogma(ctx context.Context, attrCats []models.SdeDogmaAttributeCategory, attrs []models.SdeDogmaAttribute, effects []models.SdeDogmaEffect, typeAttrs []models.SdeTypeDogmaAttribute, typeEffects []models.SdeTypeDogmaEffect) error {
	m.upsertCalls++
	return m.upsertErr
}
func (m *mockSdeDataRepo) UpsertNpcData(ctx context.Context, factions []models.SdeFaction, corps []models.SdeNpcCorporation, divs []models.SdeNpcCorporationDivision, agents []models.SdeAgent, agentsInSpace []models.SdeAgentInSpace, races []models.SdeRace, bloodlines []models.SdeBloodline, ancestries []models.SdeAncestry) error {
	m.upsertCalls++
	return m.upsertErr
}
func (m *mockSdeDataRepo) UpsertIndustryData(ctx context.Context, schematics []models.SdePlanetSchematic, schematicTypes []models.SdePlanetSchematicType, towerResources []models.SdeControlTowerResource) error {
	m.upsertCalls++
	return m.upsertErr
}
func (m *mockSdeDataRepo) UpsertMiscData(ctx context.Context, skins []models.SdeSkin, skinLicenses []models.SdeSkinLicense, skinMaterials []models.SdeSkinMaterial, certificates []models.SdeCertificate, landmarks []models.SdeLandmark, stationOps []models.SdeStationOperation, stationSvcs []models.SdeStationService, contrabandTypes []models.SdeContrabandType, researchAgents []models.SdeResearchAgent, charAttrs []models.SdeCharacterAttribute, corpActivities []models.SdeCorporationActivity, tournamentRuleSets []models.SdeTournamentRuleSet) error {
	m.upsertCalls++
	return m.upsertErr
}

// Mock existing table repositories

type mockItemTypeRepo struct{ err error }

func (m *mockItemTypeRepo) UpsertItemTypes(ctx context.Context, itemTypes []models.EveInventoryType) error {
	return m.err
}

type mockRegionRepo struct{ err error }

func (m *mockRegionRepo) Upsert(ctx context.Context, regions []models.Region) error { return m.err }

type mockConstellationRepo struct{ err error }

func (m *mockConstellationRepo) Upsert(ctx context.Context, constellations []models.Constellation) error {
	return m.err
}

type mockSolarSystemRepo struct{ err error }

func (m *mockSolarSystemRepo) Upsert(ctx context.Context, systems []models.SolarSystem) error {
	return m.err
}

type mockStationRepo struct{ err error }

func (m *mockStationRepo) Upsert(ctx context.Context, stations []models.Station) error {
	return m.err
}

func emptySdeData() *client.SdeData {
	return &client.SdeData{}
}

func Test_SdeUpdater_FullUpdateWhenChecksumDiffers(t *testing.T) {
	sdeClient := &mockSdeClient{
		checksumResult: "new-checksum-abc",
		downloadResult: "/tmp/fake.zip",
		parseResult:    emptySdeData(),
	}
	repo := newMockSdeDataRepo()

	u := updaters.NewSde(sdeClient, repo, &mockItemTypeRepo{}, &mockRegionRepo{}, &mockConstellationRepo{}, &mockSolarSystemRepo{}, &mockStationRepo{})

	err := u.Update(context.Background())
	assert.NoError(t, err)

	// Checksum should be stored
	assert.Equal(t, "new-checksum-abc", repo.metadata["checksum"].Value)
}

func Test_SdeUpdater_SkipsWhenChecksumMatches(t *testing.T) {
	sdeClient := &mockSdeClient{
		checksumResult: "existing-checksum",
	}
	repo := newMockSdeDataRepo()
	repo.metadata["checksum"] = &models.SdeMetadata{Key: "checksum", Value: "existing-checksum"}

	u := updaters.NewSde(sdeClient, repo, &mockItemTypeRepo{}, &mockRegionRepo{}, &mockConstellationRepo{}, &mockSolarSystemRepo{}, &mockStationRepo{})

	err := u.Update(context.Background())
	assert.NoError(t, err)

	// No upserts should have happened
	assert.Equal(t, 0, repo.upsertCalls)
}

func Test_SdeUpdater_ErrorGettingChecksum(t *testing.T) {
	sdeClient := &mockSdeClient{
		checksumErr: fmt.Errorf("network error"),
	}
	repo := newMockSdeDataRepo()

	u := updaters.NewSde(sdeClient, repo, &mockItemTypeRepo{}, &mockRegionRepo{}, &mockConstellationRepo{}, &mockSolarSystemRepo{}, &mockStationRepo{})

	err := u.Update(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get SDE checksum")
}

func Test_SdeUpdater_ErrorGettingStoredMetadata(t *testing.T) {
	sdeClient := &mockSdeClient{
		checksumResult: "new-checksum",
	}
	repo := newMockSdeDataRepo()
	repo.getMetaErr = fmt.Errorf("db error")

	u := updaters.NewSde(sdeClient, repo, &mockItemTypeRepo{}, &mockRegionRepo{}, &mockConstellationRepo{}, &mockSolarSystemRepo{}, &mockStationRepo{})

	err := u.Update(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get stored SDE checksum")
}

func Test_SdeUpdater_ErrorDownloading(t *testing.T) {
	sdeClient := &mockSdeClient{
		checksumResult: "new-checksum",
		downloadErr:    fmt.Errorf("download failed"),
	}
	repo := newMockSdeDataRepo()

	u := updaters.NewSde(sdeClient, repo, &mockItemTypeRepo{}, &mockRegionRepo{}, &mockConstellationRepo{}, &mockSolarSystemRepo{}, &mockStationRepo{})

	err := u.Update(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to download SDE")
}

func Test_SdeUpdater_ErrorParsing(t *testing.T) {
	sdeClient := &mockSdeClient{
		checksumResult: "new-checksum",
		downloadResult: "/tmp/fake.zip",
		parseErr:       fmt.Errorf("invalid yaml"),
	}
	repo := newMockSdeDataRepo()

	u := updaters.NewSde(sdeClient, repo, &mockItemTypeRepo{}, &mockRegionRepo{}, &mockConstellationRepo{}, &mockSolarSystemRepo{}, &mockStationRepo{})

	err := u.Update(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse SDE")
}

func Test_SdeUpdater_ErrorUpsertingRegions(t *testing.T) {
	sdeClient := &mockSdeClient{
		checksumResult: "new-checksum",
		downloadResult: "/tmp/fake.zip",
		parseResult:    emptySdeData(),
	}
	repo := newMockSdeDataRepo()

	u := updaters.NewSde(sdeClient, repo, &mockItemTypeRepo{}, &mockRegionRepo{err: fmt.Errorf("db error")}, &mockConstellationRepo{}, &mockSolarSystemRepo{}, &mockStationRepo{})

	err := u.Update(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to upsert regions")
}

func Test_SdeUpdater_ErrorUpsertingItemTypes(t *testing.T) {
	sdeClient := &mockSdeClient{
		checksumResult: "new-checksum",
		downloadResult: "/tmp/fake.zip",
		parseResult:    emptySdeData(),
	}
	repo := newMockSdeDataRepo()

	u := updaters.NewSde(sdeClient, repo, &mockItemTypeRepo{err: fmt.Errorf("db error")}, &mockRegionRepo{}, &mockConstellationRepo{}, &mockSolarSystemRepo{}, &mockStationRepo{})

	err := u.Update(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to upsert item types")
}

func Test_SdeUpdater_ErrorSettingChecksum(t *testing.T) {
	sdeClient := &mockSdeClient{
		checksumResult: "new-checksum",
		downloadResult: "/tmp/fake.zip",
		parseResult:    emptySdeData(),
	}
	repo := newMockSdeDataRepo()
	repo.setMetaErr = fmt.Errorf("db error")

	u := updaters.NewSde(sdeClient, repo, &mockItemTypeRepo{}, &mockRegionRepo{}, &mockConstellationRepo{}, &mockSolarSystemRepo{}, &mockStationRepo{})

	err := u.Update(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update SDE checksum")
}

func Test_SdeUpdater_FirstRunWithNoStoredChecksum(t *testing.T) {
	sdeClient := &mockSdeClient{
		checksumResult: "first-checksum",
		downloadResult: "/tmp/fake.zip",
		parseResult:    emptySdeData(),
	}
	repo := newMockSdeDataRepo()
	// No stored checksum — nil metadata

	u := updaters.NewSde(sdeClient, repo, &mockItemTypeRepo{}, &mockRegionRepo{}, &mockConstellationRepo{}, &mockSolarSystemRepo{}, &mockStationRepo{})

	err := u.Update(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "first-checksum", repo.metadata["checksum"].Value)
}
