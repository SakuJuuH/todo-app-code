package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CorsMiddleware(c *gin.Context) {
	allowedOrigins := getEnvWithDefault("ALLOWED_ORIGINS", "*")

	c.Header("Access-Control-Allow-Origin", allowedOrigins)
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.Next()
}
