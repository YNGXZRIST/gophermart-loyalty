package internal

import (
	"flag"
	"fmt"
	"gophermart-loyalty/internal/gopherman/logger"
	"log"
	"net"
	"strconv"

	"github.com/caarlos0/env/v11"
)

const (
	AccrualWorkerCountDefault = 10
	TypeModeDefault           = logger.TypeModeDevelopment
)

type Options struct {
	Address            string `env:"RUN_ADDRESS"`
	DatabaseURL        string `env:"DATABASE_URI"`
	AccrualAddress     string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	Mode               string `env:"MODE"`
	AccrualWorkerCount int    `env:"ACCRUAL_WORKER_COUNT"`
	AccrualUseMock     bool   `env:"ACCRUAL_USE_MOCK"`
}

func NewOptions(args []string) (*Options, error) {
	opt := new(Options)
	if err := opt.parseArgs(args); err != nil {
		return nil, fmt.Errorf("ERROR: cannot parse arguments: %w", err)
	}
	if err := opt.parseEnv(); err != nil {
		return nil, fmt.Errorf("ERROR: cannot parse environment variables: %w", err)
	}
	if opt.AccrualUseMock {
		ln, err := net.Listen("tcp", ":0")
		if err != nil {
			log.Fatal(err)
		}
		port := ln.Addr().(*net.TCPAddr).Port
		ln.Close()
		opt.AccrualAddress = "http://localhost:" + strconv.Itoa(port)
		fmt.Printf("accrual address: %s\n", opt.AccrualAddress)
	}
	return opt, nil
}

func parseArgs(args []string) (*Options, error) {
	opt := &Options{
		Address:            "localhost",
		Mode:               TypeModeDefault,
		AccrualWorkerCount: AccrualWorkerCountDefault,
	}
	if err := opt.parseArgs(args); err != nil {
		return nil, err
	}
	return opt, nil
}

func (opt *Options) parseArgs(args []string) error {
	flags := flag.NewFlagSet("server", flag.ContinueOnError)
	flags.StringVar(&opt.Address, "a", opt.Address, "Address of the server")
	flags.StringVar(&opt.DatabaseURL, "d", opt.DatabaseURL, "Database URL")
	flags.StringVar(&opt.DatabaseURL, "db", opt.DatabaseURL, "Database URL (alias)")
	flags.StringVar(&opt.AccrualAddress, "r", opt.AccrualAddress, "Address of accrual server")
	flags.StringVar(&opt.Mode, "m", TypeModeDefault, "Server mode")
	flags.IntVar(&opt.AccrualWorkerCount, "ac", AccrualWorkerCountDefault, "Worker accrual count")
	flags.BoolVar(&opt.AccrualUseMock, "accrual-mock", false, "Use free port for accrual address (start mock server on this port)")
	err := flags.Parse(args)
	if err != nil {
		return err
	}
	return nil
}
func (opt *Options) parseEnv() error {
	err := env.Parse(opt)
	if err != nil {
		return err
	}
	return nil
}
func (opt *Options) ValidateOptions() error {
	if opt.Address == "" {
		return fmt.Errorf("address is required")
	}
	if opt.DatabaseURL == "" {
		return fmt.Errorf("database URL is required")
	}
	if opt.AccrualAddress == "" {
		return fmt.Errorf("accrual address is required")
	}
	return nil
}
