package interceptor

import (
	"time"
	"log"
	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		log.Printf("[GIN] %s | %s | %v", c.Request.Method, c.Request.URL.Path, latency)
	}
}