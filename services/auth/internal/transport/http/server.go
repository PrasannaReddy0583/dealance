package http

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/dealance/services/auth/internal/transport/http/handler"
	"github.com/dealance/shared/middleware"
	dealjwt "github.com/dealance/shared/pkg/jwt"
)

// ServerConfig holds server configuration.
type ServerConfig struct {
	Port        string
	SkipAttest  bool
	SkipSigning bool
}

// NewServer creates the Gin engine with all routes and middleware.
func NewServer(
	cfg ServerConfig,
	rdb *redis.Client,
	jwtVerifier *dealjwt.Verifier,
	signupHandler *handler.SignupHandler,
	loginHandler *handler.LoginHandler,
	tokenHandler *handler.TokenHandler,
	passkeyHandler *handler.PasskeyHandler,
	kycWebhookHandler *handler.KYCWebhookHandler,
	log zerolog.Logger,
) *gin.Engine {
	// Use gin.New(), NOT gin.Default() — per spec
	r := gin.New()

	// CORS — allow Flutter app requests from any local origin
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:*", "http://10.0.2.2:*", "http://127.0.0.1:*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Device-ID", "X-Timestamp", "X-Nonce", "X-Signature", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Global middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recovery(log))
	r.Use(middleware.SecurityHeaders())

	// Health check (no auth required)
	r.GET("/health", handler.HealthCheck)

	// --- Public auth routes (no JWT required) ---
	api := r.Group("/api/v1")
	{
		// Rate limited signup routes
		signup := api.Group("/signup")
		signup.Use(middleware.AuthRateLimit(rdb))
		{
			signup.POST("/initiate", signupHandler.Initiate)
			signup.POST("/verify-email", signupHandler.VerifyEmail)
			signup.POST("/resend-otp", signupHandler.ResendOTP)
			signup.POST("/confirm-auth", signupHandler.ConfirmAuth)
			signup.POST("/set-country", signupHandler.SetCountry)
			signup.POST("/set-role", signupHandler.SetRole)
			signup.POST("/kyc/initiate", signupHandler.InitiateKYC)
			signup.GET("/status", signupHandler.Status)
		}

		// Login routes
		login := api.Group("/login")
		login.Use(middleware.AuthRateLimit(rdb))
		{
			login.POST("/passkey/begin", loginHandler.BeginPasskeyLogin)
			login.POST("/passkey/finish", loginHandler.FinishPasskeyLogin)
			login.POST("/oauth", loginHandler.OAuthLogin)
			login.POST("/email/begin", loginHandler.BeginEmailLogin)
			login.POST("/email/finish", loginHandler.FinishEmailLogin)
		}

		// Token management (refresh doesn't need JWT)
		api.POST("/token/refresh", tokenHandler.Refresh)

		// Webhook endpoints (verified by HMAC, not JWT)
		webhooks := api.Group("/webhooks")
		{
			webhooks.POST("/kyc/hyperverge", kycWebhookHandler.Hyperverge)
			webhooks.POST("/kyc/onfido", kycWebhookHandler.Onfido)
		}
	}

	// --- Protected routes (JWT required) ---
	protected := r.Group("/api/v1")
	protected.Use(middleware.JWTAuth(jwtVerifier, rdb, log))
	{
		protected.POST("/logout", tokenHandler.Logout)

		// Passkey registration on existing account
		passkeys := protected.Group("/passkeys")
		{
			passkeys.POST("/begin", passkeyHandler.BeginRegistration)
			passkeys.POST("/finish", passkeyHandler.FinishRegistration)
		}
	}

	return r
}
