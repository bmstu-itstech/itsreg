package postgres_test

import (
	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
)

func (s *RepositoryTestSuite) TestBotMetaProvider_BotMeta_Success() {
	meta, err := s.repos.BotMeta(s.ctx, "b0001")
	s.Require().NoError(err)
	s.Require().Equal(
		dto.BotMeta{
			ID:       "b0001",
			OwnerID:  1,
			ScriptID: "sc0001",
			Token:    "token_b0001",
			Deleted:  false,
		},
		meta,
	)
}

func (s *RepositoryTestSuite) TestBotMetaProvider_BotMeta_FetchDeleted() {
	meta, err := s.repos.BotMeta(s.ctx, "b0004")
	s.Require().NoError(err)
	s.Require().Equal(
		dto.BotMeta{
			ID:       "b0004",
			OwnerID:  2,
			ScriptID: "sc0004",
			Token:    "token_b0004",
			Deleted:  true,
		},
		meta,
	)
}

func (s *RepositoryTestSuite) TestBotMetaProvider_BotMeta_NotFound() {
	meta, err := s.repos.BotMeta(s.ctx, "b000x")
	s.Require().ErrorIs(err, port.ErrBotNotFound)
	s.Require().Equal(dto.BotMeta{}, meta)
}
