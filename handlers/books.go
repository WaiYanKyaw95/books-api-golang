package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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

	// retrieve the records from the database
	var books = []Book{}
	rows, err := db.Query(`SELECT id, title, author, year, subject FROM books
		OFFSET $1 LIMIT $2`,
		offset,
		limitInt)

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
		if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Year, &book.Subject); err != nil {
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
