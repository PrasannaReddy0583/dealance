package http

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/dealance/services/content/internal/transport/http/handler"
	"github.com/dealance/shared/middleware"
	dealjwt "github.com/dealance/shared/pkg/jwt"
)

type ServerConfig struct {
	Port string
}

func NewServer(
	cfg ServerConfig,
	rdb *redis.Client,
	jwtVerifier *dealjwt.Verifier,
	postHandler *handler.PostHandler,
	commentHandler *handler.CommentHandler,
	reactionHandler *handler.ReactionHandler,
	log zerolog.Logger,
) *gin.Engine {
	engine := gin.New()

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

	engine.GET("/health", handler.HealthCheck)

	// Public routes
	public := engine.Group("/api/v1/content")
	{
		public.GET("/posts/:id", postHandler.Get)
		public.GET("/posts/user/:user_id", postHandler.GetUserPosts)
		public.GET("/feed", postHandler.GetFeed)
		public.GET("/hashtag/:tag", postHandler.GetByHashtag)
		public.GET("/trending", postHandler.TrendingHashtags)
		public.GET("/posts/:id/comments", commentHandler.GetByPost)
		public.GET("/comments/:id/replies", commentHandler.GetReplies)
	}

	// Protected routes
	protected := engine.Group("/api/v1/content")
	protected.Use(middleware.JWTAuth(jwtVerifier, rdb, log))
	{
		protected.POST("/posts", postHandler.Create)
		protected.PATCH("/posts/:id", postHandler.Update)
		protected.DELETE("/posts/:id", postHandler.Delete)
		protected.POST("/posts/save", postHandler.Save)
		protected.DELETE("/posts/:id/save", postHandler.Unsave)

		protected.POST("/comments", commentHandler.Create)
		protected.PATCH("/comments/:id", commentHandler.Update)
		protected.DELETE("/comments/:id", commentHandler.Delete)

		protected.POST("/react", reactionHandler.React)
		protected.POST("/unreact", reactionHandler.Unreact)
		protected.POST("/report", reactionHandler.Report)
	}

	return engine
}
