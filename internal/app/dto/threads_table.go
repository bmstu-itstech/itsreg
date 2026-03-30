package dto

import "time"

type ThreadsTableRow struct {
	ID        string
	EntryKey  string
	UserID    int64
	Username  string
	Timestamp time.Time
	Answers   []string
}

type ThreadsTableHead struct {
	Headers []string
}

type ThreadsTable struct {
	Head ThreadsTableHead
	Body []ThreadsTableRow
}
