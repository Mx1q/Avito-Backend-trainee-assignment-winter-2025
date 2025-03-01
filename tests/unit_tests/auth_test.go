package unit_tests

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/entity"
	"Avito-Backend-trainee-assignment-winter-2025/internal/mocks"
	errs "Avito-Backend-trainee-assignment-winter-2025/internal/pkg/errors"
	"Avito-Backend-trainee-assignment-winter-2025/internal/service"
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestAuthService_Auth(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := mocks.NewMockLogger()
	repo := mocks.NewMockIAuthRepository(ctrl)
	hasher := mocks.NewMockIHashCrypto(ctrl)
	tokenManager := mocks.NewMockITokenManager(ctrl)

	svc := service.NewAuthService(repo, logger, hasher, tokenManager)

	tests := []struct {
		name        string
		authInfo    *entity.Auth
		beforeTest  func(authRepo mocks.MockIAuthRepository, crypto mocks.MockIHashCrypto)
		wantErr     bool
		requiredErr error
	}{
		{
			name: "успешная аутентификация",
			authInfo: &entity.Auth{
				Username: "username",
				Password: "pass",
			},
			beforeTest: func(authRepo mocks.MockIAuthRepository, hasher mocks.MockIHashCrypto) {
				authRepo.EXPECT().
					GetByUsername(
						context.Background(),
						"username",
					).
					Return(&entity.Auth{
						Username: "username",
						Password: "hashedPass",
					}, nil)

				hasher.EXPECT().
					VerifyPassword("pass", "hashedPass").
					Return(true)

				tokenManager.EXPECT().
					CreateToken("username").
					Return("token", nil)
			},
			wantErr: false,
		}, // успешная аутентификация
		{
			name: "успешная регистрация",
			authInfo: &entity.Auth{
				Username: "new",
				Password: "pass",
			},
			beforeTest: func(authRepo mocks.MockIAuthRepository, hasher mocks.MockIHashCrypto) {
				const hashedPass = "hashedPass"

				authRepo.EXPECT().
					GetByUsername(
						context.Background(),
						"new",
					).
					Return(nil, nil)

				hasher.EXPECT().
					HashPassword("pass").
					Return(hashedPass, nil)

				authRepo.EXPECT().
					Register(
						context.Background(),
						&entity.Auth{
							Username: "new",
							Password: hashedPass,
						},
					).
					Return(nil)

				tokenManager.EXPECT().
					CreateToken("new").
					Return("token", nil)
			},
			wantErr: false,
		}, // успешная регистрация
		{
			name:        "nil",
			authInfo:    nil,
			wantErr:     true,
			requiredErr: errs.InvalidData,
		}, // nil
		{
			name: "пустое имя пользователя",
			authInfo: &entity.Auth{
				Username: "",
				Password: "pass",
			},
			wantErr:     true,
			requiredErr: errs.InvalidData,
		}, // пустое имя пользователя
		{
			name: "пустой пароль",
			authInfo: &entity.Auth{
				Username: "username",
				Password: "",
			},
			wantErr:     true,
			requiredErr: errs.InvalidData,
		}, // пустой пароль
		{
			name: "repo getByUsername internal error",
			authInfo: &entity.Auth{
				Username: "username",
				Password: "pass",
			},
			beforeTest: func(authRepo mocks.MockIAuthRepository, hasher mocks.MockIHashCrypto) {
				authRepo.EXPECT().
					GetByUsername(
						context.Background(),
						"username",
					).
					Return(nil, fmt.Errorf("db internal error"))
			},
			wantErr:     true,
			requiredErr: errs.InternalError,
		}, // repo getByUsername internal error
		{
			name: "некорректный пароль",
			authInfo: &entity.Auth{
				Username: "username",
				Password: "pass",
			},
			beforeTest: func(authRepo mocks.MockIAuthRepository, hasher mocks.MockIHashCrypto) {
				authRepo.EXPECT().
					GetByUsername(
						context.Background(),
						"username",
					).
					Return(&entity.Auth{
						Username: "username",
						Password: "hashedPass",
					}, nil)

				hasher.EXPECT().
					VerifyPassword("pass", "hashedPass").
					Return(false)
			},
			wantErr:     true,
			requiredErr: errs.InvalidCredentials,
		}, // некорректный пароль
		{
			name: "repo register internal error",
			authInfo: &entity.Auth{
				Username: "new",
				Password: "pass",
			},
			beforeTest: func(authRepo mocks.MockIAuthRepository, hasher mocks.MockIHashCrypto) {
				const hashedPass = "hashedPass"

				authRepo.EXPECT().
					GetByUsername(
						context.Background(),
						"new",
					).
					Return(nil, nil)

				hasher.EXPECT().
					HashPassword("pass").
					Return(hashedPass, nil)

				authRepo.EXPECT().
					Register(
						context.Background(),
						&entity.Auth{
							Username: "new",
							Password: hashedPass,
						},
					).
					Return(fmt.Errorf("db internal error"))
			},
			wantErr:     true,
			requiredErr: errs.InternalError,
		}, // repo register internal error
		{
			name: "ошибка хеширования пароля",
			authInfo: &entity.Auth{
				Username: "new",
				Password: "pass",
			},
			beforeTest: func(authRepo mocks.MockIAuthRepository, hasher mocks.MockIHashCrypto) {
				authRepo.EXPECT().
					GetByUsername(
						context.Background(),
						"new",
					).
					Return(nil, nil)

				hasher.EXPECT().
					HashPassword("pass").
					Return("", fmt.Errorf("hashing error"))
			},
			wantErr:     true,
			requiredErr: errs.InternalError,
		}, // ошибка хеширования пароля
		{
			name: "ошибка получения токена",
			authInfo: &entity.Auth{
				Username: "username",
				Password: "pass",
			},
			beforeTest: func(authRepo mocks.MockIAuthRepository, hasher mocks.MockIHashCrypto) {
				authRepo.EXPECT().
					GetByUsername(
						context.Background(),
						"username",
					).
					Return(&entity.Auth{
						Username: "username",
						Password: "hashedPass",
					}, nil)

				hasher.EXPECT().
					VerifyPassword("pass", "hashedPass").
					Return(true)

				tokenManager.EXPECT().
					CreateToken("username").
					Return("", fmt.Errorf("creating token error"))
			},
			wantErr:     true,
			requiredErr: errs.InternalError,
		}, // успешная аутентификация
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeTest != nil {
				tt.beforeTest(*repo, *hasher)
			}

			_, err := svc.Auth(context.Background(), tt.authInfo)

			if tt.wantErr {
				require.Equal(t, tt.requiredErr, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
