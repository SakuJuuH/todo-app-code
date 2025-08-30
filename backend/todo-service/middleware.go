package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func CorsMiddleware(c *gin.Context) {
	var allowedOrigins = os.Getenv("ALLOWED_ORIGINS")

	c.Header("Access-Control-Allow-Origin", allowedOrigins)
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.Next()
}
