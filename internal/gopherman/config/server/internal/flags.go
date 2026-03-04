package internal

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Options struct {
	Address        string `env:"RUN_ADDRESS"`
	DatabaseURL    string `env:"DATABASE_URI"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func NewOptions(args []string) (*Options, error) {
	opt, err := parseArgs(args)
	if err != nil {
		return nil, fmt.Errorf("ERROR: cannot parse arguments: %w", err)
	}
	err = opt.parseEnv()
	if err != nil {
		return nil, fmt.Errorf("ERROR: cannot parse environment variables: %w", err)
	}
	return opt, nil
}
func parseArgs(args []string) (*Options, error) {
	flags := flag.NewFlagSet("server", flag.ContinueOnError)
	opt := new(Options)
	flags.StringVar(&opt.Address, "a", "localhost", "Address of the server")
	flags.StringVar(&opt.DatabaseURL, "d", "", "Database URL")
	flags.StringVar(&opt.DatabaseURL, "db", "", "Address of accrual server")
	err := flags.Parse(args)
	if err != nil {
		return nil, err
	}
	return opt, nil
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
