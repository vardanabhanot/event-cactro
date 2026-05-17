package db

import (
	"database/sql"
	"log"

	"eventbooking/config"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Connect() {
	var err error
	DB, err = sql.Open("sqlite", config.App.DBPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Enable WAL mode for better concurrent read/write performance
	if _, err = DB.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		log.Fatalf("Failed to set WAL mode: %v", err)
	}

	// Enforce foreign key constraints
	if _, err = DB.Exec("PRAGMA foreign_keys=ON;"); err != nil {
		log.Fatalf("Failed to enable foreign keys: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Database connected:", config.App.DBPath)
}
