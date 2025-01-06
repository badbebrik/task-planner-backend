package migration

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func RunMigrations(db *sql.DB, migrationPath string) error {
	if err := ensureMigrationTable(db); err != nil {
		return err
	}

	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(migrationPath)
	if err != nil {
		return fmt.Errorf("Failed to read migration directory: %w", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".sql" {
			continue
		}

		version := file.Name()
		if _, applied := appliedMigrations[version]; applied {
			continue
		}

		log.Printf("Applying migration: %s", version)
		if err := applyMigration(db, filepath.Join(migrationPath, version), version); err != nil {
			return err
		}
	}

	log.Println("All migration applied successfully")
	return nil

}

func ensureMigrationTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT NOW()
	)`
	_, err := db.Exec(query)
	return err
}

func getAppliedMigrations(db *sql.DB) (map[string]struct{}, error) {
	query := `SELECT version FROM schema_migrations`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	migrations := make(map[string]struct{})
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		migrations[version] = struct{}{}
	}
	return migrations, nil
}

func applyMigration(db *sql.DB, filepath, version string) error {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("Failed to read migration file %s: %w", filepath, err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("Failed to start transaction %w", err)
	}

	if _, err := tx.Exec(string(content)); err != nil {
		tx.Rollback()
		return fmt.Errorf("Failed to execute migration %s: %w", version, err)
	}

	query := `INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2)`
	if _, err := tx.Exec(query, version, time.Now()); err != nil {
		tx.Rollback()
		return fmt.Errorf("Failed to record migration %s: %w", version, err)
	}

	return tx.Commit()
}
