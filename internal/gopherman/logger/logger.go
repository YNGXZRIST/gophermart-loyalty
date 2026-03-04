package logger

import (
	"errors"
	"fmt"
	"gophermart-loyalty/internal/gopherman/constant"
	"os"

	"go.uber.org/zap"
)

const logDir = "logs/"

func Initialize(mode, cmdType string) (*zap.Logger, error) {
	var err error
	var log *zap.Logger
	switch mode {
	case constant.TypeModeProduction:
		log, err = createProductionLogger(cmdType)
	case constant.TypeModeDevelopment:
	case constant.TypeModeTest:
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
