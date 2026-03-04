package main

import (
	"fmt"
	"gophermart-loyalty/internal/gopherman/config/db"
	"gophermart-loyalty/internal/gopherman/config/server"
	"gophermart-loyalty/internal/gopherman/db/conn"
	"log"
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
	conn, err := conn.NewConn(dbConfig)
	if err != nil {
		return fmt.Errorf("error creating database connection: %w", err)
	}
	fmt.Println(conn)
	return nil
}
