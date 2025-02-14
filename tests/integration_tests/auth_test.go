package integration_tests

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/entity"
	errs "Avito-Backend-trainee-assignment-winter-2025/internal/pkg/errors"
	"Avito-Backend-trainee-assignment-winter-2025/internal/storage/postgres"
	"context"
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

const UserCoinsOnRegister = int32(1000)

type IAuthSuite struct {
	suite.Suite
	repo                entity.IAuthRepository
	builder             squirrel.StatementBuilderType
	userCoinsOnRegister int32
}

func (s *IAuthSuite) SetupSuite() {
	s.repo = postgres.NewAuthRepository(testDbInstance)
	s.builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	s.userCoinsOnRegister = UserCoinsOnRegister
}

func (s *IAuthSuite) TearDownSubTest() {
	query := `truncate table users cascade`
	_, err := testDbInstance.Exec(context.Background(), query)
	require.NoError(s.T(), err)
}

func (s *IAuthSuite) TestAuthRepository_Register() {
	testCases := []struct {
		name        string
		authInfo    *entity.Auth
		beforeTest  func()
		wantErr     bool
		requiredErr error
	}{
		{
			name: "успешная регистрация",
			authInfo: &entity.Auth{
				Username: "test",
				Password: "hashedPass",
			},
			wantErr: false,
		}, // успешная регистрация
		{
			name: "пользователь уже существует",
			authInfo: &entity.Auth{
				Username: "test",
				Password: "hashedPass",
			},
			beforeTest: func() {
				query, args, err := s.builder.
					Insert("users").
					Columns("username", "password").
					Values("test", "hashedPass").
					ToSql()
				require.NoError(s.T(), err)

				_, err = testDbInstance.Exec(
					context.Background(),
					query,
					args...,
				)
				require.NoError(s.T(), err)
			},
			wantErr:     true,
			requiredErr: errs.UserAlreadyExists,
		}, // пользователь уже существует
	}
	for _, tt := range testCases {
		s.T().Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				s.TearDownSubTest()
			})

			if tt.beforeTest != nil {
				tt.beforeTest()
			}

			checkQuery, args, err := s.builder.
				Select("username", "password", "coins").
				From("users").
				Where(squirrel.Eq{"username": tt.authInfo.Username}).
				ToSql()
			require.NoError(t, err)

			err = s.repo.Register(context.Background(), tt.authInfo)
			dbUser := new(entity.Auth)
			var coins int32
			dbErr := testDbInstance.QueryRow(
				context.Background(),
				checkQuery,
				args...,
			).Scan(
				&dbUser.Username,
				&dbUser.Password,
				&coins,
			)
			require.NoError(t, dbErr)

			if tt.wantErr {
				require.Equal(t, tt.requiredErr, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.authInfo, dbUser)
				require.Equal(t, s.userCoinsOnRegister, coins)
			}
		})
	}
}

func (s *IAuthSuite) TestAuthRepository_GetByUsername() {
	testCases := []struct {
		name        string
		authInfo    *entity.Auth
		beforeTest  func()
		wantErr     bool
		requiredErr error
	}{
		{
			name: "успешное получение",
			authInfo: &entity.Auth{
				Username: "test",
				Password: "hashedPass",
			},
			beforeTest: func() {
				query, args, err := s.builder.
					Insert("users").
					Columns("username", "password").
					Values("test", "hashedPass").
					ToSql()
				require.NoError(s.T(), err)

				_, err = testDbInstance.Exec(
					context.Background(),
					query,
					args...,
				)
				require.NoError(s.T(), err)
			},
			wantErr: false,
		}, // успешное получение
		{
			name: "успешное получение",
			authInfo: &entity.Auth{
				Username: "test",
				Password: "hashedPass",
			},
			wantErr:     true,
			requiredErr: nil,
		}, // пользователь не найден
	}
	for _, tt := range testCases {
		s.T().Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				s.TearDownSubTest()
			})

			if tt.beforeTest != nil {
				tt.beforeTest()
			}

			checkQuery, args, err := s.builder.
				Select("username", "password").
				From("users").
				Where(squirrel.Eq{"username": tt.authInfo.Username}).
				ToSql()
			require.NoError(t, err)

			user, err := s.repo.GetByUsername(context.Background(), tt.authInfo.Username)
			dbUser := new(entity.Auth)
			dbErr := testDbInstance.QueryRow(
				context.Background(),
				checkQuery,
				args...,
			).Scan(
				&dbUser.Username,
				&dbUser.Password,
			)
			if errors.Is(dbErr, pgx.ErrNoRows) {
				dbUser = nil
			} else if dbErr != nil {
				require.NoError(t, dbErr)
			}

			if tt.wantErr {
				require.Equal(t, tt.requiredErr, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, dbUser, user)
			}
		})
	}
}

func TestIAuthTestSuite(t *testing.T) {
	suite.Run(t, new(IAuthSuite))
}
