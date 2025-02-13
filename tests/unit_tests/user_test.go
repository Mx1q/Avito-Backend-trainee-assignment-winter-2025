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

func TestUserService_GetCoinsHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocks.NewMockLogger()
	repo := mocks.NewMockIUserRepository(ctrl)

	svc := service.NewUserService(repo, logger)

	tests := []struct {
		name        string
		username    string
		beforeTest  func(authRepo mocks.MockIUserRepository)
		wantErr     bool
		requiredErr error
	}{
		{
			name:     "успешное получение истории",
			username: "user",
			beforeTest: func(authRepo mocks.MockIUserRepository) {
				repo.EXPECT().
					GetCoinsHistory(context.Background(), "user").
					Return(
						int32(0),
						&entity.CoinsHistory{
							Received: make([]*entity.User, 0),
							Sent:     make([]*entity.User, 0),
						},
						nil,
					)
			},
			wantErr: false,
		}, // успешное получение истории
		{
			name:        "пустое имя пользователя",
			username:    "",
			wantErr:     true,
			requiredErr: errs.InvalidData,
		}, // пустое имя пользователя
		{
			name:     "repo get coins history error",
			username: "user",
			beforeTest: func(authRepo mocks.MockIUserRepository) {
				repo.EXPECT().
					GetCoinsHistory(context.Background(), "user").
					Return(
						int32(0),
						nil,
						fmt.Errorf("repo error"),
					)
			}, // repo get coins history error
			wantErr:     true,
			requiredErr: errs.InternalError,
		}, // repo get coins history error
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeTest != nil {
				tt.beforeTest(*repo)
			}

			coins, history, err := svc.GetCoinsHistory(context.Background(), tt.username)

			if tt.wantErr {
				require.Equal(t, tt.requiredErr, err)
				require.Nil(t, history)
				require.Zero(t, coins)
			} else {
				require.NotEmpty(t, history)
				require.GreaterOrEqual(t, coins, int32(0))
				require.Nil(t, err)
			}
		})
	}
}

func TestUserService_SendCoins(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocks.NewMockLogger()
	repo := mocks.NewMockIUserRepository(ctrl)

	svc := service.NewUserService(repo, logger)

	tests := []struct {
		name        string
		transfer    *entity.TransferCoins
		beforeTest  func(authRepo mocks.MockIUserRepository)
		wantErr     bool
		requiredErr error
	}{
		{
			name: "успешная отправка монет",
			transfer: &entity.TransferCoins{
				FromUser: "1",
				ToUser:   "2",
				Amount:   100,
			},
			beforeTest: func(authRepo mocks.MockIUserRepository) {
				repo.EXPECT().
					SendCoins(context.Background(), &entity.TransferCoins{
						FromUser: "1",
						ToUser:   "2",
						Amount:   100,
					}).
					Return(nil)
			},
			wantErr: false,
		}, // успешная отправка монет
		{
			name: "repo send coins error",
			transfer: &entity.TransferCoins{
				FromUser: "1",
				ToUser:   "2",
				Amount:   100,
			},
			beforeTest: func(authRepo mocks.MockIUserRepository) {
				repo.EXPECT().
					SendCoins(context.Background(), &entity.TransferCoins{
						FromUser: "1",
						ToUser:   "2",
						Amount:   100,
					}).
					Return(fmt.Errorf("repo error"))
			},
			wantErr:     true,
			requiredErr: errs.InternalError,
		}, // repo send coins error
		{
			name: "у пользователя недостаточно монет",
			transfer: &entity.TransferCoins{
				FromUser: "1",
				ToUser:   "2",
				Amount:   100,
			},
			beforeTest: func(authRepo mocks.MockIUserRepository) {
				repo.EXPECT().
					SendCoins(context.Background(), &entity.TransferCoins{
						FromUser: "1",
						ToUser:   "2",
						Amount:   100,
					}).
					Return(errs.NotEnoughCoins)
			},
			wantErr:     true,
			requiredErr: errs.NotEnoughCoins,
		}, // у пользователя недостаточно монет
		{
			name: "пользователь-получатель не найден",
			transfer: &entity.TransferCoins{
				FromUser: "1",
				ToUser:   "2",
				Amount:   100,
			},
			beforeTest: func(authRepo mocks.MockIUserRepository) {
				repo.EXPECT().
					SendCoins(context.Background(), &entity.TransferCoins{
						FromUser: "1",
						ToUser:   "2",
						Amount:   100,
					}).
					Return(errs.UserNotFound)
			},
			wantErr:     true,
			requiredErr: errs.UserNotFound,
		}, // у пользователя недостаточно монет
		{
			name: "пустое имя отправителя",
			transfer: &entity.TransferCoins{
				FromUser: "",
				ToUser:   "2",
				Amount:   100,
			},
			wantErr:     true,
			requiredErr: errs.InvalidData,
		}, // пустое имя отправителя
		{
			name: "пустое имя получателя",
			transfer: &entity.TransferCoins{
				FromUser: "1",
				ToUser:   "",
				Amount:   100,
			},
			wantErr:     true,
			requiredErr: errs.InvalidData,
		}, // пустое имя получателя
		{
			name: "отправка 0 монет",
			transfer: &entity.TransferCoins{
				FromUser: "1",
				ToUser:   "2",
				Amount:   0,
			},
			wantErr:     true,
			requiredErr: errs.InvalidData,
		}, // отправка 0 монет
		{
			name: "отправка отрицательного количества монет",
			transfer: &entity.TransferCoins{
				FromUser: "1",
				ToUser:   "2",
				Amount:   -10,
			},
			wantErr:     true,
			requiredErr: errs.InvalidData,
		}, // отправка отрицательного количества монет
		{
			name: "получатель и отправитель одинаковые",
			transfer: &entity.TransferCoins{
				FromUser: "1",
				ToUser:   "1",
				Amount:   100,
			},
			wantErr:     true,
			requiredErr: errs.InvalidData,
		}, // получатель и отправитель одинаковые
		{
			name:        "nil",
			transfer:    nil,
			wantErr:     true,
			requiredErr: errs.InvalidData,
		}, // nil
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeTest != nil {
				tt.beforeTest(*repo)
			}

			err := svc.SendCoins(context.Background(), tt.transfer)

			if tt.wantErr {
				require.Equal(t, tt.requiredErr, err)
			} else {
				require.Nil(t, err)
			}
		})
	}
}
