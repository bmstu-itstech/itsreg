package postgres

import (
	"fmt"
	"time"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

func operationToString(op bots.Operation) string {
	switch op.(type) {
	case bots.NoOp:
		return "noop"
	case bots.SaveOp:
		return "save"
	case bots.AppendOp:
		return "append"
	default:
		// - Кабум?
		// - Да Рико, кабум!
		panic("invalid predicate type")
	}
}

func operationFromString(s string) (bots.Operation, error) {
	switch s {
	case "noop":
		return bots.NoOp{}, nil
	case "save":
		return bots.SaveOp{}, nil
	case "append":
		return bots.AppendOp{}, nil
	default:
		return nil, fmt.Errorf("invalid operation %s, expected one of ['noop', 'save', 'append']", s)
	}
}

func predicateToStrings(p bots.Predicate) (string, string) {
	switch p := p.(type) {
	case bots.AlwaysTruePredicate:
		return "always", ""
	case bots.ExactMatchPredicate:
		return "exact", p.Text()
	case bots.RegexMatchPredicate:
		return "regexp", p.Pattern()
	default:
		// - Кабум?
		// - Да Рико, кабум!
		panic("invalid predicate type")
	}
}

func predicateFromStrings(ptype string, pdata string) (bots.Predicate, error) {
	switch ptype {
	case "always":
		return bots.AlwaysTruePredicate{}, nil
	case "exact":
		return bots.NewExactMatchPredicate(pdata)
	case "regexp":
		return bots.NewRegexMatchPredicate(pdata)
	default:
		return nil, fmt.Errorf("invalid predicate type %s, expected one of ['always', 'exact', 'regexp']", ptype)
	}
}

func botToRow(bot *bots.Bot) botRow {
	return botRow{
		ID:        string(bot.ID()),
		Token:     string(bot.Token()),
		Author:    int64(bot.Author()),
		Enabled:   bot.Enabled(),
		CreatedAt: bot.CreatedAt().In(time.UTC),
	}
}

func entryToRow(botID bots.BotID, entry bots.Entry) entryRow {
	return entryRow{
		BotID: string(botID),
		Key:   string(entry.Key()),
		Start: entry.Start().Int(),
	}
}

func entriesToRows(botID bots.BotID, entries []bots.Entry) []entryRow {
	res := make([]entryRow, len(entries))
	for i, entry := range entries {
		res[i] = entryToRow(botID, entry)
	}
	return res
}

func nodeToRow(botID bots.BotID, node bots.Node) nodeRow {
	return nodeRow{
		BotID: string(botID),
		State: node.State().Int(),
		Title: node.Title(),
	}
}

func nodesToRows(botID bots.BotID, nodes []bots.Node) []nodeRow {
	res := make([]nodeRow, len(nodes))
	for i, node := range nodes {
		res[i] = nodeToRow(botID, node)
	}
	return res
}

func edgeToRow(botID bots.BotID, state bots.State, edge bots.Edge) edgeRow {
	ptype, pdata := predicateToStrings(edge.Predicate)
	return edgeRow{
		BotID:     string(botID),
		State:     state.Int(),
		ToState:   edge.To().Int(),
		Operation: operationToString(edge.Operation()),
		PredType:  ptype,
		PredData:  pdata,
	}
}

func edgesToRows(botID bots.BotID, state bots.State, edges []bots.Edge) []edgeRow {
	res := make([]edgeRow, len(edges))
	for i, edge := range edges {
		res[i] = edgeToRow(botID, state, edge)
	}
	return res
}

func messageToRow(botID bots.BotID, state bots.State, msg bots.Message) messageRow {
	return messageRow{
		BotID: string(botID),
		State: state.Int(),
		Text:  msg.Text(),
	}
}

func messagesToRows(botID bots.BotID, state bots.State, msgs []bots.Message) []messageRow {
	res := make([]messageRow, len(msgs))
	for i, msg := range msgs {
		res[i] = messageToRow(botID, state, msg)
	}
	return res
}

func optionToRow(botID bots.BotID, state bots.State, opt bots.Option) optionRow {
	return optionRow{
		BotID: string(botID),
		State: state.Int(),
		Text:  opt.String(),
	}
}

func optionsToRows(botID bots.BotID, state bots.State, opts []bots.Option) []optionRow {
	res := make([]optionRow, len(opts))
	for i, opt := range opts {
		res[i] = optionToRow(botID, state, opt)
	}
	return res
}

func participantToRow(prt *bots.Participant) participantRow {
	var activeThreadID *string
	if thread := prt.ActiveThread(); thread != nil {
		s := string(thread.ID())
		activeThreadID = &s
	}
	return participantRow{
		BotID:        string(prt.ID().BotID()),
		UserID:       int64(prt.ID().UserID()),
		ActiveThread: activeThreadID,
	}
}

func threadToRow(botID bots.BotID, userID bots.UserID, thread *bots.Thread) threadRow {
	return threadRow{
		ID:        string(thread.ID()),
		BotID:     string(botID),
		UserID:    int64(userID),
		Key:       string(thread.Key()),
		State:     thread.State().Int(),
		StartedAt: thread.StartedAt(),
	}
}

func answerToRow(threadID bots.ThreadID, state bots.State, msg bots.Message) answerRow {
	return answerRow{
		ThreadID: string(threadID),
		State:    state.Int(),
		Text:     msg.Text(),
	}
}

func answersToRows(threadID bots.ThreadID, answers map[bots.State]bots.Message) []answerRow {
	res := make([]answerRow, 0, len(answers))
	for state, answer := range answers {
		res = append(res, answerToRow(threadID, state, answer))
	}
	return res
}
