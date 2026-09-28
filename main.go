package main

import (
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

	router.Run(":8080")
}
