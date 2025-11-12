package postgres_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg-bots/internal/app/port"
	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
)

func TestPostgresBotRepository_CreateNew(t *testing.T) {
	r, closeFn := setupRepository()
	t.Cleanup(closeFn)

	ctx := context.Background()

	id := bots.BotID(gofakeit.AppName())
	bot := bots.MustNewBot(id, "token", bots.UserID(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.MustNewState(1), "Greeting", nil, []bots.Message{
				bots.MustNewMessage("Hello, world!"),
			}, nil),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.MustNewState(1)),
		},
	))

	err := r.UpsertBot(ctx, bot)
	require.NoError(t, err)

	recv, err := r.Bot(ctx, id)
	require.NoError(t, err)
	require.Equal(t, bot, recv)
}

func TestPostgresBotRepository_ErrorIfNotExists(t *testing.T) {
	r, closeFn := setupRepository()
	t.Cleanup(closeFn)

	ctx := context.Background()
	id := bots.BotID(gofakeit.AppName())
	_, err := r.Bot(ctx, id)
	require.ErrorIs(t, err, port.ErrBotNotFound)
}

func TestPostgresBotRepository_AddNode(t *testing.T) {
	r, closeFn := setupRepository()
	t.Cleanup(closeFn)

	ctx := context.Background()

	id := bots.BotID(gofakeit.AppName())
	bot := bots.MustNewBot(id, "token", bots.UserID(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.MustNewState(1), "Greeting", nil, []bots.Message{
				bots.MustNewMessage("Hello, world!"),
			}, nil),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.MustNewState(1)),
		},
	))

	err := r.UpsertBot(ctx, bot)
	require.NoError(t, err)

	updatedBot := bots.MustNewBot(id, "token", bots.UserID(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(
				bots.MustNewState(1),
				"Greeting",
				[]bots.Edge{
					bots.NewEdge(bots.AlwaysTruePredicate{}, bots.MustNewState(2), bots.NoOp{}),
				},
				[]bots.Message{
					bots.MustNewMessage("Hello, world!"),
				},
				nil,
			),
			bots.MustNewNode(bots.MustNewState(2), "Goodbye", nil, []bots.Message{
				bots.MustNewMessage("Goodbye!"),
			}, nil),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.MustNewState(1)),
		},
	))

	err = r.UpsertBot(ctx, updatedBot)
	require.NoError(t, err)
	recv, err := r.Bot(ctx, id)
	require.NoError(t, err)
	require.Equal(t, updatedBot, recv)
}

func TestPostgresBotRepository_RemoveNode(t *testing.T) {
	r, closeFn := setupRepository()
	t.Cleanup(closeFn)

	ctx := context.Background()

	id := bots.BotID(gofakeit.AppName())
	bot := bots.MustNewBot(id, "token", bots.UserID(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(
				bots.MustNewState(1),
				"Greeting",
				[]bots.Edge{
					bots.NewEdge(bots.AlwaysTruePredicate{}, bots.MustNewState(2), bots.NoOp{}),
				},
				[]bots.Message{
					bots.MustNewMessage("Hello, world!"),
				},
				nil,
			),
			bots.MustNewNode(bots.MustNewState(2), "Goodbye", nil, []bots.Message{
				bots.MustNewMessage("Goodbye!"),
			}, nil),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.MustNewState(1)),
		},
	))

	err := r.UpsertBot(ctx, bot)
	require.NoError(t, err)

	updatedBot := bots.MustNewBot(id, "token", bots.UserID(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.MustNewState(1), "Greeting", nil, []bots.Message{
				bots.MustNewMessage("Hello, world!"),
			}, nil),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.MustNewState(1)),
		},
	))

	err = r.UpsertBot(ctx, updatedBot)
	require.NoError(t, err)
	recv, err := r.Bot(ctx, id)
	require.NoError(t, err)
	require.Equal(t, updatedBot, recv)
}

func TestPostgresBotRepository_AddEntry(t *testing.T) {
	r, closeFn := setupRepository()
	t.Cleanup(closeFn)

	ctx := context.Background()

	id := bots.BotID(gofakeit.AppName())
	bot := bots.MustNewBot(id, "token", bots.UserID(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.MustNewState(1), "Greeting", nil, []bots.Message{
				bots.MustNewMessage("Hello, world!"),
			}, nil),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.MustNewState(1)),
		},
	))

	err := r.UpsertBot(ctx, bot)
	require.NoError(t, err)

	updatedBot := bots.MustNewBot(id, "token", bots.UserID(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.MustNewState(1), "Greeting", nil, []bots.Message{
				bots.MustNewMessage("Hello, world!"),
			}, nil),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.MustNewState(1)),
			bots.MustNewEntry("mailing_1", bots.MustNewState(1)),
		},
	))

	err = r.UpsertBot(ctx, updatedBot)
	require.NoError(t, err)
	recv, err := r.Bot(ctx, id)
	require.NoError(t, err)
	require.Equal(t, updatedBot, recv)
}

func TestPostgresBotRepository_RemoveEntry(t *testing.T) {
	r, closeFn := setupRepository()
	t.Cleanup(closeFn)

	ctx := context.Background()

	id := bots.BotID(gofakeit.AppName())
	bot := bots.MustNewBot(id, "token", bots.UserID(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.MustNewState(1), "Greeting", nil, []bots.Message{
				bots.MustNewMessage("Hello, world!"),
			}, nil),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.MustNewState(1)),
			bots.MustNewEntry("mailing_1", bots.MustNewState(1)),
		},
	))

	err := r.UpsertBot(ctx, bot)
	require.NoError(t, err)

	updatedBot := bots.MustNewBot(id, "token", bots.UserID(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.MustNewState(1), "Greeting", nil, []bots.Message{
				bots.MustNewMessage("Hello, world!"),
			}, nil),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.MustNewState(1)),
		},
	))

	err = r.UpsertBot(ctx, updatedBot)
	require.NoError(t, err)
	recv, err := r.Bot(ctx, id)
	require.NoError(t, err)
	require.Equal(t, updatedBot, recv)
}

func TestPostgresBotRepository_AddEdge(t *testing.T) {
	r, closeFn := setupRepository()
	t.Cleanup(closeFn)

	ctx := context.Background()

	id := bots.BotID(gofakeit.AppName())
	bot := bots.MustNewBot(id, "token", bots.UserID(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.MustNewState(1), "Greeting", nil, []bots.Message{
				bots.MustNewMessage("Hello, world!"),
			}, nil),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.MustNewState(1)),
		},
	))

	err := r.UpsertBot(ctx, bot)
	require.NoError(t, err)

	updatedBot := bots.MustNewBot(id, "token", bots.UserID(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.MustNewState(1), "Greeting",
				[]bots.Edge{
					bots.NewEdge(bots.MustNewRegexMatchPredicate("repeat"), bots.MustNewState(1), bots.NoOp{}),
				},
				[]bots.Message{
					bots.MustNewMessage("Hello, world!"),
				},
				nil,
			),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.MustNewState(1)),
		},
	))

	err = r.UpsertBot(ctx, updatedBot)
	require.NoError(t, err)
	recv, err := r.Bot(ctx, id)
	require.NoError(t, err)
	require.Equal(t, updatedBot, recv)
}

func TestPostgresBotRepository_RemoveEdge(t *testing.T) {
	r, closeFn := setupRepository()
	t.Cleanup(closeFn)

	ctx := context.Background()

	id := bots.BotID(gofakeit.AppName())
	bot := bots.MustNewBot(id, "token", bots.UserID(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.MustNewState(1), "Greeting",
				[]bots.Edge{
					bots.NewEdge(bots.MustNewRegexMatchPredicate("repeat"), bots.MustNewState(1), bots.NoOp{}),
				},
				[]bots.Message{
					bots.MustNewMessage("Hello, world!"),
				},
				nil,
			),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.MustNewState(1)),
		},
	))

	updatedBot := bots.MustNewBot(id, "token", bots.UserID(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.MustNewState(1), "Greeting", nil, []bots.Message{
				bots.MustNewMessage("Hello, world!"),
			}, nil),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.MustNewState(1)),
		},
	))

	err := r.UpsertBot(ctx, bot)
	require.NoError(t, err)

	err = r.UpsertBot(ctx, updatedBot)
	require.NoError(t, err)
	recv, err := r.Bot(ctx, id)
	require.NoError(t, err)
	require.Equal(t, updatedBot, recv)
}

func TestPostgresBotRepository_AddMessage(t *testing.T) {
	r, closeFn := setupRepository()
	t.Cleanup(closeFn)

	ctx := context.Background()

	id := bots.BotID(gofakeit.AppName())
	bot := bots.MustNewBot(id, "token", bots.UserID(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.MustNewState(1), "Greeting", nil, []bots.Message{
				bots.MustNewMessage("Hello, world!"),
			}, nil),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.MustNewState(1)),
		},
	))

	err := r.UpsertBot(ctx, bot)
	require.NoError(t, err)

	updatedBot := bots.MustNewBot(id, "token", bots.UserID(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.MustNewState(1), "Greeting", nil, []bots.Message{
				bots.MustNewMessage("Hello, world!"),
				bots.MustNewMessage("Hello, world x2!"),
			}, nil),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.MustNewState(1)),
		},
	))

	err = r.UpsertBot(ctx, updatedBot)
	require.NoError(t, err)
	recv, err := r.Bot(ctx, id)
	require.NoError(t, err)
	require.Equal(t, updatedBot, recv)
}

func TestPostgresBotRepository_RemoveMessage(t *testing.T) {
	r, closeFn := setupRepository()
	t.Cleanup(closeFn)

	ctx := context.Background()

	id := bots.BotID(gofakeit.AppName())
	bot := bots.MustNewBot(id, "token", bots.UserID(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.MustNewState(1), "Greeting", nil, []bots.Message{
				bots.MustNewMessage("Hello, world!"),
				bots.MustNewMessage("Hello, world x2!"),
			}, nil),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.MustNewState(1)),
		},
	))

	err := r.UpsertBot(ctx, bot)
	require.NoError(t, err)

	updatedBot := bots.MustNewBot(id, "token", bots.UserID(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.MustNewState(1), "Greeting", nil, []bots.Message{
				bots.MustNewMessage("Hello, world!"),
			}, nil),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.MustNewState(1)),
		},
	))

	err = r.UpsertBot(ctx, updatedBot)
	require.NoError(t, err)
	recv, err := r.Bot(ctx, id)
	require.NoError(t, err)
	require.Equal(t, updatedBot, recv)
}

func TestPostgresBotRepository_UpdateBotData(t *testing.T) {
	r, closeFn := setupRepository()
	t.Cleanup(closeFn)

	ctx := context.Background()

	id := bots.BotID(gofakeit.AppName())
	bot := bots.MustNewBot(id, "token", bots.UserID(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.MustNewState(1), "Greeting", nil, []bots.Message{
				bots.MustNewMessage("Hello, world!"),
			}, nil),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.MustNewState(1)),
		},
	))

	err := r.UpsertBot(ctx, bot)
	require.NoError(t, err)

	updatedBot := bots.MustNewBot(id, "token2", bots.UserID(2), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.MustNewState(1), "Greeting", nil, []bots.Message{
				bots.MustNewMessage("Hello, world!"),
			}, nil),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.MustNewState(1)),
		},
	))

	err = r.UpsertBot(ctx, updatedBot)
	require.NoError(t, err)
	recv, err := r.Bot(ctx, id)
	require.NoError(t, err)
	require.Equal(t, updatedBot, recv)
}

func TestPostgresBotRepository_UpdateNodeAndEntry(t *testing.T) {
	r, closeFn := setupRepository()
	t.Cleanup(closeFn)

	ctx := context.Background()

	id := bots.BotID(gofakeit.AppName())
	bot := bots.MustNewBot(id, "token", bots.UserID(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.MustNewState(1), "Greeting", nil, []bots.Message{
				bots.MustNewMessage("Hello, world!"),
			}, nil),
		},
		[]bots.Entry{
			bots.MustNewEntry("start", bots.MustNewState(1)),
		},
	))

	err := r.UpsertBot(ctx, bot)
	require.NoError(t, err)

	updatedBot := bots.MustNewBot(id, "token", bots.UserID(1), bots.MustNewScript(
		[]bots.Node{
			bots.MustNewNode(bots.MustNewState(2), "Greeting 2", nil, []bots.Message{
				bots.MustNewMessage("Hello, world!"),
			}, nil),
		},
		[]bots.Entry{
			bots.MustNewEntry("start2", bots.MustNewState(2)),
		},
	))

	err = r.UpsertBot(ctx, updatedBot)
	require.NoError(t, err)
	recv, err := r.Bot(ctx, id)
	require.NoError(t, err)
	require.Equal(t, updatedBot, recv)
}
