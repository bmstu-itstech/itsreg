package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/zhikh23/pgutils"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

type PostgresUsernameRepository struct {
	db *sqlx.DB
}

func NewPostgresUsernameRepository(db *sqlx.DB) *PostgresUsernameRepository {
	return &PostgresUsernameRepository{
		db: db,
	}
}

func (r *PostgresUsernameRepository) Upsert(ctx context.Context, id bots.UserID, username bots.Username) error {
	return pgutils.RequireAffected(pgutils.Exec(ctx, r.db,
		`INSERT INTO
			usernames (
			    user_id,
				username
			)
		 VALUES 
			($1, $2)
		 ON CONFLICT 
			(user_id) 
		 DO UPDATE 
		 SET
			username = $2`,
		id,
		username,
	))
}

func (r *PostgresUsernameRepository) Username(ctx context.Context, id bots.UserID) (bots.Username, error) {
	var row usernameRow
	err := pgutils.Get(ctx, r.db, &row,
		`SELECT 
			user_id,
			username
		 FROM
			usernames
		 WHERE
			user_id = $1`,
		id,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return "", bots.ErrUsernameNotFound
	}

	return bots.Username(row.Username), nil
}

type usernameRow struct {
	UserID   int64  `db:"user_id"`
	Username string `db:"username"`
}
