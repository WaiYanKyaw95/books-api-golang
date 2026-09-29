package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lib/pq"
)

// search info struct
type SearchResponse struct {
	Docs []BookDoc `json:"docs"`
}

// book attribute struct
type BookDoc struct {
	Title            string   `json:"title"`
	AuthorName       []string `json:"author_name"`
	FirstPublishYear int      `json:"first_publish_year"`
	Subject          []string `json:"subject"`
	Key              string   `json:"key"`
}

func main() {
	// connect to the db
	db, err := sql.Open("pgx", "postgres://wai:password@localhost:5432/books?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// fetch from open library
	resp, err := http.Get("https://openlibrary.org/search.json?q=fiction&limit=100&fields=title,author_name,first_publish_year,subject,key")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// parse the JSON response into struct
	var result SearchResponse

	json.Unmarshal(body, &result)

	// insert each book
	var count int64 = 0
	for _, doc := range result.Docs {
		result, err := db.Exec(`INSERT INTO books (title, author, year, subject, open_library_id)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (open_library_id) DO NOTHING`,
			doc.Title,
			pq.Array(doc.AuthorName),
			doc.FirstPublishYear,
			pq.Array(doc.Subject),
			doc.Key)
		if err != nil {
			log.Fatal(err)
		}
		rows, _ := result.RowsAffected()
		count += rows
	}

	// print count
	log.Printf("Inserted %d books", count)

}
