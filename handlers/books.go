package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type Book struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Author      []string `json:"author"`
	Year        int      `json:"year"`
	Subject     []string `json:"subject"`
	Description string   `json:"description"`
}

func GetBooks(c *gin.Context, db *sql.DB) {
	// get query parameters - page and limit
	// default to page 1 and limit 10 - use c.DefaultQuery()
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	// query parameters
	author := c.Query("author")
	year := c.Query("year")
	subject := c.Query("subject")

	// sorting
	sort := c.DefaultQuery("sort", "id")    // default sort by id
	order := c.DefaultQuery("order", "asc") // default ascending

	// validate sort column to allow only known columns
	allowedSort := map[string]bool{
		"id": true, "title": true, "year": true,
	}

	if !allowedSort[sort] {
		sort = "id" // fallback to safe default
	}

	if order != "asc" && order != "desc" {
		order = "asc"
	}

	// c.Query() returns a string; thus need to convert
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page"})
		return
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
		return
	}

	// offset calculation = (page - 1) * limit
	offset := (pageInt - 1) * limitInt

	// retrieve the records from the database with a dynamic query
	// SELECT id, title, author, year, subject FROM books WHERE 1=1
	// AND "author" = ANY(author)
	// AND year = yearInt
	// AND "subject" = ANY(subject)
	// OFFSET offset LIMIT limitInt
	var books = []Book{}

	query := "SELECT id, title, author, year, subject FROM books WHERE 1=1"
	args := []any{}
	argsCount := 1

	if author != "" {
		query += fmt.Sprintf(" AND $%d = ANY(author)", argsCount)
		args = append(args, author)
		argsCount++
	}

	var yearInt int
	if year != "" {
		yearInt, err = strconv.Atoi(year)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
			return
		}
		query += fmt.Sprintf(" AND year = $%d", argsCount)
		args = append(args, yearInt)
		argsCount++
	}

	if subject != "" {
		subject = strings.ToLower(subject)
		query += fmt.Sprintf(" AND $%d = ANY(subject)", argsCount)
		args = append(args, subject)
		argsCount++
	}

	// sorting
	query += fmt.Sprintf(" ORDER BY %s %s", sort, order)

	// offset and limit
	query += fmt.Sprintf(" OFFSET $%d LIMIT $%d", argsCount, argsCount+1)
	args = append(args, offset, limitInt)

	rows, err := db.Query(query, args...)

	// database error 500
	if err != nil {
		log.Printf("GetBooks: query failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer rows.Close()

	// scan them into a struct one bye one
	for rows.Next() {
		var book Book

		if err := rows.Scan(&book.ID, &book.Title, pq.Array(&book.Author), &book.Year, pq.Array(&book.Subject)); err != nil {
			log.Printf("GetBooks: scan failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "data parsing error"})
			return
		}

		books = append(books, book)
	}
	if err = rows.Err(); err != nil {
		log.Printf("GetBooks: rows error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to retrieve records"})
		return
	}

	c.JSON(http.StatusOK, books)

}

func GetBookByID(c *gin.Context, db *sql.DB) {
	var book Book

	// get the param id
	id := c.Param("id")

	// convert id to int
	idInt, err := strconv.Atoi(id)
	// handle invalid id
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// retrieve the record
	row := db.QueryRow(`SELECT id, title, author, year, subject FROM books WHERE id = $1`, idInt)
	err = row.Scan(&book.ID, &book.Title, &book.Author, &book.Year, &book.Subject)

	// handle no records and 500 database error
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "no book found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// respond with the book
	c.JSON(http.StatusOK, book)
}
