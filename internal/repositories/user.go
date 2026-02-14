package repositories

import (
	"context"
	"database/sql"

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
