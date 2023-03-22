package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func CorsMiddleware(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "false")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Accept-Encoding, auth-token, Cache-control, Connection, Content-Length, Content-Type, Origin, X-CSRF-Token, X-Requested-with")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PATCH, PUT, DELETE")
	c.Writer.Header().Set("Allow-credentials", "false")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	c.Next()
}
