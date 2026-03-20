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

	"github.com/dealance/services/user/config"
	"github.com/dealance/services/user/internal/application"
	pgRepo "github.com/dealance/services/user/internal/infrastructure/postgres"
	redisRepo "github.com/dealance/services/user/internal/infrastructure/redis"
	httpTransport "github.com/dealance/services/user/internal/transport/http"
	"github.com/dealance/services/user/internal/transport/http/handler"
	dealjwt "github.com/dealance/shared/pkg/jwt"
	"github.com/dealance/shared/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.App.Env)
	log = logger.WithService(log, "dealance-user")

	log.Info().
		Str("env", cfg.App.Env).
		Str("port", cfg.App.Port).
		Msg("starting user service")

	// PostgreSQL
	db, err := sqlx.Connect("pgx", cfg.DB.DSN())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to PostgreSQL")
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	log.Info().Msg("connected to PostgreSQL")

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer rdb.Close()

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to Redis")
	}
	log.Info().Msg("connected to Redis")

	// JWT Verifier (user service only verifies, doesn't issue)
	jwtVerifier, err := dealjwt.NewVerifier(cfg.JWT.PublicKeyPath, cfg.JWT.Issuer, cfg.JWT.Audience)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize JWT verifier")
	}
	log.Info().Msg("JWT verifier initialized")

	// Repositories
	profileRepo := pgRepo.NewProfileRepo(db)
	followRepo := pgRepo.NewFollowRepo(db)
	blockRepo := pgRepo.NewBlockRepo(db)
	settingsRepo := pgRepo.NewSettingsRepo(db)
	cacheRepo := redisRepo.NewCacheRepo(rdb)

	// Application services
	profileSvc := application.NewProfileService(profileRepo, settingsRepo, cacheRepo, log)
	followSvc := application.NewFollowService(followRepo, blockRepo, profileRepo, cacheRepo, log)
	settingsSvc := application.NewSettingsService(settingsRepo, log)

	// Handlers
	profileHandler := handler.NewProfileHandler(profileSvc)
	followHandler := handler.NewFollowHandler(followSvc)
	settingsHandler := handler.NewSettingsHandler(settingsSvc)

	// Server
	serverCfg := httpTransport.ServerConfig{
		Port: cfg.App.Port,
	}

	engine := httpTransport.NewServer(
		serverCfg, rdb, jwtVerifier,
		profileHandler, followHandler, settingsHandler, log,
	)

	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Info().Str("port", cfg.App.Port).Msg("user service listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down user service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server forced shutdown")
	}

	log.Info().Msg("user service stopped")
}
