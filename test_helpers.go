package main

import (
	"database/sql"
	"log"
	"log/slog"
	"strings"
	"testing"
)

type TestSetup struct {
	DB      *sql.DB
	Repo    *Repository
	Service *Service
}

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

func setupTest(t *testing.T) *TestSetup {
	db := initTestDB()
	repo := NewRepository(db)
	service := NewService(repo)

	return &TestSetup{
		DB:      db,
		Repo:    repo,
		Service: service,
	}
}

func teardownTest(t *testing.T, setup *TestSetup) {
	if setup.DB != nil {
		cleanDB(t, setup)
		setup.DB.Close()
	}
}

func cleanDB(t *testing.T, setup *TestSetup) {
	rows, err := setup.DB.Query("SELECT name FROM sqlite_master WHERE type = 'table' and name not in ('sqlite_sequence', 'schema_migrations')")
	if err != nil {
		t.Fatalf("Failed to query tables: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var table string

		if err := rows.Scan(&table); err != nil {
			t.Fatalf("Failed to scan table name: %v", err)
		}
		table = strings.TrimSpace(table)

		slog.Info("Cleaning table", "table", table)

		var count int
		err = setup.DB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?", table).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check if table %s exists: %v", table, err)
		}

		if count == 0 {
			slog.Info("Table does not exist, skipping", "table", table)
			continue
		}

		stmt, err := setup.DB.Prepare("DELETE FROM " + table)
		if err != nil {
			t.Fatalf("Failed to prepare delete statement for table %s: %v", table, err)
		}
		defer stmt.Close()

		if _, err := stmt.Exec(); err != nil {
			t.Fatalf("Failed to execute delete for table %s: %v", table, err)
		}
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("Error iterating over rows: %v", err)
	}
}
