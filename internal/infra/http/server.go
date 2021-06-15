package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/imouto1994/yume/internal/infra/config"
	"go.uber.org/zap"
)

func RunServer(handler http.Handler, cfg *config.Config, logger *zap.Logger) {
	address := fmt.Sprintf("localhost:%s", cfg.HTTPPort)
	server := &http.Server{
		Handler:      handler,
		Addr:         address,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	logger.Info("Server started successfully", zap.String("address", address))
	server.ListenAndServe()
}
