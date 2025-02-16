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

type IItemRepoSuite struct {
	suite.Suite
	repo    entity.IItemRepository
	builder squirrel.StatementBuilderType
}

func (s *IItemRepoSuite) SetupSuite() {
	s.repo = postgres.NewItemRepository(testDbInstance)
	s.builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
}

func (s *IItemRepoSuite) TearDownSubTest() {
	query := `truncate table users cascade`
	_, err := testDbInstance.Exec(context.Background(), query)
	require.NoError(s.T(), err)
}

func (s *IItemRepoSuite) Test_itemRepository_BuyItem() {
	testCases := []struct {
		name        string
		purchase    *entity.Purchase
		beforeTest  func(t *testing.T)
		check       func(t *testing.T, purchase *entity.Purchase) error
		wantErr     bool
		requiredErr error
	}{
		{
			name: "успешная покупка",
			purchase: &entity.Purchase{
				Username: "user",
				ItemName: "cup",
			},
			beforeTest: func(t *testing.T) {
				query, args, err := s.builder.
					Insert("users").
					Columns("username", "password").
					Values("user", "hashedPass").
					ToSql()
				require.NoError(t, err)

				_, err = testDbInstance.Exec(
					context.Background(),
					query,
					args...,
				)
				require.NoError(t, err)
			},
			check: func(t *testing.T, purchase *entity.Purchase) error {
				query, args, err := s.builder.
					Select("item").
					From("purchases").
					Where(squirrel.Eq{"username": purchase.Username}).
					ToSql()
				require.NoError(t, err)

				var purchaseHistory string
				err = testDbInstance.QueryRow(
					context.Background(),
					query,
					args...,
				).Scan(&purchaseHistory)
				require.NoError(t, err)

				if purchase.ItemName != purchaseHistory {
					return fmt.Errorf("purchase item name %s does not match history %s", purchaseHistory, purchase.ItemName)
				}
				return nil
			},
			wantErr: false,
		}, // успешная покупка
		{
			name: "пользователю не хватает монет",
			purchase: &entity.Purchase{
				Username: "user",
				ItemName: "cup",
			},
			beforeTest: func(t *testing.T) {
				query, args, err := s.builder.
					Insert("users").
					Columns("username", "password", "coins").
					Values("user", "hashedPass", 0).
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
			requiredErr: errs.NotEnoughCoins,
		}, // пользователю не хватает монет
		{
			name: "попытка купить несуществующую вещь",
			purchase: &entity.Purchase{
				Username: "user",
				ItemName: "undefined",
			},
			beforeTest: func(t *testing.T) {
				query, args, err := s.builder.
					Insert("users").
					Columns("username", "password", "coins").
					Values("user", "hashedPass", 0).
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
			requiredErr: errs.ItemNotFound,
		}, // попытка купить несуществующую вещь
		{
			name: "пользователь не найден",
			purchase: &entity.Purchase{
				Username: "undefined",
				ItemName: "cup",
			},
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
				tt.beforeTest(s.T())
			}

			err := s.repo.BuyItem(context.Background(), tt.purchase)
			var checkErr error
			if tt.check != nil {
				checkErr = tt.check(s.T(), tt.purchase)
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

func (s *IItemRepoSuite) Test_itemRepository_GetInventory() {
	testCases := []struct {
		name        string
		username    string
		inventory   []*entity.Item
		beforeTest  func(t *testing.T, username string, inventory []*entity.Item)
		wantErr     bool
		requiredErr error
	}{
		{
			name:     "успешное получение инвентаря",
			username: "user",
			inventory: []*entity.Item{
				{
					Name:     "cup",
					Quantity: 2,
				},
				{
					Name:     "powerbank",
					Quantity: 1,
				},
			},
			beforeTest: func(t *testing.T, username string, inventory []*entity.Item) {
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

				tmpBuilder := s.builder.
					Insert("purchases").
					Columns("username", "item")
				for _, item := range inventory {
					for i := int32(0); i < item.Quantity; i++ {
						tmpBuilder = tmpBuilder.Values(username, item.Name)
					}
				}
				query, args, err = tmpBuilder.ToSql()
				require.NoError(t, err)

				_, err = testDbInstance.Exec(
					context.Background(),
					query,
					args...,
				)
				require.NoError(t, err)
			},
			wantErr: false,
		}, // успешное получение инвентаря
		{
			name:      "успешное получение пустого инвентаря",
			username:  "user",
			inventory: []*entity.Item{},
			wantErr:   false,
		}, // успешное получение пустого инвентаря
	}
	for _, tt := range testCases {
		s.T().Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				s.TearDownSubTest()
			})

			if tt.beforeTest != nil {
				tt.beforeTest(t, tt.username, tt.inventory)
			}

			inventory, err := s.repo.GetInventory(context.Background(), tt.username)

			if tt.wantErr {
				require.Equal(t, tt.requiredErr, err)
				require.Nil(t, inventory)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.inventory, inventory)
			}
		})
	}
}

func TestIItemRepoTestSuite(t *testing.T) {
	suite.Run(t, new(IItemRepoSuite))
}
