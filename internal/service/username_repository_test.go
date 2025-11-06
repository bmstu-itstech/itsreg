package service_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/bmstu-itstech/itsreg-bots/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg-bots/internal/service"
)

func setupMockUsernameRepository() *service.MockUsernameRepository {
	return service.NewMockUsernameRepository()
}

func setupPostgresUsernameRepository() (*service.PostgresUsernameRepository, func()) {
	uri := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		"localhost",
		os.Getenv("POSTGRES_EXTERNAL_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)
	db := sqlx.MustConnect("postgres", uri)
	return service.NewPostgresUsernameRepository(db), func() {
		_ = db.Close()
	}
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

func TestPostgresUsernameRepository_CreateNew(t *testing.T) {
	r, closeFn := setupPostgresUsernameRepository()
	t.Cleanup(closeFn)
	testUsernameRepositoryCreateNew(t, r, r)
}

func TestPostgresUsernameRepository_UpdateExisting(t *testing.T) {
	r, closeFn := setupPostgresUsernameRepository()
	t.Cleanup(closeFn)
	testUsernameRepositoryUpdateExisting(t, r, r)
}

func TestPostgresUsernameRepository_ErrorIfNotExists(t *testing.T) {
	r, closeFn := setupPostgresUsernameRepository()
	t.Cleanup(closeFn)
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
