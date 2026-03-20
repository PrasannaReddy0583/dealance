package http

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/dealance/services/deal/internal/transport/http/handler"
	"github.com/dealance/shared/middleware"
	dealjwt "github.com/dealance/shared/pkg/jwt"
)

type ServerConfig struct{ Port string }

func NewServer(cfg ServerConfig, rdb *redis.Client, jwtVerifier *dealjwt.Verifier, dealHandler *handler.DealHandler, log zerolog.Logger) *gin.Engine {
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

	// Public
	api := engine.Group("/api/v1/deals")
	{
		api.GET("/:id", dealHandler.Get)
		api.GET("/startup/:startup_id", dealHandler.GetByStartup)
		api.GET("/:id/participants", dealHandler.GetParticipants)
		api.GET("/:id/documents", dealHandler.GetDocuments)
		api.GET("/:id/milestones", dealHandler.GetMilestones)
		api.GET("/:id/negotiations", dealHandler.GetNegotiations)
	}

	// Protected
	protected := engine.Group("/api/v1/deals")
	protected.Use(middleware.JWTAuth(jwtVerifier, rdb, log))
	{
		protected.POST("/", dealHandler.Create)
		protected.PATCH("/:id", dealHandler.Update)
		protected.GET("/mine", dealHandler.GetMyDeals)
		protected.POST("/:id/join", dealHandler.Join)
		protected.POST("/:id/commit", dealHandler.Commit)
		protected.POST("/:id/nda/sign", dealHandler.SignNDA)
		protected.POST("/:id/negotiate", dealHandler.SendMessage)
		protected.POST("/:id/documents", dealHandler.UploadDocument)
	}

	return engine
}
