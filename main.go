package main

import (
	"bookapi/handlers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// set up dependencies
	db := initDB()
	defer db.Close()

	// create a default gin router with recovery and logger
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	router.GET("/books", func(c *gin.Context) {
		handlers.GetBooks(c, db)
	})

	router.GET("/books/:id", func(c *gin.Context) {
		handlers.GetBookByID(c, db)
	})

	router.Run(":8080")
}
