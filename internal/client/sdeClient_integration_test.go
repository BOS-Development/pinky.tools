package client_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_SdeClient_Integration_ParseRealSDE downloads the actual SDE from CCP
// and validates that all YAML files parse without errors. This catches struct
// mismatches (e.g., plain string vs localized map) against the live data format.
//
// Run with: SDE_INTEGRATION_TEST=1 go test -v -run Test_SdeClient_Integration -timeout 300s ./internal/client/
func Test_SdeClient_Integration_ParseRealSDE(t *testing.T) {
	if os.Getenv("SDE_INTEGRATION_TEST") == "" {
		t.Skip("Skipping SDE integration test (set SDE_INTEGRATION_TEST=1 to run)")
	}

	httpClient := &http.Client{}
	c := client.NewSdeClient(httpClient)
	ctx := context.Background()

	// Step 1: Verify checksum endpoint works
	checksum, err := c.GetChecksum(ctx)
	require.NoError(t, err, "GetChecksum should not fail")
	assert.NotEmpty(t, checksum, "Checksum should not be empty")
	t.Logf("SDE build number: %s", checksum)

	// Step 2: Download the SDE ZIP
	zipPath, err := c.DownloadSDE(ctx)
	require.NoError(t, err, "DownloadSDE should not fail")
	defer os.Remove(zipPath)

	info, err := os.Stat(zipPath)
	require.NoError(t, err)
	t.Logf("SDE ZIP size: %d bytes", info.Size())
	assert.Greater(t, info.Size(), int64(1000000), "SDE ZIP should be at least 1MB")

	// Step 3: Parse all YAML files — this is the key validation
	data, err := c.ParseSDE(zipPath)
	require.NoError(t, err, "ParseSDE should not fail — if this fails, a YAML struct likely has wrong field types")

	// Step 4: Validate that key data was actually parsed (not just silently skipped)
	assert.NotEmpty(t, data.Types, "Should have parsed types")
	assert.NotEmpty(t, data.Categories, "Should have parsed categories")
	assert.NotEmpty(t, data.Groups, "Should have parsed groups")
	assert.NotEmpty(t, data.Blueprints, "Should have parsed blueprints")
	assert.NotEmpty(t, data.BlueprintActivities, "Should have parsed blueprint activities")
	assert.NotEmpty(t, data.BlueprintMaterials, "Should have parsed blueprint materials")
	assert.NotEmpty(t, data.BlueprintProducts, "Should have parsed blueprint products")
	assert.NotEmpty(t, data.BlueprintSkills, "Should have parsed blueprint skills")
	assert.NotEmpty(t, data.Regions, "Should have parsed regions")
	assert.NotEmpty(t, data.Constellations, "Should have parsed constellations")
	assert.NotEmpty(t, data.SolarSystems, "Should have parsed solar systems")
	assert.NotEmpty(t, data.Stations, "Should have parsed stations")
	assert.NotEmpty(t, data.MarketGroups, "Should have parsed market groups")
	assert.NotEmpty(t, data.MetaGroups, "Should have parsed meta groups")
	assert.NotEmpty(t, data.DogmaAttributes, "Should have parsed dogma attributes")
	assert.NotEmpty(t, data.DogmaEffects, "Should have parsed dogma effects")
	assert.NotEmpty(t, data.DogmaAttributeCategories, "Should have parsed dogma attribute categories")
	assert.NotEmpty(t, data.TypeDogmaAttributes, "Should have parsed type dogma attributes")
	assert.NotEmpty(t, data.TypeDogmaEffects, "Should have parsed type dogma effects")
	assert.NotEmpty(t, data.Factions, "Should have parsed factions")
	assert.NotEmpty(t, data.NpcCorporations, "Should have parsed NPC corporations")
	assert.NotEmpty(t, data.Races, "Should have parsed races")
	assert.NotEmpty(t, data.Bloodlines, "Should have parsed bloodlines")
	assert.NotEmpty(t, data.Ancestries, "Should have parsed ancestries")
	assert.NotEmpty(t, data.Skins, "Should have parsed skins")
	assert.NotEmpty(t, data.SkinLicenses, "Should have parsed skin licenses")
	assert.NotEmpty(t, data.SkinMaterials, "Should have parsed skin materials")
	assert.NotEmpty(t, data.Certificates, "Should have parsed certificates")
	assert.NotEmpty(t, data.Icons, "Should have parsed icons")
	assert.NotEmpty(t, data.Graphics, "Should have parsed graphics")

	// Note: npcStations.yaml does not include station names — names are empty.
	// The station upsert preserves existing names when the SDE provides an empty one.

	// Log counts for visibility
	t.Logf("Parsed: %d types, %d categories, %d groups, %d blueprints, %d regions, %d systems",
		len(data.Types), len(data.Categories), len(data.Groups),
		len(data.Blueprints), len(data.Regions), len(data.SolarSystems))
	t.Logf("Parsed: %d dogma attrs, %d dogma effects, %d type dogma attrs, %d type dogma effects",
		len(data.DogmaAttributes), len(data.DogmaEffects),
		len(data.TypeDogmaAttributes), len(data.TypeDogmaEffects))
	t.Logf("Parsed: %d factions, %d NPC corps, %d races, %d certificates, %d skins",
		len(data.Factions), len(data.NpcCorporations), len(data.Races),
		len(data.Certificates), len(data.Skins))
}
