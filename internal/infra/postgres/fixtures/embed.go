package fixtures

import "embed"

// FS contains SQL fixtures for testing the Postgres repository.
//
//go:embed *.sql
var FS embed.FS
