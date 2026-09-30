package middleware

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

func AuthMiddleWare(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get authentication token
		authorization := c.GetHeader("Authorization")
		if !strings.Contains(authorization, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no valid token"})
			return
		}
		token := strings.TrimPrefix(authorization, "Bearer ")

		// check if it exists in the sessions database, grab its user_id
		var user User
		row := db.QueryRow(`SELECT user_id FROM sessions WHERE token = $1`, token)
		err := row.Scan(&user.ID)
		// if not, invalid request 401, abort and return
		if err == sql.ErrNoRows {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no valid session"})
			return
		}

		// get the username and role from users database, using user_id from above
		row = db.QueryRow(`SELECT username, role FROM users WHERE id = $1`, user.ID)
		err = row.Scan(&user.Username, &user.Role)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		// c.Set("user", user)
		c.Set("user", user)
		// c.Next() to continue with the main handler function
		c.Next()
	}
}

func AdminMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		// get the username and role from the context
		val, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		user, ok := val.(User)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		// check if the role is admin
		// if not, unauthorized 401, abort and return
		if user.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// c.Next()
		c.Next()
	}
}
