package e2e_tests

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/app"
	"Avito-Backend-trainee-assignment-winter-2025/internal/mocks"
	"Avito-Backend-trainee-assignment-winter-2025/internal/pkg/config"
	"Avito-Backend-trainee-assignment-winter-2025/internal/web/handlers"
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	logger := mocks.NewMockLogger()

	tokenAuth := jwtauth.New("HS256", []byte(cfg.Jwt.Key), nil)

	app := app.NewApp(db, cfg, logger)
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Use(middleware.Logger)

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth", handlers.AuthHandler(app))

		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(jwtauth.Authenticator(tokenAuth))

			r.Route("/buy", func(r chi.Router) {
				r.Get("/{item}", handlers.BuyItemHandler(app))
			})

			r.Post("/sendCoin", handlers.SendCoinsHandler(app))
			r.Get("/info", handlers.GetUserInfoHandler(app))
		})
	})
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler: r,
	}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		log.Println("shutting down server...")
		shutdownCtx, _ := context.WithTimeout(serverCtx, GracefulShutdownSeconds*time.Second)
		go func() {
			<-shutdownCtx.Done()
			if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	started <- true
	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

	<-serverCtx.Done()
}
