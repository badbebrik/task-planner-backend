package main

import (
	"log"
	"net/http"
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
	service := user.NewService(userRepo)

	err := service.CreateUser("example@example.com", "hashed_password", "John Doe")
	if err != nil {
		log.Fatalf("Error creating user: %v", err)
	}

	log.Println("User created successfully")

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
