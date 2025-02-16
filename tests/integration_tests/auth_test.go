package integration_tests

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/entity"
	errs "Avito-Backend-trainee-assignment-winter-2025/internal/pkg/errors"
	"Avito-Backend-trainee-assignment-winter-2025/internal/storage/postgres"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const userCoinsOnRegister = int32(1000)

type IAuthSuite struct {
	suite.Suite
	repo                entity.IAuthRepository
	builder             squirrel.StatementBuilderType
	userCoinsOnRegister int32
}

func (s *IAuthSuite) SetupSuite() {
	s.repo = postgres.NewAuthRepository(testDbInstance)
	s.builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	s.userCoinsOnRegister = userCoinsOnRegister
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
		beforeTest  func(t *testing.T)
		check       func(t *testing.T, auth *entity.Auth) error
		wantErr     bool
		requiredErr error
	}{
		{
			name: "успешная регистрация",
			authInfo: &entity.Auth{
				Username: "test",
				Password: "hashedPass",
			},
			check: func(t *testing.T, auth *entity.Auth) error {
				checkQuery, args, err := s.builder.
					Select("username", "password", "coins").
					From("users").
					Where(squirrel.Eq{"username": auth.Username}).
					ToSql()
				require.NoError(t, err)

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

				if dbUser.Username != auth.Username {
					return fmt.Errorf("invalid username")
				}
				if dbUser.Password != auth.Password {
					return fmt.Errorf("invalid password")
				}
				if coins != userCoinsOnRegister {
					return fmt.Errorf("invalid coins")
				}
				return nil
			},
			wantErr: false,
		}, // успешная регистрация
		{
			name: "пользователь уже существует",
			authInfo: &entity.Auth{
				Username: "test",
				Password: "hashedPass",
			},
			beforeTest: func(t *testing.T) {
				query, args, err := s.builder.
					Insert("users").
					Columns("username", "password").
					Values("test", "hashedPass").
					ToSql()
				require.NoError(t, err)

				_, err = testDbInstance.Exec(
					context.Background(),
					query,
					args...,
				)
				require.NoError(t, err)
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
				tt.beforeTest(s.T())
			}

			err := s.repo.Register(context.Background(), tt.authInfo)
			var checkErr error
			if tt.check != nil {
				checkErr = tt.check(t, tt.authInfo)
			}

			if tt.wantErr {
				require.Equal(t, tt.requiredErr, err)
			} else {
				require.NoError(t, err)
				require.NoError(t, checkErr)
			}
		})
	}
}

func (s *IAuthSuite) TestAuthRepository_GetByUsername() {
	testCases := []struct {
		name        string
		username    string
		beforeTest  func(t *testing.T)
		check       func(t *testing.T, username string, authFromRepo *entity.Auth) error
		wantErr     bool
		requiredErr error
	}{
		{
			name:     "успешное получение",
			username: "test",
			beforeTest: func(t *testing.T) {
				query, args, err := s.builder.
					Insert("users").
					Columns("username", "password").
					Values("test", "hashedPass").
					ToSql()
				require.NoError(t, err)

				_, err = testDbInstance.Exec(
					context.Background(),
					query,
					args...,
				)
				require.NoError(t, err)
			},
			check: func(t *testing.T, username string, authFromRepo *entity.Auth) error {
				checkQuery, args, err := s.builder.
					Select("username", "password").
					From("users").
					Where(squirrel.Eq{"username": username}).
					ToSql()
				require.NoError(t, err)

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

				if dbUser.Username != authFromRepo.Username {
					return fmt.Errorf("invalid username")
				}
				if dbUser.Password != authFromRepo.Password {
					return fmt.Errorf("invalid password")
				}

				return nil
			},
			wantErr: false,
		}, // успешное получение
		{
			name:        "пользователь не найден",
			username:    "test",
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
				tt.beforeTest(s.T())
			}

			user, err := s.repo.GetByUsername(context.Background(), tt.username)
			var checkErr error
			if tt.check != nil {
				checkErr = tt.check(s.T(), tt.username, user)
			}

			if tt.wantErr {
				require.Equal(t, tt.requiredErr, err)
			} else {
				require.NoError(t, err)
				require.NoError(t, checkErr)
			}
		})
	}
}

func TestIAuthTestSuite(t *testing.T) {
	suite.Run(t, new(IAuthSuite))
}
