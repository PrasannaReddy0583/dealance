package http

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/dealance/services/wallet/internal/transport/http/handler"
	"github.com/dealance/shared/middleware"
	dealjwt "github.com/dealance/shared/pkg/jwt"
)

type ServerConfig struct{ Port string }

func NewServer(cfg ServerConfig, rdb *redis.Client, jwtVerifier *dealjwt.Verifier, walletHandler *handler.WalletHandler, log zerolog.Logger) *gin.Engine {
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

	// ALL wallet routes are protected
	protected := engine.Group("/api/v1/wallet")
	protected.Use(middleware.JWTAuth(jwtVerifier, rdb, log))
	{
		protected.GET("/", walletHandler.GetWallet)
		protected.GET("/balance", walletHandler.GetBalance)
		protected.POST("/deposit", walletHandler.Deposit)
		protected.POST("/withdraw", walletHandler.Withdraw)
		protected.POST("/transfer", walletHandler.Transfer)
		protected.GET("/transactions", walletHandler.GetTransactions)
		protected.GET("/ledger", walletHandler.GetLedger)

		// Bank accounts
		protected.POST("/bank-accounts", walletHandler.AddBankAccount)
		protected.GET("/bank-accounts", walletHandler.GetBankAccounts)
		protected.DELETE("/bank-accounts/:id", walletHandler.RemoveBankAccount)
	}

	return engine
}
