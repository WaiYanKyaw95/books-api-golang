package main

import (
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func initDB() *sql.DB {
	// open the connection to the postgres
	db, err := sql.Open("pgx", "host=localhost port=5432 user=wai password=password dbname=books sslmode=disable")

	// handle error 500 database error
	if err != nil {
		log.Fatal(err)
	}

	// ping the database
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// create a table books if not exists
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS books (
		id				SERIAL PRIMARY KEY,
		title 			TEXT NOT NULL,
		author 			TEXT[],
		year			INTEGER,
		subject			TEXT[],
		description		TEXT,
		open_library_id TEXT UNIQUE
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// indexing
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_books_year ON books(year)`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_books_title ON books(title)`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_books_author ON books USING GIN(author)`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_books_subject ON books USING GIN(subject)`)
	if err != nil {
		log.Fatal(err)
	}

	// users table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id 			SERIAL PRIMARY KEY,
		username 	TEXT NOT NULL UNIQUE,
		password 	TEXT NOT NULL,
		role		TEXT NOT NULL DEFAULT 'user'
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// sessions table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS sessions (
		id			SERIAL PRIMARY KEY,
		user_id 	INTEGER NOT NULL,
		token		TEXT NOT NULL UNIQUE,
		created_at 	TEXT NOT NULL
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// return the db connection
	return db
}
