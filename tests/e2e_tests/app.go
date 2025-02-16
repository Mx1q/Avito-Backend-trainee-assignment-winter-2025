package e2e_tests

import (
	appPackage "Avito-Backend-trainee-assignment-winter-2025/internal/app"
	"Avito-Backend-trainee-assignment-winter-2025/internal/mocks"
	"Avito-Backend-trainee-assignment-winter-2025/internal/pkg/config"
	"Avito-Backend-trainee-assignment-winter-2025/internal/web/handlers"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	GracefulShutdownSeconds = 30
	TestingPort             = 8081
)

func RunTheApp(db *pgxpool.Pool, started chan bool) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{Port: TestingPort},
		Jwt:  config.Jwt{Key: "abcdef12345"},
	}
	svcLogger := mocks.NewMockLogger()

	app := appPackage.NewApp(db, cfg, svcLogger)

	r := fiber.New(fiber.Config{
		Prefork:       false,
		ServerHeader:  "Avito-shop",
		CaseSensitive: true,
	})
	r.Use(logger.New())
	r.Use(cors.New())

	r.Route("/api", func(r fiber.Router) {
		r.Post("/auth", handlers.AuthHandler(app))

		r.Use(jwtMiddleware(cfg.Jwt.Key))
		r.Get("/buy/:item", handlers.BuyItemHandler(app))

		r.Post("/sendCoin", handlers.SendCoinsHandler(app))
		r.Get("/info", handlers.GetUserInfoHandler(app))
	})

	go func() {
		if err := r.Listen(fmt.Sprintf(":%d", cfg.HTTP.Port)); err != nil {
			log.Panic(err)
		}
	}()
	started <- true

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-sig
	log.Info(syscall.Getpid(), " gracefully shutting down...")
	err := r.ShutdownWithTimeout(GracefulShutdownSeconds * time.Second)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Info(syscall.Getpid(), " successful graceful shutdown!")
	}
}

func jwtMiddleware(jwtKey string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		return jwtware.New(jwtware.Config{
			SigningKey: jwtware.SigningKey{Key: []byte(jwtKey)},
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"errors": err.Error(),
				})
			},
		})(ctx)
	}
}

//func RunTheApp(db *pgxpool.Pool, started chan bool) {
//	cfg := &config.Config{
//		HTTP: config.HTTPConfig{Port: TestingPort},
//		Jwt:  config.Jwt{Key: "abcdef12345"},
//	}
//	logger := mocks.NewMockLogger()
//
//	tokenAuth := jwtauth.New("HS256", []byte(cfg.Jwt.Key), nil)
//
//	app := app.NewApp(db, cfg, logger)
//	r := chi.NewRouter()
//	r.Use(cors.Handler(cors.Options{
//		AllowedOrigins:   []string{"https://*", "http://*"},
//		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
//		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
//		ExposedHeaders:   []string{"Link"},
//		AllowCredentials: false,
//		MaxAge:           300,
//	}))
//
//	r.Use(middleware.Logger)
//
//	r.Route("/api", func(r chi.Router) {
//		r.Post("/auth", handlers.AuthHandler(app))
//
//		r.Group(func(r chi.Router) {
//			r.Use(jwtauth.Verifier(tokenAuth))
//			r.Use(jwtauth.Authenticator(tokenAuth))
//
//			r.Route("/buy", func(r chi.Router) {
//				r.Get("/{item}", handlers.BuyItemHandler(app))
//			})
//
//			r.Post("/sendCoin", handlers.SendCoinsHandler(app))
//			r.Get("/info", handlers.GetUserInfoHandler(app))
//		})
//	})
//	server := &http.Server{
//		Addr:    fmt.Sprintf(":%d", cfg.HTTP.Port),
//		Handler: r,
//	}
//
//	serverCtx, serverStopCtx := context.WithCancel(context.Background())
//	sig := make(chan os.Signal, 1)
//	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
//	go func() {
//		<-sig
//
//		log.Println("shutting down server...")
//		shutdownCtx, _ := context.WithTimeout(serverCtx, GracefulShutdownSeconds*time.Second)
//		go func() {
//			<-shutdownCtx.Done()
//			if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
//				log.Fatal("graceful shutdown timed out.. forcing exit.")
//			}
//		}()
//
//		err := server.Shutdown(shutdownCtx)
//		if err != nil {
//			log.Fatal(err)
//		}
//		serverStopCtx()
//	}()
//
//	started <- true
//	err := server.ListenAndServe()
//	if err != nil && !errors.Is(err, http.ErrServerClosed) {
//		log.Fatal(err)
//	}
//
//	<-serverCtx.Done()
//}
