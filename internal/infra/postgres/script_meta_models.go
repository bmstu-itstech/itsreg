package postgres

type scriptMetaRow struct {
	ID      string `db:"id"`
	OwnerID int64  `db:"owner_id"`
	Desc    string `db:"desc"`
	Deleted bool   `db:"deleted"`
}
