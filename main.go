package main

import (
	"bookapi/handlers"
	"bookapi/middleware"
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

	router.POST("/register", func(c *gin.Context) {
		handlers.Register(c, db)
	})

	router.POST("/login", func(c *gin.Context) {
		handlers.LogIn(c, db)
	})

	// logged in users
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleWare(db))

	{
		protected.POST("/logout", func(c *gin.Context) {
			handlers.LogOut(c, db)
		})
	}

	// admin users
	admin := router.Group("/")
	admin.Use(middleware.AuthMiddleWare(db))
	admin.Use(middleware.AdminMiddleWare())

	{
		admin.POST("/books", func(c *gin.Context) {
			handlers.CreateBooks(c, db)
		})

		admin.PUT("/books/:id", func(c *gin.Context) {
			handlers.UpdateBook(c, db)
		})

		admin.DELETE("/books/:id", func(c *gin.Context) {
			// delete book handler
		})
	}

	router.Run(":8080")
}
