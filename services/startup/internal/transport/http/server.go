package http

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/dealance/services/startup/internal/transport/http/handler"
	"github.com/dealance/shared/middleware"
	dealjwt "github.com/dealance/shared/pkg/jwt"
)

type ServerConfig struct{ Port string }

func NewServer(cfg ServerConfig, rdb *redis.Client, jwtVerifier *dealjwt.Verifier, startupHandler *handler.StartupHandler, log zerolog.Logger) *gin.Engine {
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
	public := engine.Group("/api/v1/startups")
	{
		public.GET("/search", startupHandler.Search)
		public.GET("/slug/:slug", startupHandler.GetBySlug)
		public.GET("/:id", startupHandler.Get)
		public.GET("/:id/funding", startupHandler.GetFundingRounds)
		public.GET("/:id/team", startupHandler.GetTeam)
	}

	// Protected
	protected := engine.Group("/api/v1/startups")
	protected.Use(middleware.JWTAuth(jwtVerifier, rdb, log))
	{
		protected.POST("/", startupHandler.Create)
		protected.PATCH("/:id", startupHandler.Update)
		protected.GET("/mine", startupHandler.GetMyStartups)
		protected.POST("/:id/follow", startupHandler.Follow)
		protected.DELETE("/:id/follow", startupHandler.Unfollow)
		protected.POST("/:id/funding", startupHandler.CreateFundingRound)
		protected.POST("/:id/team", startupHandler.AddTeamMember)
	}

	return engine
}
