package postgres

type botMetaRow struct {
	ID       string `db:"id"`
	OwnerID  int64  `db:"owner_id"`
	ScriptID string `db:"script_id"`
	Token    string `db:"token"`
	Deleted  bool   `db:"deleted"`
}
