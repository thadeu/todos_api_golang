package test

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	. "todoapp/pkg/db"
)

type TestSetup[T any] struct {
	DB   *sql.DB
	Repo *T
}

// findProjectRoot finds the project root directory by looking for go.mod
func findProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Fallback to current working directory
	if wd, err := os.Getwd(); err == nil {
		return wd
	}

	log.Fatal("Could not find project root directory")
	return ""
}

func InitTestDB() *sql.DB {
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
	projectRoot := findProjectRoot()
	migrationsPath := filepath.Join(projectRoot, "db", "migrations")
	RunMigrations(db, migrationsPath)

	return db
}

func SetupTest[T any](t *testing.T, repo *T) *TestSetup[T] {
	db := InitTestDB()

	return &TestSetup[T]{
		DB:   db,
		Repo: repo,
	}
}

func TeardownTest[T any](t *testing.T, setup *TestSetup[T]) {
	if setup.DB != nil {
		CleanDB(t, setup)
		setup.DB.Close()
	}
}

func CleanDB[T any](t *testing.T, setup *TestSetup[T]) {
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

		// slog.Info("Cleaning table", "table", table)

		var count int
		err = setup.DB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?", table).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check if table %s exists: %v", table, err)
		}

		if count == 0 {
			// slog.Info("Table does not exist, skipping", "table", table)
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
