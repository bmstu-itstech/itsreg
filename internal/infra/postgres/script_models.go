package postgres

import (
	"time"
)

type scriptRow struct {
	// PK (ID)
	ID        string     `db:"id"`
	OwnerID   int64      `db:"owner_id"`
	Desc      string     `db:"desc"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

type nodeRow struct {
	// PK (ScriptID, State)
	ScriptID string `db:"script_id"`
	State    int    `db:"state"`
	Title    string `db:"title"`
}

type entryRow struct {
	// PK (ScriptID, Key)
	ScriptID string `db:"script_id"`
	Key      string `db:"key"`
	Start    int    `db:"start"`
}

type edgeRow struct {
	// PK (ScriptID, State, Index)
	ScriptID  string `db:"script_id"`
	State     int    `db:"state"`
	Index     int    `db:"index"`
	ToState   int    `db:"to_state"`
	Operation string `db:"operation"`
	PredType  string `db:"pred_type"`
	PredData  string `db:"pred_data"`
}

type messageRow struct {
	// PK (ScriptID, State, Index)
	ScriptID string `db:"script_id"`
	State    int    `db:"state"`
	Index    int    `db:"index"`
	Text     string `db:"text"`
}

type optionRow struct {
	// PK (ScriptID, State, Index)
	ScriptID string `db:"script_id"`
	State    int    `db:"state"`
	Index    int    `db:"index"`
	Text     string `db:"text"`
}

type nodeRowsSoA struct {
	ScriptIDs []string `db:"script_ids"`
	States    []int    `db:"states"`
	Titles    []string `db:"titles"`
}

type entryRowsSoA struct {
	ScriptIDs []string `db:"script_ids"`
	Keys      []string `db:"keys"`
	Starts    []int    `db:"starts"`
}

type edgeRowsSoA struct {
	ScriptIDs  []string `db:"script_ids"`
	States     []int    `db:"states"`
	Indices    []int    `db:"indices"`
	ToStates   []int    `db:"to_states"`
	Operations []string `db:"operations"`
	PredTypes  []string `db:"pred_types"`
	PredData   []string `db:"pred_data"`
}

type messagesRowsSoA struct {
	ScriptIDs []string `db:"script_ids"`
	States    []int    `db:"states"`
	Indices   []int    `db:"indices"`
	Texts     []string `db:"texts"`
}

type optionRowsSoA struct {
	ScriptIDs []string `db:"script_ids"`
	States    []int    `db:"states"`
	Indices   []int    `db:"indices"`
	Texts     []string `db:"texts"`
}
