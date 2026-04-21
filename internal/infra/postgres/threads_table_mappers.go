package postgres

import "github.com/bmstu-itstech/itsreg/internal/app/dto"

func groupThreadsAnswers(rows []threadTableRow) dto.ThreadsTable {
	if len(rows) == 0 {
		return dto.ThreadsTable{}
	}

	// Первым проходом считываем заголовок. Это нужно для того, чтобы знать, сколько
	// в каждой нити (строке) ячеек
	head := dto.ThreadsTableHead{
		Headers: make([]string, 0),
	}

	// Гарантировано есть
	head.Headers = append(head.Headers, rows[0].Header)
	firstID := rows[0]

	// Считываем ровно один тред, чтобы получить все возможные заголовки
	for _, row := range rows[1:] {
		if row.ID != firstID.ID {
			break
		}
		head.Headers = append(head.Headers, row.Header)
	}
	columns := len(head.Headers)

	body := make([]dto.ThreadsTableRow, 0, len(rows)/columns)
	var data dto.ThreadsTableRow
	for i, row := range rows {
		if i%columns == 0 {
			data.ID = row.ID
			data.EntryKey = row.Key
			data.UserID = row.UserID
			data.Username = row.Username
			data.Timestamp = row.Timestamp
			data.Answers = make([]string, 0, columns)
		}
		data.Answers = append(data.Answers, emptyOnNil(row.Value))
		if (i+1)%columns == 0 {
			body = append(body, data)
		}
	}

	return dto.ThreadsTable{
		Head: head,
		Body: body,
	}
}

func emptyOnNil(opt *string) string {
	if opt == nil {
		return ""
	}
	return *opt
}
