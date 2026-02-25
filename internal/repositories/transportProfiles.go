package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type TransportProfiles struct {
	db *sql.DB
}

func NewTransportProfiles(db *sql.DB) *TransportProfiles {
	return &TransportProfiles{db: db}
}

func (r *TransportProfiles) GetByUser(ctx context.Context, userID int64) ([]*models.TransportProfile, error) {
	query := `
		select p.id, p.user_id, p.name, p.transport_method, p.character_id,
		       p.cargo_m3, p.rate_per_m3_per_jump, p.collateral_rate, p.collateral_price_basis,
		       p.fuel_type_id, p.fuel_per_ly, p.fuel_conservation_level,
		       p.route_preference, p.is_default, p.created_at,
		       COALESCE(c.name, ''),
		       COALESCE(t.type_name, '')
		from transport_profiles p
		left join characters c on c.id = p.character_id
		left join asset_item_types t on t.type_id = p.fuel_type_id
		where p.user_id = $1
		order by p.created_at desc
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query transport profiles")
	}
	defer rows.Close()

	profiles := []*models.TransportProfile{}
	for rows.Next() {
		var p models.TransportProfile
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.Name, &p.TransportMethod, &p.CharacterID,
			&p.CargoM3, &p.RatePerM3PerJump, &p.CollateralRate, &p.CollateralPriceBasis,
			&p.FuelTypeID, &p.FuelPerLY, &p.FuelConservationLevel,
			&p.RoutePreference, &p.IsDefault, &p.CreatedAt,
			&p.CharacterName, &p.FuelTypeName,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan transport profile")
		}
		profiles = append(profiles, &p)
	}

	return profiles, nil
}

func (r *TransportProfiles) GetByID(ctx context.Context, id, userID int64) (*models.TransportProfile, error) {
	query := `
		select p.id, p.user_id, p.name, p.transport_method, p.character_id,
		       p.cargo_m3, p.rate_per_m3_per_jump, p.collateral_rate, p.collateral_price_basis,
		       p.fuel_type_id, p.fuel_per_ly, p.fuel_conservation_level,
		       p.route_preference, p.is_default, p.created_at,
		       COALESCE(c.name, ''),
		       COALESCE(t.type_name, '')
		from transport_profiles p
		left join characters c on c.id = p.character_id
		left join asset_item_types t on t.type_id = p.fuel_type_id
		where p.id = $1 and p.user_id = $2
	`

	var p models.TransportProfile
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&p.ID, &p.UserID, &p.Name, &p.TransportMethod, &p.CharacterID,
		&p.CargoM3, &p.RatePerM3PerJump, &p.CollateralRate, &p.CollateralPriceBasis,
		&p.FuelTypeID, &p.FuelPerLY, &p.FuelConservationLevel,
		&p.RoutePreference, &p.IsDefault, &p.CreatedAt,
		&p.CharacterName, &p.FuelTypeName,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get transport profile by ID")
	}

	return &p, nil
}

func (r *TransportProfiles) GetDefaultByMethod(ctx context.Context, userID int64, method string) (*models.TransportProfile, error) {
	query := `
		select p.id, p.user_id, p.name, p.transport_method, p.character_id,
		       p.cargo_m3, p.rate_per_m3_per_jump, p.collateral_rate, p.collateral_price_basis,
		       p.fuel_type_id, p.fuel_per_ly, p.fuel_conservation_level,
		       p.route_preference, p.is_default, p.created_at,
		       COALESCE(c.name, ''),
		       COALESCE(t.type_name, '')
		from transport_profiles p
		left join characters c on c.id = p.character_id
		left join asset_item_types t on t.type_id = p.fuel_type_id
		where p.user_id = $1 and p.transport_method = $2 and p.is_default = true
		limit 1
	`

	var p models.TransportProfile
	err := r.db.QueryRowContext(ctx, query, userID, method).Scan(
		&p.ID, &p.UserID, &p.Name, &p.TransportMethod, &p.CharacterID,
		&p.CargoM3, &p.RatePerM3PerJump, &p.CollateralRate, &p.CollateralPriceBasis,
		&p.FuelTypeID, &p.FuelPerLY, &p.FuelConservationLevel,
		&p.RoutePreference, &p.IsDefault, &p.CreatedAt,
		&p.CharacterName, &p.FuelTypeName,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get default transport profile")
	}

	return &p, nil
}

func (r *TransportProfiles) Create(ctx context.Context, p *models.TransportProfile) (*models.TransportProfile, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// If setting as default, clear existing defaults for this method
	if p.IsDefault {
		_, err = tx.ExecContext(ctx, `
			update transport_profiles set is_default = false
			where user_id = $1 and transport_method = $2 and is_default = true
		`, p.UserID, p.TransportMethod)
		if err != nil {
			return nil, errors.Wrap(err, "failed to clear existing defaults")
		}
	}

	query := `
		insert into transport_profiles
			(user_id, name, transport_method, character_id, cargo_m3,
			 rate_per_m3_per_jump, collateral_rate, collateral_price_basis,
			 fuel_type_id, fuel_per_ly, fuel_conservation_level,
			 route_preference, is_default)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		returning id, user_id, name, transport_method, character_id,
		          cargo_m3, rate_per_m3_per_jump, collateral_rate, collateral_price_basis,
		          fuel_type_id, fuel_per_ly, fuel_conservation_level,
		          route_preference, is_default, created_at
	`

	var created models.TransportProfile
	err = tx.QueryRowContext(ctx, query,
		p.UserID, p.Name, p.TransportMethod, p.CharacterID, p.CargoM3,
		p.RatePerM3PerJump, p.CollateralRate, p.CollateralPriceBasis,
		p.FuelTypeID, p.FuelPerLY, p.FuelConservationLevel,
		p.RoutePreference, p.IsDefault,
	).Scan(
		&created.ID, &created.UserID, &created.Name, &created.TransportMethod, &created.CharacterID,
		&created.CargoM3, &created.RatePerM3PerJump, &created.CollateralRate, &created.CollateralPriceBasis,
		&created.FuelTypeID, &created.FuelPerLY, &created.FuelConservationLevel,
		&created.RoutePreference, &created.IsDefault, &created.CreatedAt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create transport profile")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit transport profile")
	}

	return &created, nil
}

func (r *TransportProfiles) Update(ctx context.Context, p *models.TransportProfile) (*models.TransportProfile, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// If setting as default, clear existing defaults for this method
	if p.IsDefault {
		_, err = tx.ExecContext(ctx, `
			update transport_profiles set is_default = false
			where user_id = $1 and transport_method = $2 and is_default = true and id != $3
		`, p.UserID, p.TransportMethod, p.ID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to clear existing defaults")
		}
	}

	query := `
		update transport_profiles
		set name = $3, transport_method = $4, character_id = $5, cargo_m3 = $6,
		    rate_per_m3_per_jump = $7, collateral_rate = $8, collateral_price_basis = $9,
		    fuel_type_id = $10, fuel_per_ly = $11, fuel_conservation_level = $12,
		    route_preference = $13, is_default = $14
		where id = $1 and user_id = $2
		returning id, user_id, name, transport_method, character_id,
		          cargo_m3, rate_per_m3_per_jump, collateral_rate, collateral_price_basis,
		          fuel_type_id, fuel_per_ly, fuel_conservation_level,
		          route_preference, is_default, created_at
	`

	var updated models.TransportProfile
	err = tx.QueryRowContext(ctx, query,
		p.ID, p.UserID, p.Name, p.TransportMethod, p.CharacterID, p.CargoM3,
		p.RatePerM3PerJump, p.CollateralRate, p.CollateralPriceBasis,
		p.FuelTypeID, p.FuelPerLY, p.FuelConservationLevel,
		p.RoutePreference, p.IsDefault,
	).Scan(
		&updated.ID, &updated.UserID, &updated.Name, &updated.TransportMethod, &updated.CharacterID,
		&updated.CargoM3, &updated.RatePerM3PerJump, &updated.CollateralRate, &updated.CollateralPriceBasis,
		&updated.FuelTypeID, &updated.FuelPerLY, &updated.FuelConservationLevel,
		&updated.RoutePreference, &updated.IsDefault, &updated.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to update transport profile")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit transport profile update")
	}

	return &updated, nil
}

func (r *TransportProfiles) Delete(ctx context.Context, id, userID int64) error {
	result, err := r.db.ExecContext(ctx, `
		delete from transport_profiles where id = $1 and user_id = $2
	`, id, userID)
	if err != nil {
		return errors.Wrap(err, "failed to delete transport profile")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected for delete")
	}
	if rows == 0 {
		return errors.New("transport profile not found")
	}

	return nil
}
