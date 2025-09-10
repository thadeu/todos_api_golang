package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

func initDB() *sql.DB {
	dbPath := os.Getenv("DATABASE_PATH")

	if dbPath == "" {
		dbPath = "db/database.db"
	}

	db, err := sql.Open("sqlite3", dbPath)

	if err != nil {
		log.Fatal(err)
	}

	// Run migrations
	runMigrations(db)

	return db
}

// runMigrations executes database migrations
func runMigrations(db *sql.DB) {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		log.Fatal("Failed to create migration driver:", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"sqlite3",
		driver,
	)
	if err != nil {
		log.Fatal("Failed to create migration instance:", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("Failed to run migrations:", err)
	}
}

// initTestDB creates an in-memory SQLite database for testing
func initTestDB() *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")

	if err != nil {
		log.Fatal(err)
	}

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")

	if err != nil {
		log.Fatal(err)
	}

	// Run migrations for test database
	runMigrations(db)

	return db
}
