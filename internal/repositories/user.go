package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

type User struct {
	ID   int64
	Name string
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Get(ctx context.Context, id int64) (*User, error) {
	rows, err := r.db.QueryContext(ctx, "select id, name from users where id=$1", id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user from database")
	}
	defer rows.Close()

	ok := rows.Next()
	if !ok {
		return nil, nil
	}

	var user User
	err = rows.Scan(&user.ID, &user.Name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to scan user")
	}

	return &user, nil
}

func (r *UserRepository) GetUserName(ctx context.Context, userID int64) (string, error) {
	var name string
	err := r.db.QueryRowContext(ctx, "select name from users where id=$1", userID).Scan(&name)
	if err != nil {
		return "", errors.Wrap(err, "failed to get user name")
	}
	return name, nil
}

func (r *UserRepository) GetAllIDs(ctx context.Context) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx, "select id from users")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all user IDs from database")
	}
	defer rows.Close()

	ids := []int64{}
	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan user ID")
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (r *UserRepository) UpdateAssetsLastUpdated(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, "update users set assets_last_updated_at = now() where id = $1", userID)
	if err != nil {
		return errors.Wrap(err, "failed to update assets_last_updated_at")
	}
	return nil
}

func (r *UserRepository) GetAssetsLastUpdated(ctx context.Context, userID int64) (*time.Time, error) {
	var lastUpdated *time.Time
	err := r.db.QueryRowContext(ctx, "select assets_last_updated_at from users where id = $1", userID).Scan(&lastUpdated)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get assets_last_updated_at")
	}
	return lastUpdated, nil
}

func (r *UserRepository) Add(ctx context.Context, user *User) error {
	_, err := r.db.ExecContext(ctx, `
insert into
	users
	(id, name)
	values ($1, $2);
	`, user.ID, user.Name)
	if err != nil {
		return errors.Wrap(err, "failed to insert user into database")
	}

	return nil
}
