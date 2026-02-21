package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type PiPlanets struct {
	db *sql.DB
}

func NewPiPlanets(db *sql.DB) *PiPlanets {
	return &PiPlanets{db: db}
}

func (r *PiPlanets) UpsertPlanets(ctx context.Context, characterID, userID int64, planets []*models.PiPlanet) error {
	if len(planets) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction for pi planets upsert")
	}
	defer tx.Rollback()

	// Delete planets that are no longer in the incoming list
	incomingPlanetIDs := make([]int64, len(planets))
	for i, p := range planets {
		incomingPlanetIDs[i] = p.PlanetID
	}

	_, err = tx.ExecContext(ctx, `
		DELETE FROM pi_planets
		WHERE character_id = $1
		  AND planet_id != ALL($2::bigint[])
	`, characterID, pq.Array(incomingPlanetIDs))
	if err != nil {
		return errors.Wrap(err, "failed to delete removed pi planets")
	}

	upsertQuery := `
		INSERT INTO pi_planets
			(character_id, user_id, planet_id, planet_type, solar_system_id,
			 upgrade_level, num_pins, last_update, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now())
		ON CONFLICT (character_id, planet_id)
		DO UPDATE SET
			user_id = EXCLUDED.user_id,
			planet_type = EXCLUDED.planet_type,
			solar_system_id = EXCLUDED.solar_system_id,
			upgrade_level = EXCLUDED.upgrade_level,
			num_pins = EXCLUDED.num_pins,
			last_update = EXCLUDED.last_update,
			updated_at = now()
	`

	stmt, err := tx.PrepareContext(ctx, upsertQuery)
	if err != nil {
		return errors.Wrap(err, "failed to prepare pi planet upsert")
	}

	for _, planet := range planets {
		_, err = stmt.ExecContext(ctx,
			characterID,
			userID,
			planet.PlanetID,
			planet.PlanetType,
			planet.SolarSystemID,
			planet.UpgradeLevel,
			planet.NumPins,
			planet.LastUpdate,
		)
		if err != nil {
			return errors.Wrap(err, "failed to execute pi planet upsert")
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit pi planets transaction")
	}
	return nil
}

func (r *PiPlanets) UpsertColony(ctx context.Context, characterID, planetID int64, pins []*models.PiPin, contents []*models.PiPinContent, routes []*models.PiRoute) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction for pi colony upsert")
	}
	defer tx.Rollback()

	// Delete existing data for this colony (contents and routes first due to logical dependency)
	_, err = tx.ExecContext(ctx, `DELETE FROM pi_pin_contents WHERE character_id = $1 AND planet_id = $2`, characterID, planetID)
	if err != nil {
		return errors.Wrap(err, "failed to delete existing pi pin contents")
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM pi_routes WHERE character_id = $1 AND planet_id = $2`, characterID, planetID)
	if err != nil {
		return errors.Wrap(err, "failed to delete existing pi routes")
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM pi_pins WHERE character_id = $1 AND planet_id = $2`, characterID, planetID)
	if err != nil {
		return errors.Wrap(err, "failed to delete existing pi pins")
	}

	// Insert new pins
	if len(pins) > 0 {
		pinQuery := `
			INSERT INTO pi_pins
				(character_id, planet_id, pin_id, type_id, schematic_id,
				 latitude, longitude, install_time, expiry_time,
				 last_cycle_start, extractor_cycle_time, extractor_head_radius,
				 extractor_product_type_id, extractor_qty_per_cycle,
				 extractor_num_heads, pin_category, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, now())
		`

		pinStmt, err := tx.PrepareContext(ctx, pinQuery)
		if err != nil {
			return errors.Wrap(err, "failed to prepare pi pin insert")
		}

		for _, pin := range pins {
			_, err = pinStmt.ExecContext(ctx,
				characterID,
				planetID,
				pin.PinID,
				pin.TypeID,
				pin.SchematicID,
				pin.Latitude,
				pin.Longitude,
				pin.InstallTime,
				pin.ExpiryTime,
				pin.LastCycleStart,
				pin.ExtractorCycleTime,
				pin.ExtractorHeadRadius,
				pin.ExtractorProductTypeID,
				pin.ExtractorQtyPerCycle,
				pin.ExtractorNumHeads,
				pin.PinCategory,
			)
			if err != nil {
				return errors.Wrap(err, "failed to execute pi pin insert")
			}
		}
	}

	// Insert new contents
	if len(contents) > 0 {
		contentQuery := `
			INSERT INTO pi_pin_contents
				(character_id, planet_id, pin_id, type_id, amount)
			VALUES ($1, $2, $3, $4, $5)
		`

		contentStmt, err := tx.PrepareContext(ctx, contentQuery)
		if err != nil {
			return errors.Wrap(err, "failed to prepare pi pin content insert")
		}

		for _, content := range contents {
			_, err = contentStmt.ExecContext(ctx,
				characterID,
				planetID,
				content.PinID,
				content.TypeID,
				content.Amount,
			)
			if err != nil {
				return errors.Wrap(err, "failed to execute pi pin content insert")
			}
		}
	}

	// Insert new routes
	if len(routes) > 0 {
		routeQuery := `
			INSERT INTO pi_routes
				(character_id, planet_id, route_id, source_pin_id,
				 destination_pin_id, content_type_id, quantity)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`

		routeStmt, err := tx.PrepareContext(ctx, routeQuery)
		if err != nil {
			return errors.Wrap(err, "failed to prepare pi route insert")
		}

		for _, route := range routes {
			_, err = routeStmt.ExecContext(ctx,
				characterID,
				planetID,
				route.RouteID,
				route.SourcePinID,
				route.DestinationPinID,
				route.ContentTypeID,
				route.Quantity,
			)
			if err != nil {
				return errors.Wrap(err, "failed to execute pi route insert")
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit pi colony transaction")
	}
	return nil
}

func (r *PiPlanets) GetPlanetsForUser(ctx context.Context, userID int64) ([]*models.PiPlanet, error) {
	query := `
		SELECT id, character_id, user_id, planet_id, planet_type,
		       solar_system_id, upgrade_level, num_pins, last_update,
		       last_stall_notified_at
		FROM pi_planets
		WHERE user_id = $1
		ORDER BY character_id, planet_id
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query pi planets")
	}
	defer rows.Close()

	planets := []*models.PiPlanet{}
	for rows.Next() {
		var planet models.PiPlanet
		err = rows.Scan(
			&planet.ID,
			&planet.CharacterID,
			&planet.UserID,
			&planet.PlanetID,
			&planet.PlanetType,
			&planet.SolarSystemID,
			&planet.UpgradeLevel,
			&planet.NumPins,
			&planet.LastUpdate,
			&planet.LastStallNotifiedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan pi planet")
		}
		planets = append(planets, &planet)
	}

	return planets, nil
}

func (r *PiPlanets) GetPinsForPlanets(ctx context.Context, userID int64) ([]*models.PiPin, error) {
	query := `
		SELECT pi_pins.id, pi_pins.character_id, pi_pins.planet_id, pi_pins.pin_id,
		       pi_pins.type_id, pi_pins.schematic_id, pi_pins.latitude, pi_pins.longitude,
		       pi_pins.install_time, pi_pins.expiry_time, pi_pins.last_cycle_start,
		       pi_pins.extractor_cycle_time, pi_pins.extractor_head_radius,
		       pi_pins.extractor_product_type_id, pi_pins.extractor_qty_per_cycle,
		       pi_pins.extractor_num_heads, pi_pins.pin_category
		FROM pi_pins
		INNER JOIN pi_planets
			ON pi_pins.character_id = pi_planets.character_id
			AND pi_pins.planet_id = pi_planets.planet_id
		WHERE pi_planets.user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query pi pins")
	}
	defer rows.Close()

	pins := []*models.PiPin{}
	for rows.Next() {
		var pin models.PiPin
		err = rows.Scan(
			&pin.ID,
			&pin.CharacterID,
			&pin.PlanetID,
			&pin.PinID,
			&pin.TypeID,
			&pin.SchematicID,
			&pin.Latitude,
			&pin.Longitude,
			&pin.InstallTime,
			&pin.ExpiryTime,
			&pin.LastCycleStart,
			&pin.ExtractorCycleTime,
			&pin.ExtractorHeadRadius,
			&pin.ExtractorProductTypeID,
			&pin.ExtractorQtyPerCycle,
			&pin.ExtractorNumHeads,
			&pin.PinCategory,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan pi pin")
		}
		pins = append(pins, &pin)
	}

	return pins, nil
}

func (r *PiPlanets) GetPinContentsForUser(ctx context.Context, userID int64) ([]*models.PiPinContent, error) {
	query := `
		SELECT pi_pin_contents.character_id, pi_pin_contents.planet_id,
		       pi_pin_contents.pin_id, pi_pin_contents.type_id,
		       pi_pin_contents.amount
		FROM pi_pin_contents
		INNER JOIN pi_planets
			ON pi_pin_contents.character_id = pi_planets.character_id
			AND pi_pin_contents.planet_id = pi_planets.planet_id
		WHERE pi_planets.user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query pi pin contents")
	}
	defer rows.Close()

	contents := []*models.PiPinContent{}
	for rows.Next() {
		var content models.PiPinContent
		err = rows.Scan(
			&content.CharacterID,
			&content.PlanetID,
			&content.PinID,
			&content.TypeID,
			&content.Amount,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan pi pin content")
		}
		contents = append(contents, &content)
	}

	return contents, nil
}

func (r *PiPlanets) GetRoutesForUser(ctx context.Context, userID int64) ([]*models.PiRoute, error) {
	query := `
		SELECT pi_routes.character_id, pi_routes.planet_id, pi_routes.route_id,
		       pi_routes.source_pin_id, pi_routes.destination_pin_id,
		       pi_routes.content_type_id, pi_routes.quantity
		FROM pi_routes
		INNER JOIN pi_planets
			ON pi_routes.character_id = pi_planets.character_id
			AND pi_routes.planet_id = pi_planets.planet_id
		WHERE pi_planets.user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query pi routes")
	}
	defer rows.Close()

	routes := []*models.PiRoute{}
	for rows.Next() {
		var route models.PiRoute
		err = rows.Scan(
			&route.CharacterID,
			&route.PlanetID,
			&route.RouteID,
			&route.SourcePinID,
			&route.DestinationPinID,
			&route.ContentTypeID,
			&route.Quantity,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan pi route")
		}
		routes = append(routes, &route)
	}

	return routes, nil
}

func (r *PiPlanets) UpdateStallNotifiedAt(ctx context.Context, characterID, planetID int64, notifiedAt *time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE pi_planets
		SET last_stall_notified_at = $3
		WHERE character_id = $1 AND planet_id = $2
	`, characterID, planetID, notifiedAt)
	if err != nil {
		return errors.Wrap(err, "failed to update pi planet stall notification time")
	}
	return nil
}
