package repositories

import (
	"context"

	"github.com/pkg/errors"
)

// loadCharStations queries character stations and returns stationMap + structures slice.
func (r *Assets) loadCharStations(ctx context.Context, user int64) (map[int64]*AssetStructure, []*AssetStructure, error) {
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
		return nil, nil, errors.Wrap(err, "failed to query stations from database")
	}
	defer stations.Close()

	stationMap := map[int64]*AssetStructure{}
	var structures []*AssetStructure
	for stations.Next() {
		structure := &AssetStructure{}
		err = stations.Scan(&structure.ID, &structure.Name, &structure.SolarSystem, &structure.Region)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to scan item")
		}
		structures = append(structures, structure)

		structure.Deliveries = []*Asset{}
		structure.HangarAssets = []*Asset{}
		structure.HangarContainers = []*AssetContainer{}
		structure.AssetSafety = []*Asset{}
		structure.CorporationHangers = []*CorporationHanger{}

		stationMap[structure.ID] = structure
	}

	return stationMap, structures, nil
}

// loadCharItems loads hangared items and deliveries into stationMap.
func (r *Assets) loadCharItems(ctx context.Context, user int64, stationMap map[int64]*AssetStructure) error {
	hangaredItemsQuery := `
SELECT
	characterAssets.character_id,
	characters.name,
	characterAssets.location_id,
	characterAssets.location_flag,
    assetTypes.type_id,
    assetTypes.type_name,
    SUM(characterAssets.quantity) as quantity,
    SUM(assetTypes.volume * characterAssets.quantity) as "volume",
    stockpile.desired_quantity,
    (SUM(characterAssets.quantity) - COALESCE(stockpile.desired_quantity, 0)) as stockpile_delta,
    market.sell_price as unit_price,
    (SUM(characterAssets.quantity) * COALESCE(market.sell_price, 0)) as total_value,
    CASE
        WHEN (SUM(characterAssets.quantity) - COALESCE(stockpile.desired_quantity, 0)) < 0
        THEN ABS(SUM(characterAssets.quantity) - COALESCE(stockpile.desired_quantity, 0)) * COALESCE(market.buy_price, 0)
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
    )
GROUP BY
    characterAssets.character_id,
    characters.name,
    characterAssets.location_id,
    characterAssets.location_flag,
    assetTypes.type_id,
    assetTypes.type_name,
    stockpile.desired_quantity,
    market.sell_price,
    market.buy_price;`

	items, err := r.db.QueryContext(ctx, hangaredItemsQuery, user)
	if err != nil {
		return errors.Wrap(err, "failed to query hangared assets from database")
	}
	defer items.Close()

	for items.Next() {
		asset := &Asset{}
		var location int64
		var locationFlag string

		asset.OwnerType = "character"

		err = items.Scan(&asset.OwnerID, &asset.OwnerName, &location, &locationFlag, &asset.TypeID, &asset.Name, &asset.Quantity, &asset.Volume, &asset.DesiredQuantity, &asset.StockpileDelta, &asset.UnitPrice, &asset.TotalValue, &asset.DeficitValue)
		if err != nil {
			return errors.Wrap(err, "failed to scan item")
		}

		station, ok := stationMap[location]
		if !ok {
			return errors.Errorf("location %d not found for hanger asset", location)
		}

		switch locationFlag {
		case "Hangar":
			station.HangarAssets = append(station.HangarAssets, asset)
		case "Deliveries":
			station.Deliveries = append(station.Deliveries, asset)
		default:
			return errors.Errorf("unknown location flag %s", locationFlag)
		}
	}

	return nil
}

// loadCharContainers queries character containers and returns containerMap.
func (r *Assets) loadCharContainers(ctx context.Context, user int64, stationMap map[int64]*AssetStructure) (map[int64]*AssetContainer, error) {
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

	return containerMap, nil
}

// loadCharContainerItems loads items inside containers into containerMap.
func (r *Assets) loadCharContainerItems(ctx context.Context, user int64, containerMap map[int64]*AssetContainer) error {
	itemsInContainersQuery := `
SELECT
	characterAssets.character_id,
	characters.name,
    assetTypes.type_id,
    assetTypes.type_name,
    SUM(characterAssets.quantity) as quantity,
    SUM(assetTypes.volume * characterAssets.quantity) as "volume",
    characterAssets.location_id,
    stockpile.desired_quantity,
    (SUM(characterAssets.quantity) - COALESCE(stockpile.desired_quantity, 0)) as stockpile_delta,
    market.sell_price as unit_price,
    (SUM(characterAssets.quantity) * COALESCE(market.sell_price, 0)) as total_value,
    CASE
        WHEN (SUM(characterAssets.quantity) - COALESCE(stockpile.desired_quantity, 0)) < 0
        THEN ABS(SUM(characterAssets.quantity) - COALESCE(stockpile.desired_quantity, 0)) * COALESCE(market.buy_price, 0)
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
GROUP BY
    characterAssets.character_id,
    characters.name,
    assetTypes.type_id,
    assetTypes.type_name,
    characterAssets.location_id,
    stockpile.desired_quantity,
    market.sell_price,
    market.buy_price
ORDER BY
    assetTypes.type_name;`

	itemsInContainers, err := r.db.QueryContext(ctx, itemsInContainersQuery, user)
	if err != nil {
		return errors.Wrap(err, "failed to query items in containers items from database")
	}
	defer itemsInContainers.Close()

	for itemsInContainers.Next() {
		asset := &Asset{}
		var location int64

		asset.OwnerType = "character"

		err = itemsInContainers.Scan(&asset.OwnerID, &asset.OwnerName, &asset.TypeID, &asset.Name, &asset.Quantity, &asset.Volume, &location, &asset.DesiredQuantity, &asset.StockpileDelta, &asset.UnitPrice, &asset.TotalValue, &asset.DeficitValue)
		if err != nil {
			return errors.Wrap(err, "failed to scan container")
		}

		container, ok := containerMap[location]
		if !ok {
			continue
			//return nil, errors.Errorf("location %d not found for item container", location)
		}

		container.Assets = append(container.Assets, asset)
	}

	return nil
}
