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

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "welcome"})
	})

	// public routes - rate limit 100
	public := router.Group("/")
	public.Use(middleware.RateLimitMiddleWare(100, "public"))
	{
		public.GET("/books", func(c *gin.Context) {
			handlers.GetBooks(c, db)
		})

		public.GET("/books/:id", func(c *gin.Context) {
			handlers.GetBookByID(c, db)
		})
	}

	// auth routes - rate limit 10
	auth := router.Group("/")
	auth.Use(middleware.RateLimitMiddleWare(10, "auth"))
	{
		auth.POST("/register", func(c *gin.Context) {
			handlers.Register(c, db)
		})

		auth.POST("/login", func(c *gin.Context) {
			handlers.LogIn(c, db)
		})
	}

	// logged in users - rate limit 10
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleWare(db))
	protected.Use(middleware.RateLimitMiddleWare(10, "protected"))
	{
		protected.POST("/logout", func(c *gin.Context) {
			handlers.LogOut(c, db)
		})
	}

	// admin users - rate limit 30
	admin := router.Group("/")
	admin.Use(middleware.AuthMiddleWare(db))
	admin.Use(middleware.AdminMiddleWare())
	admin.Use(middleware.RateLimitMiddleWare(30, "admin"))
	{
		admin.POST("/books", func(c *gin.Context) {
			handlers.CreateBooks(c, db)
		})

		admin.PUT("/books/:id", func(c *gin.Context) {
			handlers.UpdateBook(c, db)
		})

		admin.DELETE("/books/:id", func(c *gin.Context) {
			handlers.DeleteBook(c, db)
		})
	}

	router.Run(":8080")
}
