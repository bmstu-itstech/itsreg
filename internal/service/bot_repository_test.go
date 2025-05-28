package service_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/internal/service"
)

func setupMockBotRepository() *service.MockBotRepository {
	return service.NewMockBotRepository()
}

func TestMockBotRepository_CreateNew(t *testing.T) {
	r := setupMockBotRepository()
	testBotRepositoryCreateNew(t, r, r)
}

func TestMockBotRepository_ErrorIfNotExists(t *testing.T) {
	r := setupMockBotRepository()
	testBotRepositoryErrorIfNotExists(t, r)
}

func TestMockBotRepository_AddNode(t *testing.T) {
	r := setupMockBotRepository()
	testBotRepositoryAddNode(t, r, r)
}

func TestMockBotRepository_RemoveNode(t *testing.T) {
	r := setupMockBotRepository()
	testBotRepositoryRemoveNode(t, r, r)
}

func TestMockBotRepository_AddEntry(t *testing.T) {
	r := setupMockBotRepository()
	testBotRepositoryAddEntry(t, r, r)
}

func TestMockBotRepository_RemoveEntry(t *testing.T) {
	r := setupMockBotRepository()
	testBotRepositoryRemoveEntry(t, r, r)
}

func TestMockBotRepository_AddEdge(t *testing.T) {
	r := setupMockBotRepository()
	testBotRepositoryAddEdge(t, r, r)
}

func TestMockBotRepository_RemoveEdge(t *testing.T) {
	r := setupMockBotRepository()
	testBotRepositoryRemoveEdge(t, r, r)
}

func TestMockBotRepository_AddMessage(t *testing.T) {
	r := setupMockBotRepository()
	testBotRepositoryAddMessage(t, r, r)
}

func TestMockBotRepository_RemoveMessage(t *testing.T) {
	r := setupMockBotRepository()
	testBotRepositoryRemoveMessage(t, r, r)
}

func TestMockBotRepository_UpdateBotData(t *testing.T) {
	r := setupMockBotRepository()
	testBotRepositoryUpdateBotData(t, r, r)
}

func TestMockBotRepository_UpdateNodeAndEntry(t *testing.T) {
	r := setupMockBotRepository()
	testBotRepositoryUpdateNodeAndEntry(t, r, r)
}

func testBotRepositoryCreateNew(t *testing.T, m bots.BotManager, p bots.BotProvider) {
	ctx := context.Background()

	id := bots.BotId(gofakeit.AppName())
	bot := bots.MustNewBot(id, "token", bots.UserId(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.State(1), "Greeting", nil, []bots.BotMessage{
				bots.MustNewBotMessage("Hello, world!", nil),
			}),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.State(1)),
		},
	))

	err := m.Upsert(ctx, bot)
	require.NoError(t, err)

	recv, err := p.Bot(ctx, id)
	require.NoError(t, err)
	require.Equal(t, bot, recv)
}

func testBotRepositoryAddNode(t *testing.T, m bots.BotManager, p bots.BotProvider) {
	ctx := context.Background()

	id := bots.BotId(gofakeit.AppName())
	bot := bots.MustNewBot(id, "token", bots.UserId(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.State(1), "Greeting", nil, []bots.BotMessage{
				bots.MustNewBotMessage("Hello, world!", nil),
			}),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.State(1)),
		},
	))

	err := m.Upsert(ctx, bot)
	require.NoError(t, err)

	updatedBot := bots.MustNewBot(id, "token", bots.UserId(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(
				bots.State(1),
				"Greeting",
				[]bots.Edge{
					bots.NewEdge(bots.AlwaysTruePredicate{}, bots.State(2), bots.NoOp{}),
				},
				[]bots.BotMessage{
					bots.MustNewBotMessage("Hello, world!", nil),
				},
			),
			bots.MustNewNode(bots.State(2), "Goodbye", nil, []bots.BotMessage{
				bots.MustNewBotMessage("Goodbye!", nil),
			}),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.State(1)),
		},
	))

	err = m.Upsert(ctx, updatedBot)
	require.NoError(t, err)
	recv, err := p.Bot(ctx, id)
	require.NoError(t, err)
	require.Equal(t, updatedBot, recv)
}

func testBotRepositoryRemoveNode(t *testing.T, m bots.BotManager, p bots.BotProvider) {
	ctx := context.Background()

	id := bots.BotId(gofakeit.AppName())
	bot := bots.MustNewBot(id, "token", bots.UserId(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(
				bots.State(1),
				"Greeting",
				[]bots.Edge{
					bots.NewEdge(bots.AlwaysTruePredicate{}, bots.State(2), bots.NoOp{}),
				},
				[]bots.BotMessage{
					bots.MustNewBotMessage("Hello, world!", nil),
				},
			),
			bots.MustNewNode(bots.State(2), "Goodbye", nil, []bots.BotMessage{
				bots.MustNewBotMessage("Goodbye!", nil),
			}),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.State(1)),
		},
	))

	err := m.Upsert(ctx, bot)
	require.NoError(t, err)

	updatedBot := bots.MustNewBot(id, "token", bots.UserId(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.State(1), "Greeting", nil, []bots.BotMessage{
				bots.MustNewBotMessage("Hello, world!", nil),
			}),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.State(1)),
		},
	))

	err = m.Upsert(ctx, updatedBot)
	require.NoError(t, err)
	recv, err := p.Bot(ctx, id)
	require.NoError(t, err)
	require.Equal(t, updatedBot, recv)
}

func testBotRepositoryAddEntry(t *testing.T, m bots.BotManager, p bots.BotProvider) {
	ctx := context.Background()

	id := bots.BotId(gofakeit.AppName())
	bot := bots.MustNewBot(id, "token", bots.UserId(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.State(1), "Greeting", nil, []bots.BotMessage{
				bots.MustNewBotMessage("Hello, world!", nil),
			}),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.State(1)),
		},
	))

	err := m.Upsert(ctx, bot)
	require.NoError(t, err)

	updatedBot := bots.MustNewBot(id, "token", bots.UserId(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.State(1), "Greeting", nil, []bots.BotMessage{
				bots.MustNewBotMessage("Hello, world!", nil),
			}),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.State(1)),
			bots.MustNewEntry("mailing_1", bots.State(1)),
		},
	))

	err = m.Upsert(ctx, updatedBot)
	require.NoError(t, err)
	recv, err := p.Bot(ctx, id)
	require.NoError(t, err)
	require.Equal(t, updatedBot, recv)
}

func testBotRepositoryRemoveEntry(t *testing.T, m bots.BotManager, p bots.BotProvider) {
	ctx := context.Background()

	id := bots.BotId(gofakeit.AppName())
	bot := bots.MustNewBot(id, "token", bots.UserId(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.State(1), "Greeting", nil, []bots.BotMessage{
				bots.MustNewBotMessage("Hello, world!", nil),
			}),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.State(1)),
			bots.MustNewEntry("mailing_1", bots.State(1)),
		},
	))

	err := m.Upsert(ctx, bot)
	require.NoError(t, err)

	updatedBot := bots.MustNewBot(id, "token", bots.UserId(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.State(1), "Greeting", nil, []bots.BotMessage{
				bots.MustNewBotMessage("Hello, world!", nil),
			}),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.State(1)),
		},
	))

	err = m.Upsert(ctx, updatedBot)
	require.NoError(t, err)
	recv, err := p.Bot(ctx, id)
	require.NoError(t, err)
	require.Equal(t, updatedBot, recv)
}

func testBotRepositoryAddEdge(t *testing.T, m bots.BotManager, p bots.BotProvider) {
	ctx := context.Background()

	id := bots.BotId(gofakeit.AppName())
	bot := bots.MustNewBot(id, "token", bots.UserId(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.State(1), "Greeting", nil, []bots.BotMessage{
				bots.MustNewBotMessage("Hello, world!", nil),
			}),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.State(1)),
		},
	))

	err := m.Upsert(ctx, bot)
	require.NoError(t, err)

	updatedBot := bots.MustNewBot(id, "token", bots.UserId(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.State(1), "Greeting",
				[]bots.Edge{
					bots.NewEdge(bots.MustNewRegexMatchPredicate("repeat"), bots.State(1), bots.NoOp{}),
				},
				[]bots.BotMessage{
					bots.MustNewBotMessage("Hello, world!", nil),
				},
			),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.State(1)),
		},
	))

	err = m.Upsert(ctx, updatedBot)
	require.NoError(t, err)
	recv, err := p.Bot(ctx, id)
	require.NoError(t, err)
	require.Equal(t, updatedBot, recv)
}

func testBotRepositoryRemoveEdge(t *testing.T, m bots.BotManager, p bots.BotProvider) {
	ctx := context.Background()

	id := bots.BotId(gofakeit.AppName())
	bot := bots.MustNewBot(id, "token", bots.UserId(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.State(1), "Greeting",
				[]bots.Edge{
					bots.NewEdge(bots.MustNewRegexMatchPredicate("repeat"), bots.State(1), bots.NoOp{}),
				},
				[]bots.BotMessage{
					bots.MustNewBotMessage("Hello, world!", nil),
				},
			),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.State(1)),
		},
	))

	updatedBot := bots.MustNewBot(id, "token", bots.UserId(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.State(1), "Greeting", nil, []bots.BotMessage{
				bots.MustNewBotMessage("Hello, world!", nil),
			}),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.State(1)),
		},
	))

	err := m.Upsert(ctx, bot)
	require.NoError(t, err)

	err = m.Upsert(ctx, updatedBot)
	require.NoError(t, err)
	recv, err := p.Bot(ctx, id)
	require.NoError(t, err)
	require.Equal(t, updatedBot, recv)
}

func testBotRepositoryAddMessage(t *testing.T, m bots.BotManager, p bots.BotProvider) {
	ctx := context.Background()

	id := bots.BotId(gofakeit.AppName())
	bot := bots.MustNewBot(id, "token", bots.UserId(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.State(1), "Greeting", nil, []bots.BotMessage{
				bots.MustNewBotMessage("Hello, world!", nil),
			}),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.State(1)),
		},
	))

	err := m.Upsert(ctx, bot)
	require.NoError(t, err)

	updatedBot := bots.MustNewBot(id, "token", bots.UserId(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.State(1), "Greeting", nil, []bots.BotMessage{
				bots.MustNewBotMessage("Hello, world!", nil),
				bots.MustNewBotMessage("Hello, world x2!", nil),
			}),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.State(1)),
		},
	))

	err = m.Upsert(ctx, updatedBot)
	require.NoError(t, err)
	recv, err := p.Bot(ctx, id)
	require.NoError(t, err)
	require.Equal(t, updatedBot, recv)
}

func testBotRepositoryRemoveMessage(t *testing.T, m bots.BotManager, p bots.BotProvider) {
	ctx := context.Background()

	id := bots.BotId(gofakeit.AppName())
	bot := bots.MustNewBot(id, "token", bots.UserId(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.State(1), "Greeting", nil, []bots.BotMessage{
				bots.MustNewBotMessage("Hello, world!", nil),
				bots.MustNewBotMessage("Hello, world x2!", nil),
			}),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.State(1)),
		},
	))

	err := m.Upsert(ctx, bot)
	require.NoError(t, err)

	updatedBot := bots.MustNewBot(id, "token", bots.UserId(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.State(1), "Greeting", nil, []bots.BotMessage{
				bots.MustNewBotMessage("Hello, world!", nil),
			}),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.State(1)),
		},
	))

	err = m.Upsert(ctx, updatedBot)
	require.NoError(t, err)
	recv, err := p.Bot(ctx, id)
	require.NoError(t, err)
	require.Equal(t, updatedBot, recv)
}

func testBotRepositoryUpdateBotData(t *testing.T, m bots.BotManager, p bots.BotProvider) {
	ctx := context.Background()

	id := bots.BotId(gofakeit.AppName())
	bot := bots.MustNewBot(id, "token", bots.UserId(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.State(1), "Greeting", nil, []bots.BotMessage{
				bots.MustNewBotMessage("Hello, world!", nil),
			}),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.State(1)),
		},
	))

	err := m.Upsert(ctx, bot)
	require.NoError(t, err)

	updatedBot := bots.MustNewBot(id, "token2", bots.UserId(2), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.State(1), "Greeting", nil, []bots.BotMessage{
				bots.MustNewBotMessage("Hello, world!", nil),
			}),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.State(1)),
		},
	))

	err = m.Upsert(ctx, updatedBot)
	require.NoError(t, err)
	recv, err := p.Bot(ctx, id)
	require.NoError(t, err)
	require.Equal(t, updatedBot, recv)
}

func testBotRepositoryUpdateNodeAndEntry(t *testing.T, m bots.BotManager, p bots.BotProvider) {
	ctx := context.Background()

	id := bots.BotId(gofakeit.AppName())
	bot := bots.MustNewBot(id, "token", bots.UserId(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.State(1), "Greeting", nil, []bots.BotMessage{
				bots.MustNewBotMessage("Hello, world!", nil),
			}),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.State(1)),
		},
	))

	err := m.Upsert(ctx, bot)
	require.NoError(t, err)

	updatedBot := bots.MustNewBot(id, "token", bots.UserId(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.State(2), "Greeting 2", nil, []bots.BotMessage{
				bots.MustNewBotMessage("Hello, world!", nil),
			}),
		},
		[]bots.Entry{
			bots.MustNewEntry("start2", bots.State(2)),
		},
	))

	err = m.Upsert(ctx, updatedBot)
	require.NoError(t, err)
	recv, err := p.Bot(ctx, id)
	require.NoError(t, err)
	require.Equal(t, updatedBot, recv)
}

func testBotRepositoryErrorIfNotExists(t *testing.T, p bots.BotProvider) {
	ctx := context.Background()
	id := bots.BotId(gofakeit.AppName())
	_, err := p.Bot(ctx, id)
	require.ErrorIs(t, err, bots.ErrBotNotFound)
}
