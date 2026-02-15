package repositories

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

type Asset struct {
	Name            string   `json:"name"`
	TypeID          int64    `json:"typeId"`
	Quantity        int64    `json:"quantity"`
	Volume          float64  `json:"volume"`
	OwnerType       string   `json:"ownerType"`
	OwnerName       string   `json:"ownerName"`
	OwnerID         int64    `json:"ownerId"`
	DesiredQuantity *int64   `json:"desiredQuantity"`
	StockpileDelta  *int64   `json:"stockpileDelta"`
	UnitPrice       *float64 `json:"unitPrice"`
	TotalValue      *float64 `json:"totalValue"`
	DeficitValue    *float64 `json:"deficitValue"`
}

type AssetsResponse struct {
	Structures []*AssetStructure `json:"structures"`
}

type AssetStructure struct {
	ID                 int64                 `json:"id"`
	Name               string                `json:"name"`
	SolarSystem        string                `json:"solarSystem"`
	Region             string                `json:"region"`
	HangarAssets       []*Asset              `json:"hangarAssets"`
	HangarContainers   []*AssetContainer     `json:"hangarContainers"`
	Deliveries         []*Asset              `json:"deliveries"`
	AssetSafety        []*Asset              `json:"assetSafety"`
	CorporationHangers []*CorporationHanger  `json:"corporationHangers"`
}

type CorporationHanger struct {
	ID               int64             `json:"id"`
	Name             string            `json:"name"`
	CorporationID    int64             `json:"corporationId"`
	CorporationName  string            `json:"corporationName"`
	Assets           []*Asset          `json:"assets"`
	HangarContainers []*AssetContainer `json:"hangarContainers"`
}

type AssetContainer struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	OwnerType string   `json:"ownerType"`
	OwnerName string   `json:"ownerName"`
	OwnerID   int64    `json:"ownerId"`
	Assets    []*Asset `json:"assets"`
}

type AssetsSummary struct {
	TotalValue   float64 `json:"totalValue"`
	TotalDeficit float64 `json:"totalDeficit"`
}

type Assets struct {
	db *sql.DB
}

func NewAssets(db *sql.DB) *Assets {
	return &Assets{
		db: db,
	}
}

func (r *Assets) GetUserAssets(ctx context.Context, user int64) (*AssetsResponse, error) {
	response := &AssetsResponse{}

	stationsQuery := `
SELECT distinct
    characterAssets.location_id,
    stations.name as "station_name",
    systems.name as "solar_system_name",
    regions.name as "region_name"
FROM
    character_assets characterAssets
INNER JOIN
    asset_item_types assetTypes
ON
    assetTypes.type_id=characterAssets.type_id
INNER JOIN
    stations stations
ON
    characterAssets.location_id=stations.station_id
INNER JOIN
    solar_systems systems
ON
    stations.solar_system_id=systems.solar_system_id
INNER JOIN
    constellations constellations
ON
    systems.constellation_id=constellations.constellation_id
INNER JOIN
    regions regions
ON
    constellations.region_id=regions.region_id
WHERE
    user_id=$1 AND
    (location_type='station' OR (location_flag='Hangar' and location_type='item'));`

	stations, err := r.db.QueryContext(ctx, stationsQuery, user)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query stations from database")
	}
	defer stations.Close()

	stationMap := map[int64]*AssetStructure{}
	for stations.Next() {
		structure := &AssetStructure{}
		err = stations.Scan(&structure.ID, &structure.Name, &structure.SolarSystem, &structure.Region)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan item")
		}
		response.Structures = append(response.Structures, structure)

		structure.Deliveries = []*Asset{}
		structure.HangarAssets = []*Asset{}
		structure.HangarContainers = []*AssetContainer{}
		structure.AssetSafety = []*Asset{}
		structure.CorporationHangers = []*CorporationHanger{}

		stationMap[structure.ID] = structure
	}

	hangaredItemsQuery := `
SELECT
	characterAssets.character_id,
	characters.name,
	characterAssets.location_id,
	characterAssets.location_flag,
    assetTypes.type_id,
    assetTypes.type_name,
    characterAssets.quantity,
    assetTypes.volume * characterAssets.quantity as "volume",
    stockpile.desired_quantity,
    (characterAssets.quantity - COALESCE(stockpile.desired_quantity, 0)) as stockpile_delta,
    market.sell_price as unit_price,
    (characterAssets.quantity * COALESCE(market.sell_price, 0)) as total_value,
    CASE
        WHEN (characterAssets.quantity - COALESCE(stockpile.desired_quantity, 0)) < 0
        THEN ABS(characterAssets.quantity - COALESCE(stockpile.desired_quantity, 0)) * COALESCE(market.buy_price, 0)
        ELSE 0
    END as deficit_value
FROM
    character_assets characterAssets
INNER JOIN
    asset_item_types assetTypes
ON
    assetTypes.type_id=characterAssets.type_id
INNER JOIN
	characters characters
ON
	characters.id=characterAssets.character_id
INNER JOIN
    stations stations
ON
    characterAssets.location_id=stations.station_id
INNER JOIN
    solar_systems systems
ON
    stations.solar_system_id=systems.solar_system_id
LEFT JOIN
    stockpile_markers stockpile
ON
    stockpile.user_id = $1
    AND stockpile.type_id = characterAssets.type_id
    AND stockpile.owner_type = 'character'
    AND stockpile.owner_id = characterAssets.character_id
    AND stockpile.location_id = characterAssets.location_id
    AND stockpile.container_id IS NULL
    AND stockpile.division_number IS NULL
LEFT JOIN
    market_prices market
ON
    market.type_id = characterAssets.type_id
    AND market.region_id = 10000002
WHERE
    characterAssets.user_id=$1
    AND NOT (is_singleton=true AND assetTypes.type_name like '%Container')
	AND NOT location_flag='AssetSafety'
    AND (
        location_type='station'
        OR (location_flag='Hangar' and location_type='item')
        OR (location_flag='Deliveries' and location_type='item')
    );`

	items, err := r.db.QueryContext(ctx, hangaredItemsQuery, user)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query hangared assets from database")
	}
	defer items.Close()

	for items.Next() {
		asset := &Asset{}
		var location int64
		var locationFlag string

		asset.OwnerType = "character"

		err = items.Scan(&asset.OwnerID, &asset.OwnerName, &location, &locationFlag, &asset.TypeID, &asset.Name, &asset.Quantity, &asset.Volume, &asset.DesiredQuantity, &asset.StockpileDelta, &asset.UnitPrice, &asset.TotalValue, &asset.DeficitValue)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan item")
		}

		station, ok := stationMap[location]
		if !ok {
			return nil, errors.Errorf("location %d not found for hanger asset", location)
		}

		switch locationFlag {
		case "Hangar":
			station.HangarAssets = append(station.HangarAssets, asset)
		case "Deliveries":
			station.Deliveries = append(station.Deliveries, asset)
		default:
			return nil, errors.Errorf("unknown location flag %s", locationFlag)
		}
	}

	// containers
	containerQuery := `
SELECT
	characterAssets.character_id,
	characters.name,
	characterAssets.item_id,
    assetTypes.type_name,
    characterAssets.location_id,
    locations.name
FROM
    character_assets characterAssets
INNER JOIN
    asset_item_types assetTypes
ON
    assetTypes.type_id=characterAssets.type_id
INNER JOIN
	characters characters
ON
	characters.id=characterAssets.character_id
INNER JOIN
    character_asset_location_names locations
ON
    locations.item_id=characterAssets.item_id
WHERE
    characterAssets.user_id=$1
    AND (is_singleton=true AND assetTypes.type_name like '%Container')
ORDER BY
    characterAssets.item_id;`

	containers, err := r.db.QueryContext(ctx, containerQuery, user)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query containers from database")
	}
	defer containers.Close()

	containerMap := map[int64]*AssetContainer{}
	for containers.Next() {
		container := &AssetContainer{}
		var location int64
		var defaultName string

		container.OwnerType = "character"

		err = containers.Scan(&container.OwnerID, &container.OwnerName, &container.ID, &defaultName, &location, &container.Name)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan container")
		}

		station, ok := stationMap[location]
		if !ok {
			continue
			//return nil, errors.Errorf("location %d not found for station container", location)
		}

		station.HangarContainers = append(station.HangarContainers, container)
		containerMap[container.ID] = container
		container.Assets = []*Asset{}
	}

	// shit in containers
	itemsInContainersQuery := `
SELECT
	characterAssets.character_id,
	characters.name,
    assetTypes.type_id,
    assetTypes.type_name,
    characterAssets.quantity,
    assetTypes.volume * characterAssets.quantity as "volume",
    characterAssets.location_id,
    stockpile.desired_quantity,
    (characterAssets.quantity - COALESCE(stockpile.desired_quantity, 0)) as stockpile_delta,
    market.sell_price as unit_price,
    (characterAssets.quantity * COALESCE(market.sell_price, 0)) as total_value,
    CASE
        WHEN (characterAssets.quantity - COALESCE(stockpile.desired_quantity, 0)) < 0
        THEN ABS(characterAssets.quantity - COALESCE(stockpile.desired_quantity, 0)) * COALESCE(market.buy_price, 0)
        ELSE 0
    END as deficit_value
FROM
    character_assets characterAssets
INNER JOIN
    asset_item_types assetTypes
ON
    assetTypes.type_id=characterAssets.type_id
INNER JOIN
	characters characters
ON
	characters.id=characterAssets.character_id
INNER JOIN
    character_asset_location_names locations
ON
    locations.item_id=characterAssets.location_id
LEFT JOIN
    stockpile_markers stockpile
ON
    stockpile.user_id = $1
    AND stockpile.type_id = characterAssets.type_id
    AND stockpile.owner_type = 'character'
    AND stockpile.owner_id = characterAssets.character_id
    AND stockpile.container_id = characterAssets.location_id
    AND stockpile.division_number IS NULL
LEFT JOIN
    market_prices market
ON
    market.type_id = characterAssets.type_id
    AND market.region_id = 10000002
WHERE
    characterAssets.user_id=$1
    AND characterAssets.location_type='item'
    AND NOT (characterAssets.is_singleton=true AND assetTypes.type_name like '%Container')
ORDER BY
    characterAssets.item_id;`

	itemsInContainers, err := r.db.QueryContext(ctx, itemsInContainersQuery, user)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query items in containers items from database")
	}
	defer itemsInContainers.Close()

	for itemsInContainers.Next() {
		asset := &Asset{}
		var location int64

		asset.OwnerType = "character"

		err = itemsInContainers.Scan(&asset.OwnerID, &asset.OwnerName, &asset.TypeID, &asset.Name, &asset.Quantity, &asset.Volume, &location, &asset.DesiredQuantity, &asset.StockpileDelta, &asset.UnitPrice, &asset.TotalValue, &asset.DeficitValue)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan container")
		}

		container, ok := containerMap[location]
		if !ok {
			continue
			//return nil, errors.Errorf("location %d not found for item container", location)
		}

		container.Assets = append(container.Assets, asset)
	}

	corpStationsQuery := `
SELECT distinct
    corporation_assets.location_id,
    corporation_assets.corporation_id,
    stations.name as "station_name",
    systems.name as "solar_system_name",
    regions.name as "region_name"
FROM
    corporation_assets corporation_assets
INNER JOIN
    asset_item_types assetTypes
ON
    assetTypes.type_id=corporation_assets.type_id
INNER JOIN
    stations stations
ON
    corporation_assets.location_id=stations.station_id
INNER JOIN
    solar_systems systems
ON
    stations.solar_system_id=systems.solar_system_id
INNER JOIN
    constellations constellations
ON
    systems.constellation_id=constellations.constellation_id
INNER JOIN
    regions regions
ON
    constellations.region_id=regions.region_id
WHERE
    corporation_assets.user_id=$1 AND
    (location_flag='OfficeFolder' OR location_flag like 'CorpSAG%');`

	stations, err = r.db.QueryContext(ctx, corpStationsQuery, user)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query corp stations from database")
	}
	defer stations.Close()

	stationCorpMap := map[int64]map[int64]bool{}
	for stations.Next() {
		var stationID int64
		var corpID int64
		var stationName string
		var solarSystem string
		var region string
		err = stations.Scan(&stationID, &corpID, &stationName, &solarSystem, &region)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan item")
		}

		_, ok := stationMap[stationID]
		if !ok {
			structure := &AssetStructure{
				ID:                 stationID,
				Name:               stationName,
				SolarSystem:        solarSystem,
				Region:             region,
				Deliveries:         []*Asset{},
				HangarAssets:       []*Asset{},
				HangarContainers:   []*AssetContainer{},
				AssetSafety:        []*Asset{},
				CorporationHangers: []*CorporationHanger{},
			}
			response.Structures = append(response.Structures, structure)
			stationMap[stationID] = structure
		}

		if stationCorpMap[stationID] == nil {
			stationCorpMap[stationID] = map[int64]bool{}
		}
		stationCorpMap[stationID][corpID] = true
	}

	corpDivisionsQuery := `
SELECT
	corporation_divisions.division_number,
	corporation_divisions.corporation_id,
	player_corporations.name,
	corporation_divisions.name
FROM
	corporation_divisions corporation_divisions
INNER JOIN
	player_corporations player_corporations
ON
	corporation_divisions.corporation_id=player_corporations.id
WHERE
	corporation_divisions.user_id=$1 AND
	corporation_divisions.division_type='hangar'
ORDER BY
	corporation_divisions.corporation_id,
	corporation_divisions.division_number;`

	corpDivisions, err := r.db.QueryContext(ctx, corpDivisionsQuery, user)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query corp divisions from database")
	}
	defer corpDivisions.Close()

	// Build a template map of divisions for each corp
	// We'll create station-specific copies as needed
	corpDivisionTemplates := map[int64]map[int64]*CorporationHanger{}
	for corpDivisions.Next() {
		hanger := &CorporationHanger{}

		err = corpDivisions.Scan(&hanger.ID, &hanger.CorporationID, &hanger.CorporationName, &hanger.Name)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan corp division")
		}

		if corpDivisionTemplates[hanger.CorporationID] == nil {
			corpDivisionTemplates[hanger.CorporationID] = map[int64]*CorporationHanger{}
		}
		corpDivisionTemplates[hanger.CorporationID][hanger.ID] = hanger
	}

	// Station-specific division map: stationID -> corpID -> divisionID -> division
	corpHangerMap := map[int64]map[int64]map[int64]*CorporationHanger{}

	// Helper function to get or create a station-specific division
	getOrCreateDivision := func(stationID, corpID, divisionID int64) *CorporationHanger {
		if corpHangerMap[stationID] == nil {
			corpHangerMap[stationID] = map[int64]map[int64]*CorporationHanger{}
		}
		if corpHangerMap[stationID][corpID] == nil {
			corpHangerMap[stationID][corpID] = map[int64]*CorporationHanger{}
		}

		// If division doesn't exist at this station yet, create it from template
		if corpHangerMap[stationID][corpID][divisionID] == nil {
			template, ok := corpDivisionTemplates[corpID][divisionID]
			if !ok {
				return nil
			}
			corpHangerMap[stationID][corpID][divisionID] = &CorporationHanger{
				ID:               template.ID,
				Name:             template.Name,
				CorporationID:    template.CorporationID,
				CorporationName:  template.CorporationName,
				Assets:           []*Asset{},
				HangarContainers: []*AssetContainer{},
			}
		}
		return corpHangerMap[stationID][corpID][divisionID]
	}

	corpHangaredItemsQuery := `
SELECT
	corporation_assets.corporation_id,
	player_corporations.name,
	corporation_assets.location_id,
	SUBSTRING(corporation_assets.location_flag, 8, 1)::int as "division_number",
	assetTypes.type_id,
	assetTypes.type_name,
	corporation_assets.quantity,
	assetTypes.volume * corporation_assets.quantity as "volume",
	stockpile.desired_quantity,
	(corporation_assets.quantity - COALESCE(stockpile.desired_quantity, 0)) as stockpile_delta,
	market.sell_price as unit_price,
	(corporation_assets.quantity * COALESCE(market.sell_price, 0)) as total_value,
	CASE
		WHEN (corporation_assets.quantity - COALESCE(stockpile.desired_quantity, 0)) < 0
		THEN ABS(corporation_assets.quantity - COALESCE(stockpile.desired_quantity, 0)) * COALESCE(market.buy_price, 0)
		ELSE 0
	END as deficit_value
FROM
	corporation_assets corporation_assets
INNER JOIN
	asset_item_types assetTypes
ON
	assetTypes.type_id=corporation_assets.type_id
INNER JOIN
	player_corporations player_corporations
ON
	player_corporations.id=corporation_assets.corporation_id
LEFT JOIN
	stockpile_markers stockpile
ON
	stockpile.user_id = $1
	AND stockpile.type_id = corporation_assets.type_id
	AND stockpile.owner_type = 'corporation'
	AND stockpile.owner_id = corporation_assets.corporation_id
	AND stockpile.location_id = corporation_assets.location_id
	AND stockpile.division_number = SUBSTRING(corporation_assets.location_flag, 8, 1)::int
	AND stockpile.container_id IS NULL
LEFT JOIN
	market_prices market
ON
	market.type_id = corporation_assets.type_id
	AND market.region_id = 10000002
WHERE
	corporation_assets.user_id=$1
	AND NOT (is_singleton=true AND assetTypes.type_name like '%Container')
	AND location_type='station'
	AND location_flag like 'CorpSAG%';`

	corpHangaredItems, err := r.db.QueryContext(ctx, corpHangaredItemsQuery, user)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query corp hangared assets from database")
	}
	defer corpHangaredItems.Close()

	for corpHangaredItems.Next() {
		asset := &Asset{}
		var location int64
		var divisionNumber int64

		asset.OwnerType = "corporation"

		err = corpHangaredItems.Scan(&asset.OwnerID, &asset.OwnerName, &location, &divisionNumber, &asset.TypeID, &asset.Name, &asset.Quantity, &asset.Volume, &asset.DesiredQuantity, &asset.StockpileDelta, &asset.UnitPrice, &asset.TotalValue, &asset.DeficitValue)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan corp hangared item")
		}

		// Get or create station-specific division
		hanger := getOrCreateDivision(location, asset.OwnerID, divisionNumber)
		if hanger == nil {
			continue
		}

		// Track that this corp has presence at this station
		if stationCorpMap[location] == nil {
			stationCorpMap[location] = map[int64]bool{}
		}
		stationCorpMap[location][asset.OwnerID] = true

		hanger.Assets = append(hanger.Assets, asset)
	}

	corpContainerQuery := `
WITH RECURSIVE container_chain AS (
	-- Start with all corp containers in divisions
	SELECT
		ca.item_id,
		ca.location_id,
		ca.location_type,
		ca.corporation_id,
		ca.location_flag,
		ca.type_id,
		ca.is_singleton,
		ca.item_id as container_id
	FROM corporation_assets ca
	WHERE ca.user_id=$1
	  AND ca.is_singleton=true
	  AND ca.location_flag LIKE 'CorpSAG%'

	UNION

	-- Recursively find parent locations until we hit a station
	SELECT
		ca.item_id,
		ca.location_id,
		ca.location_type,
		ca.corporation_id,
		ca.location_flag,
		ca.type_id,
		ca.is_singleton,
		cc.container_id
	FROM corporation_assets ca
	INNER JOIN container_chain cc ON ca.item_id = cc.location_id
	WHERE ca.user_id=$1
	  AND cc.location_type != 'station'
)
SELECT
	cc.corporation_id,
	pc.name as corp_name,
	cc.container_id as item_id,
	ait.type_name,
	-- Get station_id: find the location_id of the deepest parent in the chain
	-- (the one where no other row in the chain has item_id = this row's location_id)
	(SELECT c1.location_id
	 FROM container_chain c1
	 WHERE c1.container_id = cc.container_id
	   AND NOT EXISTS (
		 SELECT 1 FROM container_chain c2
		 WHERE c2.container_id = cc.container_id AND c2.item_id = c1.location_id
	   )
	 LIMIT 1) as station_id,
	CASE
		WHEN EXISTS (SELECT 1 FROM container_chain WHERE container_id = cc.container_id AND location_type = 'station')
		THEN 'station'
		ELSE 'item'
	END as final_location_type,
	SUBSTRING((SELECT location_flag FROM container_chain WHERE item_id = cc.container_id LIMIT 1), 8, 1)::int as division_number,
	loc.name as container_name
FROM (SELECT DISTINCT container_id, corporation_id FROM container_chain) cc
INNER JOIN asset_item_types ait ON ait.type_id = (
	SELECT type_id FROM container_chain WHERE item_id = cc.container_id LIMIT 1
)
INNER JOIN player_corporations pc ON pc.id = cc.corporation_id
INNER JOIN corporation_asset_location_names loc ON loc.item_id = cc.container_id
WHERE ait.type_name LIKE '%Container'
ORDER BY cc.container_id;`

	corpContainers, err := r.db.QueryContext(ctx, corpContainerQuery, user)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query corp containers from database")
	}
	defer corpContainers.Close()

	corpContainerMap := map[int64]*AssetContainer{}
	// Map to store containers by corp and division (for nested containers)
	corpContainersByDivision := map[int64]map[int64][]*AssetContainer{} // corpID -> divisionID -> containers

	for corpContainers.Next() {
		container := &AssetContainer{}
		var location sql.NullInt64
		var locationType string
		var divisionNumber int64
		var defaultName string

		container.OwnerType = "corporation"

		err = corpContainers.Scan(&container.OwnerID, &container.OwnerName, &container.ID, &defaultName, &location, &locationType, &divisionNumber, &container.Name)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan corp container")
		}

		container.Assets = []*Asset{}
		corpContainerMap[container.ID] = container

		// If we successfully determined a station_id, add container directly to that station
		if location.Valid {
			stationID := location.Int64
			hanger := getOrCreateDivision(stationID, container.OwnerID, divisionNumber)
			if hanger != nil {
				// Track that this corp has presence at this station
				if stationCorpMap[stationID] == nil {
					stationCorpMap[stationID] = map[int64]bool{}
				}
				stationCorpMap[stationID][container.OwnerID] = true

				hanger.HangarContainers = append(hanger.HangarContainers, container)
			}
		} else {
			// Can't determine station (very rare) - store for fallback assignment
			if corpContainersByDivision[container.OwnerID] == nil {
				corpContainersByDivision[container.OwnerID] = map[int64][]*AssetContainer{}
			}
			corpContainersByDivision[container.OwnerID][divisionNumber] = append(
				corpContainersByDivision[container.OwnerID][divisionNumber],
				container,
			)
		}
	}

	corpItemsInContainersQuery := `
SELECT
	corporation_assets.corporation_id,
	player_corporations.name,
	assetTypes.type_id,
	assetTypes.type_name,
	corporation_assets.quantity,
	assetTypes.volume * corporation_assets.quantity as "volume",
	corporation_assets.location_id,
	stockpile.desired_quantity,
	(corporation_assets.quantity - COALESCE(stockpile.desired_quantity, 0)) as stockpile_delta,
	market.sell_price as unit_price,
	(corporation_assets.quantity * COALESCE(market.sell_price, 0)) as total_value,
	CASE
		WHEN (corporation_assets.quantity - COALESCE(stockpile.desired_quantity, 0)) < 0
		THEN ABS(corporation_assets.quantity - COALESCE(stockpile.desired_quantity, 0)) * COALESCE(market.buy_price, 0)
		ELSE 0
	END as deficit_value
FROM
	corporation_assets corporation_assets
INNER JOIN
	asset_item_types assetTypes
ON
	assetTypes.type_id=corporation_assets.type_id
INNER JOIN
	player_corporations player_corporations
ON
	player_corporations.id=corporation_assets.corporation_id
LEFT JOIN
	corporation_asset_location_names locations
ON
	locations.item_id=corporation_assets.location_id
	AND locations.corporation_id=corporation_assets.corporation_id
	AND locations.user_id=corporation_assets.user_id
LEFT JOIN
	stockpile_markers stockpile
ON
	stockpile.user_id = $1
	AND stockpile.type_id = corporation_assets.type_id
	AND stockpile.owner_type = 'corporation'
	AND stockpile.owner_id = corporation_assets.corporation_id
	AND stockpile.container_id = corporation_assets.location_id
LEFT JOIN
	market_prices market
ON
	market.type_id = corporation_assets.type_id
	AND market.region_id = 10000002
WHERE
	corporation_assets.user_id=$1
	AND corporation_assets.location_type='item'
	AND NOT (corporation_assets.is_singleton=true AND assetTypes.type_name like '%Container')
ORDER BY
	corporation_assets.item_id;`

	corpItemsInContainers, err := r.db.QueryContext(ctx, corpItemsInContainersQuery, user)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query corp items in containers from database")
	}
	defer corpItemsInContainers.Close()

	for corpItemsInContainers.Next() {
		asset := &Asset{}
		var location int64

		asset.OwnerType = "corporation"

		err = corpItemsInContainers.Scan(&asset.OwnerID, &asset.OwnerName, &asset.TypeID, &asset.Name, &asset.Quantity, &asset.Volume, &location, &asset.DesiredQuantity, &asset.StockpileDelta, &asset.UnitPrice, &asset.TotalValue, &asset.DeficitValue)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan corp container item")
		}

		container, ok := corpContainerMap[location]
		if !ok {
			continue
		}

		container.Assets = append(container.Assets, asset)
	}

	// Add divisions to stations
	// If a corp has ANY assets at a station, show ALL its divisions (even empty ones)
	for stationID, corpIDs := range stationCorpMap {
		station, ok := stationMap[stationID]
		if !ok {
			continue
		}

		for corpID := range corpIDs {
			// Get all division templates for this corp
			divisionTemplates, ok := corpDivisionTemplates[corpID]
			if !ok {
				continue
			}

			// For each defined division
			for divisionID, template := range divisionTemplates {
				// Check if we already created this division at this station (has content)
				var division *CorporationHanger
				if corpHangerMap[stationID] != nil && corpHangerMap[stationID][corpID] != nil {
					division = corpHangerMap[stationID][corpID][divisionID]
				}

				// If not created yet, create an empty one
				if division == nil {
					division = &CorporationHanger{
						ID:               template.ID,
						Name:             template.Name,
						CorporationID:    template.CorporationID,
						CorporationName:  template.CorporationName,
						Assets:           []*Asset{},
						HangarContainers: []*AssetContainer{},
					}
				}

				station.CorporationHangers = append(station.CorporationHangers, division)
			}
		}
	}

	// Handle corps that have containers but no presence in stationCorpMap
	// Add their divisions to the first available station
	for corpID, divisions := range corpContainersByDivision {
		// Check if this corp already has presence at any station
		hasPresence := false
		for _, corpMap := range stationCorpMap {
			if corpMap[corpID] {
				hasPresence = true
				break
			}
		}

		// If no presence yet, add divisions to first station
		if !hasPresence && len(response.Structures) > 0 {
			firstStation := response.Structures[0]

			// Get division templates for this corp
			divisionTemplates, ok := corpDivisionTemplates[corpID]
			if ok {
				for divisionID, template := range divisionTemplates {
					division := &CorporationHanger{
						ID:               template.ID,
						Name:             template.Name,
						CorporationID:    template.CorporationID,
						CorporationName:  template.CorporationName,
						Assets:           []*Asset{},
						HangarContainers: divisions[divisionID], // Add containers for this division
					}
					firstStation.CorporationHangers = append(firstStation.CorporationHangers, division)
				}
			}
		}
	}

	return response, nil
}

type StockpileItem struct {
	Name            string   `json:"name"`
	TypeID          int64    `json:"typeId"`
	Quantity        int64    `json:"quantity"`
	Volume          float64  `json:"volume"`
	OwnerType       string   `json:"ownerType"`
	OwnerName       string   `json:"ownerName"`
	OwnerID         int64    `json:"ownerId"`
	DesiredQuantity int64    `json:"desiredQuantity"`
	StockpileDelta  int64    `json:"stockpileDelta"`
	DeficitValue    float64  `json:"deficitValue"`
	StructureName   string   `json:"structureName"`
	SolarSystem     string   `json:"solarSystem"`
	Region          string   `json:"region"`
	ContainerName   *string  `json:"containerName"`
}

type StockpilesResponse struct {
	Items []*StockpileItem `json:"items"`
}

func (r *Assets) GetStockpileDeficits(ctx context.Context, user int64) (*StockpilesResponse, error) {
	response := &StockpilesResponse{
		Items: []*StockpileItem{},
	}

	// Query for all assets with stockpile deficit (stockpile_delta < 0)
	// This combines personal and corporation assets in a single query
	query := `
		WITH all_deficits AS (
			-- Personal hangar items
			SELECT
				assetTypes.type_name as name,
				characterAssets.type_id,
				characterAssets.quantity,
				(characterAssets.quantity * assetTypes.volume) as volume,
				'character' as owner_type,
				characters.name as owner_name,
				characters.id as owner_id,
				stockpile.desired_quantity,
				(characterAssets.quantity - COALESCE(stockpile.desired_quantity, 0)) as stockpile_delta,
				ABS(characterAssets.quantity - COALESCE(stockpile.desired_quantity, 0)) * COALESCE(market.buy_price, 0) as deficit_value,
				stations.name as structure_name,
				systems.name as solar_system,
				regions.name as region,
				NULL::text as container_name
			FROM character_assets characterAssets
			INNER JOIN characters ON characters.id = characterAssets.character_id
			INNER JOIN asset_item_types assetTypes ON assetTypes.type_id = characterAssets.type_id
			INNER JOIN stations ON characterAssets.location_id = stations.station_id
			INNER JOIN solar_systems systems ON stations.solar_system_id = systems.solar_system_id
			INNER JOIN constellations ON systems.constellation_id = constellations.constellation_id
			INNER JOIN regions ON constellations.region_id = regions.region_id
			LEFT JOIN stockpile_markers stockpile ON (
				stockpile.type_id = characterAssets.type_id
				AND stockpile.location_id = characterAssets.location_id
				AND stockpile.container_id IS NULL
				AND stockpile.owner_id = characterAssets.character_id
			)
			LEFT JOIN market_prices market ON (market.type_id = characterAssets.type_id AND market.region_id = 10000002)
			WHERE characterAssets.user_id = $1
				AND characterAssets.location_type = 'station'
				AND characterAssets.location_flag IN ('Hangar', 'Deliveries', 'AssetSafety')
				AND (characterAssets.quantity - COALESCE(stockpile.desired_quantity, 0)) < 0

			UNION ALL

			-- Personal container items
			SELECT
				assetTypes.type_name as name,
				characterAssets.type_id,
				characterAssets.quantity,
				(characterAssets.quantity * assetTypes.volume) as volume,
				'character' as owner_type,
				characters.name as owner_name,
				characters.id as owner_id,
				stockpile.desired_quantity,
				(characterAssets.quantity - COALESCE(stockpile.desired_quantity, 0)) as stockpile_delta,
				ABS(characterAssets.quantity - COALESCE(stockpile.desired_quantity, 0)) * COALESCE(market.buy_price, 0) as deficit_value,
				stations.name as structure_name,
				systems.name as solar_system,
				regions.name as region,
				containerTypes.type_name as container_name
			FROM character_assets characterAssets
			INNER JOIN characters ON characters.id = characterAssets.character_id
			INNER JOIN asset_item_types assetTypes ON assetTypes.type_id = characterAssets.type_id
			INNER JOIN character_assets containers ON containers.item_id = characterAssets.location_id
			INNER JOIN asset_item_types containerTypes ON containerTypes.type_id = containers.type_id
			INNER JOIN stations ON containers.location_id = stations.station_id
			INNER JOIN solar_systems systems ON stations.solar_system_id = systems.solar_system_id
			INNER JOIN constellations ON systems.constellation_id = constellations.constellation_id
			INNER JOIN regions ON constellations.region_id = regions.region_id
			LEFT JOIN stockpile_markers stockpile ON (
				stockpile.type_id = characterAssets.type_id
				AND stockpile.container_id = characterAssets.location_id
				AND stockpile.owner_id = characterAssets.character_id
			)
			LEFT JOIN market_prices market ON (market.type_id = characterAssets.type_id AND market.region_id = 10000002)
			WHERE characterAssets.user_id = $1
				AND characterAssets.location_type = 'item'
				AND NOT (characterAssets.is_singleton = true AND assetTypes.type_name LIKE '%Container')
				AND (characterAssets.quantity - COALESCE(stockpile.desired_quantity, 0)) < 0

			UNION ALL

			-- Corporation hangar items (using view for location resolution)
			SELECT
				assetTypes.type_name as name,
				loc.type_id,
				ca.quantity,
				(ca.quantity * assetTypes.volume) as volume,
				'corporation' as owner_type,
				corps.name as owner_name,
				corps.id as owner_id,
				stockpile.desired_quantity,
				(ca.quantity - COALESCE(stockpile.desired_quantity, 0)) as stockpile_delta,
				ABS(ca.quantity - COALESCE(stockpile.desired_quantity, 0)) * COALESCE(market.buy_price, 0) as deficit_value,
				loc.station_name as structure_name,
				loc.solar_system_name as solar_system,
				loc.region_name as region,
				COALESCE(divisions.name, loc.location_flag) as container_name
			FROM corporation_asset_locations loc
			INNER JOIN corporation_assets ca ON (
				ca.item_id = loc.item_id
				AND ca.corporation_id = loc.corporation_id
				AND ca.user_id = loc.user_id
			)
			INNER JOIN player_corporations corps ON corps.id = loc.corporation_id
			INNER JOIN asset_item_types assetTypes ON assetTypes.type_id = loc.type_id
			LEFT JOIN corporation_divisions divisions ON (
				divisions.division_number = loc.division_number
				AND divisions.corporation_id = loc.corporation_id
				AND divisions.user_id = loc.user_id
				AND divisions.division_type = 'hangar'
			)
			LEFT JOIN stockpile_markers stockpile ON (
				stockpile.type_id = loc.type_id
				AND stockpile.location_id = loc.location_id
				AND stockpile.division_number = loc.division_number
				AND stockpile.container_id IS NULL
				AND stockpile.owner_id = loc.corporation_id
			)
			LEFT JOIN market_prices market ON (market.type_id = loc.type_id AND market.region_id = 10000002)
			WHERE loc.user_id = $1
				AND loc.location_type = 'station'
				AND loc.location_flag LIKE 'CorpSAG%'
				AND loc.station_id IS NOT NULL
				AND (ca.quantity - COALESCE(stockpile.desired_quantity, 0)) < 0

			UNION ALL

			-- Corporation container items (using view for location resolution)
			SELECT
				assetTypes.type_name as name,
				loc.type_id,
				ca.quantity,
				(ca.quantity * assetTypes.volume) as volume,
				'corporation' as owner_type,
				corps.name as owner_name,
				corps.id as owner_id,
				stockpile.desired_quantity,
				(ca.quantity - COALESCE(stockpile.desired_quantity, 0)) as stockpile_delta,
				ABS(ca.quantity - COALESCE(stockpile.desired_quantity, 0)) * COALESCE(market.buy_price, 0) as deficit_value,
				loc.station_name as structure_name,
				loc.solar_system_name as solar_system,
				loc.region_name as region,
				COALESCE(divisions.name, loc.container_location_flag) || ' - ' || containerTypes.type_name as container_name
			FROM corporation_asset_locations loc
			INNER JOIN corporation_assets ca ON (
				ca.item_id = loc.item_id
				AND ca.corporation_id = loc.corporation_id
				AND ca.user_id = loc.user_id
			)
			INNER JOIN player_corporations corps ON corps.id = loc.corporation_id
			INNER JOIN asset_item_types assetTypes ON assetTypes.type_id = loc.type_id
			INNER JOIN asset_item_types containerTypes ON containerTypes.type_id = loc.container_type_id
			LEFT JOIN corporation_divisions divisions ON (
				divisions.division_number = loc.division_number
				AND divisions.corporation_id = loc.corporation_id
				AND divisions.user_id = loc.user_id
				AND divisions.division_type = 'hangar'
			)
			LEFT JOIN stockpile_markers stockpile ON (
				stockpile.type_id = loc.type_id
				AND stockpile.division_number = loc.division_number
				AND stockpile.container_id = loc.container_id
				AND stockpile.owner_id = loc.corporation_id
			)
			LEFT JOIN market_prices market ON (market.type_id = loc.type_id AND market.region_id = 10000002)
			WHERE loc.user_id = $1
				AND loc.location_type = 'item'
				AND loc.container_location_flag LIKE 'CorpSAG%'
				AND loc.station_id IS NOT NULL
				AND NOT (ca.is_singleton = true AND assetTypes.type_name LIKE '%Container')
				AND (ca.quantity - COALESCE(stockpile.desired_quantity, 0)) < 0
		)
		SELECT * FROM all_deficits
		ORDER BY deficit_value DESC NULLS LAST, structure_name, name
	`

	rows, err := r.db.QueryContext(ctx, query, user)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query stockpile deficits")
	}
	defer rows.Close()

	for rows.Next() {
		item := &StockpileItem{}
		err = rows.Scan(
			&item.Name,
			&item.TypeID,
			&item.Quantity,
			&item.Volume,
			&item.OwnerType,
			&item.OwnerName,
			&item.OwnerID,
			&item.DesiredQuantity,
			&item.StockpileDelta,
			&item.DeficitValue,
			&item.StructureName,
			&item.SolarSystem,
			&item.Region,
			&item.ContainerName,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan stockpile item")
		}

		response.Items = append(response.Items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating stockpile rows")
	}

	return response, nil
}

func (r *Assets) GetUserAssetsSummary(ctx context.Context, user int64) (*AssetsSummary, error) {
	query := `
	SELECT
		COALESCE(SUM(total_value), 0) as total_value,
		COALESCE(SUM(deficit_value), 0) as total_deficit
	FROM (
		-- Character assets
		SELECT
			(characterAssets.quantity * COALESCE(prices.sell_price, 0)) as total_value,
			CASE
				WHEN stockpileMarkers.desired_quantity IS NOT NULL AND characterAssets.quantity < stockpileMarkers.desired_quantity
				THEN (stockpileMarkers.desired_quantity - characterAssets.quantity) * COALESCE(prices.buy_price, 0)
				ELSE 0
			END as deficit_value
		FROM
			character_assets characterAssets
		INNER JOIN
			characters characters
		ON
			characterAssets.character_id = characters.id
		LEFT JOIN
			market_prices prices
		ON
			characterAssets.type_id = prices.type_id
		LEFT JOIN
			stockpile_markers stockpileMarkers
		ON
			stockpileMarkers.type_id = characterAssets.type_id
			AND stockpileMarkers.location_id = characterAssets.location_id
			AND stockpileMarkers.owner_id = characterAssets.character_id
			AND stockpileMarkers.owner_type = 'character'
			AND stockpileMarkers.container_id IS NULL
			AND stockpileMarkers.division_number IS NULL
		WHERE
			characters.user_id = $1
			AND characterAssets.location_flag IN ('Hangar', 'Deliveries', 'AssetSafety')

		UNION ALL

		-- Character assets in containers
		SELECT
			(containerAssets.quantity * COALESCE(prices.sell_price, 0)) as total_value,
			CASE
				WHEN stockpileMarkers.desired_quantity IS NOT NULL AND containerAssets.quantity < stockpileMarkers.desired_quantity
				THEN (stockpileMarkers.desired_quantity - containerAssets.quantity) * COALESCE(prices.buy_price, 0)
				ELSE 0
			END as deficit_value
		FROM
			character_assets containerAssets
		INNER JOIN
			character_assets containerItem
		ON
			containerAssets.location_id = containerItem.item_id
		INNER JOIN
			characters characters
		ON
			containerAssets.character_id = characters.id
		LEFT JOIN
			market_prices prices
		ON
			containerAssets.type_id = prices.type_id
		LEFT JOIN
			stockpile_markers stockpileMarkers
		ON
			stockpileMarkers.type_id = containerAssets.type_id
			AND stockpileMarkers.location_id = containerAssets.location_id
			AND stockpileMarkers.owner_id = containerAssets.character_id
			AND stockpileMarkers.owner_type = 'character'
			AND stockpileMarkers.container_id = containerItem.item_id
		WHERE
			characters.user_id = $1
			AND containerItem.location_flag = 'Hangar'

		UNION ALL

		-- Corporation assets
		SELECT
			(corpAssets.quantity * COALESCE(prices.sell_price, 0)) as total_value,
			CASE
				WHEN stockpileMarkers.desired_quantity IS NOT NULL AND corpAssets.quantity < stockpileMarkers.desired_quantity
				THEN (stockpileMarkers.desired_quantity - corpAssets.quantity) * COALESCE(prices.buy_price, 0)
				ELSE 0
			END as deficit_value
		FROM
			corporation_assets corpAssets
		LEFT JOIN
			market_prices prices
		ON
			corpAssets.type_id = prices.type_id
		LEFT JOIN
			stockpile_markers stockpileMarkers
		ON
			stockpileMarkers.type_id = corpAssets.type_id
			AND stockpileMarkers.location_id = corpAssets.location_id
			AND stockpileMarkers.owner_id = corpAssets.corporation_id
			AND stockpileMarkers.owner_type = 'corporation'
			AND stockpileMarkers.division_number = SUBSTRING(corpAssets.location_flag, 8, 1)::int
			AND stockpileMarkers.container_id IS NULL
		WHERE
			corpAssets.user_id = $1
			AND corpAssets.location_flag LIKE 'CorpSAG%'

		UNION ALL

		-- Corporation assets in containers
		SELECT
			(containerAssets.quantity * COALESCE(prices.sell_price, 0)) as total_value,
			CASE
				WHEN stockpileMarkers.desired_quantity IS NOT NULL AND containerAssets.quantity < stockpileMarkers.desired_quantity
				THEN (stockpileMarkers.desired_quantity - containerAssets.quantity) * COALESCE(prices.buy_price, 0)
				ELSE 0
			END as deficit_value
		FROM
			corporation_assets containerAssets
		INNER JOIN
			corporation_assets containerItem
		ON
			containerAssets.location_id = containerItem.item_id
			AND containerAssets.corporation_id = containerItem.corporation_id
			AND containerAssets.user_id = containerItem.user_id
		LEFT JOIN
			market_prices prices
		ON
			containerAssets.type_id = prices.type_id
		LEFT JOIN
			stockpile_markers stockpileMarkers
		ON
			stockpileMarkers.type_id = containerAssets.type_id
			AND stockpileMarkers.location_id = containerAssets.location_id
			AND stockpileMarkers.owner_id = containerAssets.corporation_id
			AND stockpileMarkers.owner_type = 'corporation'
			AND stockpileMarkers.container_id = containerItem.item_id
			AND stockpileMarkers.division_number = SUBSTRING(containerItem.location_flag, 8, 1)::int
		WHERE
			containerAssets.user_id = $1
			AND containerAssets.location_type = 'item'
			AND containerItem.location_flag LIKE 'CorpSAG%'
	) all_assets
	`

	summary := &AssetsSummary{}
	err := r.db.QueryRowContext(ctx, query, user).Scan(&summary.TotalValue, &summary.TotalDeficit)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get assets summary")
	}

	return summary, nil
}
