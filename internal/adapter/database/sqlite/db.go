package sqlite

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/Masterminds/squirrel"

	_ "github.com/mattn/go-sqlite3"
	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"go.opentelemetry.io/otel"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/rs/zerolog"
	"github.com/simukti/sqldb-logger/logadapter/zerologadapter"
)

type DB struct {
	*sql.DB
	QueryBuilder *squirrel.StatementBuilderType
}

func New() *sql.DB {
	tracerProvider := otel.GetTracerProvider()

	if tracerProvider == nil {
		log.Fatal("TracerProvider not configured. Initialize telemetry first.")
	}

	log.Printf("Initializing database with TracerProvider: %T", tracerProvider)
	dbPath := os.Getenv("DATABASE_PATH")

	if dbPath == "" {
		dbPath = "database.db"
	}

	migrationDB, err := sql.Open("sqlite3", dbPath)

	if err != nil {
		log.Fatal(err)
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")

	if migrationsPath == "" {
		migrationsPath = "db/migrations"
	}

	RunMigrations(migrationDB, migrationsPath)
	migrationDB.Close()

	sqlDB, err := otelsql.Open("sqlite3", dbPath,
		otelsql.WithDBSystem("sqlite"),
		otelsql.WithDBName("todoapp"),
		otelsql.WithTracerProvider(tracerProvider),
	)

	if err != nil {
		log.Fatal(err)
	}

	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	logger := zerolog.New(os.Stdout)

	db := sqldblogger.OpenDriver(dbPath, sqlDB.Driver(), zerologadapter.New(logger))

	return db
}

func NewDB() (*DB, error) {
	sqlDB := New()
	queryBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	return &DB{
		DB:           sqlDB,
		QueryBuilder: &queryBuilder,
	}, nil
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
