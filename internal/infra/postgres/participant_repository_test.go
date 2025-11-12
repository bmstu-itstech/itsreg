package postgres_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/internal/infra/postgres"
	"github.com/bmstu-itstech/itsreg-bots/pkg/tests"
)

const (
	testBotID       = "test"
	testEntryKey    = "start"
	testEntryKeyAlt = "start2"
	testStartState  = 1
)

// Как сделать миграции над setupRepository?
func setupRepositoryWithParticipantFixtures() (*postgres.Repository, func()) {
	db := tests.ConnectPostgresDB()

	db.MustExecContext(
		context.Background(),
		`INSERT INTO bots (id, token, author) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
		testBotID, "token", 1,
	)

	db.MustExecContext(
		context.Background(),
		`INSERT INTO nodes (bot_id, state, title) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
		testBotID, 1, "Test",
	)

	db.MustExecContext(
		context.Background(),
		`INSERT INTO entries (bot_id, key, start) VALUES ($1, $2, $4), ($1, $3, $5) ON CONFLICT DO NOTHING`,
		testBotID, testEntryKey, testEntryKeyAlt, testStartState, testStartState,
	)

	return postgres.NewRepository(db, slog.Default()), func() {
		_ = db.Close()
	}
}

func TestPostgresParticipantRepository_CreateNew(t *testing.T) {
	r, closeFn := setupRepositoryWithParticipantFixtures()
	t.Cleanup(closeFn)

	ctx := context.Background()
	id := bots.NewParticipantID(bots.UserID(gofakeit.Int64()), testBotID)

	executed := false
	err := r.UpdateOrCreateParticipant(ctx, id, func(_ context.Context, _ *bots.Participant) error {
		executed = true
		return nil
	})
	require.NoError(t, err)
	require.True(t, executed)
}

func TestPostgresParticipantRepository_UpdateExisting(t *testing.T) {
	r, closeFn := setupRepositoryWithParticipantFixtures()
	t.Cleanup(closeFn)

	ctx := context.Background()
	id := bots.NewParticipantID(bots.UserID(gofakeit.Int64()), testBotID)
	entry := bots.MustNewEntry(testEntryKey, bots.MustNewState(testStartState))

	msg := bots.MustNewMessage("hello")
	err := r.UpdateOrCreateParticipant(ctx, id, func(_ context.Context, prt *bots.Participant) error {
		cthr, err := prt.StartThread(entry)
		cthr.SaveAnswer(msg)
		return err
	})
	require.NoError(t, err)

	err = r.UpdateOrCreateParticipant(ctx, id, func(_ context.Context, prt *bots.Participant) error {
		cthr := prt.ActiveThread()
		require.NotNil(t, cthr)
		recv, ok := cthr.Answers()[bots.MustNewState(testStartState)]
		require.True(t, ok)
		require.Equal(t, msg, recv)
		return nil
	})
	require.NoError(t, err)
}

func TestPostgresParticipantRepository_CreateMultiplyParticipants(t *testing.T) {
	r, closeFn := setupRepositoryWithParticipantFixtures()
	t.Cleanup(closeFn)

	ctx := context.Background()
	id1 := bots.NewParticipantID(bots.UserID(gofakeit.Int64()), testBotID)
	id2 := bots.NewParticipantID(bots.UserID(gofakeit.Int64()), testBotID)

	err := r.UpdateOrCreateParticipant(ctx, id1, func(_ context.Context, _ *bots.Participant) error {
		return nil
	})
	require.NoError(t, err)

	err = r.UpdateOrCreateParticipant(ctx, id2, func(_ context.Context, _ *bots.Participant) error {
		return nil
	})
	require.NoError(t, err)
}

func TestPostgresParticipantRepository_CreateNewThread(t *testing.T) {
	r, closeFn := setupRepositoryWithParticipantFixtures()
	t.Cleanup(closeFn)

	ctx := context.Background()
	id := bots.NewParticipantID(bots.UserID(gofakeit.Int64()), testBotID)

	entry1 := bots.MustNewEntry(testEntryKey, bots.MustNewState(testStartState))
	entry2 := bots.MustNewEntry(testEntryKeyAlt, bots.MustNewState(testStartState))

	var thrID1, thrID2 bots.ThreadID
	err := r.UpdateOrCreateParticipant(ctx, id, func(_ context.Context, prt *bots.Participant) error {
		_, err := prt.StartThread(entry1)
		require.NoError(t, err)
		require.NotNil(t, prt.ActiveThread())
		thrID1 = prt.ActiveThread().ID()
		return err
	})
	require.NoError(t, err)

	err = r.UpdateOrCreateParticipant(ctx, id, func(_ context.Context, prt *bots.Participant) error {
		require.Equal(t, thrID1, prt.ActiveThread().ID())
		_, err = prt.StartThread(entry2)
		require.NotNil(t, prt.ActiveThread())
		thrID2 = prt.ActiveThread().ID()
		return err
	})
	require.NoError(t, err)

	err = r.UpdateOrCreateParticipant(ctx, id, func(_ context.Context, prt *bots.Participant) error {
		require.NotNil(t, prt.ActiveThread())
		require.Equal(t, thrID2, prt.ActiveThread().ID())
		return err
	})
	require.NoError(t, err)
}
