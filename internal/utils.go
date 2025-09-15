package internal

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"go.opentelemetry.io/otel"

	"github.com/rs/zerolog"
	"github.com/simukti/sqldb-logger/logadapter/zerologadapter"
)

func InitDB() *sql.DB {
	// Garantir que o TracerProvider esteja configurado
	tracerProvider := otel.GetTracerProvider()
	if tracerProvider == nil {
		log.Fatal("TracerProvider not configured. Initialize telemetry first.")
	}

	log.Printf("Initializing database with TracerProvider: %T", tracerProvider)
	dbPath := os.Getenv("DATABASE_PATH")

	if dbPath == "" {
		dbPath = "db/database.db"
	}

	// Primeiro, executar migrações com banco não instrumentado
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

	// Agora criar o banco instrumentado para a aplicação
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
