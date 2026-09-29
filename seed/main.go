package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

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
	queries := []string{
		"fiction",
		"mystery",
		"fantasy",
		"horror",
		"thriller",
		"romance",
		"biography",
		"science",
		"history",
		"adventure",
		"poetry",
		"philosophy",
		"dystopia",
		"classic",
		"crime",
		"drama",
		"humor",
		"travel",
		"war",
		"children",
	}

	// connect to the db
	db, err := sql.Open("pgx", "postgres://wai:password@localhost:5432/books?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	var totalCount int64 = 0

	// loop through the queries
	// fetch from open library
	for _, q := range queries {
		url := fmt.Sprintf("https://openlibrary.org/search.json?q=%s&limit=100&fields=title,author_name,first_publish_year,subject,key", q)
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Fatal(err)
		}

		// parse the JSON response into struct
		var result SearchResponse

		json.Unmarshal(body, &result)

		// insert each book
		var count int64 = 0
		for _, doc := range result.Docs {
			// make all subjects lower case
			for i, s := range doc.Subject {
				doc.Subject[i] = strings.ToLower(s)
			}

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

		// update totalCount
		totalCount += count
	}
	log.Printf("Total inserted %d books", totalCount)

}
