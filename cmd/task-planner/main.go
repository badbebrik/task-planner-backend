package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"os"
	"task-planner/internal/auth"
	"task-planner/internal/db"
	"task-planner/internal/email"
	"task-planner/internal/user"
	"task-planner/migration"
	"task-planner/pkg/config"
	"time"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Unable to load config: %v", err)
	}
	database, err := db.Connect(cfg.DB)
	if err != nil {
		log.Fatalf("Could not connect to DB: %v", err)
	}
	defer database.Close()

	if err := migration.RunMigrations(database, "migration"); err != nil {
		log.Fatalf("Failed to run migration: %v", err)
	}

	userRepo := user.NewPGRepository(database)
	userService := user.NewService(userRepo)

	emailService := email.NewSMTPEmailService(
		"smtp.gmail.com",
		"587",
		os.Getenv("EMAIL_USERNAME"),
		os.Getenv("EMAIL_PASSWORD"),
		"no-reply@whatamitodo.com",
	)
	emailRepo := email.NewEmailRepository(database)
	tokenRepo := auth.NewTokenRepository(database)

	jwtCfg := auth.JWTConfig{
		AccessSecret:  cfg.JWTAccessSecret,
		RefreshSecret: cfg.JWTRefreshSecret,
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    24 * time.Hour * 7,
	}

	authService := auth.NewService(userService, emailService, emailRepo, tokenRepo, jwtCfg)
	authHandler := auth.NewHandler(authService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("App is running"))
	})

	r.Post("/register/email", authHandler.RegisterEmail)
	r.Post("/register/email/verify", authHandler.VerifyEmail)

	r.Post("/login", authHandler.Login)
	r.Post("/refresh", authHandler.Refresh)

	addr := fmt.Sprintf(":%d", cfg.AppPort)
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
