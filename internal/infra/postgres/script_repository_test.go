package postgres_test

import (
	"sort"
	"time"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func (s *RepositoryTestSuite) TestScriptRepository_Script_Success() {
	script, err := s.repos.Script(s.ctx, "sc0001")
	s.Require().NoError(err)

	// Assert script
	s.Require().NotNil(script)
	s.Require().Equal(bots.ScriptID("sc0001"), script.ID())
	s.Require().Equal(bots.UserID(1), script.OwnerID())
	s.Require().Equal("Test script sc0001", script.Desc())
	s.Require().Equal(time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC), script.CreatedAt())
	s.Require().Equal(time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC), script.UpdatedAt())
	s.Require().Nil(script.DeletedAt())

	// Assert entries
	entries := script.Entries()
	s.Require().Len(entries, 1)
	entry1 := entries[0]
	s.Require().Equal(bots.EntryKey("start"), entry1.Key())
	s.Require().Equal(bots.MustNewState(1), entry1.Start())

	// Assert nodes
	nodes := script.Nodes()
	s.Require().Len(nodes, 2)
	// Порядок узлов не гарантируется
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].State().Int() < nodes[j].State().Int()
	})
	node1 := nodes[0]
	s.Require().Equal(bots.MustNewState(1), node1.State())
	s.Require().Equal("sc0001:1", node1.Title())
	// Порядок опций гарантирован
	s.Require().Len(node1.Options(), 2)
	opt11 := node1.Options()[0]
	s.Require().Equal("Option 1 for sc0001:1", opt11.String())
	opt12 := node1.Options()[1]
	s.Require().Equal("Option 2 for sc0001:1", opt12.String())
	// Порядок сообщений гарантирован
	s.Require().Len(node1.Messages(), 2)
	msg11 := node1.Messages()[0]
	s.Require().Equal("Message 1 for sc0001:1", msg11.String())
	msg12 := node1.Messages()[1]
	s.Require().Equal("Message 2 for sc0001:1", msg12.String())
	// Порядок рёбер гарантирован
	s.Require().Len(node1.Edges(), 2)
	edge11 := node1.Edges()[0]
	s.Require().Equal(bots.MustNewState(2), edge11.To())
	s.Require().Equal(bots.SaveOp{}, edge11.Operation())
	s.Require().IsType(bots.RegexMatchPredicate{}, edge11.Predicate)
	regexPred := edge11.Predicate.(bots.RegexMatchPredicate)
	s.Require().Equal("^Далее$", regexPred.Pattern())
	edge12 := node1.Edges()[1]
	s.Require().Equal(bots.MustNewState(2), edge12.To())
	s.Require().Equal(bots.AppendOp{}, edge12.Operation())
	s.Require().IsType(bots.AlwaysTruePredicate{}, edge12.Predicate)

	node2 := nodes[1]
	s.Require().Equal(bots.MustNewState(2), node2.State())
	s.Require().Equal("sc0001:2", node2.Title())
	// Порядок опций гарантирован
	s.Require().Len(node2.Options(), 1)
	opt21 := node2.Options()[0]
	s.Require().Equal("Option for sc0001:2", opt21.String())
	// Порядок сообщений гарантирован
	s.Require().Len(node2.Messages(), 1)
	msg21 := node2.Messages()[0]
	s.Require().Equal("Message for sc0001:2", msg21.String())
	// Порядок рёбер гарантирован
	s.Require().Len(node2.Edges(), 1)
	edge21 := node2.Edges()[0]
	s.Require().Equal(bots.MustNewState(1), edge21.To())
	s.Require().Equal(bots.NoOp{}, edge21.Operation())
	s.Require().IsType(bots.ExactMatchPredicate{}, edge21.Predicate)
	exactPred := edge21.Predicate.(bots.ExactMatchPredicate)
	s.Require().Equal("Назад", exactPred.Text())
}

func (s *RepositoryTestSuite) TestScriptRepository_Script_FetchDeleted() {
	script, err := s.repos.Script(s.ctx, "sc0004")
	s.Require().NoError(err)
	s.Require().NotNil(script)
	s.Require().NotNil(script.DeletedAt())
	deletedAt := *script.DeletedAt()
	s.Require().Equal(time.Date(2026, 4, 11, 11, 0, 0, 0, time.UTC), deletedAt)
}

func (s *RepositoryTestSuite) TestScriptRepository_Script_NotFound() {
	script, err := s.repos.Script(s.ctx, "sc000x")
	s.Require().ErrorIs(err, port.ErrScriptNotFound)
	s.Require().Nil(script)
}

func (s *RepositoryTestSuite) TestScriptRepository_ScriptsByOwnerID_Empty() {
	scripts, err := s.repos.ScriptsByOwnerID(s.ctx, bots.UserID(-1))
	s.Require().NoError(err)
	s.Require().Empty(scripts)
}

func (s *RepositoryTestSuite) TestScriptRepository_ScriptsByOwnerID_OneScript() {
	// User(2) имеет две записи в таблице scripts, но одна из них удалена (deleted_at IS NOT NULL)
	scripts, err := s.repos.ScriptsByOwnerID(s.ctx, bots.UserID(2))
	s.Require().NoError(err)
	s.Require().Len(scripts, 1)
	script := scripts[0]
	s.Require().Equal(bots.ScriptID("sc0003"), script.ID())
	s.Require().Equal(bots.UserID(2), script.OwnerID())
	s.Require().Equal("Test script sc0003", script.Desc())
	s.Require().Equal(time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC), script.CreatedAt())
	s.Require().Equal(time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC), script.UpdatedAt())
	s.Require().Nil(script.DeletedAt())
}

func (s *RepositoryTestSuite) TestScriptRepository_ScriptsByOwnerID_MultipleScripts() {
	scripts, err := s.repos.ScriptsByOwnerID(s.ctx, bots.UserID(1))
	s.Require().NoError(err)
	s.Require().Len(scripts, 2)

	// User(2) имеет два сценария, и порядок определяется столбцом updated_at; второй сценарий имеет более позднее
	// время обновления, поэтому он должен быть первым в списке.
	script2 := scripts[0]
	s.Require().Equal(bots.ScriptID("sc0002"), script2.ID())
	s.Require().Equal(bots.UserID(1), script2.OwnerID())
	s.Require().Equal("Test script sc0002", script2.Desc())
	s.Require().Equal(time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC), script2.CreatedAt())
	s.Require().Equal(time.Date(2026, 4, 11, 11, 0, 0, 0, time.UTC), script2.UpdatedAt())
	s.Require().Nil(script2.DeletedAt())

	script1 := scripts[1]
	s.Require().Equal(bots.ScriptID("sc0001"), script1.ID())
	s.Require().Equal(bots.UserID(1), script1.OwnerID())
	s.Require().Equal("Test script sc0001", script1.Desc())
	s.Require().Equal(time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC), script1.CreatedAt())
	s.Require().Equal(time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC), script1.UpdatedAt())
	s.Require().Nil(script1.DeletedAt())
}

func (s *RepositoryTestSuite) TestScriptRepository_SaveScript_Success() {
	want := makeSampleScript()
	err := s.repos.SaveScript(s.ctx, want)
	s.Require().NoError(err)

	got, err := s.repos.Script(s.ctx, want.ID())
	s.Require().NoError(err)

	// Assert script
	s.Require().NotNil(got)
	s.Require().Equal(want.ID(), got.ID())
	s.Require().Equal(actorUserID, got.OwnerID())
	s.Require().Equal("Test script", got.Desc())
	s.Require().WithinDuration(want.CreatedAt(), got.CreatedAt(), time.Second)
	s.Require().WithinDuration(want.UpdatedAt(), got.UpdatedAt(), time.Second)
	s.Require().Nil(got.DeletedAt())

	// Assert entries
	entries := got.Entries()
	s.Require().Len(entries, 1)
	entry1 := entries[0]
	s.Require().Equal(bots.EntryKey("start"), entry1.Key())
	s.Require().Equal(bots.MustNewState(1), entry1.Start())

	// Assert nodes
	nodes := got.Nodes()
	s.Require().Len(nodes, 2)
	// Порядок узлов не гарантируется
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].State().Int() < nodes[j].State().Int()
	})
	node1 := nodes[0]
	s.Require().Equal(bots.MustNewState(1), node1.State())
	s.Require().Equal("Node 1", node1.Title())
	// Порядок опций гарантирован
	s.Require().Len(node1.Options(), 2)
	opt11 := node1.Options()[0]
	s.Require().Equal("Option 1 for Node 1", opt11.String())
	opt12 := node1.Options()[1]
	s.Require().Equal("Option 2 for Node 1", opt12.String())
	// Порядок сообщений гарантирован
	s.Require().Len(node1.Messages(), 2)
	msg11 := node1.Messages()[0]
	s.Require().Equal("Message 1 for Node 1", msg11.String())
	msg12 := node1.Messages()[1]
	s.Require().Equal("Message 2 for Node 1", msg12.String())
	// Порядок рёбер гарантирован
	s.Require().Len(node1.Edges(), 2)
	edge11 := node1.Edges()[0]
	s.Require().Equal(bots.MustNewState(2), edge11.To())
	s.Require().Equal(bots.SaveOp{}, edge11.Operation())
	s.Require().IsType(bots.RegexMatchPredicate{}, edge11.Predicate)
	regexPred := edge11.Predicate.(bots.RegexMatchPredicate)
	s.Require().Equal("^Далее$", regexPred.Pattern())
	edge12 := node1.Edges()[1]
	s.Require().Equal(bots.MustNewState(2), edge12.To())
	s.Require().Equal(bots.AppendOp{}, edge12.Operation())
	s.Require().IsType(bots.AlwaysTruePredicate{}, edge12.Predicate)

	node2 := nodes[1]
	s.Require().Equal(bots.MustNewState(2), node2.State())
	s.Require().Equal("Node 2", node2.Title())
	// Порядок опций гарантирован
	s.Require().Len(node2.Options(), 1)
	opt21 := node2.Options()[0]
	s.Require().Equal("Option 1 for Node 2", opt21.String())
	// Порядок сообщений гарантирован
	s.Require().Len(node2.Messages(), 1)
	msg21 := node2.Messages()[0]
	s.Require().Equal("Message 1 for Node 2", msg21.String())
	// Порядок рёбер гарантирован
	s.Require().Len(node2.Edges(), 1)
	edge21 := node2.Edges()[0]
	s.Require().Equal(bots.MustNewState(1), edge21.To())
	s.Require().Equal(bots.NoOp{}, edge21.Operation())
	s.Require().IsType(bots.ExactMatchPredicate{}, edge21.Predicate)
	exactPred := edge21.Predicate.(bots.ExactMatchPredicate)
	s.Require().Equal("Назад", exactPred.Text())
}

func (s *RepositoryTestSuite) TestScriptRepository_SaveScript_FailedToSaveScriptTwice() {
	script := makeSampleScript()
	err := s.repos.SaveScript(s.ctx, script)
	s.Require().NoError(err)

	err = s.repos.SaveScript(s.ctx, script)
	s.Require().ErrorIs(err, port.ErrScriptAlreadyExists)
}

func (s *RepositoryTestSuite) TestScriptRepository_UpdateScript_NonExistentScript() {
	script := makeSampleScript()
	err := s.repos.UpdateScript(s.ctx, script)
	s.Require().ErrorIs(err, port.ErrScriptNotFound)
}

func (s *RepositoryTestSuite) TestScriptRepository_UpdateScript_Success() {
	script := makeSampleScript()
	err := s.repos.SaveScript(s.ctx, script)
	s.Require().NoError(err)

	err = script.Replace(
		"Test script updated",
		[]bots.Node{
			bots.MustNewNode(
				bots.MustNewState(1),
				"Node 1 updated",
				[]bots.Edge{
					bots.NewEdge(
						bots.MustNewExactMatchPredicate("Далее"),
						bots.MustNewState(3),
						bots.SaveOp{},
					),
					bots.NewEdge(
						bots.AlwaysTruePredicate{},
						bots.MustNewState(3),
						bots.NoOp{},
					),
				},
				[]bots.Message{
					bots.MustNewMessage("Message 1 for Node 1"),
					bots.MustNewMessage("Message 3 for Node 1"),
				},
				[]bots.Option{
					bots.MustNewOption("Option 1 for Node 1"),
					bots.MustNewOption("Option 3 for Node 1"),
				},
			),
			bots.MustNewNode(
				bots.MustNewState(3),
				"Node 3",
				[]bots.Edge{
					bots.NewEdge(
						bots.MustNewExactMatchPredicate("Назад"),
						bots.MustNewState(1),
						bots.NoOp{},
					),
				},
				[]bots.Message{
					bots.MustNewMessage("Message 1 for Node 3"),
				},
				[]bots.Option{
					bots.MustNewOption("Option 1 for Node 3"),
				},
			),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.MustNewState(1)),
			bots.MustNewEntry("start2", bots.MustNewState(1)),
		},
	)
	s.Require().NoError(err)

	err = s.repos.UpdateScript(s.ctx, script)
	s.Require().NoError(err)

	want := script
	got, err := s.repos.Script(s.ctx, want.ID())
	s.Require().NoError(err)

	// Assert script
	s.Require().NotNil(got)
	s.Require().Equal(want.ID(), got.ID())
	s.Require().Equal(want.OwnerID(), got.OwnerID())
	s.Require().Equal(want.Desc(), got.Desc())
	s.Require().WithinDuration(want.CreatedAt(), got.CreatedAt(), time.Second)
	s.Require().WithinDuration(want.CreatedAt(), got.CreatedAt(), time.Second)
	s.Require().Nil(script.DeletedAt())

	// Assert entries
	wantEntries := want.Entries()
	sort.Slice(wantEntries, func(i, j int) bool {
		return wantEntries[i].Key().String() < wantEntries[j].Key().String()
	})
	gotEntries := got.Entries()
	sort.Slice(gotEntries, func(i, j int) bool {
		return gotEntries[i].Key().String() < gotEntries[j].Key().String()
	})
	s.Require().Equal(wantEntries, gotEntries)

	// Assert nodes
	wantNodes := want.Nodes()
	sort.Slice(wantNodes, func(i, j int) bool {
		return wantNodes[i].State().Int() < wantNodes[j].State().Int()
	})
	gotNodes := got.Nodes()
	sort.Slice(gotNodes, func(i, j int) bool {
		return gotNodes[i].State().Int() < gotNodes[j].State().Int()
	})
	s.Require().Equal(wantNodes, gotNodes)
}

func makeSampleScript() *bots.Script {
	return bots.MustNewScript(
		actorUserID,
		"Test script",
		[]bots.Node{
			bots.MustNewNode(
				bots.MustNewState(1),
				"Node 1",
				[]bots.Edge{
					bots.NewEdge(
						bots.MustNewRegexMatchPredicate("^Далее$"),
						bots.MustNewState(2),
						bots.SaveOp{},
					),
					bots.NewEdge(
						bots.AlwaysTruePredicate{},
						bots.MustNewState(2),
						bots.AppendOp{},
					),
				},
				[]bots.Message{
					bots.MustNewMessage("Message 1 for Node 1"),
					bots.MustNewMessage("Message 2 for Node 1"),
				},
				[]bots.Option{
					bots.MustNewOption("Option 1 for Node 1"),
					bots.MustNewOption("Option 2 for Node 1"),
				},
			),
			bots.MustNewNode(
				bots.MustNewState(2),
				"Node 2",
				[]bots.Edge{
					bots.NewEdge(
						bots.MustNewExactMatchPredicate("Назад"),
						bots.MustNewState(1),
						bots.NoOp{},
					),
				},
				[]bots.Message{
					bots.MustNewMessage("Message 1 for Node 2"),
				},
				[]bots.Option{
					bots.MustNewOption("Option 1 for Node 2"),
				},
			),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.MustNewState(1)),
		},
	)
}
