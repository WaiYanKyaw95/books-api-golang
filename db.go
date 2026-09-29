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

	// return the db connection
	return db
}
