// Package logger builds configured zap loggers for app commands.
package logger

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
)

const logDir = "logs/"
const (
	// TypeModeProduction enables production logger configuration.
	TypeModeProduction = "production"
	// TypeModeDevelopment enables development logger configuration.
	TypeModeDevelopment = "development"
	// TypeModeTest enables development-style logger for tests.
	TypeModeTest = "test"
)
const (
	// ServerLgr is default logger name for HTTP server command.
	ServerLgr = "server"
)

// Initialize creates logger by mode and command type.
func Initialize(mode, cmdType string) (*zap.Logger, error) {
	var err error
	var log *zap.Logger
	mode = strings.TrimSpace(mode)
	switch mode {
	case TypeModeProduction:
		log, err = createProductionLogger(cmdType)
	case TypeModeDevelopment, TypeModeTest:
		log, err = createDevelopmentLogger()
	default:
		err = errors.New("invalid mode")
	}
	if err != nil {
		return nil, fmt.Errorf("error creating logger: %w", err)
	}

	return log, nil
}
func createProductionLogger(cmdType string) (*zap.Logger, error) {
	err := os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("could not create log directory: %w", err)
	}
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{logDir + cmdType + "_info.log", "stdout"}
	config.ErrorOutputPaths = []string{logDir + cmdType + "_errors.log", "stderr"}
	logger, err := config.Build()
	return logger, err
}
func createDevelopmentLogger() (*zap.Logger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("could not create development logger: %w", err)
	}
	return logger, nil
}
