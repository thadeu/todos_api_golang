package internal

import (
	"database/sql"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
	sqldblogger "github.com/simukti/sqldb-logger"

	"github.com/rs/zerolog"
	"github.com/simukti/sqldb-logger/logadapter/zerologadapter"
)

func InitDB() *sql.DB {
	dbPath := os.Getenv("DATABASE_PATH")

	if dbPath == "" {
		dbPath = "db/database.db"
	}

	db, err := sql.Open("sqlite3", dbPath)

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	logger := zerolog.New(os.Stdout)

	db = sqldblogger.OpenDriver(dbPath, db.Driver(), zerologadapter.New(logger))

	if err != nil {
		log.Fatal(err)
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")

	if migrationsPath == "" {
		migrationsPath = "db/migrations"
	}

	RunMigrations(db, migrationsPath)

	return db
}

func RunMigrations(db *sql.DB, migrationsPath string) {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})

	if err != nil {
		log.Fatal("Failed to create migration driver:", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
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

func GetServerPort() string {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	return port
}
