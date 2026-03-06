package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()

		c.Next()

		latency := time.Since(startedAt)
		log.Printf(
			"method=%s path=%s status=%d latency=%s client_ip=%s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			latency,
			c.ClientIP(),
		)
	}
}
