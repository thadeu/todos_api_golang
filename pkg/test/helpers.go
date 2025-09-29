package test

import (
	"database/sql"
	"log"
	"path/filepath"

	"github.com/Masterminds/squirrel"

	database "todoapp/internal/adapter/database/sqlite"
	"todoapp/pkg"
)

type TestSetup[T any] struct {
	DB   *database.DB
	Repo *T
}

func InitTestDB() *database.DB {
	db, err := sql.Open("sqlite3", ":memory:")

	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON")

	if err != nil {
		log.Fatal(err)
	}

	projectRoot := pkg.FindProjectRoot()
	migrationsPath := filepath.Join(projectRoot, "infra", "migrations")

	database.RunMigrations(db, migrationsPath)

	queryBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)

	return &database.DB{
		DB:           db,
		QueryBuilder: &queryBuilder,
	}
}
