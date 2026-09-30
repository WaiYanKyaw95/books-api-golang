package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
	Role     string `json:"role"`
}

type AuthInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Register(c *gin.Context, db *sql.DB) {
	// create a AuthInput struct
	var input AuthInput
	// get the username and password from the context - ShouldBindJSON()
	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	// hash byte from the plain password
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	// handle hash error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// hash password in string
	passwordHash := string(hashBytes)

	// insert them into users table
	var id int
	err = db.QueryRow(`INSERT INTO users (username, password)
		VALUES ($1, $2) RETURNING id`,
		input.Username,
		passwordHash,
	).Scan(&id)
	// PostgreSQL way of handling error
	if pqErr, ok := err.(*pgconn.PgError); ok {
		// unique username violation
		if pqErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
			return
		}
		// handle 500 database error
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// respond with a user - build yourself; don't retrieve.
	newUser := User{ID: int(id), Username: input.Username, Role: "user"}
	c.JSON(http.StatusCreated, newUser)
}
