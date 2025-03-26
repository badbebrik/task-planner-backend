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

	rateLimiter := auth.NewRateLimiter(5*time.Minute, 5)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Group(func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ip := r.RemoteAddr
				if !rateLimiter.Allow(ip) {
					http.Error(w, "Too many requests", http.StatusTooManyRequests)
					return
				}
				next.ServeHTTP(w, r)
			})
		})

		r.Post("/register/email", authHandler.RegisterEmail)
		r.Post("/register/email/verify", authHandler.VerifyEmail)
		r.Post("/login", authHandler.Login)
	})

	r.Group(func(r chi.Router) {
		r.Use(auth.JWTAuthMiddleware(cfg.JWTAccessSecret))
		r.Post("/refresh", authHandler.Refresh)
		r.Post("/logout", authHandler.Logout)
	})

	addr := fmt.Sprintf(":%d", cfg.AppPort)
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
