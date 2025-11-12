package postgres

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
	l  *slog.Logger
}

func NewRepository(db *sqlx.DB, l *slog.Logger) *Repository {
	return &Repository{
		db: db,
		l:  l,
	}
}
