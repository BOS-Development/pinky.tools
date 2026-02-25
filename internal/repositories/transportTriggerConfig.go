package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type TransportTriggerConfig struct {
	db *sql.DB
}

func NewTransportTriggerConfig(db *sql.DB) *TransportTriggerConfig {
	return &TransportTriggerConfig{db: db}
}

func (r *TransportTriggerConfig) GetByUser(ctx context.Context, userID int64) ([]*models.TransportTriggerConfig, error) {
	query := `
		select user_id, trigger_type, default_fulfillment, allowed_fulfillments,
		       default_profile_id, default_method, courier_rate_per_m3, courier_collateral_rate
		from transport_trigger_config
		where user_id = $1
		order by trigger_type asc
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query transport trigger configs")
	}
	defer rows.Close()

	configs := []*models.TransportTriggerConfig{}
	for rows.Next() {
		var c models.TransportTriggerConfig
		if err := rows.Scan(
			&c.UserID, &c.TriggerType, &c.DefaultFulfillment, pq.Array(&c.AllowedFulfillments),
			&c.DefaultProfileID, &c.DefaultMethod, &c.CourierRatePerM3, &c.CourierCollateralRate,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan transport trigger config")
		}
		configs = append(configs, &c)
	}

	return configs, nil
}

func (r *TransportTriggerConfig) Upsert(ctx context.Context, c *models.TransportTriggerConfig) (*models.TransportTriggerConfig, error) {
	query := `
		insert into transport_trigger_config
			(user_id, trigger_type, default_fulfillment, allowed_fulfillments,
			 default_profile_id, default_method, courier_rate_per_m3, courier_collateral_rate)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
		on conflict (user_id, trigger_type) do update set
			default_fulfillment = EXCLUDED.default_fulfillment,
			allowed_fulfillments = EXCLUDED.allowed_fulfillments,
			default_profile_id = EXCLUDED.default_profile_id,
			default_method = EXCLUDED.default_method,
			courier_rate_per_m3 = EXCLUDED.courier_rate_per_m3,
			courier_collateral_rate = EXCLUDED.courier_collateral_rate
		returning user_id, trigger_type, default_fulfillment, allowed_fulfillments,
		          default_profile_id, default_method, courier_rate_per_m3, courier_collateral_rate
	`

	var result models.TransportTriggerConfig
	err := r.db.QueryRowContext(ctx, query,
		c.UserID, c.TriggerType, c.DefaultFulfillment, pq.Array(c.AllowedFulfillments),
		c.DefaultProfileID, c.DefaultMethod, c.CourierRatePerM3, c.CourierCollateralRate,
	).Scan(
		&result.UserID, &result.TriggerType, &result.DefaultFulfillment, pq.Array(&result.AllowedFulfillments),
		&result.DefaultProfileID, &result.DefaultMethod, &result.CourierRatePerM3, &result.CourierCollateralRate,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to upsert transport trigger config")
	}

	return &result, nil
}
