package postgres_test

import (
	"github.com/bmstu-itstech/itsreg/internal/app/dto"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
)

func (s *RepositoryTestSuite) TestScriptMetaProvider_ScriptMeta_Success() {
	meta, err := s.repos.ScriptMeta(s.ctx, "sc0001")
	s.Require().NoError(err)
	s.Require().Equal(
		dto.ScriptMeta{
			ID:      "sc0001",
			OwnerID: 1,
			Desc:    "Test script sc0001",
			Deleted: false,
		},
		meta,
	)
}

func (s *RepositoryTestSuite) TestScriptMetaProvider_ScriptMeta_FetchDeleted() {
	meta, err := s.repos.ScriptMeta(s.ctx, "sc0004")
	s.Require().NoError(err)
	s.Require().Equal(
		dto.ScriptMeta{
			ID:      "sc0004",
			OwnerID: 2,
			Desc:    "Test script sc0003",
			Deleted: true,
		},
		meta,
	)
}

func (s *RepositoryTestSuite) TestScriptMetaProvider_ScriptMeta_NotFound() {
	meta, err := s.repos.ScriptMeta(s.ctx, "sc000x")
	s.Require().ErrorIs(err, port.ErrScriptNotFound)
	s.Require().Equal(dto.ScriptMeta{}, meta)
}
