package integration_tests

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/entity"
	errs "Avito-Backend-trainee-assignment-winter-2025/internal/pkg/errors"
	"Avito-Backend-trainee-assignment-winter-2025/internal/storage/postgres"
	"context"
	"fmt"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type IUserRepoSuite struct {
	suite.Suite
	repo    entity.IUserRepository
	builder squirrel.StatementBuilderType
}

func (s *IUserRepoSuite) SetupSuite() {
	s.repo = postgres.NewUserRepository(testDbInstance)
	s.builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
}

func (s *IUserRepoSuite) TearDownSubTest() {
	query := `truncate table users cascade`
	_, err := testDbInstance.Exec(context.Background(), query)
	require.NoError(s.T(), err)
}

func (s *IUserRepoSuite) Test_userRepository_GetCoinsHistory() {
	testCases := []struct {
		name         string
		username     string
		coinsHistory *entity.CoinsHistory
		beforeTest   func(t *testing.T, username string, history *entity.CoinsHistory)
		wantErr      bool
		requiredErr  error
	}{
		{
			name:     "успешное получение истории транзакций",
			username: "user",
			coinsHistory: &entity.CoinsHistory{
				Received: []*entity.User{
					{
						Username: "first",
						Coins:    100,
					},
					{
						Username: "second",
						Coins:    200,
					},
				},
				Sent: []*entity.User{
					{
						Username: "first",
						Coins:    100,
					},
				},
			},
			beforeTest: func(t *testing.T, username string, history *entity.CoinsHistory) {
				query, args, err := s.builder.
					Insert("users").
					Columns("username", "password").
					Values(username, "hashedPass").
					Values("first", "hashedPass").
					Values("second", "hashedPass").
					ToSql()
				require.NoError(t, err)

				_, err = testDbInstance.Exec(
					context.Background(),
					query,
					args...,
				)
				require.NoError(t, err)

				insertingTransactions := s.builder.
					Insert("transactions").
					Columns("fromUser", "toUser", "coins")
				for _, fromUser := range history.Received {
					insertingTransactions = insertingTransactions.
						Values(fromUser.Username, username, fromUser.Coins)
				}
				for _, toUser := range history.Sent {
					insertingTransactions = insertingTransactions.
						Values(username, toUser.Username, toUser.Coins)
				}
				query, args, err = insertingTransactions.ToSql()
				require.NoError(t, err)

				_, err = testDbInstance.Exec(
					context.Background(),
					query,
					args...,
				)
				require.NoError(t, err)
			},
			wantErr: false,
		}, // успешное получение истории транзакций
		{
			name:     "успешное получение пустой истории транзакций",
			username: "user",
			coinsHistory: &entity.CoinsHistory{
				Received: []*entity.User{},
				Sent:     []*entity.User{},
			},
			beforeTest: func(t *testing.T, username string, history *entity.CoinsHistory) {
				query, args, err := s.builder.
					Insert("users").
					Columns("username", "password").
					Values(username, "hashedPass").
					ToSql()
				require.NoError(t, err)

				_, err = testDbInstance.Exec(
					context.Background(),
					query,
					args...,
				)
				require.NoError(t, err)
			},
			wantErr: false,
		}, // успешное получение пустой истории транзакций
		{
			name:        "пользователь не найден",
			username:    "user",
			wantErr:     true,
			requiredErr: errs.UserNotFound,
		}, // пользователь не найден
	}
	for _, tt := range testCases {
		s.T().Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				s.TearDownSubTest()
			})

			if tt.beforeTest != nil {
				tt.beforeTest(t, tt.username, tt.coinsHistory)
			}

			coins, history, err := s.repo.GetCoinsHistory(context.Background(), tt.username)

			if tt.wantErr {
				require.Equal(t, tt.requiredErr, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.coinsHistory, history)
				require.Equal(t, UserCoinsOnRegister, coins)
			}
		})
	}
}

func (s *IUserRepoSuite) Test_userRepository_SendCoins() {
	testCases := []struct {
		name        string
		transfer    *entity.TransferCoins
		beforeTest  func(t *testing.T, transfer *entity.TransferCoins)
		check       func(t *testing.T, transfer *entity.TransferCoins) error
		wantErr     bool
		requiredErr error
	}{
		{
			name: "успешная отправка монет",
			transfer: &entity.TransferCoins{
				FromUser: "first",
				ToUser:   "second",
				Amount:   UserCoinsOnRegister,
			},
			beforeTest: func(t *testing.T, transfer *entity.TransferCoins) {
				query, args, err := s.builder.
					Insert("users").
					Columns("username", "password").
					Values(transfer.FromUser, "hashedPass").
					Values(transfer.ToUser, "hashedPass").
					ToSql()
				require.NoError(t, err)

				_, err = testDbInstance.Exec(
					context.Background(),
					query,
					args...,
				)
				require.NoError(t, err)
			},
			check: func(t *testing.T, transfer *entity.TransferCoins) error {
				query, args, err := s.builder.
					Select("coins").
					From("users").
					Where(squirrel.Eq{"username": transfer.FromUser}).
					ToSql()
				require.NoError(t, err)

				var fromUserCoins int32
				err = testDbInstance.QueryRow(
					context.Background(),
					query,
					args...,
				).Scan(&fromUserCoins)
				require.NoError(t, err)

				query, args, err = s.builder.
					Select("coins").
					From("users").
					Where(squirrel.Eq{"username": transfer.ToUser}).
					ToSql()
				require.NoError(t, err)

				var toUserCoins int32
				err = testDbInstance.QueryRow(
					context.Background(),
					query,
					args...,
				).Scan(&toUserCoins)
				require.NoError(t, err)

				query, args, err = s.builder.
					Select("fromUser", "toUser", "coins").
					From("transactions").
					Where(squirrel.Eq{"fromUser": transfer.FromUser}).
					ToSql()
				require.NoError(t, err)

				transactionHistory := new(entity.TransferCoins)
				err = testDbInstance.QueryRow(
					context.Background(),
					query,
					args...,
				).Scan(
					&transactionHistory.FromUser,
					&transactionHistory.ToUser,
					&transactionHistory.Amount,
				)
				require.NoError(t, err)
				require.Equal(t, transfer, transactionHistory)

				if toUserCoins != UserCoinsOnRegister+transfer.Amount ||
					fromUserCoins != UserCoinsOnRegister-transfer.Amount {
					return fmt.Errorf("invalid amount of coins")
				}
				return nil
			},
			wantErr: false,
		}, // успешная отправка монет
		{
			name: "пользователю не хватает монет",
			transfer: &entity.TransferCoins{
				FromUser: "first",
				ToUser:   "second",
				Amount:   UserCoinsOnRegister + 1,
			},
			beforeTest: func(t *testing.T, transfer *entity.TransferCoins) {
				query, args, err := s.builder.
					Insert("users").
					Columns("username", "password").
					Values(transfer.FromUser, "hashedPass").
					Values(transfer.ToUser, "hashedPass").
					ToSql()
				require.NoError(t, err)

				_, err = testDbInstance.Exec(
					context.Background(),
					query,
					args...,
				)
				require.NoError(t, err)
			},
			check: func(t *testing.T, transfer *entity.TransferCoins) error {
				query, args, err := s.builder.
					Select("coins").
					From("users").
					Where(squirrel.Eq{"username": transfer.FromUser}).
					ToSql()
				require.NoError(t, err)

				var fromUserCoins int32
				err = testDbInstance.QueryRow(
					context.Background(),
					query,
					args...,
				).Scan(&fromUserCoins)
				require.NoError(t, err)

				query, args, err = s.builder.
					Select("coins").
					From("users").
					Where(squirrel.Eq{"username": transfer.ToUser}).
					ToSql()
				require.NoError(t, err)

				var toUserCoins int32
				err = testDbInstance.QueryRow(
					context.Background(),
					query,
					args...,
				).Scan(&toUserCoins)
				require.NoError(t, err)

				if fromUserCoins != UserCoinsOnRegister ||
					toUserCoins != UserCoinsOnRegister {
					return fmt.Errorf("user dont have needed coins, amount of coins was changed")
				}
				return nil
			},
			wantErr:     true,
			requiredErr: errs.NotEnoughCoins,
		}, // пользователю не хватает монет
		{
			name: "отправитель не найден",
			transfer: &entity.TransferCoins{
				FromUser: "first",
				ToUser:   "second",
				Amount:   UserCoinsOnRegister,
			},
			beforeTest: func(t *testing.T, transfer *entity.TransferCoins) {
				query, args, err := s.builder.
					Insert("users").
					Columns("username", "password").
					Values(transfer.ToUser, "hashedPass").
					ToSql()
				require.NoError(t, err)

				_, err = testDbInstance.Exec(
					context.Background(),
					query,
					args...,
				)
				require.NoError(t, err)
			},
			check: func(t *testing.T, transfer *entity.TransferCoins) error {
				query, args, err := s.builder.
					Select("coins").
					From("users").
					Where(squirrel.Eq{"username": transfer.ToUser}).
					ToSql()
				require.NoError(t, err)

				var toUserCoins int32
				err = testDbInstance.QueryRow(
					context.Background(),
					query,
					args...,
				).Scan(&toUserCoins)
				require.NoError(t, err)

				if toUserCoins != UserCoinsOnRegister {
					return fmt.Errorf("fromUser not found, but toUser coins was changed")
				}
				return nil
			},
			wantErr:     true,
			requiredErr: errs.UserNotFound,
		}, // отправитель не найден
		{
			name: "получатель не найден",
			transfer: &entity.TransferCoins{
				FromUser: "first",
				ToUser:   "second",
				Amount:   UserCoinsOnRegister,
			},
			beforeTest: func(t *testing.T, transfer *entity.TransferCoins) {
				query, args, err := s.builder.
					Insert("users").
					Columns("username", "password").
					Values(transfer.FromUser, "hashedPass").
					ToSql()
				require.NoError(t, err)

				_, err = testDbInstance.Exec(
					context.Background(),
					query,
					args...,
				)
				require.NoError(t, err)
			},
			check: func(t *testing.T, transfer *entity.TransferCoins) error {
				query, args, err := s.builder.
					Select("coins").
					From("users").
					Where(squirrel.Eq{"username": transfer.FromUser}).
					ToSql()
				require.NoError(t, err)

				var fromUserCoins int32
				err = testDbInstance.QueryRow(
					context.Background(),
					query,
					args...,
				).Scan(&fromUserCoins)
				require.NoError(t, err)

				if fromUserCoins != UserCoinsOnRegister {
					return fmt.Errorf("toUser not found, but fromUser coins was changed")
				}
				return nil
			},
			wantErr:     true,
			requiredErr: errs.UserNotFound,
		}, // получатель не найден
	}
	for _, tt := range testCases {
		s.T().Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				s.TearDownSubTest()
			})

			if tt.beforeTest != nil {
				tt.beforeTest(t, tt.transfer)
			}

			err := s.repo.SendCoins(context.Background(), tt.transfer)
			var checkErr error
			if tt.check != nil {
				checkErr = tt.check(t, tt.transfer)
			}

			if tt.wantErr {
				require.Equal(t, tt.requiredErr, err)
				require.NoError(t, checkErr)
			} else {
				require.NoError(t, err)
				require.NoError(t, checkErr)
			}
		})
	}
}

func TestIUserRepoTestSuite(t *testing.T) {
	suite.Run(t, new(IUserRepoSuite))
}
