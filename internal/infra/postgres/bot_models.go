package postgres

import (
	"time"
)

type botRow struct {
	// PK (ID)
	ID        string     `db:"id"`
	OwnerID   int64      `db:"owner_id"`
	ScriptID  string     `db:"script_id"`
	Token     string     `db:"token"`
	Desc      string     `db:"desc"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}
