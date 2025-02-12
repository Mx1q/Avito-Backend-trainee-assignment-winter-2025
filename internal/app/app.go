package app

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/entity"
	"Avito-Backend-trainee-assignment-winter-2025/internal/pkg/config"
	"Avito-Backend-trainee-assignment-winter-2025/internal/pkg/jwt"
	"Avito-Backend-trainee-assignment-winter-2025/internal/pkg/logger"
	"Avito-Backend-trainee-assignment-winter-2025/internal/service"
	"Avito-Backend-trainee-assignment-winter-2025/internal/storage/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Config      *config.Config
	Logger      logger.ILogger
	AuthService entity.IAuthService
	ItemService entity.IItemService
	UserService entity.IUserService
}

func NewApp(db *pgxpool.Pool, cfg *config.Config, logger logger.ILogger) *App {
	authRepo := postgres.NewAuthRepository(db)
	itemRepo := postgres.NewItemRepository(db)
	userRepo := postgres.NewUserRepository(db)

	return &App{
		Config: cfg,
		Logger: logger,
		AuthService: service.NewAuthService(
			authRepo,
			logger,
			jwt.NewHashCrypto(),
			cfg.Jwt.Key,
		),
		ItemService: service.NewItemService(
			itemRepo,
			logger,
		),
		UserService: service.NewUserService(
			userRepo,
			logger,
		),
	}
}
