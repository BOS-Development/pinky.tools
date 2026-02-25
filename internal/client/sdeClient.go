package client

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const defaultSdeBaseURL = "https://developers.eveonline.com/static-data/"

// SdeData holds all parsed SDE data organized by table group
type SdeData struct {
	Types          []models.EveInventoryType
	Categories     []models.SdeCategory
	Groups         []models.SdeGroup
	MetaGroups     []models.SdeMetaGroup
	MarketGroups   []models.SdeMarketGroup
	Icons          []models.SdeIcon
	Graphics       []models.SdeGraphic
	Regions        []models.Region
	Constellations []models.Constellation
	SolarSystems   []models.SolarSystem
	Stations       []models.Station

	Blueprints         []models.SdeBlueprint
	BlueprintActivities []models.SdeBlueprintActivity
	BlueprintMaterials []models.SdeBlueprintMaterial
	BlueprintProducts  []models.SdeBlueprintProduct
	BlueprintSkills    []models.SdeBlueprintSkill

	DogmaAttributeCategories []models.SdeDogmaAttributeCategory
	DogmaAttributes          []models.SdeDogmaAttribute
	DogmaEffects             []models.SdeDogmaEffect
	TypeDogmaAttributes      []models.SdeTypeDogmaAttribute
	TypeDogmaEffects         []models.SdeTypeDogmaEffect

	Factions              []models.SdeFaction
	NpcCorporations       []models.SdeNpcCorporation
	NpcCorporationDivisions []models.SdeNpcCorporationDivision
	Agents                []models.SdeAgent
	AgentsInSpace         []models.SdeAgentInSpace
	Races                 []models.SdeRace
	Bloodlines            []models.SdeBloodline
	Ancestries            []models.SdeAncestry

	PlanetSchematics     []models.SdePlanetSchematic
	PlanetSchematicTypes []models.SdePlanetSchematicType
	ControlTowerResources []models.SdeControlTowerResource

	Skins          []models.SdeSkin
	SkinLicenses   []models.SdeSkinLicense
	SkinMaterials  []models.SdeSkinMaterial
	Certificates   []models.SdeCertificate
	Landmarks      []models.SdeLandmark
	StationOperations []models.SdeStationOperation
	StationServices   []models.SdeStationService
	ContrabandTypes   []models.SdeContrabandType
	ResearchAgents    []models.SdeResearchAgent
	CharacterAttributes []models.SdeCharacterAttribute
	CorporationActivities []models.SdeCorporationActivity
	TournamentRuleSets    []models.SdeTournamentRuleSet
}

type SdeClient struct {
	httpClient HTTPDoer
	baseURL    string
}

func NewSdeClient(httpClient HTTPDoer) *SdeClient {
	return &SdeClient{
		httpClient: httpClient,
		baseURL:    defaultSdeBaseURL,
	}
}

func NewSdeClientWithBaseURL(httpClient HTTPDoer, baseURL string) *SdeClient {
	return &SdeClient{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

func (c *SdeClient) GetChecksum(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"tranquility/latest.jsonl", nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to create checksum request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to fetch checksum")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code fetching checksum: %d", resp.StatusCode)
	}

	var latest struct {
		BuildNumber int64 `json:"buildNumber"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&latest); err != nil {
		return "", errors.Wrap(err, "failed to decode latest.jsonl")
	}

	return strconv.FormatInt(latest.BuildNumber, 10), nil
}

func (c *SdeClient) DownloadSDE(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"eve-online-static-data-latest-yaml.zip", nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to create SDE download request")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to download SDE")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code downloading SDE: %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "sde-*.zip")
	if err != nil {
		return "", errors.Wrap(err, "failed to create temp file for SDE")
	}
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", errors.Wrap(err, "failed to write SDE to temp file")
	}

	return tmpFile.Name(), nil
}

func (c *SdeClient) ParseSDE(zipPath string) (*SdeData, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open SDE ZIP")
	}
	defer reader.Close()

	data := &SdeData{}

	// Build a map of filename -> zip.File for quick lookup
	fileMap := make(map[string]*zip.File)
	for _, f := range reader.File {
		fileMap[f.Name] = f
	}

	// Parse each known YAML file (new SDE format as of 2025 rework — files at zip root)
	parsers := map[string]func(*zip.File, *SdeData) error{
		"types.yaml":                    parseTypeIDs,
		"categories.yaml":               parseCategoryIDs,
		"groups.yaml":                   parseGroupIDs,
		"metaGroups.yaml":               parseMetaGroups,
		"marketGroups.yaml":             parseMarketGroups,
		"icons.yaml":                    parseIconIDs,
		"graphics.yaml":                 parseGraphicIDs,
		"blueprints.yaml":               parseBlueprints,
		"dogmaAttributeCategories.yaml": parseDogmaAttributeCategories,
		"dogmaAttributes.yaml":          parseDogmaAttributes,
		"dogmaEffects.yaml":             parseDogmaEffects,
		"typeDogma.yaml":                parseTypeDogma,
		"factions.yaml":                 parseFactions,
		"npcCorporations.yaml":          parseNpcCorporations,
		"npcCorporationDivisions.yaml":  parseNpcCorporationDivisions,
		"agentsInSpace.yaml":            parseAgentsInSpace,
		"races.yaml":                    parseRaces,
		"bloodlines.yaml":               parseBloodlines,
		"ancestries.yaml":               parseAncestries,
		"planetSchematics.yaml":         parsePlanetSchematics,
		"controlTowerResources.yaml":    parseControlTowerResources,
		"skins.yaml":                    parseSkins,
		"skinLicenses.yaml":             parseSkinLicenses,
		"skinMaterials.yaml":            parseSkinMaterials,
		"certificates.yaml":             parseCertificates,
		"landmarks.yaml":                parseLandmarks,
		"stationOperations.yaml":        parseStationOperations,
		"stationServices.yaml":          parseStationServices,
		"contrabandTypes.yaml":          parseContrabandTypes,
		"characterAttributes.yaml":      parseCharacterAttributes,
		"corporationActivities.yaml":    parseCorporationActivities,
		"mapRegions.yaml":               parseRegions,
		"mapConstellations.yaml":        parseConstellations,
		"mapSolarSystems.yaml":          parseSolarSystems,
		"npcStations.yaml":              parseStations,
	}

	for name, parser := range parsers {
		if parser == nil {
			continue
		}
		f, ok := fileMap[name]
		if !ok {
			log.Info("SDE file not found in ZIP, skipping", "file", name)
			continue
		}
		if err := parser(f, data); err != nil {
			return nil, errors.Wrapf(err, "failed to parse %s", name)
		}
	}

	return data, nil
}

// YAML parse helpers

func openZipFile(f *zip.File) (io.ReadCloser, error) {
	return f.Open()
}

func parseYAMLMap[V any](f *zip.File) (map[int64]V, error) {
	rc, err := openZipFile(f)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	var result map[int64]V
	decoder := yaml.NewDecoder(rc)
	if err := decoder.Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// localizedString handles SDE fields that may be either a plain string or a
// localized map like {en: "...", de: "..."}. Implements yaml.Unmarshaler.
type localizedString struct {
	Value string
}

func (ls *localizedString) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		ls.Value = value.Value
		return nil
	}
	if value.Kind == yaml.MappingNode {
		var m map[string]string
		if err := value.Decode(&m); err != nil {
			return err
		}
		ls.Value = m["en"]
		return nil
	}
	return nil
}

// YAML structs for CCP's SDE format

type sdeTypeYAML struct {
	Name            map[string]string `yaml:"name"`
	Description     map[string]string `yaml:"description"`
	GroupID         *int64            `yaml:"groupID"`
	Volume          *float64          `yaml:"volume"`
	PackagedVolume  *float64          `yaml:"packagedVolume"`
	Mass            *float64          `yaml:"mass"`
	Capacity        *float64          `yaml:"capacity"`
	PortionSize     *int              `yaml:"portionSize"`
	Published       *bool             `yaml:"published"`
	MarketGroupID   *int64            `yaml:"marketGroupID"`
	IconID          *int64            `yaml:"iconID"`
	GraphicID       *int64            `yaml:"graphicID"`
	RaceID          *int64            `yaml:"raceID"`
}

func parseTypeIDs(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeTypeYAML](f)
	if err != nil {
		return err
	}

	types := make([]models.EveInventoryType, 0, len(raw))
	for typeID, t := range raw {
		name := t.Name["en"]
		if name == "" {
			continue
		}
		volume := float64(0)
		if t.Volume != nil {
			volume = *t.Volume
		}

		var desc *string
		if d, ok := t.Description["en"]; ok && d != "" {
			desc = &d
		}

		types = append(types, models.EveInventoryType{
			TypeID:         typeID,
			TypeName:       name,
			Volume:         volume,
			IconID:         t.IconID,
			GroupID:        t.GroupID,
			PackagedVolume: t.PackagedVolume,
			Mass:           t.Mass,
			Capacity:       t.Capacity,
			PortionSize:    t.PortionSize,
			Published:      t.Published,
			MarketGroupID:  t.MarketGroupID,
			GraphicID:      t.GraphicID,
			RaceID:         t.RaceID,
			Description:    desc,
		})
	}
	data.Types = types
	return nil
}

type sdeCategoryYAML struct {
	Name      map[string]string `yaml:"name"`
	Published bool              `yaml:"published"`
	IconID    *int64            `yaml:"iconID"`
}

func parseCategoryIDs(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeCategoryYAML](f)
	if err != nil {
		return err
	}

	cats := make([]models.SdeCategory, 0, len(raw))
	for id, c := range raw {
		cats = append(cats, models.SdeCategory{
			CategoryID: id,
			Name:       c.Name["en"],
			Published:  c.Published,
			IconID:     c.IconID,
		})
	}
	data.Categories = cats
	return nil
}

type sdeGroupYAML struct {
	Name       map[string]string `yaml:"name"`
	CategoryID int64             `yaml:"categoryID"`
	Published  bool              `yaml:"published"`
	IconID     *int64            `yaml:"iconID"`
}

func parseGroupIDs(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeGroupYAML](f)
	if err != nil {
		return err
	}

	groups := make([]models.SdeGroup, 0, len(raw))
	for id, g := range raw {
		groups = append(groups, models.SdeGroup{
			GroupID:    id,
			Name:      g.Name["en"],
			CategoryID: g.CategoryID,
			Published:  g.Published,
			IconID:     g.IconID,
		})
	}
	data.Groups = groups
	return nil
}

type sdeMetaGroupYAML struct {
	Name map[string]string `yaml:"name"`
}

func parseMetaGroups(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeMetaGroupYAML](f)
	if err != nil {
		return err
	}

	mgs := make([]models.SdeMetaGroup, 0, len(raw))
	for id, mg := range raw {
		mgs = append(mgs, models.SdeMetaGroup{
			MetaGroupID: id,
			Name:        mg.Name["en"],
		})
	}
	data.MetaGroups = mgs
	return nil
}

type sdeMarketGroupYAML struct {
	Name          map[string]string `yaml:"name"`
	Description   map[string]string `yaml:"description"`
	ParentGroupID *int64            `yaml:"parentGroupID"`
	IconID        *int64            `yaml:"iconID"`
	HasTypes      bool              `yaml:"hasTypes"`
}

func parseMarketGroups(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeMarketGroupYAML](f)
	if err != nil {
		return err
	}

	mgs := make([]models.SdeMarketGroup, 0, len(raw))
	for id, mg := range raw {
		var desc *string
		if d := mg.Description["en"]; d != "" {
			desc = &d
		}
		mgs = append(mgs, models.SdeMarketGroup{
			MarketGroupID: id,
			ParentGroupID: mg.ParentGroupID,
			Name:          mg.Name["en"],
			Description:   desc,
			IconID:        mg.IconID,
			HasTypes:      mg.HasTypes,
		})
	}
	data.MarketGroups = mgs
	return nil
}

type sdeIconYAML struct {
	Description string `yaml:"description"`
}

func parseIconIDs(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeIconYAML](f)
	if err != nil {
		return err
	}

	icons := make([]models.SdeIcon, 0, len(raw))
	for id, icon := range raw {
		var desc *string
		if icon.Description != "" {
			desc = &icon.Description
		}
		icons = append(icons, models.SdeIcon{
			IconID:      id,
			Description: desc,
		})
	}
	data.Icons = icons
	return nil
}

type sdeGraphicYAML struct {
	Description string `yaml:"description"`
}

func parseGraphicIDs(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeGraphicYAML](f)
	if err != nil {
		return err
	}

	graphics := make([]models.SdeGraphic, 0, len(raw))
	for id, g := range raw {
		var desc *string
		if g.Description != "" {
			desc = &g.Description
		}
		graphics = append(graphics, models.SdeGraphic{
			GraphicID:   id,
			Description: desc,
		})
	}
	data.Graphics = graphics
	return nil
}

// Blueprint YAML structures

type sdeBlueprintYAML struct {
	MaxProductionLimit *int                              `yaml:"maxProductionLimit"`
	Activities         map[string]sdeBlueprintActYAML    `yaml:"activities"`
}

type sdeBlueprintActYAML struct {
	Time      int                          `yaml:"time"`
	Materials []sdeBlueprintMaterialYAML   `yaml:"materials"`
	Products  []sdeBlueprintProductYAML    `yaml:"products"`
	Skills    []sdeBlueprintSkillYAML      `yaml:"skills"`
}

type sdeBlueprintMaterialYAML struct {
	TypeID   int64 `yaml:"typeID"`
	Quantity int   `yaml:"quantity"`
}

type sdeBlueprintProductYAML struct {
	TypeID      int64    `yaml:"typeID"`
	Quantity    int      `yaml:"quantity"`
	Probability *float64 `yaml:"probability"`
}

type sdeBlueprintSkillYAML struct {
	TypeID int64 `yaml:"typeID"`
	Level  int   `yaml:"level"`
}

func parseBlueprints(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeBlueprintYAML](f)
	if err != nil {
		return err
	}

	blueprints := make([]models.SdeBlueprint, 0, len(raw))
	activities := make([]models.SdeBlueprintActivity, 0)
	materials := make([]models.SdeBlueprintMaterial, 0)
	products := make([]models.SdeBlueprintProduct, 0)
	skills := make([]models.SdeBlueprintSkill, 0)

	for bpID, bp := range raw {
		blueprints = append(blueprints, models.SdeBlueprint{
			BlueprintTypeID:    bpID,
			MaxProductionLimit: bp.MaxProductionLimit,
		})

		for actName, act := range bp.Activities {
			activities = append(activities, models.SdeBlueprintActivity{
				BlueprintTypeID: bpID,
				Activity:        actName,
				Time:            act.Time,
			})

			for _, mat := range act.Materials {
				materials = append(materials, models.SdeBlueprintMaterial{
					BlueprintTypeID: bpID,
					Activity:        actName,
					TypeID:          mat.TypeID,
					Quantity:        mat.Quantity,
				})
			}

			for _, prod := range act.Products {
				products = append(products, models.SdeBlueprintProduct{
					BlueprintTypeID: bpID,
					Activity:        actName,
					TypeID:          prod.TypeID,
					Quantity:        prod.Quantity,
					Probability:     prod.Probability,
				})
			}

			for _, skill := range act.Skills {
				skills = append(skills, models.SdeBlueprintSkill{
					BlueprintTypeID: bpID,
					Activity:        actName,
					TypeID:          skill.TypeID,
					Level:           skill.Level,
				})
			}
		}
	}

	data.Blueprints = blueprints
	data.BlueprintActivities = activities
	data.BlueprintMaterials = materials
	data.BlueprintProducts = products
	data.BlueprintSkills = skills
	return nil
}

// Dogma YAML structures

type sdeDogmaAttrCatYAML struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

func parseDogmaAttributeCategories(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeDogmaAttrCatYAML](f)
	if err != nil {
		return err
	}

	cats := make([]models.SdeDogmaAttributeCategory, 0, len(raw))
	for id, c := range raw {
		var name, desc *string
		if c.Name != "" {
			name = &c.Name
		}
		if c.Description != "" {
			desc = &c.Description
		}
		cats = append(cats, models.SdeDogmaAttributeCategory{
			CategoryID:  id,
			Name:        name,
			Description: desc,
		})
	}
	data.DogmaAttributeCategories = cats
	return nil
}

type sdeDogmaAttrYAML struct {
	Name         localizedString `yaml:"name"`
	Description  localizedString `yaml:"description"`
	DefaultValue *float64        `yaml:"defaultValue"`
	DisplayName  localizedString `yaml:"displayName"`
	CategoryID   *int64          `yaml:"attributeCategoryID"`
	HighIsGood   *bool           `yaml:"highIsGood"`
	Stackable    *bool           `yaml:"stackable"`
	Published    *bool           `yaml:"published"`
	UnitID       *int64          `yaml:"unitID"`
}

func parseDogmaAttributes(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeDogmaAttrYAML](f)
	if err != nil {
		return err
	}

	attrs := make([]models.SdeDogmaAttribute, 0, len(raw))
	for id, a := range raw {
		var name, desc, displayName *string
		if a.Name.Value != "" {
			name = &a.Name.Value
		}
		if a.Description.Value != "" {
			desc = &a.Description.Value
		}
		if a.DisplayName.Value != "" {
			displayName = &a.DisplayName.Value
		}
		attrs = append(attrs, models.SdeDogmaAttribute{
			AttributeID:  id,
			Name:         name,
			Description:  desc,
			DefaultValue: a.DefaultValue,
			DisplayName:  displayName,
			CategoryID:   a.CategoryID,
			HighIsGood:   a.HighIsGood,
			Stackable:    a.Stackable,
			Published:    a.Published,
			UnitID:       a.UnitID,
		})
	}
	data.DogmaAttributes = attrs
	return nil
}

type sdeDogmaEffectYAML struct {
	Name        map[string]string `yaml:"effectName"`
	Description map[string]string `yaml:"description"`
	DisplayName map[string]string `yaml:"displayName"`
	CategoryID  *int64            `yaml:"effectCategoryID"`
}

func parseDogmaEffects(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeDogmaEffectYAML](f)
	if err != nil {
		return err
	}

	effects := make([]models.SdeDogmaEffect, 0, len(raw))
	for id, e := range raw {
		var name, desc, displayName *string
		if n := e.Name["en"]; n != "" {
			name = &n
		}
		if d := e.Description["en"]; d != "" {
			desc = &d
		}
		if dn := e.DisplayName["en"]; dn != "" {
			displayName = &dn
		}
		effects = append(effects, models.SdeDogmaEffect{
			EffectID:    id,
			Name:        name,
			Description: desc,
			DisplayName: displayName,
			CategoryID:  e.CategoryID,
		})
	}
	data.DogmaEffects = effects
	return nil
}

type sdeTypeDogmaYAML struct {
	DogmaAttributes []sdeTypeDogmaAttrYAML  `yaml:"dogmaAttributes"`
	DogmaEffects    []sdeTypeDogmaEffYAML   `yaml:"dogmaEffects"`
}

type sdeTypeDogmaAttrYAML struct {
	AttributeID int64   `yaml:"attributeID"`
	Value       float64 `yaml:"value"`
}

type sdeTypeDogmaEffYAML struct {
	EffectID  int64 `yaml:"effectID"`
	IsDefault bool  `yaml:"isDefault"`
}

func parseTypeDogma(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeTypeDogmaYAML](f)
	if err != nil {
		return err
	}

	attrs := make([]models.SdeTypeDogmaAttribute, 0)
	effects := make([]models.SdeTypeDogmaEffect, 0)

	for typeID, td := range raw {
		for _, a := range td.DogmaAttributes {
			attrs = append(attrs, models.SdeTypeDogmaAttribute{
				TypeID:      typeID,
				AttributeID: a.AttributeID,
				Value:       a.Value,
			})
		}
		for _, e := range td.DogmaEffects {
			effects = append(effects, models.SdeTypeDogmaEffect{
				TypeID:    typeID,
				EffectID:  e.EffectID,
				IsDefault: e.IsDefault,
			})
		}
	}

	data.TypeDogmaAttributes = attrs
	data.TypeDogmaEffects = effects
	return nil
}

// NPC YAML structures

type sdeFactionYAML struct {
	Name          map[string]string `yaml:"name"`
	Description   map[string]string `yaml:"description"`
	CorporationID *int64            `yaml:"corporationID"`
	IconID        *int64            `yaml:"iconID"`
}

func parseFactions(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeFactionYAML](f)
	if err != nil {
		return err
	}

	factions := make([]models.SdeFaction, 0, len(raw))
	for id, fac := range raw {
		var desc *string
		if d := fac.Description["en"]; d != "" {
			desc = &d
		}
		factions = append(factions, models.SdeFaction{
			FactionID:     id,
			Name:          fac.Name["en"],
			Description:   desc,
			CorporationID: fac.CorporationID,
			IconID:        fac.IconID,
		})
	}
	data.Factions = factions
	return nil
}

type sdeNpcCorpYAML struct {
	Name      map[string]string `yaml:"name"`
	FactionID *int64            `yaml:"factionID"`
	IconID    *int64            `yaml:"iconID"`
}

func parseNpcCorporations(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeNpcCorpYAML](f)
	if err != nil {
		return err
	}

	corps := make([]models.SdeNpcCorporation, 0, len(raw))
	for id, c := range raw {
		corps = append(corps, models.SdeNpcCorporation{
			CorporationID: id,
			Name:          c.Name["en"],
			FactionID:     c.FactionID,
			IconID:        c.IconID,
		})
	}
	data.NpcCorporations = corps
	return nil
}

type sdeNpcCorpDivYAML struct {
	Divisions map[int64]sdeNpcDivNameYAML `yaml:"divisions"`
}

type sdeNpcDivNameYAML struct {
	Name map[string]string `yaml:"name"`
}

func parseNpcCorporationDivisions(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeNpcCorpDivYAML](f)
	if err != nil {
		return err
	}

	divs := make([]models.SdeNpcCorporationDivision, 0)
	for corpID, corp := range raw {
		for divID, div := range corp.Divisions {
			divs = append(divs, models.SdeNpcCorporationDivision{
				CorporationID: corpID,
				DivisionID:    divID,
				Name:          div.Name["en"],
			})
		}
	}
	data.NpcCorporationDivisions = divs
	return nil
}

type sdeAgentInSpaceYAML struct {
	SolarSystemID *int64 `yaml:"solarSystemID"`
}

func parseAgentsInSpace(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeAgentInSpaceYAML](f)
	if err != nil {
		return err
	}

	agents := make([]models.SdeAgentInSpace, 0, len(raw))
	for id, a := range raw {
		agents = append(agents, models.SdeAgentInSpace{
			AgentID:       id,
			SolarSystemID: a.SolarSystemID,
		})
	}
	data.AgentsInSpace = agents
	return nil
}

type sdeRaceYAML struct {
	Name        map[string]string `yaml:"name"`
	Description map[string]string `yaml:"description"`
	IconID      *int64            `yaml:"iconID"`
}

func parseRaces(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeRaceYAML](f)
	if err != nil {
		return err
	}

	races := make([]models.SdeRace, 0, len(raw))
	for id, r := range raw {
		var desc *string
		if d := r.Description["en"]; d != "" {
			desc = &d
		}
		races = append(races, models.SdeRace{
			RaceID:      id,
			Name:        r.Name["en"],
			Description: desc,
			IconID:      r.IconID,
		})
	}
	data.Races = races
	return nil
}

type sdeBloodlineYAML struct {
	Name        map[string]string `yaml:"name"`
	Description map[string]string `yaml:"description"`
	RaceID      *int64            `yaml:"raceID"`
	IconID      *int64            `yaml:"iconID"`
}

func parseBloodlines(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeBloodlineYAML](f)
	if err != nil {
		return err
	}

	bloodlines := make([]models.SdeBloodline, 0, len(raw))
	for id, b := range raw {
		var desc *string
		if d := b.Description["en"]; d != "" {
			desc = &d
		}
		bloodlines = append(bloodlines, models.SdeBloodline{
			BloodlineID: id,
			Name:        b.Name["en"],
			RaceID:      b.RaceID,
			Description: desc,
			IconID:      b.IconID,
		})
	}
	data.Bloodlines = bloodlines
	return nil
}

type sdeAncestryYAML struct {
	Name        map[string]string `yaml:"name"`
	Description map[string]string `yaml:"description"`
	BloodlineID *int64            `yaml:"bloodlineID"`
	IconID      *int64            `yaml:"iconID"`
}

func parseAncestries(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeAncestryYAML](f)
	if err != nil {
		return err
	}

	ancestries := make([]models.SdeAncestry, 0, len(raw))
	for id, a := range raw {
		var desc *string
		if d := a.Description["en"]; d != "" {
			desc = &d
		}
		ancestries = append(ancestries, models.SdeAncestry{
			AncestryID:  id,
			Name:        a.Name["en"],
			BloodlineID: a.BloodlineID,
			Description: desc,
			IconID:      a.IconID,
		})
	}
	data.Ancestries = ancestries
	return nil
}

// Industry YAML structures

type sdePlanetSchematicYAML struct {
	Name      map[string]string                        `yaml:"name"`
	CycleTime int                                       `yaml:"cycleTime"`
	Types     map[int64]sdePlanetSchematicTypeEntryYAML `yaml:"types"`
}

type sdePlanetSchematicTypeEntryYAML struct {
	Quantity int  `yaml:"quantity"`
	IsInput  bool `yaml:"isInput"`
}

func parsePlanetSchematics(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdePlanetSchematicYAML](f)
	if err != nil {
		return err
	}

	schematics := make([]models.SdePlanetSchematic, 0, len(raw))
	schematicTypes := make([]models.SdePlanetSchematicType, 0)

	for id, s := range raw {
		schematics = append(schematics, models.SdePlanetSchematic{
			SchematicID: id,
			Name:        s.Name["en"],
			CycleTime:   s.CycleTime,
		})

		for typeID, t := range s.Types {
			schematicTypes = append(schematicTypes, models.SdePlanetSchematicType{
				SchematicID: id,
				TypeID:      typeID,
				Quantity:    t.Quantity,
				IsInput:     t.IsInput,
			})
		}
	}

	data.PlanetSchematics = schematics
	data.PlanetSchematicTypes = schematicTypes
	return nil
}

type sdeControlTowerResourceYAML struct {
	Resources []sdeControlTowerResEntryYAML `yaml:"resources"`
}

type sdeControlTowerResEntryYAML struct {
	ResourceTypeID int64    `yaml:"resourceTypeID"`
	Purpose        *int     `yaml:"purpose"`
	Quantity       int      `yaml:"quantity"`
	MinSecurity    *float64 `yaml:"minSecurityLevel"`
	FactionID      *int64   `yaml:"factionID"`
}

func parseControlTowerResources(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeControlTowerResourceYAML](f)
	if err != nil {
		return err
	}

	resources := make([]models.SdeControlTowerResource, 0)
	for towerTypeID, tower := range raw {
		for _, res := range tower.Resources {
			resources = append(resources, models.SdeControlTowerResource{
				ControlTowerTypeID: towerTypeID,
				ResourceTypeID:     res.ResourceTypeID,
				Purpose:            res.Purpose,
				Quantity:           res.Quantity,
				MinSecurity:        res.MinSecurity,
				FactionID:          res.FactionID,
			})
		}
	}
	data.ControlTowerResources = resources
	return nil
}

// Misc YAML structures

type sdeSkinYAML struct {
	TypeID     *int64 `yaml:"typeID"`
	MaterialID *int64 `yaml:"skinMaterialID"`
}

func parseSkins(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeSkinYAML](f)
	if err != nil {
		return err
	}

	skins := make([]models.SdeSkin, 0, len(raw))
	for id, s := range raw {
		skins = append(skins, models.SdeSkin{
			SkinID:     id,
			TypeID:     s.TypeID,
			MaterialID: s.MaterialID,
		})
	}
	data.Skins = skins
	return nil
}

type sdeSkinLicenseYAML struct {
	SkinID   *int64 `yaml:"skinID"`
	Duration *int   `yaml:"duration"`
}

func parseSkinLicenses(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeSkinLicenseYAML](f)
	if err != nil {
		return err
	}

	licenses := make([]models.SdeSkinLicense, 0, len(raw))
	for id, l := range raw {
		licenses = append(licenses, models.SdeSkinLicense{
			LicenseTypeID: id,
			SkinID:        l.SkinID,
			Duration:      l.Duration,
		})
	}
	data.SkinLicenses = licenses
	return nil
}

type sdeSkinMaterialYAML struct {
	Name map[string]string `yaml:"name"`
}

func parseSkinMaterials(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeSkinMaterialYAML](f)
	if err != nil {
		return err
	}

	materials := make([]models.SdeSkinMaterial, 0, len(raw))
	for id, m := range raw {
		var name *string
		if n := m.Name["en"]; n != "" {
			name = &n
		}
		materials = append(materials, models.SdeSkinMaterial{
			SkinMaterialID: id,
			Name:           name,
		})
	}
	data.SkinMaterials = materials
	return nil
}

type sdeCertificateYAML struct {
	Name        map[string]string `yaml:"name"`
	Description map[string]string `yaml:"description"`
	GroupID     *int64            `yaml:"groupID"`
}

func parseCertificates(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeCertificateYAML](f)
	if err != nil {
		return err
	}

	certs := make([]models.SdeCertificate, 0, len(raw))
	for id, c := range raw {
		var name, desc *string
		if n := c.Name["en"]; n != "" {
			name = &n
		}
		if d := c.Description["en"]; d != "" {
			desc = &d
		}
		certs = append(certs, models.SdeCertificate{
			CertificateID: id,
			Name:          name,
			Description:   desc,
			GroupID:       c.GroupID,
		})
	}
	data.Certificates = certs
	return nil
}

type sdeLandmarkYAML struct {
	Name        map[string]string `yaml:"name"`
	Description map[string]string `yaml:"description"`
}

func parseLandmarks(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeLandmarkYAML](f)
	if err != nil {
		return err
	}

	landmarks := make([]models.SdeLandmark, 0, len(raw))
	for id, l := range raw {
		var name, desc *string
		if n := l.Name["en"]; n != "" {
			name = &n
		}
		if d := l.Description["en"]; d != "" {
			desc = &d
		}
		landmarks = append(landmarks, models.SdeLandmark{
			LandmarkID:  id,
			Name:        name,
			Description: desc,
		})
	}
	data.Landmarks = landmarks
	return nil
}

type sdeStationOpYAML struct {
	Name        map[string]string `yaml:"name"`
	Description map[string]string `yaml:"description"`
}

func parseStationOperations(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeStationOpYAML](f)
	if err != nil {
		return err
	}

	ops := make([]models.SdeStationOperation, 0, len(raw))
	for id, op := range raw {
		var name, desc *string
		if n := op.Name["en"]; n != "" {
			name = &n
		}
		if d := op.Description["en"]; d != "" {
			desc = &d
		}
		ops = append(ops, models.SdeStationOperation{
			OperationID: id,
			Name:        name,
			Description: desc,
		})
	}
	data.StationOperations = ops
	return nil
}

type sdeStationSvcYAML struct {
	Name        map[string]string `yaml:"name"`
	Description map[string]string `yaml:"description"`
}

func parseStationServices(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeStationSvcYAML](f)
	if err != nil {
		return err
	}

	svcs := make([]models.SdeStationService, 0, len(raw))
	for id, svc := range raw {
		var name, desc *string
		if n := svc.Name["en"]; n != "" {
			name = &n
		}
		if d := svc.Description["en"]; d != "" {
			desc = &d
		}
		svcs = append(svcs, models.SdeStationService{
			ServiceID:   id,
			Name:        name,
			Description: desc,
		})
	}
	data.StationServices = svcs
	return nil
}

type sdeContrabandTypeYAML struct {
	Factions map[int64]sdeContrabandEntryYAML `yaml:"factions"`
}

type sdeContrabandEntryYAML struct {
	StandingLoss *float64 `yaml:"standingLoss"`
	FineByValue  *float64 `yaml:"fineByValue"`
}

func parseContrabandTypes(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeContrabandTypeYAML](f)
	if err != nil {
		return err
	}

	types := make([]models.SdeContrabandType, 0)
	for typeID, ct := range raw {
		for factionID, entry := range ct.Factions {
			types = append(types, models.SdeContrabandType{
				FactionID:    factionID,
				TypeID:       typeID,
				StandingLoss: entry.StandingLoss,
				FineByValue:  entry.FineByValue,
			})
		}
	}
	data.ContrabandTypes = types
	return nil
}

type sdeCharAttrYAML struct {
	Name        localizedString `yaml:"name"`
	Description localizedString `yaml:"description"`
	IconID      *int64          `yaml:"iconID"`
}

func parseCharacterAttributes(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeCharAttrYAML](f)
	if err != nil {
		return err
	}

	attrs := make([]models.SdeCharacterAttribute, 0, len(raw))
	for id, a := range raw {
		var name, desc *string
		if a.Name.Value != "" {
			name = &a.Name.Value
		}
		if a.Description.Value != "" {
			desc = &a.Description.Value
		}
		attrs = append(attrs, models.SdeCharacterAttribute{
			AttributeID: id,
			Name:        name,
			Description: desc,
			IconID:      a.IconID,
		})
	}
	data.CharacterAttributes = attrs
	return nil
}

type sdeCorpActivityYAML struct {
	Name map[string]string `yaml:"name"`
}

func parseCorporationActivities(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeCorpActivityYAML](f)
	if err != nil {
		return err
	}

	acts := make([]models.SdeCorporationActivity, 0, len(raw))
	for id, a := range raw {
		var name *string
		if n := a.Name["en"]; n != "" {
			name = &n
		}
		acts = append(acts, models.SdeCorporationActivity{
			ActivityID: id,
			Name:       name,
		})
	}
	data.CorporationActivities = acts
	return nil
}

// Map YAML structures (formerly universe/)

type sdeRegionYAML struct {
	Name map[string]string `yaml:"name"`
}

func parseRegions(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeRegionYAML](f)
	if err != nil {
		return err
	}

	regions := make([]models.Region, 0, len(raw))
	for id, r := range raw {
		regions = append(regions, models.Region{
			ID:   id,
			Name: r.Name["en"],
		})
	}
	data.Regions = regions
	return nil
}

type sdeConstellationYAML struct {
	Name     map[string]string `yaml:"name"`
	RegionID int64             `yaml:"regionID"`
}

func parseConstellations(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeConstellationYAML](f)
	if err != nil {
		return err
	}

	constellations := make([]models.Constellation, 0, len(raw))
	for id, c := range raw {
		constellations = append(constellations, models.Constellation{
			ID:       id,
			Name:     c.Name["en"],
			RegionID: c.RegionID,
		})
	}
	data.Constellations = constellations
	return nil
}

type sdeSolarSystemPosition struct {
	X float64 `yaml:"x"`
	Y float64 `yaml:"y"`
	Z float64 `yaml:"z"`
}

type sdeSolarSystemYAML struct {
	Name            map[string]string      `yaml:"name"`
	ConstellationID int64                  `yaml:"constellationID"`
	Security        float64                `yaml:"securityStatus"`
	Position        sdeSolarSystemPosition `yaml:"position"`
}

func parseSolarSystems(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeSolarSystemYAML](f)
	if err != nil {
		return err
	}

	systems := make([]models.SolarSystem, 0, len(raw))
	for id, s := range raw {
		x := s.Position.X
		y := s.Position.Y
		z := s.Position.Z
		systems = append(systems, models.SolarSystem{
			ID:              id,
			Name:            s.Name["en"],
			ConstellationID: s.ConstellationID,
			Security:        s.Security,
			X:               &x,
			Y:               &y,
			Z:               &z,
		})
	}
	data.SolarSystems = systems
	return nil
}

type sdeStationYAML struct {
	StationName   localizedString `yaml:"stationName"`
	SolarSystemID int64           `yaml:"solarSystemID"`
	CorporationID int64           `yaml:"ownerID"`
}

func parseStations(f *zip.File, data *SdeData) error {
	raw, err := parseYAMLMap[sdeStationYAML](f)
	if err != nil {
		return err
	}

	stations := make([]models.Station, 0, len(raw))
	for id, s := range raw {
		stations = append(stations, models.Station{
			ID:            id,
			Name:          s.StationName.Value,
			SolarSystemID: s.SolarSystemID,
			CorporationID: s.CorporationID,
			IsNPC:         true,
		})
	}
	data.Stations = stations
	return nil
}
