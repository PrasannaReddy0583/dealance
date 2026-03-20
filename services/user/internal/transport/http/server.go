package http

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/dealance/services/user/internal/transport/http/handler"
	"github.com/dealance/shared/middleware"
	dealjwt "github.com/dealance/shared/pkg/jwt"
)

type ServerConfig struct {
	Port        string
	SkipAttest  bool
	SkipSigning bool
}

func NewServer(
	cfg ServerConfig,
	rdb *redis.Client,
	jwtVerifier *dealjwt.Verifier,
	profileHandler *handler.ProfileHandler,
	followHandler *handler.FollowHandler,
	settingsHandler *handler.SettingsHandler,
	log zerolog.Logger,
) *gin.Engine {
	engine := gin.New()

	// Global middleware chain
	engine.Use(
		cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:*", "http://10.0.2.2:*", "http://127.0.0.1:*", "http://10.250.134.27:*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID", "X-Device-ID", "X-Timestamp", "X-Nonce", "X-Signature"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),
		middleware.RequestID(),
		middleware.Logger(log),
		middleware.Recovery(log),
		middleware.SecurityHeaders(),
	)

	// Health
	engine.GET("/health", handler.HealthCheck)

	// API v1 group
	api := engine.Group("/api/v1")

	// Public routes
	{
		api.GET("/search", profileHandler.Search)
		api.GET("/profile/username/:username", profileHandler.GetByUsername)
		api.GET("/profile/:id", profileHandler.GetByID)
		api.GET("/followers/:id", followHandler.GetFollowers)
		api.GET("/following/:id", followHandler.GetFollowing)
		api.GET("/follow-counts/:id", followHandler.GetCounts)
	}

	// Protected routes (require JWT)
	protected := api.Group("")
	protected.Use(middleware.JWTAuth(jwtVerifier, rdb, log))
	{
		protected.POST("/profile", profileHandler.Create)
		protected.GET("/profile", profileHandler.GetMe)
		protected.PATCH("/profile", profileHandler.Update)

		// Follow
		protected.POST("/follow", followHandler.Follow)
		protected.POST("/unfollow", followHandler.Unfollow)

		// Block
		protected.POST("/block", followHandler.Block)
		protected.DELETE("/block/:id", followHandler.Unblock)

		// Settings
		protected.GET("/settings", settingsHandler.Get)
		protected.PATCH("/settings", settingsHandler.Update)
	}

	return engine
}
