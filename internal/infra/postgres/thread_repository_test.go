package postgres_test

import (
	"time"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func (s *RepositoryTestSuite) TestThreadRepository_LastUserThread_Success() {
	thread, err := s.repos.LastUserThread(s.ctx, "b0001", bots.UserID(1))
	s.Require().NoError(err)
	s.Require().NotNil(thread)

	s.Require().Equal(bots.ThreadID("t0001"), thread.ID())
	s.Require().Equal(bots.BotID("b0001"), thread.BotID())
	s.Require().Equal(bots.UserID(1), thread.UserID())
	s.Require().Equal(bots.EntryKey("start"), thread.Key())
	s.Require().Equal(bots.MustNewState(1), thread.State())
	s.Require().Equal(time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC), thread.StartedAt())
	s.Require().Equal(time.Date(2026, 4, 11, 10, 30, 0, 0, time.UTC), thread.UpdatedAt())

	answers := thread.Answers()
	s.Require().Len(answers, 2)
	s.Require().Equal("Answer 1 for t0001:1", answers[bots.MustNewState(1)].String())
	s.Require().Equal("Answer 2 for t0001:2", answers[bots.MustNewState(2)].String())
}

func (s *RepositoryTestSuite) TestThreadRepository_LastUserThread_NotFound() {
	thread, err := s.repos.LastUserThread(s.ctx, "b9999", bots.UserID(1))
	s.Require().ErrorIs(err, port.ErrUserHasNotThreads)
	s.Require().Nil(thread)
}

func (s *RepositoryTestSuite) TestThreadRepository_LastUserThread_ReturnsLatestByStartedAt() {
	entry := bots.MustNewEntry("start", bots.MustNewState(1))
	newThread := bots.MustNewThread("b0001", bots.UserID(1), entry)
	newThread.SaveAnswer(bots.MustNewMessage("latest answer"))

	err := s.repos.SaveThread(s.ctx, newThread)
	s.Require().NoError(err)

	got, err := s.repos.LastUserThread(s.ctx, "b0001", bots.UserID(1))
	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Require().Equal(newThread.ID(), got.ID())
	s.Require().Equal("latest answer", got.Answers()[bots.MustNewState(1)].String())
}

func (s *RepositoryTestSuite) TestThreadRepository_SaveThread_Success() {
	entry := bots.MustNewEntry("start", bots.MustNewState(1))
	want := bots.MustNewThread("b0003", bots.UserID(1), entry)
	want.SaveAnswer(bots.MustNewMessage("answer for state 1"))

	err := s.repos.SaveThread(s.ctx, want)
	s.Require().NoError(err)

	got, err := s.repos.LastUserThread(s.ctx, want.BotID(), want.UserID())
	s.Require().NoError(err)
	s.Require().NotNil(got)

	s.Require().Equal(want.ID(), got.ID())
	s.Require().Equal(want.BotID(), got.BotID())
	s.Require().Equal(want.UserID(), got.UserID())
	s.Require().Equal(want.Key(), got.Key())
	s.Require().Equal(want.State(), got.State())
	s.Require().Equal("answer for state 1", got.Answers()[bots.MustNewState(1)].String())
	s.Require().WithinDuration(want.StartedAt(), got.StartedAt(), time.Second)
	s.Require().WithinDuration(want.UpdatedAt(), got.UpdatedAt(), time.Second)
}

func (s *RepositoryTestSuite) TestThreadRepository_SaveThread_FailedToSaveThreadTwice() {
	entry := bots.MustNewEntry("start", bots.MustNewState(1))
	thread := bots.MustNewThread("b0003", bots.UserID(1), entry)

	err := s.repos.SaveThread(s.ctx, thread)
	s.Require().NoError(err)

	err = s.repos.SaveThread(s.ctx, thread)
	s.Require().ErrorIs(err, port.ErrThreadAlreadyExists)
}

func (s *RepositoryTestSuite) TestThreadRepository_UpdateThread_Success() {
	entry := bots.MustNewEntry("start", bots.MustNewState(1))
	thread := bots.MustNewThread("b0003", bots.UserID(1), entry)
	thread.SaveAnswer(bots.MustNewMessage("answer state1 v1"))

	err := s.repos.SaveThread(s.ctx, thread)
	s.Require().NoError(err)

	thread.StepTo(bots.MustNewState(2))
	thread.SaveAnswer(bots.MustNewMessage("answer state2 v1"))
	thread.StepTo(bots.MustNewState(1))
	thread.SaveAnswer(bots.MustNewMessage("answer state1 v2"))

	err = s.repos.UpdateThread(s.ctx, thread)
	s.Require().NoError(err)

	got, err := s.repos.LastUserThread(s.ctx, thread.BotID(), thread.UserID())
	s.Require().NoError(err)
	s.Require().NotNil(got)

	s.Require().Equal(thread.ID(), got.ID())
	s.Require().Equal(bots.MustNewState(1), got.State())
	s.Require().Equal("answer state1 v2", got.Answers()[bots.MustNewState(1)].String())
	s.Require().Equal("answer state2 v1", got.Answers()[bots.MustNewState(2)].String())
	s.Require().WithinDuration(thread.StartedAt(), got.StartedAt(), time.Second)
	s.Require().WithinDuration(thread.UpdatedAt(), got.UpdatedAt(), time.Second)
}

func (s *RepositoryTestSuite) TestThreadRepository_UpdateThread_NonExistentThread() {
	entry := bots.MustNewEntry("start", bots.MustNewState(1))
	thread := bots.MustNewThread("b0003", bots.UserID(1), entry)

	err := s.repos.UpdateThread(s.ctx, thread)
	s.Require().ErrorIs(err, port.ErrThreadNotFound)
}
