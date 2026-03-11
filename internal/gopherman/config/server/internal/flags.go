package internal

import (
	"flag"
	"fmt"
	"gophermart-loyalty/internal/gopherman/constant"

	"github.com/caarlos0/env/v11"
)

type Options struct {
	Address        string `env:"RUN_ADDRESS"`
	DatabaseURL    string `env:"DATABASE_URI"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	Mode           string `env:"MODE"`
}

func NewOptions(args []string) (*Options, error) {
	opt := new(Options)
	err := opt.parseEnv()
	if err != nil {
		return nil, fmt.Errorf("ERROR: cannot parse environment variables: %w", err)
	}
	err = opt.parseArgs(args)
	if err != nil {
		return nil, fmt.Errorf("ERROR: cannot parse arguments: %w", err)
	}

	return opt, nil
}
func (opt *Options) parseArgs(args []string) error {
	flags := flag.NewFlagSet("server", flag.ContinueOnError)
	flags.StringVar(&opt.Address, "a", opt.Address, "Address of the server")
	flags.StringVar(&opt.DatabaseURL, "d", opt.DatabaseURL, "Database URL")
	flags.StringVar(&opt.AccrualAddress, "r", opt.AccrualAddress, "Address of accrual server")
	flags.StringVar(&opt.Mode, "m", constant.TypeModeDefault, "Server mode")
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
