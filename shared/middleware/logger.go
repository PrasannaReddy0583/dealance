package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Logger returns a gin middleware that logs each request with zerolog.
func Logger(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		requestID := GetRequestID(c)

		if raw != "" {
			path = path + "?" + raw
		}

		event := log.Info()
		if status >= 500 {
			event = log.Error()
		} else if status >= 400 {
			event = log.Warn()
		}

		event.
			Str("request_id", requestID).
			Str("method", method).
			Str("path", path).
			Int("status", status).
			Dur("latency", latency).
			Str("client_ip", clientIP).
			Int("body_size", c.Writer.Size()).
			Msg("request")
	}
}
