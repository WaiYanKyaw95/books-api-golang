package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

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
	if input.Username == "" || input.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
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

func LogIn(c *gin.Context, db *sql.DB) {
	// create a AuthInput variable
	var input AuthInput
	var user User
	// bind json
	err := c.ShouldBindJSON(&input)
	// if not, invalid request 400
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// check if username exists - grab id and role as well for successful cases
	row := db.QueryRow(`SELECT id, username, password, role FROM users WHERE username = $1`, input.Username)
	err = row.Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	// if not, invalid credentials 401
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
		// 500 database error
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// check against hash password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	// if not, invalid credentials 401
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// successful login will receive token and username response
	// create a 32 bytes token and fill it with random bytes
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	// change it to a hex and store in the sessions table with user_id
	encodedTokenStr := hex.EncodeToString(tokenBytes)
	_, err = db.Exec(`INSERT INTO sessions (user_id, token, created_at) VALUES ($1, $2, $3)`, user.ID, string(encodedTokenStr), time.Now().Format(time.RFC3339))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	// respond includes token, username
	c.JSON(http.StatusOK, gin.H{"username": user.Username, "token": encodedTokenStr})
}

func LogOut(c *gin.Context, db *sql.DB) {
	// check if the user is logged in, using authorization token (c.GetHeader)
	authorization := c.GetHeader("Authorization")
	if !strings.Contains(authorization, "Bearer ") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token unavailable"})
		return
	}

	token := strings.TrimPrefix(authorization, "Bearer ")

	result, err := db.Exec(`DELETE FROM sessions WHERE token = $1`, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}
