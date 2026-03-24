// Package server builds runtime application configuration.
package server

import (
	"fmt"
	"gophermart-loyalty/internal/gopherman/config/server/internal"
	"gophermart-loyalty/internal/gopherman/errors/labelerrors"
)

const (
	labelFlags = "FLAGS"
)

// Config is validated server configuration.
type Config struct {
	internal.Options
}

// NewConfig parses and validates command-line/env options.
func NewConfig(args []string) (*Config, error) {
	opt, err := internal.NewOptions(args)
	if err != nil {
		return nil, fmt.Errorf("error parsing options: %v", err)
	}
	err = opt.ValidateOptions()
	if err != nil {
		return nil, labelerrors.NewLabelError(labelFlags, fmt.Errorf("error validating options: %w", err))
	}
	cfg := &Config{
		*opt,
	}
	return cfg, nil
}
