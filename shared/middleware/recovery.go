package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/dealance/shared/pkg/response"
)

// Recovery returns a middleware that recovers from panics and returns a 500 JSON error.
func Recovery(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID := GetRequestID(c)
				log.Error().
					Str("request_id", requestID).
					Interface("panic", err).
					Str("path", c.Request.URL.Path).
					Msg("panic recovered")

				c.JSON(http.StatusInternalServerError, response.Response{
					Success: false,
					Error: &response.ErrorBody{
						Code:    "INTERNAL_ERROR",
						Message: "An unexpected error occurred",
					},
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
