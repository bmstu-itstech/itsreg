package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/zhikh23/pgutils"
)

func (r *Repository) updateUsername(
	ctx context.Context,
	ec sqlx.ExtContext,
	userID int64,
	username string,
) error {
	return pgutils.RequireAffected(pgutils.Exec(ctx, ec, `
		INSERT INTO users 
			(id, username, updated_at) 
		VALUES 
			($1, $2, now())	
		ON CONFLICT (id)
			DO UPDATE SET 
				username = $2, 
				updated_at = now()
		`, userID, username,
	))
}
