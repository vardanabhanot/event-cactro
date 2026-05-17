package db

import (
	"log"
	"os"
)

// Migrate reads the SQL migration file and executes it against the database.
func Migrate() {
	sql, err := os.ReadFile("migrations/001_init.sql")
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	if _, err := DB.Exec(string(sql)); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Migrations applied successfully")
}
