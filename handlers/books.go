package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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

func CreateBooks(c *gin.Context, db *sql.DB) {
	// create a Book struct
	var book Book
	// get book info from user
	err := c.ShouldBindJSON(&book)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	// validate the data before inserting into the database
	// such as title (required), author (required and at least one) and year (optional, valid year) and subject (optional)
	if book.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "must include a title"})
		return
	}

	if len(book.Author) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "must include at least one author"})
		return
	}

	if book.Year != 0 {
		if book.Year <= 1000 || book.Year > time.Now().Year() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
			return
		}
	}
	// insert into the database
	var id int
	err = db.QueryRow(`INSERT INTO books (title, author, year, subject) 
							VALUES ($1, $2, $3, $4) RETURNING id`,
		book.Title, pq.Array(book.Author), book.Year, pq.Array(book.Subject)).Scan(&id)
	// handle 500 database error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// respond with 201 book added
	book.ID = id
	c.JSON(http.StatusCreated, book)
}

func UpdateBook(c *gin.Context, db *sql.DB) {

	// get the param id -> id, convert it to int and handle conversion error
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// check if the record with that id exists - through database query
	var existing Book
	row := db.QueryRow(`SELECT id, title, author, year, subject FROM books WHERE id = $1`, idInt)
	err = row.Scan(&existing.ID, &existing.Title, pq.Array(&existing.Author), &existing.Year, pq.Array(&existing.Subject))
	// if not available, no book with such id
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "no such book"})
		return
		// database error 500
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// shouldjsonbind book, a Book struct
	var book Book
	err = c.ShouldBindJSON(&book)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// check upfront if all values are there
	if book.Title == "" && len(book.Author) == 0 && book.Year == 0 && len(book.Subject) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no field to update"})
		return
	}

	// check if year is valid
	if book.Year != 0 && (book.Year < 1000 || book.Year > time.Now().Year()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
		return
	}

	// start the dynamic query - title, author, year, subject
	query := "UPDATE books SET"
	args := []any{}
	argsCount := 1

	if book.Title != "" {
		query += fmt.Sprintf(" title = $%d,", argsCount)
		args = append(args, book.Title)
		argsCount++
	}

	if len(book.Author) > 0 {
		query += fmt.Sprintf(" author = $%d,", argsCount)
		args = append(args, pq.Array(book.Author))
		argsCount++
	}

	if book.Year != 0 {
		query += fmt.Sprintf(" year = $%d,", argsCount)
		args = append(args, book.Year)
		argsCount++
	}

	if len(book.Subject) > 0 {
		query += fmt.Sprintf(" subject = $%d,", argsCount)
		args = append(args, pq.Array(book.Subject))
		argsCount++
	}

	// remove the trailing comma before WHERE clause
	query = strings.TrimSuffix(query, ",")
	query += fmt.Sprintf(" WHERE id = $%d", argsCount)
	args = append(args, idInt)

	// syntax to update -> UPDATE books SET column = value, column = value WHERE column = value;
	_, err = db.Exec(query, args...)
	// handle 500 database error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// query the updated record
	// respond with 200 updated book
	var updated Book
	row = db.QueryRow(`SELECT id, title, author, year, subject FROM books WHERE id = $1`, idInt)
	err = row.Scan(&updated.ID, &updated.Title, pq.Array(&updated.Author), &updated.Year, pq.Array(&updated.Subject))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func DeleteBook(c *gin.Context, db *sql.DB) {
	// get the param id -> id, convert it to int and handle errors
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// delete the record
	result, err := db.Exec(`DELETE FROM books WHERE id = $1`, idInt)
	// handle 500 database error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	// if deleted successfully
	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	// 200 successfully deleted.
	c.JSON(http.StatusOK, gin.H{"message": "successfully deleted"})
}
