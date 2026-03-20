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

	"github.com/dealance/services/startup/config"
	"github.com/dealance/services/startup/internal/application"
	pgRepo "github.com/dealance/services/startup/internal/infrastructure/postgres"
	redisRepo "github.com/dealance/services/startup/internal/infrastructure/redis"
	httpTransport "github.com/dealance/services/startup/internal/transport/http"
	"github.com/dealance/services/startup/internal/transport/http/handler"
	dealjwt "github.com/dealance/shared/pkg/jwt"
	"github.com/dealance/shared/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil { fmt.Fprintf(os.Stderr, "config: %v\n", err); os.Exit(1) }

	log := logger.New(cfg.App.Env)
	log = logger.WithService(log, "dealance-startup")
	log.Info().Str("env", cfg.App.Env).Str("port", cfg.App.Port).Msg("starting startup service")

	db, err := sqlx.Connect("pgx", cfg.DB.DSN())
	if err != nil { log.Fatal().Err(err).Msg("pg connect failed") }
	defer db.Close()
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	rdb := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr(), Password: cfg.Redis.Password, DB: cfg.Redis.DB})
	defer rdb.Close()
	if err := rdb.Ping(context.Background()).Err(); err != nil { log.Fatal().Err(err).Msg("redis connect failed") }

	jwtVerifier, err := dealjwt.NewVerifier(cfg.JWT.PublicKeyPath, cfg.JWT.Issuer, cfg.JWT.Audience)
	if err != nil { log.Fatal().Err(err).Msg("jwt verifier failed") }

	// Repos
	startupRepo := pgRepo.NewStartupRepo(db)
	fundingRepo := pgRepo.NewFundingRoundRepo(db)
	teamRepo := pgRepo.NewTeamMemberRepo(db)
	followRepo := pgRepo.NewStartupFollowRepo(db)
	cacheRepo := redisRepo.NewCacheRepo(rdb)

	// Service
	startupSvc := application.NewStartupService(startupRepo, fundingRepo, teamRepo, followRepo, cacheRepo, log)

	// Handler + Server
	startupHandler := handler.NewStartupHandler(startupSvc)
	engine := httpTransport.NewServer(httpTransport.ServerConfig{Port: cfg.App.Port}, rdb, jwtVerifier, startupHandler, log)

	srv := &http.Server{Addr: ":" + cfg.App.Port, Handler: engine, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second}

	go func() {
		log.Info().Str("port", cfg.App.Port).Msg("startup service listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed { log.Fatal().Err(err).Msg("server failed") }
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Info().Msg("startup service stopped")
}
