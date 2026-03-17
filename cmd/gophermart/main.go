package main

import (
	"context"
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
)

func main() {
	if err := run(os.Args[:1]); err != nil {
		log.Fatal(fmt.Errorf("failed initialization server: %w", err))
	}
}
func run(args []string) error {
	cfg, err := server.NewConfig(args)
	if err != nil {
		return fmt.Errorf("error creating config: %w", err)
	}
	lgr, err := logger.Initialize(cfg.Mode, constant.ServerType)
	if err != nil {
		return fmt.Errorf("error initializing logger: %w", err)
	}
	err = migrations.Migrate(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("error migrating database: %w", err)
	}
	dbConfig := db.NewCfg(cfg.DatabaseURL)
	newConn, err := conn.NewConn(dbConfig)
	if err != nil {
		return fmt.Errorf("error creating database connection: %w", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	repos := repository.Repositories{
		User:       repository.NewUserRepository(newConn),
		Order:      repository.NewOrderRepository(newConn),
		Withdrawal: repository.NewWithdrawalRepository(newConn),
	}
	newHandler := api.NewHandler(newConn, repos, lgr)
	if cfg.AccrualUseMock {
		_ = accrual.NewMocker(cfg)
	}
	accrualClient := accrual.NewClient(ctx, newConn, repos, cfg)
	go accrualClient.StartPoolAccrual(ctx)
	go accrualClient.CollectResults(ctx)
	if err := http.ListenAndServe(cfg.Address, router.GetRouter(newHandler)); err != nil {
		return fmt.Errorf("error starting HTTP server: %w", err)
	}

	return nil
}
