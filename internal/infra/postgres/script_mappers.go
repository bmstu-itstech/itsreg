package postgres

import (
	"fmt"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
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
		return "regex", p.Pattern()
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
	case "regex":
		return bots.NewRegexMatchPredicate(pdata)
	default:
		return nil, fmt.Errorf("invalid predicate type %s, expected one of ['always', 'exact', 'regex']", ptype)
	}
}

func edgesFromRows(rows []edgeRow) ([]bots.Edge, error) {
	res := make([]bots.Edge, len(rows))
	for i, row := range rows {
		pred, err := predicateFromStrings(row.PredType, row.PredData)
		if err != nil {
			return nil, err
		}
		op, err := operationFromString(row.Operation)
		if err != nil {
			return nil, err
		}
		toState, err := bots.NewState(row.ToState)
		if err != nil {
			return nil, err
		}
		res[i] = bots.NewEdge(pred, toState, op)
	}
	return res, nil
}

func messagesFromRows(rows []messageRow) ([]bots.Message, error) {
	res := make([]bots.Message, len(rows))
	for i, row := range rows {
		m, err := bots.NewMessage(row.Text)
		if err != nil {
			return nil, err
		}
		res[i] = m
	}
	return res, nil
}

func optionsFromRows(rows []optionRow) ([]bots.Option, error) {
	res := make([]bots.Option, len(rows))
	for i, row := range rows {
		o, err := bots.NewOption(row.Text)
		if err != nil {
			return nil, err
		}
		res[i] = o
	}
	return res, nil
}

func nodeFromRows(rNode nodeRow, rEdges []edgeRow, rMessages []messageRow, rOptions []optionRow) (bots.Node, error) {
	s, err := bots.NewState(rNode.State)
	if err != nil {
		return bots.Node{}, err
	}

	edges, err := edgesFromRows(rEdges)
	if err != nil {
		return bots.Node{}, err
	}

	messages, err := messagesFromRows(rMessages)
	if err != nil {
		return bots.Node{}, err
	}

	options, err := optionsFromRows(rOptions)
	if err != nil {
		return bots.Node{}, err
	}

	return bots.NewNode(s, rNode.Title, edges, messages, options)
}

func nodesFromRows(
	rNodes []nodeRow,
	rEdges []edgeRow,
	rMessages []messageRow,
	rOptions []optionRow,
) ([]bots.Node, error) {
	mrEdges := groupXByField(rEdges, func(e edgeRow) int { return e.State })
	mrMessages := groupXByField(rMessages, func(e messageRow) int { return e.State })
	mrOptions := groupXByField(rOptions, func(e optionRow) int { return e.State })

	res := make([]bots.Node, len(rNodes))
	for i, node := range rNodes {
		s := node.State
		n, err := nodeFromRows(rNodes[i], mrEdges[s], mrMessages[s], mrOptions[s])
		if err != nil {
			return nil, err
		}
		res[i] = n
	}

	return res, nil
}

func entryFromRow(row entryRow) (bots.Entry, error) {
	s, err := bots.NewState(row.Start)
	if err != nil {
		return bots.Entry{}, err
	}
	return bots.NewEntry(bots.EntryKey(row.Key), s)
}

func entriesFromRows(rows []entryRow) ([]bots.Entry, error) {
	res := make([]bots.Entry, len(rows))
	for i, row := range rows {
		e, err := entryFromRow(row)
		if err != nil {
			return nil, err
		}
		res[i] = e
	}
	return res, nil
}

// groupXByField группирует значения массива ts по ключу, определяемый функцией key.
func groupXByField[K comparable, V any](ts []V, key func(v V) K) map[K][]V {
	res := make(map[K][]V)
	for _, t := range ts {
		if _, ok := res[key(t)]; !ok {
			res[key(t)] = make([]V, 0, 1)
		}
		res[key(t)] = append(res[key(t)], t)
	}
	return res
}

func scriptToRow(s *bots.Script) scriptRow {
	return scriptRow{
		ID:        s.ID().String(),
		OwnerID:   s.OwnerID().Int64(),
		Desc:      s.Desc(),
		CreatedAt: s.CreatedAt(),
		UpdatedAt: s.UpdatedAt(),
		DeletedAt: s.DeletedAt(),
	}
}

func edgeToRow(e bots.Edge, scriptID string, state int, index int) edgeRow {
	o := operationToString(e.Operation())
	pt, pd := predicateToStrings(e.Predicate)
	return edgeRow{
		ScriptID:  scriptID,
		State:     state,
		Index:     index,
		ToState:   e.To().Int(),
		Operation: o,
		PredType:  pt,
		PredData:  pd,
	}
}

func edgesToRows(es []bots.Edge, scriptID string, state int) []edgeRow {
	res := make([]edgeRow, len(es))
	for i, e := range es {
		res[i] = edgeToRow(e, scriptID, state, i+1)
	}
	return res
}

func messagesToRows(ms []bots.Message, scriptID string, state int) []messageRow {
	res := make([]messageRow, len(ms))
	for i, m := range ms {
		res[i] = messageRow{
			ScriptID: scriptID,
			State:    state,
			Index:    i + 1,
			Text:     m.Text(),
		}
	}
	return res
}

func optionsToRows(os []bots.Option, scriptID string, state int) []optionRow {
	res := make([]optionRow, len(os))
	for i, o := range os {
		res[i] = optionRow{
			ScriptID: scriptID,
			State:    state,
			Index:    i + 1,
			Text:     o.String(),
		}
	}
	return res
}

func entryToRow(e bots.Entry, scriptID string) entryRow {
	return entryRow{
		ScriptID: scriptID,
		Key:      e.Key().String(),
		Start:    e.Start().Int(),
	}
}

func entriesToRows(es []bots.Entry, scriptID string) []entryRow {
	res := make([]entryRow, len(es))
	for i, e := range es {
		res[i] = entryToRow(e, scriptID)
	}
	return res
}

func nodeToRow(n bots.Node, scriptID string) nodeRow {
	return nodeRow{
		ScriptID: scriptID,
		State:    n.State().Int(),
		Title:    n.Title(),
	}
}

func nodesToRows(ns []bots.Node, scriptID string) []nodeRow {
	res := make([]nodeRow, len(ns))
	for i, n := range ns {
		res[i] = nodeToRow(n, scriptID)
	}
	return res
}

func decomposeNodeToRows(n bots.Node, scriptID string) ([]edgeRow, []messageRow, []optionRow) {
	ers := edgesToRows(n.Edges(), scriptID, n.State().Int())
	mrs := messagesToRows(n.Messages(), scriptID, n.State().Int())
	ors := optionsToRows(n.Options(), scriptID, n.State().Int())
	return ers, mrs, ors
}

func decomposeNodesToRows(ns []bots.Node, scriptID string) ([]edgeRow, []messageRow, []optionRow) {
	var rEdges []edgeRow
	var rMessages []messageRow
	var rOptions []optionRow

	for _, n := range ns {
		rEn, rMn, rOn := decomposeNodeToRows(n, scriptID)
		rEdges = append(rEdges, rEn...)
		rMessages = append(rMessages, rMn...)
		rOptions = append(rOptions, rOn...)
	}

	return rEdges, rMessages, rOptions
}

func nodesRowsAoSToSoA(rows []nodeRow) nodeRowsSoA {
	soa := nodeRowsSoA{
		ScriptIDs: make([]string, len(rows)),
		States:    make([]int, len(rows)),
		Titles:    make([]string, len(rows)),
	}
	for i, row := range rows {
		soa.ScriptIDs[i] = row.ScriptID
		soa.States[i] = row.State
		soa.Titles[i] = row.Title
	}
	return soa
}

func entriesRowsAoSToSoA(rows []entryRow) entryRowsSoA {
	soa := entryRowsSoA{
		ScriptIDs: make([]string, len(rows)),
		Keys:      make([]string, len(rows)),
		Starts:    make([]int, len(rows)),
	}
	for i, row := range rows {
		soa.ScriptIDs[i] = row.ScriptID
		soa.Keys[i] = row.Key
		soa.Starts[i] = row.Start
	}
	return soa
}

func edgesRowsAoSToSoA(rows []edgeRow) edgeRowsSoA {
	soa := edgeRowsSoA{
		ScriptIDs:  make([]string, len(rows)),
		States:     make([]int, len(rows)),
		Indices:    make([]int, len(rows)),
		ToStates:   make([]int, len(rows)),
		Operations: make([]string, len(rows)),
		PredTypes:  make([]string, len(rows)),
		PredData:   make([]string, len(rows)),
	}
	for i, row := range rows {
		soa.ScriptIDs[i] = row.ScriptID
		soa.States[i] = row.State
		soa.Indices[i] = row.Index
		soa.ToStates[i] = row.ToState
		soa.Operations[i] = row.Operation
		soa.PredTypes[i] = row.PredType
		soa.PredData[i] = row.PredData
	}
	return soa
}

func messagesRowsAoSToSoA(rows []messageRow) messagesRowsSoA {
	soa := messagesRowsSoA{
		ScriptIDs: make([]string, len(rows)),
		States:    make([]int, len(rows)),
		Indices:   make([]int, len(rows)),
		Texts:     make([]string, len(rows)),
	}
	for i, row := range rows {
		soa.ScriptIDs[i] = row.ScriptID
		soa.States[i] = row.State
		soa.Indices[i] = row.Index
		soa.Texts[i] = row.Text
	}
	return soa
}

func optionsRowsAoSToSoA(rows []optionRow) optionRowsSoA {
	soa := optionRowsSoA{
		ScriptIDs: make([]string, len(rows)),
		States:    make([]int, len(rows)),
		Indices:   make([]int, len(rows)),
		Texts:     make([]string, len(rows)),
	}
	for i, row := range rows {
		soa.ScriptIDs[i] = row.ScriptID
		soa.States[i] = row.State
		soa.Indices[i] = row.Index
		soa.Texts[i] = row.Text
	}
	return soa
}
