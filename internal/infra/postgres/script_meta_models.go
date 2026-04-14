package postgres

type scriptMetaRow struct {
	ID      string `db:"id"`
	OwnerID int64  `db:"owner_id"`
	Desc    string `db:"description"`
	Deleted bool   `db:"deleted"`
}
