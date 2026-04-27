package postgres_test

import (
	"time"

	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
)

func (s *RepositoryTestSuite) TestOwnedMailingProvider_OwnedMailing_Success() {
	got, err := s.repos.OwnedMailing(s.ctx, "m10001")
	s.Require().NoError(err)
	s.Require().Equal("m10001", got.ID)
	s.Require().Equal(int64(1), got.OwnerID)
	s.Require().Equal("b0001", got.BotID)
	s.Require().Equal("Mailing 1", got.Name)
	s.Require().Equal("start", got.EntryKey)
	s.Require().Equal("completed", got.Status)
	s.Require().Equal(1, got.SuccessCount)
	s.Require().Equal(1, got.FailCount)
	s.Require().Equal(0, got.PendingCount)
	s.Require().Equal(2, got.TotalCount)
	s.Require().Equal(time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC), got.CreatedAt)
	s.Require().NotNil(got.StartedAt)
	s.Require().Equal(time.Date(2026, 4, 12, 10, 5, 0, 0, time.UTC), *got.StartedAt)
	s.Require().NotNil(got.CompletedAt)
	s.Require().Equal(time.Date(2026, 4, 12, 10, 10, 0, 0, time.UTC), *got.CompletedAt)
	s.Require().ElementsMatch([]int64{1, 2}, got.Recipients)

	resultsByUserID := mapOwnedMailingResultsByUserID(got.Results)
	s.Require().Len(resultsByUserID, 2)
	s.Require().True(resultsByUserID[1].Success)
	s.Require().Nil(resultsByUserID[1].ErrorMsg)
	s.Require().False(resultsByUserID[2].Success)
	s.Require().NotNil(resultsByUserID[2].ErrorMsg)
	s.Require().Equal("cannot send message", *resultsByUserID[2].ErrorMsg)
}

func (s *RepositoryTestSuite) TestOwnedMailingProvider_OwnedMailing_NotFound() {
	got, err := s.repos.OwnedMailing(s.ctx, "m99999")
	s.Require().ErrorIs(err, port.ErrMailingNotFound)
	s.Require().Equal(dto.OwnedMailing{}, got)
}

func (s *RepositoryTestSuite) TestOwnedMailingProvider_OwnedMailing_ExcludeDeletedBots() {
	got, err := s.repos.OwnedMailing(s.ctx, "m10003")
	s.Require().ErrorIs(err, port.ErrMailingNotFound)
	s.Require().Equal(dto.OwnedMailing{}, got)
}

func mapOwnedMailingResultsByUserID(results []dto.UserMailingResult) map[int64]dto.UserMailingResult {
	res := make(map[int64]dto.UserMailingResult, len(results))
	for _, result := range results {
		res[result.UserID] = result
	}
	return res
}
