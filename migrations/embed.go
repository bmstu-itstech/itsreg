package migrations

import "embed"

// FS contains SQL migration files for tests and local tooling.
//
//go:embed *.sql
var FS embed.FS
