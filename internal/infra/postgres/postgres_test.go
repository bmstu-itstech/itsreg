package postgres_test

import (
	"context"
	"io/fs"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"

	"github.com/bmstu-itstech/itsreg/internal/domain/bots"
	"github.com/bmstu-itstech/itsreg/internal/infra/postgres"
	"github.com/bmstu-itstech/itsreg/internal/infra/postgres/fixtures"
	"github.com/bmstu-itstech/itsreg/migrations"
	"github.com/bmstu-itstech/itsreg/pkg/tests"
)

const actorUserID = bots.UserID(42)

func TestRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration tests in short mode")
	}
	suite.Run(t, new(RepositoryTestSuite))
}

type RepositoryTestSuite struct {
	suite.Suite

	cont  *tests.PostgresContainer
	db    *sqlx.DB
	repos *postgres.Repository
	ctx   context.Context
}

func (s *RepositoryTestSuite) SetupSuite() {
	s.ctx = context.Background()

	cont, err := tests.SetupPostgres(s.ctx)
	s.Require().NoError(err)
	s.cont = cont

	db, err := sqlx.ConnectContext(s.ctx, "postgres", cont.URI)
	s.Require().NoError(err)
	s.db = db

	err = applyMigrations(s.cont.URI)
	s.Require().NoError(err)

	repos := postgres.NewRepositoryFromDB(db)
	s.repos = repos
}

func (s *RepositoryTestSuite) SetupTest() {
	_, err := s.db.Exec(`
		TRUNCATE TABLE 
		    answers,
			threads,
			users,
			runs,
			bots,
			options,
			messages,
			edges,
			entries,
			nodes,
			scripts
		CASCADE;
	`)
	s.Require().NoError(err)

	err = applyFixtures(s.db)
	s.Require().NoError(err)
}

func (s *RepositoryTestSuite) TearDownSuite() {
	if s.db != nil {
		err := s.db.Close()
		if err != nil {
			s.T().Log(err)
		}
	}
	if s.cont != nil {
		err := s.cont.Container.Terminate(s.ctx)
		if err != nil {
			s.T().Log(err)
		}
	}
}

func applyMigrations(uri string) error {
	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, uri)
	if err != nil {
		return err
	}
	defer func(m *migrate.Migrate) {
		_, _ = m.Close()
	}(m)

	return m.Up()
}

func applyFixtures(db *sqlx.DB) error {
	fixturesFiles, err := fs.Glob(fixtures.FS, "*.sql")
	if err != nil {
		return err
	}

	for _, f := range fixturesFiles {
		query, err2 := fs.ReadFile(fixtures.FS, f)
		if err2 != nil {
			return err2
		}
		_, err2 = db.Exec(string(query))
		if err2 != nil {
			return err2
		}
	}

	return nil
}
