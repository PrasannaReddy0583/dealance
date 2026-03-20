package http

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/dealance/services/media/internal/transport/http/handler"
	"github.com/dealance/shared/middleware"
	dealjwt "github.com/dealance/shared/pkg/jwt"
)

type ServerConfig struct{ Port string }

func NewServer(cfg ServerConfig, rdb *redis.Client, jwtV *dealjwt.Verifier, h *handler.MediaHandler, log zerolog.Logger) *gin.Engine {
	e := gin.New()
	e.Use(
		cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:*", "http://10.0.2.2:*", "http://127.0.0.1:*", "http://10.250.134.27:*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID", "X-Device-ID", "X-Timestamp", "X-Nonce", "X-Signature"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),
		middleware.RequestID(), middleware.Logger(log), middleware.Recovery(log), middleware.SecurityHeaders(),
	)
	e.GET("/health", handler.HealthCheck)

	p := e.Group("/api/v1/media")
	p.Use(middleware.JWTAuth(jwtV, rdb, log))
	{
		p.POST("/upload-url", h.RequestUploadURL)
		p.POST("/confirm", h.ConfirmUpload)
		p.GET("/uploads", h.GetMyUploads)
		p.DELETE("/uploads/:id", h.DeleteUpload)
	}
	return e
}
