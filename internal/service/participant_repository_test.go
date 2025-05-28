package service_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/internal/service"
)

func setupMockParticipantRepository() *service.MockParticipantRepository {
	return service.NewMockParticipantRepository()
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

func testParticipantRepositoryCreateNew(t *testing.T, repo bots.ParticipantRepository) {
	ctx := context.Background()
	id := bots.NewParticipantId(bots.UserId(gofakeit.Int64()), bots.BotId(gofakeit.AppName()))

	executed := false
	err := repo.UpdateOrCreate(ctx, id, func(_ context.Context, prt *bots.Participant) error {
		executed = true
		return nil
	})
	require.NoError(t, err)
	require.True(t, executed)
}

func testParticipantRepositoryUpdateExisting(t *testing.T, repo bots.ParticipantRepository) {
	ctx := context.Background()
	id := bots.NewParticipantId(bots.UserId(gofakeit.Int64()), bots.BotId(gofakeit.AppName()))
	entry := bots.MustNewEntry("start", 1)

	err := repo.UpdateOrCreate(ctx, id, func(_ context.Context, prt *bots.Participant) error {
		_, err := prt.StartThread(entry)
		return err
	})
	require.NoError(t, err)

	err = repo.UpdateOrCreate(ctx, id, func(_ context.Context, prt *bots.Participant) error {
		cthr, ok := prt.CurrentThread()
		require.True(t, ok)
		require.NotNil(t, cthr)
		return nil
	})
	require.NoError(t, err)
}

func testParticipantRepositoryCreateMultiplyParticipants(t *testing.T, repo bots.ParticipantRepository) {
	ctx := context.Background()
	id1 := bots.NewParticipantId(bots.UserId(gofakeit.Int64()), bots.BotId(gofakeit.AppName()))
	id2 := bots.NewParticipantId(bots.UserId(gofakeit.Int64()), bots.BotId(gofakeit.AppName()))

	err := repo.UpdateOrCreate(ctx, id1, func(ctx context.Context, _ *bots.Participant) error {
		return nil
	})
	require.NoError(t, err)

	err = repo.UpdateOrCreate(ctx, id2, func(ctx context.Context, _ *bots.Participant) error {
		return nil
	})
	require.NoError(t, err)
}
