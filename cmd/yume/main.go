package main

import (
	"fmt"
	"log"

	"github.com/imouto1994/yume/internal/infra/config"
	httpProtocol "github.com/imouto1994/yume/internal/infra/http"
	"github.com/imouto1994/yume/internal/infra/logger"
	"github.com/imouto1994/yume/internal/infra/sqlite"
	"github.com/imouto1994/yume/internal/route"
)

func main() {
	// Initialize configuration
	cfg, err := config.Initialize()
	if err != nil {
		log.Fatalln(fmt.Errorf("Unable to load configurations: '%v'", err))
	}

	// Intitialize database client
	db, err := sqlite.Connect()
	if err != nil {
		log.Fatalln(fmt.Errorf("Unable to establish connection to database: '%v'", err))
	}

	lg := logger.Initialize()

	router := route.CreateRouter(cfg, db, lg)
	httpProtocol.RunServer(router, cfg, lg)
}
