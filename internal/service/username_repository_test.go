package service_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/internal/service"
)

func setupMockUsernameRepository() *service.MockUsernameRepository {
	return service.NewMockUsernameRepository()
}

func TestMockUsernameRepository_CreateNew(t *testing.T) {
	r := setupMockUsernameRepository()
	testUsernameRepositoryCreateNew(t, r, r)
}

func TestMockUsernameRepository_UpdateExisting(t *testing.T) {
	r := setupMockUsernameRepository()
	testUsernameRepositoryUpdateExisting(t, r, r)
}

func TestMockUsernameRepository_ErrorIfNotExists(t *testing.T) {
	r := setupMockUsernameRepository()
	testUsernameRepositoryErrorIfNotExists(t, r)
}

func testUsernameRepositoryCreateNew(t *testing.T, m bots.UsernameManager, p bots.UsernameProvider) {
	ctx := context.Background()
	uid := bots.UserId(gofakeit.Int64())

	username := bots.Username(gofakeit.Username())
	err := m.Upsert(ctx, uid, username)
	require.NoError(t, err)

	recv, err := p.Username(ctx, uid)
	require.NoError(t, err)
	require.Equal(t, username, recv)
}

func testUsernameRepositoryUpdateExisting(t *testing.T, m bots.UsernameManager, p bots.UsernameProvider) {
	ctx := context.Background()
	uid := bots.UserId(gofakeit.Int64())

	username1 := bots.Username(gofakeit.Username())
	err := m.Upsert(ctx, uid, username1)
	require.NoError(t, err)

	username2 := bots.Username(gofakeit.Username())
	err = m.Upsert(ctx, uid, username2)
	require.NoError(t, err)

	recv, err := p.Username(ctx, uid)
	require.NoError(t, err)
	require.Equal(t, username2, recv)
}

func testUsernameRepositoryErrorIfNotExists(t *testing.T, p bots.UsernameProvider) {
	ctx := context.Background()
	uid := bots.UserId(gofakeit.Int64())
	_, err := p.Username(ctx, uid)
	require.ErrorIs(t, err, bots.ErrUsernameNotFound)
}
