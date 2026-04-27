package postgres_test

import (
	"time"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func (s *RepositoryTestSuite) TestMailingRepository_Mailing_Success() {
	got, err := s.repos.Mailing(s.ctx, "m10001")
	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Require().Equal(bots.MailingID("m10001"), got.ID())
	s.Require().Equal(bots.BotID("b0001"), got.BotID())
	s.Require().Equal("Mailing 1", got.Name())
	s.Require().Equal(bots.EntryKey("start"), got.EntryKey())
	s.Require().Equal(bots.MailingStatusCompleted, got.Status())
	s.Require().Equal(time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC), got.CreatedAt())
	s.Require().NotNil(got.StartedAt())
	s.Require().Equal(time.Date(2026, 4, 12, 10, 5, 0, 0, time.UTC), *got.StartedAt())
	s.Require().NotNil(got.CompletedAt())
	s.Require().Equal(time.Date(2026, 4, 12, 10, 10, 0, 0, time.UTC), *got.CompletedAt())
	s.Require().ElementsMatch([]bots.UserID{1, 2}, got.Recipients())

	resultsByUserID := mapUserResultsByUserID(got.Results())
	s.Require().Len(resultsByUserID, 2)
	s.Require().True(resultsByUserID[1].Success())
	s.Require().Nil(resultsByUserID[1].ErrorMessage())
	s.Require().False(resultsByUserID[2].Success())
	s.Require().NotNil(resultsByUserID[2].ErrorMessage())
	s.Require().Equal("cannot send message", *resultsByUserID[2].ErrorMessage())
}

func (s *RepositoryTestSuite) TestMailingRepository_Mailing_NotFound() {
	got, err := s.repos.Mailing(s.ctx, "m99999")
	s.Require().ErrorIs(err, port.ErrMailingNotFound)
	s.Require().Nil(got)
}

func (s *RepositoryTestSuite) TestMailingRepository_MailingsByOwnerID_Empty() {
	res, err := s.repos.MailingsByOwnerID(s.ctx, bots.UserID(-1), port.MailingsFilter{})
	s.Require().NoError(err)
	s.Require().Empty(res)
}

func (s *RepositoryTestSuite) TestMailingRepository_MailingsByOwnerID_ExcludeDeletedBots() {
	res, err := s.repos.MailingsByOwnerID(s.ctx, bots.UserID(2), port.MailingsFilter{})
	s.Require().NoError(err)
	s.Require().Len(res, 1)
	s.Require().Equal(bots.MailingID("m10002"), res[0].ID())
}

func (s *RepositoryTestSuite) TestMailingRepository_MailingsByOwnerID_MultipleMailings() {
	res, err := s.repos.MailingsByOwnerID(s.ctx, bots.UserID(1), port.MailingsFilter{})
	s.Require().NoError(err)
	s.Require().Len(res, 5)

	ids := make([]bots.MailingID, 0, len(res))
	for _, mailing := range res {
		ids = append(ids, mailing.ID())
	}
	s.Require().ElementsMatch([]bots.MailingID{"m10001", "m10004", "m10005", "m10006", "m10007"}, ids)
}

func (s *RepositoryTestSuite) TestMailingRepository_MailingsByOwnerID_FilterByBotID() {
	botID := bots.BotID("b0001")
	res, err := s.repos.MailingsByOwnerID(s.ctx, bots.UserID(1), port.MailingsFilter{BotID: &botID})
	s.Require().NoError(err)
	s.Require().Len(res, 4)

	ids := make([]bots.MailingID, 0, len(res))
	for _, mailing := range res {
		ids = append(ids, mailing.ID())
	}
	s.Require().ElementsMatch([]bots.MailingID{"m10001", "m10004", "m10006", "m10007"}, ids)
}

func (s *RepositoryTestSuite) TestMailingRepository_MailingsByOwnerID_FilterByStatus() {
	status := bots.MailingStatusStarted
	res, err := s.repos.MailingsByOwnerID(s.ctx, bots.UserID(1), port.MailingsFilter{Status: &status})
	s.Require().NoError(err)
	s.Require().Len(res, 1)
	s.Require().Equal(bots.MailingID("m10006"), res[0].ID())
}

func (s *RepositoryTestSuite) TestMailingRepository_MailingsByOwnerID_FilterByBotIDAndStatus() {
	botID := bots.BotID("b0001")
	status := bots.MailingStatusStarted
	res, err := s.repos.MailingsByOwnerID(s.ctx, bots.UserID(1), port.MailingsFilter{BotID: &botID, Status: &status})
	s.Require().NoError(err)
	s.Require().Len(res, 1)
	s.Require().Equal(bots.MailingID("m10006"), res[0].ID())
}

func (s *RepositoryTestSuite) TestMailingRepository_MailingsByOwnerID_FilterByDeletedBotID_Empty() {
	botID := bots.BotID("b0004")
	res, err := s.repos.MailingsByOwnerID(s.ctx, bots.UserID(2), port.MailingsFilter{BotID: &botID})
	s.Require().NoError(err)
	s.Require().Empty(res)
}

func (s *RepositoryTestSuite) TestMailingRepository_SaveMailing_Success() {
	want, err := bots.RestoreMailing(
		"m00071",
		"b0001",
		"Mailing 1",
		"start",
		bots.MailingStatusScheduled,
		[]bots.UserID{1, 2},
		nil,
		time.Date(2026, 4, 12, 17, 0, 0, 0, time.UTC),
		nil,
		nil,
	)
	s.Require().NoError(err)

	err = s.repos.SaveMailing(s.ctx, want)
	s.Require().NoError(err)

	got, err := s.repos.Mailing(s.ctx, want.ID())
	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Require().Equal(want.ID(), got.ID())
	s.Require().Equal(want.BotID(), got.BotID())
	s.Require().Equal(want.EntryKey(), got.EntryKey())
	s.Require().Equal(want.Status(), got.Status())
	s.Require().Equal(want.CreatedAt(), got.CreatedAt())
	s.Require().Nil(got.StartedAt())
	s.Require().Nil(got.CompletedAt())
	s.Require().ElementsMatch([]bots.UserID{1, 2}, got.Recipients())
	s.Require().Empty(got.Results())
}

func (s *RepositoryTestSuite) TestMailingRepository_SaveMailing_FailedToSaveTwice() {
	m, err := bots.RestoreMailing(
		"m00081",
		"b0001",
		"Mailing 1",
		"start",
		bots.MailingStatusScheduled,
		[]bots.UserID{1},
		nil,
		time.Date(2026, 4, 12, 17, 10, 0, 0, time.UTC),
		nil,
		nil,
	)
	s.Require().NoError(err)

	err = s.repos.SaveMailing(s.ctx, m)
	s.Require().NoError(err)

	err = s.repos.SaveMailing(s.ctx, m)
	s.Require().ErrorIs(err, port.ErrMailingAlreadyExists)
}

func (s *RepositoryTestSuite) TestMailingRepository_UpdateMailing_NonExistentMailing() {
	m, err := bots.RestoreMailing(
		"m00091",
		"b0001",
		"Mailing 1",
		"start",
		bots.MailingStatusScheduled,
		[]bots.UserID{1},
		nil,
		time.Date(2026, 4, 12, 17, 20, 0, 0, time.UTC),
		nil,
		nil,
	)
	s.Require().NoError(err)

	err = s.repos.UpdateMailing(s.ctx, m)
	s.Require().ErrorIs(err, port.ErrMailingNotFound)
}

func (s *RepositoryTestSuite) TestMailingRepository_UpdateMailing_Success() {
	initial, err := bots.RestoreMailing(
		"m00101",
		"b0001",
		"Mailing 1",
		"start",
		bots.MailingStatusScheduled,
		[]bots.UserID{1, 2},
		nil,
		time.Date(2026, 4, 12, 18, 0, 0, 0, time.UTC),
		nil,
		nil,
	)
	s.Require().NoError(err)
	err = s.repos.SaveMailing(s.ctx, initial)
	s.Require().NoError(err)

	startedAt := time.Date(2026, 4, 12, 18, 1, 0, 0, time.UTC)
	updated, err := bots.RestoreMailing(
		initial.ID(),
		initial.BotID(),
		initial.Name(),
		initial.EntryKey(),
		bots.MailingStatusStarted,
		[]bots.UserID{1, 2},
		[]bots.UserMailingResult{
			bots.NewSuccessMailingResult(1),
			bots.NewErrorMailingResult(2, "temporary error"),
		},
		initial.CreatedAt(),
		&startedAt,
		nil,
	)
	s.Require().NoError(err)

	err = s.repos.UpdateMailing(s.ctx, updated)
	s.Require().NoError(err)

	got, err := s.repos.Mailing(s.ctx, initial.ID())
	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Require().Equal(bots.MailingStatusStarted, got.Status())
	s.Require().NotNil(got.StartedAt())
	s.Require().Equal(startedAt, *got.StartedAt())
	s.Require().Nil(got.CompletedAt())
	s.Require().ElementsMatch([]bots.UserID{1, 2}, got.Recipients())

	resultsByUserID := mapUserResultsByUserID(got.Results())
	s.Require().Len(resultsByUserID, 2)
	s.Require().True(resultsByUserID[1].Success())
	s.Require().False(resultsByUserID[2].Success())
	s.Require().NotNil(resultsByUserID[2].ErrorMessage())
	s.Require().Equal("temporary error", *resultsByUserID[2].ErrorMessage())
}

func mapUserResultsByUserID(results []bots.UserMailingResult) map[bots.UserID]bots.UserMailingResult {
	res := make(map[bots.UserID]bots.UserMailingResult, len(results))
	for _, result := range results {
		res[result.UserID()] = result
	}
	return res
}
