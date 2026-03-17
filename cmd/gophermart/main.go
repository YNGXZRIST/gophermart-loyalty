package main

import (
	"context"
	"errors"
	"fmt"
	"gophermart-loyalty/internal/gopherman/accrual"
	"gophermart-loyalty/internal/gopherman/config/db"
	"gophermart-loyalty/internal/gopherman/config/server"
	"gophermart-loyalty/internal/gopherman/constant"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/handler/api"
	"gophermart-loyalty/internal/gopherman/logger"
	"gophermart-loyalty/internal/gopherman/repository"
	"gophermart-loyalty/internal/gopherman/router"
	"gophermart-loyalty/migrations"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(fmt.Errorf("failed initialization server: %w", err))
	}
}
func run(args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := initConfig(args)
	if err != nil {
		return err
	}

	lgr, err := initLogger(cfg)
	if err != nil {
		return err
	}

	if err := runMigrations(cfg); err != nil {
		return err
	}

	dbConn, err := initDB(cfg)
	if err != nil {
		return err
	}

	repos := initRepos(dbConn)
	handler := initHTTPHandler(dbConn, repos, lgr)

	startAccrual(ctx, cfg, dbConn, repos)

	return startHTTPServer(ctx, cfg, handler)
}

func initConfig(args []string) (*server.Config, error) {
	cfg, err := server.NewConfig(args)
	if err != nil {
		return nil, fmt.Errorf("error creating config: %w", err)
	}
	return cfg, nil
}

func initLogger(cfg *server.Config) (*zap.Logger, error) {
	lgr, err := logger.Initialize(cfg.Mode, constant.ServerType)
	if err != nil {
		return nil, fmt.Errorf("error initializing logger: %w", err)
	}
	return lgr, nil
}
func runMigrations(cfg *server.Config) error {
	if err := migrations.Migrate(cfg.DatabaseURL); err != nil {
		return fmt.Errorf("error migrating database: %w", err)
	}
	return nil
}
func initRepos(dbConn *conn.DB) repository.Repositories {
	return repository.Repositories{
		User:       repository.NewUserRepository(dbConn),
		Order:      repository.NewOrderRepository(dbConn),
		Withdrawal: repository.NewWithdrawalRepository(dbConn),
	}
}
func initDB(cfg *server.Config) (*conn.DB, error) {
	dbConfig := db.NewCfg(cfg.DatabaseURL)
	newConn, err := conn.NewConn(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating database connection: %w", err)
	}
	return newConn, nil
}

func initHTTPHandler(dbConn *conn.DB, repos repository.Repositories, lgr *zap.Logger) *api.Handler {
	return api.NewHandler(dbConn, repos, lgr)
}

func startAccrual(ctx context.Context, cfg *server.Config, dbConn *conn.DB, repos repository.Repositories) {
	if cfg.AccrualUseMock {
		_ = accrual.NewMocker(cfg)
	}

	accrualClient := accrual.NewClient(ctx, dbConn, repos, cfg)
	go accrualClient.StartPoolAccrual(ctx)
	go accrualClient.CollectResults(ctx)
}

func startHTTPServer(ctx context.Context, cfg *server.Config, handler *api.Handler) error {
	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: router.GetRouter(handler),
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}
		return nil
	case err := <-errCh:
		return fmt.Errorf("error starting HTTP server: %w", err)
	}
}
