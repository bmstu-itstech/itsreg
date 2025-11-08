package tests

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
)

func ConnectPostgresDB() *sqlx.DB {
	uri := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		"localhost",
		os.Getenv("POSTGRES_EXTERNAL_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)
	return sqlx.MustConnect("postgres", uri)
}
