package postgres_test

import (
	"time"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
)

func (s *RepositoryTestSuite) TestOwnedRunProvider_OwnedRun_Success() {
	res, err := s.repos.OwnedRun(s.ctx, "r0002")
	s.Require().NoError(err)
	s.Require().Equal(
		dto.OwnedRun{
			ID:        "r0002",
			OwnerID:   1,
			BotID:     "b0001",
			Status:    "failed",
			ErrorMsg:  ptr("Some error occurred"),
			StartedAt: ptr(time.Date(2026, 4, 11, 10, 30, 0, 0, time.UTC)),
			StoppedAt: ptr(time.Date(2026, 4, 11, 10, 45, 0, 0, time.UTC)),
		},
		res,
	)
}

func (s *RepositoryTestSuite) TestOwnedRunProvider_OwnedRun_NotFound() {
	res, err := s.repos.OwnedRun(s.ctx, "r9999")
	s.Require().ErrorIs(err, port.ErrRunNotFound)
	s.Require().Equal(dto.OwnedRun{}, res)
}

func (s *RepositoryTestSuite) TestOwnedRunProvider_OwnedRun_ExcludeDeletedBots() {
	_, err := s.db.ExecContext(s.ctx, `
		INSERT INTO runs
			(id, bot_id, token, status, error_msg, started_at, stopped_at)
		VALUES
			($1, $2, $3, $4, NULL, $5, NULL)
	`, "r9001", "b0004", "token_b0004", "active", time.Date(2026, 4, 11, 10, 55, 0, 0, time.UTC))
	s.Require().NoError(err)

	res, err := s.repos.OwnedRun(s.ctx, "r9001")
	s.Require().ErrorIs(err, port.ErrRunNotFound)
	s.Require().Equal(dto.OwnedRun{}, res)
}
