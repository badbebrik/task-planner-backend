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
	"task-planner/internal/goal"
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

	userRepo := user.NewRepository(database)
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

	authService := auth.NewService(userService, emailService, emailRepo, tokenRepo, cfg.JWT)
	authHandler := auth.NewHandler(authService)

	rateLimiter := auth.NewRateLimiter(1*time.Minute, 60)

	goalRepo := goal.NewRepository(database)
	goalService := goal.NewService(goalRepo, os.Getenv("OPENAI_API_KEY"))
	goalHandler := goal.NewHandler(goalService)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(auth.RateLimiterMiddleware(rateLimiter))

	r.Group(func(r chi.Router) {
		r.Route("/api/auth", func(r chi.Router) {
			r.Post("/signup", authHandler.Signup)
			r.Post("/verify-email", authHandler.VerifyEmail)
			r.Post("/send-code", authHandler.SendVerificationCode)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
			r.Post("/logout", authHandler.Logout)
		})
	})

	r.Route("/api/users", func(r chi.Router) {
		r.With(auth.JWTAuthMiddleware(cfg.JWT.AccessSecret)).
			Get("/me", authHandler.GetMe)
	})

	r.Group(func(r chi.Router) {
		r.Use(auth.JWTAuthMiddleware(cfg.JWT.RefreshSecret))

		r.Route("/goals", func(r chi.Router) {
			r.Post("/", goalHandler.CreateGoal)
			r.Get("/", goalHandler.ListGoals)
			r.Get("/{id}", goalHandler.GetGoal)
			r.Put("/{id}", goalHandler.UpdateGoal)
			r.Post("/decompose", goalHandler.CreateGoalDecomposed)

			r.Post("/phases", goalHandler.CreatePhase)

			r.Post("/tasks", goalHandler.CreateTask)
			r.Put("/tasks/{id}", goalHandler.UpdateTask)
		})
	})

	addr := fmt.Sprintf(":%d", cfg.AppPort)
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
