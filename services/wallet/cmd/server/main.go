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

	"github.com/dealance/services/wallet/config"
	"github.com/dealance/services/wallet/internal/application"
	pgRepo "github.com/dealance/services/wallet/internal/infrastructure/postgres"
	redisRepo "github.com/dealance/services/wallet/internal/infrastructure/redis"
	httpTransport "github.com/dealance/services/wallet/internal/transport/http"
	"github.com/dealance/services/wallet/internal/transport/http/handler"
	dealjwt "github.com/dealance/shared/pkg/jwt"
	"github.com/dealance/shared/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil { fmt.Fprintf(os.Stderr, "config: %v\n", err); os.Exit(1) }

	log := logger.New(cfg.App.Env)
	log = logger.WithService(log, "dealance-wallet")
	log.Info().Str("env", cfg.App.Env).Str("port", cfg.App.Port).Msg("starting wallet service")

	db, err := sqlx.Connect("pgx", cfg.DB.DSN())
	if err != nil { log.Fatal().Err(err).Msg("pg connect failed") }
	defer db.Close()
	db.SetMaxOpenConns(25); db.SetMaxIdleConns(10); db.SetConnMaxLifetime(5 * time.Minute)

	rdb := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr(), Password: cfg.Redis.Password, DB: cfg.Redis.DB})
	defer rdb.Close()
	if err := rdb.Ping(context.Background()).Err(); err != nil { log.Fatal().Err(err).Msg("redis failed") }

	jwtVerifier, err := dealjwt.NewVerifier(cfg.JWT.PublicKeyPath, cfg.JWT.Issuer, cfg.JWT.Audience)
	if err != nil { log.Fatal().Err(err).Msg("jwt verifier failed") }

	walletRepo := pgRepo.NewWalletRepo(db)
	ledgerRepo := pgRepo.NewLedgerRepo(db)
	txRepo := pgRepo.NewTransactionRepo(db)
	bankRepo := pgRepo.NewBankAccountRepo(db)
	cacheRepo := redisRepo.NewCacheRepo(rdb)

	walletSvc := application.NewWalletService(walletRepo, ledgerRepo, txRepo, bankRepo, cacheRepo, log)
	walletHandler := handler.NewWalletHandler(walletSvc)
	engine := httpTransport.NewServer(httpTransport.ServerConfig{Port: cfg.App.Port}, rdb, jwtVerifier, walletHandler, log)

	srv := &http.Server{Addr: ":" + cfg.App.Port, Handler: engine, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second}
	go func() {
		log.Info().Str("port", cfg.App.Port).Msg("wallet service listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed { log.Fatal().Err(err).Msg("server failed") }
	}()

	quit := make(chan os.Signal, 1); signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM); <-quit
	log.Info().Msg("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second); defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Info().Msg("wallet service stopped")
}
