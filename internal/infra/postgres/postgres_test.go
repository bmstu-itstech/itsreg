package postgres_test

import (
	"github.com/bmstu-itstech/itsreg-bots/internal/infra/postgres"
	"github.com/bmstu-itstech/itsreg-bots/pkg/logs"
	"github.com/bmstu-itstech/itsreg-bots/pkg/tests"
)

func setupRepository() (*postgres.Repository, func()) {
	db := tests.ConnectPostgresDB()
	return postgres.NewRepository(db, logs.DefaultLogger()), func() {
		_ = db.Close()
	}
}
