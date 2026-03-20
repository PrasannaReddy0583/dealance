package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"
const RequestIDKey = "request_id"

// RequestID generates a unique request ID and sets it in the context and response header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set(RequestIDKey, requestID)
		c.Header(RequestIDHeader, requestID)
		c.Next()
	}
}

// GetRequestID retrieves the request ID from the gin context.
func GetRequestID(c *gin.Context) string {
	if id, exists := c.Get(RequestIDKey); exists {
		return id.(string)
	}
	return ""
}
