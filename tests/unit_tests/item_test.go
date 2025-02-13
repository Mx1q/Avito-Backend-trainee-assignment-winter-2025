package unit_tests

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/entity"
	"Avito-Backend-trainee-assignment-winter-2025/internal/mocks"
	errs "Avito-Backend-trainee-assignment-winter-2025/internal/pkg/errors"
	"Avito-Backend-trainee-assignment-winter-2025/internal/service"
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestItemService_BuyItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocks.NewMockLogger()
	itemRepo := mocks.NewMockIItemRepository(ctrl)

	svc := service.NewItemService(itemRepo, logger)

	tests := []struct {
		name        string
		purchase    *entity.Purchase
		beforeTest  func(authRepo mocks.MockIItemRepository)
		wantErr     bool
		requiredErr error
	}{
		{
			name: "успешная покупка",
			purchase: &entity.Purchase{
				Username: "user",
				ItemName: "cup",
			},
			beforeTest: func(itemRepo mocks.MockIItemRepository) {
				itemRepo.EXPECT().
					BuyItem(context.Background(),
						&entity.Purchase{
							Username: "user",
							ItemName: "cup",
						}).
					Return(nil)
			},
			wantErr: false,
		}, // успешная покупка
		{
			name: "пользователь не найден",
			purchase: &entity.Purchase{
				Username: "user",
				ItemName: "cup",
			},
			beforeTest: func(itemRepo mocks.MockIItemRepository) {
				itemRepo.EXPECT().
					BuyItem(context.Background(),
						&entity.Purchase{
							Username: "user",
							ItemName: "cup",
						}).
					Return(errs.UserNotFound)
			},
			wantErr:     true,
			requiredErr: errs.UserNotFound,
		}, // пользователь не найден
		{
			name: "предмет не найден",
			purchase: &entity.Purchase{
				Username: "user",
				ItemName: "cup",
			},
			beforeTest: func(itemRepo mocks.MockIItemRepository) {
				itemRepo.EXPECT().
					BuyItem(context.Background(),
						&entity.Purchase{
							Username: "user",
							ItemName: "cup",
						}).
					Return(errs.ItemNotFound)
			},
			wantErr:     true,
			requiredErr: errs.ItemNotFound,
		}, // предмет не найден
		{
			name: "пользователю не хватает монет",
			purchase: &entity.Purchase{
				Username: "user",
				ItemName: "cup",
			},
			beforeTest: func(itemRepo mocks.MockIItemRepository) {
				itemRepo.EXPECT().
					BuyItem(context.Background(),
						&entity.Purchase{
							Username: "user",
							ItemName: "cup",
						}).
					Return(errs.NotEnoughCoins)
			},
			wantErr:     true,
			requiredErr: errs.NotEnoughCoins,
		}, // пользователю не хватает монет
		{
			name: "repo buy item error",
			purchase: &entity.Purchase{
				Username: "user",
				ItemName: "cup",
			},
			beforeTest: func(itemRepo mocks.MockIItemRepository) {
				itemRepo.EXPECT().
					BuyItem(context.Background(),
						&entity.Purchase{
							Username: "user",
							ItemName: "cup",
						}).
					Return(fmt.Errorf("repo buy item error"))
			},
			wantErr:     true,
			requiredErr: errs.InternalError,
		}, // repo buy item error
		{
			name: "пустое имя пользователя",
			purchase: &entity.Purchase{
				Username: "",
				ItemName: "cup",
			},
			wantErr:     true,
			requiredErr: errs.InvalidData,
		}, // пустое имя пользователя
		{
			name: "пустое название предмета",
			purchase: &entity.Purchase{
				Username: "user",
				ItemName: "",
			},
			wantErr:     true,
			requiredErr: errs.InvalidData,
		}, // пустое название предмета
		{
			name:        "nil",
			purchase:    nil,
			wantErr:     true,
			requiredErr: errs.InvalidData,
		}, // nil
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeTest != nil {
				tt.beforeTest(*itemRepo)
			}

			err := svc.BuyItem(context.Background(), tt.purchase)

			if tt.wantErr {
				require.Equal(t, tt.requiredErr, err)
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestItemService_GetInventory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocks.NewMockLogger()
	itemRepo := mocks.NewMockIItemRepository(ctrl)

	svc := service.NewItemService(itemRepo, logger)

	tests := []struct {
		name        string
		username    string
		beforeTest  func(authRepo mocks.MockIItemRepository)
		wantErr     bool
		requiredErr error
	}{
		{
			name:     "успешное получение инвентаря",
			username: "user",
			beforeTest: func(itemRepo mocks.MockIItemRepository) {
				itemRepo.EXPECT().
					GetInventory(context.Background(), "user").
					Return([]*entity.Item{
						{
							Name:     "cup",
							Quantity: 1,
						},
					}, nil)
			},
			wantErr: false,
		}, // успешное получение инвентаря
		{
			name:     "repo get inventory error",
			username: "user",
			beforeTest: func(itemRepo mocks.MockIItemRepository) {
				itemRepo.EXPECT().
					GetInventory(context.Background(), "user").
					Return(nil, fmt.Errorf("repo get inventory error"))
			},
			wantErr:     true,
			requiredErr: errs.InternalError,
		}, // repo get inventory error
		{
			name:        "repo get inventory error",
			username:    "",
			wantErr:     true,
			requiredErr: errs.InvalidData,
		}, // пустое имя пользователя
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeTest != nil {
				tt.beforeTest(*itemRepo)
			}

			inventory, err := svc.GetInventory(context.Background(), tt.username)

			if tt.wantErr {
				require.Equal(t, tt.requiredErr, err)
				require.Nil(t, inventory)
			} else {
				require.Nil(t, err)
				require.NotNil(t, inventory)
			}
		})
	}
}
