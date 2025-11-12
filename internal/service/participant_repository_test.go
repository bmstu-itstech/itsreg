package service_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/internal/service"
	"github.com/bmstu-itstech/itsreg-bots/pkg/tests"
)

const (
	testBotID      = "test"
	testEntryKey   = "start"
	testStartState = 1
)

func setupMockParticipantRepository() *service.MockParticipantRepository {
	return service.NewMockParticipantRepository()
}

func setupPostgresParticipantRepository() (*service.PostgresParticipantRepository, func()) {
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
		`INSERT INTO entries (bot_id, key, start) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
		testBotID, testEntryKey, testStartState,
	)

	return service.NewPostgresParticipantRepository(db, slog.Default()), func() {
		_ = db.Close()
	}
}

func TestMockParticipantRepository_CreateNew(t *testing.T) {
	r := setupMockParticipantRepository()
	testParticipantRepositoryCreateNew(t, r)
}

func TestMockParticipantRepository_UpdateExisting(t *testing.T) {
	r := setupMockParticipantRepository()
	testParticipantRepositoryUpdateExisting(t, r)
}

func TestMockParticipantRepository_CreateMultiplyParticipants(t *testing.T) {
	r := setupMockParticipantRepository()
	testParticipantRepositoryCreateMultiplyParticipants(t, r)
}

func TestPostgresParticipantRepository_CreateNew(t *testing.T) {
	r, closeFn := setupPostgresParticipantRepository()
	t.Cleanup(closeFn)
	testParticipantRepositoryCreateNew(t, r)
}

func TestPostgresParticipantRepository_UpdateExisting(t *testing.T) {
	r, closeFn := setupPostgresParticipantRepository()
	t.Cleanup(closeFn)
	testParticipantRepositoryUpdateExisting(t, r)
}

func TestPostgresParticipantRepository_CreateMultiplyParticipants(t *testing.T) {
	r, closeFn := setupPostgresParticipantRepository()
	t.Cleanup(closeFn)
	testParticipantRepositoryCreateMultiplyParticipants(t, r)
}

func testParticipantRepositoryCreateNew(t *testing.T, repo bots.ParticipantRepository) {
	ctx := context.Background()
	id := bots.NewParticipantID(bots.UserID(gofakeit.Int64()), testBotID)

	executed := false
	err := repo.UpdateOrCreate(ctx, id, func(_ context.Context, _ *bots.Participant) error {
		executed = true
		return nil
	})
	require.NoError(t, err)
	require.True(t, executed)
}

func testParticipantRepositoryUpdateExisting(t *testing.T, repo bots.ParticipantRepository) {
	ctx := context.Background()
	id := bots.NewParticipantID(bots.UserID(gofakeit.Int64()), testBotID)
	entry := bots.MustNewEntry(testEntryKey, bots.MustNewState(testStartState))

	msg := bots.MustNewMessage("hello")
	err := repo.UpdateOrCreate(ctx, id, func(_ context.Context, prt *bots.Participant) error {
		cthr, err := prt.StartThread(entry)
		cthr.SaveAnswer(msg)
		return err
	})
	require.NoError(t, err)

	err = repo.UpdateOrCreate(ctx, id, func(_ context.Context, prt *bots.Participant) error {
		cthr, ok := prt.CurrentThread()
		require.True(t, ok)
		require.NotNil(t, cthr)
		recv, ok := cthr.Answers()[bots.MustNewState(testStartState)]
		require.True(t, ok)
		require.Equal(t, msg, recv)
		return nil
	})
	require.NoError(t, err)
}

func testParticipantRepositoryCreateMultiplyParticipants(t *testing.T, repo bots.ParticipantRepository) {
	ctx := context.Background()
	id1 := bots.NewParticipantID(bots.UserID(gofakeit.Int64()), testBotID)
	id2 := bots.NewParticipantID(bots.UserID(gofakeit.Int64()), testBotID)

	err := repo.UpdateOrCreate(ctx, id1, func(_ context.Context, _ *bots.Participant) error {
		return nil
	})
	require.NoError(t, err)

	err = repo.UpdateOrCreate(ctx, id2, func(_ context.Context, _ *bots.Participant) error {
		return nil
	})
	require.NoError(t, err)
}
