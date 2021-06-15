package route

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/imouto1994/yume/internal/infra/config"
	"github.com/imouto1994/yume/internal/repository"
	"github.com/imouto1994/yume/internal/service"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func CreateRouter(cfg *config.Config, db *sqlx.DB, logger *zap.Logger) http.Handler {
	// Initialize repositories
	libraryRepository := repository.NewLibraryRepository(db)

	// Initialize services
	libraryService := service.NewLibraryService(libraryRepository, logger)

	// Initialize handlers
	libraryHandler := NewLibraryHandler(libraryService)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Mount("/api/library", libraryHandler.InitializeRoutes())

	return r
}
