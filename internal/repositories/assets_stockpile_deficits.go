package repositories

import (
	"context"

	"github.com/pkg/errors"
)

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
				AND stockpile.location_id = loc.station_id
				AND stockpile.division_number = loc.division_number
				AND stockpile.container_id IS NULL
				AND stockpile.owner_id = loc.corporation_id
			)
			LEFT JOIN market_prices market ON (market.type_id = loc.type_id AND market.region_id = 10000002)
			WHERE loc.user_id = $1
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

			UNION ALL

			-- Orphan character stockpile markers (no matching asset in inventory)
			SELECT
				assetTypes.type_name as name,
				stockpile.type_id,
				0 as quantity,
				assetTypes.volume as volume,
				'character' as owner_type,
				characters.name as owner_name,
				stockpile.owner_id as owner_id,
				stockpile.desired_quantity,
				(0 - stockpile.desired_quantity) as stockpile_delta,
				stockpile.desired_quantity * COALESCE(market.buy_price, 0) as deficit_value,
				stations.name as structure_name,
				systems.name as solar_system,
				regions.name as region,
				NULL::text as container_name
			FROM stockpile_markers stockpile
			INNER JOIN characters ON characters.id = stockpile.owner_id
			INNER JOIN asset_item_types assetTypes ON assetTypes.type_id = stockpile.type_id
			INNER JOIN stations ON stockpile.location_id = stations.station_id
			INNER JOIN solar_systems systems ON stations.solar_system_id = systems.solar_system_id
			INNER JOIN constellations ON systems.constellation_id = constellations.constellation_id
			INNER JOIN regions ON constellations.region_id = regions.region_id
			LEFT JOIN market_prices market ON (market.type_id = stockpile.type_id AND market.region_id = 10000002)
			WHERE stockpile.user_id = $1
				AND stockpile.owner_type = 'character'
				AND stockpile.container_id IS NULL
				AND stockpile.division_number IS NULL
				AND NOT EXISTS (
					SELECT 1 FROM character_assets ca
					WHERE ca.user_id = $1
					  AND ca.character_id = stockpile.owner_id
					  AND ca.type_id = stockpile.type_id
					  AND ca.location_id = stockpile.location_id
					  AND ca.location_type = 'station'
					  AND ca.location_flag IN ('Hangar', 'Deliveries', 'AssetSafety')
				)

			UNION ALL

			-- Orphan corporation stockpile markers (no matching asset in inventory)
			SELECT
				assetTypes.type_name as name,
				stockpile.type_id,
				0 as quantity,
				assetTypes.volume as volume,
				'corporation' as owner_type,
				corps.name as owner_name,
				stockpile.owner_id as owner_id,
				stockpile.desired_quantity,
				(0 - stockpile.desired_quantity) as stockpile_delta,
				stockpile.desired_quantity * COALESCE(market.buy_price, 0) as deficit_value,
				stations.name as structure_name,
				systems.name as solar_system,
				regions.name as region,
				NULL::text as container_name
			FROM stockpile_markers stockpile
			INNER JOIN player_corporations corps ON corps.id = stockpile.owner_id
			INNER JOIN asset_item_types assetTypes ON assetTypes.type_id = stockpile.type_id
			INNER JOIN stations ON stockpile.location_id = stations.station_id
			INNER JOIN solar_systems systems ON stations.solar_system_id = systems.solar_system_id
			INNER JOIN constellations ON systems.constellation_id = constellations.constellation_id
			INNER JOIN regions ON constellations.region_id = regions.region_id
			LEFT JOIN market_prices market ON (market.type_id = stockpile.type_id AND market.region_id = 10000002)
			WHERE stockpile.user_id = $1
				AND stockpile.owner_type = 'corporation'
				AND stockpile.container_id IS NULL
				AND NOT EXISTS (
					SELECT 1 FROM corporation_asset_locations loc
					INNER JOIN corporation_assets ca ON (
						ca.item_id = loc.item_id
						AND ca.corporation_id = loc.corporation_id
						AND ca.user_id = loc.user_id
					)
					WHERE loc.user_id = $1
					  AND loc.corporation_id = stockpile.owner_id
					  AND loc.type_id = stockpile.type_id
					  AND loc.station_id = stockpile.location_id
					  AND loc.location_flag LIKE 'CorpSAG%'
					  AND (stockpile.division_number IS NULL OR loc.division_number = stockpile.division_number)
				)
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
