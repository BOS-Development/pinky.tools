package repositories

import (
	"context"

	"github.com/pkg/errors"
)

type orphanStockpileRow struct {
	TypeID          int64
	TypeName        string
	Volume          float64
	OwnerType       string
	OwnerID         int64
	LocationID      int64
	ContainerID     *int64
	DivisionNumber  *int
	DesiredQuantity int64
	UnitPrice       float64
	OwnerName       string
}

// InjectOrphanStockpileRows adds phantom asset rows for stockpile markers that have no matching asset.
func (r *Assets) InjectOrphanStockpileRows(ctx context.Context, userID int64, response *AssetsResponse) error {
	query := `
		SELECT
			sm.type_id, ait.type_name, ait.volume,
			sm.owner_type, sm.owner_id, sm.location_id,
			sm.container_id, sm.division_number, sm.desired_quantity,
			COALESCE(mp.buy_price, 0) as unit_price,
			COALESCE(chars.name, corps.name, '') as owner_name
		FROM stockpile_markers sm
		INNER JOIN asset_item_types ait ON ait.type_id = sm.type_id
		LEFT JOIN market_prices mp ON mp.type_id = sm.type_id AND mp.region_id = 10000002
		LEFT JOIN characters chars ON chars.id = sm.owner_id AND sm.owner_type = 'character'
		LEFT JOIN player_corporations corps ON corps.id = sm.owner_id AND sm.owner_type = 'corporation'
		WHERE sm.user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return errors.Wrap(err, "failed to query stockpile markers for phantom rows")
	}
	defer rows.Close()

	var markers []*orphanStockpileRow
	for rows.Next() {
		var m orphanStockpileRow
		err = rows.Scan(
			&m.TypeID, &m.TypeName, &m.Volume,
			&m.OwnerType, &m.OwnerID, &m.LocationID,
			&m.ContainerID, &m.DivisionNumber, &m.DesiredQuantity,
			&m.UnitPrice, &m.OwnerName,
		)
		if err != nil {
			return errors.Wrap(err, "failed to scan orphan stockpile row")
		}
		markers = append(markers, &m)
	}

	if err = rows.Err(); err != nil {
		return errors.Wrap(err, "error iterating orphan stockpile rows")
	}

	// Build set of existing asset keys to avoid duplicates
	type assetKey struct {
		typeID         int64
		ownerID        int64
		locationID     int64
		containerID    int64 // 0 for no container
		divisionNumber int   // 0 for no division
	}

	existing := make(map[assetKey]bool)
	for _, structure := range response.Structures {
		for _, a := range structure.HangarAssets {
			existing[assetKey{a.TypeID, a.OwnerID, structure.ID, 0, 0}] = true
		}
		for _, c := range structure.HangarContainers {
			for _, a := range c.Assets {
				existing[assetKey{a.TypeID, a.OwnerID, structure.ID, c.ID, 0}] = true
			}
		}
		for _, h := range structure.CorporationHangers {
			for _, a := range h.Assets {
				existing[assetKey{a.TypeID, a.OwnerID, structure.ID, 0, int(h.ID)}] = true
			}
			for _, c := range h.HangarContainers {
				for _, a := range c.Assets {
					existing[assetKey{a.TypeID, a.OwnerID, structure.ID, c.ID, int(h.ID)}] = true
				}
			}
		}
	}

	// Inject phantom rows
	for _, m := range markers {
		cid := int64(0)
		if m.ContainerID != nil {
			cid = *m.ContainerID
		}
		div := 0
		if m.DivisionNumber != nil {
			div = *m.DivisionNumber
		}
		if existing[assetKey{m.TypeID, m.OwnerID, m.LocationID, cid, div}] {
			continue
		}

		delta := -m.DesiredQuantity
		deficitValue := float64(m.DesiredQuantity) * m.UnitPrice
		phantom := &Asset{
			Name:            m.TypeName,
			TypeID:          m.TypeID,
			Quantity:        0,
			Volume:          0,
			OwnerType:       m.OwnerType,
			OwnerName:       m.OwnerName,
			OwnerID:         m.OwnerID,
			DesiredQuantity: &m.DesiredQuantity,
			StockpileDelta:  &delta,
			UnitPrice:       &m.UnitPrice,
			TotalValue:      nil,
			DeficitValue:    &deficitValue,
		}

		for _, structure := range response.Structures {
			if structure.ID != m.LocationID {
				continue
			}

			if m.ContainerID != nil {
				// Inject into a container
				if m.DivisionNumber != nil {
					// Corp container
					for _, h := range structure.CorporationHangers {
						if int(h.ID) == *m.DivisionNumber {
							for _, c := range h.HangarContainers {
								if c.ID == *m.ContainerID {
									c.Assets = append(c.Assets, phantom)
								}
							}
						}
					}
				} else {
					// Character container
					for _, c := range structure.HangarContainers {
						if c.ID == *m.ContainerID {
							c.Assets = append(c.Assets, phantom)
						}
					}
				}
			} else if m.DivisionNumber != nil {
				// Corp hanger direct assets
				for _, h := range structure.CorporationHangers {
					if int(h.ID) == *m.DivisionNumber {
						h.Assets = append(h.Assets, phantom)
					}
				}
			} else {
				// Personal hangar
				structure.HangarAssets = append(structure.HangarAssets, phantom)
			}
		}
	}

	return nil
}
