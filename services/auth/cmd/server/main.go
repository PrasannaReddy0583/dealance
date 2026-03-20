package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"

	"github.com/dealance/services/auth/config"
	"github.com/dealance/services/auth/internal/application"
	pgRepo "github.com/dealance/services/auth/internal/infrastructure/postgres"
	redisRepo "github.com/dealance/services/auth/internal/infrastructure/redis"
	scyllaRepo "github.com/dealance/services/auth/internal/infrastructure/scylla"
	httpTransport "github.com/dealance/services/auth/internal/transport/http"
	"github.com/dealance/services/auth/internal/transport/http/handler"
	dealjwt "github.com/dealance/shared/pkg/jwt"
	"github.com/dealance/shared/pkg/logger"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(cfg.App.Env)
	log = logger.WithService(log, "dealance-auth")

	log.Info().
		Str("env", cfg.App.Env).
		Str("port", cfg.App.Port).
		Bool("skip_attest", cfg.App.SkipAttest).
		Bool("skip_signing", cfg.App.SkipSigning).
		Bool("kyc_mock", cfg.App.KYCMock).
		Msg("starting auth service")

	// Initialize PostgreSQL
	db, err := sqlx.Connect("pgx", cfg.DB.DSN())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to PostgreSQL")
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Info().Msg("connected to PostgreSQL")

	// Initialize Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer rdb.Close()

	// Test Redis connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to Redis")
	}
	log.Info().Msg("connected to Redis")

	// Initialize JWT Issuer (auth service has the private key)
	jwtIssuer, err := dealjwt.NewIssuer(cfg.JWT.PrivateKeyPath, cfg.JWT.Issuer, cfg.JWT.Audience)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize JWT issuer")
	}
	log.Info().Msg("JWT issuer initialized")

	// JWT Verifier (uses public key from issuer)
	jwtVerifier := dealjwt.NewVerifierFromKey(jwtIssuer.GetPublicKey(), cfg.JWT.Issuer, cfg.JWT.Audience)

	// Initialize repositories
	userRepo := pgRepo.NewUserRepo(db)
	roleRepo := pgRepo.NewUserRoleRepo(db)
	identRepo := pgRepo.NewIdentityProviderRepo(db)
	kycRepo := pgRepo.NewKYCRepo(db)
	refreshRepo := pgRepo.NewRefreshTokenRepo(db)
	sessionRepo := redisRepo.NewSessionRepo(rdb)
	auditRepo := scyllaRepo.NewAuditLogRepo(log)

	// Initialize external services (mock for dev)
	emailSvc := &mockEmailService{}
	kycVendorSvc := &mockKYCVendorService{}

	// Initialize application services
	signupSvc := application.NewSignupService(
		userRepo, roleRepo, identRepo, kycRepo,
		sessionRepo, auditRepo, emailSvc, kycVendorSvc,
		log, cfg.App.KYCMock,
	)

	loginSvc := application.NewLoginService(
		userRepo, roleRepo, identRepo, refreshRepo,
		sessionRepo, auditRepo, emailSvc, jwtIssuer, log,
	)

	tokenSvc := application.NewTokenService(
		userRepo, roleRepo, refreshRepo,
		sessionRepo, auditRepo, jwtIssuer, jwtVerifier, log,
	)

	passkeySvc := application.NewPasskeyService(
		identRepo, sessionRepo, auditRepo, log,
	)

	kycWebhookSvc := application.NewKYCWebhookService(
		userRepo, kycRepo, auditRepo, log,
	)

	// Initialize handlers
	signupHandler := handler.NewSignupHandler(signupSvc)
	loginHandler := handler.NewLoginHandler(loginSvc)
	tokenHandler := handler.NewTokenHandler(tokenSvc)
	passkeyHandler := handler.NewPasskeyHandler(passkeySvc)
	kycWebhookHandler := handler.NewKYCWebhookHandler(kycWebhookSvc, cfg.KYC.HypervergeSecret, cfg.KYC.OnfidoWebhookSecret)

	// Create server
	serverCfg := httpTransport.ServerConfig{
		Port:        cfg.App.Port,
		SkipAttest:  cfg.App.SkipAttest,
		SkipSigning: cfg.App.SkipSigning,
	}

	engine := httpTransport.NewServer(
		serverCfg, rdb, jwtVerifier,
		signupHandler, loginHandler, tokenHandler,
		passkeyHandler, kycWebhookHandler, log,
	)

	// Start HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Info().Str("port", cfg.App.Port).Msg("auth service listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down auth service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server forced shutdown")
	}

	log.Info().Msg("auth service stopped")
}

// --- Mock services for development ---

type mockEmailService struct{}

func (s *mockEmailService) SendOTP(ctx context.Context, email, otp string) error {
	fmt.Printf("[MOCK EMAIL] OTP %s sent to %s\n", otp, email)
	return nil
}

func (s *mockEmailService) SendWelcome(ctx context.Context, email string) error {
	fmt.Printf("[MOCK EMAIL] Welcome email sent to %s\n", email)
	return nil
}

func (s *mockEmailService) SendKYCApproved(ctx context.Context, email string) error {
	fmt.Printf("[MOCK EMAIL] KYC approved email sent to %s\n", email)
	return nil
}

func (s *mockEmailService) SendKYCRejected(ctx context.Context, email, reason string) error {
	fmt.Printf("[MOCK EMAIL] KYC rejected email sent to %s: %s\n", email, reason)
	return nil
}

type mockKYCVendorService struct{}

func (s *mockKYCVendorService) InitiateSession(ctx context.Context, userID, kycType string) (string, string, error) {
	fmt.Printf("[MOCK KYC] Session initiated for user %s, type %s\n", userID, kycType)
	return "mock_session_id", "mock_sdk_token", nil
}

func (s *mockKYCVendorService) VerifyWebhookSignature(payload []byte, signature string) bool {
	return true // Always valid in dev
}
