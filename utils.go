package main

import (
	"database/sql"
	"log"
	"os"

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

	return db
}
