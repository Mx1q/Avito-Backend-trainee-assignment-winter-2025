package main

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/app"
	"Avito-Backend-trainee-assignment-winter-2025/internal/pkg/config"
	loggerPackage "Avito-Backend-trainee-assignment-winter-2025/internal/pkg/logger"
	"Avito-Backend-trainee-assignment-winter-2025/internal/storage/postgres"
	"Avito-Backend-trainee-assignment-winter-2025/internal/web"
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"log"
	"net/http"
	"os"
)

func main() {
	cfg, err := config.ReadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	logFile, err := os.OpenFile(cfg.Logger.File, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModeAppend)
	if err != nil {
		log.Fatalln(err)
	}
	defer func(logFile *os.File) {
		err := logFile.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(logFile)
	logger := loggerPackage.NewLogger(cfg.Logger.Level, logFile)

	//tokenAuth := jwtauth.New("HS256", []byte(cfg.Jwt.Key), nil)

	logger.Infof("connecting to database...")
	pool, err := postgres.NewConn(context.Background(), &cfg.Database)
	if err != nil {
		log.Fatalln(err)
	}
	logger.Infof("successfully connected to database!")

	app := app.NewApp(pool, cfg, logger)
	mux := chi.NewRouter()
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	mux.Use(middleware.Logger)

	mux.Route("/api", func(r chi.Router) {
		r.Post("/auth", web.AuthHandler(app))
	})

	http.ListenAndServe(":8080", mux)
}
