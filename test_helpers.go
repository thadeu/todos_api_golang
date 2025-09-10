package main

import (
	"database/sql"
	"testing"
)

type TestSetup struct {
	DB      *sql.DB
	Repo    *Repository
	Service *Service
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
	tables := []string{}

	rows, err := setup.DB.Query("SELECT name FROM sqlite_master WHERE type = 'table' and name not in ('sqlite_sequence', 'schema_migrations')")

	if err != nil {
		t.Fatal(err)
	}

	defer rows.Close()

	for rows.Next() {
		var table string
		rows.Scan(&table)
		tables = append(tables, table)
	}

	for _, table := range tables {
		setup.DB.Exec("DELETE FROM " + table)
	}
}
