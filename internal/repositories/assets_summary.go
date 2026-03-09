package repositories

import (
	"context"

	"github.com/pkg/errors"
)

func (r *Assets) GetUserAssetsSummary(ctx context.Context, user int64) (*AssetsSummary, error) {
	query := `
	SELECT
		COALESCE(SUM(total_value), 0) as total_value,
		COALESCE(SUM(deficit_value), 0) as total_deficit,
		(SELECT COUNT(*) FROM esi_industry_jobs WHERE user_id = $1 AND status = 'active') as active_jobs
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
	err := r.db.QueryRowContext(ctx, query, user).Scan(&summary.TotalValue, &summary.TotalDeficit, &summary.ActiveJobs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get assets summary")
	}

	return summary, nil
}
