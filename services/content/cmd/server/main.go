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

	"github.com/dealance/services/content/config"
	"github.com/dealance/services/content/internal/application"
	pgRepo "github.com/dealance/services/content/internal/infrastructure/postgres"
	redisRepo "github.com/dealance/services/content/internal/infrastructure/redis"
	httpTransport "github.com/dealance/services/content/internal/transport/http"
	"github.com/dealance/services/content/internal/transport/http/handler"
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
	log = logger.WithService(log, "dealance-content")

	log.Info().
		Str("env", cfg.App.Env).
		Str("port", cfg.App.Port).
		Msg("starting content service")

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
		Addr: cfg.Redis.Addr(), Password: cfg.Redis.Password, DB: cfg.Redis.DB,
	})
	defer rdb.Close()
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to Redis")
	}
	log.Info().Msg("connected to Redis")

	// JWT Verifier
	jwtVerifier, err := dealjwt.NewVerifier(cfg.JWT.PublicKeyPath, cfg.JWT.Issuer, cfg.JWT.Audience)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize JWT verifier")
	}

	// Repositories
	postRepo := pgRepo.NewPostRepo(db)
	mediaRepo := pgRepo.NewPostMediaRepo(db)
	commentRepo := pgRepo.NewCommentRepo(db)
	reactionRepo := pgRepo.NewReactionRepo(db)
	savedPostRepo := pgRepo.NewSavedPostRepo(db)
	hashtagRepo := pgRepo.NewHashtagRepo(db)
	reportRepo := pgRepo.NewReportRepo(db)
	cacheRepo := redisRepo.NewCacheRepo(rdb)

	// Application services
	postSvc := application.NewPostService(postRepo, mediaRepo, hashtagRepo, savedPostRepo, reactionRepo, cacheRepo, log)
	commentSvc := application.NewCommentService(commentRepo, postRepo, log)
	reactionSvc := application.NewReactionService(reactionRepo, postRepo, commentRepo, reportRepo, cacheRepo, log)

	// Handlers
	postHandler := handler.NewPostHandler(postSvc)
	commentHandler := handler.NewCommentHandler(commentSvc)
	reactionHandler := handler.NewReactionHandler(reactionSvc)

	// Server
	engine := httpTransport.NewServer(
		httpTransport.ServerConfig{Port: cfg.App.Port},
		rdb, jwtVerifier,
		postHandler, commentHandler, reactionHandler, log,
	)

	srv := &http.Server{
		Addr: ":" + cfg.App.Port, Handler: engine,
		ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second,
	}

	go func() {
		log.Info().Str("port", cfg.App.Port).Msg("content service listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down content service...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server forced shutdown")
	}
	log.Info().Msg("content service stopped")
}
