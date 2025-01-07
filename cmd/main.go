package main

import (
	"log"
	"net/http"
	"os"
	"task-planner/internal/auth"
	"task-planner/internal/email"
	"task-planner/internal/user"
	"task-planner/migration"
	"task-planner/pkg/config"
)

func main() {
	cfg := config.LoadConfig()
	db := config.ConnectDB(cfg)
	defer db.Close()

	if err := migration.RunMigrations(db, "migration"); err != nil {
		log.Fatalf("Failed to run migration: %v", err)
	}

	userRepo := user.NewPGRepository(db)
	userService := user.NewService(userRepo)

	emailService := email.NewSMTPEmailService(
		"smtp.gmail.com",
		"587",
		os.Getenv("EMAIL_USERNAME"),
		os.Getenv("EMAIL_PASSWORD"),
		"no-reply@whatamitodo.com",
	)
	emailRepo := email.NewEmailRepository(db)

	authService := auth.NewService(userService, emailService, emailRepo)
	authHandler := auth.NewHandler(authService)

	http.HandleFunc("/register/email", authHandler.RegisterEmail)
	http.HandleFunc("/register/email/verify", authHandler.VerifyEmail)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("App is running"))
		if err != nil {
			return
		}
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
