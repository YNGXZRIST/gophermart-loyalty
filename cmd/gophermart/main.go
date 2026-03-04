package main

import (
	"fmt"
	"gophermart-loyalty/internal/gopherman/config/db"
	"gophermart-loyalty/internal/gopherman/config/server"
	"gophermart-loyalty/internal/gopherman/constant"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/handler/api"
	"gophermart-loyalty/internal/gopherman/logger"
	"gophermart-loyalty/internal/gopherman/router"
	"gophermart-loyalty/migrations"
	"log"
	"net/http"
	"os"
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
	newHandler := api.NewHandler(newConn, lgr)
	if err := http.ListenAndServe(cfg.Address, router.GetRouter(newHandler)); err != nil {
		return fmt.Errorf("error starting HTTP server: %w", err)
	}

	return nil
}
