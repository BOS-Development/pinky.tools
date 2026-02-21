package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type PiTaxConfig struct {
	db *sql.DB
}

func NewPiTaxConfig(db *sql.DB) *PiTaxConfig {
	return &PiTaxConfig{db: db}
}

func (r *PiTaxConfig) GetForUser(ctx context.Context, userID int64) ([]*models.PiTaxConfig, error) {
	query := `
		SELECT id, user_id, planet_id, tax_rate
		FROM pi_tax_config
		WHERE user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query pi tax configs")
	}
	defer rows.Close()

	configs := []*models.PiTaxConfig{}
	for rows.Next() {
		var config models.PiTaxConfig
		err = rows.Scan(
			&config.ID,
			&config.UserID,
			&config.PlanetID,
			&config.TaxRate,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan pi tax config")
		}
		configs = append(configs, &config)
	}

	return configs, nil
}

func (r *PiTaxConfig) Upsert(ctx context.Context, config *models.PiTaxConfig) error {
	query := `
		INSERT INTO pi_tax_config (user_id, planet_id, tax_rate)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, planet_id)
		DO UPDATE SET tax_rate = $3
	`

	_, err := r.db.ExecContext(ctx, query,
		config.UserID,
		config.PlanetID,
		config.TaxRate,
	)
	if err != nil {
		return errors.Wrap(err, "failed to upsert pi tax config")
	}

	return nil
}

func (r *PiTaxConfig) Delete(ctx context.Context, userID int64, planetID *int64) error {
	if planetID == nil {
		_, err := r.db.ExecContext(ctx, `
			DELETE FROM pi_tax_config
			WHERE user_id = $1 AND planet_id IS NULL
		`, userID)
		if err != nil {
			return errors.Wrap(err, "failed to delete global pi tax config")
		}
		return nil
	}

	_, err := r.db.ExecContext(ctx, `
		DELETE FROM pi_tax_config
		WHERE user_id = $1 AND planet_id = $2
	`, userID, *planetID)
	if err != nil {
		return errors.Wrap(err, "failed to delete pi tax config")
	}

	return nil
}
