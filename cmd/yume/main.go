package main

import (
	"github.com/go-playground/validator"
	"go.uber.org/zap"

	"github.com/imouto1994/yume/internal/infra/config"
	httpProtocol "github.com/imouto1994/yume/internal/infra/http"
	"github.com/imouto1994/yume/internal/infra/migration"
	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/route"
)

func main() {
	// Initialize global logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	// Initialize validator
	v := validator.New()

	// Initialize configuration
	cfg, err := config.Initialize(v)
	if err != nil {
		zap.L().Fatal("failed to load configurations", zap.Error(err))
	}

	// Initialize migration
	migration.UpLatest()

	// Intitialize database client
	db, err := sqlite.Connect()
	if err != nil {
		zap.L().Fatal("failed to establish connection to database", zap.Error(err))
	}
	defer db.Close()

	router := route.CreateRouter(cfg, db, v)
	httpProtocol.RunServer(router, cfg)
}
