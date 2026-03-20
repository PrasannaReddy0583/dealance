package http

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/dealance/services/chat/internal/transport/http/handler"
	"github.com/dealance/shared/middleware"
	dealjwt "github.com/dealance/shared/pkg/jwt"
)

type ServerConfig struct{ Port string }

func NewServer(cfg ServerConfig, rdb *redis.Client, jwtVerifier *dealjwt.Verifier, chatHandler *handler.ChatHandler, log zerolog.Logger) *gin.Engine {
	engine := gin.New()
	engine.Use(
		cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:*", "http://10.0.2.2:*", "http://127.0.0.1:*", "http://10.250.134.27:*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID", "X-Device-ID", "X-Timestamp", "X-Nonce", "X-Signature"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),
		middleware.RequestID(), middleware.Logger(log), middleware.Recovery(log), middleware.SecurityHeaders(),
	)

	engine.GET("/health", handler.HealthCheck)

	protected := engine.Group("/api/v1/chat")
	protected.Use(middleware.JWTAuth(jwtVerifier, rdb, log))
	{
		protected.POST("/conversations", chatHandler.CreateConversation)
		protected.GET("/conversations", chatHandler.GetConversations)
		protected.GET("/conversations/:id/messages", chatHandler.GetMessages)
		protected.POST("/conversations/:id/read", chatHandler.MarkRead)
		protected.POST("/messages", chatHandler.SendMessage)
		protected.PATCH("/messages/:id", chatHandler.EditMessage)
		protected.DELETE("/messages/:id", chatHandler.DeleteMessage)
	}

	return engine
}
