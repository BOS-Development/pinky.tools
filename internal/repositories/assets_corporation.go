package repositories

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

// loadCorpStations queries corp stations, adds new ones to stationMap/response, returns stationCorpMap.
func (r *Assets) loadCorpStations(ctx context.Context, user int64, stationMap map[int64]*AssetStructure, response *AssetsResponse) (map[int64]map[int64]bool, error) {
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

	stations, err := r.db.QueryContext(ctx, corpStationsQuery, user)
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

	return stationCorpMap, nil
}

// loadCorpDivisionTemplates queries corp divisions and returns templates map.
func (r *Assets) loadCorpDivisionTemplates(ctx context.Context, user int64) (map[int64]map[int64]*CorporationHanger, error) {
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

	return corpDivisionTemplates, nil
}

// loadCorpItems loads corp hangared items into hangerMap / stationCorpMap.
func (r *Assets) loadCorpItems(ctx context.Context, user int64, stationCorpMap map[int64]map[int64]bool, hangerMap map[int64]map[int64]map[int64]*CorporationHanger, divisionTemplates map[int64]map[int64]*CorporationHanger) error {
	corpHangaredItemsQuery := `
SELECT
	corporation_assets.corporation_id,
	player_corporations.name,
	office.location_id as station_id,
	SUBSTRING(corporation_assets.location_flag, 8, 1)::int as "division_number",
	assetTypes.type_id,
	assetTypes.type_name,
	SUM(corporation_assets.quantity) as quantity,
	SUM(assetTypes.volume * corporation_assets.quantity) as "volume",
	stockpile.desired_quantity,
	(SUM(corporation_assets.quantity) - COALESCE(stockpile.desired_quantity, 0)) as stockpile_delta,
	market.sell_price as unit_price,
	(SUM(corporation_assets.quantity) * COALESCE(market.sell_price, 0)) as total_value,
	CASE
		WHEN (SUM(corporation_assets.quantity) - COALESCE(stockpile.desired_quantity, 0)) < 0
		THEN ABS(SUM(corporation_assets.quantity) - COALESCE(stockpile.desired_quantity, 0)) * COALESCE(market.buy_price, 0)
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
INNER JOIN
	corporation_assets office
ON
	office.item_id = corporation_assets.location_id
	AND office.location_flag = 'OfficeFolder'
	AND office.user_id = $1
LEFT JOIN
	stockpile_markers stockpile
ON
	stockpile.user_id = $1
	AND stockpile.type_id = corporation_assets.type_id
	AND stockpile.owner_type = 'corporation'
	AND stockpile.owner_id = corporation_assets.corporation_id
	AND stockpile.location_id = office.location_id
	AND stockpile.division_number = SUBSTRING(corporation_assets.location_flag, 8, 1)::int
	AND stockpile.container_id IS NULL
LEFT JOIN
	market_prices market
ON
	market.type_id = corporation_assets.type_id
	AND market.region_id = 10000002
WHERE
	corporation_assets.user_id=$1
	AND NOT (corporation_assets.is_singleton=true AND assetTypes.type_name like '%Container')
	AND corporation_assets.location_type='item'
	AND corporation_assets.location_flag like 'CorpSAG%'
GROUP BY
	corporation_assets.corporation_id,
	player_corporations.name,
	office.location_id,
	SUBSTRING(corporation_assets.location_flag, 8, 1)::int,
	assetTypes.type_id,
	assetTypes.type_name,
	stockpile.desired_quantity,
	market.sell_price,
	market.buy_price;`

	corpHangaredItems, err := r.db.QueryContext(ctx, corpHangaredItemsQuery, user)
	if err != nil {
		return errors.Wrap(err, "failed to query corp hangared assets from database")
	}
	defer corpHangaredItems.Close()

	for corpHangaredItems.Next() {
		asset := &Asset{}
		var location int64
		var divisionNumber int64

		asset.OwnerType = "corporation"

		err = corpHangaredItems.Scan(&asset.OwnerID, &asset.OwnerName, &location, &divisionNumber, &asset.TypeID, &asset.Name, &asset.Quantity, &asset.Volume, &asset.DesiredQuantity, &asset.StockpileDelta, &asset.UnitPrice, &asset.TotalValue, &asset.DeficitValue)
		if err != nil {
			return errors.Wrap(err, "failed to scan corp hangared item")
		}

		// Get or create station-specific division
		hanger := getOrCreateDivision(hangerMap, divisionTemplates, location, asset.OwnerID, divisionNumber)
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

	return nil
}

// loadCorpContainers queries corp containers (recursive CTE) and populates hangerMap.
// Returns corpContainerMap and corpContainersByDivision (for fallback).
func (r *Assets) loadCorpContainers(ctx context.Context, user int64, stationCorpMap map[int64]map[int64]bool, hangerMap map[int64]map[int64]map[int64]*CorporationHanger, divisionTemplates map[int64]map[int64]*CorporationHanger) (map[int64]*AssetContainer, map[int64]map[int64][]*AssetContainer, error) {
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
	COALESCE(loc.name, ait.type_name) as container_name
FROM (SELECT DISTINCT container_id, corporation_id FROM container_chain) cc
INNER JOIN asset_item_types ait ON ait.type_id = (
	SELECT type_id FROM container_chain WHERE item_id = cc.container_id LIMIT 1
)
INNER JOIN player_corporations pc ON pc.id = cc.corporation_id
LEFT JOIN corporation_asset_location_names loc ON loc.item_id = cc.container_id
WHERE ait.type_name LIKE '%Container'
ORDER BY cc.container_id;`

	corpContainers, err := r.db.QueryContext(ctx, corpContainerQuery, user)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to query corp containers from database")
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
			return nil, nil, errors.Wrap(err, "failed to scan corp container")
		}

		container.Assets = []*Asset{}
		corpContainerMap[container.ID] = container

		// If we successfully determined a station_id, add container directly to that station
		if location.Valid {
			stationID := location.Int64
			hanger := getOrCreateDivision(hangerMap, divisionTemplates, stationID, container.OwnerID, divisionNumber)
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

	return corpContainerMap, corpContainersByDivision, nil
}

// loadCorpContainerItems loads corp items inside containers into corpContainerMap.
func (r *Assets) loadCorpContainerItems(ctx context.Context, user int64, corpContainerMap map[int64]*AssetContainer) error {
	corpItemsInContainersQuery := `
SELECT
	corporation_assets.corporation_id,
	player_corporations.name,
	assetTypes.type_id,
	assetTypes.type_name,
	SUM(corporation_assets.quantity) as quantity,
	SUM(assetTypes.volume * corporation_assets.quantity) as "volume",
	corporation_assets.location_id,
	stockpile.desired_quantity,
	(SUM(corporation_assets.quantity) - COALESCE(stockpile.desired_quantity, 0)) as stockpile_delta,
	market.sell_price as unit_price,
	(SUM(corporation_assets.quantity) * COALESCE(market.sell_price, 0)) as total_value,
	CASE
		WHEN (SUM(corporation_assets.quantity) - COALESCE(stockpile.desired_quantity, 0)) < 0
		THEN ABS(SUM(corporation_assets.quantity) - COALESCE(stockpile.desired_quantity, 0)) * COALESCE(market.buy_price, 0)
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
GROUP BY
	corporation_assets.corporation_id,
	player_corporations.name,
	assetTypes.type_id,
	assetTypes.type_name,
	corporation_assets.location_id,
	stockpile.desired_quantity,
	market.sell_price,
	market.buy_price
ORDER BY
	assetTypes.type_name;`

	corpItemsInContainers, err := r.db.QueryContext(ctx, corpItemsInContainersQuery, user)
	if err != nil {
		return errors.Wrap(err, "failed to query corp items in containers from database")
	}
	defer corpItemsInContainers.Close()

	for corpItemsInContainers.Next() {
		asset := &Asset{}
		var location int64

		asset.OwnerType = "corporation"

		err = corpItemsInContainers.Scan(&asset.OwnerID, &asset.OwnerName, &asset.TypeID, &asset.Name, &asset.Quantity, &asset.Volume, &location, &asset.DesiredQuantity, &asset.StockpileDelta, &asset.UnitPrice, &asset.TotalValue, &asset.DeficitValue)
		if err != nil {
			return errors.Wrap(err, "failed to scan corp container item")
		}

		container, ok := corpContainerMap[location]
		if !ok {
			continue
		}

		container.Assets = append(container.Assets, asset)
	}

	return nil
}

// getOrCreateDivision is a package-level helper to lazily instantiate a station-specific division.
func getOrCreateDivision(hangerMap map[int64]map[int64]map[int64]*CorporationHanger, divisionTemplates map[int64]map[int64]*CorporationHanger, stationID, corpID, divisionID int64) *CorporationHanger {
	if hangerMap[stationID] == nil {
		hangerMap[stationID] = map[int64]map[int64]*CorporationHanger{}
	}
	if hangerMap[stationID][corpID] == nil {
		hangerMap[stationID][corpID] = map[int64]*CorporationHanger{}
	}

	// If division doesn't exist at this station yet, create it from template
	if hangerMap[stationID][corpID][divisionID] == nil {
		template, ok := divisionTemplates[corpID][divisionID]
		if !ok {
			return nil
		}
		hangerMap[stationID][corpID][divisionID] = &CorporationHanger{
			ID:               template.ID,
			Name:             template.Name,
			CorporationID:    template.CorporationID,
			CorporationName:  template.CorporationName,
			Assets:           []*Asset{},
			HangarContainers: []*AssetContainer{},
		}
	}
	return hangerMap[stationID][corpID][divisionID]
}

// attachCorpDivisions attaches all divisions to their stations in response.
func attachCorpDivisions(stationMap map[int64]*AssetStructure, stationCorpMap map[int64]map[int64]bool, divisionTemplates map[int64]map[int64]*CorporationHanger, hangerMap map[int64]map[int64]map[int64]*CorporationHanger, containersByDivision map[int64]map[int64][]*AssetContainer, response *AssetsResponse) {
	// Add divisions to stations
	// If a corp has ANY assets at a station, show ALL its divisions (even empty ones)
	for stationID, corpIDs := range stationCorpMap {
		station, ok := stationMap[stationID]
		if !ok {
			continue
		}

		for corpID := range corpIDs {
			// Get all division templates for this corp
			templates, ok := divisionTemplates[corpID]
			if !ok {
				continue
			}

			// For each defined division
			for divisionID, template := range templates {
				// Check if we already created this division at this station (has content)
				var division *CorporationHanger
				if hangerMap[stationID] != nil && hangerMap[stationID][corpID] != nil {
					division = hangerMap[stationID][corpID][divisionID]
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
	for corpID, divisions := range containersByDivision {
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
			templates, ok := divisionTemplates[corpID]
			if ok {
				for divisionID, template := range templates {
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
}
