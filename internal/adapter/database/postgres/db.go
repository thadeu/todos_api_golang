package postgres

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"todoapp/pkg"

	"github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	*pgxpool.Pool
	QueryBuilder *squirrel.StatementBuilderType
	url          string
}

func NewDB() (*DB, error) {
	url := os.Getenv("DATABASE_URL")

	ctx := context.Background()

	if url == "" {
		return nil, errors.New("DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(ctx, url)

	if err != nil {
		return nil, err
	}

	err = pool.Ping(ctx)

	if err != nil {
		pool.Close()
		return nil, err
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	db := &DB{
		Pool:         pool,
		QueryBuilder: &psql,
		url:          url,
	}

	if err := RunMigrations(pool, url); err != nil {
		pool.Close()
		return nil, err
	}

	return db, nil
}

func RunMigrations(db *pgxpool.Pool, dbURL string) error {
	migrationsPath := os.Getenv("MIGRATIONS_PATH")

	if migrationsPath == "" {
		migrationsPath = "infra/migrations"
	}

	sqlDB, err := sql.Open("pgx", dbURL)

	if err != nil {
		return err
	}

	defer sqlDB.Close()

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})

	if err != nil {
		return err
	}

	projectRoot := pkg.FindProjectRoot()
	migrationsPath = filepath.Join(projectRoot, "infra", "migrations")

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
