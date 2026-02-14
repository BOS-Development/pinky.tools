package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

type Character struct {
	ID                int64
	Name              string
	EsiToken          string
	EsiRefreshToken   string
	EsiTokenExpiresOn time.Time
	UserID            int64
}

type CharacterRepository struct {
	db *sql.DB
}

func NewCharacterRepository(db *sql.DB) *CharacterRepository {
	return &CharacterRepository{
		db: db,
	}
}

func (r *CharacterRepository) GetAll(ctx context.Context, baseUserId int64) ([]*Character, error) {
	rows, err := r.db.QueryContext(ctx, `
select
	id,
	name,
	user_id,
	esi_token,
	esi_refresh_token,
	esi_token_expires_on
from
	characters
where
	user_id = $1`, baseUserId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get character from database")
	}
	defer rows.Close()

	characters := []*Character{}

	for rows.Next() {
		var char Character

		err = rows.Scan(&char.ID, &char.Name, &char.UserID, &char.EsiToken, &char.EsiRefreshToken, &char.EsiTokenExpiresOn)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan charater")
		}
		characters = append(characters, &char)
	}

	return characters, nil
}

func (r *CharacterRepository) Get(ctx context.Context, id string) (*Character, error) {
	rows, err := r.db.QueryContext(ctx, "select id, name from characters where character_id = $1", id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get character from database")
	}
	defer rows.Close()

	ok := rows.Next()
	if !ok {
		return nil, nil
	}

	var char Character
	err = rows.Scan(&char)
	if err != nil {
		return nil, errors.Wrap(err, "failed to scan charater")
	}

	return &char, nil
}

func (r *CharacterRepository) Add(ctx context.Context, character *Character) error {
	_, err := r.db.ExecContext(ctx, `
insert into
	characters
		(id, name, user_id, esi_token, esi_refresh_token, esi_token_expires_on)
	values
		($1, $2, $3, $4, $5, $6)
on conflict
	(id, user_id)
do update set
	name = EXCLUDED.name,
	esi_token = EXCLUDED.esi_token,
	esi_refresh_token = EXCLUDED.esi_refresh_token,
	esi_token_expires_on = EXCLUDED.esi_token_expires_on;
	`, character.ID, character.Name, character.UserID, character.EsiToken, character.EsiRefreshToken, character.EsiTokenExpiresOn)
	if err != nil {
		return errors.Wrap(err, "failed to insert character into database")
	}
	return nil
}
