package server

import (
	"fmt"
	"gophermart-loyalty/internal/gopherman/config/server/internal"
	"gophermart-loyalty/internal/gopherman/constant"
	"gophermart-loyalty/internal/gopherman/errors/labelerrors"
)

type Config struct {
	internal.Options
}

func NewConfig(args []string) (*Config, error) {
	opt, err := internal.NewOptions(args)
	if err != nil {
		return nil, fmt.Errorf("error parsing options: %v", err)
	}
	err = opt.ValidateOptions()
	if err != nil {
		return nil, labelerrors.NewLabelError(constant.LabelFlags, fmt.Errorf("error validating options: %w", err))
	}
	cfg := &Config{
		*opt,
	}
	return cfg, nil
}
