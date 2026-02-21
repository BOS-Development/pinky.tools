package repositories

import (
	"context"
	"database/sql"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type PiLaunchpadLabels struct {
	db *sql.DB
}

func NewPiLaunchpadLabels(db *sql.DB) *PiLaunchpadLabels {
	return &PiLaunchpadLabels{db: db}
}

func (r *PiLaunchpadLabels) GetForUser(ctx context.Context, userID int64) ([]*models.PiLaunchpadLabel, error) {
	query := `
		SELECT user_id, character_id, planet_id, pin_id, label
		FROM pi_launchpad_labels
		WHERE user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query pi launchpad labels")
	}
	defer rows.Close()

	labels := []*models.PiLaunchpadLabel{}
	for rows.Next() {
		var label models.PiLaunchpadLabel
		err = rows.Scan(
			&label.UserID,
			&label.CharacterID,
			&label.PlanetID,
			&label.PinID,
			&label.Label,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan pi launchpad label")
		}
		labels = append(labels, &label)
	}

	return labels, nil
}

func (r *PiLaunchpadLabels) Upsert(ctx context.Context, label *models.PiLaunchpadLabel) error {
	query := `
		INSERT INTO pi_launchpad_labels (user_id, character_id, planet_id, pin_id, label)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, character_id, planet_id, pin_id)
		DO UPDATE SET label = $5
	`

	_, err := r.db.ExecContext(ctx, query,
		label.UserID,
		label.CharacterID,
		label.PlanetID,
		label.PinID,
		label.Label,
	)
	if err != nil {
		return errors.Wrap(err, "failed to upsert pi launchpad label")
	}

	return nil
}

func (r *PiLaunchpadLabels) Delete(ctx context.Context, userID, characterID, planetID, pinID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM pi_launchpad_labels
		WHERE user_id = $1 AND character_id = $2 AND planet_id = $3 AND pin_id = $4
	`, userID, characterID, planetID, pinID)
	if err != nil {
		return errors.Wrap(err, "failed to delete pi launchpad label")
	}

	return nil
}
