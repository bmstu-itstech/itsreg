package postgres_test

import (
	"time"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func (s *RepositoryTestSuite) TestRunRepository_Run_Success() {
	run, err := s.repos.Run(s.ctx, "r0002")
	s.Require().NoError(err)
	s.Require().NotNil(run)
	s.Require().Equal(bots.RunID("r0002"), run.ID())
	s.Require().Equal(bots.BotID("b0001"), run.BotID())
	s.Require().Equal(bots.Token("token_b0001"), run.Token())
	s.Require().Equal(bots.RunStatusFailed, run.Status())
	s.Require().NotNil(run.ErrorMsg())
	s.Require().Equal("Some error occurred", *run.ErrorMsg())
	s.Require().NotNil(run.StartedAt())
	s.Require().Equal(time.Date(2026, 4, 11, 10, 30, 0, 0, time.UTC), *run.StartedAt())
	s.Require().NotNil(run.StoppedAt())
	s.Require().Equal(time.Date(2026, 4, 11, 10, 45, 0, 0, time.UTC), *run.StoppedAt())
}

func (s *RepositoryTestSuite) TestRunRepository_Run_NotFound() {
	run, err := s.repos.Run(s.ctx, "r9999")
	s.Require().ErrorIs(err, port.ErrRunNotFound)
	s.Require().Nil(run)
}

func (s *RepositoryTestSuite) TestRunRepository_RunsByOwnerID_Empty() {
	res, err := s.repos.RunsByOwnerID(s.ctx, bots.UserID(-1), port.RunsFilter{})
	s.Require().NoError(err)
	s.Require().Empty(res)
}

func (s *RepositoryTestSuite) TestRunRepository_RunsByOwnerID_ExcludeDeletedBots() {
	res, err := s.repos.RunsByOwnerID(s.ctx, bots.UserID(2), port.RunsFilter{})
	s.Require().NoError(err)
	s.Require().Len(res, 1)

	run := res[0]
	s.Require().Equal(bots.RunID("r0004"), run.ID())
	s.Require().Equal(bots.BotID("b0003"), run.BotID())
	s.Require().Equal(bots.Token("token_b0003"), run.Token())
	s.Require().Equal(bots.RunStatusStopping, run.Status())
}

func (s *RepositoryTestSuite) TestRunRepository_RunsByOwnerID_MultipleRuns() {
	res, err := s.repos.RunsByOwnerID(s.ctx, bots.UserID(1), port.RunsFilter{})
	s.Require().NoError(err)
	s.Require().Len(res, 3)

	ids := make([]bots.RunID, 0, len(res))
	for _, run := range res {
		ids = append(ids, run.ID())
	}
	s.Require().ElementsMatch([]bots.RunID{"r0001", "r0002", "r0003"}, ids)
}

func (s *RepositoryTestSuite) TestRunRepository_RunsByOwnerID_FilterByBotID() {
	botID := bots.BotID("b0001")
	res, err := s.repos.RunsByOwnerID(s.ctx, bots.UserID(1), port.RunsFilter{BotID: &botID})
	s.Require().NoError(err)
	s.Require().Len(res, 2)

	ids := make([]bots.RunID, 0, len(res))
	for _, run := range res {
		ids = append(ids, run.ID())
	}
	s.Require().ElementsMatch([]bots.RunID{"r0001", "r0002"}, ids)
}

func (s *RepositoryTestSuite) TestRunRepository_RunsByOwnerID_FilterByStatus() {
	status := bots.RunStatusActive
	res, err := s.repos.RunsByOwnerID(s.ctx, bots.UserID(1), port.RunsFilter{Status: &status})
	s.Require().NoError(err)
	s.Require().Len(res, 1)
	s.Require().Equal(bots.RunID("r0003"), res[0].ID())
}

func (s *RepositoryTestSuite) TestRunRepository_RunsByOwnerID_FilterByBotIDAndStatus() {
	botID := bots.BotID("b0001")
	status := bots.RunStatusFailed
	res, err := s.repos.RunsByOwnerID(s.ctx, bots.UserID(1), port.RunsFilter{BotID: &botID, Status: &status})
	s.Require().NoError(err)
	s.Require().Len(res, 1)
	s.Require().Equal(bots.RunID("r0002"), res[0].ID())
}

func (s *RepositoryTestSuite) TestRunRepository_RunsByOwnerID_FilterByDeletedBotID_Empty() {
	botID := bots.BotID("b0004")
	res, err := s.repos.RunsByOwnerID(s.ctx, bots.UserID(2), port.RunsFilter{BotID: &botID})
	s.Require().NoError(err)
	s.Require().Empty(res)
}

func (s *RepositoryTestSuite) TestRunRepository_ActiveRuns() {
	res, err := s.repos.ActiveRuns(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(res, 2)

	ids := make([]bots.RunID, 0, len(res))
	for _, run := range res {
		ids = append(ids, run.ID())
	}
	s.Require().ElementsMatch([]bots.RunID{"r0001", "r0003"}, ids)
}

func (s *RepositoryTestSuite) TestRunRepository_SaveRun_Success() {
	run, err := bots.NewRun("b0003", "token_b0003")
	s.Require().NoError(err)

	err = s.repos.SaveRun(s.ctx, run)
	s.Require().NoError(err)

	got, err := s.repos.Run(s.ctx, run.ID())
	s.Require().NoError(err)
	s.Require().Equal(run.ID(), got.ID())
	s.Require().Equal(run.BotID(), got.BotID())
	s.Require().Equal(run.Token(), got.Token())
	s.Require().Equal(run.Status(), got.Status())
	s.Require().Nil(got.ErrorMsg())
	s.Require().Nil(got.StartedAt())
	s.Require().Nil(got.StoppedAt())
}

func (s *RepositoryTestSuite) TestRunRepository_SaveRun_FailedWhenActiveRunAlreadyExists() {
	run, err := bots.NewRun("b0001", "token_b0001")
	s.Require().NoError(err)

	err = s.repos.SaveRun(s.ctx, run)
	s.Require().ErrorIs(err, port.ErrActiveRunAlreadyExists)
}

func (s *RepositoryTestSuite) TestRunRepository_UpdateRun_Success() {
	run, err := bots.NewRun("b0003", "token_b0003")
	s.Require().NoError(err)

	err = s.repos.SaveRun(s.ctx, run)
	s.Require().NoError(err)

	err = run.Started()
	s.Require().NoError(err)

	err = s.repos.UpdateRun(s.ctx, run)
	s.Require().NoError(err)

	got, err := s.repos.Run(s.ctx, run.ID())
	s.Require().NoError(err)
	s.Require().Equal(run.ID(), got.ID())
	s.Require().Equal(run.BotID(), got.BotID())
	s.Require().Equal(run.Token(), got.Token())
	s.Require().Equal(bots.RunStatusActive, got.Status())
	s.Require().NotNil(got.StartedAt())
	s.Require().WithinDuration(*run.StartedAt(), *got.StartedAt(), time.Second)
	s.Require().Nil(got.StoppedAt())
	s.Require().Nil(got.ErrorMsg())
}

func (s *RepositoryTestSuite) TestRunRepository_UpdateRun_NonExistentRun() {
	run, err := bots.NewRun("b0003", "token_b0003")
	s.Require().NoError(err)

	err = s.repos.UpdateRun(s.ctx, run)
	s.Require().ErrorIs(err, port.ErrRunNotFound)
}
