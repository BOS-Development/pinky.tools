package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type PlayerCorporation struct {
	ID              int64
	UserID          int64
	Name            string
	EsiToken        string
	EsiRefreshToken string
	EsiExpiresOn    time.Time
	EsiScopes       string
}

type PlayerCorporations struct {
	db *sql.DB
}

func NewPlayerCorporations(db *sql.DB) *PlayerCorporations {
	return &PlayerCorporations{
		db: db,
	}
}

func (r *PlayerCorporations) Upsert(ctx context.Context, corp PlayerCorporation) error {
	upsertQuery := `
insert into
	player_corporations
	(
		id,
		user_id,
		name,
		esi_token,
		esi_refresh_token,
		esi_token_expires_on,
		esi_scopes
	)
	values
		($1,$2,$3,$4,$5,$6,$7)
on conflict
	(id, user_id)
do update set
	name = EXCLUDED.name,
	esi_token = EXCLUDED.esi_token,
	esi_refresh_token = EXCLUDED.esi_refresh_token,
	esi_token_expires_on = EXCLUDED.esi_token_expires_on,
	esi_scopes = EXCLUDED.esi_scopes;`

	_, err := r.db.ExecContext(ctx, upsertQuery, corp.ID, corp.UserID, corp.Name, corp.EsiToken, corp.EsiRefreshToken, corp.EsiExpiresOn, corp.EsiScopes)
	if err != nil {
		return errors.Wrap(err, "failed to execute player corporation upsert")
	}
	return nil
}

func (r *PlayerCorporations) UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error {
	_, err := r.db.ExecContext(ctx, `
update player_corporations set
	esi_token = $1,
	esi_refresh_token = $2,
	esi_token_expires_on = $3
where
	id = $4 and user_id = $5;
	`, token, refreshToken, expiresOn, id, userID)
	if err != nil {
		return errors.Wrap(err, "failed to update corporation tokens")
	}
	return nil
}

func (r *PlayerCorporations) Get(ctx context.Context, user int64) ([]PlayerCorporation, error) {
	query := `
select
	id,
	user_id,
	name,
	esi_token,
	esi_refresh_token,
	esi_token_expires_on,
	esi_scopes
from
	player_corporations
where
	user_id=$1;`

	rows, err := r.db.QueryContext(ctx, query, user)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query for user corps")
	}
	defer rows.Close()

	corps := []PlayerCorporation{}
	for rows.Next() {
		var corp PlayerCorporation
		err = rows.Scan(&corp.ID, &corp.UserID, &corp.Name, &corp.EsiToken, &corp.EsiRefreshToken, &corp.EsiExpiresOn, &corp.EsiScopes)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan player corporation row")
		}
		corps = append(corps, corp)
	}

	return corps, nil
}

func (r *PlayerCorporations) GetDivisions(ctx context.Context, corp, user int64) (*models.CorporationDivisions, error) {
	query := `
select
	division_number,
	division_type,
	name
from
	corporation_divisions
where
	corporation_id=$1 and
	user_id=$2;`

	rows, err := r.db.QueryContext(ctx, query, corp, user)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query corporation divisions")
	}
	defer rows.Close()

	divisions := &models.CorporationDivisions{
		Hanger: map[int]string{},
		Wallet: map[int]string{},
	}

	for rows.Next() {
		var divisionNum int
		var divisionType string
		var name string

		err = rows.Scan(&divisionNum, &divisionType, &name)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan corporation division row")
		}

		switch divisionType {
		case "hangar":
			divisions.Hanger[divisionNum] = name
		case "wallet":
			divisions.Wallet[divisionNum] = name
		}
	}

	return divisions, nil
}

func (r *PlayerCorporations) UpsertDivisions(ctx context.Context, corp, user int64, divisions *models.CorporationDivisions) error {
	if divisions == nil {
		return nil
	}

	if len(divisions.Hanger) == 0 && len(divisions.Wallet) == 0 {
		return nil
	}

	upsertQuery := `
insert into
	corporation_divisions
	(
		corporation_id,
		user_id,
		division_number,
		division_type,
		name
	)
	values
	($1, $2, $3, $4, $5)
on conflict
	(corporation_id, user_id, division_number, division_type)
do update set
	name = EXCLUDED.name;
	`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction for corporation division upsert")
	}
	defer tx.Rollback()

	smt, err := tx.PrepareContext(ctx, upsertQuery)
	if err != nil {
		return errors.Wrap(err, "failed to prepare for corporation division upsert")
	}

	for divisionNum, name := range divisions.Hanger {
		_, err := smt.ExecContext(ctx, corp, user, divisionNum, "hangar", name)
		if err != nil {
			return errors.Wrap(err, "failed to execute corporation hangar division upsert")
		}
	}

	for divisionNum, name := range divisions.Wallet {
		_, err := smt.ExecContext(ctx, corp, user, divisionNum, "wallet", name)
		if err != nil {
			return errors.Wrap(err, "failed to execute corporation wallet division upsert")
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit corporation division transaction")
	}

	return nil
}
