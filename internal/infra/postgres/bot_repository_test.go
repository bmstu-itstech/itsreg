package postgres_test

import (
	"time"

	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
)

func (s *RepositoryTestSuite) TestBotRepository_Bot_Success() {
	bot, err := s.repos.Bot(s.ctx, "b0001")
	s.Require().NoError(err)
	s.Require().NotNil(bot)
	s.Require().Equal(bots.BotID("b0001"), bot.ID())
	s.Require().Equal(bots.UserID(1), bot.OwnerID())
	s.Require().Equal(bots.ScriptID("sc0001"), bot.ScriptID())
	s.Require().Equal(bots.Token("token_b0001"), bot.Token())
	s.Require().Equal("Test bot b0001", bot.Desc())
	s.Require().Equal(time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC), bot.CreatedAt())
	s.Require().Equal(time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC), bot.UpdatedAt())
	s.Require().Nil(bot.DeletedAt())
}

func (s *RepositoryTestSuite) TestBotRepository_Bot_FetchDeleted() {
	bot, err := s.repos.Bot(s.ctx, "b0004")
	s.Require().NoError(err)
	s.Require().NotNil(bot)
	s.Require().NotNil(bot.DeletedAt())
	deletedAt := *bot.DeletedAt()
	s.Require().Equal(time.Date(2026, 4, 11, 11, 0, 0, 0, time.UTC), deletedAt)
}

func (s *RepositoryTestSuite) TestBotRepository_Bot_NotFound() {
	bot, err := s.repos.Bot(s.ctx, "b000x")
	s.Require().ErrorIs(err, port.ErrBotNotFound)
	s.Require().Nil(bot)
}

func (s *RepositoryTestSuite) TestBotRepository_BotsByOwnerID_Empty() {
	res, err := s.repos.BotsByOwnerID(s.ctx, bots.UserID(-1))
	s.Require().NoError(err)
	s.Require().Empty(res)
}

func (s *RepositoryTestSuite) TestBotRepository_BotsByOwnerID_OneBot() {
	// User(2) имеет две записи в таблице bots, но одна из них удалена (deleted_at IS NOT NULL)
	res, err := s.repos.BotsByOwnerID(s.ctx, bots.UserID(2))
	s.Require().NoError(err)
	s.Require().Len(res, 1)
	bot := res[0]
	s.Require().Equal(bots.BotID("b0003"), bot.ID())
	s.Require().Equal(bots.UserID(2), bot.OwnerID())
	s.Require().Equal(bots.ScriptID("sc0003"), bot.ScriptID())
	s.Require().Equal(bots.Token("token_b0003"), bot.Token())
	s.Require().Equal("Test bot b0003", bot.Desc())
	s.Require().Equal(time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC), bot.CreatedAt())
	s.Require().Equal(time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC), bot.UpdatedAt())
	s.Require().Nil(bot.DeletedAt())
}

func (s *RepositoryTestSuite) TestBotRepository_BotsByOwnerID_MultipleBots() {
	res, err := s.repos.BotsByOwnerID(s.ctx, bots.UserID(1))
	s.Require().NoError(err)
	s.Require().Len(res, 2)

	// User(1) имеет два бота, и порядок определяется столбцом updated_at; второй бот имеет более позднее время
	// обновления, поэтому он должен быть первым в списке.
	bot2 := res[0]
	s.Require().Equal(bots.BotID("b0002"), bot2.ID())
	s.Require().Equal(bots.UserID(1), bot2.OwnerID())
	s.Require().Equal(bots.ScriptID("sc0002"), bot2.ScriptID())
	s.Require().Equal(bots.Token("token_b0002"), bot2.Token())
	s.Require().Equal("Test bot b0002", bot2.Desc())
	s.Require().Equal(time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC), bot2.CreatedAt())
	s.Require().Equal(time.Date(2026, 4, 11, 11, 0, 0, 0, time.UTC), bot2.UpdatedAt())

	bot1 := res[1]
	s.Require().Equal(bots.BotID("b0001"), bot1.ID())
	s.Require().Equal(bots.UserID(1), bot1.OwnerID())
	s.Require().Equal(bots.ScriptID("sc0001"), bot1.ScriptID())
	s.Require().Equal(bots.Token("token_b0001"), bot1.Token())
	s.Require().Equal("Test bot b0001", bot1.Desc())
	s.Require().Equal(time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC), bot1.CreatedAt())
	s.Require().Equal(time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC), bot1.UpdatedAt())
}

func (s *RepositoryTestSuite) TestBotRepository_SaveBot_Success() {
	want := bots.MustNewBot(actorUserID, "sc0001", "token", "Test bot")
	err := s.repos.SaveBot(s.ctx, want)
	s.Require().NoError(err)

	got, err := s.repos.Bot(s.ctx, want.ID())
	s.Require().NoError(err)

	s.Require().Equal(want.ID(), got.ID())
	s.Require().Equal(want.OwnerID(), got.OwnerID())
	s.Require().Equal(want.ScriptID(), got.ScriptID())
	s.Require().Equal(want.Token(), got.Token())
	s.Require().Equal(want.Desc(), got.Desc())
	s.Require().WithinDuration(want.CreatedAt(), got.CreatedAt(), time.Second)
	s.Require().WithinDuration(want.UpdatedAt(), got.UpdatedAt(), time.Second)
	s.Require().Nil(got.DeletedAt())
}

func (s *RepositoryTestSuite) TestBotRepository_SaveBot_FailedToSaveBotTwice() {
	bot := bots.MustNewBot(actorUserID, "sc0001", "token", "Test bot")
	err := s.repos.SaveBot(s.ctx, bot)
	s.Require().NoError(err)

	err = s.repos.SaveBot(s.ctx, bot)
	s.Require().ErrorIs(err, port.ErrBotAlreadyExists)
}

func (s *RepositoryTestSuite) TestBotRepository_UpdateBot_Success() {
	bot := bots.MustNewBot(actorUserID, "sc0001", "token", "Test bot")
	err := s.repos.SaveBot(s.ctx, bot)
	s.Require().NoError(err)

	err = bot.SetScriptID("sc0002")
	s.Require().NoError(err)
	err = bot.SetToken("token_b0002")
	s.Require().NoError(err)
	err = bot.SetDesc("Test bot (updated)")
	s.Require().NoError(err)

	err = s.repos.UpdateBot(s.ctx, bot)
	s.Require().NoError(err)

	got, err := s.repos.Bot(s.ctx, bot.ID())
	s.Require().NoError(err)

	want := bot
	s.Require().Equal(want.ID(), got.ID())
	s.Require().Equal(want.OwnerID(), got.OwnerID())
	s.Require().Equal(want.ScriptID(), got.ScriptID())
	s.Require().Equal(want.Token(), got.Token())
	s.Require().Equal(want.Desc(), got.Desc())
	s.Require().WithinDuration(want.CreatedAt(), got.CreatedAt(), time.Second)
	s.Require().WithinDuration(want.UpdatedAt(), got.UpdatedAt(), time.Second)
	s.Require().Nil(got.DeletedAt())
}

func (s *RepositoryTestSuite) TestBotRepository_UpdateBot_DeleteBot() {
	bot := bots.MustNewBot(actorUserID, "sc0001", "token", "Test bot")
	err := s.repos.SaveBot(s.ctx, bot)
	s.Require().NoError(err)

	err = bot.Delete()
	s.Require().NoError(err)

	err = s.repos.UpdateBot(s.ctx, bot)
	s.Require().NoError(err)

	got, err := s.repos.Bot(s.ctx, bot.ID())
	s.Require().NoError(err)

	want := bot
	s.Require().Equal(want.ID(), got.ID())
	s.Require().Equal(want.OwnerID(), got.OwnerID())
	s.Require().Equal(want.ScriptID(), got.ScriptID())
	s.Require().Equal(want.Token(), got.Token())
	s.Require().Equal(want.Desc(), got.Desc())
	s.Require().WithinDuration(want.CreatedAt(), got.CreatedAt(), time.Second)
	s.Require().WithinDuration(want.UpdatedAt(), got.UpdatedAt(), time.Second)
	s.Require().NotNil(got.DeletedAt())
	deletedAt := *got.DeletedAt()
	s.Require().WithinDuration(want.UpdatedAt(), deletedAt, time.Second)
}

func (s *RepositoryTestSuite) TestBotRepository_UpdateBot_NonExistentBot() {
	bot := bots.MustNewBot(actorUserID, "sc0001", "token", "Test bot")
	err := s.repos.UpdateBot(s.ctx, bot)
	s.Require().ErrorIs(err, port.ErrBotNotFound)
}
