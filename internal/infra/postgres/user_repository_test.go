package postgres_test

import (
	"time"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

type userDBRow struct {
	ID        int64     `db:"id"`
	Username  string    `db:"username"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (s *RepositoryTestSuite) TestUserRepository_UpsertUsername_InsertNewUser() {
	const userID = int64(7777)
	const username = bots.Username("new_user")

	err := s.repos.UpsertUsername(s.ctx, bots.UserID(userID), username)
	s.Require().NoError(err)

	var got userDBRow
	err = s.db.GetContext(s.ctx, &got, `
		SELECT id, username, updated_at
		FROM users
		WHERE id = $1
	`, userID)
	s.Require().NoError(err)
	s.Require().Equal(userID, got.ID)
	s.Require().Equal(username.String(), got.Username)
	s.Require().True(got.UpdatedAt.After(time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC)))
}

func (s *RepositoryTestSuite) TestUserRepository_UpsertUsername_UpdateExistingUser() {
	const userID = int64(1)
	const username = bots.Username("user1_updated")

	err := s.repos.UpsertUsername(s.ctx, bots.UserID(userID), username)
	s.Require().NoError(err)

	var got userDBRow
	err = s.db.GetContext(s.ctx, &got, `
		SELECT id, username, updated_at
		FROM users
		WHERE id = $1
	`, userID)
	s.Require().NoError(err)
	s.Require().Equal(userID, got.ID)
	s.Require().Equal(username.String(), got.Username)
	s.Require().True(got.UpdatedAt.After(time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC)))
}

func (s *RepositoryTestSuite) TestUserRepository_UpsertUsername_UpdateExistingUser_RefreshesUpdatedAt() {
	const userID = int64(2)

	var before userDBRow
	err := s.db.GetContext(s.ctx, &before, `
		SELECT id, username, updated_at
		FROM users
		WHERE id = $1
	`, userID)
	s.Require().NoError(err)
	s.Require().Equal(time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC), before.UpdatedAt)

	err = s.repos.UpsertUsername(s.ctx, bots.UserID(userID), "user2_new")
	s.Require().NoError(err)

	var after userDBRow
	err = s.db.GetContext(s.ctx, &after, `
		SELECT id, username, updated_at
		FROM users
		WHERE id = $1
	`, userID)
	s.Require().NoError(err)
	s.Require().Equal("user2_new", after.Username)
	s.Require().True(after.UpdatedAt.After(before.UpdatedAt))
}
