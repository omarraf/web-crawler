package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func initDB(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err = db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Connected to database successfully!")
	return db, nil
}