package main

import (
	"fmt"
	"gophermart-loyalty/internal/gopherman/config/db"
	"gophermart-loyalty/internal/gopherman/config/server"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"gophermart-loyalty/internal/gopherman/handler/api"
	"gophermart-loyalty/internal/gopherman/router"
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
	dbConfig := db.NewCfg(cfg.DatabaseURL)
	newConn, err := conn.NewConn(dbConfig)
	if err != nil {
		return fmt.Errorf("error creating database connection: %w", err)
	}
	newHandler := api.NewHandler(newConn)
	if err := http.ListenAndServe(cfg.Address, router.GetRouter(newHandler)); err != nil {
		return fmt.Errorf("error starting HTTP server: %w", err)
	}
	fmt.Println(newConn)
	return nil
}
