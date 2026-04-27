package apiv3

import (
	"github.com/bmstu-itstech/itsreg/internal/app/dto"
)

func predicateToAPI(d dto.Predicate) Predicate {
	switch d.Type {
	case string(Always):
		p := Predicate{}
		_ = p.FromAlwaysPredicate(AlwaysPredicate{Type: Always})
		return p

	case string(Exact):
		p := Predicate{}
		_ = p.FromExactPredicate(ExactPredicate{
			Type: Exact,
			Text: d.Data,
		})
		return p

	case string(Regex):
		p := Predicate{}
		_ = p.FromRegexPredicate(RegexPredicate{
			Type:    Regex,
			Pattern: d.Data,
		})
		return p

	default:
		return Predicate{}
	}
}

func edgeToAPI(d dto.Edge) Edge {
	return Edge{
		Operation: EdgeOperation(d.Operation),
		Predicate: predicateToAPI(d.Predicate),
		To:        d.To,
	}
}

func edgesToAPI(ds []dto.Edge) []Edge {
	res := make([]Edge, len(ds))
	for i, d := range ds {
		res[i] = edgeToAPI(d)
	}
	return res
}

func messagesToAPI(ds []dto.Message) []Message {
	res := make([]Message, len(ds))
	for i, d := range ds {
		res[i] = Message{
			Text: d.Text,
		}
	}
	return res
}

func nodeToAPI(d dto.Node) Node {
	return Node{
		Edges:    nilOnEmptySlice(edgesToAPI(d.Edges)),
		Messages: messagesToAPI(d.Messages),
		Options:  nilOnEmptySlice(d.Options),
		State:    d.State,
		Title:    d.Title,
	}
}

func nodesToAPI(ds []dto.Node) []Node {
	res := make([]Node, len(ds))
	for i, d := range ds {
		res[i] = nodeToAPI(d)
	}
	return res
}

func entryToAPI(d dto.Entry) Entry {
	return Entry{
		Key:   d.Key,
		Start: d.Start,
	}
}

func entriesToAPI(ds []dto.Entry) []Entry {
	res := make([]Entry, len(ds))
	for i, d := range ds {
		res[i] = entryToAPI(d)
	}
	return res
}

func scriptToAPI(d dto.Script) Script {
	return Script{
		CreatedAt: &d.CreatedAt,
		Desc:      d.Desc,
		Entries:   entriesToAPI(d.Entries),
		Id:        &d.ID,
		Nodes:     nodesToAPI(d.Nodes),
		UpdatedAt: &d.UpdatedAt,
	}
}

func scriptsToAPI(ds []dto.Script) []Script {
	res := make([]Script, len(ds))
	for i, d := range ds {
		res[i] = scriptToAPI(d)
	}
	return res
}

func botToAPI(d dto.Bot) Bot {
	return Bot{
		CreatedAt: d.CreatedAt,
		Desc:      d.Desc,
		Id:        d.ID,
		OwnerID:   d.OwnerID,
		ScriptID:  d.ScriptID,
		UpdatedAt: d.UpdatedAt,
	}
}

func botsToAPI(ds []dto.Bot) []Bot {
	res := make([]Bot, len(ds))
	for i, d := range ds {
		res[i] = botToAPI(d)
	}
	return res
}

func ownedRunToAPI(d dto.OwnedRun) Run {
	return Run{
		BotID:     d.BotID,
		ErrorMsg:  d.ErrorMsg,
		Id:        d.ID,
		StartedAt: d.StartedAt,
		Status:    RunStatus(d.Status),
		StoppedAt: d.StoppedAt,
	}
}

func runToAPI(d dto.Run) Run {
	return Run{
		BotID:     d.BotID,
		ErrorMsg:  d.ErrorMsg,
		Id:        d.ID,
		StartedAt: d.StartedAt,
		Status:    RunStatus(d.Status),
		StoppedAt: d.StoppedAt,
	}
}

func runsToAPI(ds []dto.Run) []Run {
	res := make([]Run, len(ds))
	for i, d := range ds {
		res[i] = runToAPI(d)
	}
	return res
}

func userMailingResultToAPI(d dto.UserMailingResult) UserMailingResult {
	return UserMailingResult{
		ErrorMsg: d.ErrorMsg,
		Success:  d.Success,
		UserID:   d.UserID,
	}
}

func userMailingResultsToAPI(ds []dto.UserMailingResult) []UserMailingResult {
	res := make([]UserMailingResult, len(ds))
	for i, d := range ds {
		res[i] = userMailingResultToAPI(d)
	}
	return res
}

func mailingToAPI(d dto.Mailing) Mailing {
	return Mailing{
		BotID:        d.BotID,
		CompletedAt:  d.CompletedAt,
		CreatedAt:    d.CreatedAt,
		EntryKey:     d.EntryKey,
		FailCount:    d.FailCount,
		Id:           d.ID,
		Name:         d.Name,
		PendingCount: d.PendingCount,
		Recipients:   d.Recipients,
		Results:      userMailingResultsToAPI(d.Results),
		StartedAt:    d.StartedAt,
		Status:       MailingStatus(d.Status),
		SuccessCount: d.SuccessCount,
		TotalCount:   d.TotalCount,
	}
}

func mailingsToAPI(ds []dto.Mailing) []Mailing {
	res := make([]Mailing, len(ds))
	for i, d := range ds {
		res[i] = mailingToAPI(d)
	}
	return res
}

func ownedMailingToAPI(d dto.OwnedMailing) Mailing {
	return Mailing{
		BotID:        d.BotID,
		CompletedAt:  d.CompletedAt,
		CreatedAt:    d.CreatedAt,
		EntryKey:     d.EntryKey,
		FailCount:    d.FailCount,
		Id:           d.ID,
		Name:         d.Name,
		PendingCount: d.PendingCount,
		Recipients:   d.Recipients,
		Results:      userMailingResultsToAPI(d.Results),
		StartedAt:    d.StartedAt,
		Status:       MailingStatus(d.Status),
		SuccessCount: d.SuccessCount,
		TotalCount:   d.TotalCount,
	}
}
