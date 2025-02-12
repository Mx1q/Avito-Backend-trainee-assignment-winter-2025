package main

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/app"
	"Avito-Backend-trainee-assignment-winter-2025/internal/pkg/config"
	loggerPackage "Avito-Backend-trainee-assignment-winter-2025/internal/pkg/logger"
	"Avito-Backend-trainee-assignment-winter-2025/internal/storage/postgres"
	"Avito-Backend-trainee-assignment-winter-2025/internal/web"
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	GracefulShutdownSeconds = 30
)

func main() {
	log.Println("Reading config...")
	cfg, err := config.ReadConfig()
	if err != nil {
		log.Fatalf("reading config error: %v\n", err)
	}

	log.Println("Opening log file...")
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
	logger := loggerPackage.NewLogger(cfg.Logger.Level, logFile)

	tokenAuth := jwtauth.New("HS256", []byte(cfg.Jwt.Key), nil)

	log.Println("Connecting to database...")
	pool, err := postgres.NewConn(context.Background(), &cfg.Database)
	if err != nil {
		log.Fatalf("Connecting to database error: %v\n", err)
	}
	log.Println("Successfully connected to database!")

	app := app.NewApp(pool, cfg, logger)
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
		r.Post("/auth", web.AuthHandler(app))

		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(jwtauth.Authenticator(tokenAuth))

			r.Route("/buy", func(r chi.Router) {
				r.Get("/{item}", web.BuyItemHandler(app))
			})

			r.Post("/sendCoin", web.SendCoinsHandler(app))
			r.Get("/info", web.GetUserInfoHandler(app))
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

	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

	<-serverCtx.Done()
}
