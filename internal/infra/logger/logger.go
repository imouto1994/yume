package logger

import "go.uber.org/zap"

func Initialize() *zap.Logger {
	logger, _ := zap.NewProduction()
	return logger
}
