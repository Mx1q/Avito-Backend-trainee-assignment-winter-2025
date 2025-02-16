package main

import (
	appPackage "Avito-Backend-trainee-assignment-winter-2025/internal/app"
	"Avito-Backend-trainee-assignment-winter-2025/internal/pkg/config"
	loggerPackage "Avito-Backend-trainee-assignment-winter-2025/internal/pkg/logger"
	"Avito-Backend-trainee-assignment-winter-2025/internal/storage/postgres"
	"Avito-Backend-trainee-assignment-winter-2025/internal/web/handlers"
	"Avito-Backend-trainee-assignment-winter-2025/internal/web/middlewares"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

const (
	GracefulShutdownSeconds = 30
)

func main() {
	configPath := os.Getenv("AVITO_SHOP_CONFIG_PATH")
	cfg, err := config.ReadConfig(configPath)
	if err != nil {
		log.Fatalf("reading config error: %v\n", err)
	}

	logFile, err := os.OpenFile(cfg.Logger.File, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModeAppend)
	if err != nil {
		log.Fatalf("Opening log file error: %v\n", err)
	}
	defer func(logFile *os.File) {
		err := logFile.Close()
		if err != nil {
			log.Fatalf("Closing log file error: %v\n", err)
		}
	}(logFile)
	svcLogger := loggerPackage.NewLogger(cfg.Logger.Level, logFile)

	pool, err := postgres.NewConn(context.Background(), &cfg.Database)
	if err != nil {
		log.Fatalf("Connecting to database error: %v\n", err)
	}

	app := appPackage.NewApp(pool, cfg, svcLogger)

	r := fiber.New(fiber.Config{
		Prefork:       true,
		ServerHeader:  "Avito-shop",
		CaseSensitive: true,
	})
	r.Use(logger.New())
	r.Use(cors.New())

	r.Route("/api", func(r fiber.Router) {
		r.Post("/auth", handlers.FAuthHandler(app))

		r.Use(middlewares.JwtMiddleware(cfg.Jwt.Key))
		r.Get("/buy/:item", handlers.FBuyItemHandler(app))

		r.Post("/sendCoin", handlers.FSendCoinsHandler(app))
		r.Get("/info", handlers.FGetUserInfoHandler(app))
	})

	go func() {
		if err := r.Listen(fmt.Sprintf(":%d", cfg.HTTP.Port)); err != nil {
			log.Panic(err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-sig
	log.Info(syscall.Getpid(), " gracefully shutting down...")
	err = r.ShutdownWithTimeout(GracefulShutdownSeconds * time.Second)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Info(syscall.Getpid(), " successful graceful shutdown!")
	}
}
